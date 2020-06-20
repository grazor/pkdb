package fuse

import (
	"context"
	"syscall"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func setattr(ctx context.Context, node *kdb.KdbNode, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	//TODO: support touch -m
	time := node.Time
	out.Mode = fuse.S_IFDIR | dirMode
	out.SetTimes(&time, &time, &time)
	if !node.HasChildren {
		out.Mode = fuse.S_IFREG | fileMode
		out.Size = uint64(node.Size)
		out.Nlink = 1
	}
	out.Owner = fuse.Owner{Uid: 1000, Gid: 100}
	return fs.OK

}
