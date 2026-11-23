package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rydia's Return — Sorcery {3}{G}{G} (Final Fantasy, #4462):
//
//	"Choose one —
//	 • Creatures you control get +3/+3 until end of turn.
//	 • Return up to two target permanent cards from your graveyard to your
//	   hand."
//
// Mode 0 is Overrun's pump over the creatures you control as it resolves
// (CR 611.2c). Mode 1 returns each target still legal as it resolves
// (CR 608.2b); "permanent card" is any card with a permanent type, a
// predicate over the graveyard card's printed types.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d007479-c5e8-4aa1-b53e-fa962783deee",
		Name:         "Rydia's Return",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Creatures you control get +3/+3 until end of turn."),
			effects.Mode("Return up to two target permanent cards from your graveyard to your hand.",
				effects.TargetCardInGraveyard("up to two target permanent cards from your graveyard",
					effects.YouOwn(), effects.Permanent()).WithCount(0, 2)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 3, Toughness: 3,
					Label: "Rydia's Return — +3/+3",
				}.Apply(ctx)
			}
			for _, t := range ctx.ModeTargets(0) {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
