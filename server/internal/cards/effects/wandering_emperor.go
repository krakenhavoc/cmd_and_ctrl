package effects

// The Wandering Emperor — Legendary Planeswalker — Wanderer with
// starting loyalty 3 and three activated loyalty abilities:
//
//	+1: Create a 2/2 Samurai token with vigilance.
//	-1: Exile target tapped creature.
//	-2: Up to one target creature gets +2/+1 and gains lifelink
//	    until end of turn.
//
// Also: "Flash" and "As long as The Wandering Emperor entered this
// turn, you may activate her loyalty abilities any time you could
// cast an instant."
//
// S14 sandbox:
//   - StartingLoyalty = 3 is the ONLY piece wired this sprint. The
//     ETBEffectHook stamps "loyalty" counters equal to this value
//     via AddCounterForEffect, so the Emperor lands on the
//     battlefield with 3 loyalty visible in the UI. The 0-loyalty
//     SBA (S13.2 — cards/effects/primitives pipeline) would kill
//     a planeswalker with 0 starting counters, so stamping before
//     the next SBA pass matters.
//   - All three loyalty abilities are manual: players click the
//     same action the Emperor's "+1 / -1 / -2" tiles are bound to
//     in the gamecli or manual override. S19's ability auto-fire
//     sprint will wire OnResolve-style handlers to each loyalty
//     ability via a new Spec.Abilities field; this registration is
//     the attach point.
//   - Flash is cosmetic in S14 — cast-timing enforcement is S17.
//   - The "entered this turn" loyalty permission is an override
//     of CR 606.5; same S19 plumbing.
//
// Token template for the +1 lives in tokens.go (WhiteSamuraiToken)
// so the future ability wiring can CreateToken it without a new
// template file.
func init() {
	Register(Spec{
		OracleID:        "0c7f18d5-36cb-4bc6-a358-443b97666215",
		Name:            "The Wandering Emperor",
		StartingLoyalty: 3,
	})
}
