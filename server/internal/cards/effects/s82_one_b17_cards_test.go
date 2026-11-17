package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set ONE (March of the Machine: Phyrexia: All Will Be One), batch 17
// (#5479). Each test plays the real catalog entry through the engine.

const (
	s82B17JorKadeenOracle = "967897db-a34b-493c-bd48-04e2650902d2"
	s82B17KembaOracle     = "93c97e1a-bd7e-4307-8ada-58f53a887cf8"
	s82B17RhukOracle      = "06607cba-76ba-4145-a5b5-24d3428142ac"
	s82B17OtharriOracle   = "b081352b-d17b-4031-bbcc-c5c02cbfa6b6"
	s82B17UrtetOracle     = "6222fccf-fc08-4190-8d40-a56d6d1423df"
	s82B17SkyhunterOracle = "8ade0824-9fd0-417d-94dd-ac399e35b582"
	s82B17RoarOracle      = "a0753b28-e693-4f5b-a4e4-846e933058d3"
	s82B17TyrranaxOracle  = "6e42da0c-151e-468d-91cb-5a5b117a9298"
	s82B17AnnexOracle     = "be31fa46-d5ac-4c85-9c19-3f0e2b1f223b"
)

func TestS82B17EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82B17JorKadeenOracle: "Jor Kadeen, First Goldwarden",
		s82B17KembaOracle:     "Kemba, Kha Enduring",
		s82B17RhukOracle:      "Rhuk, Hexgold Nabber",
		s82B17OtharriOracle:   "Otharri, Suns' Glory",
		s82B17UrtetOracle:     "Urtet, Remnant of Memnarch",
		s82B17SkyhunterOracle: "Skyhunter Strike Force",
		s82B17RoarOracle:      "Roar of Resistance",
		s82B17TyrranaxOracle:  "Tyrranax Rex",
		s82B17AnnexOracle:     "Annex Sentry",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// s82B17Equip puts an Equipment of `owner`'s on the battlefield attached to host.
func s82B17Equip(t *testing.T, g *game.Game, owner, host uuid.UUID, name string) uuid.UUID {
	t.Helper()
	eq := b12Permanent(g, owner, name, "Artifact — Equipment")
	var err error
	g.WithWriteLock(func() { err = g.AttachForEffect(eq, game.TargetRef{Kind: game.TargetCard, ID: host}) })
	if err != nil {
		t.Fatalf("attach %s: %v", name, err)
	}
	return eq
}

func s82B17Host(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	return attachmentHostOf(t, g, id).ID
}

// --- Annex Sentry ---------------------------------------------------------------

func TestS82B17AnnexSentryExilesASmallArtifactOrCreatureUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	smallCreature := woeB1CreatureOfCost(g, opp.ID, "Small", "{1}{G}{G}")
	bigCreature := woeB1CreatureOfCost(g, opp.ID, "Big", "{2}{G}{G}")
	mine := woeB1CreatureOfCost(g, me.ID, "Mine", "{G}")
	trinket := b27Push(g, opp.ID, "Trinket", "Artifact", "", "{2}", 0, 0)
	enchant := b27Push(g, opp.ID, "Aura Thing", "Enchantment", "", "{1}", 0, 0)
	sentry := castCatalogSpell(t, g, "Annex Sentry", "Artifact Creature — Phyrexian Cleric", s82B17AnnexOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	for name, id := range map[string]uuid.UUID{
		"a mana value 4 creature": bigCreature, "your own creature": mine, "an enchantment": enchant,
	} {
		if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, cardRefs(id)); err == nil {
			t.Fatalf("Annex Sentry targeted %s", name)
		}
	}
	pickCard(t, g, me.ID, trinket)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(trinket) || g.Battlefield.Contains(trinket) {
		t.Fatal("the mana value 2 artifact was not exiled")
	}
	if !g.Battlefield.Contains(smallCreature) {
		t.Error("the other creature left too")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(sentry) })
	g.RunStateChecksForTest()
	if n := namedOnBattlefield(g, "Trinket", opp.ID); n != 1 {
		t.Errorf("after the Sentry left, %d of the exiled artifact under its owner's control, want 1", n)
	}
}

