package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hag Hedge-Mage — {2}{B/G} Creature — Hag Shaman, 2/2:
//
//	"When this creature enters, if you control two or more Swamps, you may have
//	 target player discard a card.
//	 When this creature enters, if you control two or more Forests, you may put
//	 target card from your graveyard on top of your library."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "364e8bd6-52e5-4397-8d2a-10c4706e5a45",
		Name:         "Hag Hedge-Mage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.InterveningIf(effects.CondControlTwoOrMoreOfSubtype("Swamp"),
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ControlTwoOrMoreOfSubtype("Swamp")),
					"Hag Hedge-Mage — if you control two or more Swamps, you may have target player discard a card",
					effects.TargetPlayerDiscards(1))),
				effects.TargetPlayer("target player")),
				"Hag Hedge-Mage — have target player discard a card?"),
			effects.Optional(effects.Targeting(effects.InterveningIf(effects.CondControlTwoOrMoreOfSubtype("Forest"),
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ControlTwoOrMoreOfSubtype("Forest")),
					"Hag Hedge-Mage — if you control two or more Forests, you may put target card from your graveyard on top of your library",
					effects.PutChosenTargetOnTopOfLibrary)),
				effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn())),
				"Hag Hedge-Mage — put target card from your graveyard on top of your library?"),
		},
	})
}
