package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_sheets3_test.go — the ability stickers that needed a seam of
// their own (ADR 0208, #4307): protection from a new quality, hellbent in a
// bundle, a divided sacrifice, a pairwise "can't block", heroic.

// stkProtectedFrom asks the engine's one protection predicate: does the
// host have protection from the source?
func stkProtectedFrom(s *stkScene, source uuid.UUID) bool {
	var protected bool
	s.g.WithWriteLock(func() {
		s.g.RecomputeLayersIfStaleLocked()
		host, _ := s.g.LookupCardForEffect(s.host)
		src, ok := s.g.LookupCardForEffect(source)
		if !ok {
			return
		}
		protected = game.ProtectedFrom(&host, game.SourceCharacteristics(&src))
	})
	return protected
}

// Notorious Sliver War.
func TestStkNotoriousSliverWarPumpsYourCreatures(t *testing.T) {
	s := newStkScene(t, "Notorious Sliver War", "Creatures you control get")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	theirs := p2Push(s.g, s.opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
	s.activate(0, "RRRRR")
	if p, _ := p2PT(s.g, other); p != 2 {
		t.Errorf("my other creature's power = %d, want 2", p)
	}
	if p, _ := p2PT(s.g, theirs); p != 1 {
		t.Errorf("an opponent's creature got the pump: power %d", p)
	}
}

func TestStkNotoriousSliverWarProtectsFromMultitypedCreatures(t *testing.T) {
	s := newStkScene(t, "Notorious Sliver War", "Protection from creatures")
	one := p2Push(s.g, s.opp.ID, "One", "Creature — Elf", "", 1, 1)
	two := p2Push(s.g, s.opp.ID, "Two", "Creature — Elf Warrior", "", 1, 1)
	if stkProtectedFrom(s, one) {
		t.Error("protected from a creature with one type")
	}
	if !stkProtectedFrom(s, two) {
		t.Error("not protected from a creature with two types")
	}
	// It stops the two-type creature from blocking the host.
	var r game.BlockRefusal
	s.g.ReadSnapshot(func() {
		host, _ := s.g.LookupCardForEffect(s.host)
		blocker, _ := s.g.LookupCardForEffect(two)
		r = s.g.BlockPairRefusalLocked(&host, &blocker)
	})
	if r.Reason != game.BlockReasonProtection {
		t.Errorf("block refusal = %q, want protection", r.Reason)
	}
}

// Weird Angel Flame.
func TestStkWeirdAngelFlameHeroic(t *testing.T) {
	s := newStkScene(t, "Weird Angel Flame", "Heroic")
	castCatalogSpell(t, s.g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{p2Card(s.host)})
	s.settle()
	if got := countersOn(s.g, s.host, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	// A spell that targets something else does not trigger it.
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 5)
	castCatalogSpell(t, s.g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{p2Card(other)})
	s.settle()
	if got := countersOn(s.g, s.host, game.CounterPlusOne); got != 2 {
		t.Errorf("a spell aimed elsewhere changed the counters: %d", got)
	}
}

func TestStkWeirdAngelFlameProtectsFromEvenManaValues(t *testing.T) {
	s := newStkScene(t, "Weird Angel Flame", "Protection from even")
	even := pushMVCreature2(s.g, s.opp, "Even", "{1}{G}")
	odd := pushMVCreature2(s.g, s.opp, "Odd", "{G}")
	if !stkProtectedFrom(s, even) {
		t.Error("not protected from a mana value 2 source")
	}
	if stkProtectedFrom(s, odd) {
		t.Error("protected from a mana value 1 source")
	}
}

func pushMVCreature2(g *game.Game, owner *game.Player, name, cost string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", ManaCost: cost,
		Power: 1, Toughness: 1, Owner: owner.ID, Controller: owner.ID,
	})
}

