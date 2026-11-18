package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const elementalMasteryGrant = "elemental-mastery/elementals"

// Elemental Mastery — Enchantment — Aura {3}{R}:
//
//	"Enchant creature
//	 Enchanted creature has "{T}: Create X 1/1 red Elemental creature
//	 tokens with haste, where X is this creature's power. Exile them at
//	 the beginning of the next end step.""
//
// Burning Anger's shape (ADR 0093): the granted row is the enchanted
// creature's own ability, so it taps that creature, its controller
// activates it, and "this creature" is the enchanted creature. X is its
// power as the ability resolves, or as it last was if it has left (CR
// 608.2h). The tokens that landed are handed to the end-step exile through
// the creation's keyed tail (Rionya, Fire Dancer's shape, CR 603.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4234f07e-583f-40db-8b47-7e16c7316e19",
		Name:         "Elemental Mastery",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Grants: []AbilityGrant{{
			Key: elementalMasteryGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}: Create X 1/1 red Elemental creature tokens with haste, where X is this creature's power. Exile them at the beginning of the next end step.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost:    TapCost(),
				Effect:  elementalMasteryTokens,
			}},
			Text: "{T}: Create X 1/1 red Elemental creature tokens with haste, where X is this creature's power. Exile them at the beginning of the next end step.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(elementalMasteryGrant)},
	})
}

func elementalMasteryTokens(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := b43PowerNowOrLastKnown(g, item.SourceCardID)
	if x <= 0 {
		return nil
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("1/1 red Elemental with haste")), Count: x}},
	}, ctx.ItemTail(elementalMasteryExile))
}

// elementalMasteryExile is "exile them at the beginning of the next end
// step". An on-disk key.
var elementalMasteryExile = socB8ExileTokensAtNextEndStep("elemental-mastery/exile-the-elementals",
	"Elemental Mastery — exile the Elementals")
