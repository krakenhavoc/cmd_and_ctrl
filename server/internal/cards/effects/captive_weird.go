package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captive Weird // Compleated Conjurer — a Phyrexian transformer (March of the Machine):
//
//	Captive Weird {U}, Creature — Weird, 1/3
//	  "Defender
//	   {3}{R/P}: Transform this creature. Activate only as a sorcery."
//
//	Compleated Conjurer, Creature — Phyrexian Weird, 3/3
//	  "When this creature transforms into Compleated Conjurer, exile the
//	   top card of your library. Until the end of your next turn, you may
//	   play that card."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with R or 2 life.
//
// No simplification.
func init() {
	const oracle = "ac5487ba-6d6e-47ee-ae29-776e186789cc"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Captive Weird",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:        "{3}{R/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{R/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Compleated Conjurer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventTransform, ThisTransformedInto(1),
				"Compleated Conjurer — exile the top card of your library; until the end of your next turn, you may play that card",
				b20ExileTopUntilEndOfNextTurn),
		},
	})
}
