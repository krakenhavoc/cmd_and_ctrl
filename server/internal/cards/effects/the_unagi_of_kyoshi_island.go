package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Unagi of Kyoshi Island — Legendary Creature — Serpent
// {3}{U}{U}, 5/5:
//
//	"Flash
//	 Ward—Waterbend {4}. (Whenever this creature becomes the target of
//	 a spell or ability an opponent controls, counter it unless that
//	 player pays {4}. They can tap their artifacts and creatures to
//	 help. Each one pays for {1}.)
//	 Whenever an opponent draws their second card each turn, you draw
//	 two cards."
//
// The proof card for a ward whose cost is a waterbend cost (#1311).
// `WardWaterbend("{4}")` queues the same CR 118.12 pay-or-counter
// prompt a "Ward {4}" does, carrying CR 701.67a's clause: the PAYER —
// the opponent whose spell targeted the Unagi — may tap untapped
// artifacts and creatures they control, each paying {1} of the {4},
// and pays the rest with mana. Charging a plain {4} instead would make
// the ward harder to pay than printed, the card stronger than it is
// (#259), which is why it waited for the prompt to learn a tap list.
//
// The draw trigger is Faerie Mastermind's, twice over:
// opponentDrewTheirSecondCardThisTurn fires once per opponent per turn,
// on the draw that brings their tally to two.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "be922360-ee9f-4cfb-bf10-d481c5de5cb1",
		Name:            "The Unagi of Kyoshi Island",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Ward(WardWaterbend("{4}"), "The Unagi of Kyoshi Island — ward—waterbend {4}"),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentDrewTheirSecondCardThisTurn(ev, source, g)
			}, "The Unagi of Kyoshi Island — draw two cards", Do(DrawCards{N: 2})),
		},
	})
}
