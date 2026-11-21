package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Body Dropper — Creature — Devil Warrior {B}{R}, 2/2 (Streets of New
// Capenna, S82):
//
//	"Whenever you sacrifice another creature, put a +1/+1 counter on this
//	 creature.
//	 {B}{R}, Sacrifice another creature: This creature gains menace until
//	 end of turn."
//
// The trigger is Dina's read of EventSacrifice (it fires before the zone
// move, so the permanent is still a creature to read), minus the source
// itself: a sacrifice of the Dropper does not grow it. A creature
// sacrificed to pay its own ability counts, and a token counts. The
// ability's sacrifice is a cost, so the menace lands on the Dropper only if
// it is still on the battlefield when the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0c465ea-4242-4737-9c3e-69222bdb8f1b",
		Name:         "Body Dropper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, effects.SncB13YouSacrificedAnotherCreature, "Body Dropper — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{R}, Sacrifice another creature: This creature gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{B}{R}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.ThisCreatureUntilEOT("Body Dropper — gains menace", 0, 0, "menace"),
		}},
	})
}
