package game

import (
	"fmt"

	"github.com/google/uuid"
)

// any_player_pays_each.go — "for each <object>, <effect> unless any
// player pays <N life>" (ADR 0187 §2, #3572): one offer for the whole
// set of objects, not one prompt per object. Cleansing's "For each
// land, destroy that land unless any player pays 1 life" is the card.
//
// THE SHAPE IS CR 101.4 LITERALLY. Every player still in the game, in
// APNAP order from the ACTIVE player, chooses which objects they will
// pay for, knowing what the earlier players chose (CR 101.4b); then all
// the payments are made. Nothing is paid until the last player has
// chosen. Two players may choose the same object and each pays for it:
// ADR 0150 decision 1's "the first payer ends the offer" does not hold
// here (owner decision 1). An object is saved if anyone paid for it.
//
// PHASE 1, CHOOSING. Each player's leg is an ordinary choose_cards
// prompt (min 0), collected one seat at a time by ADR 0175's ordered
// collector (ChoosePermanentsInOrderThenForEffect), so the kind's gate,
// its enumerator, its dock, its departure row and the restore census
// all treat it as they already do. The collector supplies what this
// needs: one seat at a time, a departed seat skipped, picks bound to
// the object by ObjectEpoch, and the earlier seats' picks carried on
// each later prompt (PendingChoice.EarlierPermanentChoicesForEffect).
// A leg is FromPlayer == Chooser, which is what stops the departure
// table reassigning it (departedChoiceObjectLocked needs FromPlayer !=
// Chooser): a chooser who leaves pays for nothing (CR 800.4f) and the
// next seat is asked. Each prompt carries the offer's plain data
// (PendingChoice.AnyPlayerOffer) and the chooser's own suggested set.
//
// PHASE 2, PAYING. Each seat, in the same order, pays EachLife once per
// object it chose, as separate payments (owner decision 2): k payments
// are k life-loss events, so a "whenever you lose life" trigger sees
// each one. They are made inside one locked step with no priority
// between them, which is "simultaneously" for every rule that can tell.
// Then OnUnsaved gets the objects nobody paid for.
//
// ONE DOOR. A per-object offer built from bare choose_cards prompts,
// outside this door and the collector, would lose the departure rule,
// the epoch check, the earlier choices and the offer data the bot and
// the view read. Cards go through QueueAnyPlayerPaysForEachForEffect.

// AnyPlayerPaysForEach is the whole of a per-object any-player offer.
type AnyPlayerPaysForEach struct {
	// Source is the card whose effect asks; Controller is that
	// effect's controller ("you" on the card).
	Source, Controller uuid.UUID
	// Objects are the battlefield objects the effect applies to,
	// frozen as the effect is applied (CR 608.2h). An object that is
	// not on the battlefield now is not offered.
	Objects []uuid.UUID
	// EachLife is the life one object costs. It is the only payment
	// this door takes; zero or less is a programming error.
	EachLife int
	// Noun names one object ("land"), for the words; the plural is
	// that with an "s".
	Noun string
	// Question is the header each prompt shows. Empty reads
	// "Choose which <nouns> to pay <N> life for".
	Question string
	// OnUnsaved gets the objects nobody paid for that are still the
	// same objects (ObjectEpoch-checked), and runs exactly once, after
	// every payment. Optional.
	OnUnsaved func(g *Game, unsaved []uuid.UUID) error
}

