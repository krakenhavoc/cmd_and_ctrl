package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farmer Cotton — Legendary Creature — Halfling Peasant {X}{G}{W},
// 1/1 (EDHREC rank 2909):
//
//	"When this creature enters, create X 1/1 white Halfling creature
//	 tokens and X Food tokens. (They're artifacts with "{2}, {T},
//	 Sacrifice this token: You gain 3 life.")"
//
// X Halflings and X Food on a one-drop body. "When this creature
// enters" is a printed ENTERS TRIGGER, not an entry replacement —
// before #1312 an enters trigger's Build had nowhere to read the
// announced X from (the resolving spell's StackItem is gone by the
// time the trigger fires), so this used to move the read into
// OnResolve, a beat before Cotton left the stack, with a declared
// caveat that the tokens arrived without a trigger to respond to.
// #1357: Build now reads source.CastX() (CastProvenance.X, CR
// 107.3m) and closes over the plain int rather than the card, so the
// trigger is a real CR 603 object on the stack — it can be countered,
// or Cotton can be removed in response to it (the trigger still
// resolves on its own last-known-X, CR 603.10), which the
// resolve-time shortcut could not model. Halflings first, then Food,
// as printed; the Food is the real Food token.
func init() {
	Register(Spec{
		OracleID:     "11f1d2b7-0c2b-40df-bf90-3c55b23449af",
		Name:         "Farmer Cotton",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Farmer Cotton — create X Halflings and X Food",
			// #1312/#1357: read once, here (ADR 0041 P9's fill-in
			// Build), and stamp the plain int on Params.Amount rather
			// than closing over the card, which the Effect must not
			// capture.
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Farmer Cotton — create X Halflings and X Food", nil)
				item.Params.Amount = source.CastX()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := item.Params.Amount
				if x <= 0 {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Halfling"), N: x}).Apply(ctx); err != nil {
					return err
				}
				return CreateToken{Controller: item.Controller, Template: FoodToken(), N: x}.Apply(ctx)
			},
		}},
	})
}
