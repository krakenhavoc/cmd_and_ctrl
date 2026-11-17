package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c0c465ea-4242-4737-9c3e-69222bdb8f1b",
		Name:         "Body Dropper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, sncB13YouSacrificedAnotherCreature, "Body Dropper — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}{R}, Sacrifice another creature: This creature gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{B}{R}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect:  thisCreatureUntilEOT("Body Dropper — gains menace", 0, 0, "menace"),
		}},
	})
}
