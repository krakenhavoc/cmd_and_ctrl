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
//
// S21 sub-PR 1: a token's abilities live on the template itself.
// `Keywords` are printed keyword abilities (flying, deathtouch) —
// the catalog's PrintedKeywords hook keys on oracle ID, which a
// token doesn't have, so before S21 every token's flying was
// cosmetic. `ManaAbilities` are the token's own mana abilities
// (Treasure, Eldrazi Spawn). Both flow through the ordinary engine
// paths: printedCharacteristic folds Keywords into the layer
// engine, ManaAbilitiesForCard prefers the intrinsic list.

// WhiteBirdToken returns a template for the Swan Song 2/2 Bird
// token with flying. S21 sub-PR 1: the flying is real now.
func WhiteBirdToken() game.Card {
	return game.Card{
		Name:      "Bird",
		TypeLine:  "Token Creature — Bird",
		Power:     2,
		Toughness: 2,
		Keywords:  []string{"flying"},
	}
}

// WhiteSamuraiToken returns a template for The Wandering Emperor's
// 2/2 Samurai token with vigilance (real since S21 sub-PR 1).
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
		Keywords:  []string{"vigilance"},
	}
}

// SpiritToken returns a template for Doomed Traveler's 1/1 white
// Spirit token with flying. Added in S19 sub-PR 4; the flying
// became real in S21 sub-PR 1.
func SpiritToken() game.Card {
	return game.Card{
		Name:      "Spirit",
		TypeLine:  "Token Creature — Spirit",
		Power:     1,
		Toughness: 1,
		Keywords:  []string{"flying"},
	}
}

// FaerieRogueToken returns a template for Bitterblossom's 1/1 black
// Faerie Rogue token with flying (real since S21 sub-PR 1). Added
// in S19 sub-PR 5.
func FaerieRogueToken() game.Card {
	return game.Card{
		Name:      "Faerie Rogue",
		TypeLine:  "Token Creature — Faerie Rogue",
		Power:     1,
		Toughness: 1,
		Keywords:  []string{"flying"},
	}
}

// EldraziSpawnToken returns a template for Awakening Zone's 0/1
// colorless Eldrazi Spawn token with "Sacrifice this creature: Add
// {C}". Added in S19 sub-PR 5; the mana ability went live in S21
// sub-PR 1 (no tap in the cost — a Spawn can be cracked the turn it
// arrives).
func EldraziSpawnToken() game.Card {
	return game.Card{
		Name:      "Eldrazi Spawn",
		TypeLine:  "Token Creature — Eldrazi Spawn",
		Power:     0,
		Toughness: 1,
		ManaAbilities: []game.ManaAbilityShape{{
			SacrificeCost: true,
			Produced:      "{C}",
			Label:         "Sacrifice this creature: Add {C}",
		}},
	}
}

// PhyrexianWurmToken returns a template for Wurmcoil Engine's 3/3
// colorless Phyrexian Wurm tokens. S19 shipped one shared vanilla
// template because token keywords were cosmetic; S21 sub-PR 1 makes
// them real, so the two halves are now distinct — see
// PhyrexianWurmDeathtouchToken / PhyrexianWurmLifelinkToken. This
// base template stays as the shared shape.
func PhyrexianWurmToken() game.Card {
	return game.Card{
		Name:      "Phyrexian Wurm",
		TypeLine:  "Token Artifact Creature — Phyrexian Wurm",
		Power:     3,
		Toughness: 3,
	}
}

// PhyrexianWurmDeathtouchToken / PhyrexianWurmLifelinkToken are
// Wurmcoil Engine's two halves (CR: the dies-trigger makes one of
// each, not two of the same). Added in S21 sub-PR 1.
func PhyrexianWurmDeathtouchToken() game.Card {
	t := PhyrexianWurmToken()
	t.Keywords = []string{"deathtouch"}
	return t
}

func PhyrexianWurmLifelinkToken() game.Card {
	t := PhyrexianWurmToken()
	t.Keywords = []string{"lifelink"}
	return t
}

// TreasureToken returns a template for the Treasure artifact token
// ("{T}, Sacrifice this artifact: Add one mana of any color").
// Added in S19 sub-PR 6 for Smothering Tithe; the sac-for-mana
// ability went live in S21 sub-PR 1, closing the S19 deferral.
//
// The five-colour pipe runs through the same commander-identity
// filter Arcane Signet uses, so cracking a Treasure in a mono-red
// deck offers {R} rather than a five-way prompt.
func TreasureToken() game.Card {
	return game.Card{
		Name:     "Treasure",
		TypeLine: "Token Artifact — Treasure",
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost:       true,
			SacrificeCost: true,
			Produced:      "{W|U|B|R|G}",
			Label:         "{T}, Sacrifice: Add one mana of any color",
		}},
	}
}

// RedGoblinToken returns a template for the 1/1 red Goblin token
// Krenko, Mob Boss makes. Added in S21 sub-PR 2.
func RedGoblinToken() game.Card {
	return game.Card{
		Name:      "Goblin",
		TypeLine:  "Token Creature — Goblin",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"R"},
	}
}
