package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teleportal — Sorcery {U}{R} (Return to Ravnica, #5566):
//
//	"Target creature you control gets +1/+0 until end of turn and can't be blocked
//	 this turn.
//	 Overload {3}{U}{R} (You may cast this spell for its overload cost. If you do,
//	 change "target" in its text to "each.")"
//
// Overloaded it names no target (CR 702.96b). The +1/+0 goes to each creature you
// control at resolution (CR 611.2c). "Can't be blocked this turn" is not a
// characteristic, so it covers every creature you control this turn, ones that
// arrive later included (Keeper of Keys' reading, #1650).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5b4876f-d28d-40dc-be7f-2c4f4c0bd6b3",
		Name:         "Teleportal",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 1, 0)),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		AlternativeCosts: []game.AlternativeCost{
			effects.CostWithPurpose(effects.Overload("{3}{U}{R}"), game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant}),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, match, ok := effects.OldB424Subject(ctx, effects.YouControl())
			if !ok {
				return nil
			}
			if err := effects.UntilEndOfTurn(ctx, id, match, "Teleportal — +1/+0", game.ModifyPTMod(1, 0)); err != nil {
				return err
			}
			if match != nil {
				return effects.RestrictUntilEOT{Scope: game.ScopeYourCreatures, Restrictions: game.CantBeBlocked, Label: "Teleportal — creatures you control can't be blocked"}.Apply(ctx)
			}
			return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Teleportal — can't be blocked"}.Apply(ctx)
		},
	})
}
