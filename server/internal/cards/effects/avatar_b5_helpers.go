package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b5_helpers.go — shared helpers for Avatar: The Last Airbender
// batch 5 (#4459).

// avatarB5CondThisIsTapped is the resolution half of "if [this
// permanent] is tapped" (Katara, Bending Prodigy), the intervening "if"
// rechecked as the ability resolves (CR 603.4). A source that has left
// the battlefield is not tapped: there is nothing to put a counter on.
// The trigger-time half is thisPermanentIsTapped. It is a fact about the
// source object, so it declares a source read.
type avatarB5CondThisIsTapped struct{}

func (avatarB5CondThisIsTapped) Holds(ctx *Context) bool {
	info, ok := ctx.SourcePermanent()
	return ok && !info.Left && info.Tapped
}

func (avatarB5CondThisIsTapped) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondSource, 0), true
}

// avatarB5LessonCardsInYourGraveyard is "the number of Lesson cards in
// your graveyard" (Master Pakku), counted as the ability resolves.
func avatarB5LessonCardsInYourGraveyard(g *game.Game, you uuid.UUID) int {
	p := g.PlayerByIDForEffect(you)
	if p == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if isLessonCard(c) {
			n++
		}
	}
	return n
}
