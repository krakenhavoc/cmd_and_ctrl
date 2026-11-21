package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Domri Rade — Legendary Planeswalker — Domri {1}{R}{G}, loyalty 3
// (Ravnica Remastered):
//
//	"+1: Look at the top card of your library. If it's a creature card, you
//	 may reveal it and put it into your hand.
//	 −2: Target creature you control fights another target creature.
//	 −7: You get an emblem with "Creatures you control have double strike,
//	 trample, hexproof, and haste.""
//
// The +1 is a one-card look with the pick Adventure Awaits registers (the
// card stays on top unless taken). The −2's second clause is Distinct: "another
// target creature" is any creature other than the fighter (CR 601.2c). The −7 emblem grants four keywords to the emblem
// owner's creatures, read through the layers as they change.
//
// No simplification.
func init() {
	creatures := creaturesTheEmblemsOwnerControls
	Register(Spec{
		OracleID:        "0c4a2c42-0874-40b4-8d12-47eeadc3e976",
		Name:            "Domri Rade",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label: "Domri Rade emblem",
			Text:  "Creatures you control have double strike, trample, hexproof, and haste.",
			Static: []game.StaticAbility{
				KeywordGrant(creatures, "double strike"),
				KeywordGrant(creatures, "trample"),
				KeywordGrant(creatures, "hexproof"),
				KeywordGrant(creatures, "haste"),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Look at the top card of your library. If it's a creature card, you may reveal it and put it into your hand.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 1),
						Match:    Creature(),
						Max:      1,
						Optional: true,
						Label:    "Domri Rade — you may reveal the creature card and put it into your hand",
						Key:      domriRadeTake,
					}.Apply(ctx)
				},
			},
			{
				Label: "−2: Target creature you control fights another target creature.",
				Cost:  LoyaltyCost(-2),
				Targets: Clauses(
					TargetCreature("target creature you control", YouControl()),
					Distinct(TargetCreature("another target creature")),
				),
				Effect: domriFight,
			},
			{
				Label:  "−7: You get an emblem with \"Creatures you control have double strike, trample, hexproof, and haste.\"",
				Cost:   LoyaltyCost(-7),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}

// domriRadeTake is the +1's pick: the revealed card goes to hand, and an
// untaken card stays where it is, on top.
var domriRadeTake = RegisterTake("choose-cards/domri-rade-take", TakeDef{Reveal: true})
