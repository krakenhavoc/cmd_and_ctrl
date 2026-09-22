package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhox Faithmender — Creature — Rhino Monk {3}{W}, 1/5 (EDHREC rank
// 1620):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)
//	 If you would gain life, you gain twice that much life instead."
//
// The lifegain doubler. A CR 614 replacement on the life-change
// event, applied only to a GAIN (a positive delta) for the
// Faithmender's own controller — a life loss or a payment is not a
// gain and goes through untouched.
//
// Since #482 every writer of a life total runs that window and lands
// in the one tail (game/life_tail.go), so this doubles a catalog
// GainLife, a drain's gain half, and the Faithmender's own printed
// lifelink (CR 702.15b: lifelink is life gain). Until then it ran on
// the public ChangePlayerLife sandbox verb alone — the life change a
// player makes by dragging their own counter — and the comment here
// claimed the lifelink it did not actually double. Two Faithmenders
// quadruple, with no prompt: two applicable replacements are normally
// a CR 616 ordering question, but these two are the same printed
// effect and every order is x4, so #792 applies them inline.
//
// DAMAGE is not this window. A Lightning Helix's three damage is not
// reduced or doubled by a life replacement (CR 120.3); only its three
// life of gain is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2abe9303-d498-4aad-b6b2-8b5064bd2ffd",
		Name:            "Rhox Faithmender",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Replacements: []game.ReplacementEffect{
			// #1222: shared with Alhammarret's Archive, which prints
			// the same sentence. This was an inline clause while it was
			// the only one; life_replacements.go is where the family
			// lives now.
			YouGainTwiceThatMuchLife("Rhox Faithmender: gain twice that much life"),
		},
	})
}
