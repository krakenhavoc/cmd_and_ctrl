package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gruul Scrapper — Creature — Human Berserker {3}{G}, 3/2 (EDHREC
// rank 31304):
//
//	"When this creature enters, if {R} was spent to cast it, it gains
//	 haste until end of turn."
//
// Ravnica's hybrid-adjacent "if {C} was spent" family, which is the
// largest reader of the mana-spent record in the corpus — thirty-odd
// cards from Azorius Herald's "sacrifice it unless {U} was spent" to
// the Mythos cycle's two-colour version. The colour half has been
// answerable since #761 (`PaidCost.SpentOfColor`); what #1212 adds is
// that a PERMANENT can ask it, because the spell has finished
// resolving before this trigger does.
//
// Note where the {R} comes from: the card costs {3}{G} and the clause
// asks about a colour its cost never requires. A Gruul player pays the
// generic half with a Mountain and gets haste; one who pays it with
// two Forests and a Sol Ring gets a 3/2 that has to wait. That is the
// printed behaviour for a payment made by hand. An AUTO-paid cast gets
// the {R} for free since #3853: WantsColors names the colour, and the
// auto-tapper pays the generic half with a source that makes it when it
// has a free choice (an ordering, never a filter — a board with no red
// source still casts it, as an ordinary 3/2). WantsManaFrom stays unset:
// that wish names a mana SOURCE kind (Treasure, creature), not a colour.
//
// CR 603.4: the clause is an intervening if, checked in AppliesTo, so
// an unhasty Scrapper puts nothing on the stack.
//
// "It gains haste" is the permanent itself, so the grant is aimed at
// the trigger's own source. A Scrapper that has already left the
// battlefield grants nothing, which GrantKeywordUntilEOT's snapshot
// handles.
func init() {
	Register(Spec{
		OracleID:     "878e8980-a2cc-4335-aafe-f5166ef48f79",
		Name:         "Gruul Scrapper",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"When “Charge mana costs” is off, or you choose Cast anyway, only mana spent from your mana pool counts, so Gruul Scrapper gains haste only when the {R} came from your mana pool."},
		WantsColors:  []string{"R"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.ManaSpentToCast().Count("R") > 0
			}), "Gruul Scrapper — gains haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"haste"},
						Label:    "Gruul Scrapper",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
