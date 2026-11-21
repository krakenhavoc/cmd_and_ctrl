package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// r37_third_spell.go — the "third spell each turn" clock, shared by
// Emeritus of Conflict (cards/e) and Captain Ripley Vance (cards/bu), which
// sit in different shards since the #5906 split (S82 GAP roll-up 37).

// R37YouCastYourThirdSpellEachTurn is "whenever you cast your third spell
// each turn" (YouCastYourSecondSpellEachTurn's shape): the cast path bumps
// the per-turn tally before it emits EventCast, so a total of exactly three
// means "this is the third".
func R37YouCastYourThirdSpellEachTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	return g.CastTallyFor(ev.Actor).Total == 3
}
