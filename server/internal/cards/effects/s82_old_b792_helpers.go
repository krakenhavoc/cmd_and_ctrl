package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b792_helpers.go — Visions, batch 792 (S82 older sets, #5631).

// Old792ThreeWishesBody is Three Wishes' delayed trigger: each card that was
// exiled and is still in exile as the same card is put into its owner's
// graveyard. A card that has been played has left exile and is a new object, so
// it is left alone (CR 400.7). An on-disk key: never renamed, never reused.
var Old792ThreeWishesBody = game.DelayedBody("s82-old-b792/three-wishes-unplayed-to-graveyard",
	func(g *game.Game, _ *game.StackItem, p game.EffectParams) error {
		for _, ref := range p.Objects {
			c, ok := g.LookupCardForEffect(ref.ID)
			if !ok || c.ObjectEpoch != ref.Epoch {
				continue
			}
			if z := g.FindCardZoneForEffect(ref.ID); z == nil || z.Kind != game.ZoneExile {
				continue
			}
			if err := g.PutIntoGraveyardForEffect(ref.ID); err != nil {
				return err
			}
		}
		return nil
	})

// Old792ThreeWishes is "Exile the top three cards of your library face down.
// You may look at those cards for as long as they remain exiled. Until your next
// turn, you may play those cards. At the beginning of your next upkeep, put any
// of those cards you didn't play into your graveyard."
//
// The exile is the face-down impulse exile (only the controller may look, for
// as long as they remain exiled), its window lasts until the controller's next
// turn begins (CR 611.2b), and the delayed trigger (CR 603.7) names each card
// as the exiled object it now is.
func Old792ThreeWishes(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	me := ctx.Controller()
	p := g.PlayerByIDForEffect(me)
	if p == nil || p.Library == nil {
		return nil
	}
	n := min(3, len(p.Library.Cards))
	if n == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, p.Library.Cards[len(p.Library.Cards)-1-i].InstanceID)
	}
	window := g.UntilYourNextTurnDuration(me)
	if err := (ExileTopWithPermission{From: me, GrantTo: me, N: n, FaceDown: true, Window: &window}).Apply(ctx); err != nil {
		return err
	}
	var refs []game.ObjectRef
	for _, id := range ids {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			refs = append(refs, game.ObjectRef{ID: id, Epoch: c.ObjectEpoch})
		}
	}
	if len(refs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Three Wishes — put the cards you didn't play into your graveyard",
		Body:               Old792ThreeWishesBody,
		Params:             game.EffectParams{Objects: refs},
	}.Apply(ctx)
}

// Old792TimeAndTide is "Simultaneously, all phased-out creatures phase in and
// all creatures with phasing phase out": both sets are fixed before either
// happens, so a creature that phases in is not also phased back out (CR
// 702.26, 608.2c).
func Old792TimeAndTide(g *game.Game, item *game.StackItem) error {
	var in, out []uuid.UUID
	for _, c := range g.PhasedOutCardsForEffect() {
		if c.IsCreature() {
			in = append(in, c.InstanceID)
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && game.HasKeyword(&c, game.KeywordPhasing) {
			out = append(out, c.InstanceID)
		}
	}
	g.PhaseInForEffect(in...)
	return g.PhaseOutForEffect(item.SourceCardID, out...)
}

// Old792PlayersSkipUntapSteps is "Players skip their untap steps" as a card's
// replacement of the untap step's transition (Stasis's shape, CR 500.11): a
// pure cancel, so a second skip on the table raises no ordering prompt.
func Old792PlayersSkipUntapSteps(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventStepTransition},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			return ev.Kind == game.RepEventStepTransition && ev.StepTransitionStep == game.StepUntap
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		PureCancel: true,
		Label:      label,
	}
}

// Old792SandsOfTimeFlip is "that player simultaneously untaps each tapped
// artifact, creature, and land they control and taps each untapped artifact,
// creature, and land they control": both sets are read before either changes,
// and a permanent that is several of the types is flipped once.
func Old792SandsOfTimeFlip(g *game.Game, item *game.StackItem) error {
	player := TriggeringActor(item)
	if player == uuid.Nil {
		return nil
	}
	ctx := NewContext(g, item)
	var untap, tap []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !(c.IsArtifact() || c.IsCreature() || c.IsLand()) {
			continue
		}
		if c.Tapped {
			untap = append(untap, c.InstanceID)
		} else {
			tap = append(tap, c.InstanceID)
		}
	}
	if err := UntapEach(ctx, untap); err != nil {
		return err
	}
	return TapEach(ctx, tap)
}
