package effects

// Repel Calamity — Instant {1}{W} (Secrets of Strixhaven Mystical Archive,
// #4834):
//
//	"Destroy target creature with power or toughness 4 or greater."
//
// Bring Down's clause with an either-or bound: the pick is a creature with
// power 4 or greater or toughness 4 or greater (PowerGE, ToughnessGE), read
// through the layers, at the cast and again as the spell resolves, so a
// creature shrunk below both numbers in response is an illegal target
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba58f508-8036-496c-8a2f-c1e410165e93",
		Name:         "Repel Calamity",
		Completeness: CompletenessFull,
		Targets: TargetCreature("target creature with power or toughness 4 or greater",
			Or(PowerGE(4), ToughnessGE(4))),
		OnResolve: destroyTheTargetPermanent,
	})
}
