package relpath

import (
	"go/build"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

// Find a template or data file
func Find(partial string) string {
	klog.V(2).Infof("Finding path to %s ...", partial)
	binpath, err := os.Executable()
	if err != nil {
		binpath = "."
	}

	try := []string{
		partial,
		filepath.Join("..", "..", partial),
		filepath.Join("..", partial),
		filepath.Join("/app", partial),
		filepath.Join(filepath.Dir(binpath), partial),
		filepath.Join(build.Default.GOPATH, "github.com/tstromberg/campwiz", partial),
	}

	for _, path := range try {
		_, err := os.Stat(path)
		if err == nil {
			return path
		}
	}

	klog.Errorf("unable to find: %s", partial)
	return partial
}
