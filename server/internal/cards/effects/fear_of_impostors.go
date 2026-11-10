package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fear of Impostors — Enchantment Creature — Nightmare {1}{U}{U}, 3/2
// (Duskmourn):
//
//	"Flash
//	 When this creature enters, counter target spell. Its controller
//	 manifests dread."
//
// The spell's controller is read BEFORE the counter (Arcane Denial's note:
// countering removes the stack record the controller lives on). A spell that
// can't be countered is still a legal target: CounterTarget does nothing to
// it and its controller still manifests dread, as the rulings say. The
// manifest is the spell controller's own dread (ManifestDread{Player}),
// their look and their pick. A spell that has left the stack makes the
// target illegal, and none of it happens (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:        "628ae0e4-5d54-48a1-a24c-b9c1d862df9d",
		Name:            "Fear of Impostors",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Fear of Impostors — counter target spell; its controller manifests dread", fearOfImpostorsCounter),
				TargetSpell("target spell"),
			),
		},
	})
}

// fearOfImpostorsCounter counters the target spell, then its controller
// manifests dread.
func fearOfImpostorsCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var spell uuid.UUID
	for _, t := range ctx.LegalTargets() {
		spell = t.ID
		break
	}
	if spell == uuid.Nil {
		return nil
	}
	victim := uuid.Nil
	if v := g.StackItemForEffect(spell); v != nil {
		victim = v.Controller
	}
	if err := (CounterTarget{StackID: spell}).Apply(ctx); err != nil {
		return err
	}
	if victim == uuid.Nil {
		return nil
	}
	return ManifestDread{Player: victim}.Apply(ctx)
}
