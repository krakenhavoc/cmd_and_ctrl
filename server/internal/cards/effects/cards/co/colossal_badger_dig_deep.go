package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const colossalBadgerOracleID = "96aab05a-49bc-418a-9980-f3322955de1e"

// Colossal Badger // Dig Deep — Adventure.
//
//	Colossal Badger {5}{G} Creature — Badger 6/5
//	  "When this creature enters, you gain 3 life."
//	Dig Deep {1}{G} Sorcery — Adventure
//	  "Choose target creature. Mill four cards, then put a +1/+1 counter on
//	   that creature for each creature card milled this way."
//
// Dig Deep targets the creature (a hexproof creature can't be chosen). The
// counters count the creature cards the mill really put into the graveyard,
// read once the mill has settled, and go on the creature if it is still there
// (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     colossalBadgerOracleID,
		Name:         "Colossal Badger",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Colossal Badger — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     colossalBadgerOracleID + "#1",
		Name:         "Dig Deep",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok {
				return nil
			}
			return ctx.Game.MillToZoneKeyedForEffect(ctx.Controller(), 4, game.ZoneGraveyard, game.MillUntil{}, ctx.ItemTail(digDeepMilled))
		},
	})
}

// digDeepMilled is "then put a +1/+1 counter on that creature for each
// creature card milled this way". An on-disk key: never renamed, never
// reused.
var digDeepMilled = effects.CardTail("tail/dig-deep-counters", func(ctx *effects.Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
})
