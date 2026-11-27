package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warren Pilferers — Creature — Goblin Rogue {4}{B}, 3/3 (Lorwyn, S82 OLD,
// #5593):
//
//	"When this creature enters, return target creature card from your graveyard
//	 to your hand. If that card is a Goblin card, this creature gains haste until
//	 end of turn."
//
// Barrow Witches' mandatory regrowth, with the haste read off the card that was
// returned (its subtypes in the graveyard, a changeling included).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f90dcd1-e372-40c5-9de2-81392b0add06",
		Name:         "Warren Pilferers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Warren Pilferers — return target creature card from your graveyard to your hand; haste if it is a Goblin card",
					effects.S82OldB555ReturnTargetCreatureHasteIfGoblin),
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.Creature(), effects.YouOwn())),
		},
	})
}
