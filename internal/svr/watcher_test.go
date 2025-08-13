package svr

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	type watcherTestSet struct {
		out   string
		qInfo *QueueInfo
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
			qInfo: &QueueInfo{
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
			qInfo: &QueueInfo{
				Name:     "testFilled",
				Waiting:  11,
				Delayed:  22,
				Reserved: 33,
				Done:     44,
			},
		},
	}

	var commands []*Cmd
	for _, test := range testSet {
		commands = append(commands, NewCmd(test.qInfo.Name, []string{"any", "cmd"}, func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			return []byte(test.out), nil
		}))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res := map[string]*QueueInfo{}
	tWatcher := NewWatcher(1*time.Millisecond, 1*time.Second)
	for qi := range tWatcher.Run(ctx, commands) {
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
