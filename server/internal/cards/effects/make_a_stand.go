package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Make a Stand — Instant for {2}{W}:
//
//	"Creatures you control get +1/+0 and gain indestructible until
//	 end of turn."
//
// The white half of the protection suite: a combat trick that wins
// the whole combat step rather than one exchange. Every blocker
// trades up, nothing you control dies, and the +1/+0 turns a wall of
// even trades into a wall of one-sided ones.
//
// Two primitives — a layer 7c pump and a layer 6 grant — for the
// reason Overrun's file sets out: one registry entry sorts into one
// layer bucket, and these genuinely belong in two.
//
// # The +1/+0 is why the toughness half stays honest
//
// `BoostUntilEOT` adds to toughness too when asked; this card asks
// for zero, so a 2/2 stays a 3/2 and dies to 2 damage the moment
// the grant expires at cleanup. That is the printed behaviour and it
// only works because indestructible does not absorb marked damage —
// the damage is still there when the keyword goes away, it just
// stops mattering until then (CR 702.12b, and see
// game/indestructible.go). In practice the creature survives anyway,
// because cleanup wipes marked damage in the same step it sweeps the
// grant.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "531f78d5-5004-4b02-99c7-b390cb342fd9",
		Name:         "Make a Stand",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			yours := And(Creature(), YouControl())
			if err := (BoostUntilEOT{
				Match:     yours,
				Power:     1,
				Toughness: 0,
				Label:     "Make a Stand — +1/+0",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    yours,
				Keywords: []string{"indestructible"},
				Label:    "Make a Stand — indestructible",
			}.Apply(ctx)
		},
	})
}
