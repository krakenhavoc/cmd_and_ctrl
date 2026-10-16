package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fluxcharger — Creature — Weird {2}{U}{R}, 1/5:
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell, you may switch this
//	 creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7). "You may" is asked as the
// trigger resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "99bba323-b5a4-4ff3-8ad1-2a19ed119069",
		Name:            "Fluxcharger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(WheneverYouCast(Or(Instant(), Sorcery()), "Fluxcharger — you may switch its power and toughness",
				switchThisPT("Fluxcharger — switch its power and toughness")),
				"Switch Fluxcharger's power and toughness until end of turn?"),
		},
	})
}
