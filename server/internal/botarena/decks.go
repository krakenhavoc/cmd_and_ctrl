package botarena

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// decks.go is the two decks an arena seat can be dealt: a synthetic
// one that needs nothing, and a curated one that needs a Scryfall
// dump.
//
// Both are here rather than in the caller because which deck a seat
// plays is a property of the MEASUREMENT, not of the harness: a win
// rate measured on BattleDeck and a win rate measured on
// mono-red-aggro-vs-esper-control are two different numbers, and a
// report that cannot say which one it took is not a report.

// BattleDeck is 65 vanilla red cards with a real curve, some evasion
// and a little burn — the deck the whole-game tests have used since
// S31 opened, moved here verbatim so the arena and those tests
// measure the same thing.
//
// Why it looks the way it does, because both halves were paid for:
//
// The SIZE. A 31-card deck ends every four-bot game the same way —
// somebody decks themselves around round 25 and the survivors win on
// 30 life. That is a fine fuzzer and a useless measurement, because
// it scores mulligan luck rather than play. 65 cards with a curve put
// the loss condition back where a policy can reach it.
//
// The EVASION. Four competent heuristics on symmetric vanilla ground
// creatures produce a perfect board stall: every attack is blocked,
// every trade is even, no life total moves, and the library decides
// it. That is the correct play of a bad deck rather than a bot that
// cannot finish — but it measures nothing. Fliers and trample are
// what every real Commander deck has, and what lets a board
// advantage convert into a win.
//
// The copy in aiseat/heuristic_game_test.go is deliberately left
// where it is. Those tests may not import this package (it holds
// internal/game, which every package under aiseat/ is banned from),
// and the duplication is a hundred lines of literal card data that
// has not changed since it was written.
func BattleDeck(owner uuid.UUID) []game.Card {
	deck := []game.Card{}
	cmdr := game.NewCommander("Commander Bear", owner)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{2}{R}"
	cmdr.Power, cmdr.Toughness = 3, 3
	deck = append(deck, cmdr)
	add := func(n int, build func() game.Card) {
		for i := 0; i < n; i++ {
			deck = append(deck, build())
		}
	}
	add(24, func() game.Card {
		c := game.NewCard("Mountain", owner)
		c.TypeLine = "Basic Land — Mountain"
		return c
	})
	add(12, func() game.Card {
		c := game.NewCard("Bear", owner)
		c.TypeLine = "Creature — Bear"
		c.ManaCost = "{1}{R}"
		c.Power, c.Toughness = 2, 2
		return c
	})
	add(8, func() game.Card {
		c := game.NewCard("Ogre", owner)
		c.TypeLine = "Creature — Ogre"
		c.ManaCost = "{3}{R}"
		c.Power, c.Toughness = 4, 4
		return c
	})
	add(8, func() game.Card {
		c := game.NewCard("Drake", owner)
		c.TypeLine = "Creature — Drake"
		c.ManaCost = "{2}{R}"
		c.Power, c.Toughness = 3, 3
		c.Keywords = []string{"flying"}
		return c
	})
	add(6, func() game.Card {
		c := game.NewCard("Wurm", owner)
		c.TypeLine = "Creature — Wurm"
		c.ManaCost = "{5}{R}"
		c.Power, c.Toughness = 7, 7
		c.Keywords = []string{"trample"}
		return c
	})
	add(8, func() game.Card {
		c := game.NewCard("Lightning Bolt", owner)
		c.TypeLine = "Instant"
		c.ManaCost = "{R}"
		c.OracleID = oracleLightningBolt
		return c
	})
	return deck
}

// oracleLightningBolt is the catalog key the engine resolves the burn
// spell through. Without it the card is an instant that does nothing,
// which would quietly remove the only non-combat interaction in the
// deck.
const oracleLightningBolt = "4457ed35-7c10-48c8-9776-456485fdf070"

// ExertDeckID is the arena deck id of ExertBattleDeck: `--decks
// exert-battle`. Synthetic, so it needs no Scryfall dump.
const ExertDeckID = "exert-battle"

