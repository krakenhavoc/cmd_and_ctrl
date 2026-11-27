package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shard Phoenix — Creature — Phoenix {4}{R}, 2/2 (Stronghold, S82 OLD b751,
// #5627):
//
//	"Flying
//	 Sacrifice this creature: It deals 2 damage to each creature without flying.
//	 {R}{R}{R}: Return this card from your graveyard to your hand. Activate only
//	 during your upkeep."
//
// Bloodfire Dwarf's sacrifice ping at 2 over the creatures without flying (read
// through the layered view as it resolves), and Brackish Trudge's graveyard
// return to hand gated on the controller's own upkeep (Firemane Angel's
// DuringYourUpkeep).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3b629793-0efb-4d60-9d82-38b7b400abfd",
		Name:            "Shard Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Sacrifice this creature: It deals 2 damage to each creature without flying.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
				Cost:    effects.SacrificeThis(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 2)
				},
			},
			{
				Label:     "{R}{R}{R}: Return this card from your graveyard to your hand. Activate only during your upkeep.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      effects.ManaCost("{R}{R}{R}"),
				Zones:     []game.ZoneKind{game.ZoneGraveyard},
				Condition: effects.DuringYourUpkeep(),
				Effect:    effects.ReturnThisFromYourGraveyardToHand,
			},
		},
	})
}
