package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mica, Reader of Ruins — Legendary Creature — Human Artificer {3}{R},
// 4/4 (Secrets of Strixhaven, #4456):
//
//	"Ward—Pay 3 life.
//	 Whenever you cast an instant or sorcery spell, you may sacrifice an
//	 artifact. If you do, copy that spell and you may choose new targets
//	 for the copy."
//
// Ward is ward.go's life cost (WardLife). The artifact is chosen as the
// trigger resolves (maySacrificeOne), over the artifacts you control then,
// Mica's own included if she is one; with none there is no question. The
// copy follows only an artifact that really was sacrificed, and is made from last-known
// information if the original has already left the stack (CR 608.2h), under
// your control, with the new-targets choice the copy's controller is owed
// (CR 707.10c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc90f74d-11f3-4bd4-b8a1-1cf2ffcd5b73",
		Name:         "Mica, Reader of Ruins",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(3), "Mica, Reader of Ruins — ward, pay 3 life"),
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Mica, Reader of Ruins — you may sacrifice an artifact; if you do, copy that spell",
				micaSacrificeAnArtifact),
		},
	})
}

// micaSacrificeAnArtifact asks the "you may sacrifice an artifact".
func micaSacrificeAnArtifact(g *game.Game, item *game.StackItem) error {
	return maySacrificeOne(g, item, "Mica, Reader of Ruins — sacrifice an artifact to copy that spell?", Artifact(), false, micaCopy)
}

// micaCopy is the "if you do": copy the spell the cast event named.
var micaCopy = RegisterSacrificedOne("prompt-run/mica-reader-of-ruins-copy", func(ctx *Context, _ uuid.UUID) error {
	return socB14CopyThatSpellOrAbility(ctx)
})
