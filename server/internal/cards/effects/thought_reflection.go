package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thought Reflection — Enchantment {4}{U}{U}{U}:
//
//	"If you would draw a card, draw two cards instead."
//
// The draw-amount seam's clean card: no "except the first", no second
// clause, just the count. The draw event has existed since S17 and
// Notion Thief has been rewriting its DRAWING PLAYER since S20; what
// there was no room on it for is HOW MANY, and #1222 added it
// (game.ReplacementEvent.DrawCount).
//
// CR 121.2 is why the count's base is one and why that is enough:
// "draw three cards" is three individual card draws, so a Thought
// Reflection on Divination draws four and on Blue Sun's Zenith for
// five draws ten, each draw its own window. The two cards of a doubled
// draw are still drawn one at a time, so Nekusar, Sheoldred,
// Consecrated Sphinx and every other per-card payoff fires per card —
// but they are ONE event, and CR 614.5's once-per-event tracking is
// what stops the Reflection from doubling its own output forever.
//
// Two Thought Reflections draw FOUR, not three, with no prompt: the
// second multiplies the count the first left, through the ordinary
// CR 616.1 apply-loop, and #792's identical-window skip means nobody
// is asked which Reflection went first. Beside an Alhammarret's
// Archive it is also four, and there the ordering prompt IS asked,
// because two different declared effects are ordered by the affected
// player even when every ordering agrees.
//
// Notion Thief composes the other way: it rewrites who draws, this
// rewrites how many, and the drawing player orders them. Thief first
// and the Thief's controller draws two; this first and the Thief takes
// a draw of two.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "692a6833-3014-42f6-b1ad-333bb3292c65",
		Name:         "Thought Reflection",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			YouDrawTwiceInstead("Thought Reflection: draw two cards instead"),
		},
	})
}
