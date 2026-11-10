package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Toby, Beastie Befriender — Legendary Creature — Human Wizard {2}{W},
// 1/1 (Duskmourn: House of Horror):
//
//	"When Toby enters, create a 4/4 white Beast creature token with
//	 "This token can't attack or block alone."
//	 As long as you control four or more creature tokens, creature
//	 tokens you control have flying."
//
// The Beast is a catalog token (tobyBeastToken) that carries both of its
// printed restrictions, so a restore keeps them: CantAttackAlone
// (game.AttackLimitNotAlone, CR 506.5 and 508.1c) and CantBlockAlone
// (game.BlockRule.NotAlone, CR 506.5 and 509.1b, #5396). Two creatures
// that each can't attack or block alone may attack or block together,
// and the other attacker or blocker need not be on the same player or
// attacker, as the rulings say.
//
// The anthem is a layer 6 grant over the creature tokens Toby's
// controller controls, applied only while they control four or more
// creature tokens, counted per recompute. Toby is not a token, so it
// neither counts nor flies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4c27c89-fa50-4cf0-bf37-039732d16fef",
		Name:         "Toby, Beastie Befriender",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Toby — create a 4/4 white Beast that can't attack or block alone",
				Do(CreateToken{Template: tobyBeastToken(), N: 1})),
		},
		Static: []game.StaticAbility{{
			Label: "As long as you control four or more creature tokens, creature tokens you control have flying.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return creatureTokensYouControl(target, g, source) && tobyCreatureTokens(g, source.Controller) >= 4
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "flying")
			},
		}},
	})
}

// tobyCreatureTokens counts the creature tokens `controller` controls,
// post-layer types.
func tobyCreatureTokens(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsToken() && c.IsCreature() {
			n++
		}
	}
	return n
}

// tobyBeastToken is the 4/4 white Beast creature token with "This token
// can't attack or block alone."
func tobyBeastToken() game.Card { return tokenFromCatalog(printedTobyBeastToken) }

// printedTobyBeastToken is the Beast as PRINTED (the Duskmourn token): not
// the plain "4/4 green Beast" row, which prints no text and is green.
func printedTobyBeastToken() tokenTemplate {
	return tokenTemplate{
		Slug: "beast-not-alone",
		Card: game.Card{
			Name:      "Beast",
			TypeLine:  "Token Creature — Beast",
			Power:     4,
			Toughness: 4,
			Colors:    []string{"W"},
		},
		AttackLimits: []game.AttackLimit{CantAttackAlone()},
		BlockRules:   []game.BlockRule{CantBlockAlone()},
		Text:         "This token can't attack or block alone.",
	}
}
