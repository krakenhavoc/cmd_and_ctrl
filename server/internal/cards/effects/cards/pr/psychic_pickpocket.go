package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Pickpocket — Creature — Octopus Rogue {4}{U}, 3/2 (Streets of
// New Capenna, #5491):
//
//	"When this creature enters, it connives. When it connives this way,
//	 return up to one target nonland permanent to its owner's hand. (To
//	 have a creature connive, draw a card, then discard a card. If you
//	 discarded a nonland card, put a +1/+1 counter on that creature.)"
//
// One printed enters trigger whose connive is followed by a reflexive
// trigger (CR 603.12, conniveThenReflexive): the permanent is chosen as that
// trigger goes on the stack, after the draw and discard, and it may be no
// target at all. Any player's nonland permanent qualifies, the Pickpocket
// itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60ead8a5-11fb-4eeb-99bf-0caaf1de1731",
		Name:         "Psychic Pickpocket",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Psychic Pickpocket — it connives", effects.PsychicPickpocketConnive),
		},
	})
}