func TestS82B17AnnexSentryCanExileACreatureAndHasToxic(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	small := woeB1CreatureOfCost(g, opp.ID, "Small", "{1}{G}{G}")
	sentry := castCatalogSpell(t, g, "Annex Sentry", "Artifact Creature — Phyrexian Cleric", s82B17AnnexOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(small) {
		t.Fatal("the creature was not exiled")
	}
	assertKeywords(t, g, sentry, "toxic 1")
}

// --- Tyrranax Rex ---------------------------------------------------------------

func TestS82B17TyrranaxRexIsUncounterableAndHasItsKeywords(t *testing.T) {
	spec, ok := Lookup(s82B17TyrranaxOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatal("Tyrranax Rex's spell can't be countered")
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	rex := b27Push(g, me.ID, "Tyrranax Rex", "Creature — Phyrexian Dinosaur", s82B17TyrranaxOracle, "{4}{G}{G}{G}", 8, 8, "G")
	assertKeywords(t, g, rex, "trample", "haste", "toxic 4", "ward")
	// Toxic 4: combat damage to a player gives four poison counters.
	opp := g.Seats[1]
	attackWith(t, g, opp.ID, rex)
	if got := opp.Counters[game.CounterPoison]; got != 4 {
		t.Errorf("poison counters = %d, want 4", got)
	}
}

// --- Skyhunter Strike Force ---------------------------------------------------------

func TestS82B17SkyhunterGivesOthersMeleeOnlyWithYourCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	force := b27Push(g, me.ID, "Skyhunter Strike Force", "Creature — Cat Knight", s82B17SkyhunterOracle, "{2}{W}", 2, 2, "W")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	if !hasEffectiveKeyword(t, g, force, "flying") || !hasEffectiveKeyword(t, g, force, game.KeywordMelee) {
		t.Fatal("the Force has flying and melee")
	}
	if hasEffectiveKeyword(t, g, bear, game.KeywordMelee) {
		t.Fatal("without your commander, nothing else has melee")
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Elf",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	if !hasEffectiveKeyword(t, g, bear, game.KeywordMelee) {
		t.Error("with your commander, other creatures you control have melee")
	}
	if hasEffectiveKeyword(t, g, theirs, game.KeywordMelee) {
		t.Error("an opponent's creature does not")
	}
}

func TestS82B17SkyhunterIgnoresAStolenCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Skyhunter Strike Force", "Creature — Cat Knight", s82B17SkyhunterOracle, "{2}{W}", 2, 2, "W")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Commander", TypeLine: "Legendary Creature — Elf",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: me.ID, IsCommander: true,
	})
	if hasEffectiveKeyword(t, g, bear, game.KeywordMelee) {
		t.Error("an opponent's commander you control is not your commander")
	}
}

func TestS82B17SkyhunterMeleeCountsOpponentsAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	force := b27Push(g, me.ID, "Skyhunter Strike Force", "Creature — Cat Knight", s82B17SkyhunterOracle, "{2}{W}", 2, 2, "W")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(force, opp1.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(bear, opp2.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p, th := effectivePower(t, g, force), effectiveToughness(t, g, force); p != 4 || th != 4 {
		t.Errorf("the Force is %d/%d, want 4/4 for two opponents attacked", p, th)
	}
}

// --- Roar of Resistance ---------------------------------------------------------------

func TestS82B17RoarGivesCreatureTokensHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Roar of Resistance", "Enchantment", s82B17RoarOracle, 0, 0)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	card := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	if !hasEffectiveKeyword(t, g, token, "haste") {
		t.Error("a creature token you control has haste")
	}
	if hasEffectiveKeyword(t, g, card, "haste") || hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("a nontoken creature, or an opponent's token, does not")
	}
}

