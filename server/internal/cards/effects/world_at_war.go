package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// World at War — Sorcery {3}{R}{R}:
//
//	"After the second main phase this turn, there's an additional combat
//	 phase followed by an additional main phase. At the beginning of
//	 that combat, untap all creatures that attacked this turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// ADR 0188 Decisions 4 and 5. "The second main phase" counts the main
// phases that occur this turn (CR 505.1b), added ones included, so the
// phases are anchored by walking the turn plan (AnchorNthMainPhase), not
// by template position:
//
//   - cast in the precombat main phase, the pair still goes after the
//     postcombat one (the second);
//   - a rebound cast in the upkeep anchors to that turn's second main
//     phase too;
//   - cast after the second main phase has ended, e.g. a second copy in
//     the third main phase this card made, it adds nothing (the
//     2010-06-15 ruling);
//   - several copies insert "after the original postcombat main phase
//     and before the newest combat phase", the newest pair first (CR 500.8).
//
// The untap is a delayed trigger bound to the combat this card added
// (DelayedTrigger.OnPhaseID), so it fires at that combat's beginning and
// at no other, and untaps every creature that attacked this turn, in any
// earlier combat (the per-object attack history). Unbound it would fire at
// the turn's regular combat, where nothing has attacked yet. Rebound is the
// engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "02a8a5cc-62ce-4ae5-aa7d-760124af7d7b",
		Name:            "World at War",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			a := ExtraCombatAndMainAfterNthMain(2)
			a.AtThatCombat = &ScheduleDelayedTrigger{
				Label: "World at War — untap all creatures that attacked this turn",
				Body:  untapAttackersAtThatCombatBody,
			}
			return a.Apply(ctx)
		},
	})
}
