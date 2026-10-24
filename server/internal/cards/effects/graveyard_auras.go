package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_auras.go — ADR 0154 (#1780): Auras that enchant a card in a
// graveyard, and the reanimation they do as they enter (Animate Dead,
// Dance of the Dead). The rules side is game/attach_graveyard.go.
//
// Append-only, mechanic-named, per the shared-file convention.

// EnchantCardInGraveyard is "Enchant <noun> card in a graveyard" (ADR
// 0154 decision 1): the existing graveyard target clause, so the cast,
// the CR 608.2b re-check, the view's legal set and the bot enumerator
// treat it exactly as Reanimate's. `noun` is the card type the label
// names ("creature" for Animate Dead, "instant" for Spellweaver
// Volute), and the predicates are the filter that goes with it.
//
// As an Aura's Spec.Targets it is also the CR 704.5m legality check, and
// its ZoneGraveyard is what lets game.AttachForEffect attach the Aura to
// a card in a graveyard at all.
func EnchantCardInGraveyard(noun string, preds ...CardPredicate) *game.TargetSpec {
	return TargetCardInGraveyard("enchant "+noun+" card in a graveyard", preds...)
}

// EnchantCreatureCardInGraveyard is Animate Dead's and Dance of the
// Dead's "Enchant creature card in a graveyard".
func EnchantCreatureCardInGraveyard() *game.TargetSpec {
	return EnchantCardInGraveyard("creature", Creature())
}

// AnimateEnchantedCard is the enters trigger Animate Dead and Dance of
// the Dead share (ADR 0154 decision 5):
//
//	"When this Aura enters, if it's on the battlefield, it loses
//	 'enchant creature card in a graveyard' and gains 'enchant creature
//	 put onto the battlefield with this Aura.' Return enchanted creature
//	 card to the battlefield [tapped] under your control and attach this
//	 Aura to it. When this Aura leaves the battlefield, that creature's
//	 controller sacrifices it."
//
// In printed order:
//
//  1. CR 603.4: the intervening "if it's on the battlefield", re-checked
//     as it resolves. An Aura that left in response does nothing at all
//     (the first ruling): no return, no delayed trigger.
//  2. The clause is swapped, even if nothing comes back.
//  3. "Enchanted creature card" is the card the Aura is attached to in a
//     graveyard, while it is that object. None means nothing to return;
//     the swapped clause then makes the Aura illegal (CR 704.5m).
//  4. The card returns under the trigger's controller ("under your
//     control", from any graveyard), through the entry pipeline.
//  5. Once it has entered — on the resume, if the entry paused on a
//     question — the clause names it and the Aura is attached to it: a
//     new object, so a new attach (CR 701.3c). Protection from black
//     makes that attach illegal, and the state-based action takes the
//     Aura (ruling 5).
//  6. The sacrifice is scheduled whether or not step 5 attached
//     (decision 6).
type AnimateEnchantedCard struct {
	// Tapped is Dance of the Dead's "onto the battlefield tapped".
	Tapped bool
}

func (a AnimateEnchantedCard) Apply(ctx *Context) error {
	g, item := ctx.Game, ctx.Item
	if item == nil || g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	auraID := item.SourceCardID
	aura, ok := g.LookupCardForEffect(auraID)
	if !ok {
		return nil
	}
	auraRef := game.ObjectRef{ID: auraID, Epoch: aura.ObjectEpoch}
	name := aura.Name
	controller := item.Controller
	g.SwapEnchantClauseForEffect(auraID)
	card, ok := g.EnchantedGraveyardCardForEffect(auraID)
	if !ok {
		return nil
	}
	return g.ReturnFromGraveyardToBattlefieldThenForEffect(card, controller, a.Tapped,
		func(g *game.Game, entered uuid.UUID) error {
			return animateAttachAndSchedule(g, auraRef, name, controller, entered)
		})
}

// animateAttachAndSchedule is steps 5 and 6, run once the return has
// settled. `entered` is the returned permanent, or uuid.Nil when nothing
// entered. It captures plain values only (the continuation contract).
func animateAttachAndSchedule(g *game.Game, aura game.ObjectRef, name string, controller, entered uuid.UUID) error {
	var returned game.ObjectRef
	if entered != uuid.Nil && onBattlefield(g, entered) {
		if c, ok := g.LookupCardForEffect(entered); ok {
			returned = game.ObjectRef{ID: entered, Epoch: c.ObjectEpoch}
		}
	}
	auraHere := false
	if c, ok := g.LookupCardForEffect(aura.ID); ok && c.ObjectEpoch == aura.Epoch && onBattlefield(g, aura.ID) {
		auraHere = true
	}
	if returned.ID != uuid.Nil && auraHere {
		g.SetEnchantReturnedForEffect(aura.ID, returned)
		if err := g.AttachForEffect(aura.ID, game.TargetRef{Kind: game.TargetCard, ID: returned.ID}); err != nil {
			return err
		}
	}
	d := game.IndefiniteDuration()
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: aura.ID,
		SourceObject: aura,
		Label:        name + " — that creature's controller sacrifices it",
		On:           []game.EventKind{game.EventLTB},
		Condition:    thisAuraLeftCondition,
		CondParams:   game.EffectParams{Object: aura},
		Duration:     &d,
		Body:         sacrificeTheReturnedCreatureBody,
		Params:       game.EffectParams{Object: returned},
	})
	if !auraHere && returned.ID != uuid.Nil {
		// The Aura left while the return waited on a question, so its
		// "leaves the battlefield" has already happened and the trigger
		// just scheduled will never see it. In one uninterrupted
		// resolution the Aura could not have left first; the creature
		// it brought back is sacrificed as it would have been.
		return sacrificeTheReturnedCreature(g, nil, game.EffectParams{Object: returned})
	}
	return nil
}

