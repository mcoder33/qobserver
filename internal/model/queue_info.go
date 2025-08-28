package model

import "fmt"

type QueueInfo struct {
	Name     string
	Waiting  int
	Delayed  int
	Reserved int
	Done     int
}

func (i QueueInfo) String() string {
	return fmt.Sprintf("QueueName: %s\nWaiting: %d\nDelayed: %d;\nReserved: %d;\nDone: %d;", i.Name, i.Waiting, i.Delayed, i.Reserved, i.Done)
}
