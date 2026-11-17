package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blightreaper Thallid // Blightsower Thallid — a Phyrexian transformer (March of the Machine):
//
//	Blightreaper Thallid {1}{B}, Creature — Fungus, 2/2
//	  "{3}{G/P}: Transform this creature. Activate only as a sorcery."
//
//	Blightsower Thallid, Creature — Phyrexian Fungus, 3/3
//	  "When this creature transforms into Blightsower Thallid or dies,
//	   create a 1/1 green Phyrexian Saproling creature token."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with G or 2 life.
//
// No simplification.
func init() {
	const oracle = "ea405cb4-a21b-4a26-98f3-251e65ea2aef"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Blightreaper Thallid",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{3}{G/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{G/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Blightsower Thallid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(OnAny([]game.EventKind{game.EventTransform, game.EventLTB}, blightsowerThallidCondition,
				"Blightsower Thallid — create a 1/1 green Phyrexian Saproling creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Phyrexian Saproling"), N: 1})), game.Purpose{Tokens: 1}),
		},
	})
}
