package app

import (
	"net/http"
	"strings"
	"time"

	"appengine"
	"appengine/taskqueue"
	"github.com/MiCHiLU/appengine-delay"

	u "logutil"
)

const ()

var (
	delayTestTQ *delay.Function

	b        = []byte("012345678")
	deadline = time.Duration(60) * time.Second
	payload  = make([]byte, 0, 65535)
)

func init() {
	delayTestTQ = delay.Func("runTestTQ", runTestTQ)
	payload = append(payload, b...)
	for i := 0; i < 16; i++ {
		payload = append(payload, payload...)
	}
}

func testTQ(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	for i := 0; i < 4; i++ {
		delayTestTQ.Call(c, "")
	}
}

func runTestTQ(
	c appengine.Context,
	r *http.Request,
) {
	//u.Debugf(c, "payload: %q", len(payload))
	c = appengine.Timeout(c, deadline)

	_, err := taskqueueAdd(c, &taskqueue.Task{
		Payload: payload,
		Method:  "PULL",
	}, "pull")
	if err != nil {
		u.Errorf(c, "%v", err)
		return
	}

	delayTestTQ.Call(c, "")
	return
}

func taskqueueAdd(c appengine.Context, task *taskqueue.Task, queueName string) (*taskqueue.Task, error) {
	for {
		task, err := taskqueue.Add(c, task, queueName)
		if isTransientError(err) {
			u.Infof(c, "%v", err.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		if isTombstonedTask(err) {
			u.Infof(c, "%v", err.Error())
			return task, nil
		}
		return task, err
	}
}

func isTransientError(err error) bool {
	if err != nil && strings.Contains(err.Error(), "taskqueue: TRANSIENT_ERROR") {
		return true
	}
	return false
}

func isTombstonedTask(err error) bool {
	if err != nil && strings.Contains(err.Error(), "taskqueue: TOMBSTONED_TASK") {
		return true
	}
	return false
}
