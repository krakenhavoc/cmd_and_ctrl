package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moss Diamond — Artifact {2} (EDHREC rank 4341):
//
//	"This artifact enters tapped.
//	 {T}: Add {G}."
//
// The green Diamond. Two colourless for a permanent {G} source that
// costs you the turn it comes down — strictly worse than a Signet in a
// two-colour deck and strictly better than nothing in a deck that
// needs its third green source on turn three.
//
// Enters-tapped is a real CR 614 self-replacement, not a hook that taps
// it a beat later: the Diamond is never untapped on the battlefield and
// no EventTapCard is emitted, so nothing watching for a tap or for an
// untapped artifact entering misreads it. That distinction is why the
// mana is genuinely unavailable the turn it lands, rather than
// available for one priority window.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02500f21-6e15-423e-93ff-891e09fe9904",
		Name:         "Moss Diamond",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
