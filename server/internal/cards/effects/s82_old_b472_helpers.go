package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b472_helpers.go — Scars of Mirrodin, S82 older sets batch 472
// (#5577): bodies the batch's cards share. Append-only.

// OldB472AnimateThisBird is "this artifact becomes a 2/2 Bird artifact
// creature with flying until end of turn" (Glint Hawk Idol, both of its
// abilities): Faerie Conclave's animation (Cmdr2B13Animate) on the source,
// which stays an artifact and is colourless. An activation or trigger whose
// artifact has left, or come back as a new object (CR 400.7), does nothing.
func OldB472AnimateThisBird(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		return Cmdr2B13Animate(NewContext(g, item), item.SourceCardID, name+" — a 2/2 Bird artifact creature with flying until end of turn",
			2, 2, "Bird", nil, "flying")
	}
}

// OldB472WheneverYouCastAnArtifactMayPay is the Smiths' cast trigger,
// "Whenever you cast an artifact spell, you may pay {1}. If you do, …": one
// trigger per artifact spell whose resolution asks for the payment (MayPay) and
// runs the registered paid branch only once it has been made.
func OldB472WheneverYouCastAnArtifactMayPay(name, label, question string, paid game.PayUnlessElse) game.TriggeredAbility {
	return WheneverYouCast(Artifact(), name+" — "+label, func(g *game.Game, item *game.StackItem) error {
		return MayPay{Chooser: item.Controller, Cost: "{1}", Question: name + " — " + question, Then: paid}.Apply(NewContext(g, item))
	})
}

// OldB472EnchantedPermanentsControllerGetsPoison is "its controller gets a
// poison counter" for a trigger on the permanent an Aura is attached to
// (Relic Putrescence): the controller of the permanent as the trigger
// resolves, which is whoever the permanent belongs to now.
func OldB472EnchantedPermanentsControllerGetsPoison(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	if !ok {
		return nil
	}
	if p := g.PlayerByIDForEffect(c.Controller); p == nil || p.Eliminated {
		return nil
	}
	return g.AddPlayerCounterByForEffect(item.SourceCardID, c.Controller, game.CounterPoison, 1)
}

// OldB472UpkeepSacrificeUnlessDiscard is "At the beginning of your upkeep,
// sacrifice this creature unless you discard a card" (Molten-Tail Masticore):
// with no card in hand it is sacrificed at once; otherwise the controller is
// asked, and declining sacrifices it. A creature that has left, or come back
// as a new object, is not sacrificed (CR 400.7).
func OldB472UpkeepSacrificeUnlessDiscard(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil || p.Eliminated {
			return nil
		}
		if p.Hand == nil || p.Hand.Size() == 0 {
			return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
		}
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   item.Controller,
			Source:   item.SourceCardID,
			N:        1,
			UpTo:     true,
			Question: name + " — discard a card, or sacrifice it?",
			Asked:    game.NewAskedItem(item),
		}, oldB472DiscardOrSacrifice, nil)
	}
}

// oldB472DiscardOrSacrifice is the answer: a card discarded keeps the
// creature, none sacrifices it. An on-disk key: never renamed, never reused.
var oldB472DiscardOrSacrifice = game.RegisterDiscardThen("choose-cards/s82-old-b472-discard-or-sacrifice", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil || len(r.Discarded.Cards()) > 0 || !SourceIsStillThisPermanent(g, r.Item) {
		return nil
	}
	return SacrificePermanent{Target: r.Item.SourceCardID}.Apply(NewContext(g, r.Item))
})

// OldB472ArtifactsYouControl are the artifacts the item's controller
// controls, for a cost that returns one (Glint Hawk).
func OldB472ArtifactsYouControl(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsArtifact() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
