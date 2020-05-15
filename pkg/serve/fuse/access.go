package fuse

import (
	"context"
	"fmt"
	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/hanwen/go-fuse/v2/fs"
	"path/filepath"
	"sync"
	"syscall"
)

func (server *FuseServer) OnAdd(ctx context.Context) {
	var populate func(kdb.TreeNode, *fs.Inode, *sync.WaitGroup)
	populate = func(node kdb.TreeNode, inode *fs.Inode, wg *sync.WaitGroup) {
		defer wg.Done()

		for _, child := range node.Children() {
			var childInode *fs.Inode
			metaData := child.MetaData()
			fmt.Println("Mounting ", metaData.Path)
			_, name := filepath.Split(metaData.Path)
			if metaData.IsLeaf {
				fmt.Println("... as leaf")
				embedder := &fs.MemRegularFile{Data: []byte("test")}
				childInode = inode.NewPersistentInode(ctx, embedder, fs.StableAttr{})
			} else {
				fmt.Println("... as dir")
				childInode = inode.NewPersistentInode(ctx, &fs.Inode{}, fs.StableAttr{Mode: syscall.S_IFDIR})
				wg.Add(1)
				go populate(child, childInode, wg)
			}
			inode.AddChild(name, childInode, true)
		}

	}

	var wg sync.WaitGroup
	wg.Add(1)
	go populate(server.dataRoot, &server.Inode, &wg)
	wg.Wait()
}
