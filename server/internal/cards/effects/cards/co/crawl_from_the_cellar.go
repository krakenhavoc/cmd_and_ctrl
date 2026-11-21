package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crawl from the Cellar — Sorcery {B} (Midnight Hunt):
//
//	"Return target creature card from your graveyard to your hand. Put a
//	 +1/+1 counter on up to one target Zombie you control.
//	 Flashback {3}{B}"
//
// Two clauses, the second "up to one": it can be left empty, and either
// one going illegal only loses its own sentence (CR 608.2b). The Zombie
// is read through the layered view. The counter is Cost of Brilliance's
// costOfBrillianceCounter ("up to one target" in clause 1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16c99713-f19e-4b25-9b03-6a8d8203adfe",
		Name:         "Crawl from the Cellar",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
			effects.TargetCreature("up to one target Zombie you control", effects.YouControl(), effects.HasSubtype("Zombie")).WithCount(0, 1),
		),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{B}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
			}),
			effects.Instruction(effects.CostOfBrillianceCounter),
		),
	})
}
