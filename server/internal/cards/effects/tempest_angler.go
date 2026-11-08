package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tempest Angler — Creature — Otter Wizard {1}{U/R}{U/R}, 2/2
// (Bloomburrow, #4467):
//
//	"Whenever you cast a noncreature spell, put a +1/+1 counter on this
//	 creature."
//
// Boar-q-pine's trigger: it fires as the spell becomes cast and resolves
// before the spell does, putting the counter on this creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60ac490c-b489-4ad2-bcc2-3babaabf8ccb",
		Name:         "Tempest Angler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Tempest Angler — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
