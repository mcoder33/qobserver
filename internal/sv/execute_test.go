package sv

import (
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

type observeTestSet struct {
	qName string
	out   string
	qInfo *QueueInfo
}

func TestObserve(t *testing.T) {
	testSet := []observeTestSet{
		{
			qName: "testZero",
			out: `
Jobs
- waiting: 0
- delayed: 0
- reserved: 0
- done: 0
`,
			qInfo: &QueueInfo{
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
			qName: "testFilled",
			qInfo: &QueueInfo{
				Waiting:  11,
				Delayed:  22,
				Reserved: 33,
				Done:     44,
			},
		},
	}

	for _, test := range testSet {
		t.Run(test.qName, func(t *testing.T) {
			cmd := &Cmd{
				name:    test.qName,
				command: []string{"any", "cmd"},
				execFn: func(name string, arg ...string) ([]byte, error) {
					return []byte(test.out), nil
				},
			}

			qi, err := cmd.Execute()

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
		execFn: func(name string, arg ...string) ([]byte, error) {
			return []byte(out), nil
		},
	}

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			_, _ = cmd.Execute()
		}(&wg)
	}
	wg.Wait()

	b.ReportAllocs()
}
