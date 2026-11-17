package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snooping Newsie — Creature — Human Rogue {U}{B}, 2/2 (Streets of New Capenna):
//
//	"When this creature enters, mill two cards.
//	 As long as there are five or more mana values among cards in your
//	 graveyard, this creature gets +1/+1 and has lifelink."
//
// The entry trigger is the ordinary mill (it goes on the stack, so it can be
// answered). The +1/+1 and the lifelink are two self statics (layers 7c and 6)
// on the same graveyard count, so the milled cards can turn them on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "347a5faa-5656-4a36-af44-ab5a3ed267f4",
		Name:         "Snooping Newsie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Snooping Newsie — mill two cards", Do(MillCards{N: 2})),
		},
		Static: []game.StaticAbility{sncSelfPT(1, 1), sncSelfKeywords("lifelink")},
	})
}