// QueueAnyPlayerPaysForEachForEffect runs the offer described at the
// top of this file. With nobody to ask, or nobody able to pay, OnUnsaved
// runs at once on every object it was given.
//
// Caller must hold g.mu.
func (g *Game) QueueAnyPlayerPaysForEachForEffect(p AnyPlayerPaysForEach) error {
	// Freeze the objects with the epoch each has now, so a land that
	// leaves and comes back during the picks is a different object.
	var objects []uuid.UUID
	epochs := make(map[uuid.UUID]int, len(p.Objects))
	for _, id := range p.Objects {
		if _, dup := epochs[id]; dup {
			continue
		}
		if c := findBattlefieldCard(g, id); c != nil {
			epochs[id] = c.ObjectEpoch
			objects = append(objects, id)
		}
	}
	unsaved := func(g *Game, saved map[uuid.UUID]bool) error {
		if p.OnUnsaved == nil {
			return nil
		}
		var out []uuid.UUID
		for _, id := range objects {
			if saved[id] {
				continue
			}
			if c := findBattlefieldCard(g, id); c != nil && c.ObjectEpoch == epochs[id] {
				out = append(out, id)
			}
		}
		return p.OnUnsaved(g, out)
	}
	if p.EachLife <= 0 {
		// A programming error, surfaced once; the card still does
		// something, as QueueAnyPlayerPaysForEffect does with an
		// unparseable cost.
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			Source:   p.Source,
			ErrorMsg: fmt.Sprintf("any player pays for each: non-positive life payment %d", p.EachLife),
		})
		return unsaved(g, nil)
	}
	if len(objects) == 0 {
		return unsaved(g, nil)
	}

	words := fmt.Sprintf("%d life per %s", p.EachLife, p.Noun)
	question := p.Question
	if question == "" {
		question = fmt.Sprintf("Choose which %ss to pay %d life for", p.Noun, p.EachLife)
	}
	order := g.votersFromLocked(g.activePlayerIDLocked())
	offer := &AnyPlayerOffer{Controller: p.Controller, Order: order, EachLife: p.EachLife}
	var legs []ChooseCardsPrompt
	for _, who := range order {
		// The most payments this seat can make one after another.
		// CR 119.4 applies to each payment and CR 119.8 to all of
		// them. No life moves before the payment phase, so the figure
		// still holds when this seat answers.
		affordable := 0
		if pl := g.playerByIDLocked(who); pl != nil && g.CanPayLifeLocked(pl, p.EachLife) {
			affordable = min(pl.Life/p.EachLife, len(objects))
		}
		if affordable == 0 {
			// A prompt whose only answer is "none" is a wasted click.
			g.EmitEvent(Event{Kind: EventAnyPlayerDeclined, Actor: who, Source: p.Source, Label: words})
			offer.Declined = append(offer.Declined, who)
			continue
		}
		legs = append(legs, ChooseCardsPrompt{
			Chooser:            who,
			FromPlayer:         who,
			Source:             p.Source,
			Question:           question,
			Cards:              objects,
			Min:                0,
			Max:                affordable,
			Zone:               ZoneBattlefield,
			anyPlayerOffer:     offer,
			suggestUnchosenOwn: true,
		})
	}

	return g.ChoosePermanentsInOrderThenForEffect(legs, func(g *Game, picks PromptedPicks) error {
		saved := make(map[uuid.UUID]bool)
		for _, entry := range picks {
			paid := 0
			for _, id := range entry.Cards {
				pl := g.playerByIDLocked(entry.Seat)
				// CR 800.4f: a seat that has left pays nothing. CR
				// 118.3 and 119.4: nor one that can't pay now.
				if pl == nil || pl.Eliminated || !g.CanPayLifeLocked(pl, p.EachLife) {
					continue
				}
				if err := g.payLifeAsCostLocked(p.Source, entry.Seat, p.EachLife); err != nil {
					continue
				}
				paid++
				saved[id] = true
			}
			if paid == 0 {
				g.EmitEvent(Event{Kind: EventAnyPlayerDeclined, Actor: entry.Seat, Source: p.Source, Label: words})
				continue
			}
			noun := p.Noun
			if paid != 1 {
				noun += "s"
			}
			g.EmitEvent(Event{
				Kind: EventAnyPlayerPaid, Actor: entry.Seat, Source: p.Source,
				Label: fmt.Sprintf("%d life (%d %s)", paid*p.EachLife, paid, noun),
			})
		}
		return unsaved(g, saved)
	})
}
