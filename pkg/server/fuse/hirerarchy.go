package fuse

import (
	"context"
	"fmt"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

var _ fs.InodeEmbedder = (*fuseNode)(nil)
var _ fs.NodeStatfser = (*fuseNode)(nil)
var _ fs.NodeReaddirer = (*fuseNode)(nil)
var _ fs.NodeLookuper = (*fuseNode)(nil)
var _ fs.NodeOpendirer = (*fuseNode)(nil)

func (node *fuseNode) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	// n files, n metadata files, dir meta
	out.Files = uint64(2*len(node.Children()) + 1)
	return fs.OK
}

func (node *fuseNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	children := node.kdbNode.Children()
	entries := make([]fuse.DirEntry, 0, 2*len(children)+1)

	entries = append(entries, fuse.DirEntry{Mode: node.server.fileMode | fuse.S_IFREG, Name: ".yml", Ino: ino(node.kdbNode, true)})
	for _, child := range children {
		entries = append(entries, fuse.DirEntry{Mode: fuse.S_IFREG | node.server.fileMode, Name: child.Name, Ino: ino(child, false)})
		if child.HasChildren {
			entries[len(entries)-1].Mode = fuse.S_IFDIR | node.server.dirMode
		} else {
			metaName := fmt.Sprintf("%s.yml", child.Name)
			entries = append(entries, fuse.DirEntry{Mode: fuse.S_IFREG | node.server.fileMode, Name: metaName, Ino: ino(child, true)})
		}
	}
	return fs.NewListDirStream(entries), fs.OK
}

func (node *fuseNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if node.isMetadataNode {
		return nil, syscall.ENOENT
	}

	n, meta, ok := childOrMeta(node.kdbNode, name, node.server.metaSuffix)
	if !ok {
		return nil, syscall.ENOENT
	}

	time := uint64(node.kdbNode.Time.Unix())
	out.Mode = fuse.S_IFDIR | node.server.dirMode
	out.Atime, out.Mtime, out.Ctime = time, time, time
	if !n.HasChildren || meta {
		out.Mode = fuse.S_IFREG | node.server.fileMode
		out.Size = uint64(n.Size)
		out.Nlink = 1
	}
	out.Owner = fuse.Owner{Uid: node.server.userID, Gid: node.server.groupID}

	embedder := &fuseNode{server: node.server, kdbNode: n, isMetadataNode: meta}
	return node.NewInode(ctx, embedder, fs.StableAttr{Mode: out.Mode, Ino: ino(n, meta)}), fs.OK
}

func (node *fuseNode) Opendir(ctx context.Context) syscall.Errno {
	return fs.OK
}
