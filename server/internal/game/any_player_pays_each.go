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
// prompt (min 0), collected one seat at a time with ADR 0175's ordered
// collector's rules (any_player_each_keyed.go), so the kind's gate,
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
// Then the registered consequence gets the objects nobody paid for.
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
	// Then is the REGISTERED consequence (RegisterAnyPlayerPaysEach,
	// #4264): it gets the objects nobody paid for that are still the same
	// objects (ObjectEpoch-checked), once, after every payment. Carry is
	// the plain data it is handed. Every leg of the offer holds no closure,
	// so a table waiting on any seat's pick is a restore point. Required:
	// the closure form, OnUnsaved, is gone (#4361 S7b).
	Then  AnyPlayerEachThen
	Carry []uuid.UUID
	// Asked is the asking stack item as data (NewAskedItem), handed to a
	// keyed consequence as AnyPlayerEachAnswer.Item. Nil when it reads
	// none.
	Asked *AskedItem
}

// QueueAnyPlayerPaysForEachForEffect runs the offer described at the
// top of this file. With nobody to ask, or nobody able to pay, the
// consequence runs at once on every object it was given.
//
// Caller must hold g.mu.
func (g *Game) QueueAnyPlayerPaysForEachForEffect(p AnyPlayerPaysForEach) error {
	if p.Then.key == "" {
		return fmt.Errorf("%w: per-object any-player offer with no registered consequence (RegisterAnyPlayerPaysEach)", ErrInvalidParam)
	}
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
	return g.queueAnyPlayerPaysForEachKeyedLocked(p, objects, epochs)
}
