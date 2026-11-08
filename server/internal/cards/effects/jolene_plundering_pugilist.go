package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jolene, Plundering Pugilist — Legendary Creature — Human Mercenary
// {1}{R}{G}, 4/2 (Outlaws of Thunder Junction, #4468):
//
//	"Whenever you attack with one or more creatures with power 4 or
//	 greater, create a Treasure token.
//	 {1}{R}, Sacrifice a Treasure: Jolene deals 1 damage to any target."
//
// The attack trigger is Bitter Work's power check on each declared
// attacker (b31CurrentPowerOf, layered power as attackers are declared),
// collapsed to one trigger per attack declaration (OncePerBatch): the
// text names no player, so attacking two opponents with big creatures
// still makes one Treasure. Jolene herself, at 4 power, counts. The
// activated ability sacrifices any Treasure you control as a cost (the
// Treasure Map row's sacrificeSpec) and always has a target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb8efd6e-b8f8-4da8-ac65-5b745c02d2bb",
		Name:         "Jolene, Plundering Pugilist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && b31CurrentPowerOf(g, ev.CardID) >= 4
			}, "Jolene, Plundering Pugilist — create a Treasure (you attacked with a creature with power 4 or greater)",
				Do(CreateToken{Template: TreasureToken(), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice a Treasure: Jolene deals 1 damage to any target.",
			Cost:    Plus(ManaCost("{1}{R}"), game.AbilityCost{SacrificeOther: sacrificeSpec("a Treasure", isTreasure)}),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
