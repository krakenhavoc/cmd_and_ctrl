package testss82m

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b22_cards_test.go — S82 SNC batch 22 (#5491): copy effects, Auras
// and odd permanents. Every test plays the real catalog card through the
// engine.

const (
	sncB22Operative = "0bd57503-b544-45f8-bcd2-f5b2170010f6"
	sncB22Facetaker = "cfb41a2b-4d8b-47e3-a074-95a1bd502227"
	sncB22Party     = "b2d1f87b-a11f-43eb-8b89-b48945320d73"
	sncB22TooDeep   = "2829ae08-9964-4679-80e8-46d5d33d10b9"
	sncB22Jinnie    = "61fbaaf2-4286-4e9a-b9cb-aa31262b596a"
	sncB22Enemy     = "c5f360dc-fe4e-4796-b5c1-468cbd0f6d24"
	sncB22Ransom    = "0c9336c0-c0ec-4fb4-af7c-a7766265e927"
	sncB22Floor     = "f6828a87-eebe-48dd-873e-251624107e38"
	sncB22TurfWar   = "8e1b84d7-b0ae-4e9d-bc1a-c84832fea9f9"
)

func TestSncB22EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncB22Operative: "Undercover Operative", sncB22Facetaker: "Cephalid Facetaker",
		sncB22Party: "Life of the Party", sncB22TooDeep: "In Too Deep",
		sncB22Jinnie: "Jinnie Fay, Jetmir's Second", sncB22Enemy: "Public Enemy",
		sncB22Ransom: "Hold for Ransom", sncB22Floor: "False Floor", sncB22TurfWar: "Turf War",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func sncB22Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
}

func TestSncB22FalseFloorEntersTappedTapsEveryCreatureAndExilesTheUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	floor := effectstest.CastCatalogSpell(t, g, "False Floor", "Artifact", sncB22Floor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B31Tapped(t, g, floor) {
		t.Fatal("False Floor should enter tapped")
	}
	// A refusal: the {T} cost cannot be paid while it is tapped.
	if err := effectstest.ActivateOn(t, g, me.ID, floor, 0, uuid.Nil); err == nil {
		t.Error("activated False Floor while it was tapped")
	}
	entered := effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	if !effectstest.B31Tapped(t, g, entered) {
		t.Error("a creature should enter tapped under False Floor, its own controller's included")
	}
	standing := effectstest.B12Creature(g, g.Seats[1].ID, "Standing", "Creature — Bear", 2, 2)
	sncB22Untap(g, entered)
	sncB22Untap(g, floor)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == standing {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := effectstest.ActivateOn(t, g, me.ID, floor, 0, uuid.Nil); err != nil {
		t.Fatalf("activate False Floor: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(floor) {
		t.Error("False Floor should have been exiled as a cost")
	}
	if g.Battlefield.Contains(entered) {
		t.Error("the untapped creature should have been exiled")
	}
	if !g.Battlefield.Contains(standing) {
		t.Error("the tapped creature should survive")
	}
}

func TestSncB22PublicEnemyAimsEveryCreatureAtTheHostsControllerAndDrawsOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Public Enemy", effectstest.AuraTypeLine, sncB22Enemy, nil); err == nil {
		t.Error("cast Public Enemy with no target")
	}
	effectstest.AuraCast(t, g, "Public Enemy", sncB22Enemy, host)
	var reqs []game.AttackRequirement
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(mine)
		reqs = c.Effective().AttackRequirements
	})
	if len(reqs) != 1 || reqs[0].MustAttack != opp.ID {
		t.Fatalf("attack requirements on my creature = %+v, want one naming the host's controller", reqs)
	}
	// The host's own controller's creature cannot obey the requirement, so
	// the host is simply never forced onto its own controller.
	hand := effectstest.HandSize(me)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.B14Kill(g, host)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Fatal("the host should have died")
	}
	if effectstest.HandSize(me)-hand != 1 {
		t.Errorf("drew %d cards when the enchanted creature died, want 1", effectstest.HandSize(me)-hand)
	}
}

