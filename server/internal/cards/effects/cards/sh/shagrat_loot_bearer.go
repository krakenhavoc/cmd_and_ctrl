package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shagrat, Loot Bearer — Legendary Creature — Orc Soldier {2}{B}{R}, 4/4 (The
// Lord of the Rings: Tales of Middle-earth):
//
//	"Whenever Shagrat attacks, attach up to one target Equipment to it.
//	 Then amass Orcs X, where X is the number of Equipment attached to
//	 Shagrat."
//
// Cloud, Ex-SOLDIER's "attach up to one target Equipment to this" (any
// Equipment, whoever controls it: control does not change), then the amass
// counted as the attach has finished, over every Equipment attached to
// Shagrat as it resolves, the one just attached included. A Shagrat that
// has left the battlefield attaches nothing (CR 701.3b) and counts none
// (the amass is then amass Orcs 0, which still finds or makes the Army,
// CR 701.47b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23db92aa-d4db-49e3-a5bc-7b120e02822f",
		Name:         "Shagrat, Loot Bearer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Shagrat, Loot Bearer — attach up to one target Equipment to Shagrat, then amass Orcs X",
					effects.Do(
						effects.Instruction(func(ctx *effects.Context) error { return effects.AttachTargetsToSource(ctx.Game, ctx.Item) }),
						effects.Instruction(func(ctx *effects.Context) error {
							x := 0
							for _, c := range ctx.Game.BattlefieldCardsForEffect() {
								if c.HasSubtype("Equipment") && c.IsAttachedTo(ctx.Source()) {
									x++
								}
							}
							return effects.Amass{Subtype: "Orc", N: x}.Apply(ctx)
						}),
					)),
				effects.TargetPermanent("up to one target Equipment", effects.HasSubtype("Equipment")).WithCount(0, 1)),
		},
	})
}
