package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b36_helpers.go — shared pieces for Star Trek Commander batch 36
// (#4974).

// trekB36AssimilateTogether is the Star Trek keyword action assimilate on
// several cards in one zone at once (trekB11Assimilate's entry: a +1/+1
// counter, a Borg artifact creature that loses its other creature types),
// put onto the battlefield under the resolving controller's control as one
// simultaneous entry (CR 603.6a). A card that is no longer in `from`, a
// token or a nonpermanent card is skipped.
func trekB36AssimilateTogether(ctx *Context, from game.ZoneKind, cards []uuid.UUID, label string) error {
	if len(cards) == 0 {
		return nil
	}
	return putCardsFromZoneTogetherWith(ctx, from, cards,
		game.ZoneEntryOptions{
			Controller:         ctx.Controller(),
			EntersWithCounters: map[string]int{game.CounterPlusOne: 1},
			EntersAs: &game.EntersAs{
				AddTypes:      []string{"Artifact"},
				CreatureTypes: ctx.TextSubtypes("Borg"),
				Label:         label,
			},
		}, game.TailRef{}, 0)
}

// trekB36OwnedByNthOpponent passes for a card owned by the live opponent
// in the n-th seat (1-based) after the chooser's, in turn order: "that
// player's graveyard" for a per-opponent graveyard clause. Eliminated
// seats keep their place, so the numbering never shifts mid-game, and an
// absent or eliminated seat admits nothing.
func trekB36OwnedByNthOpponent(n int) CardPredicate {
	return func(g *game.Game, chooser uuid.UUID, c game.Card) bool {
		p := g.OpponentSeatAfter(chooser, n)
		return p != nil && c.Owner == p.ID
	}
}

// trekB36ArtifactCreaturesYouControl is "artifact creatures you control",
// read on post-layer types, the source included when it is one.
func trekB36ArtifactCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsArtifact() && target.IsCreature() && target.Controller == source.Controller
}
