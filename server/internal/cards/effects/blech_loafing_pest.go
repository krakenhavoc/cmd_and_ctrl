package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blech, Loafing Pest — Legendary Creature — Pest {1}{B}{G}, 3/4 (Secrets
// of Strixhaven, #4456):
//
//	"Whenever you gain life, put a +1/+1 counter on each Pest, Bat,
//	 Insect, Snake, and Spider you control."
//
// One trigger per life-gain event. The set is read as the ability
// resolves: every permanent you control with any of the five subtypes,
// Blech included (it is a Pest), each given one counter as one placement.
// The subtypes are read off the permanents as they are now, so a
// changeling counts, and the five words are read through Blech's text
// (CR 612.2), so a text change that swaps one of them swaps what is
// counted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1bdd361-c7f2-4aa2-8d07-b582775551ff",
		Name:         "Blech, Loafing Pest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Blech, Loafing Pest — a +1/+1 counter on each Pest, Bat, Insect, Snake, and Spider you control",
				blechCounters),
		},
	})
}

// blechCounters puts the counters on.
func blechCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	subtypes := ctx.TextSubtypes("Pest", "Bat", "Insect", "Snake", "Spider")
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller {
			continue
		}
		for _, sub := range subtypes {
			if c.HasSubtypeWord(sub) {
				ids = append(ids, c.InstanceID)
				break
			}
		}
	}
	group := ctx.asGroupMember()
	for _, id := range ids {
		if !onBattlefield(g, id) {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(group); err != nil {
			return err
		}
	}
	return nil
}
