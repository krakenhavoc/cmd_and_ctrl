package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodghast — 2/1 Creature — Vampire Spirit for {B}{B}:
//
//	"This creature can't block.
//	 This creature has haste as long as an opponent has 10 or less life.
//	 Landfall — Whenever a land you control enters, you may return
//	 this card from your graveyard to the battlefield."
//
// The card #925 was opened for: a landfall trigger that watches from
// the GRAVEYARD, which is the only zone it can ever fire from —
// "return this card FROM YOUR GRAVEYARD" says where the ability
// lives, and the battlefield copy of the ability would have nothing
// to return.
//
// "You control" and "your graveyard" both mean the OWNER here
// (CR 108.4): a card in a graveyard has no controller, so the
// harvest reads the trigger's "you" off Card.Owner and it is the
// owner's land drop that wakes it up.
//
// The return is not a cast and not a recursion cost — Bloodghast is
// a free body every time a land enters, which is why it is a staple
// of aristocrat decks and why the "you may" (CR 603.5) matters: a
// player holding it back for a sacrifice outlet declines.
func init() {
	Register(Spec{
		OracleID:     "e97f9c2b-b41e-4f36-9245-77c0ac125647",
		Name:         "Bloodghast",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Its conditional haste is omitted: a static keyed on an opponent's life total goes stale, because a life-total change does not invalidate the layer cache (the \"Layer invalidation on hand / life / attack / graveyard state\" seam). Bloodghast is always summoning-sick on the turn it returns.",
		},
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBlock),
		},
		Triggered: []game.TriggeredAbility{
			Optional(InGraveyard(Landfall(
				"Bloodghast — return it from your graveyard to the battlefield",
				returnThisCardFromYourGraveyard)), "Bloodghast — return it to the battlefield?"),
		},
	})
}
