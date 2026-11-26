package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Duskmantle Guildmage — Creature — Human Wizard {U}{B}, 2/2 (Gatecrash, S82
// OLD b408, #5565):
//
//	"{1}{U}{B}: Whenever a card is put into an opponent's graveyard from
//	 anywhere this turn, that player loses 1 life.
//	 {2}{U}{B}: Target player mills two cards."
//
// The first ability creates a delayed triggered ability that lasts until end
// of turn and fires for every card that reaches an opponent's graveyard
// (CR 603.7b), by any route (a death, a discard, a mill, a countered spell).
// Activating it twice makes two triggers, so each card costs 2. The loser is
// the owner of the graveyard the card landed in. The second is a plain mill.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec44ca05-6b10-4be2-b4df-967535967fce",
		Name:         "Duskmantle Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{U}{B}: Whenever a card is put into an opponent's graveyard from anywhere this turn, that player loses 1 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{1}{U}{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DelayedOnEvent{
						Label:     "Duskmantle Guildmage — that player loses 1 life",
						On:        []game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill, game.EventCounterSpell},
						Condition: duskmantleGuildmageCondition,
						Body:      duskmantleGuildmageBody,
						Repeats:   true,
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{2}{U}{B}: Target player mills two cards.",
				Cost:    effects.ManaCost("{2}{U}{B}"),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					victim := effects.TargetedPlayer(ctx)
					if victim == uuid.Nil {
						return nil
					}
					return effects.MillCards{Player: victim, N: 2}.Apply(ctx)
				},
			},
		},
	})
}

// duskmantleGuildmageCondition is "a card is put into an opponent's graveyard
// from anywhere": the event moved a card into the graveyard of a player other
// than the delayed trigger's controller. An on-disk key.
var duskmantleGuildmageCondition = game.DelayedCondition("graveyard/card-into-an-opponents",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.CardID == uuid.Nil {
			return false
		}
		switch ev.Kind {
		case game.EventZoneMove, game.EventDiscardCard, game.EventMill:
			if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard {
				return false
			}
		case game.EventCounterSpell:
			if ev.Target != ev.CardID {
				return false
			}
			if z := g.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneGraveyard {
				return false
			}
		default:
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && !c.IsToken() && c.Owner != dt.Controller
	})

// duskmantleGuildmageBody makes the owner of the card the event named lose
// 1 life. An on-disk key.
var duskmantleGuildmageBody = game.SimpleDelayedBody("graveyard/the-cards-owner-loses-1-life",
	func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, id := range ctx.PayloadCards() {
			if c, ok := g.LookupCardForEffect(id); ok {
				return g.ChangePlayerLifeForEffect(ctx.Source(), c.Owner, -1)
			}
		}
		return nil
	})
