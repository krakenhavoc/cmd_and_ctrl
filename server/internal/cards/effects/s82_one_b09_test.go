package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 batch 9 — the oil-counter cards of Phyrexia: All Will Be One.

const (
	s82b09SawbladeScamp   = "8cac286c-1a17-4ee1-9858-d7a2f8c09037"
	s82b09SerumSovereign  = "d52ecb32-26ed-484b-b1a7-ee47b66f0f94"
	s82b09Chimera         = "ed54a59b-b24c-4681-b862-fa634fe759fd"
	s82b09Flamestoker     = "7ed9a887-9942-4482-942e-b3b05feb6342"
	s82b09TrawlerDrake    = "4fe38c9e-6fa0-4e05-9d3f-8fd635c457fb"
	s82b09Surgeon         = "af22ed78-7054-4f5c-84ab-2a274c3d47fb"
	s82b09Synthesizer     = "e82f46f3-143b-46ae-ab5d-4aa177437a64"
	s82b09FreeFromFlesh   = "e4a96cc7-8a4f-4007-8477-e1f30720db7b"
	s82b09MagmaticSprint  = "85d025d9-158d-4bf1-b3b2-361f3fe59663"
	s82b09CreatureTypeLn  = "Creature — Phyrexian"
	s82b09NoncreatureType = "Instant"
)

func TestS82B09CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82b09SawbladeScamp: "Sawblade Scamp", s82b09SerumSovereign: "Serum Sovereign",
		s82b09Chimera: "Serum-Core Chimera", s82b09Flamestoker: "Vindictive Flamestoker",
		s82b09TrawlerDrake: "Trawler Drake", s82b09Surgeon: "Atmosphere Surgeon",
		s82b09Synthesizer: "Ichor Synthesizer", s82b09FreeFromFlesh: "Free from Flesh",
		s82b09MagmaticSprint: "Magmatic Sprinter",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec)
		}
	}
}

// s82b09Push puts the card on the battlefield with `oil` oil counters.
func s82b09Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough, oil int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	}
	if oil > 0 {
		c.Counters = map[string]int{"oil": oil}
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

func s82b09Oil(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return layeredCard(t, g, id).Counters["oil"]
}

// Every oil creature: a noncreature spell you cast adds a counter; a
// creature spell, or an opponent's spell, does not.
func TestS82B09NoncreatureSpellsAddOilCounters(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Sawblade Scamp", s82b09SawbladeScamp}, {"Serum Sovereign", s82b09SerumSovereign},
		{"Serum-Core Chimera", s82b09Chimera}, {"Vindictive Flamestoker", s82b09Flamestoker},
		{"Trawler Drake", s82b09TrawlerDrake}, {"Atmosphere Surgeon", s82b09Surgeon},
		{"Ichor Synthesizer", s82b09Synthesizer},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := s82b09Push(g, me.ID, c.name, s82b09CreatureTypeLn, c.oracle, 1, 1, 0)
			advanceToMain(t, g)
			g.WithWriteLock(func() { me.Hand.Cards = nil })

			cr := handCardFull(me, "Bear", "Creature — Bear", "{0}", "", nil)
			if err := g.CastSpell(me.ID, cr, game.CastSpellParams{}); err != nil {
				t.Fatal(err)
			}
			drcSettle(t, g, me.ID)
			if n := s82b09Oil(t, g, id); n != 0 {
				t.Fatalf("a creature spell added %d oil counters", n)
			}

			ni := handCardFull(me, "Shock", s82b09NoncreatureType, "{0}", "", nil)
			if err := g.CastSpell(me.ID, ni, game.CastSpellParams{}); err != nil {
				t.Fatal(err)
			}
			drcSettle(t, g, me.ID)
			if n := s82b09Oil(t, g, id); n != 1 {
				t.Fatalf("a noncreature spell left %d oil counters, want 1", n)
			}
		})
	}
}

func TestS82B09SawbladeScampPingsEachOpponentForAnOilCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := s82b09Push(g, me.ID, "Sawblade Scamp", s82b09CreatureTypeLn, s82b09SawbladeScamp, 1, 1, 0)
	lives := b24Lives(g)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no oil counter")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = map[string]int{"oil": 2}
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	drcSettle(t, g, me.ID)
	got := b24Lives(g)
	for i := range got {
		want := lives[i]
		if i != 0 {
			want--
		}
		if got[i] != want {
			t.Errorf("seat %d life %d, want %d", i, got[i], want)
		}
	}
	if n := s82b09Oil(t, g, id); n != 1 {
		t.Errorf("%d oil counters left, want 1", n)
	}
}

