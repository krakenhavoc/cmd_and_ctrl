package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Graceful Antelope — Creature — Antelope {2}{W}{W}, 1/4:
//
//	"Plainswalk (This creature can't be blocked as long as defending
//	 player controls a Plains.)
//	 Whenever this creature deals combat damage to a player, you may have
//	 target land become a Plains until this creature leaves the
//	 battlefield."
//
// Plainswalk is the printed keyword (CR 702.14). The trigger is ADR 0109
// §1's (#1881) CR 305.7 type set for as long as the Antelope stays on the
// battlefield (CR 611.2b): the land is a Plains (other subtypes stay, CR
// 205.1a), loses its rules-text abilities and taps for {W} (CR 305.6).
// Each hit gives the next defending player a Plains to walk through. An
// Antelope gone before the trigger resolves changes nothing.
//
// No simplification.
func init() {
	t := WheneverThisDealsCombatDamageToAPlayer("Graceful Antelope — target land becomes a Plains",
		TargetLandBecomesWhileSourceRemains("Graceful Antelope", "Plains"))
	t.Targets = TargetPermanent("target land", Land())
	Register(Spec{
		OracleID:        "8dac693b-c5c5-4d80-9739-a3b95949a213",
		Name:            "Graceful Antelope",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"plainswalk"},
		Triggered: []game.TriggeredAbility{
			Optional(t, "Graceful Antelope — have target land become a Plains until it leaves the battlefield?"),
		},
	})
}
