package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barret, Avalanche Leader — Legendary Creature — Human Rebel
// {2}{R}{G}, 4/4 (#3656):
//
//	"Reach
//	 Avalanche! — Whenever an Equipment you control enters, create a
//	 2/2 red Rebel creature token.
//	 At the beginning of combat on your turn, attach up to one target
//	 Equipment you control to target Rebel you control."
//
// The token is made for every Equipment you control that enters,
// including one that arrives while Barret is entering (he sees them
// enter). The combat attach is Brass Squire's two-slot shape with an
// "up to one" Equipment clause: with no Equipment chosen there is
// nothing to attach, and the Rebel is still the second target as
// printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "83f776d9-f3b6-40c2-8008-1ace4d110825",
		Name:            "Barret, Avalanche Leader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype("Equipment")
			}, "Barret, Avalanche Leader — Avalanche!: create a 2/2 red Rebel creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red Rebel"), N: 1})),
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Barret, Avalanche Leader — attach up to one target Equipment you control to target Rebel you control",
					effects.AttachClauseTargets),
				effects.TwoSlotAttachTargets(
					"up to one target Equipment you control", []effects.CardPredicate{effects.YouControl()},
					"target Rebel you control", []effects.CardPredicate{effects.YouControl(), effects.HasSubtype("Rebel")},
				).WithCount(0, 1)),
		},
	})
}
