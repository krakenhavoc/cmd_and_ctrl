package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frenzied Saddlebrute — Creature — Orc Warrior {4}{R}, 5/4 (Commander
// Legends):
//
//	"Haste
//	 All creatures can attack your opponents and planeswalkers your opponents
//	 control as though those creatures had haste."
//
// Haste, and a layer-6 static that writes the Saddlebrute's controller onto
// every creature (AllCreaturesCanAttackYourOpponentsAsThoughHaste). A sick
// creature, anyone's, may then be declared attacking a player other than that
// controller or a planeswalker the controller does not control; against the
// controller and their planeswalkers it stays sick (CR 302.6, CR 609.4). The
// creatures do not gain haste, so their {T} abilities stay unavailable.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f9d72bd4-8b4c-47bb-9361-a8cb5e5f64cb",
		Name:            "Frenzied Saddlebrute",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{effects.AllCreaturesCanAttackYourOpponentsAsThoughHaste()},
	})
}
