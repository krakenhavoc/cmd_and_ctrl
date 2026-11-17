package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Samite Alchemist — Creature — Human Cleric {3}{W}, 0/2 (Homelands,
// #5753):
//
//	"{W}{W}, {T}: Prevent the next 4 damage that would be dealt this turn
//	 to target creature you control. Tap that creature. It doesn't untap
//	 during your next untap step."
//
// The shield is the charged one (PreventNextDamage, CR 615.7), pinned to
// the creature. Then TapAndFreeze taps it and gives it ADR 0058's one-shot
// mark, keyed to the ACTIVATOR's untap step ("your", UntapSkip.Player):
// if another player gains control of the creature, it untaps normally
// during that player's untap step. The mark is used up by the activator's
// next untap step whether or not the creature is still tapped, and it
// goes away if the creature leaves the battlefield (CR 400.7).
//
// A creature that is no longer a legal target as the ability resolves
// (it left, or its controller changed) gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "120b565a-e22b-4d5d-9841-b11750b4ee49",
		Name:         "Samite Alchemist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}{W}, {T}: Prevent the next 4 damage that would be dealt this turn to target creature you control. Tap that creature. It doesn't untap during your next untap step.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    Plus(ManaCost("{W}{W}"), TapCost()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (PreventNextDamage{Target: id, Amount: 4, Label: "Samite Alchemist — prevent the next 4 damage"}).Apply(ctx); err != nil {
					return err
				}
				return TapAndFreeze{
					Targets: []uuid.UUID{id},
					Player:  ctx.Controller(),
					Label:   "Samite Alchemist — doesn't untap during your next untap step",
				}.Apply(ctx)
			},
		}},
	})
}
