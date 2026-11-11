package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s41_helpers.go — helpers of the spellslinger batch (cmdr2 batch 41).

// cmdr2S41CascadeTheSpellBody is the delayed trigger's body for "the next
// <type> spell you cast this turn has cascade" (Smoldering Stagecoach), the
// cascade sibling of stormTheSpellBody. The spell the watch fired on rides
// the item as its payload; cascade (CR 702.85a) exiles until a nonland card
// with lesser mana value than that spell, which is read as the trigger
// resolves (an X spell counts its X, CR 202.3e) from last-known information
// if the spell has left the stack (CR 608.2h).
var cmdr2S41CascadeTheSpellBody = game.SimpleDelayedBody("cascade/the-spell-just-cast", cmdr2S41CascadeTheSpell)

func cmdr2S41CascadeTheSpell(g *game.Game, item *game.StackItem) error {
	cast := NewContext(g, item).PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	mv := 0
	if spell, ok := g.LookupCardForEffect(cast[0]); ok {
		mv, _ = g.ManaValueForEffect(spell)
	}
	return g.CascadeForEffect(item.Controller, item.SourceCardID, mv)
}

// cmdr2S41DifferentInstantSorceryManaValues is the number of different mana
// values among the instant and sorcery cards in a player's graveyard (Eris,
// Roar of the Storm). Mana value is read through the layered view, and an X
// in a card's cost is 0 outside the stack (CR 202.3e).
func cmdr2S41DifferentInstantSorceryManaValues(g *game.Game, controller uuid.UUID) int {
	p := g.PlayerByIDForEffect(controller)
	if p == nil {
		return 0
	}
	seen := map[int]bool{}
	for _, c := range p.Graveyard.Cards {
		if c.IsInstant() || c.IsSorcery() {
			mv, _ := g.ManaValueForEffect(c)
			seen[mv] = true
		}
	}
	return len(seen)
}

// cmdr2S41CastThreeOrMoreSpellsThisTurn is Stella Lee's activation gate,
// "Activate only if you've cast three or more spells this turn": the
// per-turn cast tally, which counts every spell the player cast (copies are
// not cast and are not counted, CR 707.10).
func cmdr2S41CastThreeOrMoreSpellsThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	return g.CastTallyFor(controller).Total >= 3
}

// cmdr2S41EightOrMoreInstantsSorceriesInGraveyard is Octavia's cost gate,
// "if you have eight or more instant and/or sorcery cards in your graveyard",
// read at CR 601.2f off the graveyard as it stands.
func cmdr2S41EightOrMoreInstantsSorceriesInGraveyard(q game.CostQuery) bool {
	return q.Game != nil && fandanielInstantsAndSorceries(q.Game, q.Controller) >= 8
}

// cmdr2S41BaseEightEight makes the chosen creature 8/8 at base until end
// of turn (Octavia, Living Thesis). It sets base power and toughness (layer
// 7b), so counters and anthems still apply on top of it.
func cmdr2S41BaseEightEight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return untilEndOfTurn(ctx, t.ID, nil,
			"Octavia, Living Thesis — base power and toughness 8/8 until end of turn",
			game.SetBasePTMods(8, 8)...)
	}
	return nil
}

// cmdr2S41FifteenOrMoreCardsInGraveyard is Case of the Shifting Visage's
// solve condition, "there are fifteen or more cards in your graveyard",
// checked at the beginning of the end step and again as the trigger
// resolves (CR 603.4).
func cmdr2S41FifteenOrMoreCardsInGraveyard(g *game.Game, controller, _ uuid.UUID) bool {
	p := g.PlayerByIDForEffect(controller)
	return p != nil && len(p.Graveyard.Cards) >= 15
}

// cmdr2S41BirdsYouControl is Murmuration's "Birds you control" AppliesTo. The
// type is read through the source's text (CR 612.2) and the creature's
// effective subtypes, so a changeling counts.
func cmdr2S41BirdsYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller &&
		target.HasSubtype(source.TextSubtype("Bird"))
}