// s82B17RoarAttack has `attacker` (controlled by someone) attack `defender`
// and settles up to the Roar's payment prompt.
func s82B17RoarSetup(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	b12Push(g, me.ID, "Roar of Resistance", "Enchantment", s82B17RoarOracle, 0, 0)
	b06AddMana(me, "R", "R")
}

func TestS82B17RoarPaidPumpsAttackersAimedAtOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	s82B17RoarSetup(t, g, me)
	atOpp1 := b16Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2, "G")
	atOpp2 := b16Creature(g, me.ID, "Bear Two", "Creature — Bear", 3, 3, "G")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atOpp1, opp1.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(atOpp2, opp2.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("one trigger for the whole attack should ask for {1}{R}")
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, atOpp1) != 4 || effectivePower(t, g, atOpp2) != 5 {
		t.Errorf("attackers are %d and %d, want 4 and 5", effectivePower(t, g, atOpp1), effectivePower(t, g, atOpp2))
	}
}

func TestS82B17RoarDeclinedPumpsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1 := g.Seats[0], g.Seats[1]
	s82B17RoarSetup(t, g, me)
	bear := b16Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2, "G")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp1.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, bear) != 2 {
		t.Errorf("power = %d, want 2", effectivePower(t, g, bear))
	}
}

// An opponent attacking another opponent triggers it, and their creature is
// "attacking your opponents" too; one attacking YOU is not.
func TestS82B17RoarCountsAnOpponentsAttackOnAnotherOpponentButNotOnYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	s82B17RoarSetup(t, g, me)
	// opp1's turn: advance the cursor so opp1 is active.
	advanceToMainOf(t, g, 1)
	bearAtOpp2 := b16Creature(g, opp1.ID, "Bear At Two", "Creature — Bear", 2, 2, "G")
	bearAtMe := b16Creature(g, opp1.ID, "Bear At Me", "Creature — Bear", 2, 2, "G")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bearAtOpp2, opp2.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(bearAtMe, me.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("an attack by any player triggers the Roar")
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, bearAtOpp2) != 4 {
		t.Errorf("the creature attacking another opponent is %d, want 4", effectivePower(t, g, bearAtOpp2))
	}
	if effectivePower(t, g, bearAtMe) != 2 {
		t.Errorf("the creature attacking you is %d, want 2", effectivePower(t, g, bearAtMe))
	}
}

// --- Urtet, Remnant of Memnarch ------------------------------------------------------

func TestS82B17UrtetMakesAMyrWhenYouCastAMyrSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Urtet, Remnant of Memnarch", "Legendary Artifact Creature — Myr", s82B17UrtetOracle, "{3}", 2, 2)
	advanceToMain(t, g)
	myr := nsHandCard(me, "Myr Spell", "Artifact Creature — Myr", "")
	bear := nsHandCard(me, "Bear Spell", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Myr", me.ID); n != 0 {
		t.Fatalf("a Bear spell made %d Myr tokens", n)
	}
	if err := g.CastSpell(me.ID, myr.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Myr", me.ID); n != 1 {
		t.Errorf("a Myr spell made %d Myr tokens, want 1", n)
	}
}

func TestS82B17UrtetUntapsEachMyrAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	urtet := b27Push(g, me.ID, "Urtet, Remnant of Memnarch", "Legendary Artifact Creature — Myr", s82B17UrtetOracle, "{3}", 2, 2)
	myr := b12Creature(g, me.ID, "Other Myr", "Artifact Creature — Myr", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{urtet, myr, bear} {
			if c := findBattlefieldCardForTest(g, id); c != nil {
				c.Tapped = true
			}
		}
	})
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Urtet": urtet, "the other Myr": myr} {
		if c, _ := battlefieldCard(g, id); c.Tapped {
			t.Errorf("%s was not untapped", name)
		}
	}
	if c, _ := battlefieldCard(g, bear); !c.Tapped {
		t.Error("the Bear is not a Myr and stays tapped")
	}
}

