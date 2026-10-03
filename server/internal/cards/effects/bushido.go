package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bushido.go — Bushido N (CR 702.45a): "Whenever this creature blocks or
// becomes blocked, it gets +N/+N until end of turn." Append-only,
// mechanic-named.
//
// One triggered ability watching two kinds: "blocks" once however many
// attackers the creature blocks (CR 509.3a, selfBlocksOnce), and "becomes
// blocked" once however many creatures block it (EventBecomesBlocked).
// Multiple instances trigger separately (CR 702.45b): a card with two
// prints Bushido twice.

// Bushido is "Bushido n" for the card named `name`.
func Bushido(n int, name string) game.TriggeredAbility {
	label := fmt.Sprintf("%s — bushido %d", name, n)
	return OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return selfBlocksOnce(ev, source) || (ev.Kind == game.EventBecomesBlocked && ev.CardID == source.InstanceID)
	}, label, thisGetsUntilEndOfTurn(n, n, label))
}
