package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Widow, Daring Operative — Legendary Creature — Human Spy Hero {4}{B},
// 3/4 (Marvel Super Heroes Commander, #4455):
//
//	"Menace
//	 When Black Widow enters, mill three cards.
//	 Whenever Black Widow deals combat damage to a player, each opponent loses
//	 life equal to the number of creature cards in your graveyard."
//
// The mill is Doom's Servo-Guards'. The drain counts the graveyard when the
// trigger resolves (CR 608.2h), and each opponent loses that much, the
// damaged player included. Black Widow's own combat damage is the condition
// (Cadira's b36SelfDealtCombatDamageToPlayer).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8815686-8c47-4067-9dfa-744b2f5c1ba2",
		Name:            "Black Widow, Daring Operative",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Black Widow, Daring Operative — mill three cards", effects.Do(effects.MillCards{N: 3})),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B36SelfDealtCombatDamageToPlayer(ev, source, g)
			}, "Black Widow, Daring Operative — each opponent loses life equal to the number of creature cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return effects.EachOpponentLosesLifeNow(g, item, effects.B11CreatureCardsInGraveyard(g, item.Controller))
				}),
		},
	})
}
