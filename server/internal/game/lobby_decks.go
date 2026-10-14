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
