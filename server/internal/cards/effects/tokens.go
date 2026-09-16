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

// FaerieRogueToken is kept as a function because a card passes it as a value; the data lives in tokens_table.go.
func FaerieRogueToken() game.Card { return TokenCard("1/1 colorless Faerie Rogue with flying") }

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

// PhyrexianWurmDeathtouchToken / PhyrexianWurmLifelinkToken are
// Wurmcoil Engine's two halves (CR: the dies-trigger makes one of
// each, not two of the same). Added in S21 sub-PR 1.
func PhyrexianWurmDeathtouchToken() game.Card {
	t := TokenCard("3/3 colorless Phyrexian Wurm artifact")
	t.Keywords = []string{"deathtouch"}
	return t
}

func PhyrexianWurmLifelinkToken() game.Card {
	t := TokenCard("3/3 colorless Phyrexian Wurm artifact")
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
//
// Cheaper than a Treasure by a tap: the printed cost is the sacrifice
// alone, so a Gold made this turn is spendable this turn and a tapped
// Gold is still spendable. Both fall out of leaving TapCost false.
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

// RedGoblinToken is kept as a function because a card passes it as a value; the data lives in tokens_table.go.
func RedGoblinToken() game.Card { return TokenCard("1/1 red Goblin") }

// BlackZombieToken is kept as a function because a card passes it as a value; the data lives in tokens_table.go.
func BlackZombieToken() game.Card { return TokenCard("2/2 black Zombie") }

// --- S27 templates ---------------------------------------------

// --- S27 Vehicle templates -------------------------------------

// BlueShapeshifterToken is Maskwood Nexus' 2/2 blue Shapeshifter with
// changeling, and ColorlessShapeshifterToken is Irregular Cohort's
// colourless one. Two constructors for what is otherwise the same
// 2/2, because the printed colour is the only difference and a shared
// helper named for neither would make the next reader check.
//
// The changeling keyword on a TOKEN is why Card.Keywords exists: a
// token has no oracle ID, so the catalog hook can never find it, and
// without the keyword on the template these would be Shapeshifters
// and nothing else â which is the whole of what they are for.
func BlueShapeshifterToken() game.Card {
	return game.Card{
		Name:      "Shapeshifter",
		TypeLine:  "Token Creature — Shapeshifter",
		Power:     2,
		Toughness: 2,
		Keywords:  []string{game.KeywordChangeling},
	}
}

// ColorlessShapeshifterToken is Irregular Cohort's 2/2 colourless
// Shapeshifter with changeling.
func ColorlessShapeshifterToken() game.Card {
	return game.Card{
		Name:      "Shapeshifter",
		TypeLine:  "Token Creature — Shapeshifter",
		Power:     2,
		Toughness: 2,
		Keywords:  []string{game.KeywordChangeling},
	}
}
