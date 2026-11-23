package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_sheets3_test.go — the ability stickers that needed a seam of
// their own (ADR 0208, #4307): protection from a new quality, hellbent in a
// bundle, a divided sacrifice, a pairwise "can't block", heroic.

// stkProtectedFrom asks the engine's one protection predicate: does the
// host have protection from the source?
func stkProtectedFrom(s *effectstest.StkScene, source uuid.UUID) bool {
	var protected bool
	s.G.WithWriteLock(func() {
		s.G.RecomputeLayersIfStaleLocked()
		host, _ := s.G.LookupCardForEffect(s.Host)
		src, ok := s.G.LookupCardForEffect(source)
		if !ok {
			return
		}
		protected = game.ProtectedFrom(&host, game.SourceCharacteristics(&src))
	})
	return protected
}

// Notorious Sliver War.
func TestStkNotoriousSliverWarPumpsYourCreatures(t *testing.T) {
	s := effectstest.NewStkScene(t, "Notorious Sliver War", "Creatures you control get")
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 1)
	theirs := effectstest.P2Push(s.G, s.Opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
	s.Activate(0, "RRRRR")
	if p, _ := effectstest.P2PT(s.G, other); p != 2 {
		t.Errorf("my other creature's power = %d, want 2", p)
	}
	if p, _ := effectstest.P2PT(s.G, theirs); p != 1 {
		t.Errorf("an opponent's creature got the pump: power %d", p)
	}
}

func TestStkNotoriousSliverWarProtectsFromMultitypedCreatures(t *testing.T) {
	s := effectstest.NewStkScene(t, "Notorious Sliver War", "Protection from creatures")
	one := effectstest.P2Push(s.G, s.Opp.ID, "One", "Creature — Elf", "", 1, 1)
	two := effectstest.P2Push(s.G, s.Opp.ID, "Two", "Creature — Elf Warrior", "", 1, 1)
	if stkProtectedFrom(s, one) {
		t.Error("protected from a creature with one type")
	}
	if !stkProtectedFrom(s, two) {
		t.Error("not protected from a creature with two types")
	}
	// It stops the two-type creature from blocking the host.
	var r game.BlockRefusal
	s.G.ReadSnapshot(func() {
		host, _ := s.G.LookupCardForEffect(s.Host)
		blocker, _ := s.G.LookupCardForEffect(two)
		r = s.G.BlockPairRefusalLocked(&host, &blocker)
	})
	if r.Reason != game.BlockReasonProtection {
		t.Errorf("block refusal = %q, want protection", r.Reason)
	}
}

// Weird Angel Flame.
func TestStkWeirdAngelFlameHeroic(t *testing.T) {
	s := effectstest.NewStkScene(t, "Weird Angel Flame", "Heroic")
	effectstest.CastCatalogSpell(t, s.G, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.P2Card(s.Host)})
	s.Settle()
	if got := effectstest.CountersOn(s.G, s.Host, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	// A spell that targets something else does not trigger it.
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 5)
	effectstest.CastCatalogSpell(t, s.G, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.P2Card(other)})
	s.Settle()
	if got := effectstest.CountersOn(s.G, s.Host, game.CounterPlusOne); got != 2 {
		t.Errorf("a spell aimed elsewhere changed the counters: %d", got)
	}
}

func TestStkWeirdAngelFlameProtectsFromEvenManaValues(t *testing.T) {
	s := effectstest.NewStkScene(t, "Weird Angel Flame", "Protection from even")
	even := pushMVCreature2(s.G, s.Opp, "Even", "{1}{G}")
	odd := pushMVCreature2(s.G, s.Opp, "Odd", "{G}")
	if !stkProtectedFrom(s, even) {
		t.Error("not protected from a mana value 2 source")
	}
	if stkProtectedFrom(s, odd) {
		t.Error("protected from a mana value 1 source")
	}
}

func pushMVCreature2(g *game.Game, owner *game.Player, name, cost string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", ManaCost: cost,
		Power: 1, Toughness: 1, Owner: owner.ID, Controller: owner.ID,
	})
}

