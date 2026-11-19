package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dimir Strandcatcher — Creature — Faerie Rogue {2}{U/B}{U/B}, 3/3
// (Ravnica: Clue Edition):
//
//	"Flying
//	 Whenever you attack, surveil X, where X is the number of opponents
//	 being attacked.
//	 At the beginning of each end step, if three or more cards were put
//	 into your graveyard from anywhere other than the battlefield this
//	 turn, draw a card."
//
// The attack trigger is Apothecary White's once-per-declaration shape; X
// is Galvanizing Tenor's count of opponents (an attack on a planeswalker
// or battle attacks no player, CR 506.3), read as the ability resolves
// with the declaration's attackers recorded. No opponent attacked means
// no surveil. The end-step trigger fires on every player's end step and is
// an intervening if (CR 603.4), checked as the step begins and again on
// resolution; the count reads the turn's graveyard arrivals
// (s82_bns_b11_helpers.go), so a card that has left the graveyard still
// counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4689206f-7907-46fb-82a7-2716687a58fc",
		Name:            "Dimir Strandcatcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Dimir Strandcatcher — surveil X, where X is the number of opponents being attacked", func(g *game.Game, item *game.StackItem) error {
				x := trekB37OpponentsAttackedThisTurn(g, item.Controller)
				if x <= 0 {
					return nil
				}
				return Surveil{Player: item.Controller, N: x}.Apply(NewContext(g, item))
			})),
			InterveningIf(CondThreeCardsToYourGraveyardNotFromBattlefield{},
				On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return cardsToYourGraveyardNotFromBattlefield(g, source.Controller) >= 3
				}, "Dimir Strandcatcher — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
