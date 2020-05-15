package fuse

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"os"
	"path/filepath"

	"github.com/grazor/pkdb/pkg/kdb"
)

type FuseServer struct {
	fs.Inode
	dataRoot kdb.TreeNode
	target   string
}

func New(target string) (*FuseServer, error) {
	absPath, err := filepath.Abs(target)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, err
	}

	return &FuseServer{target: target}, nil
}

func (server *FuseServer) Serve(root kdb.TreeNode, done chan interface{}) error {
	server.dataRoot = root

	fuse, err := fs.Mount(server.target, server, &fs.Options{MountOptions: fuse.MountOptions{Debug: true}})
	if err != nil {
		return err
	}
	fuse.Wait()

	return nil
}
