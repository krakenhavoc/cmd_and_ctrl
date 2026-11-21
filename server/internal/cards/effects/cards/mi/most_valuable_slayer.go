package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Most Valuable Slayer — Creature — Human Warrior {3}{R}, 2/4
// (Duskmourn: House of Horror, #4466):
//
//	"Whenever you attack, target attacking creature gets +1/+0 and gains
//	 first strike until end of turn."
//
// Raffine's "whenever you attack" shape: one trigger per attack
// declaration however many creatures attack (OncePerBatch), whether or
// not the Slayer is among them, aimed at an attacking creature as it goes
// on the stack. The pump and the keyword land only if that creature is
// still a legal target as the trigger resolves (pumpAndGrantFirstTarget,
// CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5bece93-afb3-4fa4-83a8-d7eaac032522",
		Name:         "Most Valuable Slayer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Most Valuable Slayer — target attacking creature gets +1/+0 and gains first strike until end of turn",
					effects.PumpAndGrantFirstTarget(1, 0, "Most Valuable Slayer — +1/+0 and first strike", "first strike")),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike"))),
		},
	})
}
