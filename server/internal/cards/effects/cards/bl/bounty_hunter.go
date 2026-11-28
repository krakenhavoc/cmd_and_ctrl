package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bounty Hunter — Creature — Human Archer Minion {2}{B}{B}, 2/2 (Tempest,
// S82 OLD batch 759, #5628):
//
//	"{T}: Put a bounty counter on target nonblack creature.
//	 {T}: Destroy target creature with a bounty counter on it."
//
// Two tap abilities with their own targets, each checked on activation and
// again as it resolves (CR 608.2b): "nonblack" is read through the layered
// colours, and the counter is a plain named counter, so a creature that
// loses it before the second ability resolves is no longer a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fd1cfc5-fbcd-4d31-ba8b-759d8c41b435",
		Name:         "Bounty Hunter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Put a bounty counter on target nonblack creature.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target nonblack creature", effects.NonBlack()),
				Effect:  effects.S82OldB759PutABountyCounter,
			},
			{
				Label:   "{T}: Destroy target creature with a bounty counter on it.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature with a bounty counter on it",
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return effects.HasCounterOfKind(c, "bounty") }),
				Effect: effects.DestroyChosenPermanent,
			},
		},
	})
}
