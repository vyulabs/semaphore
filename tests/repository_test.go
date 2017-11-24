package tests

import (
	"testing"
	"path/filepath"
	. "gopkg.in/check.v1"
	"github.com/ansible-semaphore/semaphore/api/tasks"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}
var _ = Suite(&MySuite{})

func (s *MySuite) TestFilePathMatch(c *C) {
	matched, _ := filepath.Match("test.*", "test.txt")
	c.Assert(matched, Equals, true)
}

func (s *MySuite) TestResolveNewVersionEmpty(c *C) {
	version, err := tasks.ResolveNewVersion("", "1.4.<next_index>", 12, 32)
	c.Assert(err, Equals, nil)
	c.Assert(version, Equals, "1.4.0")
}

func (s *MySuite) TestResolveNewVersion(c *C) {
	version, err := tasks.ResolveNewVersion("1.2.3", "1.<task_id>.<task_num>", 12, 32)
	c.Assert(err, Equals, nil)
	c.Assert(version, Equals, "1.12.32")
}

func (s *MySuite) TestResolveNewVersionWithNextIndexAtTheEnd(c *C) {
	version, err := tasks.ResolveNewVersion("1.30.20", "1.30.<next_index>", 12, 32)
	c.Assert(err, Equals, nil)
	c.Assert(version, Equals, "1.30.21")
}

func (s *MySuite) TestResolveNewVersionWithNextIndex(c *C) {
	version, err := tasks.ResolveNewVersion("1.10.44", "1.<next_index>.44", 12, 32)
	c.Assert(err, Equals, nil)
	c.Assert(version, Equals, "1.11.44")
}

func (s *MySuite) TestResolveNewVersionAll(c *C) {
	version, err := tasks.ResolveNewVersion("1.30.20#10", "1.30.<next_index>#<task_id>", 12, 32)
	c.Assert(err, Equals, nil)
	c.Assert(version, Equals, "1.30.21#12")
}