package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_by_opponent_helpers_test.go — the shared table for ADR 0153
// PR 2's sixteen card tests (#3168). Not a card's test file: each card
// has its own, and they all drive the same two discards.
//
//   - dboTheyDiscard: the active seat, an opponent of the victim, casts
//     Mind Rot at the victim. The discard is an EFFECT discard whose
//     cause controller is the caster.
//   - dboIDiscard: the active seat casts Mind Rot at itself, so the
//     cause controller is the discarding player and nothing here
//     applies.
//
// Mind Rot makes the target discard two, so each test gives the victim
// the card under test and one filler, and both go.

// dboTable is the game plus the two seats every test uses.
type dboTable struct {
	g      *game.Game
	caster *game.Player // the active seat
	victim *game.Player // an opponent of the caster
}

func newDBOTable(t *testing.T) dboTable {
	t.Helper()
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() {
		caster.Hand.Cards = nil
		victim.Hand.Cards = nil
	})
	return dboTable{g: g, caster: caster, victim: victim}
}

// handCard files a card with a real oracle ID in p's hand.
func (d dboTable) handCard(p *game.Player, name, typeLine, cost, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	c := game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: oracle, Power: power, Toughness: toughness,
		Owner: p.ID, Controller: p.ID,
	}
	c.AddKnower(p.ID)
	d.g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return id
}

// filler files a vanilla card in p's hand.
func (d dboTable) filler(p *game.Player) uuid.UUID {
	return d.handCard(p, "Filler", "Sorcery", "{1}", "", 0, 0)
}

// mindRot casts Mind Rot (as the active seat) at `target` and has
// `target` discard `picks`, then settles the stack.
func (d dboTable) mindRot(t *testing.T, target *game.Player, picks ...uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}})
	passPriorityAroundTable(t, d.g)
	if discardChoiceFor(d.g, target.ID) != nil {
		answerDiscard(t, d.g, target.ID, picks...)
	}
	passPriorityAroundTable(t, d.g)
}

// theyDiscard is an opponent's Mind Rot making the victim discard the
// picks.
func (d dboTable) theyDiscard(t *testing.T, picks ...uuid.UUID) {
	t.Helper()
	d.mindRot(t, d.victim, picks...)
}

// iDiscard is the caster's own Mind Rot on itself.
func (d dboTable) iDiscard(t *testing.T, picks ...uuid.UUID) {
	t.Helper()
	d.mindRot(t, d.caster, picks...)
}

// zoneOf names where a card is now: "hand", "graveyard", "battlefield",
// "library", "exile", "stack" or "" when nowhere.
func (d dboTable) zoneOf(id uuid.UUID) string {
	var out string
	d.g.ReadSnapshot(func() {
		if z := d.g.FindCardZoneForEffect(id); z != nil {
			out = string(z.Kind)
		}
	})
	return out
}

// onBattlefield returns a copy of the battlefield permanent, if it is
// there.
func (d dboTable) onBattlefield(id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var ok bool
	d.g.ReadSnapshot(func() {
		for _, c := range d.g.Battlefield.Cards {
			if c.InstanceID == id {
				out, ok = c, true
				return
			}
		}
	})
	return out, ok
}

// countNamedUnder counts the battlefield permanents with this name
// under p's control.
func (d dboTable) countNamedUnder(p *game.Player, name string) int {
	return battlefieldCountNamed(d.g, p.ID, name)
}

// toOwnEndStep walks the turn to the active seat's end step and settles
// what triggers there.
func (d dboTable) toEndStep(t *testing.T) {
	t.Helper()
	advanceToEndStepOf(t, d.g, d.g.Turn.ActiveSeat)
	passPriorityAroundTable(t, d.g)
}

// battlefieldPermanent files a permanent with the given shape under p.
func (d dboTable) permanent(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(d.g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
}

// discardEvents counts the EventDiscardCard events the game logged for
// this card, and the zone the last one named.
type dboWatcher struct{ events []game.Event }

func (w *dboWatcher) OnEvent(_ *game.Game, ev game.Event) {
	if ev.Kind == game.EventDiscardCard {
		w.events = append(w.events, ev)
	}
}

func (d dboTable) watch() *dboWatcher {
	w := &dboWatcher{}
	d.g.RegisterListener(w)
	return w
}
