package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Capricopian — Creature — Goat Hydra {X}{G}, 0/0 (#6170):
//
//	"This creature enters with X +1/+1 counters on it.
//	 {2}: Put a +1/+1 counter on this creature, then you may reselect
//	 which player this creature is attacking. Only the player this
//	 creature is attacking may activate this ability and only during the
//	 declare attackers step. (It can't attack its controller.)"
//
// The activator rule is DefendingPlayerOnly (ADR 0106 §1 amendment
// 2026-11-25): game.MayActivate opens the row to the seat the Goat is
// attacking right now and to nobody else, its controller included, so
// a Goat that is not attacking, or attacks a planeswalker or a battle,
// has no activator. "Only during the declare attackers step" is the
// row's Condition. "You" in the effect is the activator (CR 109.5).
//
// "Reselect which player" is CR 508.7 through the engine's reselect
// prompt, narrowed to players (ReselectAttackPrompt.PlayersOnly) and
// asked of the activator, with "keep attacking" as its first option
// (the "you may"). CR 508.7c keeps the choice to the Goat's controller's
// opponents, which is the reminder text: it can't attack its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "3a5a6732-bdfa-4b64-b897-6817983dbb0e",
		Name:                       "Capricopian",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}: Put a +1/+1 counter on this creature, then you may reselect which player this creature is attacking. " +
				"Only the player this creature is attacking may activate this ability and only during the declare attackers step.",
			Cost:                effects.ManaCost("{2}"),
			DefendingPlayerOnly: true,
			Condition:           effects.DuringStep(game.StepDeclareAttackers),
			// ADR 0142: a +1/+1 counter is a pump. Answers alone is what
			// a defending-player row may declare.
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect:  capricopianGrowAndReselect,
		}},
	})
}

func capricopianGrowAndReselect(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src := ctx.Source()
	if !effects.OnBattlefield(g, src) || ctx.IsNewSourceObject(src) {
		return nil
	}
	if err := effects.NothingIfGone(g.AddCounterByForEffect(item.Controller, src, game.CounterPlusOne, 1)); err != nil {
		return err
	}
	g.QueueReselectAttackForEffect(game.ReselectAttackPrompt{
		Chooser:     item.Controller,
		Attacker:    src,
		Source:      item.SourceCardID,
		Question:    "Capricopian — reselect which player it is attacking?",
		PlayersOnly: true,
	})
	return nil
}
