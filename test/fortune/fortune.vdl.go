// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was auto-generated by the vanadium vdl tool.
// Source: fortune.vdl

package fortune

import (
	// VDL system imports
	"io"
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/i18n"
	"v.io/v23/rpc"
	"v.io/v23/vdl"
	"v.io/v23/verror"

	// VDL user imports
	"v.io/v23/security/access"
)

type ComplexErrorParam struct {
	Str  string
	Num  int32
	List []uint32
}

func (ComplexErrorParam) __VDLReflect(struct {
	Name string "v.io/x/jni/test/fortune.ComplexErrorParam"
}) {
}

func init() {
	vdl.Register((*ComplexErrorParam)(nil))
}

var (
	ErrErrNoFortunes = verror.Register("v.io/x/jni/test/fortune.ErrNoFortunes", verror.NoRetry, "{1:}{2:} no fortunes added")
	ErrErrComplex    = verror.Register("v.io/x/jni/test/fortune.ErrComplex", verror.NoRetry, "{1:}{2:} this is a complex error with params {3} {4} {5}")
)

func init() {
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(ErrErrNoFortunes.ID), "{1:}{2:} no fortunes added")
	i18n.Cat().SetWithBase(i18n.LangID("en"), i18n.MsgID(ErrErrComplex.ID), "{1:}{2:} this is a complex error with params {3} {4} {5}")
}

// NewErrErrNoFortunes returns an error with the ErrErrNoFortunes ID.
func NewErrErrNoFortunes(ctx *context.T) error {
	return verror.New(ErrErrNoFortunes, ctx)
}

// NewErrErrComplex returns an error with the ErrErrComplex ID.
func NewErrErrComplex(ctx *context.T, first ComplexErrorParam, second string, third int32) error {
	return verror.New(ErrErrComplex, ctx, first, second, third)
}

// FortuneClientMethods is the client interface
// containing Fortune methods.
//
// Fortune allows clients to Get and Add fortune strings.
type FortuneClientMethods interface {
	// Add stores a fortune in the set used by Get.
	Add(ctx *context.T, Fortune string, opts ...rpc.CallOpt) error
	// Get returns a random fortune.
	Get(*context.T, ...rpc.CallOpt) (Fortune string, err error)
	// StreamingGet returns a stream that can be used to obtain fortunes.
	StreamingGet(*context.T, ...rpc.CallOpt) (FortuneStreamingGetClientCall, error)
	// GetComplexError returns (always!) ErrComplex.
	GetComplexError(*context.T, ...rpc.CallOpt) error
	// NoTags is a method without tags.
	NoTags(*context.T, ...rpc.CallOpt) error
	// TestContext is a method used for testing that the server receives a
	// correct context.
	TestContext(*context.T, ...rpc.CallOpt) error
}

// FortuneClientStub adds universal methods to FortuneClientMethods.
type FortuneClientStub interface {
	FortuneClientMethods
	rpc.UniversalServiceMethods
}

// FortuneClient returns a client stub for Fortune.
func FortuneClient(name string) FortuneClientStub {
	return implFortuneClientStub{name}
}

type implFortuneClientStub struct {
	name string
}

func (c implFortuneClientStub) Add(ctx *context.T, i0 string, opts ...rpc.CallOpt) (err error) {
	var call rpc.ClientCall
	if call, err = v23.GetClient(ctx).StartCall(ctx, c.name, "Add", []interface{}{i0}, opts...); err != nil {
		return
	}
	err = call.Finish()
	return
}

func (c implFortuneClientStub) Get(ctx *context.T, opts ...rpc.CallOpt) (o0 string, err error) {
	var call rpc.ClientCall
	if call, err = v23.GetClient(ctx).StartCall(ctx, c.name, "Get", nil, opts...); err != nil {
		return
	}
	err = call.Finish(&o0)
	return
}

func (c implFortuneClientStub) StreamingGet(ctx *context.T, opts ...rpc.CallOpt) (ocall FortuneStreamingGetClientCall, err error) {
	var call rpc.ClientCall
	if call, err = v23.GetClient(ctx).StartCall(ctx, c.name, "StreamingGet", nil, opts...); err != nil {
		return
	}
	ocall = &implFortuneStreamingGetClientCall{ClientCall: call}
	return
}