func TestSncB22HoldForRansomPacifiesAndPaysOffForSevenAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := effectstest.AuraCast(t, g, "Hold for Ransom", sncB22Ransom, host)
	if r := effectstest.AuraRestrictions(t, g, host); r&game.CantAttack == 0 || r&game.CantBlock == 0 {
		t.Errorf("restrictions %v, want can't attack or block", r)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := effectstest.ActivateOn(t, g, me.ID, host, 0, uuid.Nil); err == nil {
		t.Fatal("activated {7} with only six mana")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.SeedLibrary(me, "A", "B")
	hand := effectstest.HandSize(me)
	if err := effectstest.ActivateOn(t, g, me.ID, host, 0, uuid.Nil); err != nil {
		t.Fatalf("activate Hold for Ransom's {7}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("Hold for Ransom should have been sacrificed")
	}
	if !g.Battlefield.Contains(host) {
		t.Error("the creature should stay")
	}
	if effectstest.HandSize(me)-hand != 1 {
		t.Errorf("Hold for Ransom's controller drew %d, want 1", effectstest.HandSize(me)-hand)
	}
}

func TestSncB22PublicEnemyForcesAttacksAtTheHostsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.AuraCast(t, g, "Public Enemy", sncB22Enemy, host)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != mine {
		t.Fatalf("pass with my creature home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(mine, third.ID); !errors.As(err, &re) {
		t.Fatalf("attacking a different player = %v, want a refusal", err)
	}
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Fatalf("attacking the host's controller: %v", err)
	}
}

func TestSncB22InTooDeepMakesAColorlessClueThatLosesEverythingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(sncB22TooDeep)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != game.KeywordSplitSecond {
		t.Errorf("printed keywords %v, want split second", spec.PrintedKeywords)
	}
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flying Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Keywords: []string{"flying"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "In Too Deep", effectstest.AuraTypeLine, sncB22TooDeep, effectstest.B16TargetCard(land)); err == nil {
		t.Error("In Too Deep enchanted a land")
	}
	effectstest.AuraCast(t, g, "In Too Deep", sncB22TooDeep, bear)
	if types := effectstest.EffectiveTypes(t, g, bear); len(types) != 1 || types[0] != "Artifact" {
		t.Errorf("types %v, want only Artifact", types)
	}
	if subs := effectstest.EffectiveSubtypes(t, g, bear); len(subs) != 1 || subs[0] != "Clue" {
		t.Errorf("subtypes %v, want only Clue", subs)
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "flying") {
		t.Errorf("the Clue kept flying: %v", effectstest.EffectiveAbilities(t, g, bear))
	}
	var colors []string
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(bear); colors = c.Effective().Colors })
	if len(colors) != 0 {
		t.Errorf("colors %v, want colorless", colors)
	}
	effectstest.SeedLibrary(me, "A", "B")
	hand := effectstest.HandSize(me)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := effectstest.ActivateOn(t, g, me.ID, bear, 0, uuid.Nil); err != nil {
		t.Fatalf("crack the Clue: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Clue should have been sacrificed")
	}
	if effectstest.HandSize(me)-hand != 1 {
		t.Errorf("drew %d, want 1", effectstest.HandSize(me)-hand)
	}
}

func sncB22Counters(g *game.Game, id uuid.UUID, kind string) int {
	n := 0
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			n = c.Counters[kind]
		}
	})
	return n
}

func TestSncB22UndercoverOperativeShieldsOnlyACopyOfYourOwnCreature(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mine   bool
		shield int
	}{{"own creature", true, 1}, {"opponent's creature", false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			owner := opp.ID
			if tc.mine {
				owner = me.ID
			}
			target := effectstest.B12Creature(g, owner, "Grizzly Bears", "Creature — Bear", 2, 2)
			effectstest.AdvanceToMain(t, g)
			id := effectstest.CastCatalogSpell(t, g, "Undercover Operative", "Creature — Shapeshifter Rogue", sncB22Operative, nil)
			effectstest.ResolveWithCopyChoice(t, g, target)
			c := effectstest.CopyBattlefieldCard(t, g, id)
			if c.Name != "Grizzly Bears" {
				t.Errorf("entered as %q, want a copy of Grizzly Bears", c.Name)
			}
			if got := sncB22Counters(g, id, game.CounterShield); got != tc.shield {
				t.Errorf("shield counters = %d, want %d", got, tc.shield)
			}
		})
	}
}

