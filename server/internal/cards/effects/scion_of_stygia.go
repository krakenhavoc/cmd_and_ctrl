package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scion of Stygia — Creature — Tiefling Shaman {2}{U}, 2/1 (Adventures in
// the Forgotten Realms, #5499):
//
//	"Flash
//	 Cone of Cold — When this creature enters, choose target creature an
//	 opponent controls, then roll a d20.
//	 1—9 | Tap that creature.
//	 10—20 | Tap that creature. It doesn't untap during its controller's
//	 next untap step."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a73e1f0a-8895-4eb2-9780-94951a8f227b",
		Name:            "Scion of Stygia",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Cone of Cold — choose target creature an opponent controls, then roll a d20",
				Do(RollDice{Sides: 20, N: 1, Then: scionOfStygiaTable})),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

// scionOfStygiaTable is the results table, after the d20.
var scionOfStygiaTable = DiceTail("dice/scion-of-stygia-table", func(ctx *Context, rolls []int) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	if rolls[0] >= 10 {
		return DoesntUntapNextUntapStep{Targets: []uuid.UUID{id}, Label: "Scion of Stygia — doesn't untap during its controller's next untap step"}.Apply(ctx)
	}
	return nil
})
