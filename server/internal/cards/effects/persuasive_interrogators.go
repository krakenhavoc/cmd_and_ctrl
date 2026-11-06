package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Persuasive Interrogators — Creature — Gorgon Detective {4}{B}{B}, 5/6
// (Murders at Karlov Manor, #4469):
//
//	"When this creature enters, investigate. (Create a Clue token. It's an
//	 artifact with "{2}, Sacrifice this token: Draw a card.")
//	 Whenever you sacrifice a Clue, target opponent gets two poison
//	 counters. (A player with ten or more poison counters loses the
//	 game.)"
//
// The entry is Novice Inspector's Clue. "Whenever you sacrifice a Clue"
// is Tireless Tracker's (lonisYouSacrificedAClue: EventSacrifice is
// emitted before the Clue leaves, so its subtype is read on the
// battlefield), and any Clue counts, not only the one this made. The
// counters are placed by the trigger's controller (Swamp Mosquito's
// AddPlayerCounterByForEffect) on the opponent still targeted as it
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b401179f-eded-4973-8182-df4627133886",
		Name:         "Persuasive Interrogators",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Persuasive Interrogators — investigate", Do(CreateToken{Template: ClueToken(), N: 1})),
			Targeting(
				On(game.EventSacrifice, lonisYouSacrificedAClue,
					"Persuasive Interrogators — target opponent gets two poison counters",
					func(g *game.Game, item *game.StackItem) error {
						for _, t := range NewContext(g, item).LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return g.AddPlayerCounterByForEffect(item.Controller, t.ID, game.CounterPoison, 2)
							}
						}
						return nil
					}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
