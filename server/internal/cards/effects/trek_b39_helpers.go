package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b39_helpers.go — Star Trek Commander, batch 39 (#4974): the pieces
// the batch's cards share or need registered.

// trekB39YouAndTargetOpponentEachDraw is "You and target opponent each draw
// N cards" (Commander William Riker): the controller draws first, then the
// targeted opponent. The second draw waits for any prompt the first asks
// (CR 608.2c, #4155).
func trekB39YouAndTargetOpponentEachDraw(n int) Effect {
	return Do(
		DrawCards{N: n},
		Instruction(func(ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return DrawCards{Player: t.ID, N: n}.Apply(ctx)
				}
			}
			return nil
		}),
	)
}

// --- Chief Montgomery Scott ---------------------------------------------

const trekB39ScottName = "Chief Montgomery Scott"

// trekB39ScottPut is "You may put a creature or Spacecraft card from your
// hand onto the battlefield. It enters with X additional +1/+1 counters on
// it if it's a creature and X additional charge counters on it if it's a
// Spacecraft": the optional pick from the hand, with X counted as the
// ability resolves and carried on the prompt as a number.
func trekB39ScottPut(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	candidates := handCardsMatching(g, player, Or(Creature(), HasSubtype("Spacecraft")))
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Destination: game.ChooseOntoBattlefield,
		Chooser:     player,
		Source:      ctx.Source(),
		Question:    trekB39ScottName + " — you may put a creature or Spacecraft card from your hand onto the battlefield",
		Cards:       candidates,
		Min:         0,
		Max:         1,
		ThenKey:     trekB39ScottPicked,
		LiveZone:    game.ZoneHand,
		CarryNums:   []int{socB23DifferentlyNamedLands(ctx)},
	})
	return nil
}

// trekB39ScottPicked puts the picked card onto the battlefield with its
// typed entry counters (CR 614.1c, CR 614.12: what it is as it would
// exist on the battlefield). An on-disk key: never renamed, never reused.
var trekB39ScottPicked = game.RegisterChooseCardsThen("choose-cards/chief-montgomery-scott-put", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	x := 0
	if len(r.Nums) > 0 {
		x = r.Nums[0]
	}
	opts := game.HandEntryOptions{Controller: r.Chooser}
	if x > 0 {
		opts.EntersWithCountersIfType = []game.EntryCountersIfType{
			{CardType: "creature", Counter: game.CounterPlusOne, N: x},
			{Subtype: "Spacecraft", Counter: game.CounterCharge, N: x},
		}
	}
	return g.PutFromHandOntoBattlefieldKeyedForEffect(r.Picked[0], opts, game.KeyedTail{})
})

// --- Cmdr. Troi, Outgoing Empath ----------------------------------------

const trekB39TroiName = "Cmdr. Troi, Outgoing Empath"

// trekB39OpponentWhoseTurnItIs is "target opponent whose turn it is": an
// opponent who is the active player, checked as the target is chosen and
// again as the ability resolves (CR 608.2b).
func trekB39OpponentWhoseTurnItIs() PlayerPredicate {
	return func(g *game.Game, caster uuid.UUID, p *game.Player) bool {
		return p.ID != caster && isActivePlayer(g, p.ID)
	}
}

// trekB39TroiOffer is the ability's resolution: the targeted opponent is
// offered one free cast from their hand inside this resolution (CR 608.2g),
// and trekB39TroiAfter does the rest once they have answered. The
// ability's controller rides the offer's carry, first, so the rest knows
// who copies or draws; a player is never a card on the stack, so it is
// never read as a cast card.
func trekB39TroiOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	controller := ctx.Controller()
	hand := allHandCardIDs(g, opp)
	perm := g.FreeCastFromHandPermission(opp, ctx.Source(), trekB39TroiName)
	carry := append([]uuid.UUID{controller}, hand...)
	if !g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      opp,
		Source:       ctx.Source(),
		Reason:       trekB39TroiName,
		Cards:        hand,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordHandCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Keep it in your hand",
		Then:         trekB39TroiAfter,
		Carry:        carry,
	}) {
		// No resolution to hold open, or the opponent has left: they
		// cast nothing, so you draw.
		return DrawCards{Player: controller, N: 1}.Apply(ctx)
	}
	return nil
}