// Wild Ogre Bupkis.
func TestStkWildOgreBupkisAttackGrows(t *testing.T) {
	s := newStkScene(t, "Wild Ogre Bupkis", "put a +1/+1 counter")
	s.attack()
	if got := countersOn(s.g, s.host, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestStkWildOgreBupkisMetalcraft(t *testing.T) {
	s := newStkScene(t, "Wild Ogre Bupkis", "Metalcraft")
	rock := p2Push(s.g, s.opp.ID, "Their Rock", "Artifact", "", 0, 0)
	p2Push(s.g, s.me.ID, "Rock 1", "Artifact", "", 0, 0)
	p2Push(s.g, s.me.ID, "Rock 2", "Artifact", "", 0, 0)
	if stkProtectedFrom(s, rock) {
		t.Fatal("protected with two artifacts")
	}
	third := p2Push(s.g, s.me.ID, "Rock 3", "Artifact", "", 0, 0)
	if !stkProtectedFrom(s, rock) {
		t.Error("not protected from a noncreature permanent with three artifacts")
	}
	creature := p2Push(s.g, s.opp.ID, "Their Bear", "Creature — Bear", "", 1, 1)
	if stkProtectedFrom(s, creature) {
		t.Error("protected from a creature")
	}
	b25Destroy(s.g, third)
	if stkProtectedFrom(s, rock) {
		t.Error("still protected after an artifact left")
	}
}

// Vampire Champion Fury.
func TestStkVampireChampionFuryHellbent(t *testing.T) {
	s := newStkScene(t, "Vampire Champion Fury", "Hellbent")
	pushHandCard(s.g, s.me)
	if p, tough := p2PT(s.g, s.host); p != 2 || tough != 2 {
		t.Fatalf("with a card in hand: %d/%d, want 2/2", p, tough)
	}
	// Emptying the hand re-reads the static at once.
	s.g.WithWriteLock(func() {
		for s.me.Hand.Size() > 0 {
			c, err := s.me.Hand.PopTop()
			if err != nil {
				t.Fatalf("pop: %v", err)
			}
			s.me.Graveyard.PushTop(c)
			s.g.EmitEvent(game.Event{Kind: game.EventDiscardCard, CardID: c.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneGraveyard})
		}
	})
	if p, tough := p2PT(s.g, s.host); p != 5 || tough != 5 {
		t.Errorf("with an empty hand: %d/%d, want 5/5", p, tough)
	}
	pushHandCard(s.g, s.me)
	s.g.WithWriteLock(func() {
		s.g.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: s.me.ID, OldZone: game.ZoneLibrary, NewZone: game.ZoneHand})
	})
	if p, _ := p2PT(s.g, s.host); p != 2 {
		t.Errorf("a card back in hand: power %d, want 2", p)
	}
}

func TestStkVampireChampionFurySacrificesForDividedDamage(t *testing.T) {
	s := newStkSceneOn(t, "Vampire Champion Fury", "divided as you choose", "Creature — Vampire", 4, 4)
	a := p2Push(s.g, s.opp.ID, "Bear A", "Creature — Bear", "", 3, 3)
	b := p2Push(s.g, s.opp.ID, "Bear B", "Creature — Bear", "", 3, 3)
	floatForTest(s.g, s.me, "RR")
	if err := s.g.ActivateCatalogAbility(s.me.ID, s.host, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{p2Card(a), p2Card(b)},
		Distribution: map[uuid.UUID]int{a: 3, b: 1},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if onBattlefield(s.g, s.host) {
		t.Error("the creature was not sacrificed as the cost")
	}
	s.settle()
	if onBattlefield(s.g, a) {
		t.Error("Bear A took 3 and survived")
	}
	if !onBattlefield(s.g, b) {
		t.Error("Bear B took 1 and died")
	}
}

// A division that does not add up to the creature's power (4) is refused.
func TestStkVampireChampionFuryRefusesADivisionThatIsNotItsPower(t *testing.T) {
	s := newStkSceneOn(t, "Vampire Champion Fury", "divided as you choose", "Creature — Vampire", 4, 4)
	a := p2Push(s.g, s.opp.ID, "Bear A", "Creature — Bear", "", 3, 3)
	floatForTest(s.g, s.me, "RR")
	if err := s.g.ActivateCatalogAbility(s.me.ID, s.host, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{p2Card(a)},
		Distribution: map[uuid.UUID]int{a: 3},
	}); err == nil {
		t.Error("a division of 3 for a 4-power creature was accepted")
	}
}

// Wrinkly Monkey Shenanigans.
func TestStkWrinklyMonkeyStopsATargetBlocking(t *testing.T) {
	s := newStkScene(t, "Wrinkly Monkey Shenanigans", "can't block this creature")
	a := b12Creature(s.g, s.opp.ID, "Wall A", "Creature — Wall", 0, 4)
	b := b12Creature(s.g, s.opp.ID, "Wall B", "Creature — Wall", 0, 4)
	s.activate(0, "R", p2Card(a))
	advanceTo(t, s.g, game.StepDeclareAttackers)
	if err := s.g.DeclareAttacker(s.host, s.opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, s.g)
	advanceTo(t, s.g, game.StepDeclareBlockers)
	if err := s.g.DeclareBlocker(a, s.host); err == nil {
		t.Error("the barred creature blocked")
	}
	if err := s.g.DeclareBlocker(b, s.host); err != nil {
		t.Errorf("another creature was barred too: %v", err)
	}
}

