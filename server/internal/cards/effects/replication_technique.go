package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Replication Technique — Sorcery {4}{U} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Demonstrate (When you cast this spell, you may copy it. If you do,
//	 choose an opponent to also copy it. Players may choose new targets
//	 for their copies.)
//	 Create a token that's a copy of target permanent you control."
//
// Demonstrate is the keyword's cast trigger (Demonstrate, CR 702.144a,
// first used by Creative Technique): the controller's copy goes on the
// stack, then the opponent they choose copies it too and may choose new
// targets, so the opponent's copy resolves first and makes a token copy of
// a permanent THAT player controls.
//
// The body is Relm's Sketching's "create a token that's a copy of target
// <permanent>" (TokenCopyOfSingleTarget): the first still-legal target is
// copied under the resolving item's controller, and a target gone in
// response leaves the spell with nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "163baf58-b35d-4f2b-aa53-8384b8b5800a",
		Name:         "Replication Technique",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Demonstrate()},
		Targets:      TargetPermanent("target permanent you control", YouControl()),
		OnResolve:    TokenCopyOfSingleTarget,
	})
}
