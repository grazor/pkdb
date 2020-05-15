package serve

import (
	"fmt"
	"net/url"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/serve/fuse"
)

type Server interface {
	Serve(kdb.TreeNode, chan interface{}) error
}

func New(uri string) (Server, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "", "file":
		{
			return fuse.New(u.Hostname() + u.Path)
		}
	}

	return nil, fmt.Errorf("unexpected scheme %s", u.Scheme)
}
