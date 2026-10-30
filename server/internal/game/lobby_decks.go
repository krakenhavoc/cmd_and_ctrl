package game

import "github.com/google/uuid"

// LobbyDecks returns, for every seat of a game still in the lobby, a deep
// copy of the deck it would start with: the command-zone cards first, then
// the library in the order they were installed. It is what the lobby keeps
// as a table starts so that a rematch (#3285) can seat the same decks in a
// new game. Each card keeps its printed data and gets a fresh InstanceID,
// so the copy shares nothing with the game it came from, and the new game
// does not reuse the old game's card identities.
//
// Nil once the game has left the lobby: a started game's zones hold the
// shuffled, dealt and played cards, not the deck.
func (g *Game) LobbyDecks() map[uuid.UUID][]Card {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.State != StateLobby {
		return nil
	}
	out := make(map[uuid.UUID][]Card, len(g.Seats))
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		deck := make([]Card, 0, len(p.Command.Cards)+len(p.Library.Cards))
		for _, zone := range []*Zone{p.Command, p.Library} {
			for _, c := range zone.Cards {
				cp := cloneCard(c)
				cp.InstanceID = uuid.New()
				deck = append(deck, cp)
			}
		}
		out[p.ID] = deck
	}
	return out
}

// CommandersOf returns a deep copy of each of player's commanders, for
// the lobby's seat line (ADR 0212 phase B2), which reads them again from
// a game restored after a restart. While the game is in the lobby they
// are the commander-zone cards the deck installed, in the deck's order,
// which is also the order LobbyDecks gives. Once the game is under way a
// commander may be anywhere: the designation stays with the card through
// every zone change (CR 903.3), so every zone is walked for a card the
// player owns marked IsCommander. A token is never a commander, even one
// copying a commander: the designation is an attribute of the card, not
// a characteristic (CR 903.3), so a copy does not acquire it (CR 707.2).
// Nil for a player who is not seated, or who has none.
func (g *Game) CommandersOf(playerID uuid.UUID) []Card {
	g.mu.RLock()
	defer g.mu.RUnlock()
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil
	}
	var out []Card
	take := func(z *Zone) {
		if z == nil {
			return
		}
		for _, c := range z.Cards {
			if c.IsCommander && c.Owner == p.ID && !c.IsToken() {
				out = append(out, cloneCard(c))
			}
		}
	}
	if g.State == StateLobby {
		take(p.Command)
		return out
	}
	g.eachIndexedZoneLocked(func(_ cardSlot, _ int, z *Zone) bool {
		take(z)
		return false
	})
	return out
}
