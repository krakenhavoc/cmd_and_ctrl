package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathknell Kami — Creature — Spirit {1}{B}, 0/1 (Saviors of Kamigawa,
// #5601):
//
//	"Flying
//	 {2}: This creature gets +1/+1 until end of turn. Sacrifice it at the
//	 beginning of the next end step.
//	 Soulshift 1 (When this creature dies, you may return target Spirit card
//	 with mana value 1 or less from your graveyard to your hand.)"
//
// Each activation pumps once and schedules its own delayed sacrifice
// (CR 603.7) at the next end step of any player's turn; it takes only this
// same creature (CR 400.7). Several activations schedule several triggers, and
// the later ones find nothing left to sacrifice.
//
// No simplification.
func init() {
	const label = "Deathknell Kami — sacrifice it at the beginning of the next end step"
	effects.Register(effects.Spec{
		OracleID:        "4d324f8c-2139-4910-b3c3-731a2e820d6b",
		Name:            "Deathknell Kami",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+1 until end of turn. Sacrifice it at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ThisGetsUntilEndOfTurn(1, 1, "Deathknell Kami — +1/+1 until end of turn")(g, item); err != nil {
					return err
				}
				return effects.OldB507SacrificeThisAtTheNextEndStep(label)(g, item)
			},
		}},
		Triggered: []game.TriggeredAbility{effects.Soulshift(1, "Deathknell Kami")},
	})
}
