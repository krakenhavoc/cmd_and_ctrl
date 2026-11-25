package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB133DevotionBoostFirstTarget is the body of a Theros Beyond Death
// entry trigger that gives its one target +X/+0 or -X/-X (or similar) until
// end of turn, "where X is your devotion to <colour>" (CR 700.5). The
// devotion is counted once as the trigger resolves, from the controller's
// permanents, and `colorLetter` is read through the card's text (CR 612.2).
// powerPer and toughnessPer are the printed signs: +1/+0 is (1, 0), -1/-1 is
// (-1, -1).
func S82OldB133DevotionBoostFirstTarget(label, colorLetter string, powerPer, toughnessPer int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		x := DevotionTo(g, item.Controller, ctx.TextColor(colorLetter))
		return BoostUntilEOT{Target: target, Power: powerPer * x, Toughness: toughnessPer * x, Label: label}.Apply(ctx)
	}
}

// S82OldB133ReturnedAuraSelfForm is ReturnedAuraForm for a returned Aura that
// HAS the granted ability itself rather than handing it to what it enchants
// (Bronzehide Lion: the Aura has "{G}{W}: Enchanted creature gains
// indestructible until end of turn"; Harold and Bob's Aura gives its Forest
// the ability). The grant goes to the permanent whose Card.ReturnedAura is set,
// declared AFTER the removal so the "loses all other abilities" half has
// already emptied the card when it lands (CR 613.1f, 613.6; a grant sorted
// after the removal survives, ability_removal.go); the Aura type and the
// removal are ReturnedAuraForm's own.
func S82OldB133ReturnedAuraSelfForm(grantKey string) []game.StaticAbility {
	grant := GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.InstanceID == source.InstanceID && source.ReturnedAura
	}, grantKey)
	// The Aura's own type change (layer 4) started this effect on the permanent,
	// so its layer-6 grant continues after the removal silences its source
	// (CR 613.6, the same half the Aura's other statics are).
	grant.ContinuesAfterRemoval = true
	return append(ReturnedAuraForm(grantKey)[1:], grant)
}
