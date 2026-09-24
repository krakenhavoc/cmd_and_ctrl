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
// One sandbox simplification remains, declared and weaker: the mana
// ability is granted only to the Shapeshifters the Parade makes, not
// to every creature token you control. A mana ability cannot be
// granted to another permanent by a static (ManaAbilitiesForCard
// reads the token's own list or the catalog by oracle ID, and
// nothing in between), so it lives on the token template with a
// Condition that the token's controller still controls a Springleaf
// Parade — off when the Parade leaves, back when another arrives,
// never on a Goblin from Krenko's Command.
func init() {
	Register(Spec{
		OracleID:     "b1305916-53cc-4021-897e-bbefc65dce78",
		Name:         "Springleaf Parade",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Only the Shapeshifters it makes get \"{T}: Add one mana of any color\" — other creature tokens you control don't.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				// #1312/#1357: read once, here, and close over the
				// plain int — not the card, which the Effect below
				// must not capture (AGENTS.md: undo restores a
				// cloned game and the closure has to resolve
				// against that one).
				x := source.CastX()
				return game.NewTriggeredItem(source, "Springleaf Parade — create X Shapeshifters",
					func(g *game.Game, item *game.StackItem) error {
						if x <= 0 {
							return nil
						}
						return CreateToken{
							Controller: item.Controller,
							Template:   b18SpringleafShapeshifterToken(),
							N:          x,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
