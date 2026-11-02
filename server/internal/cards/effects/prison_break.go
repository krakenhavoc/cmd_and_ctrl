package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prison Break — Sorcery {4}{B} (Marvel's Spider-Man):
//
//	"Return target creature card from your graveyard to the battlefield
//	 with an additional +1/+1 counter on it.
//	 Mayhem {3}{B} (You may cast this card from your graveyard for
//	 {3}{B} if you discarded it this turn. Timing rules still apply.)"
//
// Generous Revival's body without the mana value cap, with mayhem
// (CR 702.187b, mayhem_keyword.go). The counter is a CR 614.1c "enters
// with", so it rides the entry event and is added to any counters the
// creature enters with on its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "9c6a6332-37f0-4d7d-a86f-478cf0eeec91",
		Name:             "Prison Break",
		Completeness:     CompletenessFull,
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{3}{B}")},
		Targets:          TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OnResolve: returnTargetCreatureWithPlusOneCounter,
	})
}
