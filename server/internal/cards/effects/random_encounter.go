package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Random Encounter — Sorcery {4}{R}{R} (Final Fantasy, #4462):
//
//	"Shuffle your library, then mill four cards. Put each creature card
//	 milled this way onto the battlefield. They gain haste. At the
//	 beginning of the next end step, return those creatures to their
//	 owner's hand.
//	 Flashback {6}{R}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Three steps, each the rest of the one before as a keyed tail: the shuffle
// and a four-card mill (only the cards that really reached the graveyard
// are "milled this way"); one simultaneous entry of the creature cards among
// them (Rise of the Dread Marn's shape, so every entering creature sees the
// others enter); then haste until end of turn on the creatures that entered
// and a delayed trigger (CR 603.7) that returns them to their owner's hand
// at the beginning of the next end step. The delayed return names each
// creature as the object that entered, so one that has left the battlefield
// and come back as a new object in the meantime stays where it is
// (CR 400.7, 603.7c). The shared Flashback cast path exiles the spell as it
// leaves the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "324493fc-a205-4177-aa94-cf4672844c81",
		Name:             "Random Encounter",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{6}{R}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			if err := ctx.Game.ShuffleLibraryForEffect(player); err != nil {
				return err
			}
			return ctx.Game.MillToZoneKeyedForEffect(player, 4, game.ZoneGraveyard, game.MillUntil{},
				ctx.ItemTail(randomEncounterMilled))
		},
	})
}

// randomEncounterMilled is the mill settled: put each creature card milled
// this way onto the battlefield, together. An on-disk key.
var randomEncounterMilled = CardTail("random-encounter/milled", func(ctx *Context, o game.TailOutcome) error {
	var creatures []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			creatures = append(creatures, id)
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:  creatures,
		ThenTail: randomEncounterEntered,
	}.Apply(ctx)
})

// randomEncounterEntered is the entry settled: the creatures that entered
// gain haste until end of turn and are returned to their owner's hand at the
// beginning of the next end step.
var randomEncounterEntered = TogetherTail("random-encounter/entered", func(ctx *Context, entered []uuid.UUID, _ int) error {
	var objects []game.ObjectRef
	var cards []uuid.UUID
	for _, id := range entered {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"haste"},
			Label:    "Random Encounter — haste until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
		objects = append(objects, game.ObjectRef{ID: id, Epoch: c.ObjectEpoch})
		cards = append(cards, id)
	}
	if len(objects) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  "Random Encounter — return those creatures to their owner's hand",
		Cards:  cards,
		Body:   returnTheObjectsToHandBody,
		Params: game.EffectParams{Objects: objects},
	}.Apply(ctx)
})

// returnTheObjectsToHandBody is "return those creatures to their owner's
// hand" at a delayed trigger: each object listed, if it is still that
// object on the battlefield (CR 400.7, 603.7c).
var returnTheObjectsToHandBody = game.DelayedBody("fin/return-the-objects-to-hand", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := NewContext(g, item)
	for _, ref := range p.Objects {
		c, ok := g.LookupCardForEffect(ref.ID)
		if !ok || c.ObjectEpoch != ref.Epoch || !onBattlefield(g, ref.ID) {
			continue
		}
		if err := (BounceToHand{Target: ref.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
