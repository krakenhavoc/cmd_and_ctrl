package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crawl from the Cellar — Sorcery {B} (Midnight Hunt):
//
//	"Return target creature card from your graveyard to your hand. Put a
//	 +1/+1 counter on up to one target Zombie you control.
//	 Flashback {3}{B}"
//
// Two clauses, the second "up to one": it can be left empty, and either
// one going illegal only loses its own sentence (CR 608.2b). The Zombie
// is read through the layered view.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16c99713-f19e-4b25-9b03-6a8d8203adfe",
		Name:         "Crawl from the Cellar",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
			TargetCreature("up to one target Zombie you control", YouControl(), HasSubtype("Zombie")).WithCount(0, 1),
		),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{B}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		),
	})
}
