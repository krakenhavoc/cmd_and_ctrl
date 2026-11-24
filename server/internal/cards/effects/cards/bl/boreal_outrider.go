package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boreal Outrider — Snow Creature — Elf Warrior {2}{G}, 3/2 (Kaldheim, #5505):
//
//	"Whenever you cast a creature spell, if {S} of any of that spell's
//	 colors was spent to cast it, that creature enters with an additional
//	 +1/+1 counter on it. ({S} is mana from a snow source.)"
//
// Wildgrowth Archaic's cast trigger with a fixed one counter, gated (CR
// 603.4) on the payment: snow mana of one of the spell's colors, read as
// the spell is cast. The counter rides the creature's entry, so Hardened
// Scales sees it. Snow mana is never paper, so the payment read is the
// pool's alone.
func init() {
	t := effects.WheneverYouCastThatEntersWithCounters("Boreal Outrider", effects.Creature(), game.CounterPlusOne,
		func(game.ManaSpent) int { return 1 },
		"Boreal Outrider — that creature enters with an additional +1/+1 counter on it")
	cast := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return cast(ev, source, lki, g) && effects.SnowManaOfASpellsColorWasSpent(ev, source, lki, g)
	}
	effects.Register(effects.Spec{
		OracleID:       "19ef24b7-25f5-4ae2-b892-3e15ed62d299",
		Name:           "Boreal Outrider",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts | game.ReadsPoolManaOnly,
		WantsManaFrom:  game.ManaSourceSnow,
		Triggered:      []game.TriggeredAbility{t},
	})
}
