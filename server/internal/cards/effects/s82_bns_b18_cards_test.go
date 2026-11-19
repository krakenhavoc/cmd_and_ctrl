package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b18_cards_test.go — S82, Ravnica: Clue Edition batch 18 (#5929).

const (
	s82bns18ShieldmateOracle = "03e04034-48b5-4f1f-9859-2b46dedf4d8b"
	s82bns18VampireOracle    = "fcedc288-5e27-45a2-a07b-06c7aa7a9230"
	s82bns18ScreecherOracle  = "4617739e-4e3e-4d06-b074-8f836261a870"
	s82bns18WatcherOracle    = "e8c0fbea-a678-46b0-b385-e36f14c07672"
	s82bns18AgentOracle      = "2ed76a99-fe5c-4d68-b572-a30848055c8a"
	s82bns18PatrolOracle     = "4fb9c040-c684-4ba8-970e-58313af5584a"
	s82bns18BodyguardOracle  = "8de57b66-cb97-44e4-8c0d-8ecf3202c299"
	s82bns18ForcemageOracle  = "785e451d-87b6-4772-97e5-d2744818e0fd"
	s82bns18ChainsOracle     = "be951382-5173-4fb6-9dda-2c5f7fddc863"
	s82bns18ImpetusOracle    = "87f22ec6-6668-4e15-b38c-fa1ac0292c6b"
	s82bns18InsuranceOracle  = "05501e88-d4c3-4474-92a9-c02ab15b107b"
	s82bns18BandOracle       = "a8631e87-bfc3-463b-bdd1-974ea1d79ae4"
)

func TestS82Bns18CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82bns18ShieldmateOracle: "Vernadi Shieldmate",
		s82bns18VampireOracle:    "Vizkopa Vampire",
		s82bns18ScreecherOracle:  "War Screecher",
		s82bns18WatcherOracle:    "Watcher in the Mist",
		s82bns18AgentOracle:      "Whisper Agent",
		s82bns18PatrolOracle:     "Wildwood Patrol",
		s82bns18BodyguardOracle:  "Wojek Bodyguard",
		s82bns18ForcemageOracle:  "Yeva's Forcemage",
		s82bns18ChainsOracle:     "Curse of Chains",
		s82bns18ImpetusOracle:    "Predatory Impetus",
		s82bns18InsuranceOracle:  "Afterlife Insurance",
		s82bns18BandOracle:       "Band Together",
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

func TestS82Bns18PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Vernadi Shieldmate", "Creature — Human Soldier", s82bns18ShieldmateOracle, 2, 2, []string{"vigilance"}},
		{"Vizkopa Vampire", "Creature — Vampire", s82bns18VampireOracle, 3, 1, []string{"lifelink"}},
		{"War Screecher", "Creature — Bird", s82bns18ScreecherOracle, 1, 3, []string{"flying"}},
		{"Watcher in the Mist", "Creature — Spirit", s82bns18WatcherOracle, 3, 4, []string{"flying"}},
		{"Whisper Agent", "Creature — Human Rogue", s82bns18AgentOracle, 3, 2, []string{"flash"}},
		{"Wildwood Patrol", "Creature — Centaur Scout", s82bns18PatrolOracle, 4, 2, []string{"trample"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keywords...)
	}
}

func TestS82Bns18WarScreecherPumpsTheOthersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := b12Push(g, me.ID, "War Screecher", "Creature — Bird", s82bns18ScreecherOracle, 1, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{W}{W}{W}{W}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("the other creature is %d/%d, want 3/3", p, tg)
	}
	if p := effectivePower(t, g, bird); p != 1 {
		t.Errorf("War Screecher has power %d, want 1", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
	if !b16Tapped(t, g, bird) {
		t.Error("the {T} cost did not tap it")
	}
	// Tapped, and with no mana left, a second activation is refused.
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a tapped War Screecher with no mana activated again")
	}
}

func s82bns18Surveil(t *testing.T, name, typeLine, oracle string, n int) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top A", "Top B", "Third")
	castAndResolveCreature(t, g, name, typeLine, oracle)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("%s entering did not queue a surveil", name)
	}
	if len(c.ScryCards) != n {
		t.Fatalf("looked at %d cards, want %d", len(c.ScryCards), n)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{c.ScryCards[0]}, c.ScryCards[1:]); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := graveyardNames(me); len(got) != 1 || got[0] != "Top A" {
		t.Errorf("graveyard = %v, want [Top A]", got)
	}
}

func TestS82Bns18WatcherInTheMistSurveilsTwo(t *testing.T) {
	s82bns18Surveil(t, "Watcher in the Mist", "Creature — Spirit", s82bns18WatcherOracle, 2)
}

