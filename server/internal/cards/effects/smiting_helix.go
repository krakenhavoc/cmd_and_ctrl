package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smiting Helix — Sorcery {3}{B} (Modern Horizons, #5519):
//
//	"Smiting Helix deals 3 damage to any target and you gain 3 life.
//	 Flashback {R}{W} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Agonizing Syphon's body at sorcery speed with a red-white flashback.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "19d19e32-af13-4bb3-925f-8ed54cce2fcb",
		Name:             "Smiting Helix",
		Completeness:     CompletenessFull,
		Targets:          TargetAny(),
		Purpose:          ForTargets(DamageToTarget(0, 3)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{R}{W}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return damageXToTheTargetAndGainX(item, ctx, 3)
		},
	})
}
