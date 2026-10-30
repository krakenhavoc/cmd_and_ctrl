package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// WheneverYourCreaturesFightOrBecomeBlocked counts a fight once even when
// you control both fighters, and a declaration once however many of your
// attackers become blocked. Separate fights in one resolution remain
// separate occurrences. Fight controllers are frozen before damage.
func WheneverYourCreaturesFightOrBecomeBlocked(label string, effect Effect) game.TriggeredAbility {
	t := OncePerBatch(OnAny([]game.EventKind{game.EventFight, game.EventBecomesBlocked},
		func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventFight {
				return ev.Actor == source.Controller
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && c.Controller == source.Controller && c.IsCreature()
		}, label, effect))
	t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string {
		if ev.Kind == game.EventFight {
			return "fight/" + strconv.FormatUint(ev.BatchSeq, 10)
		}
		return "blocked"
	}
	return t
}

// itFightsTheFirstTarget is the body of "<this permanent> fights up to one
// target creature": the trigger's source fights the first legal target, and
// nothing happens when none was chosen (the "up to one" shape) or either
// fighter has gone (b10Fight).
func itFightsTheFirstTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return b10Fight(ctx, item.SourceCardID, id)
}
