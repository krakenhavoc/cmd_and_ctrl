package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grafted Butcher — Creature — Phyrexian Samurai {1}{B}, 2/2 (March of
// the Machine):
//
//	"When this creature enters, Phyrexians you control gain menace until
//	 end of turn.
//	 Other Phyrexians you control get +1/+1.
//	 {3}{B}, Sacrifice an artifact or creature: Return this card from
//	 your graveyard to the battlefield. Activate only as a sorcery."
//
// The set of Phyrexians that gain menace is fixed as the trigger resolves
// (CR 611.2c). The lord is a layer-7c anthem. The graveyard ability pays
// {3}{B} and sacrifices an artifact or creature as costs and may be
// activated only from the graveyard at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6c4a0e3d-c63c-4acf-8386-b51f1ec37463",
		Name:         "Grafted Butcher",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Phyrexian"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Grafted Butcher — Phyrexians you control gain menace until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    And(OfCreatureType("Phyrexian"), YouControl()),
						Keywords: []string{"menace"},
						Label:    "Grafted Butcher — Phyrexians gain menace",
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}{B}, Sacrifice an artifact or creature: Return this card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost: Plus(ManaCost("{3}{B}"), game.AbilityCost{
				SacrificeOther: sacrificeSpec("an artifact or creature", Or(Artifact(), Creature())),
			}),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return returnThisFromGraveyardToBattlefield(g, item)
			},
		}},
	})
}
