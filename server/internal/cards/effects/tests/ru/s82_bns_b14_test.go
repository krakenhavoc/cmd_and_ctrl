package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

func TestBnsB14PrintedKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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
		id := effectstest.B12Push(g, me.ID, tc.name, "Creature", tc.oracle, 2, 2)
		c := effectstest.LayeredCard(t, g, id)
		for _, k := range tc.keywords {
			if !game.HasKeyword(&c, k) {
				t.Errorf("%s has no %s", tc.name, k)
			}
		}
	}
}

// --- Orzhov Racketeers ------------------------------------------------------

func TestBnsB14RacketeersDiscardOnCombatDamageOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, other := effectstest.BlbB1Seats(g)
	rack := effectstest.B12Push(g, me.ID, "Orzhov Racketeers", "Creature — Human Rogue", bnsB14RacketeersOracle, 3, 2)
	keep := effectstest.PushCatalogHandCard(opp, "Keep", "Instant", "")
	pitch := effectstest.PushCatalogHandCard(opp, "Pitch", "Sorcery", "")

	effectstest.DealNoncombatDamage(g, me.ID, rack, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		t.Fatal("noncombat damage triggered the Racketeers")
	}
	effectstest.DealCombatDamage(g, me.ID, rack, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", n)
	}
	if effectstest.DiscardChoiceFor(g, other.ID) != nil {
		t.Error("a player not dealt damage was asked to discard")
	}
	effectstest.AnswerDiscard(t, g, opp.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(pitch) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card was not the one discarded")
	}
}

func TestBnsB14RacketeersAfterlifeMakesTwoSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rack := effectstest.B12Push(g, me.ID, "Orzhov Racketeers", "Creature — Human Rogue", bnsB14RacketeersOracle, 3, 2)
	effectstest.B14Kill(g, rack)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Spirit", me.ID); got != 2 {
		t.Fatalf("Spirit tokens = %d, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Spirit" {
			continue
		}
		cc := effectstest.LayeredCard(t, g, c.InstanceID)
		if cc.Power != 1 || cc.Toughness != 1 || !game.HasKeyword(&cc, "flying") {
			t.Errorf("Spirit is %d/%d, flying=%v", cc.Power, cc.Toughness, game.HasKeyword(&cc, "flying"))
		}
	}
}

// --- Parhelion Patrol -------------------------------------------------------

func TestBnsB14PatrolMentorsAnAttackerWithLesserPower(t *testing.T) {
	g, patrol := effectstest.MentorGame(t, "Parhelion Patrol", bnsB14PatrolOracle, 2)
	me := g.Seats[g.Turn.ActiveSeat]
	small := effectstest.Ally(g, "Small", 1)
	equal := effectstest.Ally(g, "Equal", 2)
	prompt := effectstest.MentorAttack(t, g, patrol, small, equal)
	if offered := prompt.PickTargetCards; len(offered) != 1 || offered[0] != small {
		t.Fatalf("offered %v, want only the attacker with lesser power", offered)
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AllCountersOn(t, g, small)["+1/+1"]; got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// --- Passwall Adept ---------------------------------------------------------

func TestBnsB14PasswallAdeptMakesATargetUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	adept := effectstest.B12Push(g, me.ID, "Passwall Adept", "Creature — Human Wizard", bnsB14PasswallOracle, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, adept, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Fatal("activated with two mana")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, adept, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, bear, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, adept, 0)
	if effectstest.B16Tapped(t, g, adept) {
		t.Error("the ability has no tap cost, yet the Adept is tapped")
	}
}

// --- Rakdos Trumpeter -------------------------------------------------------

func TestBnsB14TrumpeterPumpsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tr := effectstest.B12Push(g, me.ID, "Rakdos Trumpeter", "Creature — Human Shaman", bnsB14TrumpeterOracle, 1, 3)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
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
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, tr); got != 3 {
		t.Errorf("power = %d, want 3", got)
	}
}

// --- Roc Charger ------------------------------------------------------------

