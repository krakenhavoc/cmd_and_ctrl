package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tadeas, Juniper Ascendant — Legendary Creature — Human Monk {2}{G}{W}, 1/3
// (Secret Lair Drop):
//
//	"Reach
//	 Tadeas has hexproof unless it's attacking.
//	 Whenever a creature you control with reach attacks, untap it and it can't
//	 be blocked by creatures with greater power this combat.
//	 Whenever one or more creatures you control deal combat damage to a player,
//	 draw a card."
//
// The hexproof is a layer 6 grant on Tadeas alone while it is not an attacking
// creature. The attack trigger fires once per attacking creature of yours that
// has reach as it attacks (read through the layers); it untaps that creature
// and gives it skulk (CR 702.118: "can't be blocked by creatures with greater
// power") until end of combat, this combat only (CR 611.2b). The last trigger is
// once for each player damaged by one or more creatures you control in a
// combat damage step (CR 603.2c).
//
// No simplification.
func init() {
	hexproofUnlessAttacking := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget == uuid.Nil
	}, "hexproof")
	hexproofUnlessAttacking.Label = "Tadeas has hexproof unless it's attacking."
	hexproofUnlessAttacking.DependsOnAttackingStatus = true
	effects.Register(effects.Spec{
		OracleID:        "b4b5b928-8b98-4f65-9f15-212859874fe4",
		Name:            "Tadeas, Juniper Ascendant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{Draws: 1},
		Static:          []game.StaticAbility{hexproofUnlessAttacking},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventAttack || ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && game.HasKeyword(&c, "reach")
			}, "Tadeas, Juniper Ascendant — untap it; it can't be blocked by creatures with greater power this combat",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					id := item.Trigger.Event.CardID
					if !effects.OnBattlefield(g, id) {
						return nil
					}
					if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
					d, ok := effects.DurationUntilEndOfCombat(ctx)
					if !ok {
						return nil
					}
					return effects.ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AddKeywordsMod("skulk")},
						Duration: d,
						Label:    "Tadeas, Juniper Ascendant — can't be blocked by creatures with greater power this combat",
					}.Apply(ctx)
				}),
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Tadeas, Juniper Ascendant — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
