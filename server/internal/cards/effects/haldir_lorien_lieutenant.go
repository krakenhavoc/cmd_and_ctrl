package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haldir, Lórien Lieutenant — Legendary Creature — Elf Soldier {X}{G}, 0/0
// (Tales of Middle-earth Commander, #5948):
//
//	"Haldir enters with X +1/+1 counters on it.
//	 Vigilance
//	 {5}{G}: Until end of turn, other Elves you control gain vigilance and get
//	 +1/+1 for each +1/+1 counter on Haldir."
//
// The counters are Benevolent Hydra's (XCounters, one replacement on the entry,
// so Hardened Scales applies). The ability's bonus is counted once, as it
// resolves (CR 611.2c): the +1/+1 counters on Haldir then, or as he last
// existed if he has left (CR 608.2h); the other Elves you control as it
// resolves are the set that gets it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "653217b1-1e92-49a7-81d9-fc4d70d4f27f",
		Name:                       "Haldir, Lórien Lieutenant",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		PrintedKeywords:            []string{"vigilance"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Activated: []ActivatedAbility{{
			Label:   "{5}{G}: Until end of turn, other Elves you control gain vigilance and get +1/+1 for each +1/+1 counter on Haldir.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{5}{G}"),
			Effect:  haldirPump,
		}},
	})
}

func haldirPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 0
	if onBattlefield(g, item.SourceCardID) {
		if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
			n = c.Counters[game.CounterPlusOne]
		}
	} else {
		n = g.LastKnownCountersForEffect(item.SourceCardID)[game.CounterPlusOne]
	}
	self, elf := item.SourceCardID, ctx.TextSubtype("Elf")
	elves := func(_ *game.Game, you uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.Controller == you && c.InstanceID != self && c.HasSubtype(elf)
	}
	if err := (GrantKeywordUntilEOT{Match: elves, Keywords: []string{"vigilance"}, Label: "Haldir, Lórien Lieutenant — vigilance"}).Apply(ctx); err != nil {
		return err
	}
	if n <= 0 {
		return nil
	}
	return BoostUntilEOT{Match: elves, Power: n, Toughness: n, Label: "Haldir, Lórien Lieutenant — +1/+1 for each +1/+1 counter"}.Apply(ctx)
}
