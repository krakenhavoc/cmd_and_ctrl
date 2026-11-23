package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_neo_b20_tokens.go — the three tokens S82 NEO batch 20 makes, each
// printed by one card of the batch and each carrying an ability of its own
// (ADR 0083), so each is a catalog template and not a tokens_table.go row.

// KeimiToken is Tatsunari, Toad Rider's Keimi: a legendary 3/3 black and
// green Frog with "Whenever you cast an enchantment spell, each opponent
// loses 1 life and you gain 1 life." "You" is the token's controller.
func KeimiToken() game.Card { return TokenFromCatalog(printedNeoB20KeimiToken) }

func printedNeoB20KeimiToken() TokenTemplate {
	return TokenTemplate{
		Slug: "keimi",
		Card: game.Card{
			Name:      "Keimi",
			TypeLine:  "Token Legendary Creature — Frog",
			Power:     3,
			Toughness: 3,
			Colors:    []string{"B", "G"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Enchantment(),
				"Keimi — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
		Text: "Whenever you cast an enchantment spell, each opponent loses 1 life and you gain 1 life.",
	}
}

// TamiyosNotebookToken is Tamiyo, Compleated Sage's legendary colorless
// Book artifact token: "Spells you cast cost {2} less to cast." and "{T}:
// Draw a card."
func TamiyosNotebookToken() game.Card { return TokenFromCatalog(printedNeoB20NotebookToken) }

func printedNeoB20NotebookToken() TokenTemplate {
	return TokenTemplate{
		Slug: "tamiyos-notebook",
		Card: game.Card{
			Name:     "Tamiyo's Notebook",
			TypeLine: "Token Legendary Artifact — Book",
		},
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Spells you cast cost {2} less to cast.", YourSpell()),
		},
		Activated: []game.ActivatedAbilityShape{{
			Label:   "{T}: Draw a card.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
		Text: "Spells you cast cost {2} less to cast.\n{T}: Draw a card.",
	}
}

// NeoB20NinjaToken is Kaito Shizuki's 1/1 blue Ninja creature token with
// "This token can't be blocked."
func NeoB20NinjaToken() game.Card { return TokenFromCatalog(printedNeoB20NinjaToken) }

func printedNeoB20NinjaToken() TokenTemplate {
	return TokenTemplate{
		Slug: "ninja-unblockable",
		Card: game.Card{
			Name:      "Ninja",
			TypeLine:  "Token Creature — Ninja",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"U"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Text:   "This creature can't be blocked.",
	}
}
