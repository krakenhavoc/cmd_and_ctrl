package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonhold — Instant {2}{R/W}:
//
//	"Target player can't play lands this turn if {R} was spent to cast
//	 this spell and can't cast creature spells this turn if {W} was
//	 spent to cast this spell. (Do both if {R}{W} was spent.)"
//
// ADR 0109 §4 (#1895). Each half reads the colours of mana the cast
// recorded (Context.ManaSpentOfColor, #761): red buys the land ban
// (CantPlayLandsThisTurn, a stored ModCantPlayLands record), white buys
// a granted creature-spell cast ban for the turn (#1316's
// CastBanOutright with a creature filter), and {R}{W} spent does both.
//
// WantsColors names both halves' colours (#3853), so an auto-paid cast
// spends a red and a white source on the hybrid symbol and the generic
// half when the board has them: the {R}{W} "do both" case is what auto-pay
// reaches for, not a thing the player taps by hand.
//
// A cast for no mana (a free cast) spends nothing, so neither half
// applies.
// A waived payment (strict mana off with too little in the pool, or Cast
// anyway) asks the caster what they paid outside the pool, because the
// card declares ReadsManaSpent, so the colours are exact (#3855, ADR 0068
// amendment 2026-10-15).
func init() {
	Register(Spec{
		OracleID:       "9b515bb8-6e37-4b2d-8d56-bb517c3b267c",
		Name:           "Moonhold",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   CompletenessFull,
		WantsColors:    []string{"R", "W"},
		Targets:        TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.ManaSpentOfColor("R") > 0 {
				if err := (CantPlayLandsThisTurn{}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.ManaSpentOfColor("W") > 0 {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if err := (RestrictCasting{
						Player:   t.ID,
						Rule:     game.CastBanRule{Kind: game.CastBanOutright, Filter: game.PermissionFilter{CreatureOnly: true}},
						Label:    "Moonhold — can't cast creature spells this turn",
						Duration: DurationUntilEndOfTurn(ctx),
					}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
