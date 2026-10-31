package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wild Growth — Enchantment — Aura {G} (#294):
//
//	"Enchant land
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional {G}."
//
// The card #763 was written for, and the first TRIGGERED MANA ABILITY
// in the catalog (CR 605.1b). Its extra {G} does NOT use the stack and
// gives nobody a priority window (CR 605.4a): it is in the pool the
// moment the land taps, which is the only way a one-mana ramp Aura
// works at all — a stack trigger would arrive after the spell it is
// paying for had already been cast.
//
// So it is a Spec.ManaTriggers entry, not a Spec.Triggered one. See
// [ADR 0074](../../../../docs/decisions/0074-triggered-mana-abilities.md).
//
// The attachment is re-read on every production
// (WheneverAttachedTapsForMana), so moving or destroying the Aura
// moves or stops the trigger with no bookkeeping, and the CR 704.5m
// legality check reruns the same "enchant land" spec that targeted.
//
// Three things it composes with for free, because the seam is the
// permanent's ability list rather than the land's:
//
//   - the land's own colour pick — a Wild Growth on a dual land fires
//     once, when the pick is answered;
//   - the auto-tapper, which fires it from its executor, so the same
//     board pays the same way whether the player clicked the land or
//     pressed "Auto-tap & cast";
//   - CR 613.1f — an Aura whose abilities were removed has no trigger.
//
// The auto-tap PLANNER counts the extra {G} (ADR 0074, amendment of
// 2026-10-30). It asks this trigger's own condition through the judge
// the executor fires with, so a Forest under Wild Growth plans as {G}{G}
// and pays a two-mana cost alone. Paying a single {G} with it still
// floats the other, since a land cannot be half-tapped.
//
// No simplification of the card itself.
func init() {
	Register(Spec{
		OracleID:     "706ae742-1807-44b7-a4fa-f2e26f61519a",
		Name:         "Wild Growth",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		ManaTriggers: []game.ManaTrigger{
			WheneverAttachedTapsForMana("Wild Growth — add an additional {G}", "{G}"),
		},
	})
}
