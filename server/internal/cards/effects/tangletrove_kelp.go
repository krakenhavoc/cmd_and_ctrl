package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tangletrove Kelp — Artifact Creature — Clue Plant {5}{U}{U}, 6/6
// (Murders at Karlov Manor Commander):
//
//	"Ward {2}
//	 At the beginning of each combat, other Clues you control become 6/6
//	 Plant creatures in addition to their other types until end of turn.
//	 {2}, Sacrifice this creature: Draw a card."
//
// The animation is BecomeCreatureUntilEOT over every other Clue the
// controller has when the combat begins (CR 611.2c: the set is fixed on
// resolution), pinned to each as the object it is (CR 400.7), adding
// Creature and Plant and setting 6/6 at layer 7b. A Clue that is already
// a creature keeps the types it has. The Kelp is a Clue itself, so it is
// left out, and a Clue that is not a token counts (CR 205.3g).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d7d9404-cad6-4c1a-bd47-7a31ebb2791c",
		Name:         "Tangletrove Kelp",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Tangletrove Kelp — ward {2}"),
			AtBeginningOfEachCombat("Tangletrove Kelp — other Clues you control become 6/6 Plant creatures until end of turn", tangletroveKelpAnimate),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}

func tangletroveKelpAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.InstanceID != item.SourceCardID && c.Controller == ctx.Controller() && c.HasSubtype("Clue") {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (BecomeCreatureUntilEOT{
			Target: id, Types: []string{"Artifact", "Creature"}, Subtypes: []string{"Plant"},
			SetPower: 6, SetToughness: 6,
			Label: "Tangletrove Kelp — becomes a 6/6 Plant creature until end of turn",
		}).applyFor(ctx.asGroupMember(), g.UntilEndOfTurnDuration()); err != nil {
			return err
		}
	}
	return nil
}
