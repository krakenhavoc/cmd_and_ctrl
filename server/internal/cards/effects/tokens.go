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

// SpiritToken returns a template for Doomed Traveler's 1/1 white
// Spirit token with flying. Flying is cosmetic (token keywords
// aren't mechanically enforced yet — see file header); the token
// enters as a vanilla 1/1 for combat. Added in S19 sub-PR 4.
func SpiritToken() game.Card {
	return game.Card{
		Name:      "Spirit",
		TypeLine:  "Token Creature — Spirit",
		Power:     1,
		Toughness: 1,
	}
}

// PhyrexianWurmToken returns a template for Wurmcoil Engine's 3/3
// colorless Phyrexian Wurm tokens. The real card makes two distinct
// tokens — one with deathtouch, one with lifelink — but token
// keywords are cosmetic in the sandbox, so both halves use this
// single 3/3 template (Wurmcoil's Build creates two of it). The
// deathtouch / lifelink split lands when token keywords go live.
// Added in S19 sub-PR 4.
func PhyrexianWurmToken() game.Card {
	return game.Card{
		Name:      "Phyrexian Wurm",
		TypeLine:  "Token Artifact Creature — Phyrexian Wurm",
		Power:     3,
		Toughness: 3,
	}
}
