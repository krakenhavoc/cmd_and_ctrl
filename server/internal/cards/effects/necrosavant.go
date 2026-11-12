package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necrosavant — Creature — Zombie Giant {3}{B}{B}{B}, 5/5 (Multiverse Gift
// Box, #5632):
//
//	"{3}{B}{B}, Sacrifice a creature: Return this card from your graveyard
//	 to the battlefield. Activate only during your upkeep."
//
// Cauldron Familiar's graveyard return (an activation from the graveyard,
// CR 113.6, through Zones) with Hell's Caretaker's window: DuringYourUpkeep
// is the step AND your turn, so an opponent's upkeep is never one. The
// creature sacrificed is one on the battlefield; the Necrosavant itself is
// in the graveyard and can't be. A card that has left the graveyard by
// the time the ability resolves is a new object (CR 400.7) and stays
// where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6732a587-e3cc-4c30-9267-f8949dd8724a",
		Name:         "Necrosavant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{3}{B}{B}, Sacrifice a creature: Return this card from your graveyard to the battlefield. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:      Plus(ManaCost("{3}{B}{B}"), SacrificeACreature()),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: DuringYourUpkeep(),
			Effect:    returnThisFromGraveyardToBattlefield,
		}},
	})
}
