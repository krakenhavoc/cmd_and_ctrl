package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Birth of the Imperium — Enchantment — Saga {2}{W}{U}{B} (Warhammer 40,000
// Commander):
//
//	I   — Create a 2/2 white Astartes Warrior creature token with vigilance
//	      for each opponent you have.
//	II  — Each opponent sacrifices a creature of their choice.
//	III — Draw two cards for each opponent who controls fewer creatures
//	      than you.
//
// "Each opponent you have" counts the opponents still in the game as the
// chapter resolves. Chapter III compares creature counts as it resolves,
// one comparison per opponent (CR 608.2c), and draws all the cards at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f0544e3-965f-4b12-baaf-d7dfea8d4d41",
		Name:         "Birth of the Imperium",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel("Birth of the Imperium", 1, "create an Astartes Warrior for each opponent"),
				birthOfTheImperiumWarriors),
			effects.ChapterTrigger(2, effects.SagaChapterLabel("Birth of the Imperium", 2, "each opponent sacrifices a creature"),
				effects.EachOpponentSacrificesACreature),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("Birth of the Imperium", 3, "draw two cards for each opponent who controls fewer creatures than you"),
				birthOfTheImperiumDraw),
		},
	})
}

func birthOfTheImperiumCreature(c game.Card) bool { return c.IsCreature() }

// birthOfTheImperiumWarriors is chapter I.
func birthOfTheImperiumWarriors(g *game.Game, item *game.StackItem) error {
	n := 0
	effects.EachOpponent(g, item.Controller, func(uuid.UUID) bool { n++; return false })
	if n == 0 {
		return nil
	}
	return effects.CreateToken{
		Controller: item.Controller,
		Template:   effects.TokenCard("2/2 white Astartes Warrior with vigilance"),
		N:          n,
	}.Apply(effects.NewContext(g, item))
}

// birthOfTheImperiumDraw is chapter III.
func birthOfTheImperiumDraw(g *game.Game, item *game.StackItem) error {
	mine := effects.CountControlled(g, item.Controller, birthOfTheImperiumCreature)
	n := 0
	effects.EachOpponent(g, item.Controller, func(opp uuid.UUID) bool {
		if effects.CountControlled(g, opp, birthOfTheImperiumCreature) < mine {
			n++
		}
		return false
	})
	if n == 0 {
		return nil
	}
	return effects.DrawCards{Player: item.Controller, N: 2 * n}.Apply(effects.NewContext(g, item))
}
