package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hired Hexblade — Creature — Elf Warlock {1}{B}, 2/2 (EDHREC rank
// 24000):
//
//	"When this creature enters, if mana from a Treasure was spent to
//	 cast it, you draw a card and you lose 1 life."
//
// The first card to read the SOURCE of the mana that paid for it
// (#1212), and the reason the source snapshot could not wait for the
// spend riders. The fact it asks about is destroyed twice over before
// it is asked:
//
//  1. The Treasure sacrifices itself to pay for its own mana ability,
//     so by the time this spell is even cast the permanent is gone —
//     and a Treasure TOKEN has ceased to exist entirely (CR 111.7),
//     so `ManaToken.Source` names nothing a lookup could find. The
//     engine answers that by snapshotting the source's kinds when the
//     mana is MADE (game/mana_source.go).
//  2. The spell has finished resolving before this trigger does.
//     "When this creature enters" is a TRIGGER, not an entry
//     replacement, so `StackItem.Paid` is gone by the time the ability
//     is on the stack — the same wall "sacrifice it unless it escaped"
//     hit (#653) and the same answer: the tokens ride
//     `Card.Provenance` onto the permanent (CR 400.7d).
//
// CR 603.4, and it is why the read is in AppliesTo rather than in the
// body: "if mana from a Treasure was spent to cast it" is an
// INTERVENING IF. An ordinary Hexblade puts nothing on the stack and
// nobody gets a window to respond to a trigger that is not happening.
// The record cannot change between the two checks the rule asks for,
// so one check is the whole of it.
//
// WantsManaFrom is the auto-tapper's ordering hint: a card that reads
// Treasure mana would rather be paid with some. DECLARED INERT for
// now — `autoTapAbilityFor` refuses every sacrifice-cost mana ability
// (`!a.TapCost || a.SacrificeCost`), so a Treasure is not an auto-tap
// source at all and the hint cannot reach one. Declared anyway,
// because the declaration is about the CARD and is right whether or
// not the planner can act on it. Cracking the Treasure by hand — which
// is the only way to spend it today — works.
func init() {
	Register(Spec{
		OracleID:      "f3a0f155-05d8-465c-b0ef-35aa12e93013",
		Name:          "Hired Hexblade",
		Completeness:  CompletenessCaveats,
		Caveats:       []string{"With strict mana off, the game doesn't see which mana you spent, so Hired Hexblade never draws the card."},
		WantsManaFrom: game.ManaSourceTreasure,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.ManaSpentToCast().FromTreasure()
			}), "Hired Hexblade — draw a card and lose 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
						return err
					}
					// "and you lose 1 life" — printed order, drawn
					// first, and a real loss rather than a cost, so
					// a life-loss replacement (#808) sees it.
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
				}),
		},
	})
}
