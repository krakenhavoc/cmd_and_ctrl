package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gandalf of the Secret Fire — Legendary Creature — Avatar Wizard
// {1}{U}{R}{W}, 3/4 (Tales of Middle-earth Commander):
//
//	"Whenever you cast an instant or sorcery spell from your hand during
//	 an opponent's turn, exile that card with three time counters on it
//	 instead of putting it into your graveyard as it resolves. Then if the
//	 exiled card doesn't have suspend, it gains suspend. (At the beginning
//	 of your upkeep, remove a time counter. When the last is removed, you
//	 may play it without paying its mana cost.)"
//
// Goliath Daydreamer's mark on the spell: the trigger is gated to a cast
// from the hand during another player's turn, and marks the spell on the
// stack so that, if it resolves, its card is exiled with three time
// counters and gains suspend if it has none (game.StackItem.ExileAndSuspend,
// resolution_exits.go). A spell that is countered or fizzles goes to the
// graveyard as usual, and a copy is not a card, so it is never marked.
//
// No simplification.
func init() {
	mark := effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
		"Gandalf of the Secret Fire — exile that card with three time counters on it instead of putting it into your graveyard as it resolves; it gains suspend",
		func(g *game.Game, item *game.StackItem) error {
			ev := effects.NewContext(g, item).Trigger().Event
			if ev.OldZone != game.ZoneHand || ev.CardID == uuid.Nil {
				return nil
			}
			g.MarkSpellExilesAndSuspendsForEffect(ev.CardID, 3)
			return nil
		})
	inner := mark.AppliesTo
	mark.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return ev.OldZone == game.ZoneHand && !effects.IsActivePlayer(g, source.Controller) && inner(ev, source, ch, g)
	}
	effects.Register(effects.Spec{
		OracleID:     "cbad35cd-9026-43ea-8333-94f3b34459e3",
		Name:         "Gandalf of the Secret Fire",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{mark},
	})
}