func TestSncB22FacetakerBecomesAOneFourCopyThatCantBeBlockedUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	taker := effectstest.B12Push(g, me.ID, "Cephalid Facetaker", "Creature — Octopus Rogue", sncB22Facetaker, 1, 4)
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Drake", TypeLine: "Creature — Drake",
		Keywords: []string{"flying"}, Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	if effectstest.AuraRestrictions(t, g, taker)&game.CantBeBlocked == 0 {
		t.Error("the Facetaker should print can't be blocked")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.CardByID(g, taker)
	if c.Name != "Big Drake" {
		t.Fatalf("the Facetaker is %q, want a copy of Big Drake", c.Name)
	}
	if p, tg := effectstest.EffectivePower(t, g, taker), effectstest.EffectiveToughness(t, g, taker); p != 1 || tg != 4 {
		t.Errorf("copy is %d/%d, want 1/4", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, taker, "flying") {
		t.Error("the copy should have the Drake's flying")
	}
	if effectstest.AuraRestrictions(t, g, taker)&game.CantBeBlocked == 0 {
		t.Error("the copy should have 'can't be blocked'")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if c := effectstest.CardByID(g, taker); c.Name != "Cephalid Facetaker" {
		t.Errorf("after end of turn the Facetaker is %q, want itself again", c.Name)
	}
}

func TestSncB22FacetakerMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	taker := effectstest.B12Push(g, me.ID, "Cephalid Facetaker", "Creature — Octopus Rogue", sncB22Facetaker, 1, 4)
	victim := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.CardByID(g, taker); c.Name != "Cephalid Facetaker" {
		t.Errorf("declined, yet it is %q", c.Name)
	}
}

func TestSncB22LifeOfThePartyGivesEachOpponentAGoadedCopyAndPumpsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	party := effectstest.CastAndResolveCreature(t, g, "Life of the Party", "Creature — Elemental", sncB22Party)
	effectstest.PassPriorityAroundTable(t, g)
	copies := 0
	for _, opp := range g.Seats[1:] {
		found := false
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Life of the Party" && c.IsToken() && c.Controller == opp.ID {
					found = true
					if !c.IsGoaded() {
						t.Errorf("%s's copy is not goaded", opp.Name)
					}
				}
			}
		})
		if found {
			copies++
		}
	}
	if copies != 3 {
		t.Fatalf("%d opponents got a goaded token copy, want 3", copies)
	}
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Life of the Party" {
				n++
			}
		}
	})
	if n != 4 {
		t.Errorf("%d Life of the Party on the battlefield, want 1 + 3 tokens (a token's enters does nothing)", n)
	}
	// The attack pump: X is the creatures you control (one).
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(party, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, party); p != 1 {
		t.Errorf("power %d, want 0 + 1 creature you control", p)
	}
	_ = me
}

func TestSncB22JinnieFayOffersCatsOrDogsInstead(t *testing.T) {
	for _, tc := range []struct {
		name, want, tokenName string
		power, toughness      int
	}{{"cats", "Cat tokens", "Cat", 2, 2}, {"dogs", "Dog tokens", "Dog", 3, 1}, {"neither", "", "Soldier", 1, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Push(g, me.ID, "Jinnie Fay, Jetmir's Second", "Legendary Creature — Elf Druid", sncB22Jinnie, 3, 3)
			g.WithWriteLock(func() {
				if _, err := g.CreateTokensForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 2, game.TokenEntryOptions{}); err != nil {
					t.Fatalf("CreateTokensForEffect: %v", err)
				}
			})
			asked := 0
			for i := 0; i < 8; i++ {
				var next *game.PendingChoice
				for _, c := range g.PendingChoices {
					if c != nil && (c.Kind == game.PendingChoiceReplacementOrder || c.Kind == game.PendingChoiceOptionalReplacement) {
						next = c
						break
					}
				}
				if next == nil {
					break
				}
				if next.Kind == game.PendingChoiceReplacementOrder {
					if err := g.ResolveReplacementOrder(next.ID, me.ID, next.ReplacementEffectIDs); err != nil {
						t.Fatalf("ResolveReplacementOrder: %v", err)
					}
					continue
				}
				asked++
				apply := tc.want != "" && strings.Contains(next.Reason, tc.want)
				if err := g.ResolveOptionalReplacement(next.ID, me.ID, apply); err != nil {
					t.Fatalf("ResolveOptionalReplacement: %v", err)
				}
			}
			if tc.want == "" && asked != 2 {
				t.Errorf("asked %d questions, want both outcomes offered", asked)
			}
			got := 0
			g.ReadSnapshot(func() {
				for _, c := range g.Battlefield.Cards {
					if c.Name == tc.tokenName && c.IsToken() {
						got++
						if c.Power != tc.power || c.Toughness != tc.toughness {
							t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.Power, c.Toughness, tc.power, tc.toughness)
						}
					}
				}
			})
			if got != 2 {
				t.Errorf("%d %s tokens, want 2 (the count is kept)", got, tc.tokenName)
			}
			// The order is Cats then Dogs: a "yes" settles it, so nothing is
			// asked after it.
			if wantAsked := map[string]int{"cats": 1, "dogs": 2}[tc.name]; tc.want != "" && asked != wantAsked {
				t.Errorf("asked %d questions, want %d (none after the yes)", asked, wantAsked)
			}
		})
	}
}

