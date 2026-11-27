package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b745_helpers.go — shared pieces of S82 OLD batch 745 (#5626).

// WithFewerCreatureCardsInGraveyardThanTheChooser narrows a player clause the
// triggering player chooses (OfThatPlayersChoice) to "whose graveyard has
// fewer creature cards in it than their graveyard does and is their
// opponent", they being that chooser (Oath of Ghouls,
// game.TargetSpec.FewerCreatureCardsInGraveyardThanChooser).
func WithFewerCreatureCardsInGraveyardThanTheChooser(spec *game.TargetSpec) *game.TargetSpec {
	spec.FewerCreatureCardsInGraveyardThanChooser = true
	return spec
}

// ThatPlayerMayReturnACreatureCardToHand is Oath of Ghouls' resolution: "The
// first player may return a creature card from their graveyard to their
// hand", the first player being the upkeep player named by the trigger's
// event, not the Oath's controller. They are asked for one creature card or
// none, and the card goes to its owner's hand.
func ThatPlayerMayReturnACreatureCardToHand(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		first := item.Trigger.Event.Actor
		if first == uuid.Nil || g.PlayerByIDForEffect(first) == nil {
			return nil
		}
		ctx := NewContext(g, item)
		cards := GraveyardCardIDs(ctx, first, game.Card.IsCreature)
		if len(cards) == 0 {
			return nil
		}
		_, err := g.ChooseCardsRunThenKeyedForEffect([]game.ChooseCardsPrompt{{
			Chooser:    first,
			FromPlayer: first,
			Source:     ctx.Source(),
			Question:   name + " — you may return a creature card from your graveyard to your hand",
			Cards:      cards,
			Min:        0,
			Max:        1,
			Zone:       game.ZoneGraveyard,
		}}, s82OldB745CreatureCardToHand, ctx.Source(), ctx.RunData([]uuid.UUID{first}))
		return err
	}
}

// s82OldB745CreatureCardToHand returns the chosen creature card to its
// owner's hand. An on-disk key: never renamed, never reused.
var s82OldB745CreatureCardToHand = RunThen("prompt-run/s82-old-b745-oath-of-ghouls-return", func(ctx *Context, r game.RunDone) error {
	picks := r.Picks()
	for _, first := range r.Carry {
		for _, id := range picks.By(first) {
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
})
