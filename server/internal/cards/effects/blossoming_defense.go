package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blossoming Defense — Instant for {G}:
//
//	"Target creature you control gets +2/+2 and gains hexproof until
//	 end of turn."
//
// One green mana that answers a removal spell and pushes two extra
// damage through in the same breath, which is why it is the default
// protection slot in every voltron shell that can cast it.
//
// Two primitives because the +2/+2 and the hexproof live in different
// CR 613 layers — 7c and 6 — and one registry entry sorts into one
// bucket. Overrun's file makes that argument at length; every card in
// this family repeats the shape.
//
// # "Target creature YOU CONTROL", and why that clause is load-bearing
//
// The predicate is not decoration. Hexproof only stops your
// OPPONENTS targeting a permanent (CR 702.11b), so the protection
// half is worthless pointed at an opponent's creature, and the +2/+2
// half would be a gift. `TargetCreature(..., YouControl())` refuses
// the pick at announce (CR 601.2c) and again at resolution
// (CR 608.2b).
//
// # It does not fizzle the removal spell it is answering
//
// The classic line — opponent casts Doom Blade, you respond with
// this — works because the Blade's target is re-checked on
// resolution against a creature that now has hexproof, so the Blade
// has no legal target and is countered by the rules. The engine
// re-runs the same predicate at resolution, so the interaction is
// real rather than reproduced by hand.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "5a851367-1c4a-4cc9-a9c3-3d2775986b4c",
		Name:     "Blossoming Defense",
		Targets:  TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (BoostUntilEOT{
				Target:    target,
				Power:     2,
				Toughness: 2,
				Label:     "Blossoming Defense — +2/+2",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"hexproof"},
				Label:    "Blossoming Defense — hexproof",
			}.Apply(ctx)
		},
	})
}
