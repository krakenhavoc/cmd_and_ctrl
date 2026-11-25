package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b095_helpers.go — shared shapes of the Zendikar Rising batch 095
// cards (#5508): the party discount on a spell, the "is also a Cleric, Rogue,
// Warrior, and Wizard" type line, and a one-target "can't be blocked this
// turn" ability body.

// OldB095AlsoPartyTypes is "<this> is also a Cleric, Rogue, Warrior, and
// Wizard" (Stonework Packbeast, Veteran Adventurer): a layer 4 static that
// adds the four party creature types to the card itself (CR 205.3d, 613.1d).
func OldB095AlsoPartyTypes(name string) game.StaticAbility {
	return game.StaticAbility{
		Label:     name + " is also a Cleric, Rogue, Warrior, and Wizard.",
		Layer:     game.Layer4Type,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, t := range FinB9PartyTypes {
				if !EotHasType(c.Subtypes, t) {
					c.Subtypes = append(c.Subtypes, t)
				}
			}
		},
	}
}

// OldB095TargetCantBeBlocked is the body of "target creature can't be blocked
// this turn" (Sneaking Guide): the first target still legal as the ability
// resolves (CR 608.2b), made unblockable until end of turn.
func OldB095TargetCantBeBlocked(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: label}.Apply(ctx)
	}
}
