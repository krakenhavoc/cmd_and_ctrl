package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boromir, Gondor's Hope — Legendary Creature — Human Warrior {2}{W}{U},
// 3/4 (Tales of Middle-earth Commander):
//
//	"Whenever Boromir enters or attacks, look at the top six cards of your
//	 library. You may reveal a Human or artifact card from among them and
//	 put it into your hand. Put the rest on the bottom of your library in a
//	 random order."
//
// Militia Bugler's look-and-take, widened to six cards and a Human or
// artifact (the library card's own printed types, as it is not yet a
// permanent). One trigger registered for both events, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bb5e4e5-c3b4-4f67-83e1-bde9ce9f18c4",
		Name:         "Boromir, Gondor's Hope",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Boromir, Gondor's Hope — look at the top six cards of your library",
				effects.LookAtTopThenMayTakeToHand(6, effects.Or(effects.HasSubtype("Human"), effects.Artifact()), 1,
					"Boromir, Gondor's Hope — you may reveal a Human or artifact card and put it into your hand")),
		},
	})
}
