package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b92_tokens.go — the two tokens Fugitive of the Judoon makes (Doctor
// Who, #5948). Catalog templates in their own file rather than rows of
// tokens_table.go, so the batch adds nothing to a shared table; the slugs are
// on-disk identities: never renamed, never reused.

// gapB92WardedHumanToken is "a 1/1 white Human creature token with ward {2}".
func gapB92WardedHumanToken() game.Card { return tokenFromCatalog(printedGapB92WardedHumanToken) }

func printedGapB92WardedHumanToken() tokenTemplate {
	return tokenTemplate{
		Slug: "human-ward-2",
		Card: game.Card{
			Name:      "Human",
			TypeLine:  "Token Creature — Human",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"W"},
		},
		Triggered: []game.TriggeredAbility{Ward(WardMana("{2}"), "Human — ward {2}")},
		Text:      "Ward {2}",
	}
}

// gapB92AlienRhinoToken is "a 4/4 white Alien Rhino creature token".
func gapB92AlienRhinoToken() game.Card { return tokenFromCatalog(printedGapB92AlienRhinoToken) }

func printedGapB92AlienRhinoToken() tokenTemplate {
	return tokenTemplate{
		Slug: "alien-rhino",
		Card: game.Card{
			Name:      "Alien Rhino",
			TypeLine:  "Token Creature — Alien Rhino",
			Power:     4,
			Toughness: 4,
			Colors:    []string{"W"},
		},
	}
}