// Wild Ogre Bupkis.
func TestStkWildOgreBupkisAttackGrows(t *testing.T) {
	s := effectstest.NewStkScene(t, "Wild Ogre Bupkis", "put a +1/+1 counter")
	s.Attack()
	if got := effectstest.CountersOn(s.G, s.Host, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestStkWildOgreBupkisMetalcraft(t *testing.T) {
	s := effectstest.NewStkScene(t, "Wild Ogre Bupkis", "Metalcraft")
	rock := effectstest.P2Push(s.G, s.Opp.ID, "Their Rock", "Artifact", "", 0, 0)
	effectstest.P2Push(s.G, s.Me.ID, "Rock 1", "Artifact", "", 0, 0)
	effectstest.P2Push(s.G, s.Me.ID, "Rock 2", "Artifact", "", 0, 0)
	if stkProtectedFrom(s, rock) {
		t.Fatal("protected with two artifacts")
	}
	third := effectstest.P2Push(s.G, s.Me.ID, "Rock 3", "Artifact", "", 0, 0)
	if !stkProtectedFrom(s, rock) {
		t.Error("not protected from a noncreature permanent with three artifacts")
	}
	creature := effectstest.P2Push(s.G, s.Opp.ID, "Their Bear", "Creature — Bear", "", 1, 1)
	if stkProtectedFrom(s, creature) {
		t.Error("protected from a creature")
	}
	effectstest.B25Destroy(s.G, third)
	if stkProtectedFrom(s, rock) {
		t.Error("still protected after an artifact left")
	}
}

// Vampire Champion Fury.
func TestStkVampireChampionFuryHellbent(t *testing.T) {
	s := effectstest.NewStkScene(t, "Vampire Champion Fury", "Hellbent")
	effectstest.PushHandCard(s.G, s.Me)
	if p, tough := effectstest.P2PT(s.G, s.Host); p != 2 || tough != 2 {
		t.Fatalf("with a card in hand: %d/%d, want 2/2", p, tough)
	}
	// Emptying the hand re-reads the static at once.
	s.G.WithWriteLock(func() {
		for s.Me.Hand.Size() > 0 {
			c, err := s.Me.Hand.PopTop()
			if err != nil {
				t.Fatalf("pop: %v", err)
			}
			s.Me.Graveyard.PushTop(c)
			s.G.EmitEvent(game.Event{Kind: game.EventDiscardCard, CardID: c.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneGraveyard})
		}
	})
	if p, tough := effectstest.P2PT(s.G, s.Host); p != 5 || tough != 5 {
		t.Errorf("with an empty hand: %d/%d, want 5/5", p, tough)
	}
	effectstest.PushHandCard(s.G, s.Me)
	s.G.WithWriteLock(func() {
		s.G.EmitEvent(game.Event{Kind: game.EventDrawCard, Actor: s.Me.ID, OldZone: game.ZoneLibrary, NewZone: game.ZoneHand})
	})
	if p, _ := effectstest.P2PT(s.G, s.Host); p != 2 {
		t.Errorf("a card back in hand: power %d, want 2", p)
	}
}

func TestStkVampireChampionFurySacrificesForDividedDamage(t *testing.T) {
	s := effectstest.NewStkSceneOn(t, "Vampire Champion Fury", "divided as you choose", "Creature — Vampire", 4, 4)
	a := effectstest.P2Push(s.G, s.Opp.ID, "Bear A", "Creature — Bear", "", 3, 3)
	b := effectstest.P2Push(s.G, s.Opp.ID, "Bear B", "Creature — Bear", "", 3, 3)
	effectstest.FloatForTest(s.G, s.Me, "RR")
	if err := s.G.ActivateCatalogAbility(s.Me.ID, s.Host, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{effectstest.P2Card(a), effectstest.P2Card(b)},
		Distribution: map[uuid.UUID]int{a: 3, b: 1},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effects.OnBattlefield(s.G, s.Host) {
		t.Error("the creature was not sacrificed as the cost")
	}
	s.Settle()
	if effects.OnBattlefield(s.G, a) {
		t.Error("Bear A took 3 and survived")
	}
	if !effects.OnBattlefield(s.G, b) {
		t.Error("Bear B took 1 and died")
	}
}

// A division that does not add up to the creature's power (4) is refused.
func TestStkVampireChampionFuryRefusesADivisionThatIsNotItsPower(t *testing.T) {
	s := effectstest.NewStkSceneOn(t, "Vampire Champion Fury", "divided as you choose", "Creature — Vampire", 4, 4)
	a := effectstest.P2Push(s.G, s.Opp.ID, "Bear A", "Creature — Bear", "", 3, 3)
	effectstest.FloatForTest(s.G, s.Me, "RR")
	if err := s.G.ActivateCatalogAbility(s.Me.ID, s.Host, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{effectstest.P2Card(a)},
		Distribution: map[uuid.UUID]int{a: 3},
	}); err == nil {
		t.Error("a division of 3 for a 4-power creature was accepted")
	}
}

// Wrinkly Monkey Shenanigans.
func TestStkWrinklyMonkeyStopsATargetBlocking(t *testing.T) {
	s := effectstest.NewStkScene(t, "Wrinkly Monkey Shenanigans", "can't block this creature")
	a := effectstest.B12Creature(s.G, s.Opp.ID, "Wall A", "Creature — Wall", 0, 4)
	b := effectstest.B12Creature(s.G, s.Opp.ID, "Wall B", "Creature — Wall", 0, 4)
	s.Activate(0, "R", effectstest.P2Card(a))
	effectstest.AdvanceTo(t, s.G, game.StepDeclareAttackers)
	if err := s.G.DeclareAttacker(s.Host, s.Opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, s.G)
	effectstest.AdvanceTo(t, s.G, game.StepDeclareBlockers)
	if err := s.G.DeclareBlocker(a, s.Host); err == nil {
		t.Error("the barred creature blocked")
	}
	if err := s.G.DeclareBlocker(b, s.Host); err != nil {
		t.Errorf("another creature was barred too: %v", err)
	}
}

func TestStkWrinklyMonkeyMorbid(t *testing.T) {
	s := effectstest.NewStkScene(t, "Wrinkly Monkey Shenanigans", "Morbid")
	effectstest.SeedLibrary(s.Me, "A", "B")
	// No creature died: nothing happens at the end step.
	hand := s.Me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, s.G, 0)
	s.Settle()
	if s.Me.Hand.Size() != hand || effectstest.CountersOn(s.G, s.Host, game.CounterMinusOne) != 0 {
		t.Fatal("morbid triggered with no creature having died")
	}
}

func TestStkWrinklyMonkeyMorbidAfterADeath(t *testing.T) {
	s := effectstest.NewStkScene(t, "Wrinkly Monkey Shenanigans", "Morbid")
	effectstest.SeedLibrary(s.Me, "A", "B")
	victim := effectstest.P2Push(s.G, s.Opp.ID, "Victim", "Creature — Bear", "", 1, 1)
	effectstest.B25Destroy(s.G, victim)
	hand := s.Me.Hand.Size()
	effectstest.AdvanceToEndStepOf(t, s.G, 0)
	s.Settle()
	if got := effectstest.CountersOn(s.G, s.Host, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters = %d, want 1", got)
	}
	if got := s.Me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// Eldrazi Guacamole Tightrope: a stickered card in its owner's graveyard
// can be cast from it, paying 2 life besides its other costs.
func TestStkTightropeCastsFromTheGraveyardForTwoLife(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 12, "Eldrazi Guacamole Tightrope")
	effectstest.AdvanceToMain(t, g)
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
	effectstest.StkGive(t, g, me, card, "cast this card from your graveyard")
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
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, card) {
		t.Fatal("the creature did not enter the battlefield")
	}
	// The sticker came with it (CR 123.5).
	if c := effectstest.StkCard(t, g, card); len(c.Stickers) != 1 {
		t.Errorf("stickers = %+v, want the Tightrope sticker", c.Stickers)
	}
}

