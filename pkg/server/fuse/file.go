package fuse

import (
	"context"
	"fmt"
	"io"
	"syscall"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/server"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func newFuseHandle(node *kdb.KdbNode) fs.FileHandle {
	return &fuseHandle{kdbNode: node}
}

type fuseHandle struct {
	kdbNode *kdb.KdbNode
}

var _ fs.FileHandle = (*fuseHandle)(nil)
var _ fs.FileReleaser = (*fuseHandle)(nil)
var _ fs.FileGetattrer = (*fuseHandle)(nil)
var _ fs.FileReader = (*fuseHandle)(nil)
var _ fs.FileWriter = (*fuseHandle)(nil)
var _ fs.FileGetlker = (*fuseHandle)(nil)
var _ fs.FileSetlker = (*fuseHandle)(nil)
var _ fs.FileLseeker = (*fuseHandle)(nil)
var _ fs.FileFlusher = (*fuseHandle)(nil)
var _ fs.FileFsyncer = (*fuseHandle)(nil)
var _ fs.FileSetattrer = (*fuseHandle)(nil)
var _ fs.FileAllocater = (*fuseHandle)(nil)

func (node *fuseNode) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	//TODO: support touch -m
	time := uint64(node.kdbNode.Time.Unix())
	out.Mode = fuse.S_IFDIR | dirMode
	out.Atime, out.Mtime, out.Ctime = time, time, time
	if !node.kdbNode.HasChildren {
		out.Mode = fuse.S_IFREG | fileMode
		out.Size = uint64(node.kdbNode.Size)
		out.Nlink = 1
	}
	out.Owner = fuse.Owner{Uid: 1000, Gid: 100}
	return fs.OK
}

func (node *fuseNode) Write(ctx context.Context, f fs.FileHandle, data []byte, off int64) (written uint32, errno syscall.Errno) {
	writeCloser, err := node.kdbNode.Writer(off)
	if err != nil {
		// TODO: use Wrap to wrap errors
		// TODO: add trace to errors
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to open for write %v", node.kdbNode.Path),
		}
		return 0, syscall.EFAULT
	}

	n, err := writeCloser.Write(data)
	if err != nil {
		writeCloser.Close()
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("failed to write %v", node.kdbNode.Path),
		}
		return 0, syscall.EFAULT
	}

	err = writeCloser.Close()
	if err != nil {
		writeCloser.Close()
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("error closing %v", node.kdbNode.Path),
		}
		return uint32(n), syscall.EFAULT
	}

	return uint32(n), fs.OK
}

func (node *fuseNode) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	return fs.OK
}

func (node *fuseNode) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	reader, err := node.kdbNode.Reader(off)
	if err != nil {
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to open node for read %v", node.kdbNode),
		}
	}
	defer reader.Close()

	if _, err := reader.Read(dest); err != nil && err != io.EOF {
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to read %v", node.kdbNode),
		}
	}
	return fuse.ReadResultData(dest), fs.OK
}

func (node *fuseNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	time := uint64(node.kdbNode.Time.Unix())
	out.Mode = fuse.S_IFDIR | dirMode
	out.Atime, out.Mtime, out.Ctime = time, time, time
	if !node.kdbNode.HasChildren {
		out.Mode = fileMode | fuse.S_IFREG
		out.Size = uint64(node.kdbNode.Size)
		out.Nlink = 1
	}
	out.Owner = fuse.Owner{Uid: 1000, Gid: 100}

	return fs.OK
}
