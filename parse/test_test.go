package parse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTest_HasDataRaceAndStatus(t *testing.T) {
	t.Parallel()

	now := time.Now()

	t.Run("passing test without race", func(t *testing.T) {
		test := &Test{
			Name:    "TestPassing",
			Package: "pkg",
			Events: []*Event{
				{Action: ActionRun, Time: now},
				{Action: ActionOutput, Output: "=== RUN   TestPassing\n", Time: now.Add(time.Millisecond)},
				{Action: ActionPass, Time: now.Add(2 * time.Millisecond)},
			},
		}
		assert.False(t, test.HasDataRace())
		assert.Equal(t, ActionPass, test.Status())
	})

	t.Run("passing test with race", func(t *testing.T) {
		test := &Test{
			Name:    "TestRacePassing",
			Package: "pkg",
			Events: []*Event{
				{Action: ActionRun, Time: now},
				{Action: ActionOutput, Output: "=== RUN   TestRacePassing\n", Time: now.Add(time.Millisecond)},
				{Action: ActionOutput, Output: "WARNING: DATA RACE\n", Time: now.Add(2 * time.Millisecond)},
				{Action: ActionPass, Time: now.Add(3 * time.Millisecond)},
			},
		}
		assert.True(t, test.HasDataRace())
		assert.Equal(t, ActionFail, test.Status())
	})

	t.Run("passing test with indented race", func(t *testing.T) {
		test := &Test{
			Name:    "TestRaceIndented",
			Package: "pkg",
			Events: []*Event{
				{Action: ActionRun, Time: now},
				{Action: ActionOutput, Output: "  WARNING: DATA RACE\n", Time: now.Add(time.Millisecond)},
				{Action: ActionPass, Time: now.Add(2 * time.Millisecond)},
			},
		}
		assert.True(t, test.HasDataRace())
		assert.Equal(t, ActionFail, test.Status())
	})

	t.Run("failing test without race", func(t *testing.T) {
		test := &Test{
			Name:    "TestFailing",
			Package: "pkg",
			Events: []*Event{
				{Action: ActionRun, Time: now},
				{Action: ActionOutput, Output: "--- FAIL: TestFailing (0.01s)\n", Time: now.Add(time.Millisecond)},
				{Action: ActionFail, Time: now.Add(2 * time.Millisecond)},
			},
		}
		assert.False(t, test.HasDataRace())
		assert.Equal(t, ActionFail, test.Status())
	})

	t.Run("skipped test without race", func(t *testing.T) {
		test := &Test{
			Name:    "TestSkipped",
			Package: "pkg",
			Events: []*Event{
				{Action: ActionRun, Time: now},
				{Action: ActionSkip, Time: now.Add(time.Millisecond)},
			},
		}
		assert.False(t, test.HasDataRace())
		assert.Equal(t, ActionSkip, test.Status())
	})

	t.Run("skipped test with race", func(t *testing.T) {
		test := &Test{
			Name:    "TestSkippedWithRace",
			Package: "pkg",
			Events: []*Event{
				{Action: ActionRun, Time: now},
				{Action: ActionOutput, Output: "WARNING: DATA RACE\n", Time: now.Add(time.Millisecond)},
				{Action: ActionSkip, Time: now.Add(2 * time.Millisecond)},
			},
		}
		assert.True(t, test.HasDataRace())
		assert.Equal(t, ActionFail, test.Status())
	})
}
