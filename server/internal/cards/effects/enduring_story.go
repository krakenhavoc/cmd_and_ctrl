package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enduring_story.go — the card vocabulary for storied and the enduring
// story (CR 702.195, #4475, ADR 0096's 2026-10-29 amendment).
//
// The designation lives on the player (game/enduring_story.go) and the
// engine grants it from the same sweep as the city's blessing, so a card
// file declares only what it does WITH it, in the blessing's words:
//
//	PrintedKeywords: []string{game.KeywordStoried}                    // the badge; the engine reads it
//	Static: []game.StaticAbility{SelfPumpWhileEnduringStory(1, 0)}    // Ori, Keeper of Songs
//	if YouHaveAnEnduringStory(g, you) { … }                           // Balin, Loremaster
//
// `storied` is a canonical keyword, so a deck-imported card gets it from
// Scryfall with no catalog entry. Where a condition is read follows the
// card, as the blessing's does: a static is read by the layer pass (the
// engine bumps it when the designation is granted, CR 702.195c), an
// untap restriction at the untap step, a trigger doubler as the trigger
// triggers, and "if you have an enduring story" as the clause resolves.

// YouHaveAnEnduringStory is "if you have an enduring story" / "as long as
// you have an enduring story". Caller holds g.mu.
func YouHaveAnEnduringStory(g *game.Game, you uuid.UUID) bool {
	return you != uuid.Nil && g.EnduringStoryForEffect(you)
}

// WhileEnduringStory is the AppliesTo of "as long as you have an enduring
// story, <these permanents> …": the permanent under test passes `inner`
// and the SOURCE's controller has an enduring story.
func WhileEnduringStory(inner func(target *game.Card, g *game.Game, source *game.Card) bool) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return whileYouHave(YouHaveAnEnduringStory, inner)
}

// SelfWhileEnduringStory is "as long as you have an enduring story, THIS
// creature …".
func SelfWhileEnduringStory() func(target *game.Card, g *game.Game, source *game.Card) bool {
	return whileYouHave(YouHaveAnEnduringStory, selfCreature)
}

// SelfPumpWhileEnduringStory is "As long as you have an enduring story,
// this creature gets +p/+t" (Óin the Brave, Ori, Keeper of Songs): a
// layer 7c modification.
func SelfPumpWhileEnduringStory(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:     "Gets " + signedPT(power, toughness) + " while you have an enduring story.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: SelfWhileEnduringStory(),
		Apply:     pumpBy(power, toughness),
	}
}

// SelfKeywordWhileEnduringStory is "As long as you have an enduring story,
// this creature has <keyword>" (Óin's haste, Ori's vigilance): a layer 6
// grant.
func SelfKeywordWhileEnduringStory(keyword string) game.StaticAbility {
	ab := KeywordGrant(SelfWhileEnduringStory(), keyword)
	ab.Label = "Has " + keyword + " while you have an enduring story."
	return ab
}

// DoesntUntapUnlessEnduringStory is "This creature doesn't untap during
// your untap step unless you have an enduring story" (Bombur, Gentle
// Dreamer): the CR 502.3 restriction on the source, asked at each of its
// controller's untap steps, so the step after the story arrives untaps it.
func DoesntUntapUnlessEnduringStory() game.UntapStepRestriction {
	r := doesntUntapDuringYourUntapStep()
	self := r.Restricts
	r.Label = "doesn't untap during your untap step unless you have an enduring story"
	r.Restricts = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return self(target, g, source) && !YouHaveAnEnduringStory(g, source.Controller)
	}
	return r
}
