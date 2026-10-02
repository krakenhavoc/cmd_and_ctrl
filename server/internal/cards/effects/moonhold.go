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
// A payment the engine did not record (a free cast, a waived charge)
// spends no mana it knows of, so neither half applies: weaker than
// printed, never stronger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b515bb8-6e37-4b2d-8d56-bb517c3b267c",
		Name:         "Moonhold",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
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
