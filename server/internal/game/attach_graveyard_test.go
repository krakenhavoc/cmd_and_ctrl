package game

import (
	"testing"

	"github.com/google/uuid"
)

// attach_graveyard_test.go pins ADR 0154 (#1780) on the rules side: an
// Aura whose enchant clause lists the graveyard attaches to a card there
// and stays legal while that card is the same object in a graveyard
// (CR 303.4, 303.4c, 400.7); every other off-battlefield attach is the
// CR 701.3b no-op it was; and the swapped clause ("enchant creature put
// onto the battlefield with this Aura") admits exactly the creature it
// names and refuses to move (CR 701.3b). The cards themselves are in
// cards/effects/animate_dead_test.go.

const (
	gyAuraOracle       = "test-graveyard-aura"
	creatureAuraOracle = "test-creature-aura"
)

// enchantCreatureCardInGraveyardSpec is Animate Dead's printed clause,
// the shape effects.EnchantCreatureCardInGraveyard builds.
func enchantCreatureCardInGraveyardSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "card_in_graveyard",
		Label: "enchant creature card in a graveyard",
		Zones: []ZoneKind{ZoneGraveyard},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
}

// withGraveyardAuraCatalog registers the two test Auras' clauses.
func withGraveyardAuraCatalog(t *testing.T) {
	t.Helper()
	withCatalogTargetSpec(t, func(oracleID string) *TargetSpec {
		switch oracleID {
		case gyAuraOracle:
			return enchantCreatureCardInGraveyardSpec()
		case creatureAuraOracle:
			return enchantCreatureSpec()
		}
		return nil
	})
}

// pushGraveyardCreature puts a creature card into p's graveyard.
func pushGraveyardCreature(p *Player, name string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Zombie",
		Power: 3, Toughness: 3, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// attachOnResolution is the Aura half of a resolving Aura spell: the
// attach announceEntryLocked runs between the zone move and EventETB,
// then the state-based actions a player would see before priority.
func attachOnResolution(t *testing.T, g *Game, aura, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.attachResolvedAuraLocked(aura, &StackItem{Targets: []TargetRef{{Kind: TargetCard, ID: host}}})
		g.runStateChecksLocked()
	})
}

func gyZoneKindOf(g *Game, id uuid.UUID) ZoneKind {
	var k ZoneKind
	g.ReadSnapshot(func() {
		if z := g.findCardZoneLocked(id); z != nil {
			k = z.Kind
		}
	})
	return k
}

// CR 303.4, CR 608.3c: an Aura whose enchant clause is "creature card in
// a graveyard" resolves attached to its target in an opponent's
// graveyard, and the state-based action finds it legal.
func TestAuraWithAGraveyardClauseAttachesToAGraveyardCard(t *testing.T) {
	withGraveyardAuraCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushGraveyardCreature(opp, "Gravecrawler")
	aura := pushCataloguedAura(g, me.ID, gyAuraOracle)

	attachOnResolution(t, g, aura, zombie)

	got, ok := battlefieldCardByID(g, aura)
	if !ok {
		t.Fatal("the Aura was put into the graveyard; a graveyard card is a legal host for its clause")
	}
	if !got.IsAttachedTo(zombie) {
		t.Fatalf("the Aura is attached to %+v, want the graveyard card", got.AttachedTo)
	}
	var epoch int
	g.ReadSnapshot(func() { epoch = g.cardObjectEpochLocked(zombie) })
	if got.AttachedEpoch != epoch {
		t.Errorf("AttachedEpoch = %d, want the host's epoch %d", got.AttachedEpoch, epoch)
	}
	// A graveyard card is never "enchanted creature" (ADR 0154 §10).
	g.ReadSnapshot(func() {
		if h := g.AttachedHostOf(&got); h != nil {
			t.Errorf("AttachedHostOf found %s on the battlefield", h.Name)
		}
	})
}

