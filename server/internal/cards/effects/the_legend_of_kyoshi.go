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
// CHAPTER II IS TWO CLAUSES ABOUT ONE LAND, and since #1178 both of
// them ship. "Earthbend X" is the keyword action (earthbend.go,
// game/earthbend.go) with X counted at resolution over the
// controller's hand; "that land becomes an Island in addition to its
// other types" is the sentence's own independent clause, a second
// layer-4 type add on the same target.
//
// THE TWO CLAUSES ARE TWO CONTINUOUS EFFECTS, not one. They are both
// layer 4 and both indefinite, but they have different lifetimes:
// earthbend's animation is pinned to the object (CR 400.7 — a land
// that dies and comes back tapped is a new object and is not animated
// any more), and the Island clause is pinned the same way for the same
// reason, so the returning land is a plain land again. Registering
// them separately is also what makes the earthbend one line: the
// keyword action owns its four parts and this card owns its fifth
// sentence.
//
// The Island matters beyond flavour when X is small — a 0/0 with no
// counters dies to the toughness state-based action and comes back
// tapped, and what you kept is a land that taps for {U}.
//
// It also ships this face at Full (#1179 registered it with the
// earthbend caveat; #1178 cleared it).
func init() {
	Register(Spec{
		OracleID:     theLegendOfKyoshiOracleID,
		Name:         "The Legend of Kyoshi",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The Legend of Kyoshi — draw cards equal to the greatest power among creatures you control",
				kyoshiChapterOneDraw),
			ChapterTriggerTargeting(2, "The Legend of Kyoshi — earthbend X, and that land becomes an Island in addition to its other types",
				EarthbendTargets(),
				kyoshiChapterTwo),
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

// kyoshiChapterTwo is the chapter's two clauses over one land, in
// printed order: "Earthbend X, where X is the number of cards in your
// hand. That land becomes an Island in addition to its other types."
//
// X is counted at RESOLUTION, over the hand the controller has when
// the chapter trigger resolves — so a discard outlet in response
// shrinks it and a draw grows it. Zero is legal: an empty hand
// earthbends 0, which animates the land into a 0/0 that dies to the
// toughness SBA and comes back tapped, and the Island below is what
// the chapter leaves behind.
func kyoshiChapterTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	hand := 0
	if p := g.PlayerByIDForEffect(item.Controller); p != nil {
		hand = p.Hand.Size()
	}
	if err := (Earthbend{Target: target, N: hand}).Apply(ctx); err != nil {
		return err
	}
	return kyoshiChapterTwoBecomesIsland(ctx, target)
}

// kyoshiChapterTwoBecomesIsland is the chapter's second clause: the
// chosen land gains the Island subtype for as long as it remains that
// object (CR 611.2a — no duration is printed, so the effect is
// indefinite; CR 611.2c pins it to the one permanent targeted, so a
// land that leaves and returns is a new object this effect no longer
// follows).
func kyoshiChapterTwoBecomesIsland(ctx *Context, target uuid.UUID) error {
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
