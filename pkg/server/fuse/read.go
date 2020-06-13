package fuse

import (
	"context"
	"fmt"
	"syscall"

	"github.com/grazor/pkdb/pkg/server"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

var _ fs.NodeOpener = (*fuseNode)(nil)
var _ fs.NodeReader = (*fuseNode)(nil)

func (node *fuseNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if node.kdbNode.HasChildren {
		return nil, 0, syscall.ENOTSUP
	}
	return node.kdbNode, fuse.FOPEN_KEEP_CACHE, fs.OK
}

func (node *fuseNode) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	reader, err := node.kdbNode.Open()
	if err != nil {
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to open node for read %v", node.kdbNode),
		}
	}
	defer reader.Close()

	_, err = reader.Read(dest)
	if err != nil {
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to read %v", node.kdbNode),
		}
	}
	return fuse.ReadResultData(dest), fs.OK
}