func TestBnsB14RocChargerGivesAnAttackerWithoutFlyingFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	roc := effectstest.B12Push(g, me.ID, "Roc Charger", "Creature — Bird", bnsB14RocOracle, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Homebody", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	for _, id := range []uuid.UUID{roc, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt")
	}
	if offered := prompt.PickTargetCards; len(offered) != 1 || offered[0] != bear {
		t.Fatalf("offered %v, want only the attacking creature without flying (not %v)", offered, home)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, bear)
	if !game.HasKeyword(&c, "flying") {
		t.Error("the targeted attacker did not gain flying")
	}
	if h := effectstest.LayeredCard(t, g, home); game.HasKeyword(&h, "flying") {
		t.Error("a creature that was not targeted gained flying")
	}
}

// --- Pyrewild Shaman --------------------------------------------------------

func TestBnsB14PyrewildBloodrushPumpsAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	shaman := effectstest.HandCardFull(me, "Pyrewild Shaman", "Creature — Goblin Shaman", "{2}{R}", bnsB14PyrewildOracle, []string{"R"})
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)

	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Fatal("activated with one mana")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(idle), Strict: true}); err == nil {
		t.Fatal("targeted a creature that is not attacking")
	}
	if !me.Hand.Contains(shaman) {
		t.Fatal("a refused activation discarded the Shaman")
	}
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("bloodrush: %v", err)
	}
	if !me.Graveyard.Contains(shaman) {
		t.Error("the discard was not the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, bear)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 3 {
		t.Errorf("the attacker is %d/%d, want 5/3", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestBnsB14PyrewildReturnsFromTheGraveyardForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	shaman := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: shaman, Name: "Pyrewild Shaman", TypeLine: "Creature — Goblin Shaman", ManaCost: "{2}{R}",
			OracleID: bnsB14PyrewildOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.DealCombatDamage(g, me.ID, bear, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment was offered")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(shaman) {
		t.Fatal("the Shaman did not return to hand after paying {3}")
	}
}

func TestBnsB14PyrewildDecliningLeavesItInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	shaman := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: shaman, Name: "Pyrewild Shaman", TypeLine: "Creature — Goblin Shaman", ManaCost: "{2}{R}",
			OracleID: bnsB14PyrewildOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamage(g, me.ID, bear, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(shaman) || me.Hand.Contains(shaman) {
		t.Error("the Shaman left the graveyard without payment")
	}
}

func TestBnsB14PyrewildDoesNothingFromTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.B12Push(g, me.ID, "Pyrewild Shaman", "Creature — Goblin Shaman", bnsB14PyrewildOracle, 3, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamage(g, me.ID, bear, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the graveyard ability triggered from the battlefield")
	}
}

// --- Resonance Technician ---------------------------------------------------

func TestBnsB14TechnicianDiscardsToInvestigateTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pitch := effectstest.PushCatalogHandCard(me, "Pitch", "Sorcery", "")
	effectstest.CastAndResolveCreature(t, g, "Resonance Technician", "Creature — Weird Detective", bnsB14TechnicianOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Fatalf("the controller was not asked to discard (owed %d)", effectstest.DiscardOwed(g, me.ID))
	}
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Clue", me.ID); got != 2 {
		t.Errorf("Clues = %d, want 2", got)
	}
}

func TestBnsB14TechnicianDecliningTheDiscardMakesNoClues(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogHandCard(me, "Keep", "Sorcery", "")
	effectstest.CastAndResolveCreature(t, g, "Resonance Technician", "Creature — Weird Detective", bnsB14TechnicianOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Clue", me.ID); got != 0 {
		t.Errorf("Clues = %d, want 0", got)
	}
}

func TestBnsB14TechnicianCopiesASpellWithManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tech := effectstest.B12Push(g, me.ID, "Resonance Technician", "Creature — Weird Detective", bnsB14TechnicianOracle, 4, 4)
	rockA := effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	rockB := effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	victim := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	effectstest.AdvanceToMain(t, g)

	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{1}{R}",
			OracleID: effectstest.LightningBoltOracle, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
		})
	})
	effectstest.FloatMana(t, g, me, "{R}{C}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err != nil {
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
	if !effectstest.B16Tapped(t, g, rockA) || !effectstest.B16Tapped(t, g, rockB) || !effectstest.B16Tapped(t, g, tech) {
		t.Error("the Technician and two artifacts were not tapped as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// CR 707.10c: the copy offers new targets; keep the Wall.
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the Wall survived: the copy did not deal its own 3 damage")
	}
}

// --- Rescuer Sphinx ---------------------------------------------------------

func castRescuerSphinx(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Rescuer Sphinx", "Creature — Sphinx", bnsB14SphinxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestBnsB14SphinxReturnsAPermanentAndEntersWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	id := castRescuerSphinx(t, g)
	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID)
	if c == nil {
		t.Fatal("the Sphinx did not ask")
	}
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
		t.Fatal("the Sphinx entered before the choice (CR 614.12)")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (a may)", c.ChooseMin, c.ChooseMax)
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != bear {
		t.Errorf("candidates %v: want only my nonland permanent (not %v)", c.ChooseCards, theirs)
	}
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the Bear did not return to its owner's hand")
	}
	if got := effectstest.B12Counter(t, g, id, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters on the Sphinx = %d, want 1", got)
	}
}

func TestBnsB14SphinxDecliningEntersWithoutACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := castRescuerSphinx(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the Bear left the battlefield though the choice was declined")
	}
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
		t.Fatal("the Sphinx did not enter")
	}
	if got := effectstest.B12Counter(t, g, id, "+1/+1"); got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0", got)
	}
}

func TestBnsB14SphinxWithNothingToReturnJustEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	id := castRescuerSphinx(t, g)
	if effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID) != nil {
		t.Fatal("asked a question with no candidates")
	}
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
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
	a = effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger A", 3, 3)
	b = effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger B", 2, 2)
	mine = effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.AttackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{a: other.ID, b: other.ID})

	effectstest.CastFlashPermanentInPlace(t, g, me.ID, "Portal Manipulator", "Creature — Human Wizard", bnsB14PortalOracle)
	effectstest.PassPriorityAroundTable(t, g)
	p = effectstest.LatestPickTarget(g, me.ID)
	if p == nil || p.Reason != "target player" {
		t.Fatalf("no target player prompt for the Manipulator's controller; pending %d", len(g.PendingChoices))
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err != nil {
		t.Fatalf("ResolvePickTarget player: %v", err)
	}
	p = effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no prompt for the attacking creatures")
	}
	return
}

func TestBnsB14PortalManipulatorRedirectsOpponentsAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, me, _, a, b, mine, p := bnsB14PortalCast(t, g)
	if len(p.PickTargetCards) != 2 || !effectstest.HasID(p.PickTargetCards, a) || !effectstest.HasID(p.PickTargetCards, b) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Fatalf("offered %v, want exactly the two attackers %v %v", p.PickTargetCards, a, b)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := effectstest.AangCardOnBF(g, id); c.AttackingTarget != me.ID {
			t.Errorf("attacker %v is attacking %v, want the chosen player", id, c.AttackingTarget)
		}
	}
}

func TestBnsB14PortalManipulatorOnlyMovesTheCreaturesChosen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, me, other, a, b, _, p := bnsB14PortalCast(t, g)
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.AangCardOnBF(g, a); c.AttackingTarget != me.ID {
		t.Errorf("the chosen attacker is attacking %v, want the chosen player", c.AttackingTarget)
	}
	if c, _ := effectstest.AangCardOnBF(g, b); c.AttackingTarget != other.ID {
		t.Errorf("the creature not chosen is attacking %v, want it left alone", c.AttackingTarget)
	}
}

func TestBnsB14PortalManipulatorOutsideDeclareAttackersDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attackerSeat, me, other := g.Seats[0], g.Seats[1], g.Seats[2]
	a := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger A", 3, 3)
	effectstest.AttackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{a: other.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.CastFlashPermanentInPlace(t, g, me.ID, "Portal Manipulator", "Creature — Human Wizard", bnsB14PortalOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the Manipulator asked for targets outside the declare attackers step")
	}
}
