package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temporal Intervention — Sorcery {2}{B} (Edge of Eternities):
//
//	"Void — This spell costs {2} less to cast if a nonland permanent left
//	 the battlefield this turn or a spell was warped this turn.
//	 Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card."
//
// The reduction is a self cost modifier (void.go, CostsLessIfVoid), priced
// at CR 601.2f against the turn as it stands when the spell is announced;
// it spends generic mana only, so the spell never costs less than {B}. The
// rest is the revealed-hand pick (ADR 0116): the whole table sees the hand,
// only a nonland card may be chosen, and a hand with none is revealed and
// nothing is discarded (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e8a2416-e311-41ac-a3cd-be1c2441d542",
		Name:         "Temporal Intervention",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfVoid(2, "This spell costs {2} less to cast if a nonland permanent left the battlefield this turn or a spell was warped this turn"),
		},
		OnResolve: TargetRevealsYouChooseDiscard(Nonland(), "nonland card"),
	})
}
