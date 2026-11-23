package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Fiend — Artifact Creature — Horror {6}, 0/0 (Darksteel, #5605):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 At the beginning of your upkeep, you may move a +1/+1 counter from
//	 target creature onto this creature.
//	 Modular 3 (This creature enters with three +1/+1 counters on it. When
//	 it dies, you may put its +1/+1 counters on target artifact
//	 creature.)"
//
// The upkeep trigger targets any creature, the Fiend included, and asks
// "you may" as it resolves (CR 603.5). The move follows CR 122.5: one
// +1/+1 counter comes off the target and goes on the Fiend, and nothing
// moves when either half is impossible: the target has no +1/+1 counter,
// the target is the Fiend itself, the Fiend has left the battlefield or
// come back as a new object, or the Fiend can't have counters put on it.
// Fear and modular are engine keywords (modular: #2012, CR 702.43a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "72e6cc74-0574-4421-a6b8-97d180484950",
		Name:            "Arcbound Fiend",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear", "modular 3"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.AtYourUpkeep("Arcbound Fiend — move a +1/+1 counter from target creature onto this creature",
				arcboundFiendMove), effects.TargetCreature("target creature")),
				"Arcbound Fiend — move a +1/+1 counter from the target creature onto it?"),
		},
	})
}

// arcboundFiendMove moves one +1/+1 counter from the trigger's target onto
// its source, or nothing at all (CR 122.5).
func arcboundFiendMove(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, self) || g.CantHaveCountersForEffect(self) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == self {
			continue
		}
		from, ok := g.LookupCardForEffect(t.ID)
		if !ok || from.Counters[game.CounterPlusOne] <= 0 {
			return nil
		}
		if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: -1}).Apply(ctx); err != nil {
			return err
		}
		return effects.AddCounter{Target: self, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	return nil
}
