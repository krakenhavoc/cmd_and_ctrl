package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Padishah Emperor Shaddam IV — Legendary Creature — Human Noble {5}{B}{B},
// 2/4 (Secret Lair Drop):
//
//	"When Padishah Emperor Shaddam IV enters, you become the monarch.
//	 Whenever another nontoken creature dies while you're the monarch, you may
//	 pay 1 life. If you do, create a Treasure token.
//	 At the beginning of your end step, destroy target creature with the
//	 greatest power among creatures your opponents control."
//
// "While you're the monarch" is read as the creature dies. The life is paid
// through the cost path (PayLifeForEffect), so a player with no life to pay
// makes no Treasure. The end-step target is a creature an opponent controls
// whose power is the greatest among them: every creature tied for it is a
// legal target.
func init() {
	dies := effects.Optional(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if g.Monarch != source.Controller || ev.CardID == source.InstanceID {
			return false
		}
		c, ok := effects.DiedCreature(ev, g)
		return ok && !c.IsToken()
	}, "Padishah Emperor Shaddam IV — you may pay 1 life; if you do, create a Treasure token",
		func(g *game.Game, item *game.StackItem) error {
			if err := g.PayLifeForEffect(item.SourceCardID, item.Controller, 1); err != nil {
				return nil
			}
			return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
		}), "Padishah Emperor Shaddam IV — pay 1 life to create a Treasure token?")
	endStep := effects.Targeting(effects.AtYourEndStep("Padishah Emperor Shaddam IV — destroy target creature with the greatest power among creatures your opponents control", effects.DestroyFirstLegalTarget),
		effects.TargetCreature("target creature with the greatest power among creatures your opponents control", effects.OldB11GreatestPowerAmongOpponentsCreatures))
	effects.Register(effects.Spec{
		OracleID:     "9726cf0e-5fd2-4699-8192-2fafa2b36963",
		Name:         "Padishah Emperor Shaddam IV",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Padishah Emperor Shaddam IV — you become the monarch", effects.Do(effects.BecomeTheMonarch{})),
			dies,
			endStep,
		},
	})
}