func (c implFortuneClientStub) GetComplexError(ctx *context.T, opts ...rpc.CallOpt) (err error) {
	var call rpc.ClientCall
	if call, err = v23.GetClient(ctx).StartCall(ctx, c.name, "GetComplexError", nil, opts...); err != nil {
		return
	}
	err = call.Finish()
	return
}

func (c implFortuneClientStub) NoTags(ctx *context.T, opts ...rpc.CallOpt) (err error) {
	var call rpc.ClientCall
	if call, err = v23.GetClient(ctx).StartCall(ctx, c.name, "NoTags", nil, opts...); err != nil {
		return
	}
	err = call.Finish()
	return
}

func (c implFortuneClientStub) TestContext(ctx *context.T, opts ...rpc.CallOpt) (err error) {
	var call rpc.ClientCall
	if call, err = v23.GetClient(ctx).StartCall(ctx, c.name, "TestContext", nil, opts...); err != nil {
		return
	}
	err = call.Finish()
	return
}

// FortuneStreamingGetClientStream is the client stream for Fortune.StreamingGet.
type FortuneStreamingGetClientStream interface {
	// RecvStream returns the receiver side of the Fortune.StreamingGet client stream.
	RecvStream() interface {
		// Advance stages an item so that it may be retrieved via Value.  Returns
		// true iff there is an item to retrieve.  Advance must be called before
		// Value is called.  May block if an item is not available.
		Advance() bool
		// Value returns the item that was staged by Advance.  May panic if Advance
		// returned false or was not called.  Never blocks.
		Value() string
		// Err returns any error encountered by Advance.  Never blocks.
		Err() error
	}
	// SendStream returns the send side of the Fortune.StreamingGet client stream.
	SendStream() interface {
		// Send places the item onto the output stream.  Returns errors
		// encountered while sending, or if Send is called after Close or
		// the stream has been canceled.  Blocks if there is no buffer
		// space; will unblock when buffer space is available or after
		// the stream has been canceled.
		Send(item bool) error
		// Close indicates to the server that no more items will be sent;
		// server Recv calls will receive io.EOF after all sent items.
		// This is an optional call - e.g. a client might call Close if it
		// needs to continue receiving items from the server after it's
		// done sending.  Returns errors encountered while closing, or if
		// Close is called after the stream has been canceled.  Like Send,
		// blocks if there is no buffer space available.
		Close() error
	}
}

// FortuneStreamingGetClientCall represents the call returned from Fortune.StreamingGet.
type FortuneStreamingGetClientCall interface {
	FortuneStreamingGetClientStream
	// Finish performs the equivalent of SendStream().Close, then blocks until
	// the server is done, and returns the positional return values for the call.
	//
	// Finish returns immediately if the call has been canceled; depending on the
	// timing the output could either be an error signaling cancelation, or the
	// valid positional return values from the server.
	//
	// Calling Finish is mandatory for releasing stream resources, unless the call
	// has been canceled or any of the other methods return an error.  Finish should
	// be called at most once.
	Finish() (total int32, err error)
}

type implFortuneStreamingGetClientCall struct {
	rpc.ClientCall
	valRecv string
	errRecv error
}

func (c *implFortuneStreamingGetClientCall) RecvStream() interface {
	Advance() bool
	Value() string
	Err() error
} {
	return implFortuneStreamingGetClientCallRecv{c}
}

type implFortuneStreamingGetClientCallRecv struct {
	c *implFortuneStreamingGetClientCall
}

func (c implFortuneStreamingGetClientCallRecv) Advance() bool {
	c.c.errRecv = c.c.Recv(&c.c.valRecv)
	return c.c.errRecv == nil
}
func (c implFortuneStreamingGetClientCallRecv) Value() string {
	return c.c.valRecv
}
func (c implFortuneStreamingGetClientCallRecv) Err() error {
	if c.c.errRecv == io.EOF {
		return nil
	}
	return c.c.errRecv
}
func (c *implFortuneStreamingGetClientCall) SendStream() interface {
	Send(item bool) error
	Close() error
} {
	return implFortuneStreamingGetClientCallSend{c}
}

type implFortuneStreamingGetClientCallSend struct {
	c *implFortuneStreamingGetClientCall
}

func (c implFortuneStreamingGetClientCallSend) Send(item bool) error {
	return c.c.Send(item)
}
func (c implFortuneStreamingGetClientCallSend) Close() error {
	return c.c.CloseSend()
}
func (c *implFortuneStreamingGetClientCall) Finish() (o0 int32, err error) {
	err = c.ClientCall.Finish(&o0)
	return
}

