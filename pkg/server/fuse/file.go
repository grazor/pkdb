package fuse

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"syscall"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/server"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func newFuseHandle(node *kdb.KdbNode, serv *fuseServer, flags uint32) fs.FileHandle {
	wSize := node.Size
	if flags&syscall.O_TRUNC != syscall.O_TRUNC {
		wSize = 0
	}

	buf := make([]byte, node.Size)
	handle := &fuseHandle{kdbNode: node, server: serv}

	// TODO: handle O_RDONLY == 0x0
	if flags&syscall.O_RDONLY == syscall.O_RDONLY || flags&syscall.O_RDWR == syscall.O_RDWR {
		reader, err := node.Reader(0)
		if err != nil {
			serv.errors <- server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("unable to open node for read %v", node),
			}
		}
		defer reader.Close()

		if _, err := reader.Read(buf); err != nil && err != io.EOF {
			serv.errors <- server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("unable to read %v", node),
			}
		}
		handle.reader = bytes.NewReader(buf)
	}

	if flags&syscall.O_APPEND == syscall.O_APPEND {
		handle.writer = bytes.NewBuffer(buf)
	} else if flags&syscall.O_WRONLY == syscall.O_WRONLY || flags&syscall.O_RDWR == syscall.O_RDWR {
		handle.writer = bytes.NewBuffer(make([]byte, wSize))
	}

	return handle
}

type fuseHandle struct {
	kdbNode *kdb.KdbNode
	server  *fuseServer

	reader *bytes.Reader
	writer *bytes.Buffer
}

var _ fs.FileHandle = (*fuseHandle)(nil)
var _ fs.FileAllocater = (*fuseHandle)(nil)
var _ fs.FileFlusher = (*fuseHandle)(nil)
var _ fs.FileFsyncer = (*fuseHandle)(nil)
var _ fs.FileGetattrer = (*fuseHandle)(nil)
var _ fs.FileLseeker = (*fuseHandle)(nil)
var _ fs.FileReader = (*fuseHandle)(nil)
var _ fs.FileReleaser = (*fuseHandle)(nil)
var _ fs.FileSetattrer = (*fuseHandle)(nil)
var _ fs.FileWriter = (*fuseHandle)(nil)

func (handle *fuseHandle) Allocate(ctx context.Context, off uint64, size uint64, mode uint32) syscall.Errno {
	return fs.OK
}

func (handle *fuseHandle) Flush(ctx context.Context) syscall.Errno {
	if handle.writer == nil {
		return fs.OK
	}

	writeCloser, err := handle.kdbNode.Writer(0)
	if err != nil {
		// TODO: use Wrap to wrap errors
		// TODO: add trace to errors
		handle.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to open for write %v", handle.kdbNode.Path),
		}
		return syscall.EFAULT
	}

	_, err = writeCloser.Write(handle.writer.Bytes())
	if err != nil {
		writeCloser.Close()
		handle.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("failed to write %v", handle.kdbNode.Path),
		}
		return syscall.EFAULT
	}

	err = writeCloser.Close()
	if err != nil {
		writeCloser.Close()
		handle.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("error closing %v", handle.kdbNode.Path),
		}
		return syscall.EFAULT
	}

	return fs.OK
}

func (handle *fuseHandle) Fsync(ctx context.Context, flags uint32) syscall.Errno {
	return handle.Flush(ctx)
}

func (handle *fuseHandle) Getattr(ctx context.Context, out *fuse.AttrOut) syscall.Errno {
	time := uint64(handle.kdbNode.Time.Unix())
	out.Mode = fuse.S_IFDIR | handle.server.dirMode
	out.Atime, out.Mtime, out.Ctime = time, time, time
	if !handle.kdbNode.HasChildren {
		out.Mode = handle.server.fileMode | fuse.S_IFREG
		out.Size = uint64(handle.kdbNode.Size)
		out.Nlink = 1
	}
	out.Owner = fuse.Owner{Uid: handle.server.userID, Gid: handle.server.groupID}

	return fs.OK
}

func (handle *fuseHandle) Lseek(ctx context.Context, off uint64, whence uint32) (uint64, syscall.Errno) {
	if handle.reader == nil {
		handle.server.errors <- server.ServerError{
			Message: fmt.Sprintf("bufer is empty %v", handle.kdbNode),
		}
		return 0, syscall.EFAULT
	}
	abs, err := handle.reader.Seek(int64(off), int(whence))
	if err != nil {
		handle.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("seek failed for %s", handle.kdbNode),
		}
		return 0, syscall.EFAULT
	}
	return uint64(abs), fs.OK
}

func (handle *fuseHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	if handle.reader == nil {
		handle.server.errors <- server.ServerError{
			Message: fmt.Sprintf("bufer is empty %v", handle.kdbNode),
		}
		return nil, syscall.EFAULT
	}
	handle.reader.Read(dest)
	return fuse.ReadResultData(dest), fs.OK
}

func (handle *fuseHandle) Release(ctx context.Context) syscall.Errno {
	return handle.Flush(ctx)
}

func (handle *fuseHandle) Setattr(ctx context.Context, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	return setattr(ctx, handle.kdbNode, handle.server, in, out)
}

func (handle *fuseHandle) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
	//TODO: dont ignore offset parameter
	if handle.writer == nil {
		handle.server.errors <- server.ServerError{
			Message: fmt.Sprintf("bufer is empty %v", handle.kdbNode),
		}
		return 0, syscall.EFAULT
	}

	n, err := handle.writer.Write(data)
	if err != nil {
		handle.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("write failed for %s", handle.kdbNode),
		}
		return 0, syscall.EFAULT
	}
	return uint32(n), fs.OK
}