// ExertBattleDeck is BattleDeck in red and white with fifteen of the
// catalogued exert cards in it: ADR 0130 §9's small arena check (ADR
// 0126 §8), since no curated deck holds an exert card. It is
// BattleDeck's curve with some of its vanilla creatures and burn swapped
// for Oketra's Avenger, Combat Celebrant, Glorybringer and Resolute
// Survivors, dealt by oracle ID so the catalog gives each its abilities.
// The printed costs and stats are the cards' own.
func ExertBattleDeck(owner uuid.UUID) []game.Card {
	deck := []game.Card{}
	cmdr := game.NewCommander("Commander Bear", owner)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{2}{R}"
	cmdr.Power, cmdr.Toughness = 3, 3
	deck = append(deck, cmdr)
	add := func(n int, build func() game.Card) {
		for i := 0; i < n; i++ {
			deck = append(deck, build())
		}
	}
	vanilla := func(name, typeLine, cost string, p, t int, kw ...string) func() game.Card {
		return func() game.Card {
			c := game.NewCard(name, owner)
			c.TypeLine, c.ManaCost = typeLine, cost
			c.Power, c.Toughness = p, t
			c.Keywords = kw
			return c
		}
	}
	exert := func(name, oracle, typeLine, cost string, p, t int) func() game.Card {
		return func() game.Card {
			c := game.NewCard(name, owner)
			c.OracleID = oracle
			c.TypeLine, c.ManaCost = typeLine, cost
			c.Power, c.Toughness = p, t
			return c
		}
	}
	add(12, func() game.Card {
		c := game.NewCard("Mountain", owner)
		c.TypeLine = "Basic Land — Mountain"
		return c
	})
	add(12, func() game.Card {
		c := game.NewCard("Plains", owner)
		c.TypeLine = "Basic Land — Plains"
		return c
	})
	add(9, vanilla("Bear", "Creature — Bear", "{1}{R}", 2, 2))
	add(4, vanilla("Ogre", "Creature — Ogre", "{3}{R}", 4, 4))
	add(4, vanilla("Drake", "Creature — Drake", "{2}{R}", 3, 3, "flying"))
	add(4, vanilla("Wurm", "Creature — Wurm", "{5}{R}", 7, 7, "trample"))
	add(6, func() game.Card {
		c := game.NewCard("Lightning Bolt", owner)
		c.TypeLine = "Instant"
		c.ManaCost = "{R}"
		c.OracleID = oracleLightningBolt
		return c
	})
	add(4, exert("Oketra's Avenger", oracleOketrasAvenger, "Creature — Human Warrior", "{1}{W}", 3, 1))
	add(4, exert("Combat Celebrant", oracleCombatCelebrant, "Creature — Human Warrior", "{2}{R}", 4, 1))
	add(3, exert("Glorybringer", oracleGlorybringer, "Creature — Dragon", "{3}{R}{R}", 4, 4))
	add(4, exert("Resolute Survivors", oracleResoluteSurvivors, "Creature — Human Warrior", "{1}{R}{W}", 3, 3))
	return deck
}

// The exert cards' catalog keys (ADR 0130 §11).
const (
	oracleOketrasAvenger    = "8f064160-3afe-408a-85b4-b335eae8571c"
	oracleCombatCelebrant   = "5e15ff93-99a0-4000-918e-4bd2c257188d"
	oracleGlorybringer      = "b75c3902-633e-4d24-acde-d7a9cc8f466e"
	oracleResoluteSurvivors = "3d699db7-cc52-4ee3-947b-0ba291bf037a"
)

// MonolithDeckID is the arena deck id of MonolithBattleDeck: `--decks
// monolith-battle`. Synthetic, so it needs no Scryfall dump.
const MonolithDeckID = "monolith-battle"

// AnswersDeckID is the arena deck id of AnswersBattleDeck: `--decks
// answers-battle`. Synthetic, so it needs no Scryfall dump.
const AnswersDeckID = "answers-battle"

// TricksDeckID is the arena deck id of TricksBattleDeck: `--decks
// tricks-battle`. Synthetic, so it needs no Scryfall dump.
const TricksDeckID = "tricks-battle"

// ThopterDeckID is the arena deck id of ThopterBattleDeck: `--decks
// thopter-battle`. Synthetic, so it needs no Scryfall dump.
const ThopterDeckID = "thopter-battle"

