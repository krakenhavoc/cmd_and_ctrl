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
// S21 sub-PR 1 put a token's abilities on the template itself, and
// #521 moved the ability half into the catalog. `Keywords` are
// printed keyword abilities (flying, deathtouch) and stay here: they
// are plain data, they always survived a snapshot, and
// printedCharacteristic folds them into the layer engine. The mana
// and activated abilities are now catalog entries under a synthetic
// token key (token_catalog.go), so ManaAbilitiesForCard and
// ActivatedAbilitiesForCard find them the ordinary way — by catalog
// key, exactly as they do for a printed card.

// FaerieRogueToken is kept as a function because a card passes it as a value; the data lives in tokens_table.go.
func FaerieRogueToken() game.Card { return TokenCard("1/1 black Faerie Rogue with flying") }

// EldraziSpawnToken returns a template for Awakening Zone's 0/1
// colorless Eldrazi Spawn token with "Sacrifice this creature: Add
// {C}". Added in S19 sub-PR 5; the mana ability went live in S21
// sub-PR 1 (no tap in the cost — a Spawn can be cracked the turn it
// arrives).
func EldraziSpawnToken() game.Card { return tokenFromCatalog(printedEldraziSpawnToken) }

// printedEldraziSpawnToken is the Eldrazi Spawn as PRINTED — the abilities
// included. It is the catalog's entry for this token
// (token_catalog.go): the abilities are registered from here at boot,
// and the template that reaches the battlefield carries the key that
// finds them rather than the closures themselves.
func printedEldraziSpawnToken() tokenTemplate {
	return tokenTemplate{
		Slug: "eldrazi-spawn",
		Card: game.Card{
			Name:      "Eldrazi Spawn",
			TypeLine:  "Token Creature — Eldrazi Spawn",
			Power:     0,
			Toughness: 1,
		},
		Mana: []game.ManaAbilityShape{{
			SacrificeCost: true,
			Produced:      "{C}",
			Label:         "Sacrifice this creature: Add {C}",
		}},
		Text: "Sacrifice this token: Add {C}.",
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
// printed on them — strip it and they're blank artifacts. Added
// across S21 sub-PRs 1 and 4, when the ability rode the template as
// a closure because the catalog's hooks keyed on oracle ID and a
// token hasn't got one.
//
// #521 gave them one: each is registered in the catalog under a
// synthetic token key ("token:treasure"), and the constructor hands
// out a template carrying that key. The ability is catalog data like
// a printed card's, which is what lets a snapshot write a board with
// a Treasure on it — before this, one live Treasure meant no restore
// point was written at all until it was spent.

// TreasureToken returns a template for the Treasure artifact token
// ("{T}, Sacrifice this artifact: Add one mana of any color").
// Added in S19 sub-PR 6 for Smothering Tithe; the sac-for-mana
// ability went live in S21 sub-PR 1, closing the S19 deferral.
//
// The five-colour pipe offers all five colours, the commander's
// identity listed first, so cracking a Treasure in a mono-red deck
// offers {R} first and the other four after it, as printed.
func TreasureToken() game.Card { return tokenFromCatalog(printedTreasureToken) }

// printedTreasureToken is the Treasure as PRINTED — the abilities
// included. It is the catalog's entry for this token
// (token_catalog.go): the abilities are registered from here at boot,
// and the template that reaches the battlefield carries the key that
// finds them rather than the closures themselves.
func printedTreasureToken() tokenTemplate {
	return tokenTemplate{
		Slug: "treasure",
		Card: game.Card{
			Name:     "Treasure",
			TypeLine: "Token Artifact — Treasure",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:       true,
			SacrificeCost: true,
			Produced:      "{W|U|B|R|G}",
			Label:         "{T}, Sacrifice: Add one mana of any color",
		}},
		Text: "{T}, Sacrifice this token: Add one mana of any color.",
	}
}

// GoldToken returns a template for the Gold artifact token
// ("Sacrifice this token: Add one mana of any color"). Treasure
// without the tap — a Gold can be cracked the turn it is made, which
// is the whole point of Curse of Opulence handing them out in the
// middle of somebody else's attack.
//
// Same five-colour pipe as Treasure: all five colours, commander
// identity first.
//
// Cheaper than a Treasure by a tap: the printed cost is the sacrifice
// alone, so a Gold made this turn is spendable this turn and a tapped
// Gold is still spendable. Both fall out of leaving TapCost false.
func GoldToken() game.Card { return tokenFromCatalog(printedGoldToken) }

// printedGoldToken is the Gold as PRINTED — the abilities
// included. It is the catalog's entry for this token
// (token_catalog.go): the abilities are registered from here at boot,
// and the template that reaches the battlefield carries the key that
// finds them rather than the closures themselves.
func printedGoldToken() tokenTemplate {
	return tokenTemplate{
		Slug: "gold",
		Card: game.Card{
			Name:     "Gold",
			TypeLine: "Token Artifact — Gold",
		},
		Mana: []game.ManaAbilityShape{{
			SacrificeCost: true,
			Produced:      "{W|U|B|R|G}",
			Label:         "Sacrifice: Add one mana of any color",
		}},
		Text: "Sacrifice this token: Add one mana of any color.",
	}
}

// FoodToken — "{2}, {T}, Sacrifice this artifact: You gain 3 life."
func FoodToken() game.Card { return tokenFromCatalog(printedFoodToken) }

// printedFoodToken is the Food as PRINTED — the abilities
// included. It is the catalog's entry for this token
// (token_catalog.go): the abilities are registered from here at boot,
// and the template that reaches the battlefield carries the key that
// finds them rather than the closures themselves.
func printedFoodToken() tokenTemplate {
	return tokenTemplate{
		Slug: "food",
		Card: game.Card{
			Name:     "Food",
			TypeLine: "Token Artifact — Food",
		},
		Text: "{2}, {T}, Sacrifice this token: You gain 3 life.",
		Activated: []game.ActivatedAbilityShape{{
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
func ClueToken() game.Card { return tokenFromCatalog(printedClueToken) }

// printedClueToken is the Clue as PRINTED — the abilities
// included. It is the catalog's entry for this token
// (token_catalog.go): the abilities are registered from here at boot,
// and the template that reaches the battlefield carries the key that
// finds them rather than the closures themselves.
func printedClueToken() tokenTemplate {
	return tokenTemplate{
		Slug: "clue",
		Card: game.Card{
			Name:     "Clue",
			TypeLine: "Token Artifact — Clue",
		},
		Text: "{2}, Sacrifice this token: Draw a card.",
		Activated: []game.ActivatedAbilityShape{{
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
func BloodToken() game.Card { return tokenFromCatalog(printedBloodToken) }

// printedBloodToken is the Blood as PRINTED — the abilities
// included. It is the catalog's entry for this token
// (token_catalog.go): the abilities are registered from here at boot,
// and the template that reaches the battlefield carries the key that
// finds them rather than the closures themselves.
func printedBloodToken() tokenTemplate {
	return tokenTemplate{
		Slug: "blood",
		Card: game.Card{
			Name:     "Blood",
			TypeLine: "Token Artifact — Blood",
		},
		Text: "{1}, {T}, Discard a card, Sacrifice this token: Draw a card.",
		Activated: []game.ActivatedAbilityShape{{
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
func PowerstoneToken() game.Card { return tokenFromCatalog(printedPowerstoneToken) }

// printedPowerstoneToken is the Powerstone as PRINTED — the abilities
// included. It is the catalog's entry for this token
// (token_catalog.go): the abilities are registered from here at boot,
// and the template that reaches the battlefield carries the key that
// finds them rather than the closures themselves.
func printedPowerstoneToken() tokenTemplate {
	return tokenTemplate{
		Slug: "powerstone",
		Card: game.Card{
			Name:     "Powerstone",
			TypeLine: "Token Artifact — Powerstone",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{C}",
			Label:    "{T}: Add {C} (spend restriction not yet modelled)",
		}},
		Text: "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.",
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
		Colors:    []string{"U"},
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

// --- tokens that print an ability of their own (ADR 0083) --------
//
// A token with a TRIGGER is a catalog template like Treasure and
// Food, for the same reason and through the same key: the ability is
// registered under game.TokenKey(slug) and the template that reaches
// the battlefield carries only the key. What #521 could not do, and
// ADR 0083 does, is declare the trigger at all — a game.Card has no
// slot for one, and putting a closure back on the instance is exactly
// what the token key exists to prevent.
//
// The two below live here rather than beside a card file because each
// is printed by more than one card, and it is the same token object
// whichever card made it.

// PestToken is Beledros Witherbloom's and Sedgemoor Witch's 1/1 black
// and green Pest with "When this token dies, you gain 1 life." — the
// example the "Triggered and static abilities on non-copy tokens" row
// of docs/engine-seams.md was written around, and the first token in
// the catalog to carry a triggered ability.
func PestToken() game.Card { return tokenFromCatalog(printedPestToken) }

// printedPestToken is the Pest as PRINTED — the trigger included. It
// is the catalog's entry for this token (token_catalog.go): the
// ability is registered from here at boot, and the template that
// reaches the battlefield carries the key that finds it.
//
// "You" in a token's own trigger is the TOKEN's controller, which is
// item.Controller on the triggered ability's stack item — the same
// read a printed card's trigger makes, because a token's trigger is a
// printed ability and uses the stack like any other (CR 603.3). It is
// emphatically not the controller of whatever created the token: a
// Pest that changes hands gains life for its new controller.
func printedPestToken() tokenTemplate {
	return tokenTemplate{
		Slug: "pest",
		Card: game.Card{
			Name:      "Pest",
			TypeLine:  "Token Creature — Pest",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"B", "G"},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Pest — you gain 1 life", func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
			}),
		},
		Text: "When this token dies, you gain 1 life.",
	}
}

// NoncreatureCastWizardToken is Cornered by Black Mages', Mysidian
// Elder's and Circle of Power's 0/1 black Wizard with "Whenever you
// cast a noncreature spell, this token deals 1 damage to each
// opponent."
//
// The damage's SOURCE is the token, which is why the ability has to
// live on the token: Cornered by Black Mages is a sorcery and is in
// the graveyard by the time anything else is cast, so there is
// nothing for it to carry the trigger on behalf of — the card file
// said exactly that before this shipped.
func NoncreatureCastWizardToken() game.Card {
	return tokenFromCatalog(printedNoncreatureCastWizardToken)
}

// printedNoncreatureCastWizardToken is that Wizard as PRINTED.
func printedNoncreatureCastWizardToken() tokenTemplate {
	return tokenTemplate{
		Slug: "wizard",
		Card: game.Card{
			Name:      "Wizard",
			TypeLine:  "Token Creature — Wizard",
			Power:     0,
			Toughness: 1,
			Colors:    []string{"B"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Wizard — 1 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					return damageToEachOpponent(g, item, 1)
				}),
		},
		Text: "Whenever you cast a noncreature spell, this token deals 1 damage to each opponent.",
	}
}
