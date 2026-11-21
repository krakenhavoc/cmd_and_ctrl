package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resistance Is Futile — Sorcery {1}{B} (Star Trek, #4973):
//
//	"Choose one —
//	 • Destroy target nonartifact creature.
//	 • Return target artifact creature card with mana value 3 or less
//	   from your graveyard to the battlefield."
//
// Each bullet's target is chosen as the spell is cast and re-checked as it
// resolves (CR 608.2b). A card with {X} in its cost counts X as 0 in the
// graveyard (CR 202.3e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49975689-85f8-4221-bb11-3d3f68a1a7ac",
		Name:         "Resistance Is Futile",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target nonartifact creature.",
				effects.TargetCreature("target nonartifact creature", effects.Not(effects.Artifact())),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Return target artifact creature card with mana value 3 or less from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target artifact creature card with mana value 3 or less from your graveyard",
					effects.YouOwn(), effects.Artifact(), effects.Creature(), effects.ManaValueLE(3)),
				effects.ReturnTheModesTargetToBattlefield),
		),
	})
}
