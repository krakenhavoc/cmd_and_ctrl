package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(NewContext(g, item))
	}
	Register(Spec{
		OracleID:     "8beaaf55-e05d-49c3-9320-c222c4d8ad12",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, gapB92ChapterLabel(name, 1, "this Saga deals 4 damage to target creature an opponent controls"),
				TargetCreature("target creature an opponent controls", OpponentControls()),
				gapB92FourDamageToTheChosenCreature),
			ChapterTrigger(2, gapB92ChapterLabel(name, 2, "exile the top card of your library; you may play that card this turn"), exileTop),
			ChapterTrigger(3, gapB92ChapterLabel(name, 3, "exile the top card of your library; you may play that card this turn"), exileTop),
			ChapterTrigger(4, gapB92ChapterLabel(name, 4, "exile the top card of your library; you may play that card this turn"), exileTop),
			ChapterTrigger(5, gapB92ChapterLabel(name, 5, "exile the top card of your library; you may play that card this turn"), exileTop),
			ChapterTrigger(6, gapB92ChapterLabel(name, 6, "add six {R}"), Do(AddMana{Produced: "{R}{R}{R}{R}{R}{R}"})),
		},
	})
}
