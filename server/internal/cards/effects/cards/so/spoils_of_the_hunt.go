package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spoils of the Hunt — Instant {2}{G} (Adventures in the Forgotten Realms):
//
//	"Target creature you control gets +1/+0 until end of turn for each mana
//	 from a Treasure that was spent to cast this spell. Then that creature
//	 deals damage equal to its power to target creature an opponent controls."
//
// The Treasure mana is counted from the spell's payment record
// (ManaSpent().CountFrom, one per mana, snapshotted when each was made), so a
// Treasure sacrificed to pay still counts. The pump comes first; the bite then
// reads the pumped power (Bite Down's oneSidedBite). Either target gone is
// the other half still happening (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "c29713ee-57e9-40b4-8ca3-a4cc6ede0f46",
		Name:           "Spoils of the Hunt",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0))},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if biter, ok := ctx.ClauseTarget(0); ok && biter.Kind == game.TargetCard {
				if n := ctx.ManaSpent().CountFrom(game.ManaSourceTreasure); n > 0 {
					if err := (effects.BoostUntilEOT{Target: biter.ID, Power: n, Label: "Spoils of the Hunt — +1/+0 for each mana from a Treasure"}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return effects.OneSidedBite(item, ctx)
		},
	})
}
