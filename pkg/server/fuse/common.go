package fuse

import (
	"context"
	"fmt"
	"strings"
	"syscall"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/server"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gopkg.in/yaml.v2"
)

const metaFileBit = 1 << 63

func getNodeAttr(ctx context.Context, node *kdb.KdbNode, serv *fuseServer, out *fuse.AttrOut) syscall.Errno {
	time := node.Time
	out.Mode = fuse.S_IFDIR | serv.dirMode
	out.SetTimes(&time, &time, &time)
	if !node.HasChildren {
		out.Mode = fuse.S_IFREG | serv.fileMode
		out.Size = uint64(node.Size)
		out.Nlink = 1
	}
	out.Owner = fuse.Owner{Uid: serv.userID, Gid: serv.groupID}
	return fs.OK
}

func getMetaAttr(ctx context.Context, node *kdb.KdbNode, serv *fuseServer, out *fuse.AttrOut) syscall.Errno {
	time := node.Time
	out.SetTimes(&time, &time, &time)
	out.Mode = fuse.S_IFREG | serv.fileMode
	out.Nlink = 1

	nodeMetaDump, err := yaml.Marshal(node.Meta())
	if err != nil {
		serv.errors <- server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("unable to dump meta for %v", node),
		}
	}
	out.Size = uint64(len([]byte(nodeMetaDump)))

	out.Owner = fuse.Owner{Uid: serv.userID, Gid: serv.groupID}
	return fs.OK
}

func getattr(ctx context.Context, node *kdb.KdbNode, isMeta bool, serv *fuseServer, out *fuse.AttrOut) syscall.Errno {
	//TODO: support touch -m
	if !isMeta {
		return getNodeAttr(ctx, node, serv, out)
	}
	return getMetaAttr(ctx, node, serv, out)
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

func newHandle(node *kdb.KdbNode, meta bool, serv *fuseServer, flags uint32) fs.FileHandle {
	if !meta {
		return newFuseHandle(node, serv, flags)
	}
	return newMetaHandle(node, serv, flags)
}
