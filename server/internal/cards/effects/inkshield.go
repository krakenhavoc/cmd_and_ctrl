package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inkshield — Instant {3}{W}{B}:
//
//	"Prevent all combat damage that would be dealt to you this turn. For each 1 damage prevented this way, create a 2/1 white and black Inkling creature token with flying."
//
// ADR 0108 owner decision 2 (#1906): Fog narrowed to you (Druid's
// Deliverance's shape) with a CR 615.5 additional effect, owed once per
// damage instance with the instance's total: a combat damage step's
// attackers are one batch of Inklings. Combat damage that can't be
// prevented is dealt and makes none (CR 615.12).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "da93264e-4e04-401a-9809-e4e1056cf604",
		Name:         "Inkshield",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PreventAllCombatDamageThisTurn{
				Player: ctx.Controller(),
				Then:   inkshieldInklingsBody,
				Label:  "Inkshield — prevent combat damage to you and make Inklings",
			}.Apply(ctx)
		},
	})
}

// The additional effect: an Inkling for each 1 damage prevented.
var inkshieldInklingsBody = game.DelayedBody("inkshield/inklings", inkshieldInklings)

func inkshieldInklings(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	return CreateToken{Controller: item.Controller, Template: TokenCard("2/1 white and black Inkling with flying"), N: p.Amount}.Apply(NewContext(g, item))
}
