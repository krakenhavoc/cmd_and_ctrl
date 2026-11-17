package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b04_cards_test.go — March of the Machine, S82 batch 4 (#5477).
// Each test plays the real catalog entry and asserts what its text says.

const (
	s82ArchpriestOracle = "7b03910a-2ca7-40cb-9c71-1c2cc856a287"
	s82BolaOracle       = "6ff995f9-080d-4316-8fbe-39bf195db91b"
	s82BrightPalmOracle = "f74888ee-5086-49a6-ae74-fd9894f46332"
	s82SledgeOracle     = "e09e7bb8-2665-4081-b75a-628a8bc0a3cd"
	s82DoomskarOracle   = "7ac00312-8963-4fee-8561-1a54e533adbb"
	s82WoodwurmOracle   = "a74e41b4-6284-453c-a586-d575735dca65"
	s82ScroungerOracle  = "d1b74fb9-8ef3-40be-bc79-45ffb4c597eb"
)

func TestS82MomB04EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		s82ArchpriestOracle: "Archpriest of Shadows",
		s82BolaOracle:       "Bola Slinger",
		s82BrightPalmOracle: "Bright-Palm, Soul Awakener",
		s82SledgeOracle:     "Conclave Sledge-Captain",
		s82DoomskarOracle:   "Doomskar Warrior",
		s82WoodwurmOracle:   "Emergent Woodwurm",
		s82ScroungerOracle:  "Hangar Scrounger",
	}
	for oracle, name := range want {
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

// s82BackupOnto casts the card and aims its backup at `target`.
func s82BackupOnto(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tough int, target func(self uuid.UUID) uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, target(id))
	passPriorityAroundTable(t, g)
	return id
}

func s82Library(p *game.Player, cards ...[3]string) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		ids[i] = uuid.New()
		p.Library.PushTop(game.Card{InstanceID: ids[i], Name: cards[i][0], TypeLine: cards[i][1], ManaCost: cards[i][2],
			Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	}
	return ids
}

// --- Archpriest of Shadows ---

func TestS82MomB04ArchpriestBackupGrantsDeathtouchAndTheReturn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dead := socB6Graveyard(me, "Dead Ogre", "Creature — Ogre", "{3}")
	rock := socB6Graveyard(me, "Dead Rock", "Artifact", "{1}")
	theirs := socB6Graveyard(opp, "Their Dead", "Creature — Ogre", "{3}")
	arch := s82BackupOnto(t, g, "Archpriest of Shadows", "Creature — Human Warlock", s82ArchpriestOracle, 4, 4,
		func(uuid.UUID) uuid.UUID { return bear })
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Bear has %d counters, want 1", n)
	}
	if countersOn(g, arch, game.CounterPlusOne) != 0 {
		t.Error("the Archpriest got the counter")
	}
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("backup did not give the Bear deathtouch")
	}
	dealCombatDamageToPlayer(g, bear, opp.ID, 3)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the granted trigger asked for no target")
	}
	for _, bad := range []uuid.UUID{rock, theirs} {
		if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bad}}); err == nil {
			t.Errorf("an illegal card %v was accepted", bad)
		}
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: dead}}); err != nil {
		t.Fatalf("pick the Ogre: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Error("the creature card did not return")
	}
	advanceToEndStepOf(t, g, 0)
	advanceToUpkeepOf(t, g, 1)
	if hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("the backup grant outlived the turn")
	}
}

func TestS82MomB04ArchpriestOnItselfHasItsOwnAbilitiesAndTriggersOnBattleDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := socB6Graveyard(me, "Dead Ogre", "Creature — Ogre", "{3}")
	arch := s82BackupOnto(t, g, "Archpriest of Shadows", "Creature — Human Warlock", s82ArchpriestOracle, 4, 4,
		func(self uuid.UUID) uuid.UUID { return self })
	if n := countersOn(g, arch, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Archpriest has %d counters, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, arch, "deathtouch") {
		t.Error("the Archpriest lacks deathtouch")
	}
	battle := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Invasion", TypeLine: "Battle — Siege",
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	dealCombatDamageToPlayer(g, arch, battle, 5)
	answerTriggerTargets(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Error("damage to a battle did not trigger the return")
	}
}

// --- Bola Slinger ---

func TestS82MomB04BolaSlingerBackupGrantsTheAttackTap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID})
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	slinger := s82BackupOnto(t, g, "Bola Slinger", "Creature — Cat Soldier", s82BolaOracle, 2, 2,
		func(uuid.UUID) uuid.UUID { return bear })
	if countersOn(g, bear, game.CounterPlusOne) != 1 || countersOn(g, slinger, game.CounterPlusOne) != 0 {
		t.Fatal("the counter did not go on the Bear alone")
	}
	declareAttack(t, g, opp.ID, bear)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the granted attack trigger asked for no target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Error("a permanent its controller owns was accepted")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err != nil {
		t.Fatalf("pick the artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !tappedOnBattlefield(t, g, rock) {
		t.Error("the artifact is not tapped")
	}
	if tappedOnBattlefield(t, g, theirs) {
		t.Error("an unchosen creature was tapped")
	}
}