func TestS82B09TrawlerDrakeEntersWithOneAndGrowsPerCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Trawler Drake", TypeLine: "Creature — Phyrexian Drake",
			OracleID: s82b09TrawlerDrake, ManaCost: "{0}", Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	drcSettle(t, g, me.ID)
	if n := s82b09Oil(t, g, id); n != 1 {
		t.Fatalf("entered with %d oil counters, want 1", n)
	}
	if p := powerOf(t, g, id); p != 1 {
		t.Fatalf("power %d with one counter, want 1", p)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, "oil", 3); err != nil {
			t.Fatal(err)
		}
	})
	if p := powerOf(t, g, id); p != 4 {
		t.Errorf("power %d with four counters, want 4", p)
	}
}

func TestS82B09IchorSynthesizerBonusAndUnblockableNeedFourCounters(t *testing.T) {
	for _, c := range []struct {
		oil     int
		power   int
		blocked bool
	}{{3, 1, false}, {4, 3, true}, {6, 3, true}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := s82b09Push(g, me.ID, "Ichor Synthesizer", s82b09CreatureTypeLn, s82b09Synthesizer, 1, 3, c.oil)
		if p := powerOf(t, g, id); p != c.power {
			t.Errorf("oil %d: power %d, want %d", c.oil, p, c.power)
		}
		if restricted := restrictionsOf(t, g, id)&game.CantBeBlocked != 0; restricted != c.blocked {
			t.Errorf("oil %d: can't be blocked = %v, want %v", c.oil, restricted, c.blocked)
		}
	}
}

func TestS82B09SerumSovereignDrawsAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := s82b09Push(g, me.ID, "Serum Sovereign", "Creature — Phyrexian Sphinx", s82b09SerumSovereign, 4, 4, 0)
	b06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no oil counter")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = map[string]int{"oil": 1}
			}
		}
	})
	seedSearchLibrary(me, searchTestLand("A", "Instant"), searchTestLand("B", "Instant"), searchTestLand("C", "Instant"))
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	drcSettle(t, g, me.ID)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if n := s82b09Oil(t, g, id); n != 0 {
		t.Errorf("%d oil counters left, want 0", n)
	}
	var scry bool
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			scry = true
		}
	}
	if !scry {
		t.Error("no scry prompt after the draw")
	}
}

func TestS82B09FlamestokerCostsOneLessPerCounter(t *testing.T) {
	for _, c := range []struct {
		oil  int
		mana []string
		ok   bool
	}{
		{0, []string{"R", "C", "C", "C", "C", "C", "C"}, true},
		{2, []string{"R", "C", "C", "C", "C"}, true},
		{2, []string{"R", "C", "C", "C"}, false},
		{9, []string{"R"}, true}, // generic reduction stops at the {R}
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		id := s82b09Push(g, me.ID, "Vindictive Flamestoker", s82b09CreatureTypeLn, s82b09Flamestoker, 1, 2, c.oil)
		g.WithWriteLock(func() { me.Hand.Cards = nil })
		hand := handCardFull(me, "Old", "Instant", "", "", nil)
		_ = hand
		b06AddMana(me, c.mana...)
		err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
		if (err == nil) != c.ok {
			t.Errorf("oil %d, pool %v: err=%v, want ok=%v", c.oil, c.mana, err, c.ok)
			continue
		}
		if c.ok {
			seedSearchLibrary(me, searchTestLand("1", "Instant"), searchTestLand("2", "Instant"), searchTestLand("3", "Instant"), searchTestLand("4", "Instant"), searchTestLand("5", "Instant"))
			drcSettle(t, g, me.ID)
			if findBattlefieldCardByID(g, id) != nil {
				t.Errorf("oil %d: the Flamestoker was not sacrificed", c.oil)
			}
			if me.Hand.Size() != 4 {
				t.Errorf("oil %d: hand %d, want 4 (hand discarded, four drawn)", c.oil, me.Hand.Size())
			}
		}
	}
}

func TestS82B09AtmosphereSurgeonGrantsFlyingAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := s82b09Push(g, me.ID, "Atmosphere Surgeon", s82b09CreatureTypeLn, s82b09Surgeon, 2, 1, 1)
	friend := pr7Creature(g, me.ID, "Friend", 2)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: friend}}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Targets: target})
	drcSettle(t, g, me.ID)
	fc := layeredCard(t, g, friend)
	if !game.HasKeyword(&fc, "flying") {
		t.Error("the target did not gain flying")
	}
	if n := s82b09Oil(t, g, id); n != 0 {
		t.Errorf("%d oil counters left, want 0", n)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Error("activated again with no counter")
	}
}

