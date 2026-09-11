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
// DECLARED SIMPLIFICATION — NO CYCLING, exactly as on Raffine's
// Tower, and for exactly the same two engine reasons:
//
//   - game.AbilityCost has no discard component. Spec.AdditionalCost's
//     DiscardCost is a spell-cast cost, not an ability cost.
//   - The CR 602 activation path only offers abilities on
//     battlefield permanents, and cycling is activated FROM HAND.
//
// Shipping without it makes the Triome strictly worse than printed,
// which is the safe direction. Cycling {3} is a late-game mana sink
// and a mulligan-smoother; the land's main job — fixing three
// colours and being fetchable — is all live.
func init() {
	Register(Spec{
		OracleID:     "6bae00e8-06cf-4ac4-a1cc-757e454109fe",
		Name:         "Ketria Triome",
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U|R}",
			Label:    "Add {G}, {U}, or {R}",
		}},
	})
}
