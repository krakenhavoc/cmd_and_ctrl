package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian of Faith — Creature — Spirit Knight {1}{W}{W}, 3/2 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Flash
//	 Vigilance
//	 When this creature enters, any number of other target creatures you
//	 control phase out. (Treat them and anything attached to them as
//	 though they don't exist until their controller's next turn.)"
//
// Clever Concealment's phasing over the ETB trigger's targets: "any
// number" is Min 0, Max unbounded, and "other" is object identity
// (Another). The targets still legal as it resolves (CR 608.2b) phase out
// together in one call (CR 702.26a), taking their Auras and Equipment with
// them; an attacker or blocker that phases out leaves combat. They phase
// in during their controller's next untap step (CR 702.26d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8e04a5e0-e22c-430b-a39a-c9f38e16ea86",
		Name:            "Guardian of Faith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Guardian of Faith — any number of other target creatures you control phase out",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PhaseOut{Targets: effects.LegalTargetCards(item, g)}.Apply(effects.NewContext(g, item))
				}),
				effects.Another(effects.TargetCreature("any number of other target creatures you control", effects.YouControl())).WithCount(0, 0)),
		},
	})
}