// CR 303.4c / 704.5m: the card leaving the graveyard (exile, the command
// zone) leaves the Aura enchanting an object that is not where its
// clause allows, and it is put into its owner's graveyard. The Aura is
// unattached on the way (Spellweaver Volute's ruling).
func TestGraveyardAuraFallsOffWhenTheCardLeavesTheGraveyard(t *testing.T) {
	for _, dest := range []string{"exile", "command"} {
		t.Run(dest, func(t *testing.T) {
			withGraveyardAuraCatalog(t)
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			zombie := pushGraveyardCreature(opp, "Gravecrawler")
			aura := pushCataloguedAura(g, me.ID, gyAuraOracle)
			attachOnResolution(t, g, aura, zombie)

			g.WithWriteLock(func() {
				to := g.Exile
				if dest == "command" {
					to = opp.Command
				}
				if _, err := MoveCard(opp.Graveyard, to, zombie); err != nil {
					t.Fatalf("MoveCard: %v", err)
				}
				g.runStateChecksLocked()
			})

			if _, ok := battlefieldCardByID(g, aura); ok {
				t.Fatal("CR 704.5m: the Aura stays on the battlefield after its host left the graveyard")
			}
			if !graveyardHas(me, aura) {
				t.Error("the Aura should be in its owner's graveyard")
			}
		})
	}
}

// CR 400.7, ADR 0154 decision 3: a card returned to hand and discarded
// again is back in the graveyard with the same instance ID, but it is a
// new object the Aura never enchanted, so the Aura falls off rather
// than silently following the card.
func TestGraveyardAuraFallsOffANewObjectWithTheSameID(t *testing.T) {
	withGraveyardAuraCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushGraveyardCreature(opp, "Gravecrawler")
	aura := pushCataloguedAura(g, me.ID, gyAuraOracle)
	attachOnResolution(t, g, aura, zombie)

	g.WithWriteLock(func() {
		if _, err := MoveCard(opp.Graveyard, opp.Hand, zombie); err != nil {
			t.Fatalf("to hand: %v", err)
		}
		if _, err := MoveCard(opp.Hand, opp.Graveyard, zombie); err != nil {
			t.Fatalf("back to the graveyard: %v", err)
		}
		g.runStateChecksLocked()
	})

	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Fatal("the Aura followed the card to a new object (CR 400.7)")
	}
	if gyZoneKindOf(g, zombie) != ZoneGraveyard {
		t.Error("the card should still be in the graveyard")
	}
}

