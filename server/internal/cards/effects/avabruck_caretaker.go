package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avabruck Caretaker // Hollowhenge Huntmaster — {4}{G}{G} Creature —
// Human Werewolf 4/4 // Creature — Werewolf 6/6 (#2586, ADR 0132):
//
//	Front: "Hexproof
//	        At the beginning of combat on your turn, put two +1/+1
//	        counters on another target creature you control.
//	        Daybound"
//	Back:  "Hexproof
//	        Other permanents you control have hexproof.
//	        At the beginning of combat on your turn, put two +1/+1
//	        counters on each creature you control.
//	        Nightbound"
//
// The front trigger targets as it goes on the stack and is removed if
// the controller has no other creature (CR 603.3d); "another" is the
// source-excluding target clause. The back face's grant covers every
// OTHER permanent its controller controls, not just creatures, and the
// counters go on as one placement per creature so a counter replacement
// applies once to each.
//
// No simplification.
func init() {
	const oracle = "730be0d6-2612-44d6-9d36-e1fc6510c6bd"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Avabruck Caretaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof", "daybound"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Avabruck Caretaker — put two +1/+1 counters on another target creature you control",
					b31CountersOnChosenCreature),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Hollowhenge Huntmaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof", "nightbound"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Abilities, "hexproof") {
					c.Abilities = append(c.Abilities, "hexproof")
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Hollowhenge Huntmaster — put two +1/+1 counters on each creature you control",
				func(g *game.Game, item *game.StackItem) error {
					return b11PutCountersOnEachCreatureYouControl(g, item, 2)
				}),
		},
	})
}
