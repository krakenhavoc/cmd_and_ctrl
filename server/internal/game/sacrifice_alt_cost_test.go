package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// sacrifice_alt_cost_test.go — #1727: a sacrifice as a component of an
// ALTERNATIVE cost. Dread Return's "Flashback—Sacrifice three
// creatures" is the shape; the card itself is pinned end to end in
// cards/effects/dread_return_test.go.
//
// Everything here is about the price, and every way it can be
// cheated fails in the caster's favour, which is the direction a
// sandbox must never err in:
//
//   - naming two creatures, or four, for a three-creature cost
//   - naming the same creature three times
//   - naming an opponent's creature (CR 701.21a)
//   - naming a permanent the clause does not admit
//   - naming the creatures on the additional cost's list instead
//
// and each refusal must leave NOTHING paid — the creatures on the
// battlefield and the spell in the graveyard. The last three tests
// are about timing: the creatures die with the spell already on the
// stack, so their dies triggers go above it, and a counter does not
// give them back.

// sacrificeFlashback is effects.FlashbackSacrifice spelled out, the
// trick flashbackOffer plays so this package need not import effects.
func sacrificeFlashback(n int) AlternativeCost {
	return AlternativeCost{
		Key:                 "flashback",
		Label:               "Flashback—Sacrifice creatures",
		FromZone:            ZoneGraveyard,
		ExileOnLeavingStack: true,
		Sacrifice:           creatureCostSpec().WithCount(n, n),
		PayLabel:            "creatures",
	}
}

// seedSacrificeFlashback wires a Dread-Return-shaped sorcery whose
// flashback sacrifices n creatures, puts it in seat 0's graveyard at a
// main phase, and gives seat 0 `have` creatures to pay with.
func seedSacrificeFlashback(t *testing.T, g *Game, n, have int) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	const oracle = "test-1727-dread-return"
	me := g.Seats[0]
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, sacrificeFlashback(n)))
	id := looterInGraveyard(t, g, me, oracle)
	fodder := make([]uuid.UUID, 0, have)
	for i := 0; i < have; i++ {
		fodder = append(fodder, pushCreatureToBattlefield(t, g, me))
	}
	return id, fodder
}

func castSacrificeFlashback(g *Game, id uuid.UUID, pay []uuid.UUID) error {
	return g.CastSpell(g.Seats[0].ID, id, CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "flashback",
		AltCostIDs:      pay,
	})
}

// assertNothingPaid is the validate-all-then-pay invariant: a refused
// cast leaves the spell where it was and every creature alive.
func assertNothingPaid(t *testing.T, g *Game, spell uuid.UUID, creatures []uuid.UUID) {
	t.Helper()
	if !g.Seats[0].Graveyard.Contains(spell) {
		t.Errorf("the refused cast moved the spell out of the graveyard")
	}
	if g.Stack.Contains(spell) {
		t.Errorf("the refused cast left the spell on the stack")
	}
	for _, id := range creatures {
		if !g.Battlefield.Contains(id) {
			t.Errorf("the refused cast sacrificed a creature anyway")
		}
	}
}

// The happy path: three creatures named, three creatures sacrificed —
// at announce, not at resolution — and the spell on the stack
// remembering the cost it was cast for.
func TestSacrificeFlashbackPaysWithTheNamedCreatures(t *testing.T) {
	g := newActiveGame(t)
	id, fodder := seedSacrificeFlashback(t, g, 3, 4)
	pay := fodder[:3]

	before := len(g.Events)
	if err := castSacrificeFlashback(g, id, pay); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	sacrificed := 0
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventSacrifice {
			sacrificed++
		}
	}
	if !g.Stack.Contains(id) {
		t.Fatalf("the flashback cast did not reach the stack")
	}
	for _, c := range pay {
		if g.Battlefield.Contains(c) {
			t.Errorf("a named creature is still on the battlefield")
		}
		if !g.Seats[0].Graveyard.Contains(c) {
			t.Errorf("a named creature did not reach its owner's graveyard")
		}
	}
	if !g.Battlefield.Contains(fodder[3]) {
		t.Errorf("the unnamed creature was sacrificed — the cost took more than it asked for")
	}
	if sacrificed != 3 {
		t.Errorf("EventSacrifice fired %d times, want 3 — a cost sacrifice is a sacrifice", sacrificed)
	}
	if item := g.StackMeta[id]; item == nil || item.AltCost != "flashback" {
		t.Errorf("the stack item does not remember the flashback cost")
	}
}

