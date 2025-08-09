package sv

import (
	"context"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestExecute(t *testing.T) {
	testSet := []struct {
		qName string
		out   string
		qInfo *QueueInfo
	}{
		{
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
		{
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

	for _, test := range testSet {
		qName := test.qInfo.Name
		t.Run(qName, func(t *testing.T) {
			cmd := &Cmd{
				name:    qName,
				command: []string{"any", "cmd"},
				execFn: func(ctx context.Context, name string, arg ...string) ([]byte, error) {
					return []byte(test.out), nil
				},
			}

			qi, err := cmd.Execute(context.Background())

			require.NoError(t, err)
			require.Equal(t, test.qInfo, qi)
		})
	}
}

func BenchmarkObserve(b *testing.B) {
	out := `
Jobs
- waiting: 11
- delayed: 22
- reserved: 33
- done: 44
`

	cmd := &Cmd{
		name:    "benchmarkQueue",
		command: []string{"any", "cmd"},
		execFn: func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			return []byte(out), nil
		},
	}

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			_, _ = cmd.Execute(context.Background())
		}(&wg)
	}
	wg.Wait()

	b.ReportAllocs()
}
