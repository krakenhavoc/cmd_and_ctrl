package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Augur of Bolas — Creature — Merfolk Wizard {1}{U}, 1/3 (Foundations
// Jumpstart, #5929):
//
//	"When this creature enters, look at the top three cards of your
//	 library. You may reveal an instant or sorcery card from among them and
//	 put it into your hand. Put the rest on the bottom of your library in
//	 any order."
//
// Quandrix Apprentice's look-three-take-one with an instant or sorcery
// as the card taken.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be5d6779-ecc9-44b9-8c4d-97a243059d55",
		Name:         "Augur of Bolas",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Augur of Bolas — look at the top three cards; you may reveal an instant or sorcery card and put it into your hand",
				effects.SocB10LookThreeTakeOne(effects.Or(effects.Instant(), effects.Sorcery()), "Augur of Bolas — you may reveal an instant or sorcery card from among them and put it into your hand")),
		},
	})
}