// CR 701.3b: an attach to a card off the battlefield is still a quiet
// no-op unless the Aura's own clause lists that zone. An "enchant
// creature" Aura, an Equipment, and an Aura aimed at a card in exile
// all stay where they are, with the skip on the event log.
func TestAttachToAnOffBattlefieldCardNeedsAClauseThatListsItsZone(t *testing.T) {
	withGraveyardAuraCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushGraveyardCreature(opp, "Gravecrawler")
	exiled := uuid.New()
	g.Exile.PushTop(Card{InstanceID: exiled, Name: "Exiled Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	creatureAura := pushCataloguedAura(g, me.ID, creatureAuraOracle)
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	gyAura := pushCataloguedAura(g, me.ID, gyAuraOracle)
	land := uuid.New()
	opp.Graveyard.PushTop(Card{InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID})

	before := countEventsOfKind(g, EventAttachSkipped)
	g.WithWriteLock(func() {
		for _, step := range []struct {
			who  uuid.UUID
			host uuid.UUID
		}{
			{creatureAura, zombie}, // battlefield-only clause
			{sword, zombie},        // not an Aura
			{gyAura, exiled},       // the clause does not list exile
			{gyAura, land},         // the clause does not accept a land card
		} {
			if err := g.AttachForEffect(step.who, TargetRef{Kind: TargetCard, ID: step.host}); err != nil {
				t.Errorf("AttachForEffect returned %v; a CR 701.3b skip is not an error", err)
			}
		}
	})
	for _, id := range []uuid.UUID{creatureAura, sword, gyAura} {
		if c, _ := battlefieldCardByID(g, id); c.IsAttached() {
			t.Errorf("%s attached to %+v", c.Name, c.AttachedTo)
		}
	}
	if got := countEventsOfKind(g, EventAttachSkipped) - before; got != 4 {
		t.Errorf("%d attach skips logged, want 4", got)
	}
}

// pushSwappedAura is an Aura that has resolved Animate Dead's trigger:
// the swapped clause, naming `returned` (zero for none).
func pushSwappedAura(t *testing.T, g *Game, controller uuid.UUID, returned ObjectRef) uuid.UUID {
	t.Helper()
	aura := pushCataloguedAura(g, controller, gyAuraOracle)
	g.WithWriteLock(func() {
		if !g.SwapEnchantClauseForEffect(aura) {
			t.Fatal("SwapEnchantClauseForEffect: the Aura is not on the battlefield")
		}
		if returned.ID != uuid.Nil && !g.SetEnchantReturnedForEffect(aura, returned) {
			t.Fatal("SetEnchantReturnedForEffect refused a swapped Aura")
		}
	})
	return aura
}

func gyObjectRefOf(g *Game, id uuid.UUID) ObjectRef {
	var ref ObjectRef
	g.ReadSnapshot(func() { ref = ObjectRef{ID: id, Epoch: g.cardObjectEpochLocked(id)} })
	return ref
}

// ADR 0154 decision 4: the swapped clause admits the creature it names,
// on the battlefield, and nothing else — not another creature, and not
// the same card once it is a new object.
func TestSwappedClauseAdmitsOnlyTheReturnedCreature(t *testing.T) {
	withGraveyardAuraCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	returned := pushAttachTestCard(g, me.ID, "Returned", "Creature — Zombie")
	other := pushAttachTestCard(g, me.ID, "Other", "Creature — Bear")
	aura := pushSwappedAura(t, g, me.ID, gyObjectRefOf(g, returned))

	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: returned}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
		g.runStateChecksLocked()
	})
	got, ok := battlefieldCardByID(g, aura)
	if !ok || !got.IsAttachedTo(returned) {
		t.Fatalf("the swapped Aura should enchant the creature it returned: on=%v %+v", ok, got.AttachedTo)
	}

	// CR 701.3b, Animate Dead's ruling: "Attempting to move Animate Dead
	// to another creature won't work." It stays where it is.
	before := countEventsOfKind(g, EventAttachSkipped)
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: other})
		g.runStateChecksLocked()
	})
	got, ok = battlefieldCardByID(g, aura)
	if !ok || !got.IsAttachedTo(returned) {
		t.Fatalf("the swapped Aura moved: on=%v %+v", ok, got.AttachedTo)
	}
	if countEventsOfKind(g, EventAttachSkipped) == before {
		t.Error("the refused move left no EventAttachSkipped")
	}

	// The same card as a new object (CR 400.7) is not the creature.
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, returned)
		c.ObjectEpoch++
		g.runStateChecksLocked()
	})
	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Error("the swapped Aura is legal on a new object with the returned creature's ID")
	}
}

// A swapped Aura with no creature named (nothing came back), or attached
// to nothing, is illegal and goes to the graveyard (CR 704.5m).
func TestSwappedClauseWithNothingReturnedIsIllegal(t *testing.T) {
	withGraveyardAuraCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushGraveyardCreature(opp, "Gravecrawler")
	aura := pushCataloguedAura(g, me.ID, gyAuraOracle)
	attachOnResolution(t, g, aura, zombie)
	g.WithWriteLock(func() {
		g.SwapEnchantClauseForEffect(aura)
		g.runStateChecksLocked()
	})
	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Fatal("a swapped Aura still on the graveyard card is legal; its clause now names a permanent")
	}
	if !graveyardHas(me, aura) {
		t.Error("the Aura should be in its owner's graveyard")
	}
}

// #3156: an Aura whose enchant ability is only the swapped clause (it
// prints none, as Necromancy, which becomes an Aura through its trigger)
// attached to nothing is illegal too. Without the clause the catalog has
// no Targets for it, which is the uncatalogued-Aura exemption.
func TestSwappedAuraWithNoPrintedClauseAttachedToNothingIsIllegal(t *testing.T) {
	withGraveyardAuraCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	aura := pushAttachTestCard(g, me.ID, "Becomes an Aura", "Enchantment — Aura")
	g.WithWriteLock(func() {
		g.runStateChecksLocked()
	})
	if _, ok := battlefieldCardByID(g, aura); !ok {
		t.Fatal("an unswapped uncatalogued Aura should keep the manual-object exemption")
	}
	g.WithWriteLock(func() {
		g.SwapEnchantClauseForEffect(aura)
		g.runStateChecksLocked()
	})
	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Fatal("a swapped Aura attached to nothing stayed on the battlefield")
	}
	if !graveyardHas(me, aura) {
		t.Error("the Aura should be in its owner's graveyard")
	}
}