// ... only by its owner, only from the graveyard, and not once the sticker
// is gone with a trip to a hand.
func TestStkTightropePermissionIsTheOwnersAndLastsWhileStickered(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 12, "Eldrazi Guacamole Tightrope")
	effectstest.AdvanceToMain(t, g)
	card := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			OracleID: "00000000-0000-0000-0000-0000000000a4", Owner: me.ID, Controller: me.ID})
	})
	effectstest.StkGive(t, g, me, card, "cast this card from your graveyard")
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
	g, me := effectstest.StkTableWith(t, 12, "Eldrazi Guacamole Tightrope")
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	effectstest.StkGive(t, g, me, land, "cast this card from your graveyard")
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
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := effectstest.CastAndResolveCreature(t, g, "Chicken Troupe", "Creature — Bird Performer", chickenTroupeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 1 {
		t.Fatalf("tickets = %d, want 1", got)
	}
	effectstest.StkPutOn(t, g, me, id, "“Night”", 0)
	if got := stkNames(effectstest.StkCard(t, g, id)); got != "Night Chicken Troupe" {
		t.Errorf("name = %q, want %q", got, "Night Chicken Troupe")
	}
	opp := g.Seats[1]
	effectstest.CastAtWardedCreature(t, g, opp, id)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
}
