// Package test implements common test suite for different transport
// implementations.
//
package test

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/cozy/go-git/plumbing"
	"github.com/cozy/go-git/plumbing/format/packfile"
	"github.com/cozy/go-git/plumbing/protocol/packp"
	"github.com/cozy/go-git/plumbing/transport"
	"github.com/cozy/go-git/storage/memory"

	"github.com/cozy/go-git/plumbing/protocol/packp/capability"
	. "gopkg.in/check.v1"
)

type UploadPackSuite struct {
	Endpoint            transport.Endpoint
	EmptyEndpoint       transport.Endpoint
	NonExistentEndpoint transport.Endpoint
	EmptyAuth           transport.AuthMethod
	Client              transport.Transport
}

func (s *UploadPackSuite) TestAdvertisedReferencesEmpty(c *C) {
	r, err := s.Client.NewUploadPackSession(s.EmptyEndpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	ar, err := r.AdvertisedReferences()
	c.Assert(err, Equals, transport.ErrEmptyRemoteRepository)
	c.Assert(ar, IsNil)
}

func (s *UploadPackSuite) TestAdvertisedReferencesNotExists(c *C) {
	r, err := s.Client.NewUploadPackSession(s.NonExistentEndpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	ar, err := r.AdvertisedReferences()
	c.Assert(err, Equals, transport.ErrRepositoryNotFound)
	c.Assert(ar, IsNil)

	r, err = s.Client.NewUploadPackSession(s.NonExistentEndpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	req := packp.NewUploadPackRequest()
	req.Wants = append(req.Wants, plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
	reader, err := r.UploadPack(req)
	c.Assert(err, Equals, transport.ErrRepositoryNotFound)
	c.Assert(reader, IsNil)
}

func (s *UploadPackSuite) TestCallAdvertisedReferenceTwice(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	ar1, err := r.AdvertisedReferences()
	c.Assert(err, IsNil)
	c.Assert(ar1, NotNil)
	ar2, err := r.AdvertisedReferences()
	c.Assert(err, IsNil)
	c.Assert(ar2, DeepEquals, ar1)
}

func (s *UploadPackSuite) TestDefaultBranch(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	defer func() { c.Assert(r.Close(), IsNil) }()

	info, err := r.AdvertisedReferences()
	c.Assert(err, IsNil)
	symrefs := info.Capabilities.Get(capability.SymRef)
	c.Assert(symrefs, HasLen, 1)
	c.Assert(symrefs[0], Equals, "HEAD:refs/heads/master")
}

func (s *UploadPackSuite) TestAdvertisedReferencesFilterUnsupported(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	defer func() { c.Assert(r.Close(), IsNil) }()

	info, err := r.AdvertisedReferences()
	c.Assert(err, IsNil)
	c.Assert(info.Capabilities.Supports(capability.MultiACK), Equals, false)
}

func (s *UploadPackSuite) TestCapabilities(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	defer func() { c.Assert(r.Close(), IsNil) }()

	info, err := r.AdvertisedReferences()
	c.Assert(err, IsNil)
	c.Assert(info.Capabilities.Get(capability.Agent), HasLen, 1)
}

func (s *UploadPackSuite) TestFullUploadPack(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	defer func() { c.Assert(r.Close(), IsNil) }()

	info, err := r.AdvertisedReferences()
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)

	req := packp.NewUploadPackRequest()
	req.Wants = append(req.Wants, plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))

	reader, err := r.UploadPack(req)
	c.Assert(err, IsNil)

	s.checkObjectNumber(c, reader, 28)
}

func (s *UploadPackSuite) TestUploadPack(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	defer func() { c.Assert(r.Close(), IsNil) }()

	req := packp.NewUploadPackRequest()
	req.Wants = append(req.Wants, plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))

	reader, err := r.UploadPack(req)
	c.Assert(err, IsNil)

	s.checkObjectNumber(c, reader, 28)
}

func (s *UploadPackSuite) TestUploadPackInvalidReq(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	defer func() { c.Assert(r.Close(), IsNil) }()

	req := packp.NewUploadPackRequest()
	req.Wants = append(req.Wants, plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
	req.Capabilities.Set(capability.Sideband)
	req.Capabilities.Set(capability.Sideband64k)

	_, err = r.UploadPack(req)
	c.Assert(err, NotNil)
}

func (s *UploadPackSuite) TestUploadPackNoChanges(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	defer func() { c.Assert(r.Close(), IsNil) }()

	req := packp.NewUploadPackRequest()
	req.Wants = append(req.Wants, plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
	req.Haves = append(req.Haves, plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))

	reader, err := r.UploadPack(req)
	c.Assert(err, Equals, transport.ErrEmptyUploadPackRequest)
	c.Assert(reader, IsNil)
}

func (s *UploadPackSuite) TestUploadPackMulti(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	defer func() { c.Assert(r.Close(), IsNil) }()

	req := packp.NewUploadPackRequest()
	req.Wants = append(req.Wants, plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
	req.Wants = append(req.Wants, plumbing.NewHash("e8d3ffab552895c19b9fcf7aa264d277cde33881"))

	reader, err := r.UploadPack(req)
	c.Assert(err, IsNil)

	s.checkObjectNumber(c, reader, 31)
}

func (s *UploadPackSuite) TestUploadPackPartial(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)
	defer func() { c.Assert(r.Close(), IsNil) }()

	req := packp.NewUploadPackRequest()
	req.Wants = append(req.Wants, plumbing.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
	req.Haves = append(req.Haves, plumbing.NewHash("918c48b83bd081e863dbe1b80f8998f058cd8294"))

	reader, err := r.UploadPack(req)
	c.Assert(err, IsNil)

	s.checkObjectNumber(c, reader, 4)
}

func (s *UploadPackSuite) TestFetchError(c *C) {
	r, err := s.Client.NewUploadPackSession(s.Endpoint, s.EmptyAuth)
	c.Assert(err, IsNil)

	req := packp.NewUploadPackRequest()
	req.Wants = append(req.Wants, plumbing.NewHash("1111111111111111111111111111111111111111"))

	reader, err := r.UploadPack(req)
	c.Assert(err, NotNil)
	c.Assert(reader, IsNil)

	//XXX: We do not test Close error, since implementations might return
	//     different errors if a previous error was found.
}

func (s *UploadPackSuite) checkObjectNumber(c *C, r io.Reader, n int) {
	b, err := ioutil.ReadAll(r)
	c.Assert(err, IsNil)
	buf := bytes.NewBuffer(b)
	scanner := packfile.NewScanner(buf)
	storage := memory.NewStorage()
	d, err := packfile.NewDecoder(scanner, storage)
	c.Assert(err, IsNil)
	_, err = d.Decode()
	c.Assert(err, IsNil)
	c.Assert(len(storage.Objects), Equals, n)
}