// IsSynthetic reports whether a deck id is dealt by this package with no
// Scryfall dump behind it: the empty id (BattleDeck), exert-battle,
// monolith-battle, answers-battle, tricks-battle, thopter-battle and
// ghast-battle, and xpumps-battle (#4976).
func IsSynthetic(id string) bool {
	if id == XPumpsDeckID {
		return true
	}
	return id == "" || id == ExertDeckID || id == MonolithDeckID || id == AnswersDeckID || id == TricksDeckID || id == ThopterDeckID ||
		id == GhastDeckID
}

// GhastDeckID is the arena deck id of GhastBattleDeck: `--decks
// ghast-battle`. Synthetic, so it needs no Scryfall dump.
const GhastDeckID = "ghast-battle"

// GhastBattleDeck is a black-green deck whose lands make more mana than
// they print through triggered mana abilities, for #4770's arena
// measurement of the turn plan counting that mana: six Crypt Ghasts
// ("Whenever you tap a Swamp for mana, add an additional {B}") and four
// Zendikar Resurgents ("Whenever you tap a land for mana, add one mana of
// any type that land produced"), beside BattleDeck's vanilla Bears,
// Drakes, Ogres and Wurms recoloured, and twelve Swamps and twelve
// Forests. Its curve tops out at the Wurm's six, so a turn with the extra
// mana has two spells to cast. No curated deck holds either card. Dealt
// by oracle ID so the catalog gives each its abilities; the printed costs
// and stats are the cards' own.
func GhastBattleDeck(owner uuid.UUID) []game.Card {
	cmdr := game.NewCommander("Commander Bear", owner)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{2}{B}"
	cmdr.Colors = []string{"B"}
	cmdr.Power, cmdr.Toughness = 3, 3
	deck := []game.Card{cmdr}
	add := func(n int, name, oracle, typeLine, cost string, colors []string, power, toughness int, keywords ...string) {
		for i := 0; i < n; i++ {
			c := game.NewCard(name, owner)
			c.OracleID, c.TypeLine, c.ManaCost = oracle, typeLine, cost
			c.Colors = colors
			c.Power, c.Toughness = power, toughness
			c.Keywords = keywords
			deck = append(deck, c)
		}
	}
	black, green := []string{"B"}, []string{"G"}
	add(12, "Swamp", "", "Basic Land — Swamp", "", nil, 0, 0)
	add(12, "Forest", "", "Basic Land — Forest", "", nil, 0, 0)
	add(6, "Crypt Ghast", oracleCryptGhast, "Creature — Spirit", "{3}{B}", black, 2, 2)
	add(4, "Zendikar Resurgent", oracleZendikarResurgent, "Enchantment", "{5}{G}{G}", green, 0, 0)
	add(10, "Bear", "", "Creature — Bear", "{1}{B}", black, 2, 2)
	add(8, "Drake", "", "Creature — Drake", "{2}{B}", black, 3, 3, "flying")
	add(6, "Ogre", "", "Creature — Ogre", "{3}{G}", green, 4, 4)
	add(6, "Wurm", "", "Creature — Wurm", "{5}{G}", green, 7, 7, "trample")
	return deck
}

// The ghast deck's catalog keys.
const (
	oracleCryptGhast        = "a3c8d817-7949-4dae-b9f5-f9d952479270"
	oracleZendikarResurgent = "a0af158b-f42a-4022-a438-e0a2e1fd6d9e"
)

// XPumpsDeckID is the arena deck id of XPumpsBattleDeck: `--decks
// xpumps-battle`. Synthetic, so it needs no Scryfall dump.
const XPumpsDeckID = "xpumps-battle"

