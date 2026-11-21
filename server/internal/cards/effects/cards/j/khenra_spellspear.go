package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Khenra Spellspear // Gitaxian Spellstalker — a Phyrexian transformer (March of the Machine):
//
//	Khenra Spellspear {1}{R}, Creature — Jackal Warrior, 2/2
//	  "Trample
//	   Prowess
//	   {3}{U/P}: Transform this creature. Activate only as a sorcery."
//
//	Gitaxian Spellstalker, Creature — Phyrexian Jackal, 3/3
//	  "Trample, ward {2}, prowess, prowess"
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with U or 2 life.
//
// No simplification.
func init() {
	const oracle = "f08d8fc5-c7fb-4913-92fc-06a0d4311d71"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Khenra Spellspear",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordProwess},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{U/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{U/P}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Gitaxian Spellstalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordProwess, game.KeywordProwess},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Gitaxian Spellstalker — ward {2}"),
		},
	})
}
