package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gain_lands.go — the Khans of Tarkir "gain lands":
//
//	"This land enters tapped.
//	 When this land enters, you gain 1 life.
//	 {T}: Add {A} or {B}."
//
// The budget dual every precon ships. Two of the ten are in the
// roadmap's batch 06 (#299), one in batch 08 (#301) and six in batch
// 09 (#302); the last belongs in this table when its batch reaches it
// — a card that belongs to an existing cycle goes in the cycle's
// table, never in a new file.
//
// Enters-tapped is the self-replacement; the life is a real ETB
// trigger with a response window, as printed; the dual is the pipe
// every two-colour land in the catalog uses. Batch 13 (#306) brings
// Rugged Highlands, the tenth — the cycle is complete. The Zendikar
// "refuge" cycle (Graypelt Refuge, Kazandu Refuge, …) prints the
// identical three lines, so its members join this table as their
// batches reach them rather than opening a second file for the same
// card.
//
// No simplification.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"d37f858e-03c8-4594-9b92-cd03699a1591", "Scoured Barrens", "W", "B"},
		{"6de714e1-446d-4fb9-9e3d-bcd3ec6af9ca", "Jungle Hollow", "B", "G"},
		{"64e29bfc-9313-4e8c-808c-bc27f6b018a6", "Bloodfell Caves", "B", "R"}, // batch 08 (#301)
		// Roadmap batch 09 (#302) — six more of the ten; that is nine.
		{"865a2194-fca0-446e-aae3-ca475cd66e00", "Dismal Backwater", "U", "B"},
		{"b0af0c54-2a59-4075-8543-d41ff20c4c87", "Wind-Scarred Crag", "R", "W"},
		{"2f4ad084-2062-44c0-9975-15f100204531", "Swiftwater Cliffs", "U", "R"},
		{"45429b2c-be3b-4b2e-9bab-a059ccbda8cd", "Blossoming Sands", "G", "W"},
		{"ec96cde2-f1e6-495c-94e2-3e8ae79e556c", "Thornwood Falls", "G", "U"},
		{"5d641bf6-0f93-4189-8dc1-ec7ea446dade", "Tranquil Cove", "W", "U"},
		// Roadmap batch 13 (#306) — the tenth.
		{"6c922206-6e68-4dcd-9559-88da1074f2c4", "Rugged Highlands", "R", "G"},
		// Roadmap batch 26 (#388): the Zendikar "refuges" are the same
		// three lines under an older name, so they share the table.
		{"60b36821-0fad-423c-98c4-f64d991719f3", "Graypelt Refuge", "G", "W"},
		// Roadmap batch 36 (#399): the Rakdos refuge.
		{"354fecd1-2371-49e3-81c6-7e47728dbb1f", "Akoum Refuge", "B", "R"},
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
