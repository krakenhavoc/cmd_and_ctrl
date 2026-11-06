package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crimestopper Sprite — Creature — Faerie Detective {2}{U}, 2/2 (Murders
// at Karlov Manor, #5193):
//
//	"As an additional cost to cast this spell, you may collect evidence 6.
//	 (Exile cards with total mana value 6 or greater from your graveyard.)
//	 Flying
//	 When this creature enters, tap target creature. If evidence was
//	 collected, put a stun counter on it."
//
// The offer is OptionalCollectEvidence. The enters trigger is Alchemax
// Slayer-Bots' tap-and-stun with the stun conditional: whether evidence was
// collected is the cast's record on the permanent (CR 400.7d, 607.2i), read
// as the trigger goes on the stack (WhenThisEntersEvidence), so a Sprite
// put onto the battlefield without being cast only taps. It is not an
// intervening if: the trigger taps either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5e238e67-00f2-4656-bf69-0115e47321a5",
		Name:            "Crimestopper Sprite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{OptionalCollectEvidence(6)},
		Triggered: []game.TriggeredAbility{Targeting(WhenThisEntersEvidence(
			"Crimestopper Sprite — tap target creature; if evidence was collected, put a stun counter on it",
			func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if !ItemEvidenceCollected(item) {
					return nil
				}
				return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
			}), TargetCreature("target creature"))},
	})
}
