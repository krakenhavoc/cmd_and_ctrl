package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana, the Last Hope — Legendary Planeswalker — Liliana {1}{B}{B}, loyalty 3 (S82 GAP, #5948):
//
//	"+1: Up to one target creature gets -2/-1 until your next turn.
//	 −2: Mill two cards, then you may return a creature card from your
//	 graveyard to your hand.
//	 −7: You get an emblem with "At the beginning of your end step, create X
//	 2/2 black Zombie creature tokens, where X is two plus the number of
//	 Zombies you control.""
//
// Mu Yanling's until-your-next-turn shrink, The Mending of Dominaria's
// mill-then-return, and Basri Ket's emblem. The emblem counts the Zombies its
// controller controls as the trigger resolves (CR 608.2h).
func init() {
	const emblemText = "At the beginning of your end step, create X 2/2 black Zombie creature tokens, where X is two plus the number of Zombies you control."
	effects.Register(effects.Spec{
		OracleID:        "64a9a6b4-26b5-4d23-be86-85c54e9c4727",
		Name:            "Liliana, the Last Hope",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Liliana, the Last Hope emblem",
			Text:  emblemText,
			Triggered: []game.TriggeredAbility{
				effects.AtYourEndStep("Liliana, the Last Hope emblem — create X 2/2 black Zombie creature tokens",
					effects.Do(effects.Instruction(lilianaLastHopeZombies))),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Up to one target creature gets -2/-1 until your next turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.ScopedEffectFor{
						Target:   t.ID,
						Mods:     []game.Mod{game.ModifyPTMod(-2, -1)},
						Duration: effects.DurationUntilYourNextTurn(ctx, item.Controller),
						Label:    "Liliana, the Last Hope — -2/-1 until your next turn",
					}.Apply(ctx)
				},
			},
			{
				Label: "−2: Mill two cards, then you may return a creature card from your graveyard to your hand.",
				Cost:  effects.LoyaltyCost(-2),
				Effect: effects.Do(effects.MillCards{N: 2}, effects.Instruction(func(ctx *effects.Context) error {
					return effects.ReturnChosenFromGraveyard{
						Player:   ctx.Controller(),
						Question: "Liliana, the Last Hope — you may return a creature card from your graveyard to your hand",
						Match:    game.Card.IsCreature,
						Max:      1,
						ToHand:   true,
					}.Apply(ctx)
				})),
			},
			{
				Label:  "−7: You get an emblem with \"" + emblemText + "\"",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// lilianaLastHopeZombies is the emblem's trigger: two plus the Zombies its
// controller controls, each a 2/2 black Zombie.
func lilianaLastHopeZombies(ctx *effects.Context) error {
	n := len(effects.MatchingBattlefield(ctx, effects.And(effects.YouControl(), effects.Subtype("Zombie"))))
	return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 2 + n}.Apply(ctx)
}