// cmdr2S41StormCrowForEachSpell is Murmuration's end-step effect: one 1/2
// blue Bird token named Storm Crow for each spell its controller has cast
// this turn (the per-turn cast tally, which counts every spell cast, from any
// zone, countered or not).
func cmdr2S41StormCrowForEachSpell(g *game.Game, item *game.StackItem) error {
	n := g.CastTallyFor(item.Controller).Total
	if n <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/2 blue Bird with flying named Storm Crow"), N: n}.Apply(NewContext(g, item))
}

// cmdr2S41FoundryCheapInstantOrSorcery is Forger's Foundry's spend filter:
// the mana paid for an instant or sorcery spell with mana value 3 or less.
func cmdr2S41FoundryCheapInstantOrSorcery(g *game.Game, _ uuid.UUID, paidFor game.Card) bool {
	if !paidFor.IsInstant() && !paidFor.IsSorcery() {
		return false
	}
	mv, _ := g.ManaValueForEffect(paidFor)
	return mv <= 3
}

// cmdr2S41FoundryExileAnswer is the "you may" of Forger's Foundry's spend
// trigger, asked as it resolves (CR 608.2d): on a yes, the spell the mana
// paid for (Carry[0]) is marked to be exiled, linked to the Foundry, instead
// of going to its owner's graveyard as it resolves. A spell that has left the
// stack, or that is a copy, is not marked, and a Foundry that has left the
// battlefield offers no link (the spell goes to the graveyard).
var cmdr2S41FoundryExileAnswer = MayYes("cmdr2-s41/forgers-foundry-exile", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	ctx.Game.MarkSpellExilesLinkedForEffect(r.Carry[0], game.PermissionCardRef(ref))
	return nil
})

// cmdr2S41FoundryExileTrigger is the spend trigger's effect: ask whether to
// exile the spell the mana paid for (Payload[0]).
func cmdr2S41FoundryExileTrigger(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cards := ctx.PayloadCards()
	if len(cards) == 0 {
		return nil
	}
	return MayChoice{
		Question: "Forger's Foundry — exile that spell instead of putting it into its owner's graveyard as it resolves?",
		YesLabel: "Exile it",
		NoLabel:  "Put it into the graveyard",
		Then:     cmdr2S41FoundryExileAnswer,
		Carry:    []uuid.UUID{cards[0]},
	}.Apply(ctx)
}

// cmdr2S41FoundryCastExiled is Forger's Foundry's second ability: the
// controller may cast any number of spells from among the cards exiled with
// it without paying their mana costs. One free, flash-speed, cast-only
// permission over each nonland card linked to the Foundry (CR 607.2a), kept
// until its holder passes priority, that a cast does not use up (AnyNumber);
// the offer is the resolution_cast prompt (CR 608.2g). An {X} in a cast
// spell's cost is 0 (CR 107.3b). A land among them cannot be cast and is left
// where it is.
func cmdr2S41FoundryCastExiled(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	controller := ctx.Controller()
	perm := game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		AnyNumber:   true,
		Source:      ctx.Source(),
		SourceName:  "Forger's Foundry",
		Label:       "Cast any number of spells from among the cards exiled with Forger's Foundry without paying their mana costs",
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsLand() {
			continue
		}
		ids = append(ids, c.InstanceID)
	}
	var granted []uuid.UUID
	for _, id := range ids {
		if g.GrantCastPermissionOverCardForEffect(id, perm) {
			granted = append(granted, id)
		}
	}
	if len(granted) == 0 {
		return nil
	}
	offerResolutionCastOver(g, controller, ctx.Source(), "Forger's Foundry", game.MayCastKeywordFreeCast, declineLeaveInExile, granted...)
	return nil
}