func TestSncB22TurfWarMarksOneLandPerPlayerAndHandsThemToTheCreatureControllerOnDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	myLand := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	oppLand := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	oppLand2 := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	third := effectstest.B12Permanent(g, g.Seats[2].ID, "Swamp", "Basic Land — Swamp")
	fourth := effectstest.B12Permanent(g, g.Seats[3].ID, "Plains", "Basic Land — Plains")
	effectstest.CastCatalogSpell(t, g, "Turf War", "Enchantment", sncB22TurfWar, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, pick := range []uuid.UUID{myLand, oppLand, third, fourth} {
		if effectstest.LatestPickTarget(g, me.ID) == nil {
			break
		}
		effectstest.PickTriggerTarget(t, g, me.ID, pick)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"mine": myLand, "opp": oppLand, "third": third, "fourth": fourth} {
		if got := sncB22Counters(g, id, "contested"); got != 1 {
			t.Errorf("%s land has %d contested counters, want 1", name, got)
		}
	}
	if got := sncB22Counters(g, oppLand2, "contested"); got != 0 {
		t.Errorf("an unchosen land has %d contested counters", got)
	}
}

// sncB22ContestedLand puts a land with a contested counter under owner.
func sncB22ContestedLand(g *game.Game, owner uuid.UUID, name string, tapped bool) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Basic Land — Mountain",
		Counters: map[string]int{"contested": 1}, Tapped: tapped, Owner: owner, Controller: owner,
	})
}

func TestSncB22TurfWarCombatDamageStealsAndUntapsTheOnlyContestedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hitter := effectstest.B12Creature(g, me.ID, "Hitter", "Creature — Bear", 2, 2)
	effectstest.B12Push(g, me.ID, "Turf War", "Enchantment", sncB22TurfWar, 0, 0)
	land := sncB22ContestedLand(g, opp.ID, "Contested Mountain", true)
	plain := effectstest.B12Permanent(g, opp.ID, "Plain Mountain", "Basic Land — Mountain")
	effectstest.AttackWith(t, g, opp.ID, hitter)
	effectstest.PassPriorityAroundTable(t, g)
	if ctl, _ := g.ControllerOfCard(land); ctl != me.ID {
		t.Fatalf("the contested land is controlled by %s, want the creature's controller", ctl)
	}
	if effectstest.B31Tapped(t, g, land) {
		t.Error("the stolen land should be untapped")
	}
	if ctl, _ := g.ControllerOfCard(plain); ctl != opp.ID {
		t.Error("a land without a counter must not be taken")
	}
}

func TestSncB22TurfWarCreatureControllerChoosesAmongSeveralContestedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hitter := effectstest.B12Creature(g, me.ID, "Hitter", "Creature — Bear", 2, 2)
	effectstest.B12Push(g, g.Seats[2].ID, "Turf War", "Enchantment", sncB22TurfWar, 0, 0)
	a := sncB22ContestedLand(g, opp.ID, "Mountain A", false)
	b := sncB22ContestedLand(g, opp.ID, "Mountain B", false)
	effectstest.AttackWith(t, g, opp.ID, hitter)
	effectstest.PassPriorityAroundTable(t, g)
	var p *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			p = c
		}
	}
	if p == nil {
		t.Fatal("no choice was offered between the two contested lands")
	}
	if p.Chooser != me.ID {
		t.Errorf("the chooser is %s, want the damaging creature's controller (not Turf War's)", p.Chooser)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if ctl, _ := g.ControllerOfCard(b); ctl != me.ID {
		t.Error("the chosen land should have changed hands")
	}
	if ctl, _ := g.ControllerOfCard(a); ctl != opp.ID {
		t.Error("only one land is taken")
	}
	// Turf War leaving does not give the land back (no duration).
}

func TestSncB22TurfWarNoncombatDamageOrNoContestedLandDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Turf War", "Enchantment", sncB22TurfWar, 0, 0)
	land := sncB22ContestedLand(g, opp.ID, "Contested Mountain", false)
	src := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, opp.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if ctl, _ := g.ControllerOfCard(land); ctl != opp.ID {
		t.Error("noncombat damage must not take a land")
	}
	// A player with no contested land: combat damage to them does nothing.
	hitter := effectstest.B12Creature(g, me.ID, "Hitter", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, g.Seats[2].ID, hitter)
	effectstest.PassPriorityAroundTable(t, g)
	if ctl, _ := g.ControllerOfCard(land); ctl != opp.ID {
		t.Error("damage to another player must not take this player's land")
	}
}
