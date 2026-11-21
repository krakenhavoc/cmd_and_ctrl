package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quake, Agent of S.H.I.E.L.D. — Legendary Creature — Inhuman Spy Hero
// {2}{W}, 3/3 (Marvel Super Heroes, #4455):
//
//	"Seismic Takedown — Whenever you cast a noncreature spell, tap target
//	 creature or land."
//
// Black Waltz No. 3's noncreature-cast condition with Devoted Grafkeeper's
// targeted tap: the target is picked when the trigger goes on the stack and
// is re-checked as it resolves (CR 608.2b). "Seismic Takedown" is an
// ability word with no rules meaning.
//
// No simplification.
func init() {
	tap := effects.WheneverYouCast(effects.Noncreature(), "Quake — tap target creature or land", effects.B36TapChosenCreature)
	tap.Targets = effects.TargetPermanent("target creature or land", effects.Or(effects.Creature(), effects.Land()))
	effects.Register(effects.Spec{
		OracleID:     "3966c94b-e597-4b3f-a4f5-001d52fff9fe",
		Name:         "Quake, Agent of S.H.I.E.L.D.",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{tap},
	})
}
