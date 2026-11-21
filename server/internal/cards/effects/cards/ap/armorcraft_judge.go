package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armorcraft Judge — Creature — Elf Artificer {3}{G}, 3/3 (EDHREC
// rank 3575):
//
//	"When this creature enters, draw a card for each creature you
//	 control with a +1/+1 counter on it."
//
// The counters deck's draw. Counted as the trigger resolves, as
// printed: a creature that picked up a counter in response counts,
// one that lost its last counter does not. The Judge itself enters
// with no counter and does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7f49243-a96e-499f-b2d7-8e9842432420",
		Name:         "Armorcraft Judge",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Armorcraft Judge — draw a card for each creature you control with a +1/+1 counter", effects.B34DrawPerCreatureYouControlMatching(effects.B34HasPlusCounter)),
		},
	})
}
