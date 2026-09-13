package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Worn Powerstone — Artifact {3}:
//
//	"Worn Powerstone enters tapped. {T}: Add {C}{C}."
//
// The enters-tapped clause is the drawback that makes a two-mana rock
// cost three, so it is implemented rather than deferred.
//
// It used to be an OnETB tap — enter untapped, then tap — with a note
// that a true enters-tapped needed a CR 614 replacement on an event the
// engine builds before the card is placed. That machinery exists now
// (SelfEntersTapped, added with the Temple cycle), so this is the real
// replacement: the Powerstone is never untapped on the battlefield.
// The old approximation was observable to anything watching for a tap
// or for an untapped permanent entering.
func init() {
	Register(Spec{
		OracleID:     "b166b670-febc-4821-855e-f8d465644c03",
		Name:         "Worn Powerstone",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}
