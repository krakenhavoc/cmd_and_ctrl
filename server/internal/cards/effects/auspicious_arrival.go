package effects

import "github.com/google/uuid"

// Auspicious Arrival — Instant {1}{W} (Murders at Karlov Manor, #4469):
//
//	"Target creature gets +2/+2 until end of turn. Investigate. (Create
//	 a Clue token. It's an artifact with "{2}, Sacrifice this token:
//	 Draw a card.")"
//
// Artful Maneuver's pump, then Deduce's Clue as the next step. A target
// that has become illegal counters the whole spell (CR 608.2b), so no
// Clue is made then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b25d5b06-44cf-42bc-90dd-3b238f702361",
		Name:         "Auspicious Arrival",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Auspicious Arrival — +2/+2"}.Apply(ctx)
			}),
			Investigate{N: 1},
		),
	})
}
