package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alhammarret's Archive — Legendary Artifact {5} (EDHREC rank ~250):
//
//	"If you would gain life, you gain twice that much life instead.
//	 If you would draw a card except the first one you draw in each of
//	 your draw steps, draw two cards instead."
//
// Two replacements on two different events, and the card has been
// waiting on the second one. The LIFE half has been writable since
// #482 put every writer of a life total through the CR 614 window —
// it is Rhox Faithmender's clause word for word, so it doubles a
// catalog GainLife, a drain's gain half and any lifelink credit. The
// DRAW half needed a COUNT on the draw event, which is #1222.
//
// The two halves are separate slots in Replacements, so they are
// separate declared effects: an Archive beside a Rhox Faithmender is
// ×4 life with a CR 616 ordering prompt, and an Archive beside a
// Thought Reflection is ×4 cards with one. Two Archives are ×4 on both
// halves with no prompt at all (#792's identical-window skip), which
// is what the printed card does.
//
// Only the first completed draw in each own draw step is exempt.
func init() {
	Register(Spec{
		OracleID:     "d5c4d36c-54b3-4149-906b-a57a670260fc",
		Name:         "Alhammarret's Archive",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			YouGainTwiceThatMuchLife("Alhammarret's Archive: gain twice that much life"),
			YouDrawTwiceInsteadExceptTheFirst("Alhammarret's Archive: draw two cards instead"),
		},
	})
}
