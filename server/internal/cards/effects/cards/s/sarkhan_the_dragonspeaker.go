package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarkhan, the Dragonspeaker — Legendary Planeswalker {3}{R}{R}, loyalty 4
// (Jumpstart 2022, #5948):
//
//	"+1: Until end of turn, Sarkhan becomes a legendary 4/4 red Dragon
//	 creature with flying, indestructible, and haste. (He doesn't lose
//	 loyalty while he's not a planeswalker.)
//	 −3: Sarkhan deals 4 damage to target creature.
//	 −6: You get an emblem with "At the beginning of your draw step, draw
//	 two additional cards" and "At the beginning of your end step, discard
//	 your hand.""
//
// The +1 replaces his types with legendary creature, so he is a 4/4 red
// Dragon creature and not a planeswalker until end of turn (CR 205.1a);
// the emblem is Sylvan Library's draw-step trigger and an end-step
// discard of the whole hand, both from the command zone (CR 114.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "80c26f58-04b0-4db0-b84e-9b136609abe0",
		Name:            "Sarkhan, the Dragonspeaker",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Sarkhan, the Dragonspeaker emblem",
			Text:  "At the beginning of your draw step, draw two additional cards. At the beginning of your end step, discard your hand.",
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventBeginDrawStep, effects.ByYou, "Sarkhan, the Dragonspeaker emblem — draw two additional cards",
					effects.Do(effects.DrawCards{N: 2})),
				effects.AtYourEndStep("Sarkhan, the Dragonspeaker emblem — discard your hand",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
						return err
					}))),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Until end of turn, Sarkhan becomes a legendary 4/4 red Dragon creature with flying, indestructible, and haste.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: []game.Mod{
							game.SetTypesMod([]string{"Creature"}, ctx.TextSubtypes("Dragon"), "Legendary"),
							game.SetColorsMod(ctx.TextColor("R")),
							game.SetBasePowerMod(4),
							game.SetBaseToughnessMod(4),
							game.AddKeywordsMod("flying", "indestructible", "haste"),
						},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Sarkhan, the Dragonspeaker — becomes a legendary 4/4 red Dragon creature",
					}.Apply(ctx)
				},
			},
			{
				Label:   "−3: Sarkhan deals 4 damage to target creature.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
				Effect:  effects.DamageToFirstLegalTarget(4),
			},
			{
				Label: "−6: You get an emblem with \"At the beginning of your draw step, draw two additional cards\" and \"At the beginning of your end step, discard your hand.\"",
				Cost:  effects.LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