// FortuneServerMethods is the interface a server writer
// implements for Fortune.
//
// Fortune allows clients to Get and Add fortune strings.
type FortuneServerMethods interface {
	// Add stores a fortune in the set used by Get.
	Add(call rpc.ServerCall, Fortune string) error
	// Get returns a random fortune.
	Get(rpc.ServerCall) (Fortune string, err error)
	// StreamingGet returns a stream that can be used to obtain fortunes.
	StreamingGet(FortuneStreamingGetServerCall) (total int32, err error)
	// GetComplexError returns (always!) ErrComplex.
	GetComplexError(rpc.ServerCall) error
	// NoTags is a method without tags.
	NoTags(rpc.ServerCall) error
	// TestContext is a method used for testing that the server receives a
	// correct context.
	TestContext(rpc.ServerCall) error
}

// FortuneServerStubMethods is the server interface containing
// Fortune methods, as expected by rpc.Server.
// The only difference between this interface and FortuneServerMethods
// is the streaming methods.
type FortuneServerStubMethods interface {
	// Add stores a fortune in the set used by Get.
	Add(call rpc.ServerCall, Fortune string) error
	// Get returns a random fortune.
	Get(rpc.ServerCall) (Fortune string, err error)
	// StreamingGet returns a stream that can be used to obtain fortunes.
	StreamingGet(*FortuneStreamingGetServerCallStub) (total int32, err error)
	// GetComplexError returns (always!) ErrComplex.
	GetComplexError(rpc.ServerCall) error
	// NoTags is a method without tags.
	NoTags(rpc.ServerCall) error
	// TestContext is a method used for testing that the server receives a
	// correct context.
	TestContext(rpc.ServerCall) error
}

// FortuneServerStub adds universal methods to FortuneServerStubMethods.
type FortuneServerStub interface {
	FortuneServerStubMethods
	// Describe the Fortune interfaces.
	Describe__() []rpc.InterfaceDesc
}

// FortuneServer returns a server stub for Fortune.
// It converts an implementation of FortuneServerMethods into
// an object that may be used by rpc.Server.
func FortuneServer(impl FortuneServerMethods) FortuneServerStub {
	stub := implFortuneServerStub{
		impl: impl,
	}
	// Initialize GlobState; always check the stub itself first, to handle the
	// case where the user has the Glob method defined in their VDL source.
	if gs := rpc.NewGlobState(stub); gs != nil {
		stub.gs = gs
	} else if gs := rpc.NewGlobState(impl); gs != nil {
		stub.gs = gs
	}
	return stub
}

type implFortuneServerStub struct {
	impl FortuneServerMethods
	gs   *rpc.GlobState
}

func (s implFortuneServerStub) Add(call rpc.ServerCall, i0 string) error {
	return s.impl.Add(call, i0)
}

func (s implFortuneServerStub) Get(call rpc.ServerCall) (string, error) {
	return s.impl.Get(call)
}

func (s implFortuneServerStub) StreamingGet(call *FortuneStreamingGetServerCallStub) (int32, error) {
	return s.impl.StreamingGet(call)
}

func (s implFortuneServerStub) GetComplexError(call rpc.ServerCall) error {
	return s.impl.GetComplexError(call)
}

func (s implFortuneServerStub) NoTags(call rpc.ServerCall) error {
	return s.impl.NoTags(call)
}

func (s implFortuneServerStub) TestContext(call rpc.ServerCall) error {
	return s.impl.TestContext(call)
}

func (s implFortuneServerStub) Globber() *rpc.GlobState {
	return s.gs
}

func (s implFortuneServerStub) Describe__() []rpc.InterfaceDesc {
	return []rpc.InterfaceDesc{FortuneDesc}
}

// FortuneDesc describes the Fortune interface.
var FortuneDesc rpc.InterfaceDesc = descFortune

