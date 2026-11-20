package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tanufel Rimespeaker — Creature — Elemental Wizard {3}{U}, 2/4 (Lorwyn
// Eclipsed, #4458):
//
//	"Whenever you cast a spell with mana value 4 or greater, draw a card."
//
// Angry Rabble's cast trigger: the spell's mana value is read off the
// stack, so {X} counts at its chosen value (CR 202.3e). Each such cast is
// its own trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf8de96b-0a75-4459-a123-dcf95e1b89f2",
		Name:         "Tanufel Rimespeaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ManaValueGE(4), "Tanufel Rimespeaker — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
