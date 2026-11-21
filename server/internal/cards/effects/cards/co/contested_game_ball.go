package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Contested Game Ball — Artifact {2} (The Lost Caverns of Ixalan, #4470):
//
//	"Whenever you're dealt combat damage, the attacking player gains
//	 control of this artifact and untaps it.
//	 {2}, {T}: Draw a card and put a point counter on this artifact. Then
//	 if it has five or more point counters on it, sacrifice it and create a
//	 Treasure token."
//
// The trigger fires once per combat damage event dealt to the controller by
// a creature; the attacking player is the controller of the creature that
// dealt it, as it last existed (CR 608.2h), and the control change lasts
// until the artifact leaves (CR 611.2a), so it can be won back. A Ball
// that left and came back is a new object and is not taken (CR 400.7). The
// activated ability draws, counts and only then looks at the counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "876c9620-216c-45dd-b8a6-b9e0cba63e3d",
		Name:         "Contested Game Ball",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.StrixhavenCreatureDealtCombatDamageToYou,
				"Contested Game Ball — the attacking player gains control of this artifact and untaps it",
				contestedGameBallSteal),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card and put a point counter on this artifact. Then if it has five or more point counters on it, sacrifice it and create a Treasure token.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Effect:  contestedGameBallDraw,
		}},
	})
}

func contestedGameBallSteal(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || effects.SourceIsNewObject(g, item) {
		return nil
	}
	dealer, ok := g.LookupCardForEffect(item.Trigger.Event.Source)
	if !ok {
		return nil
	}
	ball := item.SourceCardID
	if !g.GainControlForEffect(ball, ball, dealer.Controller, game.IndefiniteDuration(),
		"Contested Game Ball — gains control of this artifact") {
		return nil
	}
	return g.UntapTargetForEffect(ball)
}

func contestedGameBallDraw(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	if effects.SourceIsNewObject(g, item) {
		return nil
	}
	if err := (effects.CounterOnThis{Kind: "point", N: 1}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || c.Counters["point"] < 5 {
		return nil
	}
	if err := effects.SacrificeThisIfStillOnBattlefield(g, item); err != nil {
		return err
	}
	return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
}
