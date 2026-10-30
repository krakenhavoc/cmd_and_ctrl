package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// M.O.D.O.K. — Legendary Artifact Creature — Villain {3}{B}{B}, 2/2
// (Marvel Super Heroes, #4455):
//
//	"Flying, lifelink
//	 Mental Organism — Pay 3 life: M.O.D.O.K. connives. Activate only
//	 during your turn.
//	 Designed Only for Killing — Creatures your opponents control get
//	 -1/-1."
//
// The life is a cost, so it is paid even if the ability is countered,
// and an activator with fewer than 3 life cannot start it (CR 119.4).
// Connive is the shared chain (draw, discard, a +1/+1 counter for a
// nonland card). The shrink is Elesh Norn's second static at -1/-1,
// layer 7c, so a 1/1 an opponent controls dies at the next check.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe7cc414-42d0-4dad-8516-7a0be0860a67",
		Name:         "M.O.D.O.K.",
		Completeness: CompletenessFull,
		// ADR 0126 §6: a static -1/-1 removes the same creatures a sweep does.
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		PrintedKeywords: []string{"flying", "lifelink"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Creatures your opponents control get -1/-1.",
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller != source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power--
				c.Toughness--
			},
		}},
		Activated: []ActivatedAbility{{
			Label:     "Mental Organism — Pay 3 life: M.O.D.O.K. connives. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      PayLife(3),
			Condition: DuringYourTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ConniveEach{Controller: item.Controller, Creatures: []uuid.UUID{item.SourceCardID}}.Apply(NewContext(g, item))
			},
		}},
	})
}
