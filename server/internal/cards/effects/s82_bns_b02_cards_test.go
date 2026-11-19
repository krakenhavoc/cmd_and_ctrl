package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82BnsContagionEngineOracle = "482bd55a-ba73-4722-8643-94fa12818515"
	s82BnsGrindstoneOracle      = "97560adc-814e-450c-9ca7-f9364e910a0b"
	s82BnsDetentionSphereOracle = "0a0b376d-7824-4ba5-8bc9-bc0be12800c5"
	s82BnsJourneyOracle         = "2bfec482-be0d-4768-84e7-1c2c74ea1373"
	s82BnsMerrimentOracle       = "78826624-8784-4003-8813-230f299a588f"
	s82BnsArchmagesCharmOracle  = "85cbb035-25df-43bb-ae06-ab098f448b02"
	s82BnsBackForMoreOracle     = "670df8be-21cd-49d8-92e3-d7cc38b256d6"
	s82BnsHinderingLightOracle  = "945ac15c-780b-44a9-8a8e-011cf1c30876"
	s82BnsHypothesizzleOracle   = "e4e6cb61-f673-4f36-b95c-c84a68459503"
	s82BnsIonizeOracle          = "01da67a5-e3dc-44f4-8a93-513bfa4d8fbc"
	s82BnsRideDownOracle        = "9d136448-7331-4a61-b2ad-cfeb23dfa464"
	s82BnsThornadoOracle        = "7476a14a-611e-4f81-b603-fd600e88568f"
	s82BnsDoomBladeOracle       = "59e7f2ae-4535-4191-98be-3e65b6b2befa"
)

func TestS82BnsB02EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82BnsContagionEngineOracle: "Contagion Engine",
		s82BnsGrindstoneOracle:      "Grindstone",
		s82BnsDetentionSphereOracle: "Detention Sphere",
		s82BnsJourneyOracle:         "Journey to Nowhere",
		s82BnsMerrimentOracle:       "Outlaws' Merriment",
		s82BnsArchmagesCharmOracle:  "Archmage's Charm",
		s82BnsBackForMoreOracle:     "Back for More",
		s82BnsHinderingLightOracle:  "Hindering Light",
		s82BnsHypothesizzleOracle:   "Hypothesizzle",
		s82BnsIonizeOracle:          "Ionize",
		s82BnsRideDownOracle:        "Ride Down",
		s82BnsThornadoOracle:        "Thornado",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82BnsB02JourneyToNowhereExilesAndReturnsOnLeaving(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	journey := castCatalogSpell(t, g, "Journey to Nowhere", "Enchantment", s82BnsJourneyOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(journey) })
	g.RunStateChecksForTest()
	back := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Their Threat" {
				back++
				if c.Controller != opp.ID {
					t.Error("it returns under its owner's control")
				}
			}
		}
	})
	if back != 1 {
		t.Fatalf("after Journey left, %d copies are on the battlefield, want 1", back)
	}
}

func TestS82BnsB02ThornadoDestroysOnlyAFlyingCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ground := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	if err := castCatalogSpellErrStrict(t, g, "Thornado", "Instant", s82BnsThornadoOracle, cardRefs(ground)); err == nil {
		t.Fatal("Thornado targeted a creature without flying")
	}
	castCatalogSpell(t, g, "Thornado", "Instant", s82BnsThornadoOracle, cardRefs(bird))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flier survived")
	}
	if !g.Battlefield.Contains(ground) {
		t.Error("the ground creature was destroyed")
	}
	spec, _ := Lookup(s82BnsThornadoOracle)
	if len(spec.Activated) != 1 {
		t.Errorf("Thornado declares %d activated abilities, want its cycling", len(spec.Activated))
	}
}

func TestS82BnsB02IonizeCountersAndHitsTheSpellsController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	life := me.Life
	castCatalogSpell(t, g, "Ionize", "Instant", s82BnsIonizeOracle, cardRefs(victim))
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("the spell was not countered")
	}
	if me.Life != life-2 {
		t.Errorf("the spell's controller is at %d, want %d", me.Life, life-2)
	}
}

func TestS82BnsB02HindingLightCountersASpellAimedAtYouAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	toMain(t, g)
	aimedAtThem := g2CastFrom(t, g, opp, "Doom Blade", "Instant", s82BnsDoomBladeOracle, game.CastSpellParams{Targets: cardRefs(theirs)})
	if err := castCatalogSpellErrStrict(t, g, "Hindering Light", "Instant", s82BnsHinderingLightOracle, cardRefs(aimedAtThem)); err == nil {
		t.Fatal("Hindering Light targeted a spell that targets neither you nor your permanent")
	}
	aimedAtMe := g2CastFrom(t, g, opp, "Doom Blade", "Instant", s82BnsDoomBladeOracle, game.CastSpellParams{Targets: cardRefs(mine)})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Hindering Light", "Instant", s82BnsHinderingLightOracle, cardRefs(aimedAtMe))
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(aimedAtMe) || !g.Battlefield.Contains(mine) {
		t.Error("the spell was not countered")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand is %d, want %d (one card drawn)", me.Hand.Size(), hand+1)
	}
}

func TestS82BnsB02ArchmagesCharmEachMode(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		g := newCatalogGame(t)
		victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
		castModal(t, g, "Archmage's Charm", "Instant", s82BnsArchmagesCharmOracle, []int{0}, cardRefs(victim))
		passPriorityAroundTable(t, g)
		if g.Stack.Contains(victim) {
			t.Error("the spell was not countered")
		}
	})
	t.Run("draw two", func(t *testing.T) {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		hand := opp.Hand.Size()
		castModal(t, g, "Archmage's Charm", "Instant", s82BnsArchmagesCharmOracle, []int{1},
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		passPriorityAroundTable(t, g)
		if opp.Hand.Size() != hand+2 {
			t.Errorf("target player's hand is %d, want %d", opp.Hand.Size(), hand+2)
		}
	})
	t.Run("gain control", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		cheap := pushCostedPermanentForTest(g, opp.ID, "Cheap Thing", "Artifact", "{1}")
		dear := pushCostedPermanentForTest(g, opp.ID, "Dear Thing", "Artifact", "{2}")
		land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
		for _, bad := range []uuid.UUID{dear, land} {
			if err := castModalStrictErr(t, g, []int{2}, cardRefs(bad)); err == nil {
				t.Fatal("Archmage's Charm took a permanent it cannot")
			}
		}
		castModal(t, g, "Archmage's Charm", "Instant", s82BnsArchmagesCharmOracle, []int{2}, cardRefs(cheap))
		passPriorityAroundTable(t, g)
		controller := uuid.Nil
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == cheap {
					controller = c.Controller
				}
			}
		})
		if controller != me.ID {
			t.Error("the permanent is not under the caster's control")
		}
		g.RunStateChecksForTest()
		advanceToNextSeatsTurn(t, g)
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == cheap && c.Controller != me.ID {
					t.Error("control ended with the turn; it is permanent")
				}
			}
		})
	})
}

func castModalStrictErr(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Archmage's Charm", TypeLine: "Instant", OracleID: s82BnsArchmagesCharmOracle,
		Owner: active.ID, Controller: active.ID})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, Strict: true})
}
