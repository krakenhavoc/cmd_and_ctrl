package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Molderhulk — Creature — Fungus Zombie {7}{B}{G}, 6/6 (Guilds of Ravnica,
// #5525):
//
//	"Undergrowth — This spell costs {1} less to cast for each creature card
//	 in your graveyard.
//	 When this creature enters, return target land card from your graveyard
//	 to the battlefield."
//
// The discount is a cost-reduction effect on the spell (CR 601.2f), counted
// as it is cast; it never reduces the coloured part. The return is not a
// "may": with a land card in your graveyard it must be targeted and returned
// (to the battlefield untapped, under your control).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6fa840a-817c-4e84-9dd3-d28a011f2214",
		Name:         "Molderhulk",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.CreatureCardsInYourGraveyard,
				"Undergrowth — This spell costs {1} less to cast for each creature card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Molderhulk — return target land card from your graveyard to the battlefield",
					effects.ReturnTargetedGraveyardCardToBattlefield),
				effects.TargetCardInGraveyard("target land card from your graveyard", effects.Land(), effects.YouOwn())),
		},
	})
}
