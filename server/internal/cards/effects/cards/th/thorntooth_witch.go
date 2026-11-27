package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thorntooth Witch — Creature — Treefolk Shaman {5}{B}, 3/4 (Lorwyn, S82
// OLD b545, #5593):
//
//	"Whenever you cast a Treefolk spell, you may have target creature get
//	 +3/-3 until end of turn."
//
// Blightcaster's optional targeted boost on a cast trigger, with a Treefolk
// spell: the spell's type is read through the changeling rule, so a
// changeling spell is a Treefolk (CR 702.73a). The Witch's own cast does not
// trigger it, as it is not on the battlefield.
//
// No Purpose: +3/-3 is partly a benefit and partly a harm, which
// TargetPurpose.UntilEndOfTurn leaves undeclared (#4312), as it does for
// Nameless Inversion.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "942eefa3-6775-428b-84aa-48c3bd78d3b8",
		Name:         "Thorntooth Witch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WheneverYouCast(effects.OfCreatureType("Treefolk"), "Thorntooth Witch — target creature gets +3/-3 until end of turn",
					effects.BoostTargetUntilEOT("Thorntooth Witch — +3/-3 until end of turn", 3, -3)),
				effects.TargetCreature("target creature")),
				"Thorntooth Witch — have target creature get +3/-3 until end of turn?"),
		},
	})
}
