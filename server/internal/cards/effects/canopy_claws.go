package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Canopy Claws — Instant {G} (Judgment, #5611):
//
//	"Target creature loses flying until end of turn.
//	 Flashback {G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Ribbon Snake's loss pointed at a target: a layer-6 ability-removing
// effect with its own timestamp (CR 613.1f), so a flying granted later
// in the turn puts it back, ending at cleanup (CR 514.2). Declared as a
// harm, so it is aimed at an opponent's flier. Flashback is the shared
// constructor, exiled as it leaves the stack (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "195c6453-e5cc-4b3b-8518-eaccf29bc4a6",
		Name:             "Canopy Claws",
		Completeness:     CompletenessFull,
		Purpose:          ForTargets(HarmUntilEOT(0, 0, 0)),
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{G}")},
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return untilEndOfTurn(ctx, id, nil, "Canopy Claws — loses flying until end of turn",
					game.RemoveKeywordsMod("flying"))
			}),
		),
	})
}