func TestStkWrinklyMonkeyMorbid(t *testing.T) {
	s := newStkScene(t, "Wrinkly Monkey Shenanigans", "Morbid")
	seedLibrary(s.me, "A", "B")
	// No creature died: nothing happens at the end step.
	hand := s.me.Hand.Size()
	advanceToEndStepOf(t, s.g, 0)
	s.settle()
	if s.me.Hand.Size() != hand || countersOn(s.g, s.host, game.CounterMinusOne) != 0 {
		t.Fatal("morbid triggered with no creature having died")
	}
}

func TestStkWrinklyMonkeyMorbidAfterADeath(t *testing.T) {
	s := newStkScene(t, "Wrinkly Monkey Shenanigans", "Morbid")
	seedLibrary(s.me, "A", "B")
	victim := p2Push(s.g, s.opp.ID, "Victim", "Creature — Bear", "", 1, 1)
	b25Destroy(s.g, victim)
	hand := s.me.Hand.Size()
	advanceToEndStepOf(t, s.g, 0)
	s.settle()
	if got := countersOn(s.g, s.host, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters = %d, want 1", got)
	}
	if got := s.me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// Eldrazi Guacamole Tightrope: a stickered card in its owner's graveyard
// can be cast from it, paying 2 life besides its other costs.
func TestStkTightropeCastsFromTheGraveyardForTwoLife(t *testing.T) {
	g, me := stkTableWith(t, 12, "Eldrazi Guacamole Tightrope")
	advanceToMain(t, g)
	card := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, OracleID: "00000000-0000-0000-0000-0000000000a3", Owner: me.ID, Controller: me.ID})
	})
	cast := func(claim string) error {
		return g.CastSpell(me.ID, card, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: claim})
	}
	if err := cast(game.AltCostKeyStickerGraveyard); err == nil {
		t.Fatal("a card with no sticker was cast from the graveyard")
	}
	stkGive(t, g, me, card, "cast this card from your graveyard")
	life := me.Life
	if err := cast(""); err == nil {
		t.Fatal("cast without paying the life")
	}
	if err := cast(game.AltCostKeyStickerGraveyard); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	if got := life - me.Life; got != 2 {
		t.Errorf("life paid = %d, want 2", got)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, card) {
		t.Fatal("the creature did not enter the battlefield")
	}
	// The sticker came with it (CR 123.5).
	if c := stkCard(t, g, card); len(c.Stickers) != 1 {
		t.Errorf("stickers = %+v, want the Tightrope sticker", c.Stickers)
	}
}

// ... only by its owner, only from the graveyard, and not once the sticker
// is gone with a trip to a hand.
func TestStkTightropePermissionIsTheOwnersAndLastsWhileStickered(t *testing.T) {
	g, me := stkTableWith(t, 12, "Eldrazi Guacamole Tightrope")
	advanceToMain(t, g)
	card := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			OracleID: "00000000-0000-0000-0000-0000000000a4", Owner: me.ID, Controller: me.ID})
	})
	stkGive(t, g, me, card, "cast this card from your graveyard")
	perm := func(player uuid.UUID) *game.CastPermission {
		var out *game.CastPermission
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(card)
			out = g.CastPermissionForLocked(player, c, game.ZoneGraveyard)
		})
		return out
	}
	if perm(me.ID) == nil {
		t.Fatal("the owner has no permission")
	}
	if perm(g.Seats[1].ID) != nil {
		t.Error("an opponent may cast the card from the graveyard")
	}
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardForEffect(card, game.ZoneHand) })
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(card)
		if len(c.Stickers) != 0 || g.CastPermissionForLocked(me.ID, c, game.ZoneGraveyard) != nil {
			t.Errorf("a card in a hand kept its sticker or its permission: %+v", c.Stickers)
		}
	})
}

// A land has no cast to permit.
func TestStkTightropeGivesALandNothing(t *testing.T) {
	g, me := stkTableWith(t, 12, "Eldrazi Guacamole Tightrope")
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	stkGive(t, g, me, land, "cast this card from your graveyard")
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(land)
		if g.CastPermissionForLocked(me.ID, c, game.ZoneGraveyard) != nil {
			t.Error("a land card has a cast permission")
		}
	})
}

const chickenTroupeOracle = "68982532-c0c8-477b-9dd4-180c6d22731e"

// Chicken Troupe: a ticket, then a sticker, and ward {2}.
func TestStkChickenTroupeGetsATicketThenAStickerAndHasWard(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := castAndResolveCreature(t, g, "Chicken Troupe", "Creature — Bird Performer", chickenTroupeOracle)
	passPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 1 {
		t.Fatalf("tickets = %d, want 1", got)
	}
	stkPutOn(t, g, me, id, "“Night”", 0)
	if got := stkNames(stkCard(t, g, id)); got != "Night Chicken Troupe" {
		t.Errorf("name = %q, want %q", got, "Night Chicken Troupe")
	}
	opp := g.Seats[1]
	castAtWardedCreature(t, g, opp, id)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
}
