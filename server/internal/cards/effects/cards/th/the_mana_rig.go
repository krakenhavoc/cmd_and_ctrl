package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Mana Rig — Legendary Artifact {3} (Dominaria United, S82 DMU batch 15):
//
//	"Whenever you cast a multicolored spell, create a tapped Powerstone
//	 token. (It's an artifact with "{T}: Add {C}. This mana can't be spent
//	 to cast a nonartifact spell.")
//	 {X}{X}{X}, {T}: Look at the top X cards of your library. Put up to two
//	 of them into your hand and the rest on the bottom of your library in a
//	 random order."
//
// The cast trigger is the Powerstone of every other Powerstone card, entering
// tapped. {X}{X}{X} is three X slots (Treasure Vault's two): X is announced
// once and the cost is three times it. X is read as the ability resolves; a
// library shorter than X is looked at whole.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58678469-77e5-4c1b-88d6-5e6642b0c8ca",
		Name:         "The Mana Rig",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Multicolored(),
				"The Mana Rig — create a tapped Powerstone token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{X}{X}, {T}: Look at the top X cards of your library. Put up to two of them into your hand and the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{X}{X}{X}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				return effects.TakeFromLibraryToHand{
					Player:   item.Controller,
					Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, x),
					Max:      2,
					Optional: true,
					Label:    "The Mana Rig — put up to two of them into your hand",
					Key:      effects.TakeRestBottomRandom,
				}.Apply(ctx)
			},
		}},
	})
}
