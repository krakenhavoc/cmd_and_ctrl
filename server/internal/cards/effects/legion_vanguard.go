package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Legion Vanguard — Creature — Vampire Soldier {1}{B}, 2/2 (Commander
// Masters):
//
//	"{1}, Sacrifice another creature: This creature explores. (Reveal the
//	 top card of your library. Put that card into your hand if it's a land.
//	 Otherwise, put a +1/+1 counter on this creature, then put the card back
//	 or put it into your graveyard.)"
//
// Agency Coroner's sacrifice-another-creature outlet paying for an explore
// by the Vanguard itself (Explores, CR 701.44): the explorer is the source as
// it is when the ability resolves, so a Vanguard that has left gets no
// counter but the reveal still happens (CR 701.44c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f2c8398-1606-400b-a593-2635656e8c4b",
		Name:         "Legion Vanguard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: This creature explores.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Explores{Explorer: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
