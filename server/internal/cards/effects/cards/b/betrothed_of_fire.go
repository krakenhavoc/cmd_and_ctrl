package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Betrothed of Fire — Enchantment — Aura {1}{R} (Weatherlight):
//
//	"Enchant creature
//	 Sacrifice an untapped creature: Enchanted creature gets +2/+0 until end of
//	 turn.
//	 Sacrifice enchanted creature: Creatures you control get +2/+0 until end of
//	 turn."
//
// Either ability is the Aura's controller's, whoever controls the enchanted
// creature. The first sacrifices an untapped creature of the activator's (the
// enchanted one qualifies); the second sacrifices the creature the Aura is
// attached to and then pumps the activator's creatures, those left after the
// sacrifice (CR 611.2c: the set is locked as it resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77849657-c872-4d54-9149-a08bbe90f787",
		Name:         "Betrothed of Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Sacrifice an untapped creature: Enchanted creature gets +2/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
				Cost:    effects.SacrificeN(1, "an untapped creature", effects.Creature(), effects.Untapped()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					host, ok := effects.EnchantedCreatureAsItResolves(ctx)
					if !ok || !effects.OnBattlefield(g, host) {
						return nil
					}
					return effects.BoostUntilEOT{Target: host, Power: 2, Label: "Betrothed of Fire — +2/+0"}.Apply(ctx)
				},
			},
			{
				Label:   "Sacrifice enchanted creature: Creatures you control get +2/+0 until end of turn.",
				Purpose: game.Purpose{TeamPump: &game.Pump{Power: 2}, Answers: game.AnswerPump | game.AnswerSacOutlet},
				Cost:    effects.SacrificeEnchantedCreature(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2,
						Label: "Betrothed of Fire — creatures you control get +2/+0",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
