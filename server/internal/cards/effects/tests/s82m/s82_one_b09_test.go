package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
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
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

func s82b09Oil(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.LayeredCard(t, g, id).Counters["oil"]
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := s82b09Push(g, me.ID, c.name, s82b09CreatureTypeLn, c.oracle, 1, 1, 0)
			effectstest.AdvanceToMain(t, g)
			g.WithWriteLock(func() { me.Hand.Cards = nil })

			cr := effectstest.HandCardFull(me, "Bear", "Creature — Bear", "{0}", "", nil)
			if err := g.CastSpell(me.ID, cr, game.CastSpellParams{}); err != nil {
				t.Fatal(err)
			}
			effectstest.DrcSettle(t, g, me.ID)
			if n := s82b09Oil(t, g, id); n != 0 {
				t.Fatalf("a creature spell added %d oil counters", n)
			}

			ni := effectstest.HandCardFull(me, "Shock", s82b09NoncreatureType, "{0}", "", nil)
			if err := g.CastSpell(me.ID, ni, game.CastSpellParams{}); err != nil {
				t.Fatal(err)
			}
			effectstest.DrcSettle(t, g, me.ID)
			if n := s82b09Oil(t, g, id); n != 1 {
				t.Fatalf("a noncreature spell left %d oil counters, want 1", n)
			}
		})
	}
}

func TestS82B09SawbladeScampPingsEachOpponentForAnOilCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := s82b09Push(g, me.ID, "Sawblade Scamp", s82b09CreatureTypeLn, s82b09SawbladeScamp, 1, 1, 0)
	lives := effectstest.B24Lives(g)
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
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.DrcSettle(t, g, me.ID)
	got := effectstest.B24Lives(g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Trawler Drake", TypeLine: "Creature — Phyrexian Drake",
			OracleID: s82b09TrawlerDrake, ManaCost: "{0}", Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.DrcSettle(t, g, me.ID)
	if n := s82b09Oil(t, g, id); n != 1 {
		t.Fatalf("entered with %d oil counters, want 1", n)
	}
	if p := effectstest.PowerOf(t, g, id); p != 1 {
		t.Fatalf("power %d with one counter, want 1", p)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, "oil", 3); err != nil {
			t.Fatal(err)
		}
	})
	if p := effectstest.PowerOf(t, g, id); p != 4 {
		t.Errorf("power %d with four counters, want 4", p)
	}
}

func TestS82B09IchorSynthesizerBonusAndUnblockableNeedFourCounters(t *testing.T) {
	for _, c := range []struct {
		oil     int
		power   int
		blocked bool
	}{{3, 1, false}, {4, 3, true}, {6, 3, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := s82b09Push(g, me.ID, "Ichor Synthesizer", s82b09CreatureTypeLn, s82b09Synthesizer, 1, 3, c.oil)
		if p := effectstest.PowerOf(t, g, id); p != c.power {
			t.Errorf("oil %d: power %d, want %d", c.oil, p, c.power)
		}
		if restricted := effectstest.RestrictionsOf(t, g, id)&game.CantBeBlocked != 0; restricted != c.blocked {
			t.Errorf("oil %d: can't be blocked = %v, want %v", c.oil, restricted, c.blocked)
		}
	}
}

func TestS82B09SerumSovereignDrawsAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := s82b09Push(g, me.ID, "Serum Sovereign", "Creature — Phyrexian Sphinx", s82b09SerumSovereign, 4, 4, 0)
	effectstest.B06AddMana(me, "U")
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
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("A", "Instant"), effectstest.SearchTestLand("B", "Instant"), effectstest.SearchTestLand("C", "Instant"))
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.DrcSettle(t, g, me.ID)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		id := s82b09Push(g, me.ID, "Vindictive Flamestoker", s82b09CreatureTypeLn, s82b09Flamestoker, 1, 2, c.oil)
		g.WithWriteLock(func() { me.Hand.Cards = nil })
		hand := effectstest.HandCardFull(me, "Old", "Instant", "", "", nil)
		_ = hand
		effectstest.B06AddMana(me, c.mana...)
		err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
		if (err == nil) != c.ok {
			t.Errorf("oil %d, pool %v: err=%v, want ok=%v", c.oil, c.mana, err, c.ok)
			continue
		}
		if c.ok {
			effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("1", "Instant"), effectstest.SearchTestLand("2", "Instant"), effectstest.SearchTestLand("3", "Instant"), effectstest.SearchTestLand("4", "Instant"), effectstest.SearchTestLand("5", "Instant"))
			effectstest.DrcSettle(t, g, me.ID)
			if effectstest.FindBattlefieldCardByID(g, id) != nil {
				t.Errorf("oil %d: the Flamestoker was not sacrificed", c.oil)
			}
			if me.Hand.Size() != 4 {
				t.Errorf("oil %d: hand %d, want 4 (hand discarded, four drawn)", c.oil, me.Hand.Size())
			}
		}
	}
}

