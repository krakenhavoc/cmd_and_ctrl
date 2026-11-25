package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Munda, Ambush Leader — Legendary Creature — Kor Ally {2}{R}{W}, 3/4
// (Battle for Zendikar, #5547):
//
//	"Haste
//	 Rally — Whenever Munda or another Ally you control enters, you may look
//	 at the top four cards of your library. If you do, reveal any number of
//	 Ally cards from among them, then put those cards on top of your library
//	 in any order and the rest on the bottom in any order."
//
// The look is the "may"; the revealed Allies and the rest are ordered by
// the player (S82OldB313MundaLook).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4d5da73a-f22a-4446-b5f9-504dfabf1d52",
		Name:            "Munda, Ambush Leader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.S82OldB313Rally("Munda, Ambush Leader — you may look at the top four cards of your library; reveal any number of Ally cards from among them", effects.S82OldB313MundaLook),
				"Munda, Ambush Leader — look at the top four cards of your library?"),
		},
	})
}
