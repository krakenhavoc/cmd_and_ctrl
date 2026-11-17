package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fervent Charge — Enchantment {1}{R}{W}{B} (Apocalypse, #5614):
//
//	"Whenever a creature you control attacks, it gets +2/+2 until end
//	 of turn."
//
// One trigger per attacking creature you control (the text is "a
// creature", not "one or more"), read off the declaration's EventAttack
// (attackDeclaredByYou, CR 508.1m). "It" is the creature the event
// names (item.Trigger.Event.CardID). BoostUntilEOT does nothing to a
// creature that has left the battlefield, and the boost ends in the
// cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6226db2b-d1a8-4c41-b802-67d3f65d2ca3",
		Name:         "Fervent Charge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Fervent Charge — it gets +2/+2 until end of turn", ferventChargeBoost),
		},
	})
}

// ferventChargeBoost gives the creature that attacked +2/+2 until end
// of turn.
func ferventChargeBoost(g *game.Game, item *game.StackItem) error {
	return BoostUntilEOT{
		Target:    item.Trigger.Event.CardID,
		Power:     2,
		Toughness: 2,
		Label:     "Fervent Charge — +2/+2",
	}.Apply(NewContext(g, item))
}
