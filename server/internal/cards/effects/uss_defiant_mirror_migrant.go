package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// U.S.S. Defiant, Mirror Migrant — Legendary Artifact — Spacecraft {2},
// 5/5 (Star Trek Commander, #4974):
//
//	"Flash
//	 When U.S.S. Defiant enters, it and up to one target creature you
//	 control phase out. (Treat them and anything attached to them as
//	 though they don't exist until your next turn.)
//	 Station
//	 8+ | Flying"
//
// Guardian of Faith's phasing (CR 702.26) over the Defiant itself and the
// trigger's target, in one call so they phase out together, taking their
// Auras and Equipment with them (CR 702.26g). The Defiant phases out only
// while it is still the object that entered (CR 400.7); a target gone by
// resolution is skipped (CR 608.2b). They phase in during their
// controller's next untap step. Station is the ordinary keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c7c529c4-08eb-480b-a699-0d783215419b",
		Name:            "U.S.S. Defiant, Mirror Migrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated:       []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(8, 5, 5),
			ThresholdKeywords(8, "flying"),
		),
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("U.S.S. Defiant — it and up to one target creature you control phase out",
				func(g *game.Game, item *game.StackItem) error {
					targets := append([]uuid.UUID{item.SourceCardID}, legalTargetCards(item, g)...)
					return PhaseOut{Targets: targets}.Apply(NewContext(g, item))
				}),
				TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1)),
		},
	})
}
