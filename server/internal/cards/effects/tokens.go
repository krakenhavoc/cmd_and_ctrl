package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tokens.go holds token templates used by CreateToken primitives.
// Each helper returns a game.Card with the "printed" fields set;
// CreateTokenForEffect stamps fresh InstanceID / Owner / Controller
// / KnownBy at creation time.
//
// Tokens have an empty ScryfallID because they're not a Scryfall
// printing. The client renders them via the name-fallback path
// (no image lookup). A future extension could key tokens by a
// synthetic Scryfall-like ID to wire up art lookup — deferred.

// WhiteBirdToken returns a template for the Swan Song 2/2 bird
// token ("Return the White Rose Bird" — creature token with flying).
// Flying is cosmetic in S14 (keyword pipeline is S18); the token
// enters as a vanilla 2/2 for combat purposes.
func WhiteBirdToken() game.Card {
	return game.Card{
		Name:      "Bird",
		TypeLine:  "Token Creature — Bird",
		Power:     2,
		Toughness: 2,
	}
}

// WhiteSamuraiToken returns a template for The Wandering Emperor's
// 2/2 Samurai token (vigilance in the real card; cosmetic in S14).
// Unused by S14 catalog (the Emperor's +1 loyalty ability stays
// manual this sprint), but the template lands now so S19's
// ability-auto-fire flow can wire up to the existing CreateToken
// primitive without growing a new template file.
func WhiteSamuraiToken() game.Card {
	return game.Card{
		Name:      "Samurai",
		TypeLine:  "Token Creature — Samurai",
		Power:     2,
		Toughness: 2,
	}
}
