package fuse

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

var _ fs.InodeEmbedder = (*fuseNode)(nil)
var _ fs.NodeStatfser = (*fuseNode)(nil)
var _ fs.NodeReaddirer = (*fuseNode)(nil)
var _ fs.NodeLookuper = (*fuseNode)(nil)
var _ fs.NodeOpendirer = (*fuseNode)(nil)

const (
	fileMode = 0660
	dirMode  = 0751
)

func (node *fuseNode) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	out.Files = uint64(len(node.Children()))
	return fs.OK
}

func (node *fuseNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	children := node.kdbNode.Children()
	entries := make([]fuse.DirEntry, len(children))
	i := 0
	for _, child := range children {
		entries[i] = fuse.DirEntry{Mode: dirMode | fuse.S_IFDIR, Name: child.Name, Ino: 0}
		i++
	}
	return fs.NewListDirStream(entries), fs.OK
}

func (node *fuseNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	n, ok := node.kdbNode.Child(name)
	if !ok {
		return nil, syscall.ENOENT
	}

	time := uint64(node.kdbNode.Time.Unix())
	out.Mode = fuse.S_IFDIR | dirMode
	out.Atime, out.Mtime, out.Ctime = time, time, time
	if !n.HasChildren {
		out.Mode = fuse.S_IFREG | fileMode
		out.Size = uint64(n.Size)
		out.Nlink = 1
	}
	out.Owner = fuse.Owner{Uid: 1000, Gid: 100}

	embedder := &fuseNode{server: node.server, kdbNode: n}
	return node.NewInode(ctx, embedder, fs.StableAttr{Mode: out.Mode}), fs.OK
}

func (node *fuseNode) Opendir(ctx context.Context) syscall.Errno {
	return fs.OK
}
