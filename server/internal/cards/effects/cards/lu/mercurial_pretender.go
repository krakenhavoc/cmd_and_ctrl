package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mercurial Pretender — Creature — Shapeshifter {4}{U}, 0/0 (Magic 2015, #5556):
//
//	"You may have this creature enter as a copy of a creature you control,
//	 except it has "{2}{U}{U}: Return this creature to its owner's hand.""
//
// Evil Twin's enters-as-a-copy replacement (EntersAsCopyOf, CR 614.12 and
// 707.9) over the creatures its controller controls, with a granted activated
// ability that is part of the copiable values (CR 707.9a), so a Clone of the
// Pretender has it too. The ability is Brightling's return of its own source.
// Declined, it is a 0/0 and dies (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07d0bd24-7731-4abb-9dc6-912aa17943f2",
		Name:         "Mercurial Pretender",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  mercurialPretenderGrant,
			Text: "{2}{U}{U}: Return this creature to its owner's hand.",
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}{U}{U}: Return this creature to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{2}{U}{U}"),
				Effect:  effects.ReturnThisPermanentToOwnersHand,
			}},
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Mercurial Pretender", effects.ACreatureYouControl,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.GrantAbility(mercurialPretenderGrant)
				}),
		},
	})
}

// mercurialPretenderGrant names the granted ability's bundle: an on-disk
// identity.
const mercurialPretenderGrant = "mercurial-pretender/return-to-hand"
