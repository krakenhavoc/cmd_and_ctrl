package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Party Thrasher — Creature — Lizard Wizard {1}{R}, 1/4 (Modern Horizons
// 3):
//
//	"Noncreature spells you cast from exile have convoke. (Each creature
//	 you tap while casting a noncreature spell from exile pays for {1} or
//	 one mana of that creature's color.)
//	 At the beginning of your first main phase, you may discard a card. If
//	 you do, exile the top two cards of your library, then choose one of
//	 them. You may play that card this turn."
//
// The convoke is Hoarding Broodlord's grant (SpellsYouCastHaveConvoke) over
// spells cast from exile that are not creature spells. The trigger is
// Fireglass Mentor's exile-then-choose behind a discard: the discard is the
// run's own (PlayerDiscardsThenKeyedForEffect), the exile happens only once
// a card has really left the hand, and only the chosen card may be played
// (a land is played, a spell cast, paying its costs) until end of turn.
// Your first main phase is the precombat main phase, which an extra phase
// added later does not repeat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03cd7583-d054-4359-8c94-7776c54198ed",
		Name:         "Party Thrasher",
		Completeness: CompletenessFull,
		GrantedTapCosts: []game.GrantedTapCost{
			SpellsYouCastHaveConvoke("Noncreature spells you cast from exile have convoke.", gapB53NoncreatureFromExile),
		},
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Party Thrasher — you may discard a card; if you do, exile the top two cards of your library, choose one, you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						UpTo:     true,
						Question: "Party Thrasher — you may discard a card to exile the top two cards of your library",
					}, gapB53ThrasherExileTwo, nil)
				}),
		},
	})
}
