package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heroic Intervention — Instant for {1}{G}:
//
//	"Permanents you control gain hexproof and indestructible until
//	 end of turn."
//
// The card S32 explicitly declined to write. Overrun's file records
// the decision: at the time, hexproof had a consumer but
// indestructible did not, so half of this card would have been a
// declared no-op and Overrun went in as the mass acceptance case
// instead. S25 (#77) shipped the destruction-path half, and the
// reason to hold this card back went with it.
//
// Two mass grants, one entry in the turn-scoped registry. Both
// keywords are layer-6 ability additions, so unlike a pump-plus-grant
// card (Blossoming Defense, Overrun) they genuinely do belong in a
// single `StaticAbility` and a single timestamp.
//
// # What "permanents you control" means here
//
// Everything, not just creatures: the lands, the mana rocks, the
// commander. `YouControl()` on its own is the whole predicate, which
// is why this card answers a Vandalblast and an Armageddon as
// readily as a Wrath.
//
// # CR 611.2c
//
// The affected set is snapshotted as the spell resolves, so a
// creature cast afterwards this turn is unprotected, and a permanent
// that leaves and re-enters comes back as a new object without the
// grant (CR 400.7). Both fall out of `eotSnapshot`; see
// until_end_of_turn.go.
//
// # The two keywords protect against genuinely different things
//
// Hexproof stops your opponents TARGETING the permanents
// (CR 702.11b) — Doom Blade, Beast Within, a Krosan Grip on your
// rock. Indestructible stops them being DESTROYED (CR 702.12b) —
// Wrath of God, Blasphemous Act, lethal combat damage. Neither
// stops a sacrifice effect, an exile effect, or a -X/-X, which is
// exactly why the card is an answer rather than an insurance policy,
// and the engine models all three of those gaps faithfully
// (see game/indestructible.go).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "24882fa2-3fe9-4c1b-aa3d-0e6488b9db27",
		Name:         "Heroic Intervention",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Indestructible saves a permanent from single-target removal and from lethal damage, but a board wipe (\"destroy all\") still destroys it."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GrantKeywordUntilEOT{
				Match:    YouControl(),
				Keywords: []string{"hexproof", "indestructible"},
				Label:    "Heroic Intervention — hexproof and indestructible",
			}.Apply(ctx)
		},
	})
}