// Every refusal, and the one property they all share: nothing paid.
func TestSacrificeFlashbackRefusesABadPayment(t *testing.T) {
	cases := []struct {
		name string
		pay  func(t *testing.T, g *Game, fodder []uuid.UUID) []uuid.UUID
		want error
	}{
		{"too few", func(_ *testing.T, _ *Game, f []uuid.UUID) []uuid.UUID { return f[:2] }, ErrInvalidParam},
		{"too many", func(_ *testing.T, _ *Game, f []uuid.UUID) []uuid.UUID { return f[:4] }, ErrInvalidParam},
		{"none", func(_ *testing.T, _ *Game, _ []uuid.UUID) []uuid.UUID { return nil }, ErrInvalidParam},
		{"the same creature three times", func(_ *testing.T, _ *Game, f []uuid.UUID) []uuid.UUID {
			return []uuid.UUID{f[0], f[0], f[0]}
		}, ErrInvalidParam},
		{"an opponent's creature", func(t *testing.T, g *Game, f []uuid.UUID) []uuid.UUID {
			theirs := pushCreatureToBattlefield(t, g, g.Seats[1])
			return []uuid.UUID{f[0], f[1], theirs}
		}, ErrCardCallerMismatch},
		{"a noncreature permanent", func(_ *testing.T, g *Game, f []uuid.UUID) []uuid.UUID {
			rock := NewCard("Test Rock", g.Seats[0].ID)
			rock.TypeLine = "Artifact"
			g.Battlefield.PushTop(rock)
			return []uuid.UUID{f[0], f[1], rock.InstanceID}
		}, ErrIllegalTarget},
		{"a creature card in the graveyard", func(_ *testing.T, g *Game, f []uuid.UUID) []uuid.UUID {
			dead := NewCard("Test Corpse", g.Seats[0].ID)
			dead.TypeLine = "Creature — Zombie"
			g.Seats[0].Graveyard.PushTop(dead)
			return []uuid.UUID{f[0], f[1], dead.InstanceID}
		}, ErrCardNotFound},
		{"the spell itself", func(_ *testing.T, g *Game, f []uuid.UUID) []uuid.UUID {
			return []uuid.UUID{f[0], f[1], g.Seats[0].Graveyard.Cards[0].InstanceID}
		}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			id, fodder := seedSacrificeFlashback(t, g, 3, 4)
			pay := tc.pay(t, g, fodder)
			err := castSacrificeFlashback(g, id, pay)
			if err == nil {
				t.Fatalf("the cast was accepted with payment %v", pay)
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
			assertNothingPaid(t, g, id, fodder)
			if life := g.Seats[0].Life; life != StartingLife {
				t.Errorf("life = %d after a refused cast, want %d", life, StartingLife)
			}
		})
	}
}

// The creatures pay the ALTERNATIVE cost, so they ride alt_cost_ids.
// Sent on the additional cost's list instead, the cast is refused
// rather than read as a payment: the card has no additional cost, and
// the offer it claimed was not paid.
func TestSacrificeFlashbackRefusesTheAdditionalCostsList(t *testing.T) {
	g := newActiveGame(t)
	id, fodder := seedSacrificeFlashback(t, g, 3, 3)
	err := g.CastSpell(g.Seats[0].ID, id, CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "flashback",
		SacrificeIDs:    fodder,
	})
	if err == nil {
		t.Fatalf("a flashback paid on sacrifice_ids was accepted")
	}
	assertNothingPaid(t, g, id, fodder)
}

