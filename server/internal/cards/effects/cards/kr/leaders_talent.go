package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leader's Talent — {1}{W} Enchantment — Class (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 Whenever you attack, put a +1/+1 counter on target attacking
//	 creature.
//	 {2}{W}: Level 2
//	 Whenever a creature you control leaves the battlefield, if it had a
//	 counter on it, you gain 2 life.
//	 {3}{W}: Level 3
//	 Whenever you cast a spell, put a +1/+1 counter on each creature you
//	 control."
//
// ADR 0071 Class machinery (Innkeeper's Talent): LevelUp for each level,
// and the level-2 and level-3 triggers gated with AtLevel (CR 716.2a).
//
//   - Level 1 is Araña, Heart of the Spider's "whenever you attack": once
//     per declaration (OncePerBatch), the attacking creature chosen as the
//     trigger goes on the stack.
//   - Level 2 is Resourceful Defense's leave trigger narrowed to a
//     creature: any exit, the creature and its controller read as it last
//     existed (CR 603.10a), and "if it had a counter on it" (any kind) is
//     the counters the engine recorded as it left. The intervening if
//     (CR 603.4) reads the same record at resolution, which cannot change.
//   - Level 3 is Zuri, Warrior of Wakanda's counter on each creature you
//     control, on any spell you cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "134430b7-d566-485a-bd5c-7db8a285f4ca",
		Name:         "Leader's Talent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Leader's Talent — put a +1/+1 counter on target attacking creature",
					effects.Do(effects.Instruction(effects.CounterOnFirstCreatureTarget))),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()))),
			effects.AtLevel(2, effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.B14PermanentYouControlLeft(ev, source, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.LeftAsType(ev, c, "creature")
			}, "Leader's Talent — you gain 2 life", effects.Do(effects.GainLife{Amount: 2}))),
			effects.AtLevel(3, effects.WheneverYouCast(nil, "Leader's Talent — put a +1/+1 counter on each creature you control",
				effects.B11PutCounterOnEachCreatureYouControl)),
		},
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{2}{W}")),
			effects.LevelUp(3, effects.ManaCost("{3}{W}")),
		},
	})
}
