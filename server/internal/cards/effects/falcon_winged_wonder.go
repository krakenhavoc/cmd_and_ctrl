package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falcon, Winged Wonder — Legendary Creature — Human Hero {4}{U}, 3/4
// (Marvel Super Heroes, #4455):
//
//	"Flying
//	 Avian Telepathy — When Falcon enters, create Redwing, a legendary 1/1
//	 blue Bird Scout creature token with flying and "Whenever Redwing
//	 attacks, surveil 1.""
//
// The enter trigger is the ordinary token-making ETB. Redwing is a catalog
// token template (ADR 0083), so its attack trigger travels with the token
// and resolves as a real stack item, the same way Fable's Goblin Shaman
// does; the surveil is the shared Surveil primitive.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92e222c2-28e6-4d43-8b48-480ebcd88495",
		Name:         "Falcon, Winged Wonder",
		Completeness: CompletenessFull,

		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Falcon — create Redwing", Do(CreateToken{Template: mshA3RedwingToken(), N: 1})),
		},
	})
}

// mshA3RedwingToken is Redwing, the legendary 1/1 blue Bird Scout with
// flying and "Whenever Redwing attacks, surveil 1."
func mshA3RedwingToken() game.Card { return tokenFromCatalog(printedMshA3RedwingToken) }

// printedMshA3RedwingToken is Redwing as PRINTED.
func printedMshA3RedwingToken() tokenTemplate {
	return tokenTemplate{
		Slug: "redwing",
		Card: game.Card{
			Name:      "Redwing",
			TypeLine:  "Token Legendary Creature — Bird Scout",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"U"},
			Keywords:  []string{"flying"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Redwing — surveil 1", Do(Surveil{N: 1})),
		},
		Text: "Whenever Redwing attacks, surveil 1.",
	}
}
