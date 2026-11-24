package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b097_helpers.go — Zendikar Rising (S82 older sets, batch 097).

// S82OldB097AnimateAttachedLand is the Zendikon animation, "Enchanted land
// is a P/T <subtype> creature with <keywords>. It's still a land." The
// land gains the Creature type and the subtype in layer 4, the keywords in
// layer 6 and the base power and toughness in layer 7b. Zendikons set no
// colour, so the land keeps its own (none).
func S82OldB097AnimateAttachedLand(subtype string, power, toughness int, keywords ...string) []game.StaticAbility {
	typed := game.StaticAbility{
		Label:                 "Enchanted land is a " + subtype + " creature. It's still a land.",
		Layer:                 game.Layer4Type,
		ContinuesAfterRemoval: true,
		AppliesTo:             AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !TypeListHasEffects(c.Types, "Creature") {
				c.Types = append(c.Types, "Creature")
			}
			if !TypeListHasEffects(c.Subtypes, subtype) {
				c.Subtypes = append(c.Subtypes, subtype)
			}
		},
	}
	out := []game.StaticAbility{typed}
	if len(keywords) > 0 {
		grant := GrantToAttached(keywords...)
		grant.ContinuesAfterRemoval = true
		out = append(out, grant)
	}
	return append(out, SetAttachedBasePT(power, toughness))
}

// S82OldB097ReturnTheDiedCardToHand is "return that card to its owner's
// hand" for a dies trigger on the enchanted permanent: the card the dies
// event named comes back from the graveyard, and a card that has left the
// graveyard by then stays where it is (CR 400.7).
func S82OldB097ReturnTheDiedCardToHand(g *game.Game, item *game.StackItem) error {
	return ReturnFromGraveyard{Target: item.Trigger.Event.CardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}

// S82OldB097AnotherCreatureWithAPlusOneCounterYouControlDied is "whenever another
// creature you control dies, if it had a +1/+1 counter on it" (Grakmaw,
// Skyclave Ravager): another creature the source's controller controlled
// as it died, with a +1/+1 counter on it as it last existed (CR 603.10a).
// The "if" is an intervening one (CR 603.4), so it is part of the trigger
// condition.
func S82OldB097AnotherCreatureWithAPlusOneCounterYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := DiedCreature(ev, g)
	return ok && LeftUnderControlOf(ev, dead) == source.Controller &&
		g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
}

// printedS82OldB097GoblinConstructToken is Relic Robber's 0/1 colorless
// Goblin Construct artifact creature: "This token can't block." and "At
// the beginning of your upkeep, this token deals 1 damage to you." The
// "you" of the upkeep ability is the token's controller. Its slug is an
// on-disk identity.
func printedS82OldB097GoblinConstructToken() TokenTemplate {
	return TokenTemplate{
		Slug: "goblin-construct",
		Card: game.Card{
			Name:      "Goblin Construct",
			TypeLine:  "Token Artifact Creature — Goblin Construct",
			Power:     0,
			Toughness: 1,
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Goblin Construct — this token deals 1 damage to you", func(g *game.Game, item *game.StackItem) error {
				return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(NewContext(g, item))
			}),
		},
		Text: "This token can't block. At the beginning of your upkeep, this token deals 1 damage to you.",
	}
}

// S82OldB097GoblinConstructToken is the Goblin Construct's card value for
// CreateToken.
func S82OldB097GoblinConstructToken() game.Card {
	return TokenFromCatalog(printedS82OldB097GoblinConstructToken)
}

// S82OldB097YouCastAKickedSpell is "whenever you cast a kicked spell"
// (Lullmage's Familiar, Risen Riptide): your own cast of a spell whose
// kicker (or multikicker) was paid, read off the spell's stack item
// (CR 702.33a). A cast that paid no kicker does not trigger it.
func S82OldB097YouCastAKickedSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && game.KickedTimesPaid(spell, g.StackItemPaidForEffect(ev.CardID).OptionalCosts) > 0
}

// S82OldB097NahiriAttachEquipmentTail is the tail of "create a 1/1 white Kor
// Warrior creature token. You may attach an Equipment you control to it": it
// asks the controller for up to one Equipment they control that can legally
// be attached to the new token (CR 701.3b) and attaches it. An on-disk key.
var S82OldB097NahiriAttachEquipmentTail = CardTail("then/nahiri-heir-of-the-ancients-attach-equipment", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	g, warrior := ctx.Game, o.IDs[0]
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != ctx.Controller() || !c.HasSubtype("Equipment") || c.IsAttachedTo(warrior) {
			continue
		}
		if g.AttachLegalForEffect(c.InstanceID, warrior) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Nahiri, Heir of the Ancients — you may attach an Equipment you control to the Kor Warrior",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		ThenKey:  s82OldB097NahiriAttachPicked,
		LiveZone: game.ZoneBattlefield,
		Carry:    []uuid.UUID{warrior},
	})
	return nil
})

// s82OldB097NahiriAttachPicked attaches the picked Equipment to the Warrior
// riding the prompt. An on-disk key: never renamed, never reused.
var s82OldB097NahiriAttachPicked = game.RegisterChooseCardsThen("choose-cards/nahiri-heir-of-the-ancients-attach-equipment", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	return g.AttachForEffect(r.Picked[0], game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]})
})

// S82OldB097EquipmentYouControl counts the Equipment a player controls.
func S82OldB097EquipmentYouControl(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasSubtype("Equipment") {
			n++
		}
	}
	return n
}

// S82OldB097PartyCostReduction is "This spell costs {1} less to cast for each
// creature in your party" (CR 700.8): the caster's party counted when the
// cost is worked out, reducing generic mana only.
func S82OldB097PartyCostReduction() game.CostModifier {
	return CostsLessEach(func(q game.CostQuery) int {
		if q.Game == nil {
			return 0
		}
		return FinB9PartySize(q.Game, q.Controller)
	}, "This spell costs {1} less to cast for each creature in your party.")
}

// S82OldB097ThisHasBasePTUntilEOT is "this creature has base power and
// toughness P/T until end of turn" as an effect: a layer 7b set that lasts
// until end of turn, so counters and pumps still apply on top. It does
// nothing to a new object if the source has left the battlefield.
func S82OldB097ThisHasBasePTUntilEOT(label string, power, toughness int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		return UntilEndOfTurn(NewContext(g, item), item.SourceCardID, nil, label,
			game.SetBasePowerMod(power), game.SetBaseToughnessMod(toughness))
	}
}
