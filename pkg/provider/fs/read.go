
import (
	"path/filepath"

	"github.com/grazor/pkdb/pkg/provider"
)

func (entry fsEntry) absolutePath() string {
	return filepath.Join(entry.provider.basePath, entry.relativePath)
}

func (entry fsEntry) Id() string {
	//TODO: read id from attrs
	_, name := filepath.Split(entry.relativePath)
	return name
}

func (entry fsEntry) Name() string {
	_, name := filepath.Split(entry.relativePath)
	return name
}

func (entry fsEntry) Path() string {
	return entry.relativePath
}

func (entry fsEntry) Attrs() map[string]interface{} {
	//TODO: read and cache attrs
	return nil
}

func (entry fsEntry) Time() time.Time {
	return entry.fileInfo.ModTime()
}

func (entry fsEntry) HasChildren() bool {
	return entry.fileInfo.IsDir()
}

func (entry fsEntry) Children() ([]provider.Entry, error) {
	if !entry.HasChildren() {
		return nil, nil
	}

	dir, err := os.Open(entry.absolutePath())
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	dirContents, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	children := make([]provider.Entry, 0, len(dirContents))
	for _, childInfo := range dirContents {
		child := fsEntry{
			provider: entry.provider,
			relativePath: filepath.Join(entry.relativePath, childInfo.Name()),
			fileInfo: childInfo,
		}
		children = append(children, chidl)
	}
	return children, nil
}
