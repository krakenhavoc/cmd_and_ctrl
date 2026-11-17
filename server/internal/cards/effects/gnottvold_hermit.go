package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gnottvold Hermit // Chrome Host Hulk — a Phyrexian transformer (March of the Machine):
//
//	Gnottvold Hermit {3}{G}, Creature — Troll, 4/4
//	  "{5}{U/P}: Transform this creature. Activate only as a sorcery."
//
//	Chrome Host Hulk, Creature — Phyrexian Troll, 5/5
//	  "Whenever this creature attacks, up to one other target creature has
//	   base power and toughness 5/5 until end of turn."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with U or 2 life.
//
// No simplification.
func init() {
	const oracle = "3bafc004-1c90-4939-9dcb-72667c170e39"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Gnottvold Hermit",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{5}{U/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{5}{U/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Chrome Host Hulk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Chrome Host Hulk — up to one other target creature has base power and toughness 5/5 until end of turn",
				chromeHostHulkReshape),
				Another(UpToOneTargetCreature("up to one other target creature"))),
		},
	})
}
