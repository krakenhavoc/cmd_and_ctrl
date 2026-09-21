package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Legend of Kyoshi — Enchantment — Saga, {4}{G}{G}:
//
//	(As this Saga enters and after your draw step, add a lore counter.)
//	I — Draw cards equal to the greatest power among creatures you
//	    control.
//	II — Earthbend X, where X is the number of cards in your hand.
//	     That land becomes an Island in addition to its other types.
//	III — Exile this Saga, then return it to the battlefield
//	      transformed under your control.
//
// The FRONT face of a transforming Saga (ADR 0034, ADR 0079). Chapter
// III is the exile-and-return verb, exactly as
// fable_of_the_mirror_breaker.go and the_legend_of_kuruk.go use it;
// see avatar_kyoshi.go for the back face.
//
// Chapter I reads CurrentPower (layered, so an anthem counts) over
// every creature the controller controls — the same greatest-power
// read Selvala, Heart of the Wilds' mana ability and Return of the
// Wildspeaker's draw mode already share (b42GreatestPowerControlledBy).
//
// SANDBOX SIMPLIFICATION, and it is why this face is not Full:
// chapter II ships only "that land becomes an Island in addition to
// its other types". Earthbend itself — "target land you control
// becomes a 0/0 creature with haste that's still a land, put X +1/+1
// counters on it, and when it dies or is exiled return it to the
// battlefield tapped" — is a keyword action this engine has no shape
// for: turning a land into an animated creature with a delayed
// return-on-death/exile trigger, on top of the counters, is more
// machinery than any existing primitive composes. Nothing in the
// catalog does it yet — filed as #1178, since no card had needed the
// shape before this one — so a land you choose still becomes an
// Island — the clause the sentence's OWN text states independently
// of the earthbend
// — but never becomes a creature, gets no counters, and is not
// returned if it leaves. Weaker than printed (#259): the chosen land
// is real, useful mana fixing, and never a combat threat it would be
// on paper.
func init() {
	Register(Spec{
		OracleID:     theLegendOfKyoshiOracleID,
		Name:         "The Legend of Kyoshi",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Earthbend isn't implemented — the chosen land becomes an Island but is never animated into a 0/0 creature, gets no +1/+1 counters, and isn't returned to the battlefield if it dies or is exiled.",
		},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The Legend of Kyoshi — draw cards equal to the greatest power among creatures you control",
				kyoshiChapterOneDraw),
			ChapterTriggerTargeting(2, "The Legend of Kyoshi — that land becomes an Island in addition to its other types",
				TargetPermanent("target land you control", And(Land(), YouControl())),
				kyoshiChapterTwoBecomesIsland),
			ChapterTrigger(3, "The Legend of Kyoshi — exile it, then return it transformed",
				ChapterExileAndReturnTransformed),
		},
	})
}

// theLegendOfKyoshiOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const theLegendOfKyoshiOracleID = "e41e8754-028c-4399-b530-de9e134d04b0"

// kyoshiChapterOneDraw is "Draw cards equal to the greatest power
// among creatures you control" — read at resolution, over the
// controller's board.
func kyoshiChapterOneDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := b42GreatestPowerControlledBy(g, item.Controller)
	return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
}

// kyoshiChapterTwoBecomesIsland is the declared slice of Earthbend
// this card ships: the chosen land gains the Island subtype for as
// long as it remains that object (CR 611.2a — no duration is printed,
// so the effect is indefinite; CR 611.2c pins it to the one permanent
// targeted, so a land that leaves and returns is a new object this
// effect no longer follows).
func kyoshiChapterTwoBecomesIsland(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	target := item.Targets[0].ID
	applies := SnapshotAffected(ctx, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.InstanceID == target
	})
	if applies == nil {
		return nil
	}
	return StaticForDuration{
		Label:    "The Legend of Kyoshi — that land becomes an Island in addition to its other types",
		Duration: game.IndefiniteDuration(),
		Ability: game.StaticAbility{
			Layer:     game.Layer4Type,
			AppliesTo: applies,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, t := range c.Subtypes {
					if t == "Island" {
						return
					}
				}
				c.Subtypes = append(c.Subtypes, "Island")
			},
		},
	}.Apply(ctx)
}
