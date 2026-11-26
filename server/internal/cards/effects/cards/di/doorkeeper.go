package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doorkeeper — Creature — Homunculus {1}{U}, 0/4 (Return to Ravnica, #5566):
//
//	"Defender
//	 {2}{U}, {T}: Target player mills X cards, where X is the number of creatures
//	 you control with defender."
//
// X is counted as the ability resolves (CR 608.2h), through the layered view, so
// the Doorkeeper itself counts while it is on the battlefield and a creature that
// lost defender in response does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8a960cc2-3802-42bc-ba28-941bfe9743be",
		Name:            "Doorkeeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Target player mills X cards, where X is the number of creatures you control with defender.",
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				x := 0
				match := effects.And(effects.Creature(), effects.YouControl(), effects.HasKeyword("defender"))
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && match(g, ctx.Controller(), c) {
						x++
					}
				}
				return effects.MillCards{Player: player, N: x}.Apply(ctx)
			},
		}},
	})
}
