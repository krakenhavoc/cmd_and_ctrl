package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ironclad Slayer — Creature — Human Warrior {2}{W}, 3/2 (Commander Legends):
//
//	"When this creature enters, you may return target Aura or Equipment card
//	 from your graveyard to your hand."
//
// An entry trigger with a graveyard target chosen as it goes on the stack
// (re-checked on resolution, CR 608.2b) and a "you may" asked on
// resolution (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "964e20e2-2947-4d8a-a0a2-b8b5fde27ff8",
		Name:         "Ironclad Slayer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Ironclad Slayer — you may return target Aura or Equipment card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.TargetCardInGraveyard("target Aura or Equipment card from your graveyard", effects.Or(effects.OfSubtype("Aura"), effects.OfSubtype("Equipment")), effects.YouOwn())),
				"Ironclad Slayer — return the target Aura or Equipment card to your hand?"),
		},
	})
}
