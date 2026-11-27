package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Manor Gargoyle — {5} Artifact Creature — Gargoyle 4/4 (Innistrad, S82 OLD
// b446, #5572):
//
//	"Defender
//	 This creature has indestructible as long as it has defender.
//	 {1}: Until end of turn, this creature loses defender and gains flying."
//
// The indestructible grant reads defender, which the ability removes in the
// same layer, so it is a keyword-reading static (CR 613.8): it waits for the
// removal whatever the timestamps, and the Gargoyle that has lost defender is
// no longer indestructible. The ability is Gargoyle Sentinel's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "edb6c7ea-b4e1-4cb7-b5a9-8c5644a2aad9",
		Name:            "Manor Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{
			effects.WithKeywordHave(effects.SelfOnly, "defender", "This creature has indestructible as long as it has defender.", "indestructible"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Until end of turn, this creature loses defender and gains flying.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntilEndOfTurn(effects.NewContext(g, item), item.SourceCardID, nil,
					"Manor Gargoyle — loses defender and gains flying until end of turn",
					game.RemoveKeywordsMod("defender"), game.AddKeywordsMod("flying"))
			},
		}},
	})
}
