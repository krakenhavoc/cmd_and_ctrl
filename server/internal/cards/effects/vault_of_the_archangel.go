package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vault of the Archangel — Land (EDHREC top 1,000, #3654):
//
//	"{T}: Add {C}.
//	 {2}{W}{B}, {T}: Creatures you control gain deathtouch and lifelink
//	 until end of turn."
//
// A colourless mana ability and a mass keyword grant. The set of
// creatures is fixed as the ability resolves (CR 611.2c): one that
// enters later this turn does not gain the keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eeaac65a-3480-475a-bb28-e6375d53f487",
		Name:         "Vault of the Archangel",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}{B}, {T}: Creatures you control gain deathtouch and lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    Plus(ManaCost("{2}{W}{B}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"deathtouch", "lifelink"},
					Label:    "Vault of the Archangel — deathtouch and lifelink",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
