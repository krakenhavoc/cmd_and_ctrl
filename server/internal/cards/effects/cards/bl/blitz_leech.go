package cardsbl

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blitz Leech — Creature — Leech {5}{B}, 5/2 (Ikoria: Lair of Behemoths,
// S82 OLD batch 120, #5512):
//
//	"Flash
//	 When this creature enters, target creature an opponent controls gets
//	 -2/-2 until end of turn. Remove all counters from that creature."
//
// Flash is a printed keyword. The trigger does its two sentences in order on
// the still-legal target (CR 608.2b): the -2/-2 (layer 7c, ending in the
// cleanup step), then every counter of every kind comes off it. State-based
// actions wait until the trigger has finished, so a creature the shrink would
// kill has lost its counters first, and a creature that survives stays
// shrunk with none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7679942f-fcbc-4750-8e8c-fe0ad8103c11",
		Name:            "Blitz Leech",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Blitz Leech — target creature an opponent controls gets -2/-2 until end of turn; remove all counters from it",
				blitzLeechEnters),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		},
	})
}

// blitzLeechEnters shrinks the target, then strips its counters.
func blitzLeechEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.BoostUntilEOT{Target: id, Power: -2, Toughness: -2, Label: "Blitz Leech — -2/-2"}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	kinds := make([]string, 0, len(c.Counters))
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		if err := (effects.AddCounter{Target: id, Kind: kind, N: -c.Counters[kind]}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
