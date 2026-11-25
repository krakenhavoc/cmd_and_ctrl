package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Indominus Rex, Alpha — Legendary Creature — Dinosaur Mutant
// {1}{U/B}{U/B}{G}{G}, 6/6 (Jurassic World Collection, #6100):
//
//	"As Indominus Rex enters, discard any number of creature cards. It
//	 enters with a flying counter on it if a card discarded this way has
//	 flying. The same is true for first strike, double strike,
//	 deathtouch, hexproof, haste, indestructible, lifelink, menace,
//	 reach, trample, and vigilance.
//	 When Indominus Rex enters, draw a card for each counter on it."
//
// The first ability is a CR 614.1c replacement: as the Rex enters, its
// controller discards any number of creature cards from their hand
// (ADR 0098's entry discard, game.EntryCardDiscard, zero included). The
// entry waits for the answer (CR 614.12), then gets one counter of each
// listed keyword that at least one card really discarded has (CR 122.1b),
// so it arrives with them and the enters trigger sees them. A card with
// two of the keywords gives both; two cards with the same keyword give
// one counter of it.
//
// The enters trigger counts every counter on the Rex as it resolves,
// whatever its kind (keyword counters, and any +1/+1 counters it entered
// with from elsewhere), from its last-known information if it has left
// (CR 608.2h).
//
// No simplification.
const indominusName = "Indominus Rex, Alpha"

// indominusAnyNumber is "any number": the ceiling is clamped to the
// creature cards in hand (EntryCardChoice.bounds).
const indominusAnyNumber = 1 << 10

// indominusKeywords are the printed keywords, each one a keyword counter
// kind (CR 122.1b), in the printed order.
var indominusKeywords = []string{
	game.CounterFlying, game.CounterFirstStrike, game.CounterDoubleStrike, game.CounterDeathtouch,
	game.CounterHexproof, game.CounterHaste, game.CounterIndestructible, game.CounterLifelink,
	game.CounterMenace, game.CounterReach, game.CounterTrample, game.CounterVigilance,
}

func init() {
	effects.Register(effects.Spec{
		OracleID:     "723a66c3-2c2c-411c-bba5-922e8601981b",
		Name:         indominusName,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			RowLabel:        "As Indominus Rex enters, discard any number of creature cards; it enters with a counter of each listed keyword a card discarded this way has.",
			EntryOnly:       true,
			Watches:         []game.EventKind{game.EventZoneMove},
			SelfReplacement: true,
			Label:           indominusName,
			PromptQuestion:  indominusName + " — discard any number of creature cards",
			EntryCardChoice: &game.EntryCardChoice{
				Action:  game.EntryCardDiscard,
				Matches: func(c game.Card) bool { return c.IsCreature() },
				Min:     0,
				Max:     indominusAnyNumber,
				Then:    indominusCounters,
			},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove &&
					ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Controller: effects.EnteringPermanentChooser,
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(indominusName+" — draw a card for each counter on it", indominusDraw),
		},
	})
}

// indominusCounters gives the entry one counter of each listed keyword a
// card really discarded has.
func indominusCounters(g *game.Game, ev *game.ReplacementEvent, discarded []uuid.UUID) error {
	for _, kw := range indominusKeywords {
		for _, id := range discarded {
			c, ok := g.LookupCardForEffect(id)
			if ok && game.HasKeyword(&c, kw) {
				ev.AddCounterAtETB(kw, 1)
				break
			}
		}
	}
	return nil
}

// indominusDraw is "draw a card for each counter on it".
func indominusDraw(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src := item.SourceCardID
	var counters map[string]int
	if effects.OnBattlefield(g, src) && !ctx.IsNewSourceObject(src) {
		if c, ok := g.LookupCardForEffect(src); ok {
			counters = c.Counters
		}
	} else if info, ok := g.LastKnownPermanentForEffect(src); ok {
		counters = info.Counters
	}
	n := 0
	for _, k := range counters {
		if k > 0 {
			n += k
		}
	}
	if n == 0 {
		return nil
	}
	return effects.DrawCards{N: n}.Apply(ctx)
}
