package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b791_helpers.go — Visions, batch 791 (S82 older sets, #5631).

// Old791WandMillBranch is Wand of Denial's paid branch: the card looked at is
// put into its owner's graveyard, if it is still in that library.
// An on-disk key: never renamed, never reused.
var Old791WandMillBranch = PayUnlessBranches("pay-unless/s82-old-b791-wand-of-denial-mill",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		// Carry is the card's own: the card looked at.
		if len(r.Carry) < 1 {
			return nil
		}
		id := r.Carry[0]
		if !StdB9InLibrary(ctx.Game, id) {
			return nil
		}
		return ctx.Game.PutIntoGraveyardForEffect(id)
	}, nil)

// Old791WandOfDenial is "Look at the top card of target player's library. If
// it's a nonland card, you may pay 2 life. If you do, put it into that
// player's graveyard." The look is private to the controller; the offer is made
// only for a nonland card, and a controller who can't pay 2 life isn't offered
// it (CR 118.3).
func Old791WandOfDenial(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	owner := TargetedPlayerOf(ctx)
	if owner == uuid.Nil {
		return nil
	}
	looked := g.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), owner, 1)
	if len(looked) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(looked[0])
	if !ok || card.IsLand() {
		return nil
	}
	return MayPay{
		Chooser:  ctx.Controller(),
		Pay:      LifePayment(2),
		Question: "Wand of Denial — pay 2 life to put " + card.Name + " into its owner's graveyard?",
		Then:     Old791WandMillBranch,
		Carry:    []uuid.UUID{looked[0]},
	}.Apply(ctx)
}

// Old791BroodBody is Brood of Cockroaches's delayed trigger: its controller
// loses 1 life, and the card returns from the graveyard to its owner's hand if
// it is still the same card there (CR 400.7). The life is lost either way.
// An on-disk key: never renamed, never reused.
var Old791BroodBody = game.DelayedBody("s82-old-b791/brood-of-cockroaches-return",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := NewContext(g, item)
		if err := (GainLife{Player: item.Controller, Amount: -1}).Apply(ctx); err != nil {
			return err
		}
		id := p.Object.ID
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.ObjectEpoch != p.Object.Epoch {
			return nil
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
	})

// Old791BroodOfCockroaches is "At the beginning of the next end step, you lose
// 1 life and return this card to your hand", scheduled as the dies trigger
// resolves with the card as it now is in the graveyard.
func Old791BroodOfCockroaches(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := item.SourceCardID
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  "Brood of Cockroaches — you lose 1 life and return this card to your hand",
		Body:   Old791BroodBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
}

// Old791TalruumOther is the other creature of the blocker-attacker pair the
// Talruum Champion is in, or the zero ID when it is in neither.
func Old791TalruumOther(ev game.Event, source *game.Card) uuid.UUID {
	switch {
	case ev.Target == uuid.Nil:
		return uuid.Nil
	case ev.CardID == source.InstanceID:
		return ev.Target
	case ev.Target == source.InstanceID:
		return ev.CardID
	}
	return uuid.Nil
}

// Old791TalruumChampionPairs is "Whenever this creature blocks or becomes
// blocked by a creature, that creature loses first strike until end of turn":
// one trigger for each blocker-attacker pair that names the Champion (CR
// 509.3b, 509.3d), the other creature read as the trigger resolves.
func Old791TalruumChampionPairs(label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBlock},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return Old791TalruumOther(ev, source) != uuid.Nil
		},
		Key: label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			other := Old791TalruumOther(ev, source)
			item.Params.Object = game.ObjectRef{ID: other, Epoch: g.ObjectEpochForEffect(other)}
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			obj := item.Params.Object
			if obj.ID == uuid.Nil || g.ObjectEpochForEffect(obj.ID) != obj.Epoch || !OnBattlefield(g, obj.ID) {
				return nil
			}
			return UntilEndOfTurn(NewContext(g, item), obj.ID, nil, "Talruum Champion — loses first strike",
				game.RemoveKeywordsMod("first strike"))
		},
	}
}

// Old791DefenderGetsAPoisonCounter is "defending player gets a poison counter".
func Old791DefenderGetsAPoisonCounter(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
	if g.PlayerByIDForEffect(defender) == nil {
		return nil
	}
	return g.AddPlayerCounterByForEffect(item.Controller, defender, game.CounterPoison, 1)
}

// Old791WaterspoutDjinnReturn is Waterspout Djinn's upkeep payment: sacrifice
// it unless you return an untapped Island you control to its owner's hand.
var Old791WaterspoutDjinnReturn = NewReturnLandsOrSacrifice(
	"Waterspout Djinn — return an untapped Island you control to its owner's hand (or sacrifice Waterspout Djinn)", 1,
	"choose-cards/s82-old-b791-waterspout-djinn-return-an-untapped-island", "s82-old-b791/waterspout-djinn-after-return",
	UntappedLandOfType("Island"))

// Old791UpkeepTrigger is the "At the beginning of your upkeep, sacrifice this
// creature unless you return <lands>" trigger of a returnLandsOrSacrifice. A
// Djinn that has left by resolution asks nothing (CR 400.7).
func (r *returnLandsOrSacrifice) Old791UpkeepTrigger(label string) game.TriggeredAbility {
	return AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
			return nil
		}
		return r.enters(g, item)
	})
}

// Old791FlyingCreaturesCantActivateTapAbilities is Katabatic Winds's second
// half: a creature with flying (as the layers leave it) can't have an activated
// ability with {T} in its cost activated, mana abilities included (CR 602.5).
func Old791FlyingCreaturesCantActivateTapAbilities(label string) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			if !q.Ability.TapCost || q.FromZone != game.ZoneBattlefield {
				return false
			}
			return q.Card.IsCreature() && game.HasKeyword(&q.Card, "flying")
		},
	}
}

// Old791FlyingCreaturesCantAttackOrBlock is Katabatic Winds's first half: "Creatures
// with flying can't attack or block", every player's.
func Old791FlyingCreaturesCantAttackOrBlock(label string) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature() && game.HasKeyword(target, "flying")
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= game.CantAttack | game.CantBlock
		},
	}
}

// Old791CounterUnlessPaysManaAndLife is "Counter target spell unless its
// controller pays {1} and 1 life" for the first legal spell target: the payment
// is one, mana and life together (CR 118.12).
func Old791CounterUnlessPaysManaAndLife(question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		t, ok := ctx.ClauseTarget(0)
		if !ok {
			return nil
		}
		return g.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{
			StackItem: t.ID,
			Source:    ctx.Source(),
			Cost:      "{1}",
			LifeCost:  1,
			Question:  question,
		})
	}
}
