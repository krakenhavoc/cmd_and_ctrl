package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Springleaf Parade — Enchantment {X}{G}{G} (EDHREC rank 1927):
//
//	"When this enchantment enters, create X 1/1 colorless
//	 Shapeshifter creature tokens with changeling. (They're every
//	 creature type.)
//	 Creature tokens you control have "{T}: Add one mana of any
//	 color.""
//
// X changelings that each tap for any colour: the tribal deck's
// ramp-and-bodies spell. The changeling keyword is real on the
// token (Card.Keywords, the Maskwood Nexus posture).
//
// "When this enchantment enters" is a printed ENTERS TRIGGER, not an
// entry replacement — before #1312 an enters trigger's Build had
// nowhere to read the announced X from (the resolving spell's
// StackItem is gone by the time the trigger fires), so this used to
// move the read into OnResolve, a beat before the Parade left the
// stack, with a declared caveat that the tokens arrived without a
// trigger to respond to. #1357: Build now reads source.CastX()
// (CastProvenance.X, CR 107.3m) and closes over the plain int rather
// than the card, so the trigger is a real CR 603 object on the
// stack — it can be countered, or the Parade can be removed in
// response to it (the trigger still resolves on its own
// last-known-X, CR 603.10), which the resolve-time shortcut could not
// model.
//
// The mana ability is ADR 0093's layer-6 grant to every creature token
// you control — the Shapeshifters, and a Goblin from Krenko's Command
// too. Until ADR 0093 it lived on the Shapeshifter's token template,
// gated on the controller still controlling a Parade, and reached no
// other token; the template is gone and the Shapeshifter is a plain
// changeling row.
//
// No simplification.
const springleafParadeGrant = "springleaf-parade/any-color"

func init() {
	Register(Spec{
		OracleID:     "b1305916-53cc-4021-897e-bbefc65dce78",
		Name:         "Springleaf Parade",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(springleafParadeGrant)},
		Static:       []game.StaticAbility{GrantAbilities(creatureTokensYouControl, springleafParadeGrant)},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Springleaf Parade — create X Shapeshifters",
			// #1312/#1357: read once, here (ADR 0041 P9's fill-in
			// Build), and stamp the plain int on Params.Amount rather
			// than closing over the card, which the Effect must not
			// capture.
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Springleaf Parade — create X Shapeshifters", nil)
				item.Params.Amount = source.CastX()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := item.Params.Amount
				if x <= 0 {
					return nil
				}
				return CreateToken{
					Controller: item.Controller,
					Template:   TokenCard("1/1 colorless Shapeshifter with changeling"),
					N:          x,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
