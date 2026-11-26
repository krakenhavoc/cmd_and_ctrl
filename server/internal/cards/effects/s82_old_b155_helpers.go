package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b155_helpers.go — shared bodies for the Core Set 2020 auras and
// planeswalkers of S82 set OLD, batch 155 (#5518).

// OldB155ReturnDiedCardWithCounter is Unholy Indenture's dies trigger: "return
// that card to the battlefield under your control with a +1/+1 counter on it".
// The card comes back as a new object (CR 400.7) under the Aura's controller,
// the counter on it as it enters; a card that has left the graveyard by then
// (or a token, which has no card) is "nothing to return".
func OldB155ReturnDiedCardWithCounter(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	_, err := g.ReturnFromGraveyardWithCountersForEffect(item.Trigger.Event.CardID, item.Controller, false,
		map[string]int{game.CounterPlusOne: 1})
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	return err
}

// OldB155DamageByToughness is Gauntlets of Light's layer 6 static: the
// enchanted creature assigns combat damage equal to its toughness rather than
// its power (CR 510.1c), whether or not toughness is the greater.
func OldB155DamageByToughness() game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		Label:     "Enchanted creature assigns combat damage equal to its toughness rather than its power.",
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.DamageByToughnessAlways = true
		},
	}
}

// OldB155CastFromYourGraveyardExileIt is Chandra, Acolyte of Flame's −2: you
// may cast the target instant or sorcery card from your graveyard (paying its
// mana cost); if that spell would be put into your graveyard, exile it
// instead. The cast is offered during the resolution (CR 608.2g); declining
// leaves the card in the graveyard.
func OldB155CastFromYourGraveyardExileIt(sourceName string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			z := g.FindCardZoneForEffect(t.ID)
			c, ok := g.LookupCardForEffect(t.ID)
			if !ok || z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{
				Player:                  item.Controller,
				Zone:                    game.ZoneGraveyard,
				Duration:                g.UntilEndOfTurnDuration(),
				AltCostKey:              game.AltCostKeyGrantedCost,
				Timing:                  game.TimingFlash,
				CastOnly:                true,
				ExileInsteadOfGraveyard: true,
				LapseOnPass:             game.LapseStaysInHand,
				Source:                  item.SourceCardID,
				SourceName:              sourceName,
				Label:                   sourceName + " — cast " + c.Name + " from your graveyard",
			}, []game.Card{c})
			OfferResolutionCastOver(g, item.Controller, item.SourceCardID, sourceName, game.MayCastKeywordPaidCast, DeclineLeaveInGraveyard, c.InstanceID)
			return nil
		}
		return nil
	}
}

// OldB155ElementalsHasteSacrificed is Chandra, Acolyte of Flame's second 0:
// two 1/1 red Elemental tokens that gain haste, sacrificed at the beginning of
// the next end step. The tokens are made through the keyed creation door, and
// the tail is handed exactly the ones that landed.
func OldB155ElementalsHasteSacrificed(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("1/1 red Elemental")), Count: 2}},
	}, ctx.ItemTail(oldB155ElementalsTail))
}

// oldB155ElementalsTail gives the new tokens haste and schedules their
// sacrifice. An on-disk key: never renamed, never reused.
var oldB155ElementalsTail = CardTail("then/chandra-acolyte-of-flame-elementals", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Chandra, Acolyte of Flame — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return TrekSacrificeTokensAtNextEndStep("Chandra, Acolyte of Flame — sacrifice the Elementals")(ctx, o)
})

// OldB155LoyaltyOnEachRedPlaneswalker is "Put a loyalty counter on each red
// planeswalker you control": the colour is read through the layers, a word of
// the source's text (CR 612.2).
func OldB155LoyaltyOnEachRedPlaneswalker(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	red := ctx.TextColor("R")
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsPlaneswalker() || !c.HasColorWord(red) {
			continue
		}
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterLoyalty, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// oldB155AjaniGain is the "its controller gains 2 life" half of the −3, told
// (as the tail's carry) who controlled the creature before the exile
// (CR 608.2h) and run only for a card that really reached exile. An on-disk
// key: never renamed, never reused.
var oldB155AjaniGain = CardTail("then/ajani-inspiring-leader-controller-gains", func(ctx *Context, o game.TailOutcome) error {
	controller := TailCarry0(o)
	if len(o.IDs) == 0 || controller == uuid.Nil || ctx.Game.PlayerByIDForEffect(controller) == nil {
		return nil
	}
	return GainLife{Player: controller, Amount: 2}.Apply(ctx)
})

// OldB155AjaniExile is Ajani, Inspiring Leader's −3: exile target creature; its
// controller gains 2 life. The controller is read before the exile
// (CR 608.2h) and the life is gained only if the card was really exiled.
func OldB155AjaniExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, found := g.LookupCardForEffect(id)
	if !found {
		return nil
	}
	return ExileTarget{Target: id, ThenTail: WithCarry(ctx.ItemTail(oldB155AjaniGain), c.Controller)}.Apply(ctx)
}

// OldB155CountersOnTheTarget is "Put N +1/+1 counters on up to one target
// creature" as a whole effect.
func OldB155CountersOnTheTarget(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	}
}

// OldB155AjaniPlusTwo is Ajani, Inspiring Leader's +2: you gain 2 life, then put
// two +1/+1 counters on up to one target creature.
func OldB155AjaniPlusTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GainLife{Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
}

// OldB155VivienPlusOne is Vivien, Arkbow Ranger's +1: distribute two +1/+1
// counters among up to two target creatures; they gain trample until end of
// turn.
func OldB155VivienPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := PutDividedCounters(ctx, game.CounterPlusOne); err != nil {
		return err
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Vivien, Arkbow Ranger — trample"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// oldB155IsCreatureCard is the creature-card clause of Vivien's −5.
func oldB155IsCreatureCard(c game.Card) bool { return c.IsCreature() }

// OldB155VivienWish is Vivien, Arkbow Ranger's keyed pick: a creature card you
// own from outside the game. An on-disk key: never renamed, never reused.
var OldB155VivienWish = RegisterOutsideTake("choose-cards/vivien-arkbow-ranger-creature", oldB155IsCreatureCard, true)

// OldB155MuYanlingPlusOne is Mu Yanling, Celestial Wind's +1: until your next
// turn, up to one target creature gets -5/-0.
func OldB155MuYanlingPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return ScopedEffectFor{
		Target:   t.ID,
		Mods:     []game.Mod{game.ModifyPTMod(-5, 0)},
		Duration: DurationUntilYourNextTurn(ctx, item.Controller),
		Label:    "Mu Yanling, Celestial Wind — -5/-0 until your next turn",
	}.Apply(ctx)
}
