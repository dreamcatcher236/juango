package integration_test

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/juanfont/juango/tasks"
)

// syncBuffer is a goroutine-safe buffer. os/exec writes process output from a
// separate goroutine while the test polls it, so a plain bytes.Buffer would race.
type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// TestTasksWorkerProcessesDummyTask exercises the full background-task path of a
// scaffolded project: it starts the generated `worker` binary, enqueues a dummy
// task through juango's task client, and confirms the worker processes it.
//
// The task type and payload mirror the scaffolded internal/tasks/dummy.go
// (TypeDummyTask = "dummy:hello").
func TestTasksWorkerProcessesDummyTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping worker test in short mode")
	}
	if testServer == nil {
		t.Fatal("test server not initialized")
	}
	if !redisReachable(testServer.redisAddr) {
		t.Skipf("skipping worker test: Redis not reachable at %s", testServer.redisAddr)
	}

	// Start the scaffolded worker as a subprocess.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerCmd := exec.CommandContext(ctx, testServer.binPath, "worker", "-c", testServer.configPath)
	workerCmd.Dir = testServer.projectDir

	var workerOutput syncBuffer
	workerCmd.Stdout = &workerOutput
	workerCmd.Stderr = &workerOutput

	if err := workerCmd.Start(); err != nil {
		t.Fatalf("failed to start worker: %v", err)
	}
	defer func() {
		cancel()
		workerCmd.Wait()
	}()

	// Give the worker a moment to connect to Redis and start polling.
	if !waitForOutput(&workerOutput, "waiting for tasks", 10*time.Second) {
		t.Fatalf("worker did not start in time.\nOutput:\n%s", workerOutput.String())
	}

	// Enqueue a dummy task via juango's task client.
	client := tasks.NewClient(testServer.redisAddr, "", 0)
	defer client.Close()

	const dummyTaskType = "dummy:hello"
	info, err := client.Enqueue(dummyTaskType, map[string]interface{}{
		"message": "smoke-test",
		"count":   1,
	})
	if err != nil {
		t.Fatalf("failed to enqueue dummy task: %v", err)
	}
	t.Logf("enqueued task %s on queue %s", info.ID, info.Queue)

	// The handler logs this line on success.
	if !waitForOutput(&workerOutput, "Dummy task completed successfully", 20*time.Second) {
		t.Fatalf("worker did not process the dummy task in time.\nOutput:\n%s", workerOutput.String())
	}
}

// waitForOutput polls buf until it contains substr or the timeout elapses.
func waitForOutput(buf *syncBuffer, substr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), substr) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return strings.Contains(buf.String(), substr)
}
