package filesystem

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"path/filepath"
)

func (node *FsNode) Watch() (chan interface{}, error) {
	watch := func(watcher *fsnotify.Watcher, done chan interface{}, root *FsNode) {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Chmod != fsnotify.Chmod {
					// TODO: handle deletion
					nodePath := filepath.Dir(event.Name)
					nodeModified, ok := root.index[nodePath]
					fmt.Println(event)
					if ok {
						nodeModified.Load(1)
					}
				}
			case <-done:
				return
			}
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	done := make(chan interface{})
	go watch(watcher, done, node)

	children := []*FsNode{node}
	for len(children) > 0 {
		n := children[0]
		children = children[1:]
		if n.children != nil && len(n.children) > 0 {
			watcher.Add(n.path)
			children = append(children, n.children...)
		}
	}

	return done, nil
}
