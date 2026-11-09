package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sarkhan, Dragon Ascendant — Legendary Creature — Human Druid {1}{R},
// 2/2 (Tarkir: Dragonstorm, #4463):
//
//	"When Sarkhan enters, you may behold a Dragon. If you do, create a
//	 Treasure token. (To behold a Dragon, choose a Dragon you control or
//	 reveal a Dragon card from your hand.)
//	 Whenever a Dragon you control enters, put a +1/+1 counter on Sarkhan.
//	 Until end of turn, Sarkhan becomes a Dragon in addition to his other
//	 types and gains flying."
//
// The enters trigger asks MayBehold (std_b11_helpers.go): a Dragon you
// control or a Dragon card in hand is the choice, a card from hand is
// revealed, and the Treasure follows the choice, not what happens to the
// Dragon afterwards. With no Dragon to behold nobody is asked. The second
// trigger is Dragon's Hoard's (CR 603.2: it triggers on a Dragon entering
// under your control, Sarkhan himself once he is one included), the
// counter first and the type and flying until end of turn through the
// same effect, bound to the object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71cb48e2-2328-471c-a04a-7b6461f378f4",
		Name:         "Sarkhan, Dragon Ascendant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sarkhan, Dragon Ascendant — you may behold a Dragon; if you do, create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					return sarkhanBehold.Ask(NewContext(g, item))
				}),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				// "Dragon" is a word of the source's text (CR 612.2, #3984).
				return ok && c.HasSubtype(source.TextSubtype("Dragon"))
			}, "Sarkhan, Dragon Ascendant — put a +1/+1 counter on Sarkhan; until end of turn he becomes a Dragon and gains flying",
				sarkhanGrow),
		},
	})
}

var sarkhanBehold = NewMayBehold("Sarkhan, Dragon Ascendant", "Dragon", "choose-cards/sarkhan-dragon-ascendant-behold",
	func(ctx *Context) error {
		return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
	})

func sarkhanGrow(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if err := (BecomeCreatureUntilEOT{
		Target: item.SourceCardID, Types: []string{"Creature"}, Subtypes: []string{"Dragon"},
		Label: "Sarkhan, Dragon Ascendant — Dragon",
	}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"},
		Label: "Sarkhan, Dragon Ascendant — flying"}.Apply(ctx)
}
