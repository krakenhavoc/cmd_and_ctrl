package game

import (
	"testing"

	"github.com/google/uuid"
)

// sacrifice_test.go — S21 sub-PR 1: sacrifice as an engine
// operation, and the sacrifice-cost mana abilities that ride on it.

func pushIntrinsicPermanent(g *Game, owner *Player, name, typeLine string, abilities []ManaAbilityShape, keywords []string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = typeLine
	c.ManaAbilities = abilities
	c.Keywords = keywords
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func hasEvent(g *Game, kind EventKind, cardID uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == kind && ev.CardID == cardID {
			return true
		}
	}
	return false
}

func TestSacrificeRoutesToGraveyardAndEmitsBothEvents(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)
	if err := g.SacrificePermanent(me.ID, id); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Errorf("sacrificed permanent should be in its owner's graveyard")
	}
	if !hasEvent(g, EventSacrifice, id) {
		t.Errorf("no EventSacrifice emitted")
	}
	// A sacrificed creature also dies: the ordinary LTB fires so
	// dies-triggers (Blood Artist, Doomed Traveler) still see it.
	var died bool
	for _, ev := range g.Events {
		if ev.Kind == EventLTB && ev.CardID == id && ev.NewZone == ZoneGraveyard {
			died = true
		}
	}
	if !died {
		t.Errorf("sacrifice must also emit the dies (LTB → graveyard) event")
	}
}

// EventSacrifice fires while the permanent is still on the
// battlefield, so a payoff can read its characteristics.
func TestSacrificeEventPrecedesTheMove(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)
	probe := &sacrificeProbe{watch: id}
	g.RegisterListener(probe)
	if err := g.SacrificePermanent(me.ID, id); err != nil {
		t.Fatal(err)
	}
	if !probe.sawOnBattlefield {
		t.Errorf("EventSacrifice should fire before the permanent leaves the battlefield")
	}
}

// sacrificeProbe records whether the sacrificed permanent was still
// on the battlefield when EventSacrifice reached the listeners.
type sacrificeProbe struct {
	watch            uuid.UUID
	sawOnBattlefield bool
}

func (p *sacrificeProbe) OnEvent(g *Game, ev Event) {
	if ev.Kind == EventSacrifice && ev.CardID == p.watch {
		p.sawOnBattlefield = g.Battlefield.Contains(p.watch)
	}
}

func TestSacrificeRejectsAnotherPlayersPermanent(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushIntrinsicPermanent(g, opp, "Bear", "Creature — Bear", nil, nil)
	if err := g.SacrificePermanent(me.ID, theirs); err != ErrCardCallerMismatch {
		t.Fatalf("sacrificing an opponent's permanent: %v, want ErrCardCallerMismatch", err)
	}
	if !g.Battlefield.Contains(theirs) {
		t.Errorf("rejected sacrifice must leave the permanent alone")
	}
	if err := g.SacrificePermanent(me.ID, uuid.New()); err != ErrCardNotFound {
		t.Errorf("unknown card: want ErrCardNotFound")
	}
}

// --- sacrifice-cost mana abilities -------------------------------

func treasureAbility() []ManaAbilityShape {
	return []ManaAbilityShape{{
		TapCost:       true,
		SacrificeCost: true,
		Produced:      "{B}",
		Label:         "{T}, Sacrifice: Add {B}",
	}}
}

func TestSacrificeCostManaAbilityCracksThePermanent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Treasure", "Token Artifact — Treasure", treasureAbility(), nil)
	if err := g.ActivateManaAbility(me.ID, id, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if g.Battlefield.Contains(id) {
		t.Errorf("a cracked Treasure should leave the battlefield")
	}
	// #596 / CR 704.5d: a Treasure is a TOKEN. It reaches the
	// graveyard — EventSacrifice and the dies triggers below are
	// judged on that — and the state-based check that the sacrifice
	// runs on its way out then removes it. It does not sit in the
	// graveyard as a card anyone can reanimate.
	if me.Graveyard.Contains(id) {
		t.Errorf("a cracked Treasure token must cease to exist, not stay in the graveyard")
	}
	if !hasEvent(g, EventSacrifice, id) {
		t.Errorf("no EventSacrifice from the mana ability's cost")
	}
	if !hasEvent(g, EventLTB, id) {
		t.Errorf("the cracked Treasure's dies event went missing — a token still dies before it ceases to exist")
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "B" {
		t.Errorf("mana pool = %+v, want one {B}", me.ManaPool)
	}
}

// A tapped permanent can't pay a {T} cost — and the failed check
// must not sacrifice it on the way out.
func TestSacrificeCostNotPaidWhenTapCostFails(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Treasure", "Token Artifact — Treasure", treasureAbility(), nil)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, ManaAbilityParams{}); err != ErrAlreadyTapped {
		t.Fatalf("tapped Treasure: %v, want ErrAlreadyTapped", err)
	}
	if !g.Battlefield.Contains(id) {
		t.Errorf("failed activation must not sacrifice the permanent")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("failed activation must not add mana")
	}
}

// Eldrazi Spawn shape: sacrifice with no tap, so it works the turn
// it arrives and while tapped.
func TestSacrificeOnlyCostIgnoresTapState(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Eldrazi Spawn", "Token Creature — Eldrazi Spawn",
		[]ManaAbilityShape{{SacrificeCost: true, Produced: "{C}", Label: "Sacrifice: Add {C}"}}, nil)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("sacrifice-only ability while tapped: %v", err)
	}
	if g.Battlefield.Contains(id) {
		t.Errorf("Spawn should be sacrificed")
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "C" {
		t.Errorf("mana pool = %+v, want one {C}", me.ManaPool)
	}
}

// --- intrinsic card abilities ------------------------------------

func TestIntrinsicKeywordsReachTheLayerEngine(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Spirit", "Token Creature — Spirit", nil, []string{"flying"})
	g.WithWriteLock(func() { g.recomputeLayersLocked() })
	var found *Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			found = &g.Battlefield.Cards[i]
		}
	}
	if found == nil {
		t.Fatalf("token missing from the battlefield")
	}
	if !HasKeyword(found, "flying") {
		t.Errorf("a token's intrinsic keyword should be live: %v", found.Effective().Abilities)
	}
}

func TestIntrinsicManaAbilitiesWinOverTheCatalog(t *testing.T) {
	c := Card{
		Name:          "Treasure",
		TypeLine:      "Token Artifact — Treasure",
		ManaAbilities: treasureAbility(),
	}
	got := ManaAbilitiesForCard(c)
	if len(got) != 1 || !got[0].SacrificeCost {
		t.Errorf("intrinsic abilities = %+v", got)
	}
	// A basic land with no intrinsic list still gets the synthetic one.
	forest := Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
	if syn := ManaAbilitiesForCard(forest); len(syn) != 1 || syn[0].Produced != "{G}" {
		t.Errorf("synthetic basic-land path broken: %+v", syn)
	}
}
