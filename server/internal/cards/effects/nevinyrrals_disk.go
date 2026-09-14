package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nevinyrral's Disk — Artifact {4}:
//
//	"This artifact enters tapped.
//	 {1}, {T}: Destroy all artifacts, creatures, and enchantments."
//
// The oldest repeatable wrath in the game and still one of the best
// in Commander: it costs four, it announces itself a full turn in
// advance, and everyone at the table has to play around it from the
// moment it lands.
//
// # Enters tapped is a replacement, not an ETB tap
//
// SelfEntersTapped, the CR 614 self-replacement — so the Disk is
// never untapped on the battlefield for even an instant. That is the
// difference that makes the card fair: you cannot crack it the turn
// it arrives, because it enters already tapped and has to survive an
// untap step first. An OnETB tap would enter untapped and be tapped
// a beat later, and anything watching could act in between.
//
// # It destroys itself
//
// The Disk IS an artifact and the sweep includes it. That is printed
// behaviour, not an oversight: the Disk goes to the graveyard with
// everything else. The activation cost has already been paid by then
// ({1} and the tap, paid at announce, CR 602.2), so destroying the
// source mid-resolution costs nothing — and the sweep works off a
// snapshot taken before anything moves, so the Disk being in that
// snapshot is exactly right.
//
// One sweep over the three-way Or rather than three sweeps: the card
// prints one event, and an artifact creature should be destroyed
// once.
func init() {
	Register(Spec{
		OracleID:     "96230edf-568a-47dd-b877-9d92aa58fac8",
		Name:         "Nevinyrral's Disk",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Destroy all artifacts, creatures, and enchantments",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return DestroyAllMatching{
					Match: Or(Artifact(), Creature(), Enchantment()),
				}.Apply(ctx)
			},
		}},
	})
}
