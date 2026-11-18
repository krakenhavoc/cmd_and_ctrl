package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anowon, the Ruin Sage — Legendary Creature — Vampire Shaman {3}{B}{B},
// 4/3:
//
//	"At the beginning of your upkeep, each player sacrifices a
//	 non-Vampire creature of their choice."
//
// An edict over the non-Vampire creatures: every player, Anowon's
// controller included, chooses their own and the choices are made in
// APNAP order and sacrificed together (EachPlayerSacrifices; CR 701.21a).
// A player with nothing eligible sacrifices nothing; Anowon, being a
// Vampire, is never eligible.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c201f79-9eca-48c8-88ba-a20fa6918a07",
		Name:         "Anowon, the Ruin Sage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerEdict(AtYourUpkeep("Anowon, the Ruin Sage — each player sacrifices a non-Vampire creature of their choice",
				Do(EachPlayerSacrifices{
					Match: And(Creature(), Not(HasSubtype("Vampire"))),
					Label: "a non-Vampire creature",
				}))),
		},
	})
}