var (
	// "When this Aura leaves the battlefield": an EventLTB of the Aura
	// CondParams.Object names. It fires the first time that card leaves,
	// which is the object it was scheduled for, since the Aura was on
	// the battlefield when it was scheduled.
	thisAuraLeftCondition = game.DelayedCondition("attach/this-aura-left", thisAuraLeft)

	// "That creature's controller sacrifices it" (ADR 0154 decision 6):
	// the permanent Params.Object names, while it is still that object
	// on the battlefield (CR 603.7c), sacrificed by its CURRENT
	// controller (CR 701.21a).
	sacrificeTheReturnedCreatureBody = game.DelayedBody("attach/sacrifice-the-returned-creature", sacrificeTheReturnedCreature)
)

func thisAuraLeft(ev game.Event, _ *game.DelayedTrigger, _ *game.Game, p game.EffectParams) bool {
	return ev.Kind == game.EventLTB && p.Object.ID != uuid.Nil && ev.CardID == p.Object.ID
}

func sacrificeTheReturnedCreature(g *game.Game, _ *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	if id == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !onBattlefield(g, id) {
		return nil
	}
	return g.SacrificePermanentForEffect(id)
}

// UpkeepOfEnchantedCreaturesController is the When of "At the beginning
// of the upkeep of enchanted creature's controller" (Dance of the Dead):
// the upkeep of whoever controls the creature this Aura is attached to
// now. An Aura enchanting nothing on the battlefield never triggers.
func UpkeepOfEnchantedCreaturesController(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	host := g.AttachedHostOf(source)
	return host != nil && ev.Actor != uuid.Nil && ev.Actor == host.Controller
}

// mayPayToUntapEnchantedCreature is "that player may pay <cost>. If the
// player does, untap that creature." "That player" is the upkeep's
// player, who pays; "that creature" is the creature the Aura enchants as
// the trigger resolves (or last enchanted, if the Aura has left), and it
// is untapped only while it is still that object on the battlefield.
func mayPayToUntapEnchantedCreature(name, cost string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item == nil || item.Trigger == nil {
			return nil
		}
		ctx := NewContext(g, item)
		aura, ok := ctx.SourcePermanent()
		if !ok || aura.AttachedTo.Kind != game.TargetCard {
			return nil
		}
		host, ok := g.LookupCardForEffect(aura.AttachedTo.ID)
		if !ok || !onBattlefield(g, host.InstanceID) {
			return nil
		}
		creature := game.ObjectRef{ID: host.InstanceID, Epoch: host.ObjectEpoch}
		return MayPay{
			Chooser:      item.Trigger.Event.Actor,
			Cost:         cost,
			Question:     name + " — pay " + cost + " to untap the enchanted creature?",
			Then:         mayPayUntapEnchantedCreaturePaid,
			CarryObjects: []game.ObjectRef{creature},
		}.Apply(ctx)
	}
}

