package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wilfred Mott — Creature — Human Soldier {3}{W}, 2/4 (Doctor Who, #5948):
//
//	"Look to the Stars — At the beginning of your upkeep, put a time counter on
//	 Wilfred Mott. Then look at the top X cards of your library, where X is the
//	 number of time counters on Wilfred Mott. You may put a nonland permanent
//	 card with mana value 3 or less from among them onto the battlefield. Put
//	 the rest on the bottom of your library in a random order."
//
// One upkeep trigger: the counter goes on Wilfred, then X is read from the
// counters on him now (if he has left, X is 0 and nothing is looked at), the
// top X cards are looked at, one nonland permanent card with mana value 3 or
// less may be put onto the battlefield, and the rest go to the bottom in a
// random order (Armored Skyhunter's pile). "Look to the Stars" is an ability
// word (CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a970742-687f-49ed-8822-e2953f4530b6",
		Name:         "Wilfred Mott",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Wilfred Mott — put a time counter on Wilfred Mott, then look at the top X cards of your library; you may put a nonland permanent card with mana value 3 or less onto the battlefield; put the rest on the bottom in a random order",
				wilfredMottUpkeep),
		},
	})
}

func wilfredMottUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if onBattlefield(g, item.SourceCardID) {
		if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	x := 0
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && onBattlefield(g, item.SourceCardID) {
		x = c.Counters[game.CounterTime]
	}
	if x <= 0 {
		return nil
	}
	return PutFromLibraryOntoBattlefield{
		Player:   item.Controller,
		Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, x),
		Match:    And(Permanent(), Nonland(), ManaValueLE(3)),
		Max:      1,
		Optional: true,
		Label:    "Wilfred Mott — you may put a nonland permanent card with mana value 3 or less from among them onto the battlefield",
		ThenTail: wilfredMottRest,
	}.Apply(ctx)
}

var wilfredMottRest = PutFromLibraryTail("wilfred-mott/rest-to-bottom",
	func(ctx *Context, res PutFromLibraryResult) error {
		return PutRestOnBottomInRandomOrder(ctx.Game, res)
	})
