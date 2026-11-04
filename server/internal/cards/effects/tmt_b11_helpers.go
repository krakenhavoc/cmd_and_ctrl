package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b11_helpers.go — Teenage Mutant Ninja Turtles, batch 11 (#4457):
// the sneak cards (CR 702.190, sneak.go).

// tmtB11PaidSneak reports whether `altCost`, a cast record's alternative
// cost key (CastProvenance.AltCost, ObjectSnapshot.AltCost), is the
// card's own sneak offer: "if its sneak cost was paid" (CR 702.190).
func tmtB11PaidSneak(altCost string) bool { return altCost == "sneak" }

// tmtB11EnteredSneaked is the intervening "if his sneak cost was paid"
// of an enters trigger (CR 603.4): the entering permanent's own cast
// record.
func tmtB11EnteredSneaked(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return tmtB11PaidSneak(source.Provenance.AltCost)
}

// tmtB11OtherCreaturesYouControl counts the creatures the source's
// controller controls other than the source (Leonardo, Big Brother).
// Read inside a layer pass, the way High Sentinels of Arashin counts.
func tmtB11OtherCreaturesYouControl(g *game.Game, source *game.Card) int {
	if g == nil || g.Battlefield == nil || source == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != source.InstanceID && c.Controller == source.Controller && c.IsCreature() {
			n++
		}
	}
	return n
}

// tmtB11NinjasYouControl counts the Ninjas `player` controls (Dark Leo &
// Shredder), by post-layer subtype, so a changeling counts. The word is
// read through the source's text (CR 612.2).
func tmtB11NinjasYouControl(ctx *Context, player uuid.UUID) int {
	ninja := ctx.TextSubtype("Ninja")
	return countControlled(ctx.Game, player, func(c game.Card) bool { return c.HasSubtype(ninja) })
}
