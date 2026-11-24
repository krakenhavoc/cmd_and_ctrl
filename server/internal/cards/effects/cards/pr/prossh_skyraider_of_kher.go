package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prossh, Skyraider of Kher — Legendary Creature — Dragon {3}{B}{R}{G}, 5/5
// (Secret Lair Drop):
//
//	"When you cast this spell, create X 0/1 red Kobold creature tokens named
//	 Kobolds of Kher Keep, where X is the amount of mana spent to cast it.
//	 Flying
//	 Sacrifice another creature: Prossh gets +1/+0 until end of turn."
//
// The cast trigger goes on the stack above the spell and resolves even if the
// spell is countered (WhenYouCastThisSpell). X is the total mana the payment
// of the spell was made of, read off the spell's own record.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "868882d2-ed4e-4171-a17c-478a341080fb",
		Name:            "Prossh, Skyraider of Kher",
		ReadsManaSpent:  game.ReadsOwnPayment,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("Prossh, Skyraider of Kher — create X 0/1 red Kobold tokens named Kobolds of Kher Keep, X the mana spent to cast it",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("0/1 red Kobolds of Kher Keep"),
					Name:     "the mana spent to cast it",
					Count:    func(ctx *effects.Context) int { return effects.OldB11ManaSpentOnCastTrigger(ctx) },
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature: Prossh gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
			Cost:    effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Prossh, Skyraider of Kher — +1/+0"),
		}},
	})
}
