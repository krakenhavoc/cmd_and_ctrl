package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Creeping Trailblazer — Creature — Elemental {R}{G}, 2/2 (Core Set 2020, #5518):
//
//	"Other Elementals you control get +1/+0.
//	 {2}{R}{G}: This creature gets +1/+1 until end of turn for each Elemental
//	 you control."
//
// Bladestitched Skaab's tribal anthem over other Elementals you control.
// The pump counts the Elementals as it resolves, the Trailblazer itself
// included, Slumbering Keepguard's shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94431642-cfcf-48f4-9e0b-fcddc5b051ef",
		Name:         "Creeping Trailblazer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Elemental"}, Others: true, YoursOnly: true}, 1, 0),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}{G}: This creature gets +1/+1 until end of turn for each Elemental you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
					return nil
				}
				n := m20ElementalsYouControl(g, item.Controller)
				return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n,
					Label: "Creeping Trailblazer — +1/+1 for each Elemental you control"}.Apply(NewContext(g, item))
			},
		}},
	})
}

// m20ElementalsYouControl is "the number of Elementals you control":
// every permanent the player controls with the Elemental subtype, read
// through the layers (Creeping Trailblazer, Lavakin Brawler).
func m20ElementalsYouControl(g *game.Game, player uuid.UUID) int {
	return len(permanentsControlledByMatching(g, player, Subtype("Elemental")))
}
