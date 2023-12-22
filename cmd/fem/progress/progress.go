package progress

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Progress struct {
	timer     time.Time
	msg       string
	start     int32
	stop      int32
	step      int32
	current   int32
	old       int32
	isStopped *bool
}

func NewProgress(msg string, start, stop, step int) *Progress {
	fmt.Printf("\r%s... 0%%", msg)
	return &Progress{time.Now(), msg, int32(start), int32(stop), int32(step), 0, 0, nil}
}

func NewUnlimitedProgress(msg string) *Progress {
	i := 0
	isStopped := false
	go func(isStopped *bool) {
		chr := []byte{'|', '/', '-', '\\'}
		for !*isStopped {
			fmt.Printf("\r%s... %c", msg, chr[i%4])
			i += 1
			if i > 3 {
				i = 0
			}
			time.Sleep(100 * time.Millisecond)
		}
	}(&isStopped)
	return &Progress{time.Now(), msg, 0, 0, 0, 0, 0, &isStopped}
}

func (p *Progress) AddProgress() {
	atomic.AddInt32(&p.current, 1)
	//p.current += 1
	duration := time.Since(p.timer)
	percent := int32((100.0 * float64(p.current)) / float64(p.stop-p.start))
	if p.current >= p.stop {
		//fmt.Printf("\r%s... 100%%", p.msg)
		fmt.Printf("\r%s... 100%%\n", p.msg)
		fmt.Printf("Done in: %0.2f sec\n", duration.Seconds())
		return
	}
	if percent == p.old {
		return
	}
	if percent%p.step == 0 {
		fmt.Printf("\r%s... %d%%", p.msg, percent)
	}
	p.old = percent
}

func (p *Progress) StopProgress() {
	duration := time.Since(p.timer)
	if p.isStopped != nil {
		*p.isStopped = true
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("\r%s... 100%%\n", p.msg)
	fmt.Printf("Done in: %0.2f sec\n", duration.Seconds())
}
