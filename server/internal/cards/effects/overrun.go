package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overrun — Sorcery for {2}{G}{G}{G}:
//
//	"Creatures you control get +3/+3 and gain trample until end of
//	 turn."
//
// The mass half of the S32 acceptance pair, and deliberately chosen
// over Heroic Intervention: at the time indestructible was still
// unenforced (#176), so Heroic Intervention would have shipped as a
// declared no-op, whereas every word of Overrun is live. (S25's #77
// closed that gap and Heroic Intervention now ships for real — see
// heroic_intervention.go.) Trample IS
// one of the twelve keywords the combat code honours — the
// combat-damage assignment reads HasKeyword(attacker, "trample")
// off the post-layer effective characteristic, so a granted trample
// really does push excess damage through.
//
// Two primitives rather than one because the +3/+3 and the trample
// genuinely belong to different CR 613 layers — 7c and 6 — and one
// registry entry can only sort into one bucket.
//
// CR 611.2c IS OBSERVED: the affected set is snapshotted when
// Overrun resolves, so a creature cast afterwards this turn gets
// neither the pump nor the trample, and a creature flickered out
// and back loses both (CR 400.7 — it returns as a new object). The
// snapshot lives in the `BoostUntilEOT` / `GrantKeywordUntilEOT`
// primitives; see until_end_of_turn.go.
//
// Sorcery, so no timing subtlety: it can only be cast in a main
// phase, which in practice means precombat main if you want the
// attack. Casting it postcombat still grants for the rest of the
// turn.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "204f9afe-c20b-4933-b5cd-aa572784762a",
		Name:     "Overrun",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			yours := And(Creature(), YouControl())
			if err := (BoostUntilEOT{
				Match:     yours,
				Power:     3,
				Toughness: 3,
				Label:     "Overrun — +3/+3",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    yours,
				Keywords: []string{"trample"},
				Label:    "Overrun — trample",
			}.Apply(ctx)
		},
	})
}
