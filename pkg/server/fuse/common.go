package fuse

import (
	"context"
	"strings"
	"syscall"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

const metaFileBit = 1 << 63

func getattr(ctx context.Context, node *kdb.KdbNode, server *fuseServer, out *fuse.AttrOut) syscall.Errno {
	//TODO: support touch -m
	time := node.Time
	out.Mode = fuse.S_IFDIR | server.dirMode
	out.SetTimes(&time, &time, &time)
	if !node.HasChildren {
		out.Mode = fuse.S_IFREG | server.fileMode
		out.Size = uint64(node.Size)
		out.Nlink = 1
	}
	out.Owner = fuse.Owner{Uid: server.userID, Gid: server.groupID}
	return fs.OK

}

func childOrMeta(node *kdb.KdbNode, name, suffix string) (childNode *kdb.KdbNode, meta bool, ok bool) {
	if strings.HasSuffix(name, suffix) {
		fileName := strings.TrimSuffix(name, suffix)
		childNode, ok = node.Child(fileName)
		if ok {
			return childNode, true, true
		}
	}
	childNode, ok = node.Child(name)
	return childNode, false, ok
}

func ino(node *kdb.KdbNode, metadata bool) (ino uint64) {
	ino = node.NodeIndex
	if metadata {
		ino |= metaFileBit
	} else {
		ino &^= metaFileBit
	}
	return ino
}
