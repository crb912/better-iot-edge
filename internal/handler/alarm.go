// internal/handler/alarm.go
// AlarmHandler 实现自定义路由 POST /api/alarm。

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

// AlarmHandler 持有 logger，供处理函数使用。
type AlarmHandler struct {
	lc logger.LoggingClient
}

// NewAlarmHandler 构造 AlarmHandler。
func NewAlarmHandler(lc logger.LoggingClient) *AlarmHandler {
	return &AlarmHandler{lc: lc}
}

// alarmRequest 是请求体的反序列化结构。
type alarmRequest struct {
	Values []int `json:"values"`
}

// alarmResponse 是响应体的序列化结构。
type alarmResponse struct {
	First int `json:"first"`
}

// HandleAlarm 处理 POST /api/alarm。
// 请求体示例：{"values": [3, 1, 4, 1, 5]}
// 响应体示例：{"first": 3}
func (h *AlarmHandler) HandleAlarm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req alarmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.lc.Warnf("AlarmHandler: failed to decode request body: %v", err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if len(req.Values) == 0 {
		http.Error(w, "values array must not be empty", http.StatusBadRequest)
		return
	}

	h.lc.Debugf("AlarmHandler: received %d values, first=%d", len(req.Values), req.Values[0])

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(alarmResponse{First: req.Values[0]})
}
