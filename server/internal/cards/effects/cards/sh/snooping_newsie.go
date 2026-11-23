package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "347a5faa-5656-4a36-af44-ab5a3ed267f4",
		Name:         "Snooping Newsie",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Snooping Newsie — mill two cards", effects.Do(effects.MillCards{N: 2})),
		},
		Static: []game.StaticAbility{effects.SncSelfPT(1, 1), effects.SncSelfKeywords("lifelink")},
	})
}