func TestS82B17UrtetPutsThreeCountersOnEachMyrDuringYourTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	urtet := b27Push(g, me.ID, "Urtet, Remnant of Memnarch", "Legendary Artifact Creature — Myr", s82B17UrtetOracle, "{3}", 2, 2)
	myr := b12Creature(g, me.ID, "Other Myr", "Artifact Creature — Myr", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{U}{B}{R}{G}"); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if c := findBattlefieldCardForTest(g, urtet); c != nil {
			c.SummonedThisTurn = false
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, urtet, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if effectivePowerWithCounters(t, g, urtet) != 5 || effectivePowerWithCounters(t, g, myr) != 4 {
		t.Errorf("Myr are %d and %d, want 5 and 4", effectivePowerWithCounters(t, g, urtet), effectivePowerWithCounters(t, g, myr))
	}
	if effectivePowerWithCounters(t, g, bear) != 2 {
		t.Errorf("the Bear is %d, want 2", effectivePowerWithCounters(t, g, bear))
	}
}

func TestS82B17UrtetCantActivateOnAnotherPlayersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	urtet := b27Push(g, me.ID, "Urtet, Remnant of Memnarch", "Legendary Artifact Creature — Myr", s82B17UrtetOracle, "{3}", 2, 2)
	advanceToMainOf(t, g, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{U}{B}{R}{G}"); err != nil {
		t.Fatal(err)
	}
	err := g.ActivateCatalogAbility(me.ID, urtet, 0, game.ActivateAbilityParams{Strict: true})
	if err == nil {
		t.Error("Urtet activated on an opponent's turn")
	}
}

// --- Kemba, Kha Enduring --------------------------------------------------------------

func TestS82B17KembaAttachesAnEquipmentToAnEnteringCatAndPumpsEquipped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kemba := b27Push(g, me.ID, "Kemba, Kha Enduring", "Legendary Creature — Cat Cleric", s82B17KembaOracle, "{1}{W}", 2, 2, "W")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	sword := s82B17Equip(t, g, me.ID, bear, "Sword")
	if effectivePower(t, g, bear) != 3 || effectivePower(t, g, kemba) != 2 {
		t.Fatalf("the equipped Bear is %d (want 3), unequipped Kemba %d (want 2)", effectivePower(t, g, bear), effectivePower(t, g, kemba))
	}
	cat := castCatalogSpell(t, g, "Cat Cub", "Creature — Cat", "", nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	if got := s82B17Host(t, g, sword); got != cat {
		t.Errorf("the Equipment is on %s, want the entering Cat %s", got, cat)
	}
}

func TestS82B17KembaTriggersForItselfAndNotForANonCat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	sword := s82B17Equip(t, g, me.ID, bear, "Sword")
	// A Bear entering is not a Cat: no prompt.
	b27Push(g, me.ID, "Kemba, Kha Enduring", "Legendary Creature — Cat Cleric", s82B17KembaOracle, "{1}{W}", 2, 2, "W")
	castCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a non-Cat entering triggered Kemba")
	}
	if got := s82B17Host(t, g, sword); got != bear {
		t.Error("the Equipment moved")
	}
}

func TestS82B17KembaCreatesACatForFiveMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kemba := b27Push(g, me.ID, "Kemba, Kha Enduring", "Legendary Creature — Cat Cleric", s82B17KembaOracle, "{1}{W}", 2, 2, "W")
	advanceToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{W}{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, kemba, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Cat", me.ID); n != 1 {
		t.Errorf("%d Cat tokens, want 1", n)
	}
	// The Cat token entering triggers Kemba, who may attach an Equipment: none here.
}

