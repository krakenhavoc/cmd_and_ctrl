package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ulamog, the Defiler — Legendary Creature — Eldrazi {10}, 7/7:
//
//	"When you cast this spell, target opponent exiles the top half of
//	 their library, rounded up.
//	 Ward—Sacrifice two permanents.
//	 Ulamog enters with a number of +1/+1 counters on it equal to the
//	 greatest mana value among cards in exile.
//	 Ulamog has annihilator X, where X is the number of +1/+1 counters
//	 on it."
//
// The cast trigger targets an opponent as Ulamog is cast and resolves
// above the spell, so it happens even if Ulamog is countered (the
// 2024-06-07 ruling). The half is counted as it resolves, rounded up,
// and exiling is not milling.
//
// Ward is the sacrifice ward: two permanents in one payment, or the
// spell or ability is countered.
//
// The entry counters are a CR 614.1c replacement read as Ulamog enters,
// before it moves, so a Ulamog entering from exile sees itself among the
// cards in exile (the ruling). A card exiled face down has no mana value
// (CR 406.3a), and a split card's is both halves' (CR 709.4b).
//
// "Annihilator X" is not a fixed token, so it is the catalog's
// AnnihilatorCounted row: the engine's annihilator trigger with X read
// as it resolves — the +1/+1 counters on Ulamog then, or as it last
// existed if it has left (the ruling; CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97836c48-8777-4b4e-98fb-e99204f38bdd",
		Name:         "Ulamog, the Defiler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ulamogDefilerCastTrigger(),
			Ward(WardSacrificeN(2, "two permanents", Permanent()),
				"Ulamog, the Defiler — ward, sacrifice two permanents"),
			AnnihilatorCounted("Ulamog, the Defiler — annihilator X, where X is the number of +1/+1 counters on it",
				CountersOnThisAsItResolves(game.CounterPlusOne)),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterPlusOne, greatestManaValueInExile(g))
				return nil
			},
			Label: "Ulamog, the Defiler: enters with +1/+1 counters equal to the greatest mana value among cards in exile",
		}},
	})
}

// ulamogDefilerCastTrigger is "When you cast this spell, target
// opponent exiles the top half of their library, rounded up."
func ulamogDefilerCastTrigger() game.TriggeredAbility {
	t := WhenYouCastThisSpell("Ulamog, the Defiler — target opponent exiles the top half of their library",
		targetOpponentExilesTopHalfOfTheirLibrary)
	t.Targets = TargetPlayer("target opponent", Opponent())
	return t
}

// greatestManaValueInExile is the greatest mana value among the cards
// in exile, zero when there are none. A card exiled face down has no
// mana value (CR 406.3a).
func greatestManaValueInExile(g *game.Game) int {
	best := 0
	if g.Exile == nil {
		return 0
	}
	for _, c := range g.Exile.Cards {
		if c.FaceDown {
			continue
		}
		if mv := c.ManaValue(); mv > best {
			best = mv
		}
	}
	return best
}
