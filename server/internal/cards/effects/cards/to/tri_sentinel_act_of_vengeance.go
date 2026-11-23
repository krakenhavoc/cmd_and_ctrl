package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tri-Sentinel, Act of Vengeance — Legendary Artifact Creature — Robot
// Villain {7}, 7/7 (Marvel Super Heroes Commander, #4455):
//
//	"Flying, menace, trample
//	 When Tri-Sentinel enters, for each opponent, Tri-Sentinel deals 3
//	 damage to up to one target creature that player controls.
//	 Unearth {7} (...)"
//
// "For each opponent, up to one target creature that player controls" is
// Molten Primordial's clause list (moltenPrimordialClauses, reused as is:
// one clause per opponent seat, each bound to its player and re-checked on
// resolution, CR 608.2b), and the damage is sourceDealsDamageToEachLegal-
// Target(3): all the picks are dealt damage as one event by Tri-Sentinel
// itself, from its last-known information if it has already left
// (CR 608.2h). Unearth is Dregscape Zombie's keyword at {7}.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Tri-Sentinel, Act of Vengeance — 3 damage to up to one target creature each opponent controls",
		effects.SourceDealsDamageToEachLegalTarget(3))
	enters.TargetsFrom = effects.MoltenPrimordialClauses
	effects.Register(effects.Spec{
		OracleID:        "d9cc9fc9-522e-4598-9a09-c95824d77acd",
		Name:            "Tri-Sentinel, Act of Vengeance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "menace", "trample"},
		Triggered:       []game.TriggeredAbility{enters},
		Activated:       []effects.ActivatedAbility{effects.Unearth("{7}")},
	})
}
