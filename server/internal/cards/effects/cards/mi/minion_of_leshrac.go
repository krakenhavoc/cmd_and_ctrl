package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minion of Leshrac — Creature — Demon Minion {4}{B}{B}{B}, 5/5 (Ice Age, S82 OLD, #5636):
//
//	"Protection from black
//	 At the beginning of your upkeep, this creature deals 5 damage to you unless you sacrifice a
//	 creature other than this creature. If this creature deals damage to you this way, tap it.
//	 {T}: Destroy target creature or land."
//
// The upkeep is a sacrifice-or-else prompt (the sacrifice can't be the Minion itself); on a "no" the
// Minion deals 5 to you and, if damage was dealt, is tapped. The {T} ability destroys any creature or
// land. Protection from black is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8af3b981-6eda-46b0-a7b8-226bc719d137",
		Name:            "Minion of Leshrac",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Minion of Leshrac — 5 damage to you unless you sacrifice another creature", effects.S82OldB842MinionOfLeshracUpkeep),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target creature or land.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target creature or land", effects.Or(effects.Creature(), effects.Land())),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
