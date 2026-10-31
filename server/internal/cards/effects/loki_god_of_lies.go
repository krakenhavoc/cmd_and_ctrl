package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loki, God of Lies — Legendary Creature — God Sorcerer Villain {1}{R}{R},
// 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever you cast a spell that targets only a single creature, gain
//	 control of that creature until end of turn. If it's your turn, untap
//	 that creature and it gains haste until end of turn."
//
// The trigger reads the spell's stack item (Muck Drubb's
// TargetsOnlyACreature: one target slot, and it names a creature on the
// battlefield). The spell is still on the stack when the trigger resolves,
// so the creature is read from it then; if the creature has left, nothing
// happens. The untap and haste depend on it being the controller's turn at
// resolution, in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f72ce67-3c17-4338-904b-26e2bf4aecdc",
		Name:         "Loki, God of Lies",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(TargetsOnlyACreature(),
				"Loki, God of Lies — gain control of the spell's only target until end of turn; on your turn untap it and it gains haste",
				lokiGodOfLiesEffect),
		},
	})
}

// lokiGodOfLiesEffect is the resolution body.
//
// Caller holds g.mu.
func lokiGodOfLiesEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spell := g.StackItemForEffect(ctx.Trigger().Event.CardID)
	if spell == nil || len(spell.Targets) != 1 || spell.Targets[0].Kind != game.TargetCard {
		return nil
	}
	target := spell.Targets[0].ID
	if !onBattlefield(g, target) {
		return nil
	}
	if err := (GainControl{
		Target:   target,
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Loki, God of Lies — gain control until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	if g.Seats[g.Turn.ActiveSeat].ID != item.Controller {
		return nil
	}
	if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   target,
		Keywords: []string{"haste"},
		Label:    "Loki, God of Lies — haste until end of turn",
	}.Apply(ctx)
}
