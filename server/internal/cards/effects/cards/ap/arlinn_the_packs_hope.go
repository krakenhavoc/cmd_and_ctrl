package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arlinn, the Pack's Hope // Arlinn, the Moon's Fury — {2}{R}{G}
// Legendary Planeswalker — Arlinn, loyalty 4 each side (#2586, ADR
// 0132):
//
//	Front: "Daybound
//	        +1: Until your next turn, you may cast creature spells as
//	        though they had flash, and each creature you control enters
//	        with an additional +1/+1 counter on it.
//	        −3: Create two 2/2 green Wolf creature tokens."
//	Back:  "Nightbound
//	        +2: Add {R}{G}.
//	        0: Until end of turn, Arlinn becomes a 5/5 Werewolf creature
//	        with trample, indestructible, and haste."
//
// The permanent turns over like any daybound permanent and keeps its
// loyalty counters (CR 712.18). The back face's 0 is Gideon, Ally of
// Zendikar's animation: a scoped effect with a type, a subtype, a base
// power and toughness and three keywords, ending with the turn. It stays
// a planeswalker, so it can still be attacked and takes damage as one
// (CR 306.8), which is a printed drawback and not a gap.
//
// The +1's second half (#3172) is a PlayerStatic grant with an until-your-
// next-turn duration (game.GrantCreaturesEnterWithCountersForEffect) read by
// a built-in entry replacement, so every creature entering under the
// controller until then, tokens included, gets the counter through the
// counter pipeline. The flash half is the same timing grant Teferi, Time
// Raveler uses for sorceries, filtered to creature spells.
func init() {
	const oracle = "f227ce07-7e96-4a36-ab7c-9be6e777d649"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Arlinn, the Pack's Hope",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Until your next turn, you may cast creature spells as though they had flash, and each creature you control enters with an additional +1/+1 counter on it.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GrantCastTiming{
						Filter:            game.PermissionFilter{CreatureOnly: true},
						UntilYourNextTurn: true,
						Label:             "Until your next turn, you may cast creature spells as though they had flash.",
					}).Apply(ctx); err != nil {
						return err
					}
					g.GrantCreaturesEnterWithCountersForEffect(item.Controller, game.CounterPlusOne, 1,
						"Arlinn, the Pack's Hope: each creature you control enters with an additional +1/+1 counter",
						item.SourceCardID, effects.DurationUntilYourNextTurn(ctx, item.Controller))
					return nil
				},
			},
			{
				Label: "−3: Create two 2/2 green Wolf creature tokens.",
				Cost:  effects.LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 2}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Arlinn, the Moon's Fury",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+2: Add {R}{G}.",
				Cost:   effects.LoyaltyCost(2),
				Effect: effects.Do(effects.AddMana{Produced: "{R}{G}"}),
			},
			{
				Label: "0: Until end of turn, Arlinn becomes a 5/5 Werewolf creature with trample, indestructible, and haste.",
				Cost:  effects.LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: []game.Mod{
							game.AddTypesMod("Creature"),
							game.AddSubtypesMod(ctx.TextSubtypes("Werewolf")...),
							game.SetBasePowerMod(5),
							game.SetBaseToughnessMod(5),
							game.AddKeywordsMod("trample", "indestructible", "haste"),
						},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Arlinn, the Moon's Fury — becomes a 5/5 Werewolf creature",
					}.Apply(ctx)
				},
			},
		},
	})
}
