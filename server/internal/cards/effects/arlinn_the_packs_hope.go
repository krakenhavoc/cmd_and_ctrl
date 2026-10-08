package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// DECLARED SIMPLIFICATION, weaker than printed: the +1 gives only the
// flash half. A "creatures you control enter with an additional +1/+1
// counter until your next turn" effect needs a replacement that lasts
// for a duration, and the engine has only the ones a permanent carries
// while it is in play. The flash half is the same timing grant Teferi,
// Time Raveler uses for sorceries, filtered to creature spells.
func init() {
	const oracle = "f227ce07-7e96-4a36-ab7c-9be6e777d649"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Arlinn, the Pack's Hope",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The +1 lets you cast creature spells as though they had flash, but your creatures don't enter with an additional +1/+1 counter."},
		PrintedKeywords: []string{"daybound"},
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Until your next turn, you may cast creature spells as though they had flash, and each creature you control enters with an additional +1/+1 counter on it.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GrantCastTiming{
						Filter:            game.PermissionFilter{CreatureOnly: true},
						UntilYourNextTurn: true,
						Label:             "Until your next turn, you may cast creature spells as though they had flash.",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−3: Create two 2/2 green Wolf creature tokens.",
				Cost:  LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: TokenCard("2/2 green Wolf"), N: 2}.Apply(NewContext(g, item))
				},
			},
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Arlinn, the Moon's Fury",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label:  "+2: Add {R}{G}.",
				Cost:   LoyaltyCost(2),
				Effect: Do(AddMana{Produced: "{R}{G}"}),
			},
			{
				Label: "0: Until end of turn, Arlinn becomes a 5/5 Werewolf creature with trample, indestructible, and haste.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					ctx := NewContext(g, item)
					return ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: []game.Mod{
							game.AddTypesMod("Creature"),
							game.AddSubtypesMod("Werewolf"),
							game.SetBasePowerMod(5),
							game.SetBaseToughnessMod(5),
							game.AddKeywordsMod("trample", "indestructible", "haste"),
						},
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Arlinn, the Moon's Fury — becomes a 5/5 Werewolf creature",
					}.Apply(ctx)
				},
			},
		},
	})
}