// --- Jor Kadeen, First Goldwarden ----------------------------------------------------------

func TestS82B17JorKadeenGrowsPerEquippedCreatureAndDrawsAtFourPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jor := b27Push(g, me.ID, "Jor Kadeen, First Goldwarden", "Legendary Creature — Human Rebel", s82B17JorKadeenOracle, "{R}{W}", 2, 2, "R", "W")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	s82B17Equip(t, g, me.ID, bear, "Sword")
	s82B17Equip(t, g, me.ID, jor, "Axe")
	hand := me.Hand.Size()
	attackWithTrigger(t, g, jor, opp.ID, uuid.Nil)
	passPriorityAroundTable(t, g)
	// Two equipped creatures: +2/+2, power 4, draw.
	if p := effectivePower(t, g, jor); p != 4 {
		t.Errorf("Jor Kadeen's power = %d, want 4", p)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestS82B17JorKadeenDoesNotDrawBelowFourPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jor := b27Push(g, me.ID, "Jor Kadeen, First Goldwarden", "Legendary Creature — Human Rebel", s82B17JorKadeenOracle, "{R}{W}", 2, 2, "R", "W")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	s82B17Equip(t, g, me.ID, bear, "Sword")
	hand := me.Hand.Size()
	attackWithTrigger(t, g, jor, opp.ID, uuid.Nil)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, jor); p != 3 {
		t.Errorf("Jor Kadeen's power = %d, want 3", p)
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards at power 3, want 0", got)
	}
}

// --- Otharri, Suns' Glory ----------------------------------------------------------------

func TestS82B17OtharriMakesRebelsPerExperienceCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	otharri := b27Push(g, me.ID, "Otharri, Suns' Glory", "Legendary Creature — Phoenix", s82B17OtharriOracle, "{3}{R}{W}", 3, 3, "R", "W")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(otharri, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passAndSettleAttackSplit(t, g, me.ID)
	if got := me.Counters[game.CounterExperience]; got != 1 {
		t.Fatalf("experience counters = %d, want 1", got)
	}
	if n := countOnBattlefield(g, "Rebel", me.ID); n != 1 {
		t.Fatalf("%d Rebels after the first attack, want 1", n)
	}
	var tapped, attacking bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Rebel" && c.Controller == me.ID {
			tapped, attacking = c.Tapped, c.AttackingTarget == opp.ID
		}
	}
	if !tapped || !attacking {
		t.Errorf("the Rebel is tapped=%v attacking=%v, want both", tapped, attacking)
	}
	assertKeywords(t, g, otharri, "flying", "lifelink", "haste")
}

func TestS82B17OtharriSecondAttackMakesTwoRebels(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	otharri := b27Push(g, me.ID, "Otharri, Suns' Glory", "Legendary Creature — Phoenix", s82B17OtharriOracle, "{3}{R}{W}", 3, 3, "R", "W")
	if err := g.AddPlayerCounterForEffect(me.ID, game.CounterExperience, 1); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(otharri, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passAndSettleAttackSplit(t, g, me.ID)
	if n := countOnBattlefield(g, "Rebel", me.ID); n != 2 {
		t.Errorf("%d Rebels with two experience counters, want 2", n)
	}
}

func TestS82B17OtharriReturnsFromTheGraveyardByTappingARebel(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rebel := b12Creature(g, me.ID, "Rebel Scout", "Creature — Human Rebel", 1, 1)
	id, _ := activateFromGraveyard(t, g, "Otharri, Suns' Glory", "Legendary Creature — Phoenix", s82B17OtharriOracle, 3, 3, "{C}{C}{R}{W}", game.ActivateAbilityParams{})
	if c, _ := battlefieldCard(g, rebel); !c.Tapped {
		t.Error("the Rebel was not tapped to pay")
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Otharri did not return")
	}
	if !c.Tapped {
		t.Error("Otharri returns tapped")
	}
}

func TestS82B17OtharriCannotReturnWithoutAnUntappedRebel(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tappedRebel := b12Creature(g, me.ID, "Rebel Scout", "Creature — Human Rebel", 1, 1)
	g.WithWriteLock(func() {
		if c := findBattlefieldCardForTest(g, tappedRebel); c != nil {
			c.Tapped = true
		}
	})
	advanceToMain(t, g)
	id := pushCatalogGraveyardCard(me, "Otharri, Suns' Glory", "Legendary Creature — Phoenix", s82B17OtharriOracle, 3, 3)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Otharri returned with no untapped Rebel to tap")
	}
}

