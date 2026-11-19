package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bnsB14RacketeersOracle = "4f6c9cee-5af7-4b00-a2a4-2770dab0a2dc"
	bnsB14PatrolOracle     = "d4891bfc-12b7-4412-be11-d02e6d014463"
	bnsB14PasswallOracle   = "9a2bc266-4076-4923-b201-5898b18cbad9"
	bnsB14GorgonOracle     = "4704ce78-b287-4038-abdd-ddb20116874d"
	bnsB14PortalOracle     = "0dd3127e-de13-4785-b5f2-0af996e9f15f"
	bnsB14PyrewildOracle   = "a1e71473-3eba-4d29-87e2-7f14acab2ccd"
	bnsB14ShredFreakOracle = "e900b2b5-4bb2-482d-89ee-b5bbaf246698"
	bnsB14TrumpeterOracle  = "2047a613-809d-4a47-9c78-f58a380ba18d"
	bnsB14SphinxOracle     = "d4c4ad11-7e71-4ad0-aee7-58ac7b1b38f8"
	bnsB14TechnicianOracle = "ec8a4b26-633a-4e88-aa8a-82e9704b3439"
	bnsB14PredatorOracle   = "4c297085-5ec1-4d29-9a8e-1785c016ce09"
	bnsB14RocOracle        = "67789c0c-3f05-4b0b-a56d-0757ac801d22"
)

func TestBnsB14EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB14RacketeersOracle: "Orzhov Racketeers",
		bnsB14PatrolOracle:     "Parhelion Patrol",
		bnsB14PasswallOracle:   "Passwall Adept",
		bnsB14GorgonOracle:     "Pitiless Gorgon",
		bnsB14PortalOracle:     "Portal Manipulator",
		bnsB14PyrewildOracle:   "Pyrewild Shaman",
		bnsB14ShredFreakOracle: "Rakdos Shred-Freak",
		bnsB14TrumpeterOracle:  "Rakdos Trumpeter",
		bnsB14SphinxOracle:     "Rescuer Sphinx",
		bnsB14TechnicianOracle: "Resonance Technician",
		bnsB14PredatorOracle:   "Ripscale Predator",
		bnsB14RocOracle:        "Roc Charger",
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

func TestBnsB14PrintedKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, oracle string
		keywords     []string
	}{
		{"Parhelion Patrol", bnsB14PatrolOracle, []string{"flying", "vigilance"}},
		{"Pitiless Gorgon", bnsB14GorgonOracle, []string{"deathtouch"}},
		{"Rakdos Shred-Freak", bnsB14ShredFreakOracle, []string{"haste"}},
		{"Rakdos Trumpeter", bnsB14TrumpeterOracle, []string{"menace"}},
		{"Ripscale Predator", bnsB14PredatorOracle, []string{"menace"}},
		{"Roc Charger", bnsB14RocOracle, []string{"flying"}},
		{"Rescuer Sphinx", bnsB14SphinxOracle, []string{"flying"}},
		{"Resonance Technician", bnsB14TechnicianOracle, []string{"flying"}},
		{"Portal Manipulator", bnsB14PortalOracle, []string{"flash"}},
	} {
		id := b12Push(g, me.ID, tc.name, "Creature", tc.oracle, 2, 2)
		c := layeredCard(t, g, id)
		for _, k := range tc.keywords {
			if !game.HasKeyword(&c, k) {
				t.Errorf("%s has no %s", tc.name, k)
			}
		}
	}
}

// --- Orzhov Racketeers ------------------------------------------------------

func TestBnsB14RacketeersDiscardOnCombatDamageOnly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, other := blbB1Seats(g)
	rack := b12Push(g, me.ID, "Orzhov Racketeers", "Creature — Human Rogue", bnsB14RacketeersOracle, 3, 2)
	keep := pushCatalogHandCard(opp, "Keep", "Instant", "")
	pitch := pushCatalogHandCard(opp, "Pitch", "Sorcery", "")

	dealNoncombatDamage(g, me.ID, rack, opp.ID)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, opp.ID) != nil {
		t.Fatal("noncombat damage triggered the Racketeers")
	}
	dealCombatDamage(g, me.ID, rack, opp.ID)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", n)
	}
	if discardChoiceFor(g, other.ID) != nil {
		t.Error("a player not dealt damage was asked to discard")
	}
	answerDiscard(t, g, opp.ID, pitch)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(pitch) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card was not the one discarded")
	}
}

