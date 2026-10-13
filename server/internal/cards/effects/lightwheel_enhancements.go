package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightwheel Enhancements — Enchantment — Aura {W}:
//
//	"Enchant creature or Vehicle
//	 Start your engines! (If you have no speed, it starts at 1. It
//	 increases once on each of your turns when an opponent loses life.
//	 Max speed is 4.)
//	 Enchanted permanent gets +1/+1 and has vigilance.
//	 Max speed — You may cast this card from your graveyard."
//
// The card #3105 was opened for, and the proof of Spec.CastableZonesIf
// (the 2026-10-13 amendment to ADR 0066): the graveyard is a cast zone
// for this card ONLY while its owner has max speed. The cast button
// appears in the graveyard the moment speed reaches 4 and goes again if
// speed ever fell, because the engine, the view, the enumerator and the
// bot all read the condition through one predicate. The rest of the
// card is ordinary: ADR 0138's start your engines! keyword, and an
// Aura whose "enchanted permanent" pump lands on a Vehicle too (an
// uncrewed Vehicle's +1/+1 is in its characteristics, ready for when it
// becomes a creature).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "17fae94e-558f-449a-b30c-e3c37a068b79",
		Name:            "Lightwheel Enhancements",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Targets:         TargetPermanent("enchant creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("vigilance"),
		},
		CastableZonesIf: []game.ConditionalCastZone{MaxSpeedCastableFrom(game.ZoneGraveyard)},
	})
}
