package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/events"
)

type EventStreamHandler struct {
	Hub *events.Hub
}

func NewEventStreamHandler(
	hub *events.Hub,
) *EventStreamHandler {
	return &EventStreamHandler{
		Hub: hub,
	}
}

func (h *EventStreamHandler) Stream(
	w http.ResponseWriter,
	r *http.Request,
) {
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(
			w,
			"streaming unsupported",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Access-Control-Allow-Origin",
		"http://localhost:3000",
	)

	w.Header().Set(
		"Access-Control-Allow-Methods",
		"GET",
	)

	w.Header().Set(
		"Access-Control-Allow-Headers",
		"Content-Type",
	)

	w.Header().Set(
		"Content-Type",
		"text/event-stream",
	)

	w.Header().Set(
		"Cache-Control",
		"no-cache",
	)

	w.Header().Set(
		"Connection",
		"keep-alive",
	)

	w.Header().Set(
		"X-Accel-Buffering",
		"no",
	)

	subscriber :=
		h.Hub.Subscribe()

	defer h.Hub.Unsubscribe(
		subscriber,
	)

	fmt.Fprint(
		w,
		": AgentShield event stream connected\n\n",
	)

	flusher.Flush()

	heartbeat :=
		time.NewTicker(15 * time.Second)

	defer heartbeat.Stop()

	for {
		select {
		case event, ok :=
			<-subscriber:

			if !ok {
				return
			}

			payload, err :=
				json.Marshal(event)

			if err != nil {
				continue
			}

			fmt.Fprintf(
				w,
				"event: security_event\n",
			)

			fmt.Fprintf(
				w,
				"data: %s\n\n",
				payload,
			)

			flusher.Flush()

		case <-heartbeat.C:
			fmt.Fprint(
				w,
				": heartbeat\n\n",
			)

			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}