// XPumpsBattleDeck is BattleDeck in red and green with the pumps of the
// announced X #4976 prices in it, for its arena measurement: four Tyvar's
// Stand ("+X/+X and gains hexproof and indestructible"), four Mishra's
// Command (its last bullet "+X/+0 and gains haste"), two Primal Might
// ("+X/+X", then a fight) and two Kessig Wolf Runs ("{X}{R}{G}, {T}: +X/+0
// and gains trample"), with ten Forests for the green. No curated deck
// holds any of them. Hatred is left out: its {B}{B} would want a third
// colour. Dealt by oracle ID so the catalog gives each its effect; the
// printed costs are the cards' own.
func XPumpsBattleDeck(owner uuid.UUID) []game.Card {
	deck := BattleDeck(owner)
	out := make([]game.Card, 0, len(deck))
	drop := map[string]int{"Bear": 8, "Ogre": 4, "Mountain": 10}
	for _, c := range deck {
		if drop[c.Name] > 0 {
			drop[c.Name]--
			continue
		}
		out = append(out, c)
	}
	add := func(n int, name, oracle, typeLine, cost string, colors []string) {
		for i := 0; i < n; i++ {
			c := game.NewCard(name, owner)
			c.OracleID, c.TypeLine, c.ManaCost = oracle, typeLine, cost
			c.Colors = colors
			out = append(out, c)
		}
	}
	add(4, "Tyvar's Stand", oracleTyvarsStand, "Instant", "{X}{G}", []string{"G"})
	add(4, "Mishra's Command", oracleMishrasCommand, "Sorcery", "{X}{R}", []string{"R"})
	add(2, "Primal Might", oraclePrimalMight, "Sorcery", "{X}{G}", []string{"G"})
	add(2, "Kessig Wolf Run", oracleKessigWolfRun, "Land", "", nil)
	add(10, "Forest", "", "Basic Land — Forest", "", nil)
	return out
}

// The X-pumps deck's catalog keys.
const (
	oracleTyvarsStand    = "4247b667-a0a8-4995-97c7-622d20132f7d"
	oracleMishrasCommand = "25434ea3-bcfa-4dae-a16e-10ab87ce32af"
	oraclePrimalMight    = "58d7b8b3-ac47-4805-8259-4ba8e70d7dc3"
	oracleKessigWolfRun  = "c6911265-54ef-4c16-bcf2-1ffb24b7d426"
)

// ThopterBattleDeck is a mono-blue deck led by Sai, Master Thopterist,
// for #4991's arena measurement: Sai makes a Thopter for each artifact
// spell and trades two artifacts for a card. Its artifacts are eight
// Ornithopters, six Memnites, four Hangarback Walkers, four Burnished
// Harts and four Mind Stones, beside BattleDeck's vanilla Drakes, Ogres
// and Wurms recoloured blue and twenty-four Islands. The curated Thopter
// Swarm deck holds the same artifacts but needs a Scryfall dump; this
// one does not. Dealt by oracle ID so the catalog gives each card its
// abilities; the printed costs and stats are the cards' own.
func ThopterBattleDeck(owner uuid.UUID) []game.Card {
	cmdr := game.NewCommander("Sai, Master Thopterist", owner)
	cmdr.OracleID = oracleSai
	cmdr.TypeLine, cmdr.ManaCost = "Legendary Creature — Human Artificer", "{2}{U}"
	cmdr.Colors = []string{"U"}
	cmdr.Power, cmdr.Toughness = 1, 4
	deck := []game.Card{cmdr}
	add := func(n int, name, oracle, typeLine, cost string, blue bool, power, toughness int, keywords ...string) {
		for i := 0; i < n; i++ {
			c := game.NewCard(name, owner)
			c.OracleID, c.TypeLine, c.ManaCost = oracle, typeLine, cost
			if blue {
				c.Colors = []string{"U"}
			}
			c.Power, c.Toughness = power, toughness
			c.Keywords = keywords
			deck = append(deck, c)
		}
	}
	add(24, "Island", "", "Basic Land — Island", "", false, 0, 0)
	add(8, "Ornithopter", oracleOrnithopter, "Artifact Creature — Thopter", "{0}", false, 0, 2, "flying")
	add(6, "Memnite", oracleMemnite, "Artifact Creature — Construct", "{0}", false, 1, 1)
	add(4, "Hangarback Walker", oracleHangarbackWalker, "Artifact Creature — Construct", "{X}{X}", false, 0, 0)
	add(4, "Burnished Hart", oracleBurnishedHart, "Artifact Creature — Elk", "{3}", false, 2, 2)
	add(4, "Mind Stone", oracleMindStone, "Artifact", "{2}", false, 0, 0)
	add(6, "Drake", "", "Creature — Drake", "{2}{U}", true, 3, 3, "flying")
	add(4, "Ogre", "", "Creature — Ogre", "{3}{U}", true, 4, 4)
	add(4, "Wurm", "", "Creature — Wurm", "{5}{U}", true, 7, 7, "trample")
	return deck
}

