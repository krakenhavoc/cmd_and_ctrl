package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Citizen's Arrest — Enchantment {1}{W}{W} (Dominaria United):
//
//	"When this enchantment enters, exile target creature or planeswalker
//	 an opponent controls until this enchantment leaves the battlefield."
//
// Oblivion Ring's shape (CR 610.3): the exile records its own return, so
// the permanent comes back as a new object under its owner's control the
// moment this leaves. If the enchantment leaves before the trigger
// resolves, nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f554418-9ee6-4a03-b81e-f93d5c70873e",
		Name:         "Citizen's Arrest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Targets:   effects.TargetPermanent("target creature or planeswalker an opponent controls", effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()),
				Key:       "Citizen's Arrest — exile target creature or planeswalker an opponent controls",
				Effect:    effects.ExileChosenTargetUntilThisLeaves("Citizen's Arrest — the exiled card returns when Citizen's Arrest leaves the battlefield"),
			},
		},
	})
}
