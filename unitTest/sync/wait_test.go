package sync_test

import (
	"redisGo/lib/sync/wait"
	"testing"
	"time"
)

func TestWaitWithTimeout(t *testing.T) {
	w := &wait.Wait{}

	// 测试正常完成的情况
	w.Add(1)
	go func() {
		defer w.Done()
		time.Sleep(100 * time.Millisecond) // 模拟工作
	}()
	if timeout := w.WaitWithTimeout(200 * time.Millisecond); timeout {
		t.Errorf("WaitWithTimeout() = true, want false")
	}
	// 测试超时的情况
	w.Add(1)
	go func() {
		defer w.Done()
		time.Sleep(300 * time.Millisecond) // 模拟工作
	}()

	if timeout := w.WaitWithTimeout(200 * time.Millisecond); !timeout {
		t.Errorf("WaitWithTimeout() = false, want true")
	}
}
