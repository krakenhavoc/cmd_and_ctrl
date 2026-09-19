package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flawless Maneuver — Instant {2}{W}:
//
//	"If you control a commander, you may cast this spell without
//	 paying its mana cost.
//	 Creatures you control gain indestructible until end of turn."
//
// The white member of the Commander Legends free-spell cycle, and a
// Teferi's Protection for one board wipe. Two pieces, both of which
// exist: `FreeIfYouControlCommander` is the conditional alternative
// cost (S28), and `GrantKeywordUntilEOT` is the mass layer-6 grant
// (S32). The batch-01 triage filed it under three blockers at once —
// cost modification, until-end-of-turn, protection — and all three
// have since landed.
//
// "Creatures you control", not "permanents": Heroic Intervention
// answers an Armageddon and this does not. The affected set is
// snapshotted as the spell resolves (CR 611.2c), so a creature cast
// afterwards this turn is not indestructible, and one that leaves and
// comes back is a new object without the grant (CR 400.7).
//
// Indestructible stops destruction and lethal damage (CR 702.12b) and
// nothing else — a sacrifice, an exile or a -X/-X still kills, which
// is why this answers a Wrath and not a Toxic Deluge.
//
// "You control a commander" is a commander PERMANENT you control: one
// in the command zone does not switch the offer on, and one you have
// stolen does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e183439-17d2-47ff-9d99-5e22821d91e3",
		Name:         "Flawless Maneuver",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			FreeIfYouControlCommander("Cast without paying its mana cost (you control a commander)"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"indestructible"},
				Label:    "Flawless Maneuver — indestructible",
			}.Apply(ctx)
		},
	})
}
