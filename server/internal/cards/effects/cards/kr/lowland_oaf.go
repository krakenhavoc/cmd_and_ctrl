package cardskr

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lowland Oaf — Creature — Giant Warrior {3}{R}, 3/3 (Lorwyn, S82 OLD b545,
// #5593):
//
//	"{T}: Target Goblin creature you control gets +1/+0 and gains flying
//	 until end of turn. Sacrifice that creature at the beginning of the next
//	 end step."
//
// The target is a Goblin you control (the Oaf is a Giant, not one). On
// resolution the creature gets the pump and the flying, and a delayed
// trigger (CR 603.7) sacrifices it at the beginning of the next end step.
// A target that is illegal by resolution does nothing at all (CR 608.2b),
// so nothing is scheduled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c14f973e-b2b0-4d22-87f6-c37f7eb92864",
		Name:         "Lowland Oaf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target Goblin creature you control gets +1/+0 and gains flying until end of turn. Sacrifice that creature at the beginning of the next end step.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "flying")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target Goblin creature you control", effects.OfCreatureType("Goblin"), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ResolveSteps(
					effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
						return effects.BoostUntilEOT{Target: id, Power: 1, Label: "Lowland Oaf — +1/+0"}.Apply(ctx)
					}),
					effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Lowland Oaf — flying"}.Apply(ctx)
					}),
					effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
						return effects.ScheduleDelayedTrigger{
							Label: "Lowland Oaf — sacrifice that creature",
							Cards: []uuid.UUID{id},
							Body:  effects.SacrificeListedCardsBody,
						}.Apply(ctx)
					}),
				)(item, effects.NewContext(g, item))
			},
		}},
	})
}
