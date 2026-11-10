package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hardlight Containment — Enchantment — Aura {W} (Edge of Eternities,
// #4461):
//
//	"Enchant artifact you control
//	 When this Aura enters, exile target creature an opponent controls
//	 until this Aura leaves the battlefield.
//	 Enchanted permanent has ward {1}."
//
// Dimensional Exile on Mechanized Production's enchant clause, with
// Lavaspur Boots' granted ward. The exile is a one-shot and the return a
// second one-shot created as the Aura leaves (CR 610.3), under the card's
// owner's control (CR 610.3c); an Aura removed before its trigger
// resolves exiles nothing (CR 610.3b). The enchant clause is re-checked
// as a state-based action (CR 303.4c): an artifact you stop controlling
// sends the Aura to the graveyard, which returns the creature. The ward
// is the trigger WardAttached carries, watching the permanent the Aura
// is attached to, so the host keeps it only while enchanted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8054092a-1940-4f4b-991d-81f4d384cdb1",
		Name:         "Hardlight Containment",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact you control", Artifact(), YouControl()),
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets:   TargetCreature("target creature an opponent controls", OpponentControls()),
				Key:       "Hardlight Containment — exile target creature an opponent controls until this Aura leaves the battlefield",
				Effect: exileChosenTargetUntilThisLeaves(
					"Hardlight Containment — the exiled card returns when Hardlight Containment leaves the battlefield"),
			},
			WardAttached(WardMana("{1}"), "Hardlight Containment — ward {1}"),
		},
	})
}
