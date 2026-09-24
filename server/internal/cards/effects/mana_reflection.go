package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Reflection — Enchantment {4}{G}{G} (EDHREC rank ~1300):
//
//	"If you tap a permanent for mana, it produces twice as much of
//	 that mana instead."
//
// The card the mana-production seam was waiting on, named on the
// registry's "Mana-production replacement event" row since the batch-02
// census. The engine could already put mana in a pool from four
// places; what there was no event for is the AMOUNT, and #1222 built
// it (game.RepEventProduceMana, one window per production, opened
// before any of it is in the pool).
//
// So the whole card is one line. What is worth knowing is where the
// window opens for each shape of source, because "twice as much of
// THAT mana" is a question about a settled colour:
//
//	Forest, Sol Ring     doubled at the activation — the colour is
//	                     printed, so {G} becomes {G}{G} and {C}{C}
//	                     becomes {C}{C}{C}{C}
//	Birds of Paradise    doubled at the PICK. It is still ONE choice
//	                     (CR 106.12b replaces how much mana is
//	                     produced, not how many choices you get), so
//	                     the two mana are the same colour and a
//	                     doubled Birds cannot pay {W}{U}
//	Gilded Lotus         one pick, six mana
//	the auto-tapper      the same window, from the executor's own
//	                     taps — and the PLANNER prices through the
//	                     same predicate, so a cast that one
//	                     Mana-Reflected land pays taps one land
//
// What it does NOT double, and each is the printed text rather than a
// gap: a spell's "Add {B}{B}{B}" (Dark Ritual was not tapped), a mana
// ability with no {T} in its cost, and a triggered mana ability's own
// output (Wild Growth's extra {G} comes from the Aura, which nobody
// tapped). game.ReplacementEvent.ManaFromTap is the bit all three
// read.
//
// A second doubler composes: with Nyxbloom Ancient out, a Forest makes
// six. The CR 616.1 ordering is not put to anybody — a production
// cannot pause (CR 605.3b) — and it does not need to be, because
// multiplication commutes. Two Mana Reflections are ×4 with no prompt
// either, by #792's identical-window skip.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a3a8a044-283d-443e-bc40-c2f826d70c22",
		Name:         "Mana Reflection",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			YouTapForTwiceAsMuchMana("Mana Reflection: twice as much of that mana"),
		},
	})
}
