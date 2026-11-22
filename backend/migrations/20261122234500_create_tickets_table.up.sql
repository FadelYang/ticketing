CREATE TABLE IF NOT EXISTS tickets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title        VARCHAR(250) NOT NULL,
    description  TEXT NOT NULL,
    category     VARCHAR(100) NOT NULL,
    priority     VARCHAR(20) NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    status       VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'resolved')),
    requester_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assignee_id  UUID NULL DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deadline_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tickets_created_at ON tickets (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets (status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets (priority);
CREATE INDEX IF NOT EXISTS idx_tickets_requester_id ON tickets (requester_id);
CREATE INDEX IF NOT EXISTS idx_tickets_assignee_id ON tickets (assignee_id);
