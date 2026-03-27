// internal/handler/alarm_test.go
package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func newTestHandler() *AlarmHandler {
	return NewAlarmHandler(logger.NewMockClient())
}

func TestHandleAlarm_Success(t *testing.T) {
	body := `{"values": [3, 1, 4, 1, 5]}`
	req := httptest.NewRequest(http.MethodPost, "/api/alarm", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newTestHandler().HandleAlarm(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp alarmResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.First != 3 {
		t.Errorf("expected first=3, got %d", resp.First)
	}
}

func TestHandleAlarm_EmptyArray(t *testing.T) {
	body := `{"values": []}`
	req := httptest.NewRequest(http.MethodPost, "/api/alarm", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	newTestHandler().HandleAlarm(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleAlarm_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/alarm", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()

	newTestHandler().HandleAlarm(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleAlarm_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/alarm", nil)
	w := httptest.NewRecorder()

	newTestHandler().HandleAlarm(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
