package tests

import (
	"testing"
	"path/filepath"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}
var _ = Suite(&MySuite{})

func (s *MySuite) TestFilePathMatch(c *C) {
	matched, _ := filepath.Match("test.*", "test.txt")
	c.Assert(matched, Equals, true)
}