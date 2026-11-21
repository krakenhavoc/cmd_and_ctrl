package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morgul-Knife Wound — Enchantment — Aura {1}{B} (The Lord of the Rings:
// Tales of Middle-earth, #5948):
//
//	"Enchant creature
//	 Enchanted creature gets -3/-0 and has "At the beginning of your
//	 upkeep, exile this creature unless you pay 2 life.""
//
// The granted trigger belongs to the enchanted creature, so "your" and "you"
// are that creature's controller and "this creature" is the creature, not the
// Aura. As it resolves the controller may pay 2 life (CR 119.4: a player with
// less than 2 life, or one whose life can't change, can't pay) or the creature
// is exiled; with no way to pay it is exiled without a question. A creature
// that has left or come back as a new object is not exiled (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f48b2349-797e-4e79-aea7-2413e78129e1",
		Name:         "Morgul-Knife Wound",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Targets:      EnchantCreature(),
		Grants: []AbilityGrant{{
			Key: morgulKnifeWoundGrant,
			Triggered: []game.TriggeredAbility{
				AtYourUpkeep("Morgul-Knife Wound — exile this creature unless you pay 2 life", morgulKnifeWoundUpkeep),
			},
			Text: "At the beginning of your upkeep, exile this creature unless you pay 2 life.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(-3, 0),
			GrantAbilitiesToAttached(morgulKnifeWoundGrant),
		},
	})
}

const morgulKnifeWoundGrant = "morgul-knife-wound/exile-unless-pay-life"

func morgulKnifeWoundUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || !g.CanPayLifeLocked(p, 2) {
		return ExileTarget{Target: item.SourceCardID}.Apply(ctx)
	}
	return PickOption{
		Question: "Morgul-Knife Wound — pay 2 life, or exile this creature",
		Options:  []game.ChoiceOption{{Label: "Exile this creature"}, {Label: "Pay 2 life", LifeCost: 2}},
		ThenKey:  morgulKnifeWoundChose,
		WithItem: true,
	}.Apply(ctx)
}

// morgulKnifeWoundChose is the keyed answer: paying life keeps the creature,
// anything else exiles it. An on-disk key: never renamed, never reused.
var morgulKnifeWoundChose = OptionPickThen("option-pick/morgul-knife-wound-pay-or-exile", func(ctx *Context, r game.OptionPicked) error {
	item := ctx.Item
	if !onBattlefield(ctx.Game, item.SourceCardID) || sourceIsNewObject(ctx.Game, item) {
		return nil
	}
	if r.Option != nil && r.Option.Label == "Pay 2 life" {
		if err := ctx.Game.PayLifeForEffect(item.SourceCardID, item.Controller, 2); err == nil {
			return nil
		}
	}
	return ExileTarget{Target: item.SourceCardID}.Apply(ctx)
})
