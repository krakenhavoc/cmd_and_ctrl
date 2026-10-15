package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archway of Innovation — Land (Duskmourn: House of Horror):
//
//	"This land enters tapped unless you control an Island.
//	 {T}: Add {U}.
//	 {U}, {T}: The next spell you cast this turn has improvise. (Your
//	 artifacts can help cast that spell. Each artifact you tap after
//	 you're done activating mana abilities pays for {1}.)"
//
// The grant is a one-use NextSpellPromise carrying Improvise
// (game/next_spell_promise.go): while it stands, the caster's next spell
// is offered the improvise tap cost (Improvise(), CR 702.126). It is
// spent by the next spell cast this turn, whether or not that spell used
// it, and lapses at end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bfa20bc7-4626-4a52-87f4-6e2763cb8ed5",
		Name:         "Archway of Innovation",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b06EntersTappedUnlessLandType("island")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{U}, {T}: The next spell you cast this turn has improvise.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{U}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantNextSpellPromise{From: "Archway of Innovation", Promise: game.NextSpellPromise{
					Improvise: true,
					Text:      "The next spell you cast this turn has improvise.",
				}}.Apply(NewContext(g, item))
			},
		}},
	})
}
