package models

import (
	"encoding/json"
	"time"
)

type SecurityIncident struct {
	ID                string          `json:"id"`
	AgentID           string          `json:"agent_id"`
	SessionID         *string         `json:"session_id,omitempty"`
	IncidentType      string          `json:"incident_type"`
	Severity          string          `json:"severity"`
	Title             string          `json:"title"`
	Description       *string         `json:"description,omitempty"`
	Status            string          `json:"status"`
	EventCount        int             `json:"event_count"`
	AssignedTo        *string         `json:"assigned_to,omitempty"`
	InvestigationNote *string         `json:"investigation_note,omitempty"`
	Resolution        *string         `json:"resolution,omitempty"`
	InvestigatingAt   *time.Time      `json:"investigating_at,omitempty"`
	FirstSeenAt       time.Time       `json:"first_seen_at"`
	LastSeenAt        time.Time       `json:"last_seen_at"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	ResolvedAt        *time.Time      `json:"resolved_at,omitempty"`
	Metadata          json.RawMessage `json:"metadata"`
}

type InvestigateIncidentRequest struct {
	AssignedTo        string `json:"assigned_to"`
	InvestigationNote string `json:"investigation_note"`
}

type ResolveIncidentRequest struct {
	Resolution string `json:"resolution"`
}

type IncidentListFilter struct {
	Status     string
	Severity   string
	AssignedTo string
	Limit      int
	Offset     int
}

type IncidentListResponse struct {
	Incidents  []SecurityIncident `json:"incidents"`
	Pagination IncidentPagination `json:"pagination"`
}

type IncidentPagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type IncidentStatusResponse struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

type IncidentMetricsSummary struct {
	Open           uint64
	Investigating  uint64
	Resolved       uint64
	Dismissed      uint64
	CriticalOpen   uint64
	UnassignedOpen uint64
	Total          uint64
}
