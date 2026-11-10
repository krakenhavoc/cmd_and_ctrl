package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// conspiracy_unraveler_test.go — Conspiracy Unraveler (Murders at Karlov
// Manor, #4469): "You may collect evidence 10 rather than pay the mana
// cost for spells you cast." A granted alternative cost whose whole price
// is collect evidence 10 (CR 118.9, 701.59a), through the real cast path.

const conspiracyUnravelerOracle = "e4541d02-2860-46ef-b630-1d6307b60766"

func TestConspiracyUnravelerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(conspiracyUnravelerOracle)
	if !ok {
		t.Fatal("Conspiracy Unraveler is not registered")
	}
	if spec.Name != "Conspiracy Unraveler" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
	if len(spec.GrantedAlternativeCosts) != 1 || spec.GrantedAlternativeCosts[0].Offer.CollectEvidence.ManaValueAtLeast != 10 {
		t.Error("want one granted collect evidence 10 offer")
	}
}

// cuTable is a main phase with the Unraveler under the active seat and a
// {6}{U} sorcery in hand.
func cuTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	src := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Conspiracy Unraveler", TypeLine: "Creature — Sphinx Detective",
		OracleID: conspiracyUnravelerOracle, Owner: me.ID, Controller: me.ID,
	})
	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{U}",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	return g, me, src, spell
}

func cuCast(g *game.Game, me *game.Player, spell uuid.UUID, ids ...uuid.UUID) error {
	return g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, AlternativeCost: GrantedAltCostCollectEvidenceTen, AltCostIDs: ids,
	})
}

// With 10 or more in the graveyard the offer is there; casting with it
// and no mana exiles exactly the named cards, puts the spell on the
// stack and announces one collection.
func TestConspiracyUnravelerCastsForEvidenceTen(t *testing.T) {
	g, me, _, spell := cuTable(t)
	gy := mkmCEEvidence(me, "{5}", "{3}{U}{U}", "{1}")
	got := offerWithKey(grantedCardOffers(g, me.ID, spell, game.ZoneHand), GrantedAltCostCollectEvidenceTen)
	if got == nil || !got.Granted || got.ManaCost != "" {
		t.Fatalf("offer = %+v", got)
	}
	if err := cuCast(g, me, spell, gy[0], gy[1]); err != nil {
		t.Fatalf("cast for collect evidence 10: %v", err)
	}
	if me.Hand.Contains(spell) || g.Stack.Size() != 1 {
		t.Error("the spell is not on the stack")
	}
	if !inExile(g, gy[0]) || !inExile(g, gy[1]) || !me.Graveyard.Contains(gy[2]) {
		t.Error("only the two named cards are exiled")
	}
	wantOneCollection(t, g, me.ID, 10, gy[0], gy[1])
}

// Below 10 the offer is withheld (CR 701.59b), and a short or foreign
// pick is refused with nothing paid.
func TestConspiracyUnravelerRefusesShortEvidence(t *testing.T) {
	g, me, _, spell := cuTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gy := mkmCEEvidence(me, "{5}", "{4}")
	theirs := mkmCEEvidence(opp, "{7}")
	if offerWithKey(grantedCardOffers(g, me.ID, spell, game.ZoneHand), GrantedAltCostCollectEvidenceTen) != nil {
		t.Error("offered over a graveyard totalling 9")
	}
	for name, err := range map[string]error{
		"a total of 9":       cuCast(g, me, spell, gy...),
		"nothing":            cuCast(g, me, spell),
		"an opponent's card": cuCast(g, me, spell, gy[0], theirs[0]),
		"a card twice":       cuCast(g, me, spell, gy[0], gy[0]),
	} {
		if err == nil {
			t.Errorf("%s: cast accepted", name)
		}
	}
	if !me.Hand.Contains(spell) || me.Graveyard.Size() != 2 || len(evidenceEvents(g)) != 0 {
		t.Fatal("a refused cast paid something")
	}
	gy = append(gy, gyCard(me, "More", "Sorcery", "{1}", ""))
	if offerWithKey(grantedCardOffers(g, me.ID, spell, game.ZoneHand), GrantedAltCostCollectEvidenceTen) == nil {
		t.Error("not offered over a graveyard totalling 10")
	}
}

