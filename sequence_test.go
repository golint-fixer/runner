package runner

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSequence_Interfaces(t *testing.T) {
	seq := &sequence{}
	var _ Command = seq
	var _ Rollbacker = seq
	var _ DryRunner = seq
	var _ fmt.Stringer = seq
}

func TestSequence_String(t *testing.T) {
	is := assert.New(t)
	seq := NewSequence()
	is.Contains(fmt.Sprint(seq), "0")

	seq = NewSequence(&MockCommand{}, &MockCommand{})
	is.Contains(fmt.Sprint(seq), "2")
}

func TestSequence_Run_EmptySequence(t *testing.T) {
	is := assert.New(t)
	p, out := getTestPrinter()

	ctx := NewContext(0)
	seq := NewSequence()

	seq.Run(ctx, p)
	is.NoError(ctx.Err())
	is.Empty(out.String())
}

func TestSequence_Run_Success(t *testing.T) {
	is := assert.New(t)

	cmdA := &MockCommand{name: "A"}
	cmdB := &MockCommand{name: "B"}
	cmdC := &MockCommand{name: "C"}

	ctx := NewContext(3)
	seq := NewSequence(cmdA, cmdB, cmdC)
	seq.Run(ctx, DefaultPrinter)

	is.NoError(ctx.Err())
	for _, cmd := range []*MockCommand{cmdA, cmdB, cmdC} {
		is.True(cmd.ran)
		is.False(cmd.failed)
		is.False(cmd.rolledBack)
	}
}

func TestSequence_Run_Rollback(t *testing.T) {
	is := assert.New(t)
	err := errors.New("foobar")

	cmdA := &MockCommand{name: "A"}
	cmdB := &MockCommand{name: "B"}
	cmdC := &MockCommand{name: "C", err: err}
	cmdD := &MockCommand{name: "D"}

	ctx := NewContext(3)
	seq := NewSequence(cmdA, cmdB, cmdC, cmdD)
	seq.Run(ctx, DefaultPrinter)

	is.Equal(err, ctx.Err())
	is.True(cmdA.ran)
	is.True(cmdA.rolledBack)
	is.True(cmdB.ran)
	is.True(cmdB.rolledBack)
	is.True(cmdC.ran)
	is.True(cmdC.failed)
	is.False(cmdC.rolledBack)
	is.False(cmdD.ran)
}

func TestSequence_DryRun_EmptySequence(t *testing.T) {
	is := assert.New(t)
	p, out := getTestPrinter()

	ctx := NewContext(0)
	seq := NewSequence().(*sequence)

	seq.DryRun(ctx, p)
	is.NoError(ctx.Err())
	is.Empty(out.String())
}

func TestSequence_DryRun_Success(t *testing.T) {
	is := assert.New(t)

	cmdA := &MockCommand{name: "A"}
	cmdB := &MockCommand{name: "B"}
	cmdC := &MockCommand{name: "C"}

	ctx := NewContext(3)
	seq := NewSequence(cmdA, cmdB, cmdC).(*sequence)
	seq.DryRun(ctx, DefaultPrinter)
	is.NoError(ctx.Err())
}

func TestSequence_Rollback_EmptySequence(t *testing.T) {
	is := assert.New(t)
	p, out := getTestPrinter()

	ctx := NewContext(0)
	seq := NewSequence().(*sequence)
	err := errors.New("foobar")

	seq.Run(ctx, p)
	out.Reset()
	ctx.SetErr(err)
	seq.Rollback(ctx, p)

	is.Equal(err, ctx.Err())
	is.Empty(out.String())
}

func TestSequence_Rollback_Success(t *testing.T) {
	is := assert.New(t)

	cmdA := &MockCommand{name: "A"}
	cmdB := &MockCommand{name: "B"}
	cmdC := &MockCommand{name: "C"}

	ctx := NewContext(3)
	seq := NewSequence(cmdA, cmdB, cmdC).(*sequence)
	seq.Run(ctx, DefaultPrinter)
	seq.Rollback(ctx, DefaultPrinter)

	is.NoError(ctx.Err())
	for _, cmd := range []*MockCommand{cmdA, cmdB, cmdC} {
		is.True(cmd.ran)
		is.False(cmd.failed)
		is.True(cmd.rolledBack)
	}
}