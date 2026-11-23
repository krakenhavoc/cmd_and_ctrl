package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b158_helpers.go — shared bodies of the GAP batch 158 cards (#5948):
// Rooting Moloch's cycling-card filter and exile-and-play, Spiteful Prankster's
// turn-gated first strike and Sarkhan's Whelp's activation watcher.

// gapB158HasCyclingAbility is "a card with a cycling ability" (CR 702.29): the
// card's own activated rows, as the engine reads them, include a cycling one.
// Typecycling counts (CR 702.29e).
func GapB158HasCyclingAbility(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	for _, ab := range game.ActivatedAbilitiesForCard(c) {
		if ab.Cycling {
			return true
		}
	}
	return false
}

// gapB158ExileTargetPlayUntilEndOfYourNextTurn is "exile target card from your
// graveyard. Until the end of your next turn, you may play that card": the
// target is re-checked as the ability resolves (CR 608.2b) and, once it is in
// exile, its controller gets the window.
func GapB158ExileTargetPlayUntilEndOfYourNextTurn(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			return nil
		}
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   item.Controller,
			Zone:     game.ZoneExile,
			Duration: g.UntilEndOfYourNextTurnDuration(item.Controller),
			Label:    name + " — play it until the end of your next turn",
		})
		return nil
	}
}

// gapB158ThisOnYourTurn is "during your turn, this creature has <keyword>":
// the grant applies to the permanent itself while its controller is the
// active player.
func GapB158ThisOnYourTurn(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID && IsActivePlayer(g, source.Controller)
}
