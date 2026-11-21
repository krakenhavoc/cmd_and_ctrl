package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archaeomancer's Map — Artifact {2}{W} (EDHREC top 1,000, #3654):
//
//	"When this artifact enters, search your library for up to two basic
//	 Plains cards, reveal them, put them into your hand, then shuffle.
//	 Whenever a land an opponent controls enters, if that player
//	 controls more lands than you, you may put a land card from your
//	 hand onto the battlefield."
//
// The search is a Plains-only fetch (up to two, to hand, revealed). The
// catch-up half is an intervening-if (CR 603.4): the land that entered
// belongs to an opponent who now controls strictly more lands than you.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef833546-f755-4ac5-867a-3926984e68a0",
		Name:         "Archaeomancer's Map",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Archaeomancer's Map — search for up to two basic Plains", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLandWithSubtype("plains"),
					Dest:      game.ZoneHand,
					Limit:     2,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Archaeomancer's Map — up to two basic Plains cards",
				}.Apply(effects.NewContext(g, item))
			}),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source == nil {
					return false
				}
				land, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !land.IsLand() || land.Controller == source.Controller {
					return false
				}
				theirs, mine := 0, 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if !c.IsLand() {
						continue
					}
					switch c.Controller {
					case land.Controller:
						theirs++
					case source.Controller:
						mine++
					}
				}
				return theirs > mine
			}, "Archaeomancer's Map — you may put a land card from your hand onto the battlefield",
				effects.Do(effects.MayPutALandFromHand("Archaeomancer's Map"))),
		},
	})
}
