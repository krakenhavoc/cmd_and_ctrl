package effectstest

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
type DboTable struct {
	G      *game.Game
	Caster *game.Player // the active seat
	Victim *game.Player // an opponent of the caster
}

func NewDBOTable(t *testing.T) DboTable {
	t.Helper()
	g := NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() {
		caster.Hand.Cards = nil
		victim.Hand.Cards = nil
	})
	return DboTable{G: g, Caster: caster, Victim: victim}
}

// handCard files a card with a real oracle ID in p's hand.
func (d DboTable) HandCard(p *game.Player, name, typeLine, cost, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	c := game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: oracle, Power: power, Toughness: toughness,
		Owner: p.ID, Controller: p.ID,
	}
	c.AddKnower(p.ID)
	d.G.WithWriteLock(func() { p.Hand.PushTop(c) })
	return id
}

// filler files a vanilla card in p's hand.
func (d DboTable) Filler(p *game.Player) uuid.UUID {
	return d.HandCard(p, "Filler", "Sorcery", "{1}", "", 0, 0)
}

// mindRot casts Mind Rot (as the active seat) at `target` and has
// `target` discard `picks`, then settles the stack.
func (d DboTable) mindRot(t *testing.T, target *game.Player, picks ...uuid.UUID) {
	t.Helper()
	CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", MindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}})
	PassPriorityAroundTable(t, d.G)
	if DiscardChoiceFor(d.G, target.ID) != nil {
		AnswerDiscard(t, d.G, target.ID, picks...)
	}
	PassPriorityAroundTable(t, d.G)
}

// theyDiscard is an opponent's Mind Rot making the victim discard the
// picks.
func (d DboTable) TheyDiscard(t *testing.T, picks ...uuid.UUID) {
	t.Helper()
	d.mindRot(t, d.Victim, picks...)
}

// iDiscard is the caster's own Mind Rot on itself.
func (d DboTable) IDiscard(t *testing.T, picks ...uuid.UUID) {
	t.Helper()
	d.mindRot(t, d.Caster, picks...)
}

// zoneOf names where a card is now: "hand", "graveyard", "battlefield",
// "library", "exile", "stack" or "" when nowhere.
func (d DboTable) ZoneOf(id uuid.UUID) string {
	var out string
	d.G.ReadSnapshot(func() {
		if z := d.G.FindCardZoneForEffect(id); z != nil {
			out = string(z.Kind)
		}
	})
	return out
}

// onBattlefield returns a copy of the battlefield permanent, if it is
// there.
func (d DboTable) OnBattlefield(id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var ok bool
	d.G.ReadSnapshot(func() {
		for _, c := range d.G.Battlefield.Cards {
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
func (d DboTable) CountNamedUnder(p *game.Player, name string) int {
	return BattlefieldCountNamed(d.G, p.ID, name)
}

// toOwnEndStep walks the turn to the active seat's end step and settles
// what triggers there.
func (d DboTable) ToEndStep(t *testing.T) {
	t.Helper()
	AdvanceToEndStepOf(t, d.G, d.G.Turn.ActiveSeat)
	PassPriorityAroundTable(t, d.G)
}

// battlefieldPermanent files a permanent with the given shape under p.
func (d DboTable) Permanent(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(d.G, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
}

// discardEvents counts the EventDiscardCard events the game logged for
// this card, and the zone the last one named.
type dboWatcher struct{ Events []game.Event }

func (w *dboWatcher) OnEvent(_ *game.Game, ev game.Event) {
	if ev.Kind == game.EventDiscardCard {
		w.Events = append(w.Events, ev)
	}
}

func (d DboTable) Watch() *dboWatcher {
	w := &dboWatcher{}
	d.G.RegisterListener(w)
	return w
}
