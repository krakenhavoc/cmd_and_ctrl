package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b04_helpers.go — the shared body of "When this creature enters,
// it connives. When it connives this way, <do something>" (Obscura
// Interceptor, Psychic Pickpocket; Streets of New Capenna, #5491).
//
// The second sentence is a reflexive triggered ability (CR 603.12) created
// once the connive is complete (CR 701.50f): it goes on the stack above the
// parent and its target is chosen as it goes there. It exists only when the
// connive happened, so a connive a replacement cancelled (N of 0) makes
// none. A conniving creature that has already left the battlefield still
// connived (CR 701.50c), so the trigger is made all the same.

// conniveThenReflexive builds the enters-trigger effect for a creature that
// connives itself and then makes the reflexive trigger `body`. `key` names
// (a lowercase slug) the registered tail and discard continuation (on-disk keys: never renamed,
// never reused) and `label` is the reflexive trigger's stack label.
func conniveThenReflexive(key, label string, body game.BodyRef) func(g *game.Game, item *game.StackItem) error {
	discarded := game.RegisterDiscardThen("choose-cards/snc-b04-"+key+"-discarded", func(g *game.Game, r game.DiscardsDone) error {
		if len(r.Carry) == 0 {
			return nil
		}
		creature := r.Carry[0]
		nonland := 0
		for _, id := range r.Discarded.By(r.Player) {
			if card, ok := g.LookupCardForEffect(id); ok && !card.IsLand() {
				nonland++
			}
		}
		if nonland > 0 && OnBattlefield(g, creature) {
			if err := g.AddCounterForEffect(creature, game.CounterPlusOne, nonland); err != nil {
				return err
			}
		}
		g.ConnivedForEffect(creature, r.Player)
		if r.Item != nil {
			g.QueueReflexiveTriggerForEffect(r.Item, game.ReflexiveTrigger{Label: label, Body: body})
		}
		return nil
	})
	tail := game.RegisterTail("snc-b04/"+key+"-connive", func(g *game.Game, o game.TailOutcome) error {
		if o.Tail.N <= 0 || len(o.Tail.Carry) == 0 {
			return nil
		}
		controller, creature := o.Tail.Player, o.Tail.Carry[0]
		for range o.Amount {
			if err := g.DrawNForEffect(controller, 1); err != nil {
				return err
			}
		}
		if err := g.DrawNForEffect(controller, o.Tail.N); err != nil {
			return err
		}
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   controller,
			Source:   creature,
			N:        o.Tail.N,
			Question: "Connive: discard a card",
			Asked:    o.Tail.Asked,
		}, discarded, []uuid.UUID{creature})
	})
	return func(g *game.Game, item *game.StackItem) error {
		creature := item.SourceCardID
		return g.ConniveForEffect(item.Controller, creature, 1, game.KeyedTail{
			Ref:    tail,
			Player: item.Controller,
			Carry:  []uuid.UUID{creature},
			Asked:  game.NewAskedItem(item),
		})
	}
}

// Obscura Interceptor: "When it connives this way, return up to one target
// spell to its owner's hand." Not a counter (Reprieve's reading): the spell
// is put into its owner's hand and nothing about "countered" applies.
var (
	obscuraInterceptorBody = game.ReflexiveBody("snc-b04/obscura-interceptor-return-spell", SimpleBody(obscuraInterceptorReturn),
		ConstTargets(func() *game.TargetSpec {
			return TargetSpell("up to one target spell").WithCount(0, 1)
		}))

	// Psychic Pickpocket: "return up to one target nonland permanent to its
	// owner's hand".
	psychicPickpocketBody = game.ReflexiveBody("snc-b04/psychic-pickpocket-return-permanent", SimpleBody(BounceEachLegalTarget),
		ConstTargets(func() *game.TargetSpec {
			return TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1)
		}))

	ObscuraInterceptorConnive = conniveThenReflexive("obscura-interceptor",
		"Obscura Interceptor — return up to one target spell to its owner's hand", obscuraInterceptorBody)
	PsychicPickpocketConnive = conniveThenReflexive("psychic-pickpocket",
		"Psychic Pickpocket — return up to one target nonland permanent to its owner's hand", psychicPickpocketBody)
)

func obscuraInterceptorReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if err := (ReturnSpellToHand{StackID: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// theEquippedCreatureConnivesForDamage is Mask of the Schemer's trigger
// body: "it connives X, where X is the amount of damage it dealt to that
// player". The creature is the one the damage event names (the equipped
// creature, as it dealt the damage) and X is the event's amount.
func TheEquippedCreatureConnivesForDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ev := ctx.Trigger().Event
	creature, ok := g.LookupCardForEffect(ev.Source)
	if !ok || ev.Amount <= 0 {
		return nil
	}
	return ConniveX{Controller: creature.Controller, Creature: ev.Source, X: ev.Amount}.Apply(ctx)
}

// conniveGrantedCreature is "it connives" for a trigger granted to a
// creature: the creature is the ability's source and its controller draws
// and discards (Security Bypass).
func ConniveGrantedCreature(g *game.Game, item *game.StackItem) error {
	return SourceConnives(NewContext(g, item))
}
