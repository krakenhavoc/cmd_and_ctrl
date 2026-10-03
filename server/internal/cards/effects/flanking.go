package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flanking.go — Flanking (CR 702.25a): "Whenever this creature becomes
// blocked by a creature without flanking, the blocking creature gets
// -1/-1 until end of turn." Append-only, mechanic-named.
//
// A trigger per blocker (EventBlock names the blocker in CardID and the
// attacker in Target, once per pair of the final declaration), so a
// creature blocked by two creatures without flanking shrinks both. Whether
// the blocker has flanking is read as it blocks (CR 509.3), and multiple
// instances trigger separately (CR 702.25b).

// Flanking is "Flanking" for the card named `name`.
func Flanking(name string) game.TriggeredAbility {
	label := fmt.Sprintf("%s — flanking", name)
	return On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventBlock || ev.Target != source.InstanceID || ev.CardID == source.InstanceID {
			return false
		}
		blocker, ok := g.LookupCardForEffect(ev.CardID)
		return ok && !game.HasKeyword(&blocker, "flanking")
	}, label, flankingShrink)
}

// flankingShrink gives the blocker the trigger saw -1/-1 until end of
// turn. A blocker that has left the battlefield is gone (CR 400.7).
func flankingShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
		return nil
	}
	return BoostUntilEOT{Target: ctx.Trigger().Event.CardID, Power: -1, Toughness: -1}.Apply(ctx)
}
