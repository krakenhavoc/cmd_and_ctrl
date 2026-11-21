package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lictor — Creature — Tyranid {3}{G}, 3/3:
//
//	"Flash
//	 Pheromone Trail — When this creature enters, if a creature entered the
//	 battlefield under an opponent's control this turn, create a 3/3 green
//	 Tyranid Warrior creature token with trample."
//
// An intervening "if" (CR 603.4): checked as the trigger would go on the stack
// and again as it resolves, from the per-turn entry tally the engine records
// as each permanent enters (Game.EnteredWithCardTypeThisTurn), so a creature
// that has since died still counts.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3fdf064d-a5be-4925-87b1-3547997f0840",
		Name:            "Lictor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{lictorTrigger()},
	})
}

func lictorOpponentCreatureEntered(g *game.Game, you uuid.UUID) bool {
	for _, seat := range g.Seats {
		if seat != nil && seat.ID != you && g.EnteredWithCardTypeThisTurn(seat.ID, "creature") > 0 {
			return true
		}
	}
	return false
}

func lictorTrigger() game.TriggeredAbility {
	t := effects.InterveningIf(effects.CondBoard{
		Name: "a creature entered the battlefield under an opponent's control this turn",
		Test: func(ctx *effects.Context) bool { return lictorOpponentCreatureEntered(ctx.Game, ctx.Controller()) },
	}, effects.TriggerWithPurpose(effects.WhenThisEnters("Lictor — create a 3/3 green Tyranid Warrior creature token with trample",
		effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Tyranid Warrior with trample"), N: 1})),
		game.Purpose{Tokens: 1, CreatureTokens: 1}))
	enters := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return enters(ev, source, lki, g) && lictorOpponentCreatureEntered(g, source.Controller)
	}
	return t
}