// descFortune hides the desc to keep godoc clean.
var descFortune = rpc.InterfaceDesc{
	Name:    "Fortune",
	PkgPath: "v.io/x/jni/test/fortune",
	Doc:     "// Fortune allows clients to Get and Add fortune strings.",
	Methods: []rpc.MethodDesc{
		{
			Name: "Add",
			Doc:  "// Add stores a fortune in the set used by Get.",
			InArgs: []rpc.ArgDesc{
				{"Fortune", ``}, // string
			},
			Tags: []*vdl.Value{vdl.ValueOf(access.Tag("Write"))},
		},
		{
			Name: "Get",
			Doc:  "// Get returns a random fortune.",
			OutArgs: []rpc.ArgDesc{
				{"Fortune", ``}, // string
			},
			Tags: []*vdl.Value{vdl.ValueOf(access.Tag("Read"))},
		},
		{
			Name: "StreamingGet",
			Doc:  "// StreamingGet returns a stream that can be used to obtain fortunes.",
			OutArgs: []rpc.ArgDesc{
				{"total", ``}, // int32
			},
			Tags: []*vdl.Value{vdl.ValueOf(access.Tag("Read"))},
		},
		{
			Name: "GetComplexError",
			Doc:  "// GetComplexError returns (always!) ErrComplex.",
			Tags: []*vdl.Value{vdl.ValueOf(access.Tag("Read"))},
		},
		{
			Name: "NoTags",
			Doc:  "// NoTags is a method without tags.",
		},
		{
			Name: "TestContext",
			Doc:  "// TestContext is a method used for testing that the server receives a\n// correct context.",
			Tags: []*vdl.Value{vdl.ValueOf(access.Tag("Read"))},
		},
	},
}

// FortuneStreamingGetServerStream is the server stream for Fortune.StreamingGet.
type FortuneStreamingGetServerStream interface {
	// RecvStream returns the receiver side of the Fortune.StreamingGet server stream.
	RecvStream() interface {
		// Advance stages an item so that it may be retrieved via Value.  Returns
		// true iff there is an item to retrieve.  Advance must be called before
		// Value is called.  May block if an item is not available.
		Advance() bool
		// Value returns the item that was staged by Advance.  May panic if Advance
		// returned false or was not called.  Never blocks.
		Value() bool
		// Err returns any error encountered by Advance.  Never blocks.
		Err() error
	}
	// SendStream returns the send side of the Fortune.StreamingGet server stream.
	SendStream() interface {
		// Send places the item onto the output stream.  Returns errors encountered
		// while sending.  Blocks if there is no buffer space; will unblock when
		// buffer space is available.
		Send(item string) error
	}
}

// FortuneStreamingGetServerCall represents the context passed to Fortune.StreamingGet.
type FortuneStreamingGetServerCall interface {
	rpc.ServerCall
	FortuneStreamingGetServerStream
}

// FortuneStreamingGetServerCallStub is a wrapper that converts rpc.StreamServerCall into
// a typesafe stub that implements FortuneStreamingGetServerCall.
type FortuneStreamingGetServerCallStub struct {
	rpc.StreamServerCall
	valRecv bool
	errRecv error
}

// Init initializes FortuneStreamingGetServerCallStub from rpc.StreamServerCall.
func (s *FortuneStreamingGetServerCallStub) Init(call rpc.StreamServerCall) {
	s.StreamServerCall = call
}

// RecvStream returns the receiver side of the Fortune.StreamingGet server stream.
func (s *FortuneStreamingGetServerCallStub) RecvStream() interface {
	Advance() bool
	Value() bool
	Err() error
} {
	return implFortuneStreamingGetServerCallRecv{s}
}

type implFortuneStreamingGetServerCallRecv struct {
	s *FortuneStreamingGetServerCallStub
}

func (s implFortuneStreamingGetServerCallRecv) Advance() bool {
	s.s.errRecv = s.s.Recv(&s.s.valRecv)
	return s.s.errRecv == nil
}
func (s implFortuneStreamingGetServerCallRecv) Value() bool {
	return s.s.valRecv
}
func (s implFortuneStreamingGetServerCallRecv) Err() error {
	if s.s.errRecv == io.EOF {
		return nil
	}
	return s.s.errRecv
}

// SendStream returns the send side of the Fortune.StreamingGet server stream.
func (s *FortuneStreamingGetServerCallStub) SendStream() interface {
	Send(item string) error
} {
	return implFortuneStreamingGetServerCallSend{s}
}

type implFortuneStreamingGetServerCallSend struct {
	s *FortuneStreamingGetServerCallStub
}

func (s implFortuneStreamingGetServerCallSend) Send(item string) error {
	return s.s.Send(item)
}
