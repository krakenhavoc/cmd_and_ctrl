package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vivien, Arkbow Ranger — Legendary Planeswalker — Vivien {1}{G}{G}{G}, loyalty 4
// (Core Set 2020, #5518):
//
//	"+1: Distribute two +1/+1 counters among up to two target creatures. They
//	 gain trample until end of turn.
//	 −3: Target creature you control deals damage equal to its power to target
//	 creature or planeswalker.
//	 −5: You may reveal a creature card you own from outside the game and put
//	 it into your hand."
//
// The +1 is Abzan Charm's distribution over up to two targets (each at least
// one counter, so one target takes both), then trample for every target still
// legal. The −3 is Hunter's Mark's one-sided bite over two clauses (the damage
// is dealt only if both targets are still legal, CR 608.2b). The −5 is a
// wish (ADR 0191): the sideboard, or the card database when the deck seated
// none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3e83dec6-e1a0-4137-aaab-308dfd3122e1",
		Name:            "Vivien, Arkbow Ranger",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Distribute two +1/+1 counters among up to two target creatures. They gain trample until end of turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2).Dividing(effects.Divide(2)),
				Effect:  effects.OldB155VivienPlusOne,
			},
			{
				Label: "−3: Target creature you control deals damage equal to its power to target creature or planeswalker.",
				Cost:  effects.LoyaltyCost(-3),
				Targets: effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.OneSidedBite(item, effects.NewContext(g, item))
				},
			},
			{
				Label: "−5: You may reveal a creature card you own from outside the game and put it into your hand.",
				Cost:  effects.LoyaltyCost(-5),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.TakeFromOutside{
						Pick:     effects.OldB155VivienWish,
						Question: "Vivien, Arkbow Ranger — reveal a creature card you own from outside the game, or none",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