// --- Rhuk, Hexgold Nabber ----------------------------------------------------------------

func TestS82B17RhukTakesEquipmentFromAnEquippedCreatureThatAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhuk := b27Push(g, me.ID, "Rhuk, Hexgold Nabber", "Legendary Creature — Goblin Rebel", s82B17RhukOracle, "{2}{R}", 2, 2, "R")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	sword := s82B17Equip(t, g, me.ID, bear, "Sword")
	axe := s82B17Equip(t, g, me.ID, bear, "Axe")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	for name, eq := range map[string]uuid.UUID{"Sword": sword, "Axe": axe} {
		if got := s82B17Host(t, g, eq); got != rhuk {
			t.Errorf("%s is on %s, want Rhuk", name, got)
		}
	}
}

func TestS82B17RhukMayDeclineAndIgnoresUnequippedAndItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhuk := b27Push(g, me.ID, "Rhuk, Hexgold Nabber", "Legendary Creature — Goblin Rebel", s82B17RhukOracle, "{2}{R}", 2, 2, "R")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	sword := s82B17Equip(t, g, me.ID, bear, "Sword")
	// Rhuk himself attacking equipped is not "other than Rhuk".
	s82B17Equip(t, g, me.ID, rhuk, "Own Axe")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rhuk, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if owedMays(g, me.ID) != 0 {
		t.Fatal("Rhuk attacking triggered himself")
	}
	if got := s82B17Host(t, g, sword); got != bear {
		t.Error("the Sword moved")
	}
}

func TestS82B17RhukDeclinedLeavesTheEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Rhuk, Hexgold Nabber", "Legendary Creature — Goblin Rebel", s82B17RhukOracle, "{2}{R}", 2, 2, "R")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	sword := s82B17Equip(t, g, me.ID, bear, "Sword")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := s82B17Host(t, g, sword); got != bear {
		t.Error("the Sword moved though Rhuk's controller declined")
	}
}

func TestS82B17RhukTakesTheEquipmentOfAnEquippedCreatureThatDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rhuk := b27Push(g, me.ID, "Rhuk, Hexgold Nabber", "Legendary Creature — Goblin Rebel", s82B17RhukOracle, "{2}{R}", 2, 2, "R")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	sword := s82B17Equip(t, g, me.ID, bear, "Sword")
	axe := s82B17Equip(t, g, me.ID, bear, "Axe")
	b27Kill(g, bear)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	for name, eq := range map[string]uuid.UUID{"Sword": sword, "Axe": axe} {
		if got := s82B17Host(t, g, eq); got != rhuk {
			t.Errorf("%s is on %s after the Bear died, want Rhuk", name, got)
		}
	}
}

func TestS82B17RhukIgnoresAnUnequippedCreatureThatDiesAndAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Rhuk, Hexgold Nabber", "Legendary Creature — Goblin Rebel", s82B17RhukOracle, "{2}{R}", 2, 2, "R")
	bare := b16Creature(g, me.ID, "Bare Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	s82B17Equip(t, g, opp.ID, theirs, "Their Sword")
	b27Kill(g, bare)
	b27Kill(g, theirs)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if owedMays(g, me.ID) != 0 {
		t.Error("Rhuk triggered for an unequipped creature or an opponent's creature")
	}
}
