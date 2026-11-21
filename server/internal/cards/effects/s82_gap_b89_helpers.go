package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB89YouCastNotFromHand is "whenever you cast a spell from anywhere
// other than your hand": a cast event by the source's controller whose
// origin zone is known and is not the hand (Osgood's Paradox).
func GapB89YouCastNotFromHand(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller &&
		ev.OldZone != "" && ev.OldZone != game.ZoneHand
}

// gapB89FirstCastNotFromHandThisTurn is the same cast, and the first such
// cast the controller made this turn. The cast tally is bumped before the
// cast event is announced (game.CastTally.NotFromHand), so the first one
// reads 1.
func GapB89FirstCastNotFromHandThisTurn(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
	return GapB89YouCastNotFromHand(ev, source, c, g) && g.CastTallyFor(source.Controller).NotFromHand == 1
}

// gapB89DemonstrateTheSpellCast is Demonstrate's body (CR 702.144a) for a
// spell that is the one just cast rather than the trigger's own source: the
// controller may copy it with new targets, and if they do an opponent they
// choose also copies it. The "you may" is the trigger's Optional.
func GapB89DemonstrateTheSpellCast(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
		return nil
	}
	ctx := NewContext(g, item)
	spell := item.Trigger.Event.CardID
	if err := (CopySpell{StackID: spell, ChooseNewTargets: true, FromLastKnown: true}).Apply(ctx); err != nil {
		return err
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Demonstrate — choose an opponent to also copy the spell",
		ThenKey:  DemonstrateOpponentCopies,
		Carry:    []uuid.UUID{spell},
	}.Apply(ctx)
}
