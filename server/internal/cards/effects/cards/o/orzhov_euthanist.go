package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orzhov Euthanist — Creature — Human Assassin {2}{B}, 2/2 (Guildpact):
//
//	"Haunt (When this creature dies, exile it haunting target creature.)
//	 When this creature enters or the creature it haunts dies, destroy
//	 target creature that was dealt damage this turn."
//
// Haunt is CR 702.55 (#6083): the haunt ability exiles the card linked to
// the creature it targeted, and the second ability, one ability with two
// conditions, triggers on the card's entry and, from exile, on that
// creature's death (CR 702.55c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3cc82a02-2906-4411-aebf-474b0735bb53",
		Name:         "Orzhov Euthanist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Haunt("Orzhov Euthanist"),
			effects.Targeting(
				effects.WhenThisEntersOrTheCreatureItHauntsDies("Orzhov Euthanist — destroy target creature that was dealt damage this turn", effects.DestroyChosenPermanent),
				effects.TargetCreature("target creature that was dealt damage this turn", effects.DealtDamageThisTurn()),
			),
		},
	})
}
