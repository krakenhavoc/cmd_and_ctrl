package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyretic Prankster // Glistening Goremonger — a Phyrexian transformer (March of the Machine):
//
//	Pyretic Prankster {1}{R}, Creature — Devil, 2/1
//	  "{3}{B/P}: Transform this creature. Activate only as a sorcery."
//
//	Glistening Goremonger, Creature — Phyrexian Devil, 3/2
//	  "When this creature dies, each opponent sacrifices an artifact or
//	   creature of their choice."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with B or 2 life.
//
// No simplification.
func init() {
	const oracle = "f810e185-4498-483b-ab37-0c583dae0eaf"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Pyretic Prankster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{3}{B/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{B/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Glistening Goremonger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisDies("Glistening Goremonger — each opponent sacrifices an artifact or creature of their choice",
				Do(EachPlayerSacrifices{ExceptController: true, Match: Or(Artifact(), Creature()), Label: "an artifact or creature"}))),
		},
	})
}
