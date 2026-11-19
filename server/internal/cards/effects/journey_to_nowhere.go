package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Journey to Nowhere — Enchantment {1}{W}:
//
//	"When this enchantment enters, exile target creature.
//	 When this enchantment leaves the battlefield, return the exiled
//	 card to the battlefield under its owner's control."
//
// Oblivion Ring's shape over any creature: the return is CR 610.3's
// one-shot recorded by ExileUntil, and if the enchantment leaves before
// the trigger resolves nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2bfec482-be0d-4768-84e7-1c2c74ea1373",
		Name:         "Journey to Nowhere",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets:   TargetCreature("target creature"),
				Key:       journeyToNowhereExileLabel,
				Effect:    exileChosenTargetUntilThisLeaves("Journey to Nowhere — the exiled card returns when Journey to Nowhere leaves the battlefield"),
			},
			UntilThisLeavesLegacyReturn("Journey to Nowhere — return the exiled card", journeyToNowhereExileLabel),
		},
	})
}

const journeyToNowhereExileLabel = "Journey to Nowhere — exile target creature"
