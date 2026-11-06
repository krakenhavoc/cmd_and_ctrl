package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b6_helpers.go — shared helpers for Avatar: The Last Airbender
// batch 6 (#4459).

// avatarB6BoostAndUnblockableSelf is "this creature gets +P/+T until end
// of turn and can't be blocked this turn" (Otter-Penguin). It does
// nothing to a new object if this one has left (CR 400.7); once the
// creature is blocked the restriction does not unblock it (the ruling,
// CR 509.1h).
func avatarB6BoostAndUnblockableSelf(name string, power, toughness int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !sourceIsStillThisPermanent(g, item) {
			return nil
		}
		ctx := NewContext(g, item)
		if err := (BoostUntilEOT{Target: item.SourceCardID, Power: power, Toughness: toughness,
			Label: name + " — " + signedPT(power, toughness)}).Apply(ctx); err != nil {
			return err
		}
		return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
			Label: name + " — can't be blocked"}.Apply(ctx)
	}
}

// avatarB6ShrinesYouControl is "the number of Shrines you control" (The
// Spirit Oasis), the subtype read through the layered view.
func avatarB6ShrinesYouControl(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.HasSubtype("Shrine") {
			n++
		}
	}
	return n
}

// avatarB6AnotherShrineEnteredUnderYourControl is "whenever another
// Shrine you control enters" (The Spirit Oasis).
func avatarB6AnotherShrineEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype("Shrine")
}

// avatarB6ThreeLessonsInYourGraveyard is "if there are three or more
// Lesson cards in your graveyard" (Serpent of the Pass), read as the
// cast begins (CR 601.3).
func avatarB6ThreeLessonsInYourGraveyard(g *game.Game, you uuid.UUID) bool {
	return len(b42CardsInGraveyardOf(g, you, isLessonCard)) >= 3
}
