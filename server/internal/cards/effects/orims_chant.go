package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orim's Chant — Instant {W} (Modern Horizons 3, #5948):
//
//	"Kicker {W} (You may pay an additional {W} as you cast this spell.)
//	 Target player can't cast spells this turn. If this spell was kicked,
//	 creatures can't attack this turn."
//
// Silence's outright cast ban, granted to the TARGET player until end of
// turn (CR 611.2; a spell already on the stack is unaffected). The kicked
// sentence is a live rule over every creature (game.ScopeCreatures), so a
// creature that enters later this turn can't attack either. Kicker is read
// back with ctx.WasKicked (CR 702.33). If the target player is gone the
// spell does nothing (CR 608.2b), the kicked sentence included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "26d3a818-7faf-443a-a6bd-9773edaedf17",
		Name:          "Orim's Chant",
		Completeness:  CompletenessFull,
		Targets:       TargetPlayer("target player"),
		OptionalCosts: []game.AdditionalCost{Kicker("{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			d := DurationUntilEndOfTurn(ctx)
			if err := (RestrictCasting{
				Player:   t.ID,
				Rule:     game.CastBanRule{Kind: game.CastBanOutright},
				Label:    "Orim's Chant — can't cast spells this turn",
				Duration: d,
			}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeCreatures, ctx.Controller(),
				[]game.Mod{game.AddRestrictionsMod(game.CantAttack)}, d,
				"Orim's Chant — creatures can't attack this turn")
			return nil
		},
	})
}
