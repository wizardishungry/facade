package generator

import (
	"errors"
	"go/build"
	"os"
	"testing"
)

func TestAllPackages(t *testing.T) {
	t.Skipf("skipping suite")
	t.Parallel() // This is slow
	runPackageCases(t, AllPackages...)
}

func TestSomePackages(t *testing.T) {
	runPackageCases(t,
		"net/http",
		"os",
		"crypto/md5", // has a reference to hash.Hash
	)
}

func runPackageCases(t *testing.T, packages ...string) {
	for _, tC := range packages {
		t.Run(tC, func(t *testing.T) {
			pkg, err := New(tC, "foobar", false)
			if err != nil {
				if !errors.Is(err, &build.NoGoError{}) {
					return
				}
				t.Fatalf("New: %v", err)
			}

			err = pkg.Write(os.Stderr)
			if err != nil {
				t.Fatalf("Write: %v", err)
			}
		})
	}
}
