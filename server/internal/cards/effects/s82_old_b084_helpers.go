package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB084AttachedIsColorInAddition is "Equipped creature is <colour> in addition
// to its other colors" (Angelic Armaments): a layer-5 addition. The colour is a
// word of the Equipment's text (CR 612.2, #3984), read through it as the pass
// applies it, and CR 613.6 keeps the effect applying once layer 4 has started.
func OldB084AttachedIsColorInAddition(color, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:                 label,
		Layer:                 game.Layer5Color,
		ContinuesAfterRemoval: true,
		AppliesTo:             AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			want := source.TextColor(color)
			if !HasFold(c.Colors, want) {
				c.Colors = append(c.Colors, want)
			}
		},
	}
}

// OldB084UpkeepCounterIfEquippedIs is Ring of Valkas / Ring of Thune: "At the
// beginning of your upkeep, put a +1/+1 counter on equipped creature if it's
// <colour>." The "if" closes the sentence, so it is read as the ability
// resolves (CR 608.2c), not as an intervening "if" (CR 603.4): the host and its
// colour are the ones it has then. An unattached Equipment does nothing.
func OldB084UpkeepCounterIfEquippedIs(label, color string) game.TriggeredAbility {
	return AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		host, ok := Cmdr2B21EquippedCreature(g, ctx.Source())
		if !ok {
			return nil
		}
		word := color
		if src, ok := g.LookupCardForEffect(item.SourceCardID); ok {
			word = src.TextColor(color)
		}
		if !HasFold(host.EffectiveColors(), word) {
			return nil
		}
		return AddCounter{Target: host.InstanceID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	})
}
