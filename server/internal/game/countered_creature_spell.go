package game

import "github.com/google/uuid"

// CreatureSpellCounteredByOpponentThisTurn reports whether a creature spell
// `player` cast this turn was countered by a spell or ability an opponent of
// `player` controlled — the condition of Summoning Trap ("If a creature spell
// you cast this turn was countered by a spell or ability an opponent
// controlled"). It reads this turn's event log: a cast by `player`, then an
// EventCounterSpell naming that spell whose cause is a spell or ability an
// opponent controls. The countered card must be a creature card, as it was
// cast. It adds no stored state.
//
// Caller must hold g.mu.
func (g *Game) CreatureSpellCounteredByOpponentThisTurn(player uuid.UUID) bool {
	if player == uuid.Nil {
		return false
	}
	cast := map[uuid.UUID]bool{}
	for _, ev := range g.EventsThisTurn() {
		switch ev.Kind {
		case EventCast:
			if ev.Actor == player {
				cast[ev.CardID] = true
			}
		case EventCounterSpell:
			if !cast[ev.Target] || ev.CauseController == uuid.Nil || ev.CauseController == player {
				continue
			}
			if c, ok := g.LookupCardForEffect(ev.Target); ok && c.HasCardType("creature") {
				return true
			}
		}
	}
	return false
}
