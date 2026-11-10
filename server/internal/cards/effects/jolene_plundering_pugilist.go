package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fb8efd6e-b8f8-4da8-ac65-5b745c02d2bb",
		Name:         "Jolene, Plundering Pugilist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.CurrentPower() >= 4
			}, "Jolene, Plundering Pugilist — create a Treasure", Do(CreateToken{Template: TreasureToken(), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice a Treasure: Jolene deals 1 damage to any target.",
			Cost:    Plus(ManaCost("{1}{R}"), game.AbilityCost{SacrificeOther: sacrificeSpec("a Treasure", isTreasure)}),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Targets: TargetAny(),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
