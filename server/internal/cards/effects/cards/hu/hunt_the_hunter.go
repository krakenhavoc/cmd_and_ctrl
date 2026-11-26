package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunt the Hunter — Sorcery {G} (Theros, S82 OLD b389, #5561):
//
//	"Target green creature you control gets +2/+2 until end of turn. It
//	 fights target green creature an opponent controls."
//
// Wild Instincts' two clauses with a green requirement on each pick: the
// pump comes first so the fight reads the bigger power (CR 701.14), and
// when the opponent's creature is gone the pump still applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00afc724-be09-4e92-a50c-88011bd10ccd",
		Name:         "Hunt the Hunter",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target green creature you control", effects.OfColor("G"), effects.YouControl()),
			effects.TargetCreature("target green creature an opponent controls", effects.OfColor("G"), effects.OpponentControls()),
		),
		Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: mine.ID, Power: 2, Toughness: 2, Label: "Hunt the Hunter — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			return effects.B10Fight(ctx, mine.ID, theirs.ID)
		},
	})
}
