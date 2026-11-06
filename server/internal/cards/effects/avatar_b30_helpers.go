package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b30_helpers.go — shared pieces of Avatar: The Last Airbender
// Eternal batches 30–33 (#5075). Append-only.

// avB30aPowerEqualsCreaturesYouControl is "<this>'s power is equal to the
// number of creatures you control" (Freedom Fighter Recruit): Adeline's
// layer-7a characteristic-defining ability (CR 604.3), self-only. It
// functions in every zone (CR 604.3, the card's ruling): the engine lends a
// card off the battlefield its own reading (Game.withPTAnywhereLocked).
func avB30aPowerEqualsCreaturesYouControl(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7A_CDA,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			c.Power = b04CreaturesControlled(g, source.Controller)
		},
	}
}

// avB30aPutCounterOnEachOfSubtypeYouControl is "put a +1/+1 counter on each
// <subtype> you control" (Lo and Li's Advisors): the set is read as the
// ability resolves and every counter is placed by b13PutCounterOnEach. The
// subtype is a word of the source's text (CR 612.2).
func avB30aPutCounterOnEachOfSubtypeYouControl(subtype string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		word := g.TextReaderForEffect(item).TextSubtype(subtype)
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == item.Controller && c.HasSubtype(word) {
				ids = append(ids, c.InstanceID)
			}
		}
		return b13PutCounterOnEach(NewContext(g, item), ids)
	}
}

// avB30aYouControlAnAttackingLegendaryCreature is Scarring Memories' "if you
// control an attacking legendary creature": a legendary creature you control
// that is currently declared as an attacker (Card.AttackingTarget, cleared
// as it leaves combat).
func avB30aYouControlAnAttackingLegendaryCreature(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsCreature() && c.IsLegendary() && c.AttackingTarget != uuid.Nil {
			return true
		}
	}
	return false
}

// avB30cEachPlayerCreatesAFood is "each player creates a Food token": one
// Food for every seat, in turn order from the active player, each created
// under that player's control (Party Dude, Elephant-Mandrill).
func avB30cEachPlayerCreatesAFood(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, p := range seatsFromActive(g) {
		if err := (CreateToken{Controller: p, Template: FoodToken(), N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// avB30cSacrificeHalfRoundedUpCost is "As an additional cost to cast this
// spell, sacrifice half the <permanents> you control, rounded up":
// avB30cSacrificeHalfRoundedUpCost("lands you control", Land(),
// YouControl()) is Tectonic Split's. The caster names the permanents on
// cast_spell's sacrifice_ids, and CastSpell refuses any count but half of
// the matching permanents they control, rounded up, counted as the cost
// is determined (game.AdditionalCost.SacrificeHalfRoundedUp). The clause
// carries the "any number" bounds, as SacrificeAllCost's does, so the
// validator checks what was named and the flag says how many. Only the
// mandatory slot may carry it (avB30cCheckSacrificeHalfCost).
func avB30cSacrificeHalfRoundedUpCost(label string, preds ...CardPredicate) *game.AdditionalCost {
	return &game.AdditionalCost{
		Sacrifice:              sacrificeSpec(label, preds...).WithCount(0, 0),
		SacrificeHalfRoundedUp: true,
		Label:                  "Sacrifice half the " + label + ", rounded up",
	}
}

// avB30cCheckSacrificeHalfCost is Register's guard for
// SacrificeHalfRoundedUp, refusing the shapes checkSacrificeAllCost
// refuses for its flag: none on an optional cost or an either/or branch,
// none without an "any number" sacrifice clause, none with a set rule,
// and never both flags on one cost.
func avB30cCheckSacrificeHalfCost(spec Spec) {
	for _, oc := range spec.OptionalCosts {
		if oc.SacrificeHalfRoundedUp {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q sacrifices half — only a cast's mandatory additional cost prints that", spec.Name, oc.Key))
		}
	}
	ac := spec.AdditionalCost
	if ac == nil {
		return
	}
	for i, b := range ac.Either {
		if b.SacrificeHalfRoundedUp {
			panic(fmt.Sprintf("effects.Register: %q either/or branch %d sacrifices half — only a cast's mandatory additional cost prints that", spec.Name, i))
		}
	}
	if !ac.SacrificeHalfRoundedUp {
		return
	}
	switch {
	case ac.Sacrifice == nil:
		panic(fmt.Sprintf("effects.Register: %q sets SacrificeHalfRoundedUp with no sacrifice clause — build it with avB30cSacrificeHalfRoundedUpCost", spec.Name))
	case !game.SacrificeAnyNumber(ac.Sacrifice):
		panic(fmt.Sprintf("effects.Register: %q sets SacrificeHalfRoundedUp on a clause with a count — its bounds are \"any number\"", spec.Name))
	case len(ac.Sacrifice.EachOf) > 0:
		panic(fmt.Sprintf("effects.Register: %q sets SacrificeHalfRoundedUp on a clause with a set rule", spec.Name))
	case ac.SacrificeAll:
		panic(fmt.Sprintf("effects.Register: %q sets both SacrificeAll and SacrificeHalfRoundedUp", spec.Name))
	}
}
