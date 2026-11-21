package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hoverstone Pilgrim — Artifact Creature — Golem {5}, 2/5 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Flying
//	 Ward {2}
//	 {2}: Put target card from a graveyard on the bottom of its owner's
//	 library."
//
// Ward is the table's ward trigger (CR 702.21a). The activated row is
// Chrome Companion's without the {T}: the card goes to the bottom of its
// OWNER's library, whichever graveyard it was in, and a card that has
// left the graveyard by resolution stays where it is (CR 608.2b). Every
// use has a target, so the row declares no Answers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "64399663-2c1e-447e-a3ed-c49ebd9f2231",
		Name:            "Hoverstone Pilgrim",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Hoverstone Pilgrim — ward {2}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.SocB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
