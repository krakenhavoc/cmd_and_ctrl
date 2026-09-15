package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obyra, Dreaming Duelist — Legendary Creature — Faerie Warrior
// {U}{B}, 2/2 (EDHREC rank 3401):
//
//	"Flash
//	 Flying
//	 Whenever another Faerie you control enters, each opponent loses
//	 1 life."
//
// The Faerie deck's drain. Both keywords ride PrintedKeywords — flash
// is what lets her enter at instant speed, and the cast gate reads
// it off the card in hand. The trigger is "another Faerie", read
// through the effective subtypes so a changeling counts, and a
// Faerie token's entry counts since it emits the same event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4162e60f-9d70-4b8c-a838-df57b83c208d",
		Name:            "Obyra, Dreaming Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b32AnotherFaerieYouControlEntered(ev, source, g)
			}, "Obyra, Dreaming Duelist — each opponent loses 1 life", func(g *game.Game, item *game.StackItem) error {
				return eachOpponentLosesLife(g, item, 1)
			}),
		},
	})
}
