package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b39_helpers.go — shared bodies for the S82 older-sets batch 39
// (Modern Horizons 2, #5501).

// S82OldB39NoColoredManaWasSpent is "if no colored mana was spent to cast
// it" for a spell already on the stack (Void Mirror). Colorless mana and a
// free cast count as no colored mana; a payment the engine did not see
// answers false, because the player paid something it did not record and a
// counterspell must not fire on a guess.
func S82OldB39NoColoredManaWasSpent(g *game.Game, spellID uuid.UUID) bool {
	paid := g.StackItemPaidForEffect(spellID)
	return paid.Known() && paid.ColorsSpentCount() == 0
}

// S82OldB39CounterTheSpellIfNoColoredManaWasSpent is "counter that spell"
// behind the intervening "if" (CR 603.4): the "if" is read again on
// resolution.
func S82OldB39CounterTheSpellIfNoColoredManaWasSpent(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell := item.Trigger.Event.CardID
	if !S82OldB39NoColoredManaWasSpent(g, spell) {
		return nil
	}
	return CounterTarget{StackID: spell}.Apply(NewContext(g, item))
}

// S82OldB39ControlsThreeOrMoreTokens is the condition "you control three
// or more tokens" (Sanctuary Raptor).
func S82OldB39ControlsThreeOrMoreTokens() Condition {
	return CondBoard{Name: "you control three or more tokens", Test: func(ctx *Context) bool {
		n := 0
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if c.Controller == ctx.Controller() && c.IsToken() {
				n++
			}
		}
		return n >= 3
	}}
}

// S82OldB39PutArtifactFromHandOrGraveyard is Dakkon's "you may put an
// artifact card from your hand or graveyard onto the battlefield": the
// pick over the artifact cards in both zones, the chosen card entering from
// whichever holds it when the answer arrives (Riveteers Confluence's
// shape).
func S82OldB39PutArtifactFromHandOrGraveyard(source string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil {
			return nil
		}
		var ids []uuid.UUID
		for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
			if z == nil {
				continue
			}
			for _, c := range z.Cards {
				if c.IsArtifact() && !c.IsToken() {
					ids = append(ids, c.InstanceID)
				}
			}
		}
		if len(ids) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  item.Controller,
			Source:   item.SourceCardID,
			Question: source + " — you may put an artifact card from your hand or graveyard onto the battlefield",
			Cards:    ids,
			Min:      0,
			Max:      1,
			ThenKey:  s82OldB39ArtifactPicked,
		})
		return nil
	}
}

// s82OldB39ArtifactPicked puts the picked artifact onto the battlefield from
// the zone that holds it now. An on-disk key: never renamed, never reused.
var s82OldB39ArtifactPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b39-artifact-from-hand-or-graveyard", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return nil
	}
	switch z.Kind {
	case game.ZoneHand:
		return g.PutFromHandOntoBattlefieldKeyedForEffect(id, game.HandEntryOptions{Controller: r.Chooser}, game.KeyedTail{})
	case game.ZoneGraveyard:
		return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, r.Chooser, false, game.KeyedTail{})
	}
	return nil
})
