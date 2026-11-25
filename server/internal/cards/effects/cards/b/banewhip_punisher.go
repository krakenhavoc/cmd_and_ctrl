package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Banewhip Punisher — Creature — Human Warrior {2}{B}, 2/2 (Hour of
// Devastation, #5534):
//
//	"When this creature enters, you may put a -1/-1 counter on target
//	 creature.
//	 {B}, Sacrifice this creature: Destroy target creature that has a -1/-1
//	 counter on it."
//
// The enters trigger is Affectionate Indrik's targeted "you may": the target
// is chosen as the trigger goes on the stack and the question is asked as it
// resolves, so a "no" does nothing. The activated ability sacrifices as a cost
// (CR 118.3) and can target only a creature that has a -1/-1 counter on it,
// checked on announcement and again on resolution (CR 608.2b), so a counter
// removed in response makes the target illegal and nothing is destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d46c397c-e772-4a29-bbd2-1d7b47a461e8",
		Name:         "Banewhip Punisher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Banewhip Punisher — put a -1/-1 counter on target creature",
					effects.R35NightmareSowerCounter),
				effects.TargetCreature("target creature")),
				"Banewhip Punisher — put a -1/-1 counter on that creature?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice this creature: Destroy target creature that has a -1/-1 counter on it.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature that has a -1/-1 counter on it",
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters[game.CounterMinusOne] > 0 }),
			Effect: effects.DestroyFirstLegalTarget,
		}},
	})
}
