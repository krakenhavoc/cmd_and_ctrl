package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quick Draw — Instant {R} (Outlaws of Thunder Junction):
//
//	"Target creature you control gets +1/+1 and gains first strike until
//	 end of turn. Creatures target opponent controls lose first strike and
//	 double strike until end of turn."
//
// Two clauses: your creature, then an opponent. The opponent's creatures
// are the ones that player controls as the spell resolves (CR 611.2c), one
// layer-6 removal over that set (a plain "lose", not "can't have").
// If the opponent is gone the second half does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae0d17f6-859c-4b41-8b66-b68abb837bbb",
		Name:         "Quick Draw",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "first strike")),
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetPlayer("target opponent", Opponent()),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && ctx.IsTargetLegal(t) {
				if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Quick Draw — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"}, Label: "Quick Draw — first strike"}).Apply(ctx); err != nil {
					return err
				}
			}
			t, ok := ctx.ClauseTarget(1)
			if !ok || !ctx.IsTargetLegal(t) {
				return nil
			}
			opp := t.ID
			return ScopedEffectFor{
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.IsCreature() && c.Controller == opp
				},
				Mods:     []game.Mod{game.RemoveKeywordsMod("first strike", "double strike")},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Quick Draw — lose first strike and double strike",
			}.Apply(ctx)
		},
	})
}
