package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beluna Grandsquall // Seek Thrills — Legendary Creature — Giant Noble
// {G}{U}{R}, 4/4 // Instant — Adventure {2}{G}{U}{R} (Wilds of Eldraine,
// #4471):
//
//	Beluna Grandsquall — "Trample
//	 Permanent spells you cast that have an Adventure cost {1} less to
//	 cast."
//	Seek Thrills — "Mill seven cards. Then put all cards that have an
//	 Adventure from among the milled cards into your hand."
//
// The discount is one cost modifier (CR 601.2f) over the card's permanent
// half, so the Adventure half of the same card is not discounted (it is
// an instant or sorcery spell), nor is anything without an Adventure. Seek
// Thrills is Fetch Quest's mill and keyed tail, taking every adventurer
// card that really landed in the graveyard and is still there.
//
// No simplification.
const belunaOracleID = "06374e6d-62fa-4833-81cd-b4d0a35e7365"

func init() {
	Register(Spec{
		OracleID:        belunaOracleID,
		Name:            "Beluna Grandsquall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Permanent spells you cast that have an Adventure cost {1} less to cast.",
				YourSpell(), PermanentSpellWithAnAdventure()),
		},
	})
	Register(Spec{
		OracleID:     belunaOracleID + "#1",
		Name:         "Seek Thrills",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return MillToZone{N: 7, ThenTail: ctx.ItemTail(seekThrillsMilled)}.Apply(ctx)
		},
	})
}

// seekThrillsMilled puts every card that has an Adventure among the cards
// milled this way, and still in the graveyard, into its owner's hand.
var seekThrillsMilled = CardTail("seek-thrills/after-mill", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && cardHasAdventure(c) {
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
})
