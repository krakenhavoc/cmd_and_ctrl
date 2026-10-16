package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kodama of the East Tree — Legendary Creature — Spirit {4}{G}{G}, 6/6
// (#3933):
//
//	"Reach
//	 Whenever another permanent you control enters, if it wasn't put onto
//	 the battlefield with this ability, you may put a permanent card with
//	 equal or lesser mana value from your hand onto the battlefield.
//	 Partner"
//
// The intervening "if" (CR 603.4) is TriggeredAbility.NotPutByThisAbility:
// every entry's EventETB names the ability that put the permanent there
// (its source object and catalog row, game/entry_provenance.go), and the
// entry this Kodama's own ability made does not trigger it again, so one
// Kodama puts at most one card per arrival. A land played, a spell
// resolving, a token made and a card another Kodama put all trigger it:
// none was put with THIS ability (two Kodamas feed each other, as printed,
// until the hands run out of cheap enough cards). The cause is fixed with
// the event, so the check as it resolves agrees with the one as it
// triggered.
//
// "Equal or lesser mana value" is the entered permanent's, read as it
// resolves (CR 608.2h): as it is while on the battlefield, as it last
// existed there once gone. The "you may" is the put's own optional pick.
// Partner is a deck-construction rule internal/deck reads off the oracle
// text (ADR 0144).
//
// No simplification.
func init() {
	const label = "Kodama of the East Tree — you may put a permanent card with equal or lesser mana value from your hand onto the battlefield"
	t := On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		_, ok := enteredUnderYourControl(ev, source, g, true)
		return ok
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		entered, ok := ctx.TriggeringPermanent()
		if !ok {
			return nil
		}
		return PutFromHandOntoBattlefield{
			Match:    ManaValueLE(entered.ManaValue),
			Optional: true,
			Label:    label,
		}.Apply(ctx)
	})
	t.NotPutByThisAbility = true
	Register(Spec{
		OracleID:        "cb1afab6-b9c8-4b9c-840b-24106523a783",
		Name:            "Kodama of the East Tree",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
