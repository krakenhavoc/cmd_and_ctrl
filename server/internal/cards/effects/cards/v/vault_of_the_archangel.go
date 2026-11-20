package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "eeaac65a-3480-475a-bb28-e6375d53f487",
		Name:         "Vault of the Archangel",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{B}, {T}: Creatures you control gain deathtouch and lifelink until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"deathtouch", "lifelink"}}, Answers: game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{2}{W}{B}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"deathtouch", "lifelink"},
					Label:    "Vault of the Archangel — deathtouch and lifelink",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
