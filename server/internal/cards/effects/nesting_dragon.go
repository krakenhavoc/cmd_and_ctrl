package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nesting Dragon — Creature — Dragon {3}{R}{R}, 5/4 (batch 33, #397):
//
//	"Flying
//	 Landfall — Whenever a land you control enters, create a 0/2 red
//	 Dragon Egg creature token with defender and "When this token
//	 dies, create a 2/2 red Dragon creature token with flying and
//	 '{R}: This token gets +1/+0 until end of turn.'""
//
// Skipped on batch 33 for the "Triggered and static abilities on
// non-copy tokens" seam; it ships with ADR 0083 (#1248).
//
// The card is worth having as a proof because its two tokens use
// DIFFERENT slots of the one declaration. The Egg is printed
// characteristics (`defender` is a keyword and plain data on the
// Card) plus a TRIGGERED slot; the Dragon it hatches into is printed
// characteristics (`flying`) plus an ACTIVATED slot. Both are found
// by the ordinary accessors through `game.CatalogKey`'s token-key
// fallback, so neither needed a dispatch path of its own — and the
// Dragon's firebreathing proves an activated ability composes with a
// trigger on the same token rather than replacing it, which is the
// question #521's two-slot registration could not be asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0acc9372-58b4-43bf-ab82-1f95831c81d4",
		Name:            "Nesting Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Landfall("Nesting Dragon — create a 0/2 red Dragon Egg",
				Do(CreateToken{Template: dragonEggToken(), N: 1})),
		},
	})
}

// dragonEggToken is the 0/2 red Dragon Egg with defender and "When
// this token dies, create a 2/2 red Dragon …".
func dragonEggToken() game.Card { return tokenFromCatalog(printedDragonEggToken) }

// nestingDragonDragonToken is the 2/2 red Dragon with flying and
// "{R}: This token gets +1/+0 until end of turn."
func nestingDragonDragonToken() game.Card { return tokenFromCatalog(printedNestingDragonDragonToken) }

// printedDragonEggToken is the Dragon Egg as PRINTED.
func printedDragonEggToken() tokenTemplate {
	return tokenTemplate{
		Slug: "dragon-egg",
		Card: game.Card{
			Name:      "Dragon Egg",
			TypeLine:  "Token Creature — Dragon",
			Power:     0,
			Toughness: 2,
			Colors:    []string{"R"},
			Keywords:  []string{"defender"},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Dragon Egg — create a 2/2 red Dragon",
				Do(CreateToken{Template: nestingDragonDragonToken(), N: 1})),
		},
		Text: "Defender\nWhen this token dies, create a 2/2 red Dragon creature token with flying and " +
			"\"{R}: This token gets +1/+0 until end of turn.\"",
	}
}

// printedNestingDragonDragonToken is the hatched Dragon as PRINTED.
//
// The pump targets the ABILITY's source rather than a chosen
// permanent: "this token" on a token is the same self-reference "this
// creature" is on a printed card, and item.SourceCardID is the
// activating permanent for an activated ability (CR 113.7a).
func printedNestingDragonDragonToken() tokenTemplate {
	return tokenTemplate{
		Slug: "dragon-nesting",
		Card: game.Card{
			Name:      "Dragon",
			TypeLine:  "Token Creature — Dragon",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"R"},
			Keywords:  []string{"flying"},
		},
		Activated: []game.ActivatedAbilityShape{{
			Label: "{R}: This token gets +1/+0 until end of turn",
			Cost:  game.AbilityCost{Mana: "{R}"},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{
					Target: ctx.Source(),
					Power:  1,
					Label:  "Dragon — +1/+0",
				}.Apply(ctx)
			},
		}},
		Text: "Flying\n{R}: This token gets +1/+0 until end of turn.",
	}
}