func TestBnsB14RacketeersAfterlifeMakesTwoSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rack := b12Push(g, me.ID, "Orzhov Racketeers", "Creature — Human Rogue", bnsB14RacketeersOracle, 3, 2)
	b14Kill(g, rack)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Spirit", me.ID); got != 2 {
		t.Fatalf("Spirit tokens = %d, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Spirit" {
			continue
		}
		cc := layeredCard(t, g, c.InstanceID)
		if cc.Power != 1 || cc.Toughness != 1 || !game.HasKeyword(&cc, "flying") {
			t.Errorf("Spirit is %d/%d, flying=%v", cc.Power, cc.Toughness, game.HasKeyword(&cc, "flying"))
		}
	}
}

// --- Parhelion Patrol -------------------------------------------------------

func TestBnsB14PatrolMentorsAnAttackerWithLesserPower(t *testing.T) {
	g, patrol := mentorGame(t, "Parhelion Patrol", bnsB14PatrolOracle, 2)
	me := g.Seats[g.Turn.ActiveSeat]
	small := ally(g, "Small", 1)
	equal := ally(g, "Equal", 2)
	prompt := mentorAttack(t, g, patrol, small, equal)
	if offered := prompt.PickTargetCards; len(offered) != 1 || offered[0] != small {
		t.Fatalf("offered %v, want only the attacker with lesser power", offered)
	}
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if got := allCountersOn(t, g, small)["+1/+1"]; got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// --- Passwall Adept ---------------------------------------------------------

func TestBnsB14PasswallAdeptMakesATargetUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	adept := b12Push(g, me.ID, "Passwall Adept", "Creature — Human Wizard", bnsB14PasswallOracle, 1, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToStepInTurn(t, g, game.StepPrecombatMain)

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, adept, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("activated with two mana")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, adept, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, bear, game.CantBeBlocked)
	assertRestrictions(t, g, adept, 0)
	if b16Tapped(t, g, adept) {
		t.Error("the ability has no tap cost, yet the Adept is tapped")
	}
}

// --- Rakdos Trumpeter -------------------------------------------------------

func TestBnsB14TrumpeterPumpsItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tr := b12Push(g, me.ID, "Rakdos Trumpeter", "Creature — Human Shaman", bnsB14TrumpeterOracle, 1, 3)
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, tr, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with three mana")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, tr, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, tr); got != 3 {
		t.Errorf("power = %d, want 3", got)
	}
}

// --- Roc Charger ------------------------------------------------------------

func TestBnsB14RocChargerGivesAnAttackerWithoutFlyingFlying(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	roc := b12Push(g, me.ID, "Roc Charger", "Creature — Bird", bnsB14RocOracle, 1, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	home := b12Creature(g, me.ID, "Homebody", "Creature — Bear", 2, 2)
	advanceToDeclareAttackersOf(t, g, seat)
	for _, id := range []uuid.UUID{roc, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt")
	}
	if offered := prompt.PickTargetCards; len(offered) != 1 || offered[0] != bear {
		t.Fatalf("offered %v, want only the attacking creature without flying (not %v)", offered, home)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, bear)
	if !game.HasKeyword(&c, "flying") {
		t.Error("the targeted attacker did not gain flying")
	}
	if h := layeredCard(t, g, home); game.HasKeyword(&h, "flying") {
		t.Error("a creature that was not targeted gained flying")
	}
}

// --- Pyrewild Shaman --------------------------------------------------------

func TestBnsB14PyrewildBloodrushPumpsAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	shaman := handCardFull(me, "Pyrewild Shaman", "Creature — Goblin Shaman", "{2}{R}", bnsB14PyrewildOracle, []string{"R"})
	advanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)

	floatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("activated with one mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Targets: cardRefs(idle), Strict: true}); err == nil {
		t.Fatal("targeted a creature that is not attacking")
	}
	if !me.Hand.Contains(shaman) {
		t.Fatal("a refused activation discarded the Shaman")
	}
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("bloodrush: %v", err)
	}
	if !me.Graveyard.Contains(shaman) {
		t.Error("the discard was not the cost")
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, bear)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 3 {
		t.Errorf("the attacker is %d/%d, want 5/3", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestBnsB14PyrewildReturnsFromTheGraveyardForThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	shaman := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: shaman, Name: "Pyrewild Shaman", TypeLine: "Creature — Goblin Shaman", ManaCost: "{2}{R}",
			OracleID: bnsB14PyrewildOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	dealCombatDamage(g, me.ID, bear, opp.ID)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment was offered")
	}
	floatMana(t, g, me, "{C}{C}{C}")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(shaman) {
		t.Fatal("the Shaman did not return to hand after paying {3}")
	}
}

func TestBnsB14PyrewildDecliningLeavesItInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	shaman := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: shaman, Name: "Pyrewild Shaman", TypeLine: "Creature — Goblin Shaman", ManaCost: "{2}{R}",
			OracleID: bnsB14PyrewildOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dealCombatDamage(g, me.ID, bear, opp.ID)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(shaman) || me.Hand.Contains(shaman) {
		t.Error("the Shaman left the graveyard without payment")
	}
}

func TestBnsB14PyrewildDoesNothingFromTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	b12Push(g, me.ID, "Pyrewild Shaman", "Creature — Goblin Shaman", bnsB14PyrewildOracle, 3, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dealCombatDamage(g, me.ID, bear, opp.ID)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the graveyard ability triggered from the battlefield")
	}
}

// --- Resonance Technician ---------------------------------------------------

func TestBnsB14TechnicianDiscardsToInvestigateTwice(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pitch := pushCatalogHandCard(me, "Pitch", "Sorcery", "")
	castAndResolveCreature(t, g, "Resonance Technician", "Creature — Weird Detective", bnsB14TechnicianOracle)
	passPriorityAroundTable(t, g)
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("the controller was not asked to discard (owed %d)", discardOwed(g, me.ID))
	}
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Clue", me.ID); got != 2 {
		t.Errorf("Clues = %d, want 2", got)
	}
}

func TestBnsB14TechnicianDecliningTheDiscardMakesNoClues(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogHandCard(me, "Keep", "Sorcery", "")
	castAndResolveCreature(t, g, "Resonance Technician", "Creature — Weird Detective", bnsB14TechnicianOracle)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Clue", me.ID); got != 0 {
		t.Errorf("Clues = %d, want 0", got)
	}
}

func TestBnsB14TechnicianCopiesASpellWithManaValueX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tech := b12Push(g, me.ID, "Resonance Technician", "Creature — Weird Detective", bnsB14TechnicianOracle, 4, 4)
	rockA := b12Permanent(g, me.ID, "Rock A", "Artifact")
	rockB := b12Permanent(g, me.ID, "Rock B", "Artifact")
	victim := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	advanceToMain(t, g)

	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{1}{R}",
			OracleID: lightningBoltOracle, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
		})
	})
	floatMana(t, g, me, "{R}{C}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: cardRefs(victim)}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
	spell := bolt

	// X = 1 names a mana value the Bolt does not have.
	if err := g.ActivateCatalogAbility(me.ID, tech, 0, game.ActivateAbilityParams{
		XValue: 1, TapIDs: []uuid.UUID{rockA}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}, Strict: true,
	}); err == nil {
		t.Fatal("copied a spell whose mana value is not X")
	}
	if err := g.ActivateCatalogAbility(me.ID, tech, 0, game.ActivateAbilityParams{
		XValue: 2, TapIDs: []uuid.UUID{rockA, rockB}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}, Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !b16Tapped(t, g, rockA) || !b16Tapped(t, g, rockB) || !b16Tapped(t, g, tech) {
		t.Error("the Technician and two artifacts were not tapped as the cost")
	}
	passPriorityAroundTable(t, g)
	// CR 707.10c: the copy offers new targets; keep the Wall.
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, victim); ok {
		t.Error("the Wall survived: the copy did not deal its own 3 damage")
	}
}

// --- Rescuer Sphinx ---------------------------------------------------------

