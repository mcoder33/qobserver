package main

import (
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

type observeTestSet struct {
	qName string
	out   string
	qInfo *queueInfo
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
			qInfo: &queueInfo{
				waiting:  0,
				delayed:  0,
				reserved: 0,
				done:     0,
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
			qInfo: &queueInfo{
				waiting:  11,
				delayed:  22,
				reserved: 33,
				done:     44,
			},
		},
	}

	for _, test := range testSet {
		t.Run(test.qName, func(t *testing.T) {
			cmd := &svCmd{
				name:    test.qName,
				command: []string{"any", "command"},
				execFn: func(name string, arg ...string) ([]byte, error) {
					return []byte(test.out), nil
				},
			}

			qi, err := cmd.execute()

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

	cmd := &svCmd{
		name:    "benchmarkQueue",
		command: []string{"any", "command"},
		execFn: func(name string, arg ...string) ([]byte, error) {
			return []byte(out), nil
		},
	}

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			_, _ = cmd.execute()
		}(&wg)
	}
	wg.Wait()

	b.ReportAllocs()
}
