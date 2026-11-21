package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stalking Vengeance — Creature — Avatar {5}{R}{R}, 5/5:
//
//	"Haste
//	 Whenever another creature you control dies, it deals damage equal to
//	 its power to target player or planeswalker."
//
// The dead creature, not the Vengeance, is the damage source, named by
// OBJECT (Aether Charge's shape), so its keywords as it last existed apply.
// The amount is that creature's power as it last existed on the battlefield
// (CR 603.10, last-known information), read as the ability resolves. The
// trigger is "another creature you control dies": the type and controller
// it had as it left (diedCreature, leftUnderControlOf), the Vengeance
// excluded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5f4ff27f-ebc1-4a86-8b0b-eeea470a25fb",
		Name:            "Stalking Vengeance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AnotherCreatureYouControlDied(ev, source, g)
			}, "Stalking Vengeance — the creature deals damage equal to its power to target player or planeswalker",
				stalkingVengeanceDamage),
			effects.TargetPlayerOrPlaneswalker())},
	})
}

func stalkingVengeanceDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	obj := ctx.Trigger().Object
	if obj == nil {
		return nil
	}
	info, ok := ctx.TriggeringPermanent()
	if !ok {
		return nil
	}
	ref := obj.Ref()
	for _, t := range ctx.LegalTargets() {
		return effects.DealDamage{SourceObject: &ref, Target: t.ID, Amount: max(info.Power, 0)}.Apply(ctx)
	}
	return nil
}
