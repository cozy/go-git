package filesystem

import (
	"testing"

	"github.com/cozy/go-git/storage/test"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-billy.v2/memfs"
	"gopkg.in/src-d/go-billy.v2/osfs"
)

func Test(t *testing.T) { TestingT(t) }

type StorageSuite struct {
	test.BaseStorageSuite
}

var _ = Suite(&StorageSuite{})

func (s *StorageSuite) SetUpTest(c *C) {
	storage, err := NewStorage(osfs.New(c.MkDir()))
	c.Assert(err, IsNil)

	s.BaseStorageSuite = test.NewBaseStorageSuite(storage)
}

func (s *StorageSuite) TestNewStorage(c *C) {
	fs := memfs.New()
	storage, err := NewStorage(fs)
	c.Assert(err, IsNil)
	c.Assert(storage, NotNil)

	_, err = fs.Stat("refs/tags")
	c.Assert(err, IsNil)
}

func (s *StorageSuite) TestFilesystem(c *C) {
	fs := memfs.New()
	storage, err := NewStorage(fs)
	c.Assert(err, IsNil)

	c.Assert(storage.Filesystem(), Equals, fs)
}
