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
// Two sandbox simplifications, both declared, both weaker:
//
//   - The X tokens are created as the spell RESOLVES, a beat before
//     the Parade moves from the stack to the battlefield — the
//     Goldvein Hydra posture: an ETB trigger cannot see the X
//     announced for the spell (the stack item is gone by the time
//     the trigger fires), and resolution is the last moment X is
//     readable. So the tokens come without a trigger on the stack to
//     respond to; nothing is stronger for it, because the spell
//     itself could be countered.
//   - The mana ability is granted only to the Shapeshifters the
//     Parade makes, not to every creature token you control. A mana
//     ability cannot be granted to another permanent by a static
//     (ManaAbilitiesForCard reads the token's own list or the
//     catalog by oracle ID, and nothing in between), so it lives on
//     the token template with a Condition that the token's
//     controller still controls a Springleaf Parade — off when the
//     Parade leaves, back when another arrives, never on a Goblin
//     from Krenko's Command.
func init() {
	Register(Spec{
		OracleID:     "b1305916-53cc-4021-897e-bbefc65dce78",
		Name:         "Springleaf Parade",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Shapeshifters are created as the spell resolves rather than by an enters trigger you can respond to.",
			"Only the Shapeshifters it makes get \"{T}: Add one mana of any color\" — other creature tokens you control don't.",
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := ctx.X()
			if n <= 0 {
				return nil
			}
			return CreateToken{
				Controller: item.Controller,
				Template:   b18SpringleafShapeshifterToken(),
				N:          n,
			}.Apply(ctx)
		},
	})
}