func TestS82MomB04BolaSlingerOwnAttackTapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	slinger := b12Push(g, me.ID, "Bola Slinger", "Creature — Cat Soldier", s82BolaOracle, 2, 2)
	declareAttack(t, g, opp.ID, slinger)
	answerTriggerTargets(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !tappedOnBattlefield(t, g, theirs) {
		t.Error("the opposing creature is not tapped")
	}
}

// --- Bright-Palm, Soul Awakener ---

func TestS82MomB04BrightPalmBackupGrantsTheDoublingAndEvasion(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	small := b12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big Ogre", "Creature — Ogre", 3, 3)
	s82BackupOnto(t, g, "Bright-Palm, Soul Awakener", "Legendary Creature — Fox Shaman", s82BrightPalmOracle, 4, 3,
		func(uuid.UUID) uuid.UUID { return bear })
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Bear has %d counters, want 1", n)
	}
	declareAttack(t, g, opp.ID, bear)
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Fatalf("the Bear has %d counters after doubling, want 2", n)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	brRefusal(t, g.DeclareBlocker(small, bear), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, bear); err != nil {
		t.Errorf("a creature with power 3 blocks it: %v", err)
	}
}

func TestS82MomB04BrightPalmDoublesAnotherCreatureAndLeavesNoCountersAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	palm := b12Push(g, me.ID, "Bright-Palm, Soul Awakener", "Legendary Creature — Fox Shaman", s82BrightPalmOracle, 4, 3)
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ally {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 3}
			}
		}
	})
	declareAttack(t, g, opp.ID, palm)
	answerTriggerTargets(t, g, me.ID, ally)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, ally, game.CounterPlusOne); n != 6 {
		t.Errorf("the ally has %d counters, want 6", n)
	}
	if n := countersOn(g, palm, game.CounterPlusOne); n != 0 {
		t.Errorf("Bright-Palm has %d counters, want 0", n)
	}
}

// --- Conclave Sledge-Captain ---

func TestS82MomB04SledgeCaptainHasThreeBackupsAndGrowsByItsDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spec, _ := Lookup(s82SledgeOracle)
	backups := 0
	for _, tr := range spec.Triggered {
		if tr.Keyword == game.KeywordBackup {
			backups++
		}
	}
	if backups != 3 {
		t.Fatalf("%d backup abilities, want 3", backups)
	}
	captain := b12Push(g, me.ID, "Conclave Sledge-Captain", "Creature — Elephant Soldier", s82SledgeOracle, 4, 4)
	dealCombatDamageToPlayer(g, captain, opp.ID, 4)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, captain, game.CounterPlusOne); n != 4 {
		t.Errorf("the Captain has %d counters, want 4 (that many)", n)
	}
	if !hasEffectiveKeyword(t, g, captain, "trample") {
		t.Error("the Captain lacks trample")
	}
}

func TestS82MomB04SledgeCaptainBackupsOnAnotherCreatureGrantTheTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Conclave Sledge-Captain", TypeLine: "Creature — Elephant Soldier",
		OracleID: s82SledgeOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		answerTriggerTargets(t, g, me.ID, bear)
	}
	answerTriggerOrderInOfferedOrder(t, g)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 3 {
		t.Fatalf("the Bear has %d counters, want 3 (three backups)", n)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
	dealCombatDamageToPlayer(g, bear, opp.ID, 5)
	passPriorityAroundTable(t, g)
	// Three backups grant three instances of the trigger (CR 702.165), each
	// putting "that many" counters: 3 + 3 x 5.
	if n := countersOn(g, bear, game.CounterPlusOne); n != 18 {
		t.Errorf("the Bear has %d counters, want 18", n)
	}
	if countersOn(g, id, game.CounterPlusOne) != 0 {
		t.Error("the Captain itself got a counter")
	}
}

// --- Doomskar Warrior ---

func TestS82MomB04DoomskarWarriorLooksAtThatManyAndTakesACreatureOrLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	top := s82Library(me, [3]string{"Shock", "Instant", "{R}"}, [3]string{"Elf", "Creature — Elf", "{G}"},
		[3]string{"Forest", "Basic Land — Forest", ""}, [3]string{"Beyond", "Creature — Ogre", "{4}"})
	warrior := b12Push(g, me.ID, "Doomskar Warrior", "Creature — Human Warrior", s82DoomskarOracle, 4, 3)
	dealCombatDamageToPlayer(g, warrior, opp.ID, 3)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt for the look")
	}
	want := map[uuid.UUID]bool{top[1]: true, top[2]: true}
	if len(prompt.ChooseCards) != 2 {
		t.Fatalf("offered %d cards, want the creature and land among the top three", len(prompt.ChooseCards))
	}
	for _, id := range prompt.ChooseCards {
		if !want[id] {
			t.Errorf("offered %s, which is not among the top three creature or land cards", id)
		}
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[1]) {
		t.Error("the chosen creature is not in hand")
	}
	if me.Library.Contains(top[3]) == false {
		t.Error("the fourth card should still be in the library")
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:2] {
		bottom[c.InstanceID] = true
	}
	if !bottom[top[0]] || !bottom[top[2]] {
		t.Error("the unchosen cards are not on the bottom")
	}
}

