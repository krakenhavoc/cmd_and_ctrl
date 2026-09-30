package effects

// Tecutlan, the Searing Rift — Legendary Land — Cave, the back face of
// Brass's Tunnel-Grinder (brasss_tunnel_grinder.go), registered under
// "<oracle_id>#1":
//
//	"(Transforms from Brass's Tunnel-Grinder.)
//	 {T}: Add {R}.
//	 Whenever you cast a permanent spell using mana produced by
//	 Tecutlan, discover X, where X is that spell's mana value."
//
// The mana ability is ordinary. The trigger is not implemented:
// discover (CR 701.57) has no primitive in the engine — exile cards
// from the top of the library until a nonland card with mana value X
// or less, cast it free or put it into hand, the rest on the bottom at
// random — and no seam names it yet. The land still taps for {R}, so
// the card is weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     brassTunnelGrinderOracleID + "#1",
		Name:         "Tecutlan, the Searing Rift",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Tecutlan's discover isn't implemented — casting a permanent spell with its mana does nothing extra.",
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}",
		}},
	})
}
