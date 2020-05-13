package fuse

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"os"
	"path/filepath"

	"github.com/grazor/pkdb/pkg/load"
)

type FuseServer struct {
	path string
}

func NewFuseServer(path string) (*FuseServer, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, err
	}

	return &FuseServer{path: path}, nil
}

func (server *FuseServer) Serve(root load.TreeNode) (chan interface{}, error) {
	fuse, err := fs.Mount(server.path, nil, &fs.Options{})
	if err != nil {
		return nil, err
	}
	fuse.Wait()

	done := make(chan interface{})
	return done, nil
}
