package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Flux — Enchantment — Saga {2}{R}{R} (Doctor Who, #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after VI.)
//	 I — This Saga deals 4 damage to target creature an opponent controls.
//	 II, III, IV, V — Exile the top card of your library. You may play that card
//	 this turn.
//	 VI — Add six {R}."
//
// Chapter I is damage from the Saga (its last-known values if it has left,
// CR 608.2h). Chapters II to V are Charforger's exile with permission
// (ExileTopWithPermission): the card may be played, a land too, until the
// end of this turn. Chapter VI adds {R}{R}{R}{R}{R}{R}; the mana empties at
// the end of the step like any mana.
//
// No simplification.
func init() {
	const name = "The Flux"
	exileTop := func(g *game.Game, item *game.StackItem) error {
		return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
	}
	effects.Register(effects.Spec{
		OracleID:     "8beaaf55-e05d-49c3-9320-c222c4d8ad12",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.GapB92ChapterLabel(name, 1, "this Saga deals 4 damage to target creature an opponent controls"),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				effects.GapB92FourDamageToTheChosenCreature),
			effects.ChapterTrigger(2, effects.GapB92ChapterLabel(name, 2, "exile the top card of your library; you may play that card this turn"), exileTop),
			effects.ChapterTrigger(3, effects.GapB92ChapterLabel(name, 3, "exile the top card of your library; you may play that card this turn"), exileTop),
			effects.ChapterTrigger(4, effects.GapB92ChapterLabel(name, 4, "exile the top card of your library; you may play that card this turn"), exileTop),
			effects.ChapterTrigger(5, effects.GapB92ChapterLabel(name, 5, "exile the top card of your library; you may play that card this turn"), exileTop),
			effects.ChapterTrigger(6, effects.GapB92ChapterLabel(name, 6, "add six {R}"), effects.Do(effects.AddMana{Produced: "{R}{R}{R}{R}{R}{R}"})),
		},
	})
}
