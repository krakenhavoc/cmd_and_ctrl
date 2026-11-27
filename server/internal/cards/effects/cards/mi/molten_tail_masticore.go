package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Molten-Tail Masticore — Artifact Creature — Masticore {4}, 4/4 (Scars of
// Mirrodin, S82 OLD batch 472, #5577):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you
//	 discard a card.
//	 {4}, Exile a creature card from your graveyard: This creature deals 4
//	 damage to any target.
//	 {2}: Regenerate this creature."
//
// Argentum Masticore's upkeep discard (OldB472UpkeepSacrificeUnlessDiscard),
// Moorland Haunt's exile-a-creature-card cost (paid as it is activated, CR
// 602.2b), and the Regenerate row Birthing Hulk uses. A target that is gone by
// resolution takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15fb3a7a-a6c3-4bef-9f72-1d04b623063e",
		Name:         "Molten-Tail Masticore",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Molten-Tail Masticore — sacrifice this creature unless you discard a card",
				effects.OldB472UpkeepSacrificeUnlessDiscard("Molten-Tail Masticore")),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{4}, Exile a creature card from your graveyard: This creature deals 4 damage to any target.",
				Cost:    effects.Plus(effects.ManaCost("{4}"), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
				Effect:  effects.B33DamageChosenTargetFromSource(4),
			},
			effects.LrwB1RegenerateThisRow("{2}"),
		},
	})
}
