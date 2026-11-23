package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Contaminant Grafter — Creature — Phyrexian Druid {4}{G}, 5/5:
//
//	"Trample, toxic 1
//	 Whenever one or more creatures you control deal combat damage to
//	 one or more players, proliferate.
//	 Corrupted — At the beginning of your end step, if an opponent has
//	 three or more poison counters, draw a card, then you may put a land
//	 card from your hand onto the battlefield."
//
// The damage trigger is one per combat damage step (CR 603.2c) however
// many creatures connect and however many players they hit, so it is
// batched with no per-player key (compare Bloated Contaminator's own
// "this creature" trigger). Toxic's poison is given in the same step as
// the damage (CR 702.164c), so the proliferate finds it.
//
// Corrupted is the intervening "if" (CR 603.4) on the end-step trigger:
// some opponent has three or more poison counters when the step begins
// and again as the trigger resolves. The land is a put, not a play (CR
// 305.4), asked after the draw, so a land just drawn can be put.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ba922414-878f-4261-9bce-500c5c9e9506",
		Name:            "Contaminant Grafter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "toxic 1"},
		Purpose:         game.Purpose{Draws: 1, Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			}, "Contaminant Grafter — proliferate", effects.Do(effects.Proliferate{}))),
			effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, contaminantGrafterCorrupted),
				"Corrupted — Contaminant Grafter — draw a card, then you may put a land card from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SeedcoreCorrupted(g, item.Controller, uuid.Nil) {
						return nil
					}
					return effects.Do(effects.DrawCards{N: 1}, effects.MayPutALandFromHand("Contaminant Grafter"))(g, item)
				}),
		},
	})
}

// contaminantGrafterCorrupted is corrupted as a trigger condition.
func contaminantGrafterCorrupted(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return effects.SeedcoreCorrupted(g, source.Controller, uuid.Nil)
}