func TestS82MomB04DoomskarWarriorBackupGrantsTrampleAndTheLook(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	top := s82Library(me, [3]string{"Elf", "Creature — Elf", "{G}"}, [3]string{"Shock", "Instant", "{R}"})
	s82BackupOnto(t, g, "Doomskar Warrior", "Creature — Human Warrior", s82DoomskarOracle, 4, 3,
		func(uuid.UUID) uuid.UUID { return bear })
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatal("the Bear did not gain trample")
	}
	dealCombatDamageToPlayer(g, bear, opp.ID, 1)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil || len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != top[0] {
		t.Fatalf("want a look at one card offering the Elf, got %+v", prompt)
	}
}

// --- Emergent Woodwurm ---

func TestS82MomB04WoodwurmLooksAtXAndPutsAPermanentOfManaValueXOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	top := s82Library(me, [3]string{"Shock", "Instant", "{R}"}, [3]string{"Big", "Artifact", "{5}"},
		[3]string{"Cheap", "Artifact", "{3}"}, [3]string{"Pricey", "Artifact", "{6}"},
		[3]string{"Out of Reach", "Artifact", "{1}"})
	worm := b12Push(g, me.ID, "Emergent Woodwurm", "Creature — Wurm", s82WoodwurmOracle, 4, 4)
	declareAttack(t, g, opp.ID, worm)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt for the look at the top four")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != top[2] {
		t.Fatalf("offered %v, want only the Cheap artifact (the top four, mana value 4 or less, a permanent)", prompt.ChooseCards)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(top[2]) {
		t.Error("the chosen card did not enter the battlefield")
	}
	if !me.Library.Contains(top[4]) {
		t.Error("the fifth card must not have been looked at or moved")
	}
}

func TestS82MomB04WoodwurmBackupThreeGrantsTheAttackTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	top := s82Library(me, [3]string{"Cheap", "Artifact", "{2}"}, [3]string{"Pricey", "Artifact", "{6}"},
		[3]string{"Filler", "Artifact", "{1}"}, [3]string{"Filler2", "Artifact", "{1}"}, [3]string{"Filler3", "Artifact", "{1}"})
	s82BackupOnto(t, g, "Emergent Woodwurm", "Creature — Wurm", s82WoodwurmOracle, 4, 4,
		func(uuid.UUID) uuid.UUID { return bear })
	if n := countersOn(g, bear, game.CounterPlusOne); n != 3 {
		t.Fatalf("the Bear has %d counters, want 3", n)
	}
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("the granted attack trigger asked for nothing")
	}
	// The Bear is 5/5: X = 5, so the top five are looked at and the
	// mana value 6 card is out of range.
	got := map[uuid.UUID]bool{}
	for _, id := range prompt.ChooseCards {
		got[id] = true
	}
	if got[top[1]] || !got[top[0]] || len(got) != 4 {
		t.Errorf("offered %v, want every artifact of the top five except the mana value 6 one", prompt.ChooseCards)
	}
}

// --- Hangar Scrounger ---

func TestS82MomB04ScroungerRummagesWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	fodder := handCard(me, "Fodder", "Instant")
	scrounger := b12Push(g, me.ID, "Hangar Scrounger", "Creature — Dwarf Pilot", s82ScroungerOracle, 2, 1)
	declareAttack(t, g, opp.ID, scrounger)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt when the Scrounger became tapped")
	}
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, fodder)
	if !me.Graveyard.Contains(fodder) || me.Hand.Size() != hand {
		t.Errorf("want the Fodder discarded and a card drawn (hand %d -> %d)", hand, me.Hand.Size())
	}
}

func TestS82MomB04ScroungerDrawsNothingWithoutADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	handCard(me, "Kept", "Instant")
	scrounger := b12Push(g, me.ID, "Hangar Scrounger", "Creature — Dwarf Pilot", s82ScroungerOracle, 2, 1)
	declareAttack(t, g, opp.ID, scrounger)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no change when nothing is discarded", hand, me.Hand.Size())
	}
}

func TestS82MomB04ScroungerBackupGrantsTheTapTriggerForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "A", "B")
	fodder := handCard(me, "Fodder", "Instant")
	s82BackupOnto(t, g, "Hangar Scrounger", "Creature — Dwarf Pilot", s82ScroungerOracle, 2, 1,
		func(uuid.UUID) uuid.UUID { return bear })
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Bear has %d counters, want 1", n)
	}
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("the granted tap trigger asked for no discard")
	}
	answerDiscard(t, g, me.ID, fodder)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the Fodder was not discarded")
	}
}
