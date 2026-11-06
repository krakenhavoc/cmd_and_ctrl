package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f5bece93-afb3-4fa4-83a8-d7eaac032522",
		Name:         "Most Valuable Slayer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Most Valuable Slayer — target attacking creature gets +1/+0 and gains first strike until end of turn",
					pumpAndGrantFirstTarget(1, 0, "Most Valuable Slayer — +1/+0 and first strike", "first strike")),
				TargetCreature("target attacking creature", AttackingCreature()))),
				ForTargets(BenefitUntilEOT(0, 1, 0, "first strike"))),
		},
	})
}
