package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feasting Troll King — Creature — Troll Noble {2}{G}{G}{G}{G}, 7/6
// (Throne of Eldraine, S82 old sets):
//
//	"Vigilance, trample
//	 When this creature enters, if you cast it from your hand, create three
//	 Food tokens.
//	 Sacrifice three Foods: Return this card from your graveyard to the
//	 battlefield. Activate only during your turn."
//
// The enters trigger is Breaching Leviathan's intervening "if" (CR 603.4)
// off the cast record, checked as it enters and as it resolves. The third
// ability works from the graveyard (CR 113.6) with a cost of three Foods
// (any Food permanents, tokens or not), at any time of your own turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6f0c6876-dcae-46e5-8c62-6929733c3525",
		Name:            "Feasting Troll King",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{Name: "you cast it from your hand", Test: func(ctx *effects.Context) bool {
				return effects.B30CastFromHand(ctx.Game, ctx.Item.SourceCardID)
			}}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && effects.B30CastFromHand(g, source.InstanceID)
			}, "Feasting Troll King — create three Food tokens",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 3}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "Sacrifice three Foods: Return this card from your graveyard to the battlefield. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.SacrificeN(3, "three Foods", effects.HasSubtype("Food")),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: effects.DuringYourTurn(),
			Effect:    effects.ReturnThisFromGraveyardToBattlefield,
		}},
	})
}
