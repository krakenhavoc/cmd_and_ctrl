package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b673_helpers.go — S82 OLD batch 673 (Prophecy, #5617).

// s82OldB673ReturnThisBody is the delayed trigger of "return it to its owner's
// hand at end of combat" (Windscouter): it returns the delayed trigger's
// source, only if that permanent is still on the battlefield and has not left
// and come back as a new object meanwhile (CR 400.7). A registered key
// (ADR 0041 phase 3).
var s82OldB673ReturnThisBody = game.SimpleDelayedBody("s82-old-b673/return-this-to-hand-at-end-of-combat", func(g *game.Game, item *game.StackItem) error {
	if g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
})

// S82OldB673ReturnThisAtEndOfCombat is the body of "When this creature attacks
// or blocks, return it to its owner's hand at end of combat": it schedules the
// return for the beginning of this combat's end of combat step (CR 511.2).
func S82OldB673ReturnThisAtEndOfCombat(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: label,
			Body:  s82OldB673ReturnThisBody,
		}.Apply(NewContext(g, item))
	}
}

// S82OldB673PsychicTheftPick is Psychic Theft's continuation, keyed under the
// card: the picked instant or sorcery card is in exile, so its chooser may cast
// it for as long as it remains exiled, and a delayed trigger at the beginning
// of the next end step returns it to its owner's hand if it is still exiled
// (it was not cast). The key is an on-disk identity: never renamed, never
// reused.
func S82OldB673PsychicTheftPick(key, sourceName string) game.RevealedPickThen {
	return RevealedPickThen(key, func(ctx *Context, pick game.RevealedPick) error {
		g := ctx.Game
		var ids []uuid.UUID
		for _, chosen := range pick.Chosen {
			if z := g.FindCardZoneForEffect(chosen.InstanceID); z == nil || z.Kind != game.ZoneExile {
				continue
			}
			c, ok := g.LookupCardForEffect(chosen.InstanceID)
			if !ok {
				continue
			}
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{
				Player:     pick.Chooser,
				Zone:       game.ZoneExile,
				CastOnly:   true,
				Duration:   game.WhileInZoneDuration(),
				Source:     pick.Source,
				SourceName: sourceName,
				Label:      "You may cast it for as long as it remains exiled",
			}, []game.Card{c})
			ids = append(ids, c.InstanceID)
		}
		if len(ids) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label: sourceName + " — return the card to its owner's hand if you haven't cast it",
			Cards: ids,
			Body:  s82OldB673PsychicTheftReturn,
		}.Apply(ctx)
	})
}

// s82OldB673PsychicTheftReturn is the delayed half: each listed card still in
// exile (not cast) returns to its owner's hand. An on-disk key: never renamed,
// never reused.
var s82OldB673PsychicTheftReturn = game.SimpleDelayedBody("exile/s82-old-b673-psychic-theft-return", func(g *game.Game, item *game.StackItem) error {
	var left []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z != nil && z.Kind == game.ZoneExile {
			left = append(left, t.ID)
		}
	}
	if len(left) == 0 {
		return nil
	}
	return g.BounceCardsToHandKeyedForEffect(left, game.KeyedTail{})
})
