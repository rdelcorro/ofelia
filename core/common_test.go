package core

import (
	"errors"
	"fmt"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type SuiteCommon struct{}

var _ = Suite(&SuiteCommon{})

func (s *SuiteCommon) TestNewContext(c *C) {
	h := NewScheduler(&TestLogger{})
	j := &TestJob{}
	j.Use(&TestMiddleware{})

	e := NewExecution()

	ctx := NewContext(h, j, e)
	c.Assert(ctx.Scheduler, DeepEquals, h)
	c.Assert(ctx.Job, DeepEquals, j)
	c.Assert(ctx.Execution, DeepEquals, e)
	c.Assert(ctx.middlewares, HasLen, 1)
}

func (s *SuiteCommon) TestContextNextError(c *C) {
	mA := &TestMiddlewareAltA{}
	mB := &TestMiddlewareAltB{}
	mC := &TestMiddlewareAltC{}
	mB.Error, mC.Error = fmt.Errorf("foo"), fmt.Errorf("foo")

	j := &TestJob{}
	j.Use(mA, mB, mC)

	e := NewExecution()

	h := NewScheduler(&TestLogger{})
	ctx := NewContext(h, j, e)
	ctx.Start()

	err := ctx.Next()
	c.Assert(err, IsNil)
	c.Assert(mA.Called, Equals, 1)
	c.Assert(mB.Called, Equals, 0)
	c.Assert(mC.Called, Equals, 0)
	c.Assert(j.Called, Equals, 0)
	c.Assert(ctx.Execution.IsRunning, Equals, true)

	err = ctx.Next()
	c.Assert(err, IsNil)
	c.Assert(mB.Called, Equals, 1)
	c.Assert(mC.Called, Equals, 0)
	c.Assert(j.Called, Equals, 0)
	c.Assert(ctx.Execution.IsRunning, Equals, false)

	err = ctx.Next()
	c.Assert(err, IsNil)
	c.Assert(mC.Called, Equals, 0)
	c.Assert(j.Called, Equals, 0)

	err = ctx.Next()
	c.Assert(err, IsNil)
	c.Assert(j.Called, Equals, 0)
}

func (s *SuiteCommon) TestContextNextNested(c *C) {
	mA := &TestMiddlewareAltA{}
	mB := &TestMiddlewareAltB{}
	mC := &TestMiddlewareAltC{}
	mA.Nested, mB.Nested, mC.Nested = true, true, true

	j := &TestJob{}
	j.Use(mA, mB, mC)

	e := NewExecution()

	h := NewScheduler(&TestLogger{})
	ctx := NewContext(h, j, e)
	ctx.Start()

	err := ctx.Next()
	c.Assert(err, IsNil)
}

func (s *SuiteCommon) TestContextNext(c *C) {
	mA := &TestMiddlewareAltA{}
	mB := &TestMiddlewareAltB{}
	mC := &TestMiddlewareAltC{}

	j := &TestJob{}
	j.Use(mA, mB, mC)

	e := NewExecution()

	h := NewScheduler(&TestLogger{})
	ctx := NewContext(h, j, e)
	ctx.Start()

	err := ctx.Next()
	c.Assert(err, IsNil)
	c.Assert(mA.Called, Equals, 1)
	c.Assert(mB.Called, Equals, 0)
	c.Assert(mC.Called, Equals, 0)
	c.Assert(j.Called, Equals, 0)
	c.Assert(ctx.Execution.IsRunning, Equals, true)

	err = ctx.Next()
	c.Assert(err, IsNil)
	c.Assert(mB.Called, Equals, 1)
	c.Assert(mC.Called, Equals, 0)
	c.Assert(j.Called, Equals, 0)
	c.Assert(ctx.Execution.IsRunning, Equals, true)

	err = ctx.Next()
	c.Assert(err, IsNil)
	c.Assert(mC.Called, Equals, 1)
	c.Assert(j.Called, Equals, 0)

	err = ctx.Next()
	c.Assert(err, IsNil)
	c.Assert(j.Called, Equals, 1)

	err = ctx.Next()
	c.Assert(err, IsNil)
	c.Assert(j.Called, Equals, 1)
}

func (s *SuiteCommon) TestExecutionStart(c *C) {
	exe := &Execution{}
	exe.Start()

	c.Assert(exe.IsRunning, Equals, true)
	c.Assert(exe.Date.IsZero(), Equals, false)
}

func (s *SuiteCommon) TestExecutionStop(c *C) {
	exe := &Execution{}
	exe.Start()
	exe.Stop(nil)

	c.Assert(exe.IsRunning, Equals, false)
	c.Assert(exe.Failed, Equals, false)
	c.Assert(exe.Skipped, Equals, false)
	c.Assert(exe.Error, Equals, nil)
	c.Assert(exe.Duration.Seconds() > .0, Equals, true)
}

func (s *SuiteCommon) TestExecutionStopError(c *C) {
	err := errors.New("foo")

	exe := &Execution{}
	exe.Start()
	exe.Stop(err)

	c.Assert(exe.IsRunning, Equals, false)
	c.Assert(exe.Failed, Equals, true)
	c.Assert(exe.Skipped, Equals, false)
	c.Assert(exe.Error, Equals, err)
	c.Assert(exe.Duration.Seconds() > .0, Equals, true)
}

func (s *SuiteCommon) TestMiddlewareContainerUseTwice(c *C) {
	mA := &TestMiddleware{}
	mB := &TestMiddleware{}

	container := &middlewareContainer{}
	container.Use(mA)
	container.Use(mB)

	ms := container.Middlewares()
	c.Assert(ms, HasLen, 1)
	c.Assert(ms[0], Equals, mA)
}

func (s *SuiteCommon) TestMiddlewareContainerUseNil(c *C) {
	var m Middleware

	container := &middlewareContainer{}
	container.Use(m)

	ms := container.Middlewares()
	c.Assert(ms, HasLen, 0)
}

func (s *SuiteCommon) TestMiddlewareContainerUseOder(c *C) {
	mA := &TestMiddleware{}
	mB := &TestMiddlewareAltA{}

	container := &middlewareContainer{}
	container.Use(mB)
	container.Use(mA)

	ms := container.Middlewares()
	c.Assert(ms, HasLen, 2)
	c.Assert(ms[0], Equals, mB)
	c.Assert(ms[1], Equals, mA)
}

type TestMiddleware struct {
	Called int
	Nested bool
	Error  error
}

func (m *TestMiddleware) Run(ctx *Context) error {
	m.Called++

	if m.Nested {
		ctx.Next()
	}

	return m.Error
}

type TestMiddlewareAltA struct{ TestMiddleware }
type TestMiddlewareAltB struct{ TestMiddleware }
type TestMiddlewareAltC struct{ TestMiddleware }

type TestJob struct {
	BareJob
	Called int
}

func (j *TestJob) Run(ctx *Context) error {
	j.Called++
	time.Sleep(time.Millisecond * 500)

	return nil
}

type TestLogger struct{}

func (*TestLogger) Critical(format string, args ...interface{}) {}
func (*TestLogger) Debug(format string, args ...interface{})    {}
func (*TestLogger) Error(format string, args ...interface{})    {}
func (*TestLogger) Notice(format string, args ...interface{})   {}
func (*TestLogger) Warning(format string, args ...interface{})  {}
