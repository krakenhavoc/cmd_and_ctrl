package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hunter's Talent — Enchantment — Class {1}{G} (Bloomburrow, #4467):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, target creature you control deals damage
//	 equal to its power to target creature you don't control.
//	 {1}{G}: Level 2
//	 Whenever you attack, target attacking creature gets +1/+0 and gains
//	 trample until end of turn.
//	 {3}{G}: Level 3
//	 At the beginning of your end step, if you control a creature with
//	 power 4 or greater, draw a card."
//
// ADR 0071 Class machinery, as Innkeeper's Talent: LevelUp for each
// level, and each later level's ability gated with AtLevel, so it exists
// only from that level on (CR 716.2a).
//
//   - Level 1 is Bite Down as an entry trigger: two target clauses, slot
//     0 deals damage equal to its power as the trigger resolves, slot 1
//     takes it (oneSidedBite). Either one gone means no damage.
//   - Level 2 is Most Valuable Slayer's "whenever you attack" with
//     trample in place of first strike.
//   - Level 3 is an end-step trigger with an intervening "if" (CR
//     603.4): the creature with power 4 or greater is checked as the
//     step begins and again as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8a9b103-0cb8-4810-a2d7-42f6a10f80bd",
		Name:         "Hunter's Talent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Hunter's Talent — target creature you control deals damage equal to its power to target creature you don't control",
				func(g *game.Game, item *game.StackItem) error { return oneSidedBite(item, NewContext(g, item)) }),
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature you don't control", OpponentControls()),
				)),
			AtLevel(2, TriggerWithPurpose(OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Hunter's Talent — target attacking creature gets +1/+0 and gains trample until end of turn",
					pumpAndGrantFirstTarget(1, 0, "Hunter's Talent — +1/+0 and trample", "trample")),
				TargetCreature("target attacking creature", AttackingCreature()))),
				ForTargets(BenefitUntilEOT(0, 1, 0, "trample")))),
			AtLevel(3, InterveningIf(CondBoard{
				Name: "you control a creature with power 4 or greater",
				Test: func(ctx *Context) bool { return huntersTalentBigCreature(ctx.Game, ctx.Controller(), ctx.Source()) },
			}, On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return huntersTalentBigCreature(g, source.Controller, source.InstanceID)
			}), "Hunter's Talent — draw a card (you control a creature with power 4 or greater)", Do(DrawCards{N: 1})))),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{G}")),
			LevelUp(3, ManaCost("{3}{G}")),
		},
	})
}

// huntersTalentBigCreature is level 3's "if you control a creature with
// power 4 or greater".
func huntersTalentBigCreature(g *game.Game, controller, source uuid.UUID) bool {
	return ControlsAtLeast(1, MatchCreatureWithPowerAtLeast(4))(g, controller, source)
}
