package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e8815686-8c47-4067-9dfa-744b2f5c1ba2",
		Name:            "Black Widow, Daring Operative",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Black Widow, Daring Operative — mill three cards", Do(MillCards{N: 3})),
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b36SelfDealtCombatDamageToPlayer(ev, source, g)
			}, "Black Widow, Daring Operative — each opponent loses life equal to the number of creature cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return eachOpponentLosesLife(g, item, b11CreatureCardsInGraveyard(g, item.Controller))
				}),
		},
	})
}