// The offer goes with the Unraveler.
func TestConspiracyUnravelerOfferLeavesWithIt(t *testing.T) {
	g, me, src, spell := cuTable(t)
	mkmCEEvidence(me, "{5}", "{5}")
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(g.Battlefield, me.Graveyard, src); err != nil {
			t.Fatal(err)
		}
	})
	if offerWithKey(grantedCardOffers(g, me.ID, spell, game.ZoneHand), GrantedAltCostCollectEvidenceTen) != nil {
		t.Error("the offer outlived the Unraveler")
	}
}

// The alternative cost is not the spell's own collect evidence: a
// Crimestopper Sprite cast for it alone doesn't stun ("if evidence was
// collected" is false, CR 701.59c); with its own collect evidence 6 paid
// too, it stuns and two collections are announced.
func TestConspiracyUnravelerIsNotTheSpellsOwnEvidence(t *testing.T) {
	for _, own := range []bool{false, true} {
		g, me, _, _ := cuTable(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
		ten := mkmCEEvidence(me, "{5}", "{5}")
		params := game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostCollectEvidenceTen, AltCostIDs: ten}
		if own {
			six := mkmCEEvidence(me, "{6}")
			params.OptionalCosts = []int{0}
			params.ExileIDs = six
		}
		sprite := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: sprite, Name: "Crimestopper Sprite", TypeLine: "Creature — Faerie Detective",
			OracleID: mkmCESpriteOracle, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, sprite, params); err != nil {
			t.Fatalf("own=%v: cast: %v", own, err)
		}
		wantEvents := 1
		if own {
			wantEvents = 2
		}
		if n := len(evidenceEvents(g)); n != wantEvents {
			t.Errorf("own=%v: %d collections, want %d", own, n, wantEvents)
		}
		passPriorityAroundTable(t, g)
		answerPickTargetIfOpen(t, g, victim)
		passPriorityAroundTable(t, g)
		want := 0
		if own {
			want = 1
		}
		if got := countersOn(g, victim, game.CounterStun); got != want {
			t.Errorf("own=%v: %d stun counters, want %d", own, got, want)
		}
	}
}

// The view ships the offer's floor and its graveyard options; the bot's
// enumerated cast names enough evidence and, dispatched, casts the spell.
func TestConspiracyUnravelerViewAndEnumerator(t *testing.T) {
	g, me, _, spell := cuTable(t)
	gy := mkmCEEvidence(me, "{5}", "{5}", "{2}")
	activatedRowOfMarkKnown(g)
	var offer *protocol.AlternativeCostView
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		if s.ID != me.ID.String() {
			continue
		}
		for _, c := range s.Hand.Cards {
			if c.InstanceID != spell.String() {
				continue
			}
			for i := range c.AlternativeCosts {
				if c.AlternativeCosts[i].Key == GrantedAltCostCollectEvidenceTen {
					offer = &c.AlternativeCosts[i]
				}
			}
		}
	}
	if offer == nil {
		t.Fatal("no collect evidence offer in the view")
	}
	if offer.CollectEvidence != 10 || offer.PayOptions == nil || len(offer.PayOptions.Cards) != 3 || offer.PayOptions.Min != 1 {
		t.Errorf("offer view %+v, want a floor of 10 over three graveyard cards", offer)
	}
	var move *legal.Move
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != spell {
			continue
		}
		var p struct {
			AlternativeCost string   `json:"alternative_cost"`
			AltCostIDs      []string `json:"alt_cost_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.AlternativeCost == GrantedAltCostCollectEvidenceTen {
			if len(p.AltCostIDs) == 0 {
				t.Fatalf("move %q names no evidence", m.Label)
			}
			move = &m
		}
	}
	if move == nil {
		t.Fatal("no enumerated cast for collect evidence 10")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(move.Type), Player: me.ID, Caller: me.ID, Params: move.Params}); err != nil {
		t.Fatalf("move %q: %v", move.Label, err)
	}
	if me.Hand.Contains(spell) || len(evidenceEvents(g)) != 1 {
		t.Error("the enumerated cast did not collect evidence")
	}
	_ = gy
}
