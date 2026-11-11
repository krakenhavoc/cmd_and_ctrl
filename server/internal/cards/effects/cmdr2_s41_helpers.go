package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s41_helpers.go — helpers of the spellslinger batch (cmdr2 batch 41).

// cmdr2S41CascadeTheSpellBody is the delayed trigger's body for "the next
// <type> spell you cast this turn has cascade" (Smoldering Stagecoach), the
// cascade sibling of stormTheSpellBody. The spell the watch fired on rides
// the item as its payload; cascade (CR 702.85a) exiles until a nonland card
// with lesser mana value than that spell, which is read as the trigger
// resolves (an X spell counts its X, CR 202.3e) from last-known information
// if the spell has left the stack (CR 608.2h).
var cmdr2S41CascadeTheSpellBody = game.SimpleDelayedBody("cascade/the-spell-just-cast", cmdr2S41CascadeTheSpell)

func cmdr2S41CascadeTheSpell(g *game.Game, item *game.StackItem) error {
	cast := NewContext(g, item).PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	mv := 0
	if spell, ok := g.LookupCardForEffect(cast[0]); ok {
		mv, _ = g.ManaValueForEffect(spell)
	}
	return g.CascadeForEffect(item.Controller, item.SourceCardID, mv)
}

// cmdr2S41DifferentInstantSorceryManaValues is the number of different mana
// values among the instant and sorcery cards in a player's graveyard (Eris,
// Roar of the Storm). Mana value is read through the layered view, and an X
// in a card's cost is 0 outside the stack (CR 202.3e).
func cmdr2S41DifferentInstantSorceryManaValues(g *game.Game, controller uuid.UUID) int {
	p := g.PlayerByIDForEffect(controller)
	if p == nil {
		return 0
	}
	seen := map[int]bool{}
	for _, c := range p.Graveyard.Cards {
		if c.IsInstant() || c.IsSorcery() {
			mv, _ := g.ManaValueForEffect(c)
			seen[mv] = true
		}
	}
	return len(seen)
}

// cmdr2S41CastThreeOrMoreSpellsThisTurn is Stella Lee's activation gate,
// "Activate only if you've cast three or more spells this turn": the
// per-turn cast tally, which counts every spell the player cast (copies are
// not cast and are not counted, CR 707.10).
func cmdr2S41CastThreeOrMoreSpellsThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	return g.CastTallyFor(controller).Total >= 3
}
