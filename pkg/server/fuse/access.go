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
var _ fs.NodeOpener = (*fuseNode)(nil)
var _ fs.NodeRenamer = (*fuseNode)(nil)
var _ fs.NodeRmdirer = (*fuseNode)(nil)
var _ fs.NodeGetattrer = (*fuseNode)(nil)
var _ fs.NodeUnlinker = (*fuseNode)(nil)
var _ fs.NodeGetxattrer = (*fuseNode)(nil)
var _ fs.NodeSetattrer = (*fuseNode)(nil)
var _ fs.NodeLinker = (*fuseNode)(nil)

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

	var childMode uint32 = fuse.S_IFDIR | node.server.dirMode
	if !kdbChild.HasChildren {
		childMode = fuse.S_IFREG | node.server.fileMode
	}

	embedder := &fuseNode{server: node.server, kdbNode: kdbChild}
	newNode = node.NewInode(ctx, embedder, fs.StableAttr{Mode: childMode, Ino: kdbChild.NodeIndex})
	return newNode, newFuseHandle(kdbChild, node.server, flags), 0, fs.OK
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

	var childMode uint32 = fuse.S_IFDIR | node.server.dirMode
	if !kdbChild.HasChildren {
		childMode = fuse.S_IFREG | node.server.fileMode
	}

	embedder := &fuseNode{server: node.server, kdbNode: kdbChild}
	newNode := node.NewInode(ctx, embedder, fs.StableAttr{Mode: childMode, Ino: kdbChild.NodeIndex})
	return newNode, fs.OK
}

// On meta delete: clean metadata
// On target delete: unlink both target and meta
func (node *fuseNode) Unlink(ctx context.Context, name string) syscall.Errno {
	nodeToDelete, isMeta, ok := childOrMeta(node.kdbNode, name, node.server.metaSuffix)
	if !ok {
		node.server.errors <- server.ServerError{
			Message: fmt.Sprintf("node to delete does not exist %s / %s", node.kdbNode, name),
		}
		return syscall.ENOENT
	}

	var err error
	if isMeta {
		err = nodeToDelete.SetMeta(nil)
	} else {
		err = nodeToDelete.Delete()
	}

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
	nodeToDelete, _, ok := childOrMeta(node.kdbNode, name, node.server.metaSuffix)
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
	moveNode, _, ok := childOrMeta(node.kdbNode, name, node.server.metaSuffix)
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

func (node *fuseNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if node.kdbNode.HasChildren && !node.isMetadataNode {
		return nil, 0, syscall.ENOTSUP
	}
	return newHandle(node.kdbNode, node.isMetadataNode, node.server, flags), 0, fs.OK
}

func (node *fuseNode) Getxattr(ctx context.Context, attr string, dest []byte) (uint32, syscall.Errno) {
	return 0, syscall.ENODATA
}

func (node *fuseNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	return getattr(ctx, node.kdbNode, node.isMetadataNode, node.server, out)
}

func (node *fuseNode) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	return getattr(ctx, node.kdbNode, node.isMetadataNode, node.server, out)
}

func (node *fuseNode) Link(ctx context.Context, target fs.InodeEmbedder, name string, out *fuse.EntryOut) (newNode *fs.Inode, errno syscall.Errno) {
	targetNode, ok := node.kdbNode.Child(name)
	if !ok {
		node.server.errors <- server.ServerError{
			Message: fmt.Sprintf("linking node does not exist %v %v", node, name),
		}
		return nil, syscall.ENOENT
	}

	embedder := &fuseNode{server: node.server, kdbNode: targetNode}
	newNode = node.NewInode(ctx, embedder, fs.StableAttr{Mode: target.EmbeddedInode().StableAttr().Mode, Ino: targetNode.NodeIndex})
	return newNode, fs.OK
}
