package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lily Bowen, Raging Grandma — Legendary Creature — Mutant Warrior {3}{G},
// 0/0 (Fallout, #5948):
//
//	"Vigilance
//	 Lily Bowen enters with two +1/+1 counters on it.
//	 At the beginning of your upkeep, double the number of +1/+1 counters
//	 on Lily Bowen if its power is 16 or less. Otherwise, remove all but
//	 one +1/+1 counter from it, then you gain 1 life for each +1/+1
//	 counter removed this way."
//
// The counters are a CR 614.1c self-replacement as it enters. The upkeep
// trigger asks about Lily's power as it resolves (the "if" is in the
// effect, not an intervening one): at 16 or less it puts on as many +1/+1
// counters as she has, the placement a counter doubler applies to; above
// 16 it removes all but one and the controller gains life for the counters
// that really came off. A Lily that has left, or come back as a new
// object, is not there to count, so nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6b2d92f5-b0a1-4f30-9140-da490bf01bda",
		Name:            "Lily Bowen, Raging Grandma",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 2, "Lily Bowen enters with two +1/+1 counters on it"),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Lily Bowen — double the +1/+1 counters if its power is 16 or less; otherwise remove all but one and gain that much life",
				lilyBowenUpkeep),
		},
	})
}

func lilyBowenUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := item.SourceCardID
	if ctx.isNewSourceObject(id) || !onBattlefield(g, id) {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	n := c.Counters[game.CounterPlusOne]
	if b31CurrentPowerOf(g, id) <= 16 {
		if n <= 0 {
			return nil
		}
		return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	}
	if n <= 1 {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: -(n - 1)}).Apply(ctx); err != nil {
		return err
	}
	after, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if removed := n - after.Counters[game.CounterPlusOne]; removed > 0 {
		return GainLife{Player: ctx.Controller(), Amount: removed}.Apply(ctx)
	}
	return nil
}
