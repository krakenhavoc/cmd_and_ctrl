package game

import (
	"testing"

	"github.com/google/uuid"
)

// copy_graveyard_source_test.go — #4797, ADR 0043's 2026-11-02 amendment:
// an "enters as a copy" replacement whose candidates are cards in a
// graveyard copies the picked card's printed values (CR 707.2), and its
// except clause can read the copied object's ID.

const testGraveyardCopyOracle = "test-enters-as-copy-of-a-graveyard-card"

func graveyardCopySelector(sawSource *uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches:         []EventKind{EventZoneMove},
		SelfReplacement: true,
		Label:           "Test Body Double",
		AppliesTo: func(ev *ReplacementEvent, _ *Game, src *Card) bool {
			return ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield && src != nil && ev.CardID == src.InstanceID
		},
		Controller: func(ev *ReplacementEvent, _ *Game, _ *Card) uuid.UUID { return ev.Actor },
		CopySelector: &CopySelector{
			Candidates: func(_ *ReplacementEvent, g *Game, _ *Card) []uuid.UUID {
				var out []uuid.UUID
				for _, p := range g.Seats {
					for _, c := range p.Graveyard.Cards {
						if c.IsCreature() {
							out = append(out, c.InstanceID)
						}
					}
				}
				return out
			},
			Except: func(ev *ReplacementEvent, _ *PrintedValues, _ *Game, _ *Card) {
				*sawSource = ev.CopySourceForEffect()
			},
		},
	}
}

func TestEntersAsACopyOfACardInAGraveyard(t *testing.T) {
	var sawSource uuid.UUID
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		testGraveyardCopyOracle: {graveyardCopySelector(&sawSource)},
	})
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceTo(t, g, StepPrecombatMain)

	dead := NewCard("Hill Giant", opp.ID)
	dead.TypeLine, dead.ManaCost = "Creature — Giant", "{3}{R}"
	dead.Power, dead.Toughness = 3, 3
	opp.Graveyard.PushTop(dead)

	double := NewCard("Test Body Double", me.ID)
	double.TypeLine, double.ManaCost, double.OracleID, double.Controller = "Creature — Shapeshifter", "{4}{U}", testGraveyardCopyOracle, me.ID
	me.Hand.PushTop(double)
	floatMana(me, "UUUUU")
	if err := g.CastSpell(me.ID, double.InstanceID, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)

	var prompt *PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceCopyTarget {
			prompt = c
		}
	}
	if prompt == nil || len(prompt.CopyOptions) != 1 || prompt.CopyOptions[0] != dead.InstanceID {
		t.Fatalf("prompt = %+v, want one offering the graveyard card", prompt)
	}
	if err := g.ResolveCopyTarget(prompt.ID, prompt.Chooser, dead.InstanceID); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	got, ok := g.LookupCardForEffect(double.InstanceID)
	if !ok || got.Name != "Hill Giant" || got.CurrentPower() != 3 || !g.Battlefield.Contains(double.InstanceID) {
		t.Fatalf("entered as %q %d (on battlefield %v), want a 3/3 Hill Giant", got.Name, got.CurrentPower(), ok)
	}
	if sawSource != dead.InstanceID {
		t.Errorf("the except clause read source %v, want the graveyard card %v", sawSource, dead.InstanceID)
	}
	if !opp.Graveyard.Contains(dead.InstanceID) {
		t.Error("copying a card leaves it where it is")
	}
}
