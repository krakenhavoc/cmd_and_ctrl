package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crippling Fatigue — Sorcery {1}{B}{B} (Torment, #5612):
//
//	"Target creature gets -2/-2 until end of turn.
//	 Flashback—{1}{B}, Pay 3 life. (You may cast this card from your
//	 graveyard for its flashback cost. Then exile it.)"
//
// The shared boostTheTargetUntilEOT body at -2/-2, which ends at cleanup
// (CR 514.2), declared as a harm so it is aimed at an opponent's
// creature. The flashback price is mana and life
// (torB1FlashbackPayingLife): a caster with less than 3 life can't
// claim it (CR 118.3), and the card is exiled as it leaves the stack
// (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d87190d0-bda3-4ad9-84b2-019f751999ce",
		Name:             "Crippling Fatigue",
		Completeness:     CompletenessFull,
		Purpose:          ForTargets(HarmUntilEOT(0, -2, -2)),
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{torB1FlashbackPayingLife("{1}{B}", 3)},
		OnResolve:        boostTheTargetUntilEOT(-2, -2, "Crippling Fatigue — -2/-2"),
	})
}
