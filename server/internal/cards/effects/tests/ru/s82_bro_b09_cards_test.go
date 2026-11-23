package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b09_cards_test.go — The Brothers' War, batch 9 (S82, #5482): Auras,
// an Equipment and exile-until-leaves enchantments. Each test plays the real
// catalog entry through the engine.

const (
	broB09Wings        = "e3bad2e2-1024-4261-b3d7-82a6998416fd"
	broB09Audacity     = "4eb8d877-ee46-4bbe-99da-fee9e238ca58"
	broB09Destiny      = "f034137e-c19f-48af-a2f2-f767bf87a275"
	broB09Discipline   = "c910c1c0-2b78-4e0f-bb55-e0581b2e07e8"
	broB09Domination   = "ede2354a-64b0-402c-8fbe-9da90f811a9b"
	broB09Sentence     = "6d9f2fe0-5657-45b4-9431-36e0f6c8eba2"
	broB09TakeFlight   = "b4cf61e5-b2b7-4fb9-addf-db84dec49681"
	broB09Subjugation  = "4c3745ff-6e1e-4660-abe6-1a30611ec00c"
	broB09Cooldown     = "4640ce0e-f82f-4bfd-93fc-2d865863fc2e"
	broB09Trenches     = "1612a08f-fc1f-45cf-ae03-5f021a46869d"
	broB09Intervention = "c1a7d108-8073-42a5-ade0-a0643d078268"
)

func TestBroB09EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		broB09Wings: "Aeronaut's Wings", broB09Audacity: "Audacity", broB09Destiny: "Draconic Destiny",
		broB09Discipline: "Military Discipline", broB09Domination: "Mishra's Domination",
		broB09Sentence: "Prison Sentence", broB09TakeFlight: "Take Flight",
		broB09Subjugation: "Weakstone's Subjugation", broB09Cooldown: "Involuntary Cooldown",
		broB09Trenches: "In the Trenches", broB09Intervention: "Ashnod's Intervention",
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

func broB09Pow(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id)
}

func TestBroB09AeronautsWingsGiveFlyingAndPowerOnceEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	wings := effectstest.PushCatalogPermanent(g, me.ID, "Aeronaut's Wings", "Artifact — Equipment", broB09Wings, false)
	if p, _ := broB09Pow(t, g, bear); p != 2 || effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Fatal("an unequipped Bear should be a plain 2/2")
	}
	effectstest.FillPool(me, 2)
	effectstest.EquipTo(t, g, me.ID, wings, bear)
	if p, tg := broB09Pow(t, g, bear); p != 3 || tg != 2 {
		t.Errorf("equipped Bear is %d/%d, want 3/2", p, tg)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("equipped Bear lacks flying")
	}
}

func TestBroB09EquipCostsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	wings := effectstest.PushCatalogPermanent(g, me.ID, "Aeronaut's Wings", "Artifact — Equipment", broB09Wings, false)
	effectstest.FillPool(me, 1)
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: broB09Wings})
	if err := g.ActivateCatalogAbility(me.ID, wings, len(abilities)-1, game.ActivateAbilityParams{
		Targets: effectstest.B16TargetCard(bear), Strict: true,
	}); err == nil {
		t.Error("equipped for {1}")
	}
}

func TestBroB09AudacityPumpsTramplesAndDrawsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Audacity", effectstest.AuraTypeLine, broB09Audacity, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := broB09Pow(t, g, bear); p != 4 || tg != 2 {
		t.Errorf("enchanted Bear is %d/%d, want 4/2", p, tg)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("enchanted Bear lacks trample")
	}
	before := len(me.Hand.Cards)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, aura); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand went %d -> %d, want one card drawn", before, got)
	}
}

func TestBroB09AudacityDrawsWhenTheCreatureDiesToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Audacity", effectstest.AuraTypeLine, broB09Audacity, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	before := len(me.Hand.Cards)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand went %d -> %d, want one card drawn", before, got)
	}
}

func TestBroB09TakeFlightGivesTheCreatureAnAttackDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Take Flight", effectstest.AuraTypeLine, broB09TakeFlight, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := broB09Pow(t, g, bear); p != 3 || tg != 2 {
		t.Errorf("enchanted Bear is %d/%d, want 3/2", p, tg)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("enchanted Bear lacks flying")
	}
	before := len(me.Hand.Cards)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand went %d -> %d, want one card drawn on attack", before, got)
	}
}

