package fuse

import (
	"context"
	"fmt"
	"syscall"

	"github.com/grazor/pkdb/pkg/server"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

var _ fs.NodeCreater = (*fuseNode)(nil)
var _ fs.NodeSetattrer = (*fuseNode)(nil)
var _ fs.NodeWriter = (*fuseNode)(nil)
var _ fs.NodeUnlinker = (*fuseNode)(nil)
var _ fs.NodeRmdirer = (*fuseNode)(nil)
var _ fs.NodeRenamer = (*fuseNode)(nil)
var _ fs.NodeFsyncer = (*fuseNode)(nil)

func (node *fuseNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (newNode *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	kdbChild, err := node.kdbNode.AddChild(name, false)
	if err != nil {
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to create child for %v: %v", node.kdbNode.Path, name),
		}
		return nil, nil, 0, syscall.EFAULT
	}

	var childMode uint32 = fuse.S_IFDIR | dirMode
	if !kdbChild.HasChildren {
		childMode = fuse.S_IFREG | fileMode
	}

	embedder := &fuseNode{server: node.server, kdbNode: kdbChild}
	newNode = node.NewInode(ctx, embedder, fs.StableAttr{Mode: childMode})
	return newNode, nil, 0, fs.OK
}

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
	return 0, fs.OK
}

func (node *fuseNode) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	return fs.OK
}

func (node *fuseNode) Unlink(ctx context.Context, name string) syscall.Errno {
	return fs.OK
}

func (node *fuseNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	return fs.OK
}

func (node *fuseNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	return fs.OK
}