func TestS82B09AtmosphereSurgeonGrantsFlyingAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := s82b09Push(g, me.ID, "Atmosphere Surgeon", s82b09CreatureTypeLn, s82b09Surgeon, 2, 1, 1)
	friend := effectstest.Pr7Creature(g, me.ID, "Friend", 2)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: friend}}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Targets: target})
	effectstest.DrcSettle(t, g, me.ID)
	fc := effectstest.LayeredCard(t, g, friend)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.Pr7Creature(g, opp.ID, "Victim", 2)
	effectstest.CastCatalogSpell(t, g, "Free from Flesh", "Instant", s82b09FreeFromFlesh,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.DrcSettle(t, g, me.ID)
	if n := s82b09Oil(t, g, victim); n != 2 {
		t.Errorf("%d oil counters, want 2", n)
	}
	if p := effectstest.PowerOf(t, g, victim); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
}

func TestS82B09MagmaticSprinterEntersWithOilOnATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	friend := effectstest.Pr7Creature(g, me.ID, "Friend", 1)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Magmatic Sprinter", TypeLine: "Creature — Phyrexian Warrior",
			OracleID: s82b09MagmaticSprint, ManaCost: "{0}", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, friend)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := s82b09Push(g, me.ID, "Magmatic Sprinter", s82b09CreatureTypeLn, s82b09MagmaticSprint, 3, 2, c.oil)
			effectstest.AdvanceTo(t, g, game.StepEnd)
			effectstest.PassPriorityAroundTable(t, g)
			prompt := s82b09PayPrompt(g, me.ID)
			if (prompt != nil) != (c.answer != nil) {
				t.Fatalf("prompt open = %v, want %v", prompt != nil, c.answer != nil)
			}
			if prompt != nil {
				if err := g.ResolvePayUnless(prompt.ID, me.ID, *c.answer); err != nil {
					t.Fatal(err)
				}
				effectstest.PassPriorityAroundTable(t, g)
			}
			if on := effectstest.FindBattlefieldCardByID(g, id) != nil; on != c.stays {
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
			g, me, opp := effectstest.HarmonizeTable(t)
			id := s82b09Push(g, me.ID, "Serum-Core Chimera", "Creature — Phyrexian Chimera", s82b09Chimera, 2, 4, 3)
			var pitch uuid.UUID
			if c.discard != "" {
				pitch = effectstest.HandCardOf(g, me, "Pitch", c.discard, "", "")
			}
			victim := effectstest.Pr7Creature(g, opp.ID, "Victim", 2, "B")
			effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Drawn", "Instant"))
			effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
			effectstest.PassPriorityAroundTable(t, g)
			if c.discard != "" {
				effectstest.AnswerDiscard(t, g, me.ID, pitch)
			} else {
				effectstest.AnswerDiscard(t, g, me.ID)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
				effectstest.PickCard(t, g, me.ID, victim)
				effectstest.PassPriorityAroundTable(t, g)
			}
			damage := 0
			g.WithWriteLock(func() {
				if v := effectstest.FindBattlefieldCardForTest(g, victim); v != nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := s82b09Push(g, me.ID, "Serum-Core Chimera", "Creature — Phyrexian Chimera", s82b09Chimera, 2, 4, 3)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
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
