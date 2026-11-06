package lobby

// personal_blocks.go: a personal block keeps the blocked person out of
// the blocker's tables (ADR 0213 §3, owner decision 6 (a), slice 3 of
// #3313).
//
// The block list (internal/social, PUT /me/blocks/{user_id}) already
// reaches lists, requests and invites. This is the fourth reach: "a seat
// at every table the blocker hosts". It acts as a standing table_blocks
// entry (blocks.go) on each table its owner hosts, so it is checked where
// a table block is: when a player invite is previewed and when a seat is
// claimed, by invite link (POST /games/{id}/join), by code (POST /join)
// and by the Discord sign-in that carries an invite (the OAuth callback).
// All three of the last reach Lobby.join.
//
// WHOSE BLOCK. The HOST's block of the joiner, one direction. A joiner
// who blocked the host is not kept out: the block is the blocker's own
// list, and the ADR gives it no reach over a table somebody else hosts
// ("at someone else's table, that host's tools apply", ADR 0075 §2.6).
// The host is the seat the room names (it moves when the host passes the
// table on or leaves, so the check follows whoever hosts at that moment);
// a table nobody human has sat at yet is hosted by its signed-in creator,
// unless it is waiting for a named host (HostDiscordID).
//
// WHAT IT DOES NOT TOUCH.
//
//   - Someone already seated stays seated. A block made later does not
//     remove a seat, close a socket or stop a reclaim; that is the host's
//     Remove control (seat_remove.go), a decision about one table. The
//     check also comes after "already seated" in join, so a blocked
//     person who is at the table is still told so, not that the invite
//     has gone.
//   - Watching. Decision 6 names a seat, so a spectator invite is
//     neither refused at its preview nor at Spectate.
//   - A guest or an agent seat: they have no account to block.
//
// THE REFUSAL NEVER SAYS WHY. The blocked person is told what a host who
// rotated the invite would have told them: ErrInvalidInvite. It is not
// ErrBlockedFromTable ("you were removed from this table"), which names
// a decision about the person, and not a made-up "full" or "started",
// which a lobby page that lists the seats would contradict. The same
// answer comes from the preview, the join and the join by code, and it
// is byte for byte the answer for a wrong token
// (TestPersonalBlockKeepsTheBlockedPersonOutOfTheHostsTable).

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// PersonalBlocks answers whether one signed-in person has blocked
// another. *social.Store is the implementation; the interface keeps the
// lobby from importing the store.
type PersonalBlocks interface {
	// IsBlocking reports whether user has blocked target (the blocker's
	// own list, one direction). Ids are users(id) strings.
	IsBlocking(ctx context.Context, user, target string) (bool, error)
}

// SetPersonalBlocks wires the block list in after construction. Nil, or
// never calling it, means no personal block reaches a table.
func (l *Lobby) SetPersonalBlocks(b PersonalBlocks) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.personalBlocks = b
}

// hostUserLocked is the account that hosts the table now, or uuid.Nil
// when nobody with an account does. Caller holds l.mu.
func (e *gameEntry) hostUserLocked() uuid.UUID {
	if e.room != nil {
		if h := e.room.HostPlayerID(); h != uuid.Nil {
			for _, s := range e.meta.Players {
				if s.PlayerID == h {
					id, _ := uuid.Parse(s.UserID)
					return id
				}
			}
			return uuid.Nil
		}
	}
	// Nobody has sat down to host yet. A named host who has not sat is
	// the host-to-be; otherwise the creator is the one who manages it.
	if e.meta.HostDiscordID != "" {
		return uuid.Nil
	}
	return e.meta.CreatedBy
}

// personalBlockLocked reports whether the table's host has blocked
// userID. A guest, an agent and a table with no account host are never
// blocked. A failed lookup is an error, not a pass: it would otherwise
// seat the person a host has blocked. Caller holds l.mu.
func (l *Lobby) personalBlockLocked(entry *gameEntry, userID uuid.UUID) (bool, error) {
	if l.personalBlocks == nil || userID == uuid.Nil {
		return false, nil
	}
	host := entry.hostUserLocked()
	if host == uuid.Nil || host == userID {
		return false, nil
	}
	blocked, err := l.personalBlocks.IsBlocking(context.Background(), host.String(), userID.String())
	if err != nil {
		return false, fmt.Errorf("checking the host's block list: %w", err)
	}
	return blocked, nil
}
