package load

import (
	"fmt"
	"net/url"

	"github.com/grazor/pkdb/pkg/kdb"
	"github.com/grazor/pkdb/pkg/load/filesystem"
)

func New(uri string) (kdb.TreeNode, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "", "file":
		{
			return filesystem.New(u.Hostname() + u.Path)
		}
	}

	return nil, fmt.Errorf("unexpected scheme %s", u.Scheme)
}
