package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Xyris, the Writhing Storm — Legendary Creature — Snake Leviathan
// {2}{G}{U}{R}, 3/5:
//
//	"Flying
//	 Whenever an opponent draws a card except the first one they draw in
//	 each of their draw steps, create a 1/1 green Snake creature token.
//	 Whenever Xyris deals combat damage to a player, you and that player
//	 each draw that many cards."
//
// The first trigger is Orcish Bowmasters' draw condition. The second reads
// the damage and the damaged player off the triggering event: you draw
// that many, then the damaged player does (a player that has left draws
// nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8687948c-456b-495b-9b31-f818c65624b6",
		Name:            "Xyris, the Writhing Storm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, effects.RprB19OpponentDrewNotFirstInDrawStep,
				"Xyris, the Writhing Storm — create a 1/1 green Snake creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Snake"), N: 1})),
			effects.WheneverThisDealsCombatDamageToAPlayer("Xyris, the Writhing Storm — you and that player each draw that many cards",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					n, victim := item.Trigger.Event.Amount, item.Trigger.Event.Target
					ctx := effects.NewContext(g, item)
					if err := (effects.DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
						return err
					}
					if g.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					return effects.DrawCards{Player: victim, N: n}.Apply(ctx)
				}),
		},
	})
}
