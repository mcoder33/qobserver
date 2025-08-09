package main

import (
	"context"
	"github.com/stretchr/testify/require"
	"qobserver/internal/sv"
	"testing"
)

type observeTestSet struct {
	qName string
	out   string
	qInfo *sv.QueueInfo
}

func TestObserve(t *testing.T) {
	testSet := map[string]observeTestSet{
		"testZero": {
			out: `
Jobs
- waiting: 0
- delayed: 0
- reserved: 0
- done: 0
`,
			qInfo: &sv.QueueInfo{
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
			qInfo: &sv.QueueInfo{
				Name:     "testFilled",
				Waiting:  11,
				Delayed:  22,
				Reserved: 33,
				Done:     44,
			},
		},
	}

	var cmdPoolTest []*sv.Cmd
	for _, test := range testSet {
		cmdPoolTest = append(cmdPoolTest, sv.NewCmd(test.qInfo.Name, []string{"any", "cmd"}, func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			return []byte(test.out), nil
		}))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res := map[string]*sv.QueueInfo{}
	for qi := range observe(ctx, 1, cmdPoolTest) {
		res[qi.Name] = qi
		if len(res) == len(cmdPoolTest) {
			cancel()
		}
	}

	require.Equal(t, len(cmdPoolTest), len(res))
	for qName, data := range testSet {
		require.Equal(t, data.qInfo, res[qName])
	}
}
