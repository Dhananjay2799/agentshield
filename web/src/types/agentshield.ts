export type Agent = {
  id: string
  name: string
  agent_type: string
  owner: string
  framework: string
  model: string
  environment: string
  status: string
  created_at: string
  updated_at: string
}

export type AgentSession = {
  id: string
  agent_id: string
  task_id: string
  status: string
  started_at: string
  ended_at?: string
  expires_at?: string
}

export type SessionSecurity = {
  id: string
  agent_id: string
  task_id: string
  status: string
  started_at: string
  ended_at?: string | null
  expires_at?: string | null

  action_count: number
  allowed_count: number
  denied_count: number
  approval_count: number
  highest_risk_score: number

  last_action_at?: string | null
}

export type AuditMetadata = {
  grant_id?: string
  agent_type?: string
  approval_id?: string
  environment?: string
  risk_reason?: string
  policy_engine?: string
  policy_reason?: string
  request_reason?: string
  [key: string]: unknown
}

export type AuditEvent = {
  id: string
  agent_id: string
  session_id: string
  event_type: string
  action: string
  resource: string
  decision: string
  risk_score: number
  metadata: AuditMetadata
  created_at: string
}

export type ApprovalRequest = {
  id: string
  agent_id: string
  session_id: string
  action: string
  resource: string
  reason: string
  risk_score: number
  status: string
  requested_at: string
  approved_at?: string
  denied_at?: string
  expires_at?: string
}

export type IncidentMetadata = {
  action?: string
  resource?: string
  risk_score?: number
  event_count_window?: number
  [key: string]: unknown
}

export type Incident = {
  id: string
  agent_id: string
  session_id?: string
  incident_type: string
  severity: string
  title: string
  description?: string
  status: string
  event_count: number
  first_seen_at: string
  last_seen_at: string
  created_at: string
  resolved_at?: string
  metadata?: IncidentMetadata
}

export type SecurityEventMetadata = {
  grant_id?: string
  agent_type?: string
  approval_id?: string
  environment?: string
  risk_reason?: string
  policy_engine?: string
  policy_reason?: string
  request_reason?: string
  [key: string]: unknown
}

export type SecurityEvent = {
  id: string
  agent_id: string
  session_id: string
  event_type: string
  action: string
  resource: string
  decision: string
  risk_score: number
  metadata: SecurityEventMetadata
  created_at: string
}