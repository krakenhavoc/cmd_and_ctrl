package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b04_helpers.go — shared shapes for March of the Machine, S82
// batch 4 (#5477): backup cards that grant a triggered ability.

// s82MomB04NthBackup is Backup for a card that prints the keyword more than
// once (Conclave Sledge-Captain: "Backup 1, backup 1, backup 1"). Each
// instance is its own triggered ability (CR 702.165, 702.165f), so each row
// carries its own Key, which a restore reads to name the row.
func S82MomB04NthBackup(name string, n, idx int, grantKey string, keywords ...string) game.TriggeredAbility {
	t := Backup(name, n, grantKey, keywords...)
	t.Key += " (" + strconv.Itoa(idx) + ")"
	return t
}

// s82MomB04ThisDealtCombatDamage is "whenever this creature deals combat
// damage to a player" for an ability whose effect reads "that many": the
// damage the event carried is captured on the item (Barrowgoyf's shape), a
// fact about the damage dealt, not about the board as the ability resolves.
func S82MomB04ThisDealtCombatDamage(when When, label string, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventDealDamage},
		AppliesTo: when,
		Key:       label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = ev.Amount
			return item
		},
		Effect: effect,
	}
}