// CR 118.3: one creature pays one sacrifice. A card with BOTH a
// sacrifice additional cost and a sacrifice alternative cost cannot
// name the same creature to each — both validators would pass it, and
// the second payment would find it already dead with the spell on the
// stack.
func TestSacrificeAltCostRefusesACreatureNamedToBothCosts(t *testing.T) {
	g := newActiveGame(t)
	id, fodder := seedSacrificeFlashback(t, g, 1, 2)
	withCatalogAdditionalCost(t, func(o string) *AdditionalCost {
		if o == "test-1727-dread-return" {
			return &AdditionalCost{Sacrifice: creatureCostSpec(), Label: "Sacrifice a creature"}
		}
		return nil
	})
	err := g.CastSpell(g.Seats[0].ID, id, CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "flashback",
		AltCostIDs:      fodder[:1],
		SacrificeIDs:    fodder[:1],
	})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("err = %v, want ErrInvalidParam for one creature paying two sacrifices", err)
	}
	assertNothingPaid(t, g, id, fodder)

	// And the two DIFFERENT creatures the two costs ask for are fine.
	if err := g.CastSpell(g.Seats[0].ID, id, CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "flashback",
		AltCostIDs:      fodder[:1],
		SacrificeIDs:    fodder[1:2],
	}); err != nil {
		t.Fatalf("two creatures for two sacrifice costs: %v", err)
	}
	for _, c := range fodder {
		if g.Battlefield.Contains(c) {
			t.Errorf("a creature named to a cost survived the cast")
		}
	}
}

// The offer is on the table only when it can be paid (#695): two
// creatures cannot pay a three-creature flashback, so the card is not
// castable from the graveyard at all, and the view and the bot read
// that from the same predicate the announce gate uses.
func TestSacrificeFlashbackIsNotOfferedWithTooFewCreatures(t *testing.T) {
	g := newActiveGame(t)
	id, fodder := seedSacrificeFlashback(t, g, 3, 2)
	me := g.Seats[0]
	card := me.Graveyard.Cards[len(me.Graveyard.Cards)-1]
	if card.InstanceID != id {
		t.Fatalf("the spell is not on top of the graveyard")
	}
	alt := sacrificeFlashback(3)

	g.WithWriteLock(func() {
		if g.AlternativeCostPayableLocked(me.ID, id, &alt) {
			t.Errorf("a three-creature cost reads payable with two creatures")
		}
		if offers := g.CastOffersForLocked(me.ID, card, ZoneGraveyard, nil); len(offers) != 0 {
			t.Errorf("offers from the graveyard = %d, want none", len(offers))
		}
	})
	// An opponent's creatures do not make it payable: "your" creatures
	// is the rule, not "creatures".
	pushCreatureToBattlefield(t, g, g.Seats[1])
	pushCreatureToBattlefield(t, g, g.Seats[1])
	g.WithWriteLock(func() {
		if g.AlternativeCostPayableLocked(me.ID, id, &alt) {
			t.Errorf("an opponent's creatures made a sacrifice cost payable")
		}
	})
	// A third creature of the caster's own does.
	pushCreatureToBattlefield(t, g, me)
	g.WithWriteLock(func() {
		if !g.AlternativeCostPayableLocked(me.ID, id, &alt) {
			t.Errorf("three creatures cannot pay a three-creature cost")
		}
		cands := g.AltCostCandidatesLocked(me.ID, id, &alt)
		if len(cands) != 3 {
			t.Errorf("candidates = %d, want the caster's three creatures", len(cands))
		}
		if n := alt.CardPaymentCount(); n != 3 {
			t.Errorf("CardPaymentCount = %d, want 3", n)
		}
	})
	_ = fodder
}

