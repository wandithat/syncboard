CREATE TABLE IF NOT EXISTS board_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    invitee_email VARCHAR(255) NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    responded_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_board_invitations_pending_unique
    ON board_invitations(board_id, LOWER(invitee_email))
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_board_invitations_invitee
    ON board_invitations(LOWER(invitee_email), status);