func TestS82Bns18WhisperAgentSurveilsOne(t *testing.T) {
	s82bns18Surveil(t, "Whisper Agent", "Creature — Human Rogue", s82bns18AgentOracle, 1)
}

func TestS82Bns18WojekBodyguardCantAttackAloneButMentors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := b12Push(g, me.ID, "Wojek Bodyguard", "Creature — Human Soldier", s82bns18BodyguardOracle, 3, 3)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(guard, opp.ID); err != nil {
		t.Fatalf("declaring the Bodyguard first was refused: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the Bodyguard alone: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(small, opp.ID); err != nil {
		t.Fatalf("the companion was refused: %v", err)
	}
	lockInAttacks(t, g)
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if got := allCountersOn(t, g, small)["+1/+1"]; got != 1 {
		t.Errorf("mentor put %d +1/+1 counters on the smaller attacker, want 1", got)
	}
}

func TestS82Bns18WojekBodyguardCantBlockAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, def := g.Seats[0], g.Seats[1]
	atk := pushVanillaCreature(g, me.ID, "Raider", 2, 2)
	guard := apaPush(g, def.ID, def.ID, apaCreature("Wojek Bodyguard", s82bns18BodyguardOracle, 3, 3))
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: atk, Target: def.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	err := g.DeclareBlocker(guard, atk)
	var re *game.BlockRefusedError
	if !errors.As(err, &re) || re.Reason != game.BlockReasonCantBlockAlone {
		t.Fatalf("blocking alone: %v, want cant_block_alone", err)
	}
}

func TestS82Bns18YevasForcemageGivesATargetCreaturePlusTwoPlusTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castAndResolveCreature(t, g, "Yeva's Forcemage", "Creature — Elf Shaman", s82bns18ForcemageOracle)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tg)
	}
}

func TestS82Bns18CurseOfChainsTapsTheCreatureEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := auraBear(g, opp.ID)
	auraCast(t, g, "Curse of Chains", s82bns18ChainsOracle, theirs)
	if b16Tapped(t, g, theirs) {
		t.Fatal("the creature was tapped before any upkeep")
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Error("the enchanted creature is untapped after the upkeep trigger")
	}
}

func TestS82Bns18PredatoryImpetusPumpsGoadsAndForcesABlock(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	auraCast(t, g, "Predatory Impetus", s82bns18ImpetusOracle, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Errorf("enchanted bear is %d/%d, want 5/5", p, tg)
	}
	c, ok := g.LookupCardForEffect(bear)
	if !ok || !c.Goaded() {
		t.Error("the enchanted creature is not goaded")
	}
	var reqs []game.BlockRequirement
	g.ReadSnapshot(func() {
		if cc := findBattlefieldCardForTest(g, bear); cc != nil {
			reqs = cc.Effective().BlockRequirements
		}
	})
	found := false
	for _, r := range reqs {
		if r.Kind == game.BlockRequirementMustBeBlocked {
			found = true
		}
	}
	if !found {
		t.Errorf("no must-be-blocked requirement on the enchanted creature: %+v", reqs)
	}
}

func TestS82Bns18AfterlifeInsuranceGrantsAfterlifeThisTurnAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Afterlife Insurance", "Instant", s82bns18InsuranceOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn)", me.Hand.Size(), hand+1)
	}
	killCreature(t, g, g.Seats[1].ID, theirs)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, g.Seats[1].ID, "Spirit"); got != 0 {
		t.Errorf("an opponent's creature left %d Spirits", got)
	}
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Spirit"); got != 1 {
		t.Fatalf("%d Spirit tokens, want 1", got)
	}
	s := mshB1Token(t, g, me.ID, "Spirit")
	if s.Power != 1 || s.Toughness != 1 || len(s.Colors) != 2 || !hasKeywordOnBattlefield(t, g, s.InstanceID, "flying") {
		t.Errorf("token = %d/%d %v, want a 1/1 white and black flier", s.Power, s.Toughness, s.Colors)
	}
}

func TestS82Bns18BandTogetherDealsDamageEqualToPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Band Together", "Instant", s82bns18BandOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("two 2-power creatures did not kill a 4/4")
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	a = b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	victim = b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Band Together", "Instant", s82bns18BandOracle, s82b12Pair(a, victim))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("one 2-power creature killed a 4/4")
	}

	// The victim is "another" creature: a source is no legal second target.
	g = newCatalogGame(t)
	a = b12Creature(g, g.Seats[0].ID, "Bear A", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Band Together", "Instant", s82bns18BandOracle, s82b12Pair(a, a)); err == nil {
		t.Error("a creature was both a source and the other target")
	}
}
