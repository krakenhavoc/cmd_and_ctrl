package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jolene, Plundering Pugilist — Legendary Creature — Human Mercenary
// {1}{R}{G}, 3/3 (Outlaws of Thunder Junction, #4468):
//
//	"Whenever you attack with one or more creatures with power 4 or
//	 greater, create a Treasure token.
//	 {1}{R}, Sacrifice a Treasure: Jolene deals 1 damage to any target."
//
// One attack declaration is one trigger however many big creatures attack
// (OncePerBatch over the engine's one-event-per-attacker); power is read
// through the layers as the attack is declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb8efd6e-b8f8-4da8-ac65-5b745c02d2bb",
		Name:         "Jolene, Plundering Pugilist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.CurrentPower() >= 4
			}, "Jolene, Plundering Pugilist — create a Treasure", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice a Treasure: Jolene deals 1 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Treasure", effects.IsTreasure)}),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
