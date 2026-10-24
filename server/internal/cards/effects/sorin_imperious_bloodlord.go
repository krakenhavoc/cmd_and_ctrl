package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorin, Imperious Bloodlord — Legendary Planeswalker — Sorin {2}{B},
// loyalty 4 (#4077):
//
//	"+1: Target creature you control gains deathtouch and lifelink until
//	     end of turn. If it's a Vampire, put a +1/+1 counter on it.
//	 +1: You may sacrifice a Vampire. When you do, Sorin deals 3 damage
//	     to any target and you gain 3 life.
//	 −3: You may put a Vampire creature card from your hand onto the
//	     battlefield."
//
// FIRST +1. The grant is a layer-6 keyword effect locked to the creature
// as it is now (CR 611.2c); "if it's a Vampire" is read as the ability
// resolves, off the creature's characteristics then, through Sorin's own
// text (CR 612.2), so a changeling counts. A creature that left in
// response is skipped (CR 608.2b), grant and counter alike.
//
// SECOND +1 is the CR 603.12 shape: "you may sacrifice a Vampire" is
// asked as the ability resolves, over whatever you control THEN (any
// permanent with the type, not only a creature), and "when you do" is a
// reflexive trigger created only by a sacrifice that really happened. It
// goes on the stack above the ability and picks its "any target" as it
// does (CR 603.3d); the damage and the life gain are one body, so a
// target that left still lets the 3 life through. The trigger's source is
// Sorin, so lifelink or a prevention on him sees the damage.
//
// THE −3 is Quicksilver Amulet's put (PutFromHandOntoBattlefield): the
// card is put, not cast, so no cast trigger sees it, and "may" is a real
// decline. Vampire creature cards only; a Vampire that is not a creature
// card could not enter as a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fee680c6-afe2-4321-9d8a-8f81a612a53b",
		Name:            "Sorin, Imperious Bloodlord",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Target creature you control gains deathtouch and lifelink until end of turn. If it's a Vampire, put a +1/+1 counter on it.",
				Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch", "lifelink")),
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("target creature you control", YouControl()),
				Effect:  sorinImperiousGrant,
			},
			{
				Label:  "+1: You may sacrifice a Vampire. When you do, Sorin deals 3 damage to any target and you gain 3 life.",
				Cost:   LoyaltyCost(1),
				Effect: sorinImperiousMaySacrifice,
			},
			{
				Label: "−3: You may put a Vampire creature card from your hand onto the battlefield.",
				Cost:  LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return PutFromHandOntoBattlefield{
						Match:    And(Creature(), HasSubtype("Vampire")),
						Optional: true,
						Label:    "Sorin, Imperious Bloodlord — you may put a Vampire creature card from your hand onto the battlefield",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// sorinImperiousGrant is the first +1, in printed order.
func sorinImperiousGrant(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (GrantKeywordUntilEOT{
		Target:   target,
		Keywords: []string{"deathtouch", "lifelink"},
		Label:    "Sorin, Imperious Bloodlord — deathtouch and lifelink",
	}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(target)
	if !ok || !c.HasSubtype(ctx.TextSubtype("Vampire")) {
		return nil
	}
	return AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}

// sorinImperiousMaySacrifice is the second +1: the optional sacrifice,
// then the reflexive strike for a Vampire that really went.
func sorinImperiousMaySacrifice(g *game.Game, item *game.StackItem) error {
	return maySacrificeOne(g, item, "Sorin, Imperious Bloodlord — sacrifice a Vampire to deal 3 damage to any target and gain 3 life?",
		HasSubtype("Vampire"), false,
		func(ctx *Context, _ uuid.UUID) error {
			return WhenYouDo("Sorin, Imperious Bloodlord — 3 damage to any target and gain 3 life",
				sorinImperiousStrikeBody).Apply(ctx)
		})
}

// sorinImperiousStrike is the reflexive trigger's body: Sorin deals 3
// damage to the target chosen as it went on the stack, if still legal,
// and its controller gains 3 life whether or not the damage landed.
func sorinImperiousStrike(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if legal := ctx.LegalTargets(); len(legal) > 0 {
		if err := (DealDamage{Source: item.SourceCardID, Target: legal[0].ID, Amount: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
}