func TestS82B09FreeFromFleshPumpsAndPutsTwoOilCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pr7Creature(g, opp.ID, "Victim", 2)
	castCatalogSpell(t, g, "Free from Flesh", "Instant", s82b09FreeFromFlesh,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	drcSettle(t, g, me.ID)
	if n := s82b09Oil(t, g, victim); n != 2 {
		t.Errorf("%d oil counters, want 2", n)
	}
	if p := powerOf(t, g, victim); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
}

func TestS82B09MagmaticSprinterEntersWithOilOnATarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	friend := pr7Creature(g, me.ID, "Friend", 1)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Magmatic Sprinter", TypeLine: "Creature — Phyrexian Warrior",
			OracleID: s82b09MagmaticSprint, ManaCost: "{0}", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, friend)
	}
	passPriorityAroundTable(t, g)
	if n := s82b09Oil(t, g, friend); n != 2 {
		t.Errorf("target has %d oil counters, want 2", n)
	}
}

func s82b09PayPrompt(g *game.Game, p uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == p {
			return c
		}
	}
	return nil
}

func TestS82B09MagmaticSprinterEndStep(t *testing.T) {
	for _, c := range []struct {
		name   string
		oil    int
		answer *bool
		stays  bool
		left   int
	}{
		{"two counters, removes them", 2, ptrBool(true), true, 0},
		{"two counters, declines", 2, ptrBool(false), false, 2},
		{"three counters, removes two", 3, ptrBool(true), true, 1},
		{"one counter, is not asked", 1, nil, false, 1},
		{"no counters, is not asked", 0, nil, false, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := s82b09Push(g, me.ID, "Magmatic Sprinter", s82b09CreatureTypeLn, s82b09MagmaticSprint, 3, 2, c.oil)
			advanceTo(t, g, game.StepEnd)
			passPriorityAroundTable(t, g)
			prompt := s82b09PayPrompt(g, me.ID)
			if (prompt != nil) != (c.answer != nil) {
				t.Fatalf("prompt open = %v, want %v", prompt != nil, c.answer != nil)
			}
			if prompt != nil {
				if err := g.ResolvePayUnless(prompt.ID, me.ID, *c.answer); err != nil {
					t.Fatal(err)
				}
				passPriorityAroundTable(t, g)
			}
			if on := findBattlefieldCardByID(g, id) != nil; on != c.stays {
				t.Errorf("on the battlefield = %v, want %v", on, c.stays)
			}
			if c.stays {
				if n := s82b09Oil(t, g, id); n != c.left {
					t.Errorf("%d oil counters left, want %d", n, c.left)
				}
			} else if !c.stays {
				found := false
				g.ReadSnapshot(func() {
					for _, h := range me.Hand.Cards {
						if h.InstanceID == id {
							found = true
						}
					}
				})
				if !found {
					t.Error("the Sprinter did not return to its owner's hand")
				}
			}
		})
	}
}

func ptrBool(b bool) *bool { return &b }

func TestS82B09ChimeraDiscardsOnlyANonlandCardForDamage(t *testing.T) {
	for _, c := range []struct {
		name    string
		discard string // "" declines
		hits    bool
	}{
		{"a nonland card", "Instant", true},
		{"nothing", "", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			g, me, opp := harmonizeTable(t)
			id := s82b09Push(g, me.ID, "Serum-Core Chimera", "Creature — Phyrexian Chimera", s82b09Chimera, 2, 4, 3)
			var pitch uuid.UUID
			if c.discard != "" {
				pitch = handCardOf(g, me, "Pitch", c.discard, "", "")
			}
			victim := pr7Creature(g, opp.ID, "Victim", 2, "B")
			seedSearchLibrary(me, searchTestLand("Drawn", "Instant"))
			b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
			passPriorityAroundTable(t, g)
			if c.discard != "" {
				answerDiscard(t, g, me.ID, pitch)
			} else {
				answerDiscard(t, g, me.ID)
			}
			passPriorityAroundTable(t, g)
			if p := latestPickTarget(g, me.ID); p != nil {
				pickCard(t, g, me.ID, victim)
				passPriorityAroundTable(t, g)
			}
			damage := 0
			g.WithWriteLock(func() {
				if v := findBattlefieldCardForTest(g, victim); v != nil {
					damage = v.DamageMarked
				} else {
					damage = 3
				}
			})
			if c.hits != (damage == 3) {
				t.Errorf("damage %d, want hit=%v", damage, c.hits)
			}
			if n := s82b09Oil(t, g, id); n != 0 {
				t.Errorf("%d oil counters left, want 0", n)
			}
		})
	}
}

func TestS82B09ChimeraNeedsThreeCountersAndSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := s82b09Push(g, me.ID, "Serum-Core Chimera", "Creature — Phyrexian Chimera", s82b09Chimera, 2, 4, 3)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated outside a main phase")
	}
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters["oil"] = 2
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated with two counters")
	}
}
