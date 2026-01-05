package service

import (
	"context"
	"testing"
	"time"

	"github.com/mcoder33/qobserver/internal/cmd"
	"github.com/mcoder33/qobserver/internal/model"
	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	type watcherTestSet struct {
		out   string
		qInfo *model.QueueInfo
	}

	testSet := map[string]watcherTestSet{
		"testZero": {
			out: `
Jobs
- waiting: 0
- delayed: 0
- reserved: 0
- done: 0
`,
			qInfo: &model.QueueInfo{
				Name:     "testZero",
				Waiting:  0,
				Delayed:  0,
				Reserved: 0,
				Done:     0,
			},
		},
		"testFilled": {
			out: `
Jobs
- waiting: 11
- delayed: 22
- reserved: 33
- done: 44
`,
			qInfo: &model.QueueInfo{
				Name:     "testFilled",
				Waiting:  11,
				Delayed:  22,
				Reserved: 33,
				Done:     44,
			},
		},
	}

	commands := make(map[string]*cmd.Process, len(testSet))
	for _, test := range testSet {
		commands[test.qInfo.Name] = cmd.New(test.qInfo.Name, []string{"any", "cmd"}, func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			return []byte(test.out), nil
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := cmd.Pool{
		Commands: commands,
	}

	res := map[string]*model.QueueInfo{}
	tWatcher := NewWatcher(1*time.Millisecond, 1*time.Second)
	for qi := range tWatcher.Run(ctx, &pool) {
		res[qi.Name] = qi
		if len(res) == len(commands) {
			cancel()
		}
	}

	require.Equal(t, len(commands), len(res))
	for qName, data := range testSet {
		require.Equal(t, data.qInfo, res[qName])
	}
}
