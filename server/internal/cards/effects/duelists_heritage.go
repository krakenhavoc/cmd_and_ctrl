package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duelist's Heritage — Enchantment {2}{W} (Commander 2016, Jumpstart):
//
//	"Whenever one or more creatures attack, you may have target attacking
//	 creature gain double strike until end of turn."
//
// Any player's attack, not only yours: one trigger per declaration
// (OncePerBatch, since the engine emits EventAttack per creature), and the
// target may be any attacking creature, as printed. The target is chosen as
// the trigger goes on the stack and re-checked on resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5cdc4654-b8c5-4599-a1d2-9e9d5f6509cb",
		Name:         "Duelist's Heritage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Optional(Targeting(On(game.EventAttack, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventAttack
			}, "Duelist's Heritage — target attacking creature gains double strike until end of turn",
				jmpB7FirstLegalTargetGains("Duelist's Heritage — double strike", "double strike")),
				TargetCreature("target attacking creature", AttackingCreature())),
				"Have the target gain double strike?")),
		},
	})
}