func castRescuerSphinx(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Rescuer Sphinx", "Creature — Sphinx", bnsB14SphinxOracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

func TestBnsB14SphinxReturnsAPermanentAndEntersWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	id := castRescuerSphinx(t, g)
	c := entryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID)
	if c == nil {
		t.Fatal("the Sphinx did not ask")
	}
	if _, ok := battlefieldCard(g, id); ok {
		t.Fatal("the Sphinx entered before the choice (CR 614.12)")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (a may)", c.ChooseMin, c.ChooseMax)
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != bear {
		t.Errorf("candidates %v: want only my nonland permanent (not %v)", c.ChooseCards, theirs)
	}
	answerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the Bear did not return to its owner's hand")
	}
	if got := b12Counter(t, g, id, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters on the Sphinx = %d, want 1", got)
	}
}

func TestBnsB14SphinxDecliningEntersWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := castRescuerSphinx(t, g)
	answerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, bear); !ok {
		t.Error("the Bear left the battlefield though the choice was declined")
	}
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("the Sphinx did not enter")
	}
	if got := b12Counter(t, g, id, "+1/+1"); got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0", got)
	}
}

func TestBnsB14SphinxWithNothingToReturnJustEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	id := castRescuerSphinx(t, g)
	if entryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID) != nil {
		t.Fatal("asked a question with no candidates")
	}
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("the Sphinx did not enter")
	}
}

// --- Portal Manipulator -----------------------------------------------------

// bnsB14PortalCast sets up: seat 0's two attackers are aimed at seat 2, seat 1
// flashes in the Manipulator in the declare attackers step and names itself,
// then returns the creature prompt.
func bnsB14PortalCast(t *testing.T, g *game.Game) (attackerSeat, me, other *game.Player, a, b, mine uuid.UUID, p *game.PendingChoice) {
	t.Helper()
	attackerSeat, me, other = g.Seats[0], g.Seats[1], g.Seats[2]
	a = pushVanillaCreature(g, attackerSeat.ID, "Charger A", 3, 3)
	b = pushVanillaCreature(g, attackerSeat.ID, "Charger B", 2, 2)
	mine = pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	attackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{a: other.ID, b: other.ID})

	castFlashPermanentInPlace(t, g, me.ID, "Portal Manipulator", "Creature — Human Wizard", bnsB14PortalOracle)
	passPriorityAroundTable(t, g)
	p = latestPickTarget(g, me.ID)
	if p == nil || p.Reason != "target player" {
		t.Fatalf("no target player prompt for the Manipulator's controller; pending %d", len(g.PendingChoices))
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err != nil {
		t.Fatalf("ResolvePickTarget player: %v", err)
	}
	p = latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no prompt for the attacking creatures")
	}
	return
}

func TestBnsB14PortalManipulatorRedirectsOpponentsAttackers(t *testing.T) {
	g := newCatalogGame(t)
	_, me, _, a, b, mine, p := bnsB14PortalCast(t, g)
	if len(p.PickTargetCards) != 2 || !hasID(p.PickTargetCards, a) || !hasID(p.PickTargetCards, b) || hasID(p.PickTargetCards, mine) {
		t.Fatalf("offered %v, want exactly the two attackers %v %v", p.PickTargetCards, a, b)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := aangCardOnBF(g, id); c.AttackingTarget != me.ID {
			t.Errorf("attacker %v is attacking %v, want the chosen player", id, c.AttackingTarget)
		}
	}
}

func TestBnsB14PortalManipulatorOnlyMovesTheCreaturesChosen(t *testing.T) {
	g := newCatalogGame(t)
	_, me, other, a, b, _, p := bnsB14PortalCast(t, g)
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := aangCardOnBF(g, a); c.AttackingTarget != me.ID {
		t.Errorf("the chosen attacker is attacking %v, want the chosen player", c.AttackingTarget)
	}
	if c, _ := aangCardOnBF(g, b); c.AttackingTarget != other.ID {
		t.Errorf("the creature not chosen is attacking %v, want it left alone", c.AttackingTarget)
	}
}

func TestBnsB14PortalManipulatorOutsideDeclareAttackersDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	attackerSeat, me, other := g.Seats[0], g.Seats[1], g.Seats[2]
	a := pushVanillaCreature(g, attackerSeat.ID, "Charger A", 3, 3)
	attackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{a: other.ID})
	advanceTo(t, g, game.StepDeclareBlockers)
	castFlashPermanentInPlace(t, g, me.ID, "Portal Manipulator", "Creature — Human Wizard", bnsB14PortalOracle)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the Manipulator asked for targets outside the declare attackers step")
	}
}
