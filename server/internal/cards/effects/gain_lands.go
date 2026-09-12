package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gain_lands.go — the Khans of Tarkir "gain lands":
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {A} or {B}."
//
// The budget dual every precon ships. Two of the ten are in the
// roadmap's batch 06 (#299); the other eight belong in this table when
// their batches reach them — a card that belongs to an existing cycle
// goes in the cycle's table, never in a new file.
//
// Enters-tapped is the self-replacement; the life is a real ETB
// trigger with a response window, as printed; the dual is the pipe
// every two-colour land in the catalog uses.
//
// No simplification.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"d37f858e-03c8-4594-9b92-cd03699a1591", "Scoured Barrens", "W", "B"},
		{"6de714e1-446d-4fb9-9e3d-bcd3ec6af9ca", "Jungle Hollow", "B", "G"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			Completeness:  CompletenessFull,
			Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
			Triggered: []game.TriggeredAbility{{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, source.Name+" — you gain 1 life",
						func(g *game.Game, item *game.StackItem) error {
							return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
						})
				},
			}},
		})
	}
}