// trekB39TroiAfter is "If they do, you copy that spell. You may choose new
// targets for the copy. If they don't, you draw a card." Picked is the card
// the opponent cast through the offer, still on the stack. An on-disk key:
// never renamed, never reused.
var trekB39TroiAfter = game.RegisterChooseCardsThen("resolution-cast/cmdr-troi-copy-or-draw", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	controller := r.Carry[0]
	if g.PlayerByIDForEffect(controller) == nil {
		return nil
	}
	for _, id := range r.Picked {
		if id == controller {
			continue
		}
		return g.CopySpellForEffect(id, controller, true, nil)
	}
	return g.DrawNForEffect(controller, 1)
})

// --- Collective Override -------------------------------------------------

// trekB39AttackingCreatureCount is the number of attacking creatures, every
// player's, counted as the bullet resolves.
func trekB39AttackingCreatureCount(ctx *Context) int {
	return len(AttackingCreatures(ctx.Game))
}

// --- Curzon Dax, Bound by Blood ------------------------------------------

// trekB39AnotherKlingonOrWarriorYouControlDied is "another Klingon or
// Warrior you control dies", read off the dead permanent's last-known
// characteristics (CR 603.10a).
func trekB39AnotherKlingonOrWarriorYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	if !ok || leftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	return leftAsSubtype(ev, dead, source.TextSubtype("Klingon")) || leftAsSubtype(ev, dead, source.TextSubtype("Warrior"))
}

// --- Hugh, First of Many -------------------------------------------------

// trekB39HughPaid is how many times the replicate cost a permanent granted
// was paid for the spell `spell` (CR 702.56a).
func trekB39HughPaid(g *game.Game, spell uuid.UUID) int {
	return game.GrantedOptionalTimesPaid(g.StackItemPaidForEffect(spell), game.ReplicateKey)
}

// trekB39HughIsTheGrantor reports whether `source` is the permanent whose
// replicate grant the cast took: the engine files one granted cost per key,
// from the first permanent the caster controls that grants it
// (game.OptionalCostsForCastLocked's battlefield order), so with two Hughs
// only that one copies the spell.
func trekB39HughIsTheGrantor(g *game.Game, source *game.Card, caster uuid.UUID, spell game.Card) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != caster {
			continue
		}
		for _, gr := range game.CatalogGrantedOptionalCosts(game.CatalogAbilityKey(c)) {
			if gr.Cost.Key == game.ReplicateKey && (gr.Spells == nil || gr.Spells(g, caster, spell)) {
				return c.InstanceID == source.InstanceID
			}
		}
	}
	return false
}

// trekB39HughReplicate is the replicate trigger of an artifact spell Hugh
// gave the keyword to: "When you cast this spell, if a replicate cost was
// paid for it, copy it for each time its replicate cost was paid." The
// spell had the ability as it was cast, so the trigger is Hugh's, watching
// the cast (Prismari's storm shape). The count is read as the trigger is
// built and carried on the item (an intervening if, CR 603.4), and the
// spell is named, not targeted, so one countered in response is still
// copied from last-known information (CR 608.2h). A copy of an artifact
// spell becomes a token as it resolves (CR 111.13, CR 608.3f).
func trekB39HughReplicate(label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Keyword: KeywordReplicate,
		Key:     label,
		Watches: []game.EventKind{game.EventCast},
		AppliesTo: socB32YouCastASpellThat(func(spell game.Card, source *game.Card, g *game.Game) bool {
			return spell.IsArtifact() && trekB39HughPaid(g, spell.InstanceID) > 0 &&
				trekB39HughIsTheGrantor(g, source, source.Controller, spell)
		}),
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Controller, item.Owner = ev.Actor, ev.Actor
			item.Params.Object = game.ObjectRef{ID: ev.CardID}
			item.Params.Amount = trekB39HughPaid(g, ev.CardID)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			if item.Params.Amount <= 0 {
				return nil
			}
			return CopySpell{
				StackID:          item.Params.Object.ID,
				Controller:       item.Controller,
				Count:            item.Params.Amount,
				ChooseNewTargets: true,
				FromLastKnown:    true,
			}.Apply(NewContext(g, item))
		},
	}
}

// --- Jurati, Cordial Queen -----------------------------------------------

// trekB39ArtifactCreaturesHitAPlayer is "an artifact creature you control
// dealt combat damage to a player", the dealer's types read as it dealt the
// damage.
func trekB39ArtifactCreaturesHitAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.Controller == source.Controller && dealer.IsArtifact() && dealer.IsCreature()
}

// trekB39TokenCopyOfTheTarget is "create a token that's a copy of target
// <token> you control": one copy of the first target still legal.
func trekB39TokenCopyOfTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return CreateTokenCopy{Controller: ctx.Controller(), Copy: t.ID, N: 1}.Apply(ctx)
		}
	}
	return nil
}