// cmdr2S41FeatherTargetsOnlyThis is Feather's trigger condition: its
// controller cast a noncreature spell that targets only Feather ("only" is
// every target, in each instance of the word, being Feather: a spell that
// names Feather twice qualifies, one that names anything else does not).
func cmdr2S41FeatherTargetsOnlyThis(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || spell.IsCreature() {
		return false
	}
	sole, ok := g.StackItemSoleTargetForEffect(ev.CardID)
	return ok && sole.Kind == game.TargetCard && sole.ID == source.InstanceID
}

// cmdr2S41FeatherCandidates is every other creature, whoever controls it,
// that a copy of the spell (controlled by its caster) could target in
// Feather's place: hexproof from the caster, protection from the spell, a
// clause that excludes it, or a failed target of a multi-target spell rules
// a creature out (game.SpellCouldTargetInsteadForEffect).
func cmdr2S41FeatherCandidates(g *game.Game, spell, caster, feather uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == feather || !c.IsCreature() {
			continue
		}
		if g.SpellCouldTargetInsteadForEffect(spell, caster, feather, c.InstanceID) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// cmdr2S41FeatherOffer is the trigger's effect: if the spell could target
// any other creature, its caster may pay {2} for each creature they will
// copy it for, up to the number of such creatures. The number is chosen first
// and the creatures after, which is the same choice (CR 608.2d) with the
// payment known.
func cmdr2S41FeatherOffer(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell, caster := item.Trigger.Event.CardID, item.Trigger.Event.Actor
	n := len(cmdr2S41FeatherCandidates(g, spell, caster, item.SourceCardID))
	if n == 0 {
		return nil
	}
	return MayPayX{
		Symbol: "{2}",
		Max:    n,
		Label:  "Feather, Radiant Arbiter",
		Buys:   "copy that spell for X other creatures",
		Unit:   game.PayAmountOther,
		Then:   cmdr2S41FeatherPaid,
		Carry:  []uuid.UUID{spell},
	}.Apply(NewContext(g, item))
}

// cmdr2S41FeatherPaid is the paid half: for each of the creatures chosen,
// copy the spell with that creature as the target of every target that was
// Feather. With as many creatures as could be chosen, nothing is asked.
var cmdr2S41FeatherPaid = PayUnlessBranches("pay-unless/cmdr2-s41-feather-radiant-arbiter",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 || len(r.Carry) == 0 {
			return nil
		}
		spell, feather := r.Carry[0], ctx.Source()
		cands := cmdr2S41FeatherCandidates(ctx.Game, spell, r.Chooser, feather)
		if len(cands) == 0 {
			return nil
		}
		if r.Number >= len(cands) {
			return cmdr2S41FeatherCopies(ctx.Game, spell, r.Chooser, feather, cands)
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  r.Chooser,
			Source:   feather,
			Question: "Feather, Radiant Arbiter — choose the creatures to copy that spell for (the copies go on the stack in the order chosen)",
			Cards:    cands,
			Min:      r.Number,
			Max:      r.Number,
			ThenKey:  cmdr2S41FeatherChosen,
			Carry:    []uuid.UUID{spell, feather},
		})
		return nil
	}, nil)

var cmdr2S41FeatherChosen = game.RegisterChooseCardsThen("choose-cards/cmdr2-s41-feather-radiant-arbiter", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) < 2 || !r.Answered {
		return nil
	}
	return cmdr2S41FeatherCopies(g, r.Carry[0], r.Chooser, r.Carry[1], r.Picked)
})

// cmdr2S41FeatherCopies makes one copy of the spell for each creature, in
// order. A creature that is no longer a legal target by now gets none.
func cmdr2S41FeatherCopies(g *game.Game, spell, caster, feather uuid.UUID, creatures []uuid.UUID) error {
	for _, id := range creatures {
		g.CopySpellRetargetedForEffect(spell, caster, feather, id)
	}
	return nil
}
