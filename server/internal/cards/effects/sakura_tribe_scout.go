package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sakura-Tribe Scout — Creature — Snake Shaman Scout {G}, 1/1 (Saviors of
// Kamigawa, #5601):
//
//	"{T}: You may put a land card from your hand onto the battlefield."
//
// PuPu UFO's row: the land card is picked from the hand as the ability
// resolves, and putting it onto the battlefield is not playing a land, so
// it does not use the turn's land drop. The {T} cost keeps it from being
// activated while the Scout is summoning sick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c319f45-cb0f-4641-acd8-145768acb3fc",
		Name:         "Sakura-Tribe Scout",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: You may put a land card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  Do(MayPutALandFromHand("Sakura-Tribe Scout")),
		}},
	})
}
