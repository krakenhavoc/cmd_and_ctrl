package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Last Light of Durin's Day — Enchantment {1}{R} (The Hobbit):
//
//	"Whenever a Mountain you control enters, put a quest counter on this
//	 enchantment. If it has six or more quest counters on it, sacrifice
//	 it. If you do, search your hand and/or library for a Dragon card and
//	 put it onto the battlefield. If you search your library this way,
//	 shuffle.
//	 Mountaincycling {2} ({2}, Discard this card: Search your library for
//	 a Mountain card, reveal it, put it into your hand, then shuffle.)"
//
// Dragon's Hoard's counter-on-a-type trigger, with the count read as the
// trigger resolves (not an intervening if: the printed sentence is a
// resolution check), then Ajani's Last Stand's "sacrifice it, and only if
// you do" (SacrificeThisThen). The search is SearchLibrary{AlsoHand}: your
// hand is searched beside your library in one prompt, "and/or" letting you
// look in either, so a Dragon may be found in hand without touching the
// library, and a take wholly from the hand does not shuffle. A search that
// finds nothing is a library search, as AlsoGraveyard documents. The
// Dragon enters through the ordinary entry pipeline.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "539583f4-21c8-4b61-a2c5-f612fadc3ffa",
		Name:         "Last Light of Durin's Day",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				// "Mountain" is a word of the source's text (CR 612.2).
				return ok && c.HasSubtype(source.TextSubtype("Mountain"))
			}, "Last Light of Durin's Day — put a quest counter on it; with six or more, sacrifice it and search your hand and/or library for a Dragon card",
				lastLightQuest),
		},
		Activated: []ActivatedAbility{
			Typecycling("Mountaincycling", "{2}", "a Mountain card", IsLandWithSubtype("Mountain")),
		},
	})
}

// lastLightSearch is "sacrifice it. If you do, search your hand and/or
// library for a Dragon card and put it onto the battlefield". An on-disk key.
var lastLightSearch = SacrificeThisThen("then/last-light-of-durins-day-search", func(ctx *Context) error {
	// "Dragon" is a word of this card's text (CR 612.2).
	dragon := ctx.TextSubtype("Dragon")
	return SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(c game.Card) bool { return c.HasSubtype(dragon) },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		AlsoHand:  true,
		Reason:    "Last Light of Durin's Day — a Dragon card from your hand or library",
	}.Apply(ctx)
})

// lastLightQuest is the trigger's body.
func lastLightQuest(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}).Apply(ctx); err != nil {
		return err
	}
	self, ok := ctx.SourcePermanent()
	if !ok || self.Counters["quest"] < 6 {
		return nil
	}
	return lastLightSearch(g, item)
}
