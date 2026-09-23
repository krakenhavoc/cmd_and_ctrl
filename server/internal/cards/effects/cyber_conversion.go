package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyber Conversion — Instant for {U}{U}:
//
//	"Turn target creature face down. It's a 2/2 Cyberman artifact
//	 creature."
//
// The removal spell shape of #1209's primitive, and the only card in
// the ranked 6000-card roadmap that turns a permanent face down
// (batch 47, #454 — its skip line read "#656 face-down objects:
// Cyber Conversion"). No morph anywhere on it, so CR 708.7 answers
// plainly: nothing gave this permanent a way back up and nothing ever
// will. It stays a 2/2 until it leaves the battlefield.
//
// It targets ANY creature, tokens included. A token turned face down
// is not a rules problem — CR 111 has nothing to say about it, the
// token stays on the battlefield as a nameless 2/2, and CR 704.5d
// still removes it the moment it goes anywhere else.
//
// DECLARED SIMPLIFICATION — the second sentence's TYPES. CR 708.2
// lets the spell that turns a permanent face down list that object's
// characteristics, and this one lists four: 2/2, Cyberman, artifact,
// creature. The engine's face-down body is CR 708.2a's default (a 2/2
// creature with no name, no subtypes and no mana cost) and is one
// shape per face-down kind, so the power, toughness and creature type
// all land and "Cyberman" and "artifact" do not. What that costs is
// narrow and real: Shatter cannot hit it and a Cyberman lord will not
// pump it. Widening FaceDownBody into per-object listed
// characteristics is a change to the face-down OBJECT model (ADR
// 0069), not to this card — #1270, which Yedora, Grave Gardener and
// Cybership wait on too. When it lands, this caveat goes.
func init() {
	Register(Spec{
		OracleID:     "33761c1a-9848-45bf-b934-123cebab566b",
		Name:         "Cyber Conversion",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The creature is turned face down as a 2/2 as printed, but it isn't also given the Cyberman artifact types.",
		},
		Targets: TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return turnSingleTargetFaceDown(ctx.Game, item)
		},
	})
}
