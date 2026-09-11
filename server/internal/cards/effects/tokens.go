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

// --- the artifact token cycle ------------------------------------
//
// Treasure, Food, Clue and Blood are defined entirely by the ability
// printed on them — strip it and they're blank artifacts. They carry
// it on the template (ManaAbilities for Treasure, ActivatedAbilities
// for the rest), because the catalog's hooks key on oracle ID and a
// token hasn't got one. Added across S21 sub-PRs 1 and 4.

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

// GoldToken returns a template for the Gold artifact token
// ("Sacrifice this token: Add one mana of any color"). Treasure
// without the tap — a Gold can be cracked the turn it is made, which
// is the whole point of Curse of Opulence handing them out in the
// middle of somebody else's attack.
//
// Same five-colour pipe as Treasure, so it runs through the same
// commander-identity filter rather than offering a five-way prompt in
// a mono-coloured deck.
func GoldToken() game.Card {
	return game.Card{
		Name:     "Gold",
		TypeLine: "Token Artifact — Gold",
		ManaAbilities: []game.ManaAbilityShape{{
			SacrificeCost: true,
			Produced:      "{W|U|B|R|G}",
			Label:         "Sacrifice: Add one mana of any color",
		}},
	}
}

// FoodToken — "{2}, {T}, Sacrifice this artifact: You gain 3 life."
func FoodToken() game.Card {
	return game.Card{
		Name:     "Food",
		TypeLine: "Token Artifact — Food",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{2}, {T}, Sacrifice this artifact: You gain 3 life",
			Cost: game.AbilityCost{
				Tap:           true,
				SacrificeSelf: true,
				Mana:          "{2}",
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
			},
		}},
	}
}

// ClueToken — "{2}, Sacrifice this artifact: Draw a card." No tap in
// the cost, so a Clue can be cracked the turn it's made.
func ClueToken() game.Card {
	return game.Card{
		Name:     "Clue",
		TypeLine: "Token Artifact — Clue",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{2}, Sacrifice this artifact: Draw a card",
			Cost: game.AbilityCost{
				SacrificeSelf: true,
				Mana:          "{2}",
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	}
}

// BloodToken — "{1}, {T}, Discard a card, Sacrifice this artifact:
// Draw a card."
//
// Sandbox gap: the discard component of the cost isn't modelled —
// AbilityCost has no "discard a card" field, and adding one wants
// the same card-choice plumbing a sacrifice cost has. Until then
// the Blood token loots for free, which is strictly better than
// printed. Noted rather than silently wrong.
func BloodToken() game.Card {
	return game.Card{
		Name:     "Blood",
		TypeLine: "Token Artifact — Blood",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{1}, {T}, Sacrifice this artifact: Draw a card (discard cost not yet modelled)",
			Cost: game.AbilityCost{
				Tap:           true,
				SacrificeSelf: true,
				Mana:          "{1}",
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	}
}

// PowerstoneToken — "{T}: Add {C}. This mana can't be spent to cast
// a nonartifact spell."
//
// Sandbox gap: the spend restriction isn't modelled. ManaToken
// carries a colour and a source, not a restriction, so enforcing it
// wants a restriction field on the pool plus a check at every spend
// site. The Powerstone therefore taps for unrestricted {C} — a real
// power increase over the printed card, so it stays out of any
// deck fixture until the restriction lands.
func PowerstoneToken() game.Card {
	return game.Card{
		Name:     "Powerstone",
		TypeLine: "Token Artifact — Powerstone",
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{C}",
			Label:    "{T}: Add {C} (spend restriction not yet modelled)",
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

// GreenBeastToken is the 3/3 green Beast that Beast Within hands to
// the permanent's controller — the drawback half of the card, and
// the reason it's "destroy anything" rather than pure removal.
func GreenBeastToken() game.Card {
	return game.Card{
		Name:      "Beast",
		TypeLine:  "Token Creature — Beast",
		Power:     3,
		Toughness: 3,
	}
}

// WhiteElephantToken is Generous Gift's 3/3 Elephant. Same shape as
// the Beast; kept as its own constructor so the two cards read as
// the printed cards do rather than sharing a misleading name.
func WhiteElephantToken() game.Card {
	return game.Card{
		Name:      "Elephant",
		TypeLine:  "Token Creature — Elephant",
		Power:     3,
		Toughness: 3,
	}
}

// WhiteAllyToken is Appa, Steadfast Guardian's 1/1 white Ally.
// Vanilla — no keywords — so the template carries only the printed
// fields. Like every other template in this file it leaves Colors
// unset; nothing in the engine reads a token's color yet, and
// stamping one here alone would be an untested special case.
func WhiteAllyToken() game.Card {
	return game.Card{
		Name:      "Ally",
		TypeLine:  "Token Creature — Ally",
		Power:     1,
		Toughness: 1,
	}
}

// --- S27 templates ---------------------------------------------

// KnightVigilanceToken is History of Benalia's 2/2 white Knight with
// vigilance. The subtype matters as much as the stats: the Saga's
// third chapter pumps "Knights you control", which reads the
// effective subtype off exactly this type line.
func KnightVigilanceToken() game.Card {
	return game.Card{
		Name:      "Knight",
		TypeLine:  "Token Creature — Knight",
		Power:     2,
		Toughness: 2,
		Keywords:  []string{"vigilance"},
	}
}

// HumanSoldierToken is The First Iroan Games' 1/1 white Human
// Soldier. Vanilla.
func HumanSoldierToken() game.Card {
	return game.Card{
		Name:      "Human Soldier",
		TypeLine:  "Token Creature — Human Soldier",
		Power:     1,
		Toughness: 1,
	}
}

// WallDefenderToken is The Birth of Meletis' 0/4 colorless Wall
// artifact creature with defender.
//
// Its 0 power is the reason the toughness has to be written down:
// the CR 704.5f state-based action skips a creature printed 0/0 with
// no counters (the demo-seed placeholder convention documented on
// Card.Power), and a 0/4 must not be mistaken for that case.
func WallDefenderToken() game.Card {
	return game.Card{
		Name:      "Wall",
		TypeLine:  "Token Artifact Creature — Wall",
		Power:     0,
		Toughness: 4,
		Keywords:  []string{"defender"},
	}
}

// GoldToken is The First Iroan Games' Gold artifact token —
// "Sacrifice this token: Add one mana of any color".
//
// Cheaper than a Treasure by a tap: the printed cost is the
// sacrifice alone, so a Gold made this turn is spendable this turn
// and a tapped Gold is still spendable. Both fall out of leaving
// TapCost false, which is the whole difference between the two
// templates.
func GoldToken() game.Card {
	return game.Card{
		Name:     "Gold",
		TypeLine: "Token Artifact — Gold",
		ManaAbilities: []game.ManaAbilityShape{{
			SacrificeCost: true,
			Produced:      "{W|U|B|R|G}",
			Label:         "Sacrifice: Add one mana of any color",
		}},
	}
}

// GoldToken's sibling templates for the other S27 card types
// (Esika's Chariot's Cats, Parhelion II's Angels, Elspeth's Soldiers)
// live with the sub-PRs that use them, so each lands with the card
// that reads it.
