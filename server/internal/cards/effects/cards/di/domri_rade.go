package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	creatures := effects.CreaturesTheEmblemsOwnerControls
	effects.Register(effects.Spec{
		OracleID:        "0c4a2c42-0874-40b4-8d12-47eeadc3e976",
		Name:            "Domri Rade",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Domri Rade emblem",
			Text:  "Creatures you control have double strike, trample, hexproof, and haste.",
			Static: []game.StaticAbility{
				effects.KeywordGrant(creatures, "double strike"),
				effects.KeywordGrant(creatures, "trample"),
				effects.KeywordGrant(creatures, "hexproof"),
				effects.KeywordGrant(creatures, "haste"),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Look at the top card of your library. If it's a creature card, you may reveal it and put it into your hand.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 1),
						Match:    effects.Creature(),
						Max:      1,
						Optional: true,
						Label:    "Domri Rade — you may reveal the creature card and put it into your hand",
						Key:      domriRadeTake,
					}.Apply(ctx)
				},
			},
			{
				Label: "−2: Target creature you control fights another target creature.",
				Cost:  effects.LoyaltyCost(-2),
				Targets: effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.Distinct(effects.TargetCreature("another target creature")),
				),
				Effect: effects.DomriFight,
			},
			{
				Label:  "−7: You get an emblem with \"Creatures you control have double strike, trample, hexproof, and haste.\"",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// domriRadeTake is the +1's pick: the revealed card goes to hand, and an
// untaken card stays where it is, on top.
var domriRadeTake = effects.RegisterTake("choose-cards/domri-rade-take", effects.TakeDef{Reveal: true})
