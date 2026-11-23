package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Floodgate — Creature — Wall {3}{U}, 0/5 (Dominaria Remastered, #5948):
//
//	"Defender
//	 When this creature has flying, sacrifice it.
//	 When this creature leaves the battlefield, it deals damage to each
//	 nonblue creature without flying equal to half the number of Islands
//	 you control, rounded down."
//
// The first is a state trigger (CR 603.8). The second counts the Islands
// its controller controls as it resolves, with the Floodgate's last-known
// controller (CR 608.2h), and deals that damage from the Floodgate to each
// creature that is not blue and has no flying, read through the layers.
// Nothing happens for fewer than two Islands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f5a85674-fd09-42d8-bba7-5656be25fcea",
		Name:            "Floodgate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenState("Floodgate — sacrifice it while it has flying",
				func(_ *game.Game, source *game.Card, _ uuid.UUID) bool {
					return game.HasKeyword(source, "flying")
				},
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
				}))),
			effects.WhenThisLeaves("Floodgate — deal damage to each nonblue creature without flying equal to half the number of Islands you control, rounded down",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					islands := ctx.TextSubtype("Island")
					n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype(islands) }) / 2
					if n <= 0 {
						return nil
					}
					blue := ctx.TextColor("U")
					return ctx.Game.DamageInstanceForEffect(func() error {
						for _, c := range ctx.Game.BattlefieldCardsForEffect() {
							if !c.IsCreature() || c.HasColor(blue) || game.HasKeyword(&c, "flying") {
								continue
							}
							if err := (effects.DealDamage{Source: ctx.Item.SourceCardID, Target: c.InstanceID, Amount: n}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				}))),
		},
	})
}
