package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keeper of the Dead — Creature — Human Wizard {B}{B}, 1/2 (Exodus, S82 OLD batch 744, #5626):
//
//	"{B}, {T}: Choose target opponent who has at least two fewer creature
//	 cards in their graveyard than you do as you activate this ability.
//	 Destroy target nonblack creature that player controls."
//
// Two target clauses: the opponent, who is only a condition (the graveyards are
// judged as the ability is activated, CR 602.2b), and a nonblack creature that
// player controls (ControlledByTargetsController, Breaking of the Fellowship's
// relation). As it resolves each target is re-checked (CR 608.2b): the creature
// is destroyed if it is still a legal target, whatever has happened to the
// graveyards since. Destruction can be regenerated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50ddd513-e06e-4728-b6cd-981839d0fba6",
		Name:         "Keeper of the Dead",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{B}, {T}: Choose target opponent who has at least two fewer creature cards in their graveyard than you do as you activate this ability. " +
				"Destroy target nonblack creature that player controls.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
			Targets: effects.Clauses(
				effects.S82OldB744TargetOpponentWithTwoFewerCreatureCards(),
				effects.ControlledByTargetsController(effects.TargetCreature("target nonblack creature that player controls", effects.NonBlack()), 0),
			),
			Effect: effects.S82OldB744DestroyTheTargetCreature,
		}},
	})
}