// BecomeAnAuraAndReanimateTarget is Necromancy's enters trigger (#3156):
//
//	"When this enchantment enters, if it's on the battlefield, it becomes
//	 an Aura with 'enchant creature put onto the battlefield with
//	 Necromancy.' Put target creature card from a graveyard onto the
//	 battlefield under your control and attach this enchantment to it.
//	 When this enchantment leaves the battlefield, that creature's
//	 controller sacrifices it."
//
// The enchantment enters as a plain enchantment and "then becomes an
// Enchant Creature Aura as a triggered ability upon entering" (the
// 2005-08-01 ruling). Unlike Animate Dead's, the trigger targets the
// card, so it is chosen as the trigger goes on the stack, and a target
// gone by resolution leaves the trigger to do nothing at all (CR
// 608.2b): the enchantment stays on the battlefield and is not an Aura.
//
// In printed order:
//
//  1. CR 603.4: the intervening "if it's on the battlefield".
//  2. It becomes an Aura: the subtype is a layer-4 effect from a
//     resolved ability (CR 611.2c, CR 613.1d), pinned to this object for
//     as long as it is on the battlefield. The enchant clause is ADR
//     0154's swapped clause (Card.EnchantSwapped), which an Aura that
//     never becomes attached fails (CR 704.5m).
//  3. The target returns under the trigger's controller's control, and
//     steps 5 and 6 of AnimateEnchantedCard follow: the clause names the
//     new permanent, the enchantment is attached to it, and the delayed
//     "that creature's controller sacrifices it" is scheduled.
type BecomeAnAuraAndReanimateTarget struct{}

func (BecomeAnAuraAndReanimateTarget) Apply(ctx *Context) error {
	g, item := ctx.Game, ctx.Item
	if item == nil || g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	auraID := item.SourceCardID
	aura, ok := g.LookupCardForEffect(auraID)
	if !ok {
		return nil
	}
	targets := ctx.Targets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	auraRef := game.ObjectRef{ID: auraID, Epoch: aura.ObjectEpoch}
	name := aura.Name
	controller := item.Controller
	if err := (ScopedEffectFor{
		Target:   auraID,
		Mods:     []game.Mod{game.AddSubtypesMod("Aura")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), auraID),
		Label:    name + " — it becomes an Aura",
	}).Apply(ctx); err != nil {
		return err
	}
	g.SwapEnchantClauseForEffect(auraID)
	return g.ReturnFromGraveyardToBattlefieldThenForEffect(targets[0].ID, controller, false,
		func(g *game.Game, entered uuid.UUID) error {
			return animateAttachAndSchedule(g, auraRef, name, controller, entered)
		})
}

// mayPayUntapEnchantedCreaturePaid is the paid half of
// mayPayToUntapEnchantedCreature, keyed (#4264): the creature, with the
// epoch it had as the trigger resolved, rides the prompt, so a creature
// that left and came back is a new object and is not untapped (CR 400.7).
var mayPayUntapEnchantedCreaturePaid = PayUnlessBranches("pay-unless/untap-enchanted-creature",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Objects) == 0 {
			return nil
		}
		creature := r.Objects[0]
		c, ok := ctx.Game.LookupCardForEffect(creature.ID)
		if !ok || c.ObjectEpoch != creature.Epoch || !onBattlefield(ctx.Game, creature.ID) {
			return nil
		}
		return UntapTarget{Target: creature.ID}.Apply(ctx)
	}, nil)