// The thopter deck's catalog keys.
const (
	oracleSai              = "52241b9c-7a69-4176-9234-8bdab09d8e64"
	oracleOrnithopter      = "a3a98bc9-caa0-49b7-951c-fe4e4f54e4ba"
	oracleMemnite          = "7663ac7c-1de3-4250-b96a-fae9dbd66a27"
	oracleHangarbackWalker = "dde55256-5259-44e7-a267-fca45a7f0d04"
	oracleBurnishedHart    = "893fed41-c144-433f-af88-bc7d419b7fb3"
	oracleMindStone        = "c97361b5-af16-4a7b-af85-a429dbaf4ad2"
)

// TricksBattleDeck is BattleDeck with the changes until end of turn #4312
// prices in it, for its arena measurement: four Brute Force ("target
// creature gets +3/+3 until end of turn"), three Dismember (-5/-5), four
// Mother of Runes and three Center Soul (protection from the colour of
// your choice), and a Rogue's Passage ("can't be blocked this turn"),
// with seven Plains for the white. No curated deck holds a pump spell, and
// only two hold Mother of Runes. Dealt by oracle ID so the catalog gives
// each its effect; the printed costs and stats are the cards' own.
func TricksBattleDeck(owner uuid.UUID) []game.Card {
	deck := BattleDeck(owner)
	out := make([]game.Card, 0, len(deck))
	drop := map[string]int{"Bear": 8, "Ogre": 4, "Drake": 2, "Mountain": 8}
	for _, c := range deck {
		if drop[c.Name] > 0 {
			drop[c.Name]--
			continue
		}
		out = append(out, c)
	}
	add := func(n int, name, oracle, typeLine, cost string, colors []string, power, toughness int) {
		for i := 0; i < n; i++ {
			c := game.NewCard(name, owner)
			c.OracleID, c.TypeLine, c.ManaCost = oracle, typeLine, cost
			c.Colors = colors
			c.Power, c.Toughness = power, toughness
			out = append(out, c)
		}
	}
	add(4, "Brute Force", oracleBruteForce, "Instant", "{R}", []string{"R"}, 0, 0)
	add(3, "Dismember", oracleDismember, "Instant", "{1}{B/P}{B/P}", []string{"B"}, 0, 0)
	add(4, "Mother of Runes", oracleMotherOfRunes, "Creature — Human Cleric", "{W}", []string{"W"}, 1, 1)
	add(3, "Center Soul", oracleCenterSoul, "Instant", "{1}{W}", []string{"W"}, 0, 0)
	add(7, "Plains", "", "Basic Land — Plains", "", nil, 0, 0)
	add(1, "Rogue's Passage", oracleRoguesPassage, "Land", "", nil, 0, 0)
	return out
}

// The tricks deck's catalog keys.
const (
	oracleBruteForce    = "9880ba09-d5b8-4675-bfb4-2161d86d2d41"
	oracleDismember     = "fd74f8eb-0253-42dd-8277-186d4934da38"
	oracleMotherOfRunes = "60433b48-d27f-413c-905c-43839b1943f1"
	oracleCenterSoul    = "686b44ec-3446-4e1f-a15f-9d8557db6d70"
	oracleRoguesPassage = "f29dc596-2121-4421-8463-15f6c2e8b9b3"
)

