package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_reprint_b_helpers.go — helpers for the older
// cards reprinted in Reality Fracture products (slice fra-reprint-b).
// Names are prefixed rfReprintB so they cannot collide with another
// slice's.

// rfReprintBMiteSlug is the catalog identity of the Phyrexian Mite.
const rfReprintBMiteSlug = "phyrexian-mite"

// printedPhyrexianMiteToken is the 1/1 colourless Phyrexian Mite
// artifact creature token with toxic 1 and "This token can't block."
// (Skrelv's Hive, White Sun's Twilight). A catalog template and not a
// tokens_table row because the can't-block clause is a static the
// token always prints.
func printedPhyrexianMiteToken() tokenTemplate {
	return tokenTemplate{
		Slug: rfReprintBMiteSlug,
		Card: game.Card{
			Name:      "Phyrexian Mite",
			TypeLine:  "Token Artifact Creature — Phyrexian Mite",
			Power:     1,
			Toughness: 1,
			Keywords:  []string{"toxic 1"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "Toxic 1 (Players dealt combat damage by this creature also get a poison counter.) This token can't block.",
	}
}

// PhyrexianMiteToken is the Mite's card value for CreateToken.
func PhyrexianMiteToken() game.Card { return tokenFromCatalog(printedPhyrexianMiteToken) }

// rfReprintBFirstStrikeWhileCreatureOnYourTurn is "During your turn,
// this creature has first strike" granted to a manland: the keyword is
// on the permanent only while it is a creature and it is its
// controller's turn, so an idle land carries no badge.
func rfReprintBFirstStrikeWhileCreatureOnYourTurn() game.StaticAbility {
	return KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.IsCreature() && isActivePlayer(g, source.Controller)
	}, "first strike")
}

// rfReprintBLowPowerOrToughness is "power or toughness 1 or less",
// read live.
func rfReprintBLowPowerOrToughness() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && (c.CurrentPower() <= 1 || c.CurrentToughness() <= 1)
	}
}

// rfReprintBSharkToken creates an X/X blue Shark creature token with
// flying under `controller`'s control (Shark Typhoon).
func rfReprintBSharkToken(ctx *Context, controller uuid.UUID, x int) error {
	return CreateToken{
		Controller: controller,
		Template: game.Card{
			Name:      "Shark",
			TypeLine:  "Token Creature — Shark",
			Colors:    []string{"U"},
			Power:     x,
			Toughness: x,
			Keywords:  []string{"flying"},
		},
		N: 1,
	}.Apply(ctx)
}

// rfReprintBYouCreatedACreatureToken is the trigger condition "you
// create a creature token": a token entered under your control and is
// a creature (Staff of the Storyteller).
func rfReprintBYouCreatedACreatureToken(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsToken() && c.IsCreature()
}

// rfReprintBCorrupted is corrupted: an opponent of `controller` has
// three or more poison counters.
func rfReprintBCorrupted(g *game.Game, controller uuid.UUID) bool {
	return seedcoreCorrupted(g, controller, uuid.Nil)
}
