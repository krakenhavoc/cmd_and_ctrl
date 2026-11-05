package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossal Clone Master — Creature — Giant {5}{U}, 0/0 (Star Trek
// Commander, #4974):
//
//	"Strange new worlds — You may have this creature enter as a copy of
//	 any creature on the battlefield, except it isn't legendary and it has
//	 "This creature's power and toughness are each equal to the number of
//	 differently named lands you control.""
//
// Phantasmal Image's shape (EntersAsCopyOf over any creature, a copy
// grant) with Spark Double's "isn't legendary". Both exceptions are part
// of the copy's copiable values (CR 707.9a, CR 707.9b), so a later copy of
// the Clone Master copies them too. The granted ability is a
// characteristic-defining ability (CR 604.3a) in layer 7a, which a copy
// grant may carry: it overrides the copied creature's printed power and
// toughness and counts each land you control once per name, every pass.
//
// Declining the copy leaves the printed 0/0 Giant, which dies to CR
// 704.5f. "Strange new worlds" is an ability word (CR 207.2c).
//
// No simplification.
const colossalCloneMasterGrant = "colossal-clone-master/differently-named-lands-pt"

func init() {
	const text = "This creature's power and toughness are each equal to the number of differently named lands you control."
	Register(Spec{
		OracleID:     "94fee2cf-7604-48fa-9d3c-03ca4af0edfd",
		Name:         "Colossal Clone Master",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:    colossalCloneMasterGrant,
			Text:   text,
			Static: []game.StaticAbility{trekB34DifferentlyNamedLandsPT(text)},
		}},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Colossal Clone Master",
				anyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.RemoveSupertype("Legendary")
					v.GrantAbility(colossalCloneMasterGrant)
				},
			),
		},
	})
}
