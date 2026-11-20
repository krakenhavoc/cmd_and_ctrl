package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crowd's Favor — Instant {R} (Magic 2015, #5556):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Target creature gets +1/+0 and gains first strike until end of turn.
//	 (It deals combat damage before creatures without first strike.)"
//
// Guided Strike's pump-and-first-strike (socB27TargetGetsAndGains)
// without the draw, with the convoke TapCost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a23915f-8bce-4aad-8395-0c736450a62d",
		Name:         "Crowd's Favor",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(BenefitUntilEOT(0, 1, 0, "first strike"))},
		OnResolve:    socB27TargetGetsAndGains(1, 0, "first strike", "Crowd's Favor"),
	})
}
