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
// izzet-aggro-vs-esper-control are two different numbers, and a
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

// CuratedDeck deals one of the four pre-built decks (docs/bot.md,
// "The four curated decks"), which is what the model tiers are
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
