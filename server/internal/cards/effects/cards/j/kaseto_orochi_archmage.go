package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kaseto, Orochi Archmage — Legendary Creature — Snake Wizard {1}{G}{U},
// 2/2 (Commander 2015, #5546):
//
//	"{G}{U}: Target creature can't be blocked this turn. If that creature is
//	 a Snake, it gets +2/+2 until end of turn."
//
// The evasion is Zhuge Jin's targetCantBeBlockedThisTurn. The Snake check
// is Ingenious Leonin's: read off the target as it is when the ability
// resolves, after the layers (a changeling is a Snake), and the word is
// read through Kaseto's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "198b6c6f-a480-4672-9d87-b0cb5d8ed1c0",
		Name:         "Kaseto, Orochi Archmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{U}: Target creature can't be blocked this turn. If that creature is a Snake, it gets +2/+2 until end of turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{G}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.TargetCantBeBlockedThisTurn("Kaseto, Orochi Archmage")(ctx.Game, ctx.Item)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok || !effects.OnBattlefield(ctx.Game, id) {
						return nil
					}
					ctx.Game.RecomputeLayersIfStaleLocked()
					c, ok := ctx.Game.LookupCardForEffect(id)
					if !ok || !c.HasSubtype(ctx.TextSubtype("Snake")) {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Kaseto, Orochi Archmage — +2/+2"}.Apply(ctx)
				}),
			),
		}},
	})
}
