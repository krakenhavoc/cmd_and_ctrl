package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scg_b1_helpers.go — shared shapes for Scourge, batch 1 (#5608, S82).

// scgB1PermanentCardOfSubtype passes a permanent card with the subtype:
// Dragonstorm's "Dragon permanent card". A Dragon creature card counts, and
// so does a kindred Dragon artifact or enchantment; a Dragon instant or
// sorcery does not. A changeling card is a Dragon (CR 702.73a). The caller
// reads the word through the source's text (ctx.TextSubtype, CR 612.2).
func scgB1PermanentCardOfSubtype(subtype string) func(game.Card) bool {
	return func(c game.Card) bool { return isPermanentCard(c) && c.HasSubtype(subtype) }
}

// scgB1GreatestManaValue is "the greatest mana value among permanents you
// control", read as the spell or ability resolves (Rush of Knowledge,
// Torrent of Fire, Accelerated Mutation, Dispersal Shield, Reward the
// Faithful, Cabal Conditioning). Zero with no permanents. A face-down
// permanent's mana value is 0 (CR 708.2), and {X} in a permanent's cost is
// 0 (CR 202.3e).
func scgB1GreatestManaValue(ctx *Context) int {
	return cmdrB13GreatestManaValueYouControl(ctx.Game, ctx.Controller())
}

// scgB1DamageThatPlayer is a cast trigger's "this deals N damage to that
// player", where that player is the one who cast the spell (Pyrostatic
// Pillar). The trigger's source deals the damage.
func scgB1DamageThatPlayer(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: n}.Apply(NewContext(g, item))
	}
}

// scgB1HostGainsUntilEOT is "enchanted creature gains <keyword> until end
// of turn" for an Aura's own ability (Pemmin's Aura). The host is found
// through the Aura as the ability resolves, so an Aura that has fallen off
// by then grants nothing.
func scgB1HostGainsUntilEOT(keyword, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		host := attachedHostFor(g, item.SourceCardID)
		if host == nil {
			return nil
		}
		return GrantKeywordUntilEOT{Target: host.InstanceID, Keywords: []string{keyword}, Label: label}.Apply(NewContext(g, item))
	}
}

// scgB1HostGetsUntilEOT is "enchanted creature gets +N/+M until end of
// turn" for an Aura's own ability, read the same way as
// scgB1HostGainsUntilEOT.
func scgB1HostGetsUntilEOT(ctx *Context, power, toughness int, label string) error {
	host := attachedHostFor(ctx.Game, ctx.Item.SourceCardID)
	if host == nil {
		return nil
	}
	return BoostUntilEOT{Target: host.InstanceID, Power: power, Toughness: toughness, Label: label}.Apply(ctx)
}