// The order is the whole point of paying a cost with the spell on the
// stack: a "whenever a creature dies" trigger fires for each of the
// three and goes ABOVE the spell, so it resolves first.
func TestSacrificeFlashbackDiesTriggersGoAboveTheSpell(t *testing.T) {
	g := newActiveGame(t)
	id, fodder := seedSacrificeFlashback(t, g, 3, 3)
	const artistOracle = "test-1727-artist"
	artist := NewCard("Test Artist", g.Seats[0].ID)
	artist.TypeLine = "Creature — Vampire"
	artist.OracleID = artistOracle
	g.Battlefield.PushTop(artist)
	withCatalogTriggers(t, func(o string) []TriggeredAbility {
		if o != artistOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventLTB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.NewZone == ZoneGraveyard && ev.CardID != source.InstanceID
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return &StackItem{
					Kind:         StackItemTriggered,
					Controller:   source.Controller,
					Owner:        source.Owner,
					SourceCardID: source.InstanceID,
					Label:        "Test Artist — a creature died",
				}
			},
		}}
	})

	if err := castSacrificeFlashback(g, id, fodder); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	spell := g.StackMeta[id]
	if spell == nil {
		t.Fatalf("the spell is not on the stack")
	}
	// Stack order is StackMeta's Seq (a triggered ability has no card
	// in the stack zone): higher resolves first.
	above := 0
	for _, item := range g.StackMeta {
		if item.Kind != StackItemTriggered {
			continue
		}
		if item.Seq < spell.Seq {
			t.Errorf("a dies trigger went UNDER the spell (seq %d < %d) — the sacrifice was not paid with the spell on the stack",
				item.Seq, spell.Seq)
			continue
		}
		above++
	}
	if above != 3 {
		t.Errorf("%d dies triggers above the spell, want 3 — one per sacrificed creature", above)
	}
}

// A countered sacrifice-flashback spell: the creatures stay dead (the
// cost was paid, and countering a spell refunds nothing), and CR
// 702.34a exiles the card rather than letting it back into the
// graveyard to be flashed back again.
func TestCounteredSacrificeFlashbackKeepsTheCreaturesDeadAndIsExiled(t *testing.T) {
	g := newActiveGame(t)
	id, fodder := seedSacrificeFlashback(t, g, 3, 3)
	if err := castSacrificeFlashback(g, id, fodder); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	if err := g.CounterSpell(id, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	for _, c := range fodder {
		if g.Battlefield.Contains(c) {
			t.Errorf("a creature came back when the spell was countered")
		}
		if !g.Seats[0].Graveyard.Contains(c) {
			t.Errorf("a sacrificed creature left the graveyard when the spell was countered")
		}
	}
	if g.Seats[0].Graveyard.Contains(id) {
		t.Errorf("a countered flashback spell returned to the graveyard")
	}
	if !g.Exile.Contains(id) {
		t.Errorf("a countered flashback spell did not reach exile")
	}
}

// And on resolution, the same replacement: the card is exiled.
func TestResolvedSacrificeFlashbackIsExiled(t *testing.T) {
	g := newActiveGame(t)
	id, fodder := seedSacrificeFlashback(t, g, 3, 3)
	if err := castSacrificeFlashback(g, id, fodder); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	resolveTop(t, g)
	if g.Seats[0].Graveyard.Contains(id) {
		t.Errorf("a flashed-back spell returned to the graveyard")
	}
	if !g.Exile.Contains(id) {
		t.Errorf("a flashed-back spell did not reach exile")
	}
}

// Cost, not target (CR 601.2h): the caster's own hexproof creature
// pays the cost, exactly as Carrion Feeder may eat it.
func TestSacrificeAltCostIgnoresHexproof(t *testing.T) {
	g := newActiveGame(t)
	id, fodder := seedSacrificeFlashback(t, g, 3, 2)
	shy := NewCard("Test Hexproof", g.Seats[0].ID)
	shy.TypeLine = "Creature — Elf"
	shy.Keywords = []string{"hexproof"}
	g.Battlefield.PushTop(shy)
	if err := castSacrificeFlashback(g, id, append(fodder, shy.InstanceID)); err != nil {
		t.Fatalf("a hexproof creature could not pay its controller's own cost: %v", err)
	}
	if g.Battlefield.Contains(shy.InstanceID) {
		t.Errorf("the hexproof creature survived")
	}
}

// #1727: what the cost names is not the auto-tapper's to spend. A
// sacrifice alternative cost has no mana of its own, but a cost
// increase gives it some, and an Eldrazi Spawn named to the sacrifice
// must not be cracked for that mana first.
func TestCastAutoTapExclusionsCoverTheAlternativeCostsCards(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	ex := CastAutoTapExclusions(CastSpellParams{AltCostIDs: []uuid.UUID{a, b}})
	if !ex[a] || !ex[b] {
		t.Fatalf("exclusions = %v, want both alt_cost_ids", ex)
	}
}