func TestBroB09PrisonSentenceStopsAttackBlockAndActivations(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Prison Sentence", effectstest.AuraTypeLine, broB09Sentence, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	broB09AnswerScry(t, g, me.ID)
	var r game.Restriction
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		r = effectstest.FindBattlefieldCardByID(g, bear).Effective().Restrictions
	})
	for _, want := range []game.Restriction{game.CantAttackOrBlock, game.CantActivate} {
		if r&want != want {
			t.Errorf("restrictions %v lack %v", r, want)
		}
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: bear, Target: opp.ID}}); err == nil {
		t.Error("the enchanted creature attacked")
	}
}

func TestBroB09PrisonSentenceScriesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Prison Sentence", effectstest.AuraTypeLine, broB09Sentence, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("no scry prompt for the Aura's controller")
	}
}

func TestBroB09MilitaryDisciplineGrantsFirstStrikeOnlyThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Military Discipline", effectstest.AuraTypeLine, broB09Discipline, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := broB09Pow(t, g, bear); p != 3 {
		t.Errorf("enchanted Bear has power %d, want 3", p)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("enchanted Bear lacks first strike this turn")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("first strike outlived the turn")
	}
	if p, _ := broB09Pow(t, g, bear); p != 3 {
		t.Errorf("the +1/+0 should stay: power %d", p)
	}
}

func TestBroB09MilitaryDisciplineHasFlash(t *testing.T) {
	spec, _ := effects.Lookup(broB09Discipline)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestBroB09MishrasDominationPumpsOnlyWhileYouControlTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", effectstest.TestCreatureTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Mishra's Domination", effectstest.AuraTypeLine, broB09Domination, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Mishra's Domination", effectstest.AuraTypeLine, broB09Domination, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	restr := func(id uuid.UUID) game.Restriction {
		var r game.Restriction
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			r = effectstest.FindBattlefieldCardByID(g, id).Effective().Restrictions
		})
		return r
	}
	if p, tg := broB09Pow(t, g, mine); p != 4 || tg != 4 {
		t.Errorf("my enchanted creature is %d/%d, want 4/4", p, tg)
	}
	if restr(mine)&game.CantBlock != 0 {
		t.Error("my enchanted creature can't block")
	}
	if p, tg := broB09Pow(t, g, theirs); p != 2 || tg != 2 {
		t.Errorf("their enchanted creature is %d/%d, want 2/2", p, tg)
	}
	if restr(theirs)&game.CantBlock == 0 {
		t.Error("their enchanted creature can block")
	}
	// It changes hands: the halves swap.
	g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, theirs, me.ID, game.IndefiniteDuration(), "theft") })
	if p, tg := broB09Pow(t, g, theirs); p != 4 || tg != 4 {
		t.Errorf("after the swap it is %d/%d, want 4/4", p, tg)
	}
	if restr(theirs)&game.CantBlock != 0 {
		t.Error("after the swap it still can't block")
	}
}

func TestBroB09WeakstonesSubjugationPaidTapsAndKeepsItTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Theirs", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.FillPool(me, 3)
	effectstest.CastCatalogSpell(t, g, "Weakstone's Subjugation", effectstest.AuraTypeLine, broB09Subjugation, effectstest.B16TargetCard(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, victim) {
		t.Fatal("paid {3} but the creature is untapped")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("it untapped in its controller's untap step")
	}
}

func TestBroB09WeakstonesSubjugationUnpaidDoesNotTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Theirs", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Weakstone's Subjugation", effectstest.AuraTypeLine, broB09Subjugation, effectstest.B16TargetCard(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, victim) {
		t.Error("tapped without paying")
	}
}

func TestBroB09WeakstonesSubjugationEnchantsAnArtifactButNotALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Weakstone's Subjugation", effectstest.AuraTypeLine, broB09Subjugation, effectstest.B16TargetCard(land)); err == nil {
		t.Error("enchanted a land")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Weakstone's Subjugation", effectstest.AuraTypeLine, broB09Subjugation, effectstest.B16TargetCard(rock)); err != nil {
		t.Errorf("could not enchant an artifact: %v", err)
	}
}

func TestBroB09InvoluntaryCooldownTapsAndStunsTwoTargetsTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Involuntary Cooldown", "Sorcery", broB09Cooldown, effectstest.S58Refs(bear, rock))
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"creature": bear, "artifact": rock} {
		if !effectstest.B16Tapped(t, g, id) {
			t.Errorf("the %s is untapped", name)
		}
		if n := effectstest.CountersOn(g, id, game.CounterStun); n != 2 {
			t.Errorf("the %s has %d stun counters, want 2", name, n)
		}
	}
}

