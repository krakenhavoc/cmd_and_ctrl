package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_d_helpers.go — shared bodies for the Rooms of ADR 0103 PR 3
// (#1756): a "from among them" pick with a count, and a distinct-name
// count over unlocked doors.

// mayTakeUpToFromAmongThem is "you may put up to N <match> cards from
// among them into your hand" after a mill: mayTakeOneFromAmongThem's
// count-N sibling (Rickety Gazebo). Candidates are the matching cards
// of `milled` that are still in the controller's graveyard. The Then
// rebuilds its Context from the game the resolver hands back, the
// contract every queued continuation follows.
func mayTakeUpToFromAmongThem(ctx *Context, milled []uuid.UUID, max int, match CardPredicate, question string) error {
	player := ctx.Controller()
	var candidates []uuid.UUID
	for _, id := range milled {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !match(ctx.Game, player, c) {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: question,
		Cards:    candidates,
		Min:      0,
		Max:      max,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  mayTakeUpToFromMilled,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// distinctDoorNames counts the different names among the unlocked
// doors of the Rooms `player` controls (Promising Stairs).
func distinctDoorNames(g *game.Game, player uuid.UUID) int {
	seen := map[string]bool{}
	for _, n := range UnlockedDoorNamesYouControl(g, player) {
		seen[n] = true
	}
	return len(seen)
}

// mayTakeUpToFromMilled is the cards chosen going to the hand, keyed
// (#4253): the asking item rides the prompt as data.
var mayTakeUpToFromMilled = askedPick("choose-cards/rooms-d-take-up-to-from-milled", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
