package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wear // Tear — split card, oracle 9842734c-1eac-4509-a731-4c22017ae586:
//
//	Wear — Instant {1}{R}: "Destroy target artifact."
//	Tear — Instant {W}: "Destroy target enchantment."
//	       Fuse (You may cast one or both halves of this card from
//	       your hand.)
//
// game.Card.CastableFaces documents the engine's own deferral for
// this layout in plain words: "split — front only for now — split
// needs fusing" (game/face.go). So only Wear, face 0, is reachable
// today; Tear and a fused cast of both halves for {1}{R}{W} are not.
// Registered with a caveat rather than left out entirely, because
// Wear alone is a complete, correct {1}{R} Shatter and the caveat
// records exactly what fusing would add.
//
// Face 0 keeps the bare oracle ID (ADR 0034 §5's composite-key
// scheme, the same one Bonecrusher Giant // Stomp registers under);
// there is no face-1 entry because CastableFaces never offers it, so
// one would be unreachable dead code.
func init() {
	Register(Spec{
		OracleID:     "9842734c-1eac-4509-a731-4c22017ae586",
		Name:         "Wear",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Only Wear (destroy target artifact, for {1}{R}) can be cast. Tear (destroy target enchantment, for {W}) and Fuse — casting both halves together for {1}{R}{W} — aren't available; the engine can only cast a split card's front half.",
		},
		Targets: TargetPermanent("target artifact", Artifact()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyChosenPermanent(ctx.Game, item)
		},
	})
}
