package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prison Realm — Enchantment {2}{W} (War of the Spark, #5520):
//
//	"When this enchantment enters, exile target creature or planeswalker an
//	 opponent controls until this enchantment leaves the battlefield.
//	 When this enchantment enters, scry 1."
//
// Banishing Light's two rows (the exile until this leaves, CR 610.3, and the
// return) with the target narrowed to a creature or planeswalker an opponent
// controls, and a second enters trigger for the scry 1.
//
// No simplification.
func init() {
	const name = "Prison Realm"
	exileLabel := name + " — exile target creature or planeswalker an opponent controls"
	effects.Register(effects.Spec{
		OracleID:     "808a8491-2966-4388-8590-a46ba147f65a",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Targets: effects.TargetPermanent("target creature or planeswalker an opponent controls",
					effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())),
				Key:    exileLabel,
				Effect: effects.ExileChosenTargetUntilThisLeaves(name + " — the exiled card returns when " + name + " leaves the battlefield"),
			},
			effects.UntilThisLeavesLegacyReturn(name+" — return the exiled card", exileLabel),
			effects.WhenThisEnters(name+" — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
