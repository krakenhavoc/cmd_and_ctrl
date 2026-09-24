package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloated Contaminator — Creature — Phyrexian Beast {2}{G}, 4/4:
//
//	"Trample
//	 Toxic 1
//	 Whenever this creature deals combat damage to a player,
//	 proliferate."
//
// Toxic is the engine's damage tail (ADR 0056, #748): combat damage to
// a player also gives them one poison counter, in the same step as the
// damage and before this trigger resolves — so the proliferate finds
// the poison already there and adds another.
//
// The proliferate picks for its controller, as every catalogued
// proliferate does (Flux Channeler, Evolution Sage): no choose-what-to-
// proliferate prompt exists yet.
func init() {
	Register(Spec{
		OracleID:        "090018e0-4dcb-4b3c-b4e0-7ba62de0484d",
		Name:            "Bloated Contaminator",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"You don't choose what to proliferate — the game picks for you, adding every counter that helps you and every counter that hurts an opponent."},
		PrintedKeywords: []string{"trample", "toxic 1"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Bloated Contaminator — proliferate", Do(Proliferate{})),
		},
	})
}
