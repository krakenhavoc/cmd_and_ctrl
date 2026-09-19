package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thornglint Bridge — Artifact Land (EDHREC rank 4511):
//
//	"This land enters tapped.
//	 Indestructible
//	 {T}: Add {G} or {W}."
//
// Darksteel Citadel's shape with a colour pair: an ARTIFACT that taps
// for two colours and cannot be destroyed. The artifact type is the
// reason it is played — it feeds affinity, metalcraft and every
// artifact count in a Selesnya artifact deck — and the
// indestructibility is why it beats an ordinary tapland there:
// the sweepers those decks fear (Vandalblast, Bane of Progress) miss
// it.
//
// Indestructible has been live since S25 (#77) for single-target
// destruction and since S30 (#470) for the mass destroy path, so both
// halves of "survives a board wipe" are true of the implementation
// as well as the card. It rides PrintedKeywords, which is what makes
// the destruction path see it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "99720c65-be96-4220-8ed4-720660bf6928",
		Name:            "Thornglint Bridge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities:   []ManaAbility{dualManaAbility("G", "W")},
	})
}
