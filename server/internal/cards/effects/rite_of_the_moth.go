package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rite of the Moth — Sorcery {1}{W}{B}{B} (Duskmourn: House of Horror,
// #4466):
//
//	"Return target creature card from your graveyard to the battlefield
//	 with a finality counter on it. (If a creature with a finality counter
//	 on it would die, exile it instead.)
//	 Flashback {3}{W}{W}{B}"
//
// Zombify's target with Grim Reaper's Scythe's return
// (grimReapersScytheReturn): the counter is on the creature as it enters
// (CR 614.1c), and a creature with a finality counter that would go to a
// graveyard is exiled instead (CR 122.1h). A target that left the
// graveyard in response makes the spell do nothing (CR 608.2b). Flashback
// is the shared alternative cost (CR 702.34).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a7fea839-d7d7-48b4-aeaf-133767532e32",
		Name:             "Rite of the Moth",
		Completeness:     CompletenessFull,
		Targets:          targetCreatureInYourGraveyard(),
		Purpose:          ForTargets(Reanimates(0)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{W}{W}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return grimReapersScytheReturn(ctx.Game, item)
		},
	})
}
