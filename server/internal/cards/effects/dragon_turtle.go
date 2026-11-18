package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragon Turtle — Creature — Dragon Turtle {1}{U}{U}, 3/5:
//
//	"Flash
//	 Drag Below — When this creature enters, tap it and up to one target
//	 creature an opponent controls. They don't untap during their
//	 controllers' next untap steps."
//
// TapAndFreeze over the Turtle and the chosen creature: both are tapped,
// then each skips its own controller's next untap step (a Turtle that
// changed controllers, or a victim that did, follows its current
// controller: CR 611.2c, ADR 0058's controller-keyed mark). The target is
// "up to one": with none chosen the Turtle still taps and freezes itself,
// and a target gone by resolution leaves the Turtle's half alone
// (CR 608.2b). "Drag Below" is an ability word (CR 207.2c).
//
// No simplification.
func init() {
	etb := WhenThisEnters("Dragon Turtle — tap it and up to one target creature an opponent controls; they don't untap during their controllers' next untap steps",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			ids := append([]uuid.UUID{ctx.Source()}, legalTargetCards(item, g)...)
			return TapAndFreeze{Targets: ids, Label: "Dragon Turtle — they don't untap during their controllers' next untap steps"}.Apply(ctx)
		})
	etb.Targets = Harmful(TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1))
	Register(Spec{
		OracleID:        "a3270e64-38d2-49e5-8a1e-a5e81205776b",
		Name:            "Dragon Turtle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Answers: game.AnswerRestrict},
		Triggered:       []game.TriggeredAbility{etb},
	})
}
