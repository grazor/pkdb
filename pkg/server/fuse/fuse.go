package fuse

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/server"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type fuseServer struct {
	mountPoint        string
	mountpointCreated bool

	tree     *kdb.KdbTree
	fuseRoot *fuseNode

	errors chan error
}

type fuseNode struct {
	fs.Inode
	server  *fuseServer
	kdbNode *kdb.KdbNode
}

func New(mountPoint string) (*fuseServer, error) {
	mountpointCreated := false
	absPath, err := filepath.Abs(mountPoint)
	if err != nil {
		return nil, server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("failed to handle path %v", mountPoint),
		}
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		parentPath := filepath.Dir(absPath)
		if _, err := os.Stat(parentPath); os.IsNotExist(err) {
			return nil, server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("mountpoint base path %v does not exist", parentPath),
			}

		}
		err := os.Mkdir(absPath, 0751)
		if err != nil {
			return nil, server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("failed to create mountpoint %v", parentPath),
			}
		}
		mountpointCreated = true
	}

	server := &fuseServer{
		mountPoint:        mountPoint,
		mountpointCreated: mountpointCreated,
		errors:            make(chan error),
	}
	return server, nil
}

func (fserver *fuseServer) Errors() <-chan error {
	return fserver.errors
}

func (fserver *fuseServer) String() string {
	return fmt.Sprintf("fuse(%v)", fserver.mountPoint)
}

func (fserver *fuseServer) Serve(ctx context.Context, wg *sync.WaitGroup, tree *kdb.KdbTree) error {
	serve := func(c context.Context, w *sync.WaitGroup, fs *fuseServer, s *fuse.Server) {
		defer w.Done()
		<-c.Done()
		s.Unmount()
		if fs.mountpointCreated {
			os.Remove(fs.mountPoint)
		}
	}

	fserver.tree = tree
	fserver.fuseRoot = &fuseNode{server: fserver, kdbNode: tree.Root}

	fuse, err := fs.Mount(fserver.mountPoint, fserver.fuseRoot, &fs.Options{})
	if err != nil {
		return server.ServerError{
			Inner:   err,
			Message: "failed to mount fuse",
		}
	}

	wg.Add(1)
	go serve(ctx, wg, fserver, fuse)
	return nil
}
