package fuse

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/server"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type fuseServer struct {
	mountPoint        string
	mountpointCreated bool
	userID, groupID   uint32
	fileMode, dirMode uint32

	tree     *kdb.KdbTree
	fuseRoot *fuseNode

	errors chan error
}

type fuseNode struct {
	fs.Inode
	server  *fuseServer
	kdbNode *kdb.KdbNode
}

const (
	defaultFileMode = 0660
	defaultDirMode  = 0751
)

func New(mountPoint string, options map[string][]string) (*fuseServer, error) {
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

	serv := &fuseServer{
		mountPoint:        mountPoint,
		mountpointCreated: mountpointCreated,
		errors:            make(chan error),
		fileMode:          defaultFileMode,
		dirMode:           defaultDirMode,
	}
	err = serv.configure(options)
	if err != nil {
		return nil, server.ServerError{
			Inner:   err,
			Message: fmt.Sprintf("failed to configure fuse server %s", mountPoint),
		}
	}
	return serv, nil
}

func (fserver *fuseServer) configure(options map[string][]string) error {
	var err error
	u, _ := user.Current()
	if usernames, ok := options["user"]; ok && len(usernames) > 0 {
		u, err = user.Lookup(usernames[0])
		if err != nil {
			return server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("user not found %s", usernames[0]),
			}
		}
	}

	var g *user.Group
	if groupnames, ok := options["group"]; ok && len(groupnames) > 0 {
		g, err = user.LookupGroup(groupnames[0])
		if err != nil {
			return server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("group not found %s", groupnames[0]),
			}
		}
	}

	if uids, ok := options["uid"]; ok && len(uids) > 0 {
		uid, err := strconv.Atoi(uids[0])
		if err != nil {
			return server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("invalid uid %s", uids[0]),
			}
		}
		fserver.userID = uint32(uid)
	} else if u != nil {
		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			return server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("invalid uid %s", u.Uid),
			}
		}
		fserver.userID = uint32(uid)
	}

	if gids, ok := options["gid"]; ok && len(gids) > 0 {
		gid, err := strconv.Atoi(gids[0])
		if err != nil {
			return server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("invalid gid %s", gids[0]),
			}
		}
		fserver.groupID = uint32(gid)
	} else if g != nil {
		gid, err := strconv.Atoi(g.Gid)
		if err != nil {
			return server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("invalid user gid %s", g.Gid),
			}
		}
		fserver.groupID = uint32(gid)
	} else if u != nil {
		gid, err := strconv.Atoi(u.Gid)
		if err != nil {
			return server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("invalid group gid %s", g.Gid),
			}
		}
		fserver.groupID = uint32(gid)
	}

	if fileModes, ok := options["file"]; ok && len(fileModes) > 0 {
		fileMode, err := strconv.ParseInt(fileModes[0], 8, 0)
		if err != nil {
			return server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("invalid file mode %s", fileModes[0]),
			}
		}
		fserver.fileMode = uint32(fileMode)
	}

	if dirModes, ok := options["dir"]; ok && len(dirModes) > 0 {
		dirMode, err := strconv.ParseInt(dirModes[0], 8, 0)
		if err != nil {
			return server.ServerError{
				Inner:   err,
				Message: fmt.Sprintf("invalid dir mode %s", dirModes[0]),
			}
		}
		fserver.dirMode = uint32(dirMode)
	}

	fmt.Printf("%#o %#o\n", fserver.fileMode, fserver.dirMode)
	return nil
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

	fuse, err := fs.Mount(fserver.mountPoint, fserver.fuseRoot, &fs.Options{
		UID: fserver.userID,
		GID: fserver.groupID,
		MountOptions: fuse.MountOptions{
			Name:   "pkdb",
			FsName: "pkdb",
			Debug:  false,
		},
	})
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
