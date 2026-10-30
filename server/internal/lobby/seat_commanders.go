package lobby

// seat_commanders.go: a seat's commanders on the lobby card (ADR 0212
// phase B2).
//
// A seat carries its deck's commanders, resolved against the card index
// by libraryCommanders, the code GET /me/decks, GET /decks and GET
// /bot/options already resolve a deck's commanders with, so a seat, a
// saved deck's tile and a pre-built deck's tile always draw the same
// printing. They are set wherever a deck goes into a seat (SetDeck, a
// bot, a rematch seat, the practice table) from the cards the engine is
// given, and read again from the restored game after a restart.
//
// KEPT ACROSS A RESTART BY DERIVING, NOT BY A COLUMN. The commanders are
// the engine's own commander cards, which the restore point already
// holds with the rest of the deck, so loadEntry reads them back from the
// restored game (game.CommandersOf) the way it reads DeckUploaded, the
// Discord avatar and the agent badge, rather than keeping a second copy
// in seats that could disagree with the cards the seat plays. A table
// only comes back when its restore point does, so there is always a game
// to read. No migration means nothing new is stored about anyone and a
// rollback to the previous binary stays a plain swap (db.ErrSchemaTooNew
// would refuse it otherwise).
//
// WHO SEES THEM (owner decision 3). The table's own people: its seats,
// its spectators, its creator and the admin. deck_name and deck_power
// reach every signed-in session through GET /games, and the commanders
// must not, so the lobby never puts them on its own copy of a seat:
// they live in SeatInfo.commanders, which has no JSON name, and only
// redactMetaFor copies them onto the wire fields, for a viewer this file
// says may see them. Any path that serializes a seat without passing
// through redactMetaFor (the invite preview a visitor reads before they
// have a session, a join, spectate or session answer) therefore sends
// none. The Discord bot lists tables as the admin, so it is sent them,
// and /c2-games names no seat's deck.
//
// The same rule holds once the game is under way: the command zone is
// public to the game's players (CR 400.2), which is the table, not
// everyone who can list it, and the in-play card draws no art (ADR 0212
// §7). The board's own view, which shows the command zone to the
// table's seats and spectators, is unchanged.

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seatCommanders is what libraryCommanders makes of a seated deck's
// commanders: one row each, and the union of their identities.
type seatCommanders struct {
	cards    []myDeckCommander
	identity []string
}

// SetCardIndex gives the lobby the card index a seat's commanders are
// resolved against. Call it once at boot, before RestoreFromDisk, so a
// restored table's seats get theirs back; nil (the default) gives no
// seat any.
func (l *Lobby) SetCardIndex(idx *cards.Index) {
	l.cardIndex.Store(idx)
}

// commandersOf resolves the commanders among a seat's cards. It needs
// no lock: the index pointer is atomic and the index locks itself.
func (l *Lobby) commandersOf(seated []game.Card) *seatCommanders {
	return resolveSeatCommanders(l.cardIndex.Load(), seated)
}

// resolveSeatCommanders hands the cards marked as commanders to
// libraryCommanders, with each one's printing as the hint, so a deck
// that names a printing (a saved deck's pinned art, a pasted list's set
// and number) draws that one, as its Decks page card does. Nil with no
// index, or for a deck with no commander.
func resolveSeatCommanders(idx *cards.Index, seated []game.Card) *seatCommanders {
	if idx == nil {
		return nil
	}
	var names []string
	printings := map[string]deck.Entry{}
	for _, c := range seated {
		if !c.IsCommander || strings.TrimSpace(c.Name) == "" {
			continue
		}
		names = append(names, c.Name)
		if id, err := uuid.Parse(c.ScryfallID); err == nil {
			printings[strings.ToLower(strings.TrimSpace(c.Name))] = deck.Entry{PrintingID: id}
		}
	}
	rows, identity := libraryCommanders(idx, names, printings)
	if len(rows) == 0 {
		return nil
	}
	return &seatCommanders{cards: rows, identity: identity}
}

// tablesOwnPeople reports whether p is one of table id's own people,
// who may see the commanders in its seats (ADR 0212 owner decision 3):
//
//   - the admin, and the table's creator (creator is redactMetaFor's
//     IsCreator, already worked out for p);
//   - a seat: a player session bound to this table whose seat is still
//     here, or a signed-in person holding a seat here through another
//     session (their seat is theirs by account, as GET /me/games says);
//   - a spectator: a spectator session bound to this table. One the
//     table removed is refused before it gets here (auth.WithSeatCheck).
//
// Everyone else who can list the table, any signed-in session, is not.
func tablesOwnPeople(p auth.Principal, admin, creator bool, id uuid.UUID, meta GameMeta) bool {
	if admin || creator {
		return true
	}
	if p.GameID == id && id != uuid.Nil {
		switch p.Role {
		case auth.RoleSpectator:
			return true
		case auth.RolePlayer:
			for _, s := range meta.Players {
				if p.PlayerID != uuid.Nil && s.PlayerID == p.PlayerID {
					return true
				}
			}
		}
	}
	if p.UserID != uuid.Nil {
		if _, seated := seatOfUser(meta.Players, p.UserID); seated {
			return true
		}
	}
	return false
}

// withSeatCommanders is meta with each seat's commanders on the wire
// fields when reveal is true, and with neither field on any seat when it
// is false, whatever the caller's copy held. The seats slice is copied
// first, so the caller's meta is not touched.
func withSeatCommanders(meta GameMeta, reveal bool) GameMeta {
	seats := make([]SeatInfo, len(meta.Players))
	copy(seats, meta.Players)
	for i := range seats {
		seats[i].CommanderCards, seats[i].ColorIdentity = nil, nil
		if c := seats[i].commanders; reveal && c != nil {
			seats[i].CommanderCards = c.cards
			seats[i].ColorIdentity = c.identity
		}
	}
	meta.Players = seats
	return meta
}
