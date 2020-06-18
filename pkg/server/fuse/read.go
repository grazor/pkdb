package fuse

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
)

var _ fs.NodeOpener = (*fuseNode)(nil)

func (node *fuseNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if node.kdbNode.HasChildren {
		return nil, 0, syscall.ENOTSUP
	}
	return newFuseHandle(node.kdbNode), 0, fs.OK
}

