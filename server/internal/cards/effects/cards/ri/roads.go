package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Roads — Land, the five-colour Aetherdrift cycle (#3552):
//
//	"This land enters tapped unless you control a Mount or Vehicle.
//	 {T}: Add {C}.
//	 {1}{C}, {T}, Sacrifice this land: Create a 1/1 colorless Pilot
//	 creature token with "This token saddles Mounts and crews Vehicles
//	 as though its power were 2 greater." Activate only as a sorcery."
//
// Country Roads {W}, Foul Roads {B}, Reef Roads {U}, Rocky Roads {R},
// Wild Roads {G}. Written as a table, like the Temples, because five
// near-identical files are five places to fix one mistake.
//
// The enters-tapped clause is Castle Ardenvale's shape with a different
// condition (SelfEntersTappedUnless); Mount and Vehicle are read as
// effective subtypes of the controller's permanents. The sacrifice is
// part of the cost, so the land is gone before the ability can be
// responded to, and the token is made for the ability's controller.
//
// No simplification.
var RoadsCycle = []struct {
	OracleID string
	Name     string
	Colour   string
}{
	{"c5a39f76-dd1b-442c-9f52-08561ecb91ad", "Country Roads", "W"},
	{"d4e4c8a5-e97b-4295-a403-d17834f73502", "Foul Roads", "B"},
	{"32438050-5ae7-4c19-bcaf-5a07a673e0e0", "Reef Roads", "U"},
	{"a659c29f-aaca-44c5-8426-cdafcb195f86", "Rocky Roads", "R"},
	{"36fbc8ba-bb4c-4e5e-9031-78c36e376851", "Wild Roads", "G"},
}

// youControlAMountOrVehicle is the Roads' untapped condition.
func youControlAMountOrVehicle(g *game.Game, controller uuid.UUID) bool {
	return effects.ControlsA("Mount")(g, controller) || effects.ControlsA("Vehicle")(g, controller)
}

func init() {
	for _, r := range RoadsCycle {
		tapped := effects.SelfEntersTappedUnless(youControlAMountOrVehicle)
		tapped.RowLabel = "Enters tapped unless you control a Mount or Vehicle."
		effects.Register(effects.Spec{
			OracleID:     r.OracleID,
			Name:         r.Name,
			Completeness: effects.CompletenessFull,
			Replacements: []game.ReplacementEffect{tapped},
			ManaAbilities: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{" + r.Colour + "}",
				Label:    "Add {" + r.Colour + "}",
			}},
			Activated: []effects.ActivatedAbility{{
				Label:        "{1}{" + r.Colour + "}, {T}, Sacrifice this land: Create a 1/1 colorless Pilot creature token with \"This token saddles Mounts and crews Vehicles as though its power were 2 greater.\" Activate only as a sorcery.",
				Purpose:      game.Purpose{Tokens: 1},
				Cost:         effects.Plus(effects.ManaCost("{1}{"+r.Colour+"}"), effects.TapCost(), effects.SacrificeThis()),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.PilotToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			}},
		})
	}
}
