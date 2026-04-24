package effects

// Ambush Viper — "Flash, deathtouch."
//
// 2/1 for 1G. The flash keyword lets this creature be cast at
// instant speed (cast_spell accepts it outside the normal sorcery
// window). HasKeyword reads the off-battlefield fallback via
// CatalogPrintedKeywords when gating the cast from hand — this is
// the forcing function for the S18 PrintedKeywords slot design
// choice (ADR 0014 §8).
//
// Note: the cast-time flash gate itself lives in CastSpell's
// sorcery-speed check (the gate reads HasKeyword on the hand
// card); this sub-PR adds the card; cast-flash gating enforcement
// lands with the S18 cast-speed changes. Off-battlefield reader
// is already live.
func init() {
	Register(Spec{
		OracleID:        "8957f7c2-040c-4048-9f21-efa7c97682b7",
		Name:            "Ambush Viper",
		PrintedKeywords: []string{"flash", "deathtouch"},
	})
}