func TestBroB09InvoluntaryCooldownRefusesALandAndAcceptsNoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Involuntary Cooldown", "Sorcery", broB09Cooldown, effectstest.B16TargetCard(land)); err == nil {
		t.Error("targeted a land")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Involuntary Cooldown", "Sorcery", broB09Cooldown, nil); err != nil {
		t.Errorf("cast with no targets: %v", err)
	}
}

func broB09Trenches_(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return effectstest.PushCatalogPermanent(g, me.ID, "In the Trenches", "Enchantment", broB09Trenches, false)
}

func broB09Activate(g *game.Game, me *game.Player, src, target uuid.UUID) error {
	idx := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: broB09Trenches})) - 1
	return g.ActivateCatalogAbility(me.ID, src, idx, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(target), Strict: true})
}

func TestBroB09InTheTrenchesPumpsYourCreaturesNotTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", effectstest.TestCreatureTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", effectstest.TestCreatureTypeLine, 2, 2)
	broB09Trenches_(t, g, me)
	if p, tg := broB09Pow(t, g, mine); p != 3 || tg != 3 {
		t.Errorf("mine is %d/%d, want 3/3", p, tg)
	}
	if p, tg := broB09Pow(t, g, theirs); p != 2 || tg != 2 {
		t.Errorf("theirs is %d/%d, want 2/2", p, tg)
	}
}

func TestBroB09InTheTrenchesExilesOnceUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", effectstest.TestCreatureTypeLine, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", effectstest.TestCreatureTypeLine, 2, 2)
	src := broB09Trenches_(t, g, me)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPool(me, 11)
	if err := broB09Activate(g, me, src, theirs); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the target was not exiled")
	}
	if err := broB09Activate(g, me, src, other); err == nil {
		t.Error("activated a second time")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(src) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldNamed(g, opp.ID, "Theirs") != 1 {
		t.Error("the exiled permanent did not return when In the Trenches left")
	}
}

func TestBroB09InTheTrenchesIsSorceryOnlyAndNeverTargetsYourOwnOrLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", effectstest.TestCreatureTypeLine, 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", effectstest.TestCreatureTypeLine, 2, 2)
	src := broB09Trenches_(t, g, me)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPool(me, 5)
	if err := broB09Activate(g, me, src, mine); err == nil {
		t.Error("exiled a permanent you control")
	}
	if err := broB09Activate(g, me, src, land); err == nil {
		t.Error("exiled a land")
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := broB09Activate(g, me, src, theirs); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestBroB09AshnodsInterventionPumpsAndReturnsOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ashnod's Intervention", "Instant", broB09Intervention, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := broB09Pow(t, g, bear); p != 4 || tg != 2 {
		t.Errorf("Bear is %d/%d, want 4/2", p, tg)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestBroB09AshnodsInterventionReturnsAnExiledCreatureButNotABouncedOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	exiled := effectstest.B12Creature(g, me.ID, "Exiled", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ashnod's Intervention", "Instant", broB09Intervention, effectstest.B16TargetCard(exiled))
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneExile}, exiled); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(exiled) {
		t.Error("the exiled creature did not return to its owner's hand")
	}
}

func TestBroB09AshnodsInterventionWearsOffAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ashnod's Intervention", "Instant", broB09Intervention, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	if p, _ := broB09Pow(t, g, bear); p != 2 {
		t.Errorf("the pump outlived the turn: power %d", p)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bear) {
		t.Error("the return outlived the turn")
	}
}

func TestBroB09DraconicDestinyGrantsFlyingHasteDragonAndFirebreathing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Draconic Destiny", effectstest.AuraTypeLine, broB09Destiny, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := broB09Pow(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("enchanted Bear is %d/%d, want 3/3", p, tg)
	}
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("enchanted Bear lacks %s", kw)
		}
	}
	var dragon, bearType bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c := effectstest.FindBattlefieldCardByID(g, bear)
		dragon, bearType = c.HasSubtype("Dragon"), c.HasSubtype("Bear")
	})
	if !dragon || !bearType {
		t.Errorf("Dragon %v, Bear %v: it should be both", dragon, bearType)
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the creature has no granted activated ability")
	}
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := broB09Pow(t, g, bear); p != 4 {
		t.Errorf("after {1}: power %d, want 4", p)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("Draconic Destiny did not return to its owner's hand")
	}
}

// broB09AnswerScry keeps every card of the pending scry on top.
func broB09AnswerScry(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == player && c.Kind == game.PendingChoiceScry {
			if err := g.ResolveScry(c.ID, player, nil, c.ScryCards); err != nil {
				t.Fatalf("ResolveScry: %v", err)
			}
			return
		}
	}
	t.Fatal("no scry prompt")
}
