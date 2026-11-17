package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aetherblade Agent // Gitaxian Mindstinger — a Phyrexian transformer (March of the Machine):
//
//	Aetherblade Agent {1}{B}, Creature — Human Rogue, 1/1
//	  "Deathtouch
//	   {4}{U/P}: Transform this creature. Activate only as a sorcery."
//
//	Gitaxian Mindstinger, Creature — Phyrexian Rogue, 3/3
//	  "Deathtouch
//	   Whenever this creature deals combat damage to a player or battle,
//	   draw a card."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with U or 2 life.
//
// No simplification.
func init() {
	const oracle = "80ecb069-36e2-490d-9f6c-ef553c00e997"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Aetherblade Agent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []ActivatedAbility{{
			Label:        "{4}{U/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{4}{U/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Gitaxian Mindstinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventDealDamage, thisDealtCombatDamageToAPlayerOrBattle,
				"Gitaxian Mindstinger — draw a card", Do(DrawCards{N: 1})), game.Purpose{Draws: 1}),
		},
	})
}