// AnswersBattleDeck is BattleDeck with ten Basalt Gargoyles and eight
// Selfless Spirits in it, for #2939's arena checks of ADR 0142's answer
// pricing: a 3/2 flier whose "{R}: +0/+1" needs two activations to live
// through the deck's Lightning Bolts, and a 2/1 flier whose sacrifice
// gives the rest of the team indestructible. No curated deck holds a
// pump creature of the kind. Dealt by oracle ID so the catalog gives each
// its abilities; the printed costs and stats are the cards' own.
func AnswersBattleDeck(owner uuid.UUID) []game.Card {
	deck := BattleDeck(owner)
	out := make([]game.Card, 0, len(deck))
	dropBears, dropOgres, dropMountains := 12, 6, 8
	for _, c := range deck {
		switch {
		case c.Name == "Bear" && dropBears > 0:
			dropBears--
		case c.Name == "Ogre" && dropOgres > 0:
			dropOgres--
		case c.Name == "Mountain" && dropMountains > 0:
			dropMountains--
		default:
			out = append(out, c)
		}
	}
	for i := 0; i < 10; i++ {
		c := game.NewCard("Basalt Gargoyle", owner)
		c.OracleID = oracleBasaltGargoyle
		c.TypeLine, c.ManaCost = "Creature — Gargoyle", "{2}{R}"
		c.Power, c.Toughness = 3, 2
		c.Keywords = []string{"flying"}
		out = append(out, c)
	}
	// Selfless Spirit (#2939 gap 3): eight of them, with the Plains that
	// cast them.
	for i := 0; i < 8; i++ {
		c := game.NewCard("Selfless Spirit", owner)
		c.OracleID = oracleSelflessSpirit
		c.TypeLine, c.ManaCost = "Creature — Spirit Cleric", "{1}{W}"
		c.Power, c.Toughness = 2, 1
		c.Keywords = []string{"flying"}
		out = append(out, c)
		p := game.NewCard("Plains", owner)
		p.TypeLine = "Basic Land — Plains"
		out = append(out, p)
	}
	return out
}

// The answers deck's catalog keys.
const (
	oracleBasaltGargoyle = "83a38c8d-e2f6-404b-83be-9cfce037d6e2"
	oracleSelflessSpirit = "71d785a9-ddc8-472e-a778-a551b444a4bd"
)

// MonolithBattleDeck is BattleDeck with ten of the two Monoliths in it,
// for #2500's arena measurement: Basalt Monolith ("{T}: Add {C}{C}{C}",
// "{3}: Untap") and Grim Monolith ("{4}: Untap") are the cards whose
// self-untap a heuristic seat once repeated for no mana until the CR
// 732 breaker parked it. No curated deck holds either. Dealt by oracle
// ID so the catalog gives each its abilities; the printed costs are the
// cards' own.
func MonolithBattleDeck(owner uuid.UUID) []game.Card {
	deck := BattleDeck(owner)
	out := make([]game.Card, 0, len(deck))
	dropBears, dropOgres := 6, 4
	for _, c := range deck {
		switch {
		case c.Name == "Bear" && dropBears > 0:
			dropBears--
		case c.Name == "Ogre" && dropOgres > 0:
			dropOgres--
		default:
			out = append(out, c)
		}
	}
	mono := func(name, oracle string) game.Card {
		c := game.NewCard(name, owner)
		c.OracleID = oracle
		c.TypeLine, c.ManaCost = "Artifact", "{3}"
		return c
	}
	for i := 0; i < 6; i++ {
		out = append(out, mono("Basalt Monolith", oracleBasaltMonolith))
	}
	for i := 0; i < 4; i++ {
		out = append(out, mono("Grim Monolith", oracleGrimMonolith))
	}
	return out
}

// The Monoliths' catalog keys.
const (
	oracleBasaltMonolith = "6b8cf2a0-b045-4d91-9d91-c602d40c6237"
	oracleGrimMonolith   = "229d6627-1292-4ae1-8849-b0f956fa6540"
)

// CuratedDeck deals one of the pre-built decks (docs/bot.md,
// "The curated decks"), which is what the model tiers are
// actually configured for: the prompt's static block is that deck's
// profile, so a model seat playing BattleDeck is being asked to
// reason about a list it was never shown.
//
// It needs a loaded Scryfall index because a curated decklist is a
// list of NAMES; the type lines, costs and oracle IDs come from the
// dump. A nil index is therefore an error rather than a thinner deck:
// silently seating everyone on vanilla bears while the report says
// "esper-control" would make the measurement a lie.
func CuratedDeck(idx *cards.Index, id string) ([]game.Card, error) {
	if idx == nil {
		return nil, fmt.Errorf("botarena: deck %q needs a Scryfall dump (--dump or $CMDCTRL_SCRYFALL_DUMP)", id)
	}
	list, err := decks.Load(idx, id)
	if err != nil {
		return nil, fmt.Errorf("botarena: %w", err)
	}
	return list.ToGameCards(), nil
}
