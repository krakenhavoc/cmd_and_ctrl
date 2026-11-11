package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knickknack Ouphe — Creature — Ouphe {X}{G}, 1/1 (Wilds of Eldraine
// Commander, #5368):
//
//	"This creature enters with X +1/+1 counters on it.
//	 When this creature enters, reveal the top X cards of your library. You
//	 may put any number of Aura cards with mana value X or less from among
//	 them onto the battlefield. Then put all cards revealed this way that
//	 weren't put onto the battlefield on the bottom of your library in a
//	 random order."
//
// X is the X the spell was cast with (CR 107.3m), read by the enters trigger
// as Hugs, Grisly Guardian's is and as the counters are
// (EntersWithCountersFromCast). The reveal is public. The put is Armored
// Skyhunter's: any number of the matching cards go onto the battlefield as
// one entry, each Aura choosing its host as it enters (CR 303.4f, so the
// Aura that finds no legal host stays in the revealed pile and goes to the
// bottom with the rest), and the rest go to the bottom in a random order.
// Mana value is read off the card in the library ({X} counts as 0 there,
// CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "55adf2f0-f023-4f41-ae9d-3c002e9e6e8d",
		Name:                       "Knickknack Ouphe",
		XMatters:                   true,
		Completeness:               CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Knickknack Ouphe — reveal the top X cards; you may put any number of Aura cards with mana value X or less onto the battlefield; the rest on the bottom in a random order",
			XFromCast: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, x,
					"Knickknack Ouphe — reveal the top X cards of your library")
				if len(revealed) == 0 {
					return nil
				}
				return PutFromLibraryOntoBattlefield{
					Player:   item.Controller,
					Cards:    revealed,
					Match:    And(Subtype("Aura"), ManaValueLE(x)),
					Max:      0,
					Optional: true,
					Label:    "Knickknack Ouphe — you may put any number of Aura cards with mana value X or less from among them onto the battlefield",
					ThenTail: PutRestOnBottomInRandomOrderTail,
				}.Apply(ctx)
			},
		}},
	})
}
