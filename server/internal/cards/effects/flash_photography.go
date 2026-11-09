package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flash Photography — Sorcery {2}{U}{U} (FIN, #4462, #5363):
//
//	"You may cast this spell as though it had flash if it targets a
//	 permanent you control.
//	 Create a token that's a copy of target permanent.
//	 Flashback {4}{U}{U}"
//
// The conditional flash is FlashIfTargets, the card's own timing
// permission (Spec.FlashIf, #5207). A cast that is open only because of
// it is announced under "target permanent you control" (CR 601.3b: "If
// any such choices could cause that effect to apply, that player may
// begin to cast that spell as though it had flash"); a sorcery-speed
// cast keeps the printed "target permanent", anybody's. The permission
// is the card's own and reads no zone, so it opens the window for the
// flashback cast from the graveyard too: flashback is an alternative
// cost (CR 702.34a) and the timing permission composes with it.
//
// The body is Relm's Sketching's and Cackling Counterpart's,
// TokenCopyOfSingleTarget: the token takes the copiable values (CR
// 707.2), so what the original is itself copying comes along and its
// counters, Auras and other effects do not. An {X} in the copied mana
// cost is treated as 0 off the stack (CR 107.3g), and the token enters
// through the ordinary pipeline, so the copy's enters abilities trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a92ae7b0-cbbb-44a8-bfd8-2a68f4c1ce1a",
		Name:             "Flash Photography",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{U}{U}")},
		Targets:          TargetPermanent("target permanent"),
		FlashIf:          FlashIfTargets("a permanent you control", TargetPermanent("target permanent you control", YouControl())),
		OnResolve:        TokenCopyOfSingleTarget,
	})
}
