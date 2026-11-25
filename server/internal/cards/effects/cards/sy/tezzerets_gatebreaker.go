package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret's Gatebreaker — Artifact {4} (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"When this artifact enters, look at the top five cards of your library.
//	 You may reveal a blue or artifact card from among them and put it into
//	 your hand. Put the rest on the bottom of your library in a random order.
//	 {5}{U}, {T}, Sacrifice this artifact: Creatures you control can't be
//	 blocked this turn."
//
// Skalla Wolf's look-at-five for a blue or artifact card, and Jace, Arcane
// Strategist's unblockable grant to the creatures you control as it resolves,
// paid with a tap and a sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb35d7e5-eddd-4e02-99ac-1a42cd575041",
		Name:         "Tezzeret's Gatebreaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tezzeret's Gatebreaker — look at the top five cards of your library",
				effects.LookAtTopThenMayTakeToHand(5, effects.Or(effects.OfColor("U"), effects.Artifact()), 1,
					"Tezzeret's Gatebreaker — you may reveal a blue or artifact card and put it into your hand")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{U}, {T}, Sacrifice this artifact: Creatures you control can't be blocked this turn.",
			Cost:    effects.Plus(effects.ManaCost("{5}{U}"), effects.TapCost(), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.RestrictUntilEOT{
					Scope:        game.ScopeYourCreatures,
					Restrictions: game.CantBeBlocked,
					Label:        "Tezzeret's Gatebreaker — can't be blocked",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
