package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorching Missile — Sorcery {3}{R} (Odyssey, #5613):
//
//	"Scorching Missile deals 4 damage to target player or planeswalker.
//	 Flashback {9}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Boros Charm's "target player or planeswalker" clause and the shared
// damageToFirstTarget body (CR 120.3a, 120.3c).
//
// Flashback is the shared constructor: the graveyard cast, the price,
// and CR 702.34a's exile as the spell leaves the stack (resolved,
// countered, or with its target gone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "02e2dbde-2373-4f29-bd9d-787f1805b061",
		Name:             "Scorching Missile",
		Completeness:     CompletenessFull,
		Targets:          targetPlayerOrPlaneswalker(),
		Purpose:          ForTargets(DamageToTarget(0, 4)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{9}{R}")},
		OnResolve:        damageToFirstTarget(4),
	})
}
