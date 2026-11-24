package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coastline Marauders — {2}{R} Creature — Human Pirate 0/3, Commander Legends (#5507):
//
//	"Trample
//	 Whenever this creature attacks, it gets +1/+0 until end of turn for each land
//	 defending player controls.
//	 Encore {4}{R}{R}"
//
// Unlike Elturel Survivors' static, this is a trigger: the lands are counted as
// it resolves, for the player the attack was declared against (CR 508.5), and
// the bonus lasts until end of turn. Encore is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c93bef6e-b676-43df-9f0f-026aaff25604",
		Name:            "Coastline Marauders",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{effects.Encore("{4}{R}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self, "Coastline Marauders — +1/+0 until end of turn for each land defending player controls",
				effects.S82OldB082ThisGetsPowerPerLandDefendingPlayerControls("Coastline Marauders — +1/+0 for each land defending player controls")),
		},
	})
}
