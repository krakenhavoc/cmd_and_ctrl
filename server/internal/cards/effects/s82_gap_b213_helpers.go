package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b213_helpers.go — shared shapes for GAP batch 213 (#5948).

// GapB213NoPermanentLeft reports "no permanents left the battlefield this
// turn" (Sarevok, Deathbringer): the revolt tally, summed over every seat,
// so a land, a token or a permanent that came back all count, whoever
// controlled it. Caller must hold g.mu (every effect body and trigger
// predicate does).
func GapB213NoPermanentLeft(g *game.Game) bool {
	for _, s := range g.Seats {
		if s != nil && g.PermanentLeftThisTurn(s.ID) {
			return false
		}
	}
	return true
}

// GapB213CondNoPermanentLeft is that fact as an intervening-if Condition.
var GapB213CondNoPermanentLeft = CondBoard{
	Name: "no permanents left the battlefield this turn",
	Test: func(ctx *Context) bool { return GapB213NoPermanentLeft(ctx.Game) },
}

// GapB213CreaturesThatBlockedThisTurn is "each creature that blocked this
// turn" (Demonic Stench): the blockers named by this turn's EventBlock
// events that are creatures on the battlefield now, each once. It reads the
// turn's event log, so a creature removed from combat since still counts, and
// one that has left the battlefield does not (it is not there to be
// affected). Caller must hold g.mu.
func GapB213CreaturesThatBlockedThisTurn(g *game.Game) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventBlock || ev.CardID == uuid.Nil || out[ev.CardID] {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.IsCreature() && OnBattlefield(g, ev.CardID) {
			out[ev.CardID] = true
		}
	}
	return out
}

// GapB213ExileFromHandUntilLeaves is "each opponent exiles a card from
// their hand until this creature leaves the battlefield" (Intellect
// Devourer): each opponent, in turn order, picks a card from their own hand;
// it is exiled linked to the source object (CR 607.2a) and goes back to its
// owner's hand when that object leaves (CR 610.3). A source that has already
// left by resolution exiles nothing (CR 610.3b), and an empty hand exiles
// nothing (CR 609.3).
func GapB213ExileFromHandUntilLeaves(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok || g.UntilEventHappenedForEffect(game.UntilReturn{Leaves: ref}, item) {
		return nil
	}
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			continue
		}
		cards := make([]uuid.UUID, 0, p.Hand.Size())
		for _, c := range p.Hand.Cards {
			cards = append(cards, c.InstanceID)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    opp,
			FromPlayer: opp,
			Source:     item.SourceCardID,
			Question:   "Intellect Devourer — exile a card from your hand until it leaves the battlefield",
			Cards:      cards,
			Min:        1,
			Max:        1,
			ThenKey:    gapB213DevourerPick,
			LiveZone:   game.ZoneHand,
			Asked:      game.NewAskedItem(item),
		})
	}
	return nil
}

// gapB213DevourerPick exiles the chosen hand card linked to the Devourer. An
// on-disk key: never renamed, never reused.
var gapB213DevourerPick = AskedPick("choose-cards/gap-b213-intellect-devourer-exile", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (ExileTarget{Target: id, Linked: true, ThenTail: ctx.ItemTail(gapB213DevourerHold)}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// gapB213DevourerHold records that the card now in exile returns to its
// owner's hand when the Devourer leaves the battlefield. On-disk key.
var gapB213DevourerHold = CardTail("tail/gap-b213-intellect-devourer-hold", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	id := o.IDs[0]
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || c.ExiledWith.ID == uuid.Nil {
		return nil
	}
	ctx.Game.ScheduleUntilReturnForEffect(id, game.UntilReturn{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        "Intellect Devourer — the exiled card returns to its owner's hand when Intellect Devourer leaves the battlefield",
		Leaves:       game.ObjectRef(c.ExiledWith),
		ToHand:       true,
	})
	return nil
})

// GapB213SivrissMillThenReturnUnlessPaid is "For each opponent, you mill a
// card, then return that card from your graveyard to your hand unless that
// player pays 3 life" (Sivriss, Nightmare Speaker): per opponent in turn
// order the controller mills the top card of their library, and that
// opponent is asked to pay 3 life. A decline, or life they cannot pay
// (CR 119.4), returns the card if it is still in the graveyard. A card that
// was not milled (an empty library, or a replacement) gives that opponent
// nothing to pay for.
func GapB213SivrissMillThenReturnUnlessPaid(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	me := g.PlayerByIDForEffect(ctx.Controller())
	if me == nil || me.Library == nil {
		return nil
	}
	for _, opp := range ctx.Opponents() {
		top, err := me.Library.Top()
		if err != nil {
			return nil
		}
		id := top.InstanceID
		if err := (MillCards{N: 1}).Apply(ctx); err != nil {
			return err
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (B23PayLifeUnless{
			Chooser:  opp,
			Life:     3,
			Question: "Sivriss, Nightmare Speaker — pay 3 life? If you don't, its controller returns the milled card to their hand.",
			Else:     gapB213SivrissReturn,
			Carry:    []uuid.UUID{id},
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// gapB213SivrissReturn is the decline: the milled card goes from the
// graveyard to its owner's hand. An on-disk key: never renamed, never reused.
var gapB213SivrissReturn = PayUnlessElse("pay-unless/gap-b213-sivriss-return", func(ctx *Context, r game.PayUnlessAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	id := r.Carry[0]
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
})

// GapB213HitCountersOwnedBy counts the cards `owner` owns in exile that
// carry a hit counter ("each card they own in exile with a hit counter on
// it", Ravenloft Adventurer), one per card however many counters it has.
// Caller must hold g.mu.
func GapB213HitCountersOwnedBy(g *game.Game, owner uuid.UUID) int {
	if g.Exile == nil {
		return 0
	}
	n := 0
	for _, c := range g.Exile.Cards {
		if c.Owner == owner && c.Counters["hit"] > 0 {
			n++
		}
	}
	return n
}
