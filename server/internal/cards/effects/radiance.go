package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// radiance.go — "target X and each other X that shares a color with it"
// (#5736), the Ravnica: City of Guilds radiance cards.
//
// Radiance is an ability word (CR 207.2c): it has no rules meaning of its
// own, so each card's text is followed as printed. What those texts share
// is one set, built here once:
//
//   - The others are found as the spell or ability RESOLVES (CR 608.2,
//     608.2h), not when it is cast or activated, so a creature that
//     arrived in response is included and one that left is not.
//   - "Shares a color" is read against the TARGET's colours at that
//     moment, so a target recoloured in response brings a different set
//     (the cards' rulings: "You check which creatures share a color with
//     the target when the spell resolves").
//   - A colourless target has no colour to share (CR 105.2c), so it is
//     the only object affected.
//   - Only the first object is a target. The others are affected without
//     being targeted (CR 115.10a), so hexproof, shroud and protection's
//     "can't be targeted" (CR 702.11b, 702.18a, 702.16b) do not keep
//     them out; protection's damage prevention still applies to damage
//     dealt to them (CR 702.16e), because that is the damage event's
//     business, not this set's.
//   - A target no longer legal makes the whole spell or ability do
//     nothing (CR 608.2b): there is one target, so no target is left and
//     the engine never resolves it. The set is empty in that case too,
//     so a caller that is reached anyway affects nothing.
//   - "Each other X" is the card's own kind (a creature, an enchantment),
//     a CardPredicate, and "other" is object identity.
//
// The set is read once (CR 608.2h), and a continuous effect a card makes
// from it locks its affected objects then (CR 611.2c): a creature that
// changes colour later in the turn keeps what it was given.

// radianceSet is the objects a radiance spell or ability affects, read
// now: its target first, then each other battlefield permanent `kind`
// admits that shares a colour with the target as it is now. Nil when the
// target is no longer legal.
func radianceSet(ctx *Context, kind CardPredicate) []uuid.UUID {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		target, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		ids := []uuid.UUID{target.InstanceID}
		for _, c := range MatchingBattlefield(ctx, kind) {
			if c.InstanceID != target.InstanceID && sharesAColorNow(c, target) {
				ids = append(ids, c.InstanceID)
			}
		}
		return ids
	}
	return nil
}

// eachRadiant applies `act` to every object of a radiance set, the target
// first. The others are members of a set the text names by group, so
// they are acted on through asGroupMember (#1463): the source of a
// radiance ability that is itself among them is acted on as printed.
func eachRadiant(ctx *Context, ids []uuid.UUID, act func(c *Context, id uuid.UUID) error) error {
	for i, id := range ids {
		c := ctx
		if i > 0 {
			c = ctx.asGroupMember()
		}
		if err := act(c, id); err != nil {
			return err
		}
	}
	return nil
}

// radianceDamage is "<source> deals N damage to target creature and each
// other creature that shares a color with it" (Cleansing Beam, Wojek
// Embermage, Brightflame): one damage instruction over the whole set,
// with `then`, when set, told the total actually dealt after prevention
// and replacement (Brightflame's "You gain life equal to the damage dealt
// this way").
func radianceDamage(ctx *Context, n int, then game.KeyedTail) error {
	ids := radianceSet(ctx, Creature())
	if len(ids) == 0 || n <= 0 {
		return nil
	}
	return ctx.Game.DealDamageEachKeyedForEffect(ctx.Source(), ids, n, then)
}

// radianceProtectionFromChosenColor is the rest of Bathe in Light once
// its colour is chosen: every creature the tail carries (the radiance
// set, read as the spell resolved) gains protection from that colour
// until end of turn. The card's name is the tail's first Word, the colour
// its Word. An on-disk key.
var radianceProtectionFromChosenColor = CardTail("then/radiance-protection-from-chosen-color-until-eot", func(ctx *Context, o game.TailOutcome) error {
	token := game.ProtectionFromColor(o.Tail.Word)
	if token == "" || len(o.Tail.Carry) == 0 || len(o.Tail.Words) == 0 {
		return nil
	}
	return GrantKeywordUntilEOT{
		Match:    b10OneOf(o.Tail.Carry),
		Keywords: []string{token},
		Label:    o.Tail.Words[0] + " — " + token,
	}.Apply(ctx)
})
