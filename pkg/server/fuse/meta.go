package fuse

import (
	"bytes"
	"context"
	"fmt"
	"syscall"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/server"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gopkg.in/yaml.v2"
)

func newMetaHandle(node *kdb.KdbNode, serv *fuseServer, flags uint32) fs.FileHandle {
	wSize := node.Size
	if flags&syscall.O_TRUNC != syscall.O_TRUNC {
		wSize = 0
	}

	handle := &FuseMetaHandle{kdbNode: node, server: serv}
	meta, err := node.Meta()
	if err != nil {
		serv.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to get meta for %v", node),
		}
	}

	nodeMetaDump, err := yaml.Marshal(meta)
	if err != nil {
		serv.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to dump meta for %v", node),
		}
	}

	if flags&syscall.O_RDONLY == syscall.O_RDONLY || flags&syscall.O_RDWR == syscall.O_RDWR {
		handle.reader = bytes.NewReader(nodeMetaDump)
	}

	if flags&syscall.O_APPEND == syscall.O_APPEND {
		handle.writer = bytes.NewBuffer(nodeMetaDump)
	} else if flags&syscall.O_WRONLY == syscall.O_WRONLY || flags&syscall.O_RDWR == syscall.O_RDWR {
		handle.writer = bytes.NewBuffer(make([]byte, wSize))
	}

	return handle
}

type FuseMetaHandle fuseHandle

var _ fs.FileHandle = (*FuseMetaHandle)(nil)
var _ fs.FileAllocater = (*FuseMetaHandle)(nil)
var _ fs.FileFlusher = (*FuseMetaHandle)(nil)
var _ fs.FileFsyncer = (*FuseMetaHandle)(nil)
var _ fs.FileGetattrer = (*FuseMetaHandle)(nil)
var _ fs.FileLseeker = (*FuseMetaHandle)(nil)
var _ fs.FileReader = (*FuseMetaHandle)(nil)
var _ fs.FileReleaser = (*FuseMetaHandle)(nil)
var _ fs.FileSetattrer = (*FuseMetaHandle)(nil)
var _ fs.FileWriter = (*FuseMetaHandle)(nil)

func (handle *FuseMetaHandle) Allocate(ctx context.Context, off uint64, size uint64, mode uint32) syscall.Errno {
	return fs.OK
}

func (handle *FuseMetaHandle) Flush(ctx context.Context) syscall.Errno {
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

func (handle *FuseMetaHandle) Fsync(ctx context.Context, flags uint32) syscall.Errno {
	return handle.Flush(ctx)
}

func (handle *FuseMetaHandle) Getattr(ctx context.Context, out *fuse.AttrOut) syscall.Errno {
	return getattr(ctx, handle.kdbNode, true, handle.server, out)
}

func (handle *FuseMetaHandle) Lseek(ctx context.Context, off uint64, whence uint32) (uint64, syscall.Errno) {
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

func (handle *FuseMetaHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	if handle.reader == nil {
		handle.server.errors <- server.ServerError{
			Message: fmt.Sprintf("bufer is empty %v", handle.kdbNode),
		}
		return nil, syscall.EFAULT
	}
	handle.reader.Read(dest)
	return fuse.ReadResultData(dest), fs.OK
}

func (handle *FuseMetaHandle) Release(ctx context.Context) syscall.Errno {
	return handle.Flush(ctx)
}

func (handle *FuseMetaHandle) Setattr(ctx context.Context, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	return getattr(ctx, handle.kdbNode, true, handle.server, out)
}

func (handle *FuseMetaHandle) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
	if handle.writer == nil {
		handle.server.errors <- server.ServerError{
			Message: fmt.Sprintf("bufer is empty %v", handle.kdbNode),
		}
		return 0, syscall.EFAULT
	}

	meta := make(map[string]interface{})
	err := yaml.Unmarshal(data, &meta)
	if err == nil {
		handle.server.errors <- server.ServerError{
			Message: fmt.Sprintf("failed to parse yaml for %v", handle.kdbNode),
		}
		return 0, syscall.EFAULT
	}

	err = handle.kdbNode.SetMeta(meta)
	if err == nil {
		handle.server.errors <- server.ServerError{
			Message: fmt.Sprintf("failed save metadata for %v", handle.kdbNode),
		}
		return 0, syscall.EFAULT
	}

	return uint32(len(data)), fs.OK
}
