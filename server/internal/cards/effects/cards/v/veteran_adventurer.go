package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Veteran Adventurer — Creature — Human {5}{G}, 5/5 (Zendikar Rising,
// #5508):
//
//	"Veteran Adventurer is also a Cleric, Rogue, Warrior, and Wizard.
//	 This spell costs {1} less to cast for each creature in your party.
//	 Vigilance"
//
// A layer 4 static adds the party types to the card itself. The discount is
// read as the spell is cast, with the Adventurer on the stack and so not yet
// in the party (CR 700.8 counts creatures you control).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "ea3401a9-76db-4f38-b973-2b7754087b97",
		Name:              "Veteran Adventurer",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"vigilance"},
		Static:            []game.StaticAbility{effects.OldB095AlsoPartyTypes("Veteran Adventurer")},
		SelfCostModifiers: []game.CostModifier{effects.OldB095PartyCostsLess()},
	})
}
