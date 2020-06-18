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
var _ fs.NodeMkdirer = (*fuseNode)(nil)
var _ fs.NodeUnlinker = (*fuseNode)(nil)
var _ fs.NodeRmdirer = (*fuseNode)(nil)
var _ fs.NodeRenamer = (*fuseNode)(nil)

func (node *fuseNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (newNode *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	kdbChild, err := node.kdbNode.AddChild(name, false)
	if err != nil {
		// TODO: select case errors<- default
		// TODO: here and everywhere else
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
	return newNode, newFuseHandle(kdbChild), 0, fs.OK
}

func (node *fuseNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	kdbChild, err := node.kdbNode.AddChild(name, true)
	if err != nil {
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to create child for %v: %v", node.kdbNode.Path, name),
		}
		return nil, syscall.EFAULT
	}

	var childMode uint32 = fuse.S_IFDIR | dirMode
	if !kdbChild.HasChildren {
		childMode = fuse.S_IFREG | fileMode
	}

	embedder := &fuseNode{server: node.server, kdbNode: kdbChild}
	newNode := node.NewInode(ctx, embedder, fs.StableAttr{Mode: childMode})
	return newNode, fs.OK
}

func (node *fuseNode) Unlink(ctx context.Context, name string) syscall.Errno {
	nodeToDelete, ok := node.kdbNode.Child(name)
	if !ok {
		node.server.errors <- server.ServerError{
			Message: fmt.Sprintf("node to delete does not exist %s / %s", node.kdbNode, name),
		}
		return syscall.ENOENT
	}

	err := nodeToDelete.Delete()
	if err != nil {
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("failed to delete %s", nodeToDelete),
		}
		return syscall.EFAULT
	}
	return fs.OK
}

func (node *fuseNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	nodeToDelete, ok := node.kdbNode.Child(name)
	if !ok {
		node.server.errors <- server.ServerError{
			Message: fmt.Sprintf("node to delete does not exist %s / %s", node.kdbNode, name),
		}
		return syscall.ENOENT
	}

	err := nodeToDelete.Delete()

	if err != nil {
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("failed to delete %s", nodeToDelete),
		}
		return syscall.ENOTEMPTY
	}
	return fs.OK
}

func (node *fuseNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	moveNode, ok := node.kdbNode.Child(name)
	if !ok {
		node.server.errors <- server.ServerError{
			Message: fmt.Sprintf("node to move does not exist %v", moveNode),
		}
		return syscall.ENOENT
	}

	targetParent, ok := newParent.(*fuseNode)
	if !ok {
		node.server.errors <- server.ServerError{
			Message: "invalid move destination",
		}
		return syscall.EFAULT
	}

	err := moveNode.Move(targetParent.kdbNode, newName)
	if err != nil {
		node.server.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("move failed from %s to %s", moveNode, targetParent),
		}
		return syscall.EFAULT
	}

	return fs.OK
}
