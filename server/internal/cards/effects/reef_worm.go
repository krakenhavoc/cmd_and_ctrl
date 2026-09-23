package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reef Worm — Creature — Worm {3}{U}, 0/1 (batch 48, #459):
//
//	"When this creature dies, create a 3/3 blue Fish creature token
//	 with "When this token dies, create a 6/6 blue Whale creature
//	 token with 'When this token dies, create a 9/9 blue Kraken
//	 creature token.'""
//
// The card the "Triggered and static abilities on non-copy tokens"
// seam was worth closing for: it is nothing BUT token triggers, three
// deep, and before ADR 0083 (#1248) not one of the three could be
// written. Skipped on batch 48 for exactly that reason.
//
// Three catalog templates and no chain machinery. `token:fish` prints
// a dies-trigger that creates `token:whale`, which prints a
// dies-trigger that creates the plain 9/9 Kraken from the token
// table — each one an ordinary `WhenThisDies` found through
// `game.CatalogKey`'s token-key fallback, exactly as the Worm's own
// dies-trigger is found through its oracle ID. The recursion is in
// the CARDS, not in the engine.
//
// CR 111.7 is what makes the chain work at all: a token that dies
// reaches the graveyard and stays there long enough for the
// leaves-the-battlefield harvest to read it, and only the NEXT
// state-based check removes it (CR 704.5d, game/token_existence.go).
// So the Fish's trigger goes on the stack off a Fish that is already
// a former token, and the Whale it makes is a real permanent by the
// time anybody has priority.
//
// The Kraken is a plain 9/9 with no text, so it is a row in
// tokens_table.go and not a fourth template — a token with no ability
// needs no catalog key (token_catalog.go refuses one).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3ad5cbc-4245-4e1e-8204-509381fa0c1c",
		Name:         "Reef Worm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Reef Worm — create a 3/3 blue Fish",
				Do(CreateToken{Template: reefWormFishToken(), N: 1})),
		},
	})
}

// reefWormFishToken is the 3/3 blue Fish with "When this token dies,
// create a 6/6 blue Whale creature token with …".
func reefWormFishToken() game.Card { return tokenFromCatalog(printedReefWormFishToken) }

// reefWormWhaleToken is the 6/6 blue Whale with "When this token
// dies, create a 9/9 blue Kraken creature token."
func reefWormWhaleToken() game.Card { return tokenFromCatalog(printedReefWormWhaleToken) }

// printedReefWormFishToken is the Fish as PRINTED.
func printedReefWormFishToken() tokenTemplate {
	return tokenTemplate{
		Slug: "fish",
		Card: game.Card{
			Name:      "Fish",
			TypeLine:  "Token Creature — Fish",
			Power:     3,
			Toughness: 3,
			Colors:    []string{"U"},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Fish — create a 6/6 blue Whale",
				Do(CreateToken{Template: reefWormWhaleToken(), N: 1})),
		},
		Text: "When this token dies, create a 6/6 blue Whale creature token with " +
			"\"When this token dies, create a 9/9 blue Kraken creature token.\"",
	}
}

// printedReefWormWhaleToken is the Whale as PRINTED.
func printedReefWormWhaleToken() tokenTemplate {
	return tokenTemplate{
		Slug: "whale",
		Card: game.Card{
			Name:      "Whale",
			TypeLine:  "Token Creature — Whale",
			Power:     6,
			Toughness: 6,
			Colors:    []string{"U"},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Whale — create a 9/9 blue Kraken",
				Do(CreateToken{Template: TokenCard("9/9 blue Kraken"), N: 1})),
		},
		Text: "When this token dies, create a 9/9 blue Kraken creature token.",
	}
}
