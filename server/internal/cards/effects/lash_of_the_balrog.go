package effects

// Lash of the Balrog — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 pay {4}. Destroy target creature."
//
// The either/or cost (ADR 0100 §2). The {4} branch joins the total
// before the cost modifiers (CR 601.2f), so Thalia taxes it once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66499407-3ad2-464b-bbfe-95d860f53616",
		Name:         "Lash of the Balrog",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("a creature", Creature()).Keyed("sacrifice"),
			ManaAdditionalCost("{4}").Keyed("mana"),
		),
		Targets:   TargetCreature("target creature"),
		OnResolve: destroyTheTargetPermanent,
	})
}
