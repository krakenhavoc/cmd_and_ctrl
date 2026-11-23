package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Circle of the Land Druid — {1}{G} Creature — Gnome Druid 1/1
//
//	"When this creature enters, you may mill four cards.
//	 Natural Recovery — When this creature dies, return target land card from
//	 your graveyard to your hand."
//
// The mill is optional as the trigger goes on the stack. The death trigger's
// target is chosen then and re-checked as it resolves (CR 608.2b); it can name
// a land the enters trigger milled, which is in the graveyard by then.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6cc351e-6afd-4427-998d-e06e430cd0e6",
		Name:         "Circle of the Land Druid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEnters("Circle of the Land Druid — you may mill four cards", effects.Do(effects.MillCards{N: 4})),
				"Circle of the Land Druid — mill four cards?"),
			effects.Targeting(
				effects.WhenThisDies("Natural Recovery — return target land card from your graveyard to your hand",
					effects.S82Mom16ReturnTargetToHand),
				effects.TargetCardInGraveyard("target land card from your graveyard", effects.Land(), effects.YouOwn())),
		},
	})
}
