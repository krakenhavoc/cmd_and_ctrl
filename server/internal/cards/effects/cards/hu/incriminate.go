package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incriminate — Sorcery {1}{B} (Streets of New Capenna, S82):
//
//	"Choose two target creatures controlled by the same player. That
//	 player sacrifices one of them of their choice."
//
// Two target clauses: any creature, then another creature controlled by
// the same player (ControlledByTargetsController, Breaking of the
// Fellowship's relation), judged as the spell is cast (CR 601.2c) and again
// as it resolves (CR 608.2b). The target's controller, not the caster,
// picks which of the creatures they sacrifice (SacrificeChoice), and the
// sacrifice is not a destroy, so indestructible and protection don't save
// it. A target that became illegal is skipped (CR 608.2b): when only one
// remains the controller sacrifices it. When both are illegal the spell
// does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8b59cc3-766f-4350-971f-edf721fcc5be",
		Name:         "Incriminate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.ControlledByTargetsController(effects.TargetCreature("another target creature controlled by the same player"), 0)),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var candidates []uuid.UUID
			var player uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || !ctx.Game.Battlefield.Contains(t.ID) || ctx.Game.CantBeSacrificedForEffect(t.ID) {
					continue
				}
				player = c.Controller
				candidates = append(candidates, t.ID)
			}
			if len(candidates) == 0 {
				return nil
			}
			return effects.SacrificeChoice{
				Player:     player,
				Candidates: candidates,
				Question:   "Incriminate — sacrifice one of the chosen creatures",
			}.Apply(ctx)
		},
	})
}
