package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Library of Leng — Artifact {1} (EDHREC rank 3000):
//
//	"You have no maximum hand size.
//	 If an effect causes you to discard a card, discard it, but you may
//	 put it on top of your library instead of into your graveyard."
//
// Skipped by batch 28 (#390) and waiting ever since, because no
// discard reached the CR 614 window and — once #853 sent one through —
// nothing on the event said it was a discard or why. #650 gave it
// both, and the card is now two ordinary declarations.
//
// The clause the rules actually draw is the CAUSE, not "voluntary"
// (ADR 0013 §10a withdrew that framing; ADR 0061 is the shape that
// replaced it):
//
//   - It applies to an EFFECT's discard — Mind Rot, looting — whoever
//     controls the spell or ability (Gatherer ruling, 2004-10-04).
//   - It does NOT apply to a discard paid as a COST, because costs are
//     not effects (same ruling). Thirst for Knowledge's "unless you
//     discard an artifact card" branch is a cost too (CR 118.12a), and
//     it is out for the same reason.
//   - It does NOT apply to the CLEANUP discard, which is a turn-based
//     action (CR 514.1, 703.1) rather than anybody's effect. Leng
//     usually removes your maximum hand size — but a later hand-size
//     effect can set one again (hand-size effects apply in timestamp
//     order; ruling, 2009-10-01), so the cleanup discard can still
//     happen, and Leng still does not replace it.
//
// "Discard it, but you may put it on top of your library instead" is
// exactly what the engine does: the discard HAPPENED — EventDiscardCard
// fires with the destination the window settled on (CR 701.8a defines
// a discard by the move out of the hand) — and only where the card
// ended up changed. So Megrim, Containment Construct and every other
// "whenever you discard" payoff still sees it.
//
// It is a "may", so a yes/no prompt is offered per card,
// which is the printed granularity when one effect discards several.
//
// Sandbox simplification, declared: when one effect discards several
// cards and more than one is sent to the library, they go in the order
// the discard batch fixed rather than in an order Leng's controller
// chooses. Nothing about which cards are saved changes — only the
// library order of cards nobody has seen.
func init() {
	Register(Spec{
		OracleID:      "867def48-4be8-4056-bcf1-d6b00450b9a3",
		Name:          "Library of Leng",
		Completeness:  CompletenessCaveats,
		Caveats:       []string{"When one effect discards several cards and you send more than one to your library, you don't choose the order they end up in."},
		NoMaxHandSize: true,
		Replacements: []game.ReplacementEffect{
			DiscardBecomes{
				Dst:      game.ZoneLibrary,
				Causes:   []game.DiscardCause{game.DiscardCauseEffect},
				Optional: true,
				Label:    "Library of Leng: on top of your library instead",
				Question: "Put the discarded card on top of your library instead?",
			}.Build(),
		},
	})
}
