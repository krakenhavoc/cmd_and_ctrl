package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyclave Shade — Creature — Shade {1}{B}, 3/1 (Zendikar Rising):
//
//	"Kicker {2}{B} (You may pay an additional {2}{B} as you cast this
//	 spell.)
//	 This creature can't block.
//	 If this creature was kicked, it enters with two +1/+1 counters on it.
//	 Landfall — Whenever a land you control enters, if this card is in
//	 your graveyard and it's your turn, you may cast it from your
//	 graveyard this turn."
//
// The kicker, the entry counters and "can't block" are Gnarlid Colony's and
// Aesthir Glider's. The landfall ability works from the graveyard
// (TriggeredAbility.Zones, CR 113.6b), where "you" is the card's owner. Its
// intervening "if" is the card being in the graveyard and the turn being
// yours (CR 603.4): the turn half is read as the land enters, the graveyard
// half again as the ability resolves, as the same object (CR 400.7). The
// grant is a cast-only permission on this card for the turn, its mana cost
// paid as for any cast (kicker included); a second land only grants it again.
//
// No simplification.
func init() {
	grant := func(g *game.Game, item *game.StackItem) error {
		if !cardObjectStillInGraveyard(g, item) || !IsYourTurn(g, item.Controller) {
			return nil
		}
		return CastTargetFromYourGraveyardThisTurn{
			Target: item.SourceCardID,
			Label:  "Skyclave Shade — cast from your graveyard this turn",
		}.Apply(NewContext(g, item))
	}
	Register(Spec{
		OracleID:                   "2345b243-ed0d-4556-8017-d533db1fbc74",
		Name:                       "Skyclave Shade",
		Completeness:               CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{Kicker("{2}{B}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 2)},
		Static:                     []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{InGraveyard(On(game.EventETB,
			func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return LandEnteredUnderYourControl(ev, source, ch, g) && IsYourTurn(g, source.Controller)
			},
			"Skyclave Shade — landfall: you may cast it from your graveyard this turn", grant))},
	})
}
