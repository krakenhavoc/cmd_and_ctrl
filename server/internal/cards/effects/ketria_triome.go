package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ketria Triome — Land — Forest Island Mountain:
//
//	"({T}: Add {G}, {U}, or {R}.)"
//	"This land enters tapped."
//	"Cycling {3}"
//
// The Ikoria Triome. Its three real basic land types are the reason
// it is played over a plain tri-land: a Misty Rainforest fetches it,
// a checkland and a Shifting Woodland see it. Those all fall out of
// the printed TypeLine through IsLandWithSubtype and need nothing
// here.
//
// The mana ability is declared rather than derived for the reason
// every nonbasic dual in the catalog declares it: the engine's
// synthetic land ability only fires for lands with the BASIC
// supertype (ManaAbilitiesForCard → basicLandColor), so a Triome's
// reminder-text ability has to be spelled out or the land taps for
// nothing.
//
// Cycling {3} arrived with #660 — the late-game mana sink and
// mulligan-smoother that made this a four-of over the tri-lands. It
// is an activated ability from hand (CR 702.29a), which is why it
// took a zone dimension on the activation path rather than an
// alternative cast cost (ADR 0062).
func init() {
	Register(Spec{
		OracleID:     "6bae00e8-06cf-4ac4-a1cc-757e454109fe",
		Name:         "Ketria Triome",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U|R}",
			Label:    "Add {G}, {U}, or {R}",
		}},
		Activated: []ActivatedAbility{Cycling("{3}")},
	})
}