// CR 400.7: the swapped clause belongs to the object. Every zone change
// clears both fields, and so does the exile return's new-object reset.
func TestMoveCardClearsTheSwappedClause(t *testing.T) {
	withGraveyardAuraCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	returned := pushAttachTestCard(g, me.ID, "Returned", "Creature — Zombie")
	aura := pushSwappedAura(t, g, me.ID, gyObjectRefOf(g, returned))
	var moved Card
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: returned})
		c, err := MoveCard(g.Battlefield, me.Hand, aura)
		if err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
		moved = c
	})
	if moved.EnchantSwapped || moved.EnchantReturned != (ObjectRef{}) || moved.AttachedEpoch != 0 || moved.IsAttached() {
		t.Errorf("the card in hand kept the old object's state: swapped=%v returned=%+v epoch=%d attached=%+v",
			moved.EnchantSwapped, moved.EnchantReturned, moved.AttachedEpoch, moved.AttachedTo)
	}
}

// The three fields ride the undo clone and the restore point (ADR 0154
// decision 7), with the graveyard-host attachment intact: a restore
// mid-trigger still has the Aura legally on the graveyard card.
func TestGraveyardAttachmentSurvivesCloneAndSnapshot(t *testing.T) {
	withGraveyardAuraCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushGraveyardCreature(opp, "Gravecrawler")
	aura := pushCataloguedAura(g, me.ID, gyAuraOracle)
	attachOnResolution(t, g, aura, zombie)
	returned := pushAttachTestCard(g, me.ID, "Returned", "Creature — Zombie")
	swapped := pushSwappedAura(t, g, me.ID, gyObjectRefOf(g, returned))
	g.WithWriteLock(func() { _ = g.AttachForEffect(swapped, TargetRef{Kind: TargetCard, ID: returned}) })

	want, _ := battlefieldCardByID(g, aura)
	wantSwapped, _ := battlefieldCardByID(g, swapped)
	check := func(label string, cards []Card) {
		t.Helper()
		for _, c := range cards {
			switch c.InstanceID {
			case aura:
				if c.AttachedTo != want.AttachedTo || c.AttachedEpoch != want.AttachedEpoch {
					t.Errorf("%s: graveyard attachment = %+v@%d, want %+v@%d", label, c.AttachedTo, c.AttachedEpoch, want.AttachedTo, want.AttachedEpoch)
				}
			case swapped:
				if !c.EnchantSwapped || c.EnchantReturned != wantSwapped.EnchantReturned {
					t.Errorf("%s: swapped=%v returned=%+v, want true %+v", label, c.EnchantSwapped, c.EnchantReturned, wantSwapped.EnchantReturned)
				}
			}
		}
	}
	check("clone", g.Clone().Battlefield.Cards)
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	var back []Card
	restored.ReadSnapshot(func() { back = append(back, restored.Battlefield.Cards...) })
	check("snapshot", back)
	restored.WithWriteLock(func() { restored.runStateChecksLocked() })
	if _, ok := battlefieldCardByID(restored, aura); !ok {
		t.Error("after a restore the graveyard Aura is judged illegal")
	}
	if _, ok := battlefieldCardByID(restored, swapped); !ok {
		t.Error("after a restore the swapped Aura is judged illegal")
	}
}

// The epoch is carried when it is not zero: a host that has moved before
// (a creature that died) has epoch 1 or more in the graveyard.
func TestAttachedEpochRoundTripsWhenNonZero(t *testing.T) {
	withGraveyardAuraCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := pushGraveyardCreature(opp, "Gravecrawler")
	g.WithWriteLock(func() {
		_, _ = MoveCard(opp.Graveyard, opp.Hand, zombie)
		_, _ = MoveCard(opp.Hand, opp.Graveyard, zombie)
	})
	aura := pushCataloguedAura(g, me.ID, gyAuraOracle)
	attachOnResolution(t, g, aura, zombie)
	got, _ := battlefieldCardByID(g, aura)
	if got.AttachedEpoch != 2 {
		t.Fatalf("AttachedEpoch = %d, want 2", got.AttachedEpoch)
	}
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	back, _ := battlefieldCardByID(restored, aura)
	if back.AttachedEpoch != 2 {
		t.Errorf("restored AttachedEpoch = %d, want 2", back.AttachedEpoch)
	}
}
