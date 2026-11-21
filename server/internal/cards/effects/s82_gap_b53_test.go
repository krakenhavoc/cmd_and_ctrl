package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapSkitteringPrecursorOracle = "7b1519b1-665c-47b8-8f96-0027410ed81c"
	gapSpawnGangOracle           = "3641ceef-4f26-4647-ac12-c958524d68ca"
	gapConduitGoblinOracle       = "0d3a38ff-4b75-4e97-9d8f-c54f441f8208"
	gapPyrosurferOracle          = "65c35934-796c-4495-a6b5-eea654dfcd55"
	gapVoidpouncerOracle         = "bb4ebe30-2501-48aa-879a-2402a9eb1be7"
	gapInfernalCaptorOracle      = "2f0dec6c-797f-4dc3-9994-066c00cbf1a4"
	gapSkoaOracle                = "0e49ee63-a293-4b33-98cc-ebdcc2c64e82"
	gapPartyThrasherOracle       = "03cd7583-d054-4359-8c94-7776c54198ed"
	gapCopycrookOracle           = "ce184565-71fb-4e23-b45b-8786727d21f7"
	gapShapeshifterOracle        = "5c4cbf1c-95ee-4555-8351-ea962e17ae1f"
	gapDreamtideWhaleOracle      = "ee8c0b4c-03c5-4da5-b4f1-028c845384af"
)

func TestGapSkitteringPrecursorMakesASpawnPerNontokenSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Skittering Precursor", "Creature — Eldrazi Drone", gapSkitteringPrecursorOracle, false)
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	if err := g.SacrificePermanent(me.ID, fodder); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 1 {
		t.Fatalf("after a nontoken sacrifice: %d Spawn, want 1", n)
	}
	// Sacrificing the Spawn itself (a token) does not trigger it.
	var spawn = gapFindNamed(t, g, me.ID, "Eldrazi Spawn")
	if err := g.SacrificePermanent(me.ID, spawn); err != nil {
		t.Fatalf("sacrifice token: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 0 {
		t.Errorf("a token sacrifice made %d Spawn, want 0", n)
	}
	// An opponent's sacrifice is not "you".
	if err := g.SacrificePermanent(opp.ID, theirs); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 0 {
		t.Errorf("an opponent's sacrifice made %d Spawn", n)
	}
	assertKeywords(t, g, gapFindNamed(t, g, me.ID, "Skittering Precursor"), "menace", "devoid")
}

func TestGapSpawnGangCommanderCastTriggerAndSacrificeDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castWithCost(t, g, "Spawn-Gang Commander", "Creature — Eldrazi Goblin", "", gapSpawnGangOracle)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 3 {
		t.Fatalf("cast trigger made %d Spawn, want 3", n)
	}
	cmd := gapFindNamed(t, g, me.ID, "Spawn-Gang Commander")
	spawn := gapFindNamed(t, g, me.ID, "Eldrazi Spawn")

	// Refused with no colourless mana: {1}{C} needs a real {C}.
	if err := g.AddManaForEffect(me.ID, me.ID, "{R}{R}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	refusedForMana(t, g.ActivateCatalogAbility(me.ID, cmd, 0, game.ActivateAbilityParams{
		Targets: playerTargets(opp.ID), SacrificeIDs: ids(spawn), Strict: true,
	}), "red mana for {C}")
	if err := g.AddManaForEffect(me.ID, me.ID, "{C}{R}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	// Sacrificing a non-Eldrazi is refused.
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, cmd, 0, game.ActivateAbilityParams{
		Targets: playerTargets(opp.ID), SacrificeIDs: ids(bear), Strict: true,
	}); err == nil {
		t.Error("sacrificed a non-Eldrazi")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, cmd, 0, game.ActivateAbilityParams{
		Targets: playerTargets(opp.ID), SacrificeIDs: ids(spawn), Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, opp.Life)
	}
}

// gapFindNamed is the first battlefield permanent a player controls with
// this name.
func gapFindNamed(t *testing.T, g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	t.Helper()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.Name == name {
			return c.InstanceID
		}
	}
	t.Fatalf("no %s under %s's control", name, controller)
	return uuid.Nil
}

func playerTargets(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func TestGapConduitGoblinEntersWithEnergyAndPaysForPumpAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gob := b12Push(g, me.ID, "Conduit Goblin", "Creature — Goblin Warrior", gapConduitGoblinOracle, 2, 2)
	bear := b12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	setEnergy(t, g, me, 1)
	advanceTo(t, g, game.StepBeginCombat)
	if n := settleEnergy(t, g, me.ID, erPlay{target: bear, pay: true}); n != 1 {
		t.Fatalf("answered %d energy prompts, want 1", n)
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("target's power %d, want 3", got)
	}
	assertKeywords(t, g, bear, "haste")
	if energyOf(me) != 0 {
		t.Errorf("energy %d, want 0", energyOf(me))
	}
	if got := effectivePower(t, g, gob); got != 2 {
		t.Errorf("the Goblin itself got the pump: power %d", got)
	}
}

func TestGapConduitGoblinDeclinedOrBrokeGivesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Conduit Goblin", "Creature — Goblin Warrior", gapConduitGoblinOracle, 2, 2)
	bear := b12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	setEnergy(t, g, me, 1)
	advanceTo(t, g, game.StepBeginCombat)
	settleEnergy(t, g, me.ID, erPlay{target: bear, pay: false})
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("declined: power %d, want 2", got)
	}
	if energyOf(me) != 1 {
		t.Errorf("declined: energy %d, want 1", energyOf(me))
	}
}

func TestGapConduitGoblinEntryGivesTwoEnergy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	setEnergy(t, g, me, 0)
	castWithCost(t, g, "Conduit Goblin", "Creature — Goblin Warrior", "", gapConduitGoblinOracle)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Errorf("energy %d after entering, want 2", energyOf(me))
	}
}

func TestGapRecklessPyrosurferLandfallGivesBattleCryForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	surfer := b12Push(g, me.ID, "Reckless Pyrosurfer", "Creature — Human Scout", gapPyrosurferOracle, 2, 2)
	bear := b12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	b31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{surfer, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("the other attacker's power %d, want 3", got)
	}
	if got := effectivePower(t, g, surfer); got != 2 {
		t.Errorf("the Pyrosurfer's own power %d, want 2", got)
	}
}

func TestGapRecklessPyrosurferWithoutALandHasNoBattleCry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	surfer := b12Push(g, me.ID, "Reckless Pyrosurfer", "Creature — Human Scout", gapPyrosurferOracle, 2, 2)
	bear := b12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{surfer, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("no landfall: the other attacker's power %d, want 2", got)
	}
}

func TestGapVoidpouncerKickedEntersBigWithTrampleAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	pouncer := pushHandCardWithManaCost(me, "Voidpouncer", "Creature — Eldrazi", "{1}{R}")
	setOracleInHand(me, pouncer, gapVoidpouncerOracle, 3, 1)
	if err := g.AddManaForEffect(me.ID, me.ID, "{R}{R}{C}{C}{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, pouncer, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast kicked: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePowerWithCounters(t, g, pouncer); got != 5 {
		t.Errorf("kicked power %d, want 5", got)
	}
	assertKeywords(t, g, pouncer, "haste", "trample")
}

func TestGapVoidpouncerUnkickedIsAPlain3By1(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	pouncer := pushHandCardWithManaCost(me, "Voidpouncer", "Creature — Eldrazi", "{1}{R}")
	setOracleInHand(me, pouncer, gapVoidpouncerOracle, 3, 1)
	if err := g.AddManaForEffect(me.ID, me.ID, "{R}{R}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, pouncer, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, pouncer); got != 3 {
		t.Errorf("unkicked power %d, want 3", got)
	}
	c := layeredCard(t, g, pouncer)
	if game.HasKeyword(&c, "haste") || game.HasKeyword(&c, "trample") {
		t.Error("an unkicked Voidpouncer has haste or trample")
	}
}

// setOracleInHand stamps an oracle ID and printed power and toughness onto a card already in a hand.
func setOracleInHand(p *game.Player, id uuid.UUID, oracle string, power, toughness int) {
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			p.Hand.Cards[i].OracleID = oracle
			p.Hand.Cards[i].Power, p.Hand.Cards[i].Toughness = power, toughness
		}
	}
}

func TestGapInfernalCaptorExploitStealsUntapsAndHastes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	victim := pushTappedVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	exploitCast(t, g, me, "Infernal Captor", "Creature — Devil Rogue", gapInfernalCaptorOracle)
	answerOwnPermanents(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s, want the exploiter's controller", got)
	}
	if c, _ := battlefieldCard(g, victim); c.Tapped {
		t.Error("the stolen creature was not untapped")
	}
	assertKeywords(t, g, victim, "haste")
	advanceToNextSeatsTurn(t, g)
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("control did not revert at end of turn: %s", got)
	}
}

func TestGapInfernalCaptorDecliningStealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	exploitCast(t, g, me, "Infernal Captor", "Creature — Devil Rogue", gapInfernalCaptorOracle)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Errorf("a declined exploit moved control to %s", got)
	}
}

func TestGapSkoaEntersForFourAndGrandeurNeedsTheNamedCardAndTwoMountains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skoa := b12Push(g, me.ID, "Skoa, Embermage", "Legendary Creature — Goblin Wizard", gapSkoaOracle, 4, 4)
	m1 := pushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	m2 := pushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	forest := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	toMainForCost(t, g)

	// No other Skoa in hand: refused.
	if err := g.ActivateCatalogAbility(me.ID, skoa, 0, game.ActivateAbilityParams{
		Targets: playerTargets(opp.ID), SacrificeIDs: ids(m1, m2), Strict: true,
	}); err == nil {
		t.Fatal("grandeur activated with no other Skoa in hand")
	}
	copyID := pushCatalogHandCard(me, "Skoa, Embermage", "Legendary Creature — Goblin Wizard", gapSkoaOracle)
	// A Forest is not a Mountain.
	if err := g.ActivateCatalogAbility(me.ID, skoa, 0, game.ActivateAbilityParams{
		Targets: playerTargets(opp.ID), SacrificeIDs: ids(m1, forest), Strict: true, DiscardIDs: ids(copyID),
	}); err == nil {
		t.Fatal("grandeur sacrificed a Forest as a Mountain")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, skoa, 0, game.ActivateAbilityParams{
		Targets: playerTargets(opp.ID), SacrificeIDs: ids(m1, m2), Strict: true, DiscardIDs: ids(copyID),
	}); err != nil {
		t.Fatalf("grandeur: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d -> %d, want -4", life, opp.Life)
	}
	if g.Battlefield.Contains(m1) || g.Battlefield.Contains(m2) || !g.Battlefield.Contains(forest) {
		t.Error("grandeur did not sacrifice exactly the two Mountains")
	}
}

func TestGapSkoaEnterTriggerDealsFourToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castWithCost(t, g, "Skoa, Embermage", "Legendary Creature — Goblin Wizard", "", gapSkoaOracle)
	passPriorityAroundTable(t, g)
	answerTriggerTargets2(t, g, me.ID, opp.ID)
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d -> %d, want -4", life, opp.Life)
	}
}

func TestGapDreamtideWhaleVanishesAndProliferatesOnTheSecondSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	putCounters(t, g, ally, game.CounterPlusOne, 1)
	toMainForCost(t, g)
	whale := pushHandCardWithManaCost(me, "Dreamtide Whale", "Creature — Whale", "")
	setOracleInHand(me, whale, gapDreamtideWhaleOracle, 7, 5)
	if err := g.CastSpell(me.ID, whale, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, whale, game.CounterTime); got != 2 {
		t.Fatalf("time counters %d, want 2 (vanishing 2)", got)
	}
	// The first spell was the Whale itself; the second proliferates.
	castWithCost(t, g, "Shock Stand-in", "Instant", "", "")
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("after the second spell: counters %d, want 2", got)
	}
	// The third does not.
	castWithCost(t, g, "Another Stand-in", "Instant", "", "")
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("after the third spell: counters %d, want 2", got)
	}
}

func TestGapDreamtideWhaleProliferatesOnAnOpponentsSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Dreamtide Whale", "Creature — Whale", gapDreamtideWhaleOracle, 7, 5)
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	putCounters(t, g, ally, game.CounterPlusOne, 1)
	for i := 0; i < 2; i++ {
		castFor(t, g, opp, "Opp Spell", "Instant", "", game.CastSpellParams{})
		passPriorityAroundTable(t, g)
	}
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent's second spell: counters %d, want 2", got)
	}
}

func TestGapCorruptedShapeshifterBecomesTheChosenForm(t *testing.T) {
	for _, tc := range []struct {
		pick       int
		power, tgh int
		keyword    string
	}{
		{0, 3, 3, "flying"},
		{1, 2, 5, "vigilance"},
		{2, 0, 12, "defender"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		toMainForCost(t, g)
		id := pushHandCardWithManaCost(me, "Corrupted Shapeshifter", "Creature — Eldrazi Shapeshifter", "")
		setOracleInHand(me, id, gapShapeshifterOracle, 0, 0)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		answerOptionPick(t, g, me.ID, tc.pick)
		passPriorityAroundTable(t, g)
		if p, th := effectivePower(t, g, id), effectiveToughness(t, g, id); p != tc.power || th != tc.tgh {
			t.Errorf("form %d: %d/%d, want %d/%d", tc.pick, p, th, tc.power, tc.tgh)
		}
		for _, kw := range []string{"flying", "vigilance", "defender"} {
			if got := hasEffectiveKeyword(t, g, id, kw); got != (kw == tc.keyword) {
				t.Errorf("form %d: %s = %v", tc.pick, kw, got)
			}
		}
		assertKeywords(t, g, id, "devoid")
	}
}

func TestGapCopycrookCopiesACreatureAndConnivesWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := castCatalogSpell(t, g, "Copycrook", "Creature — Shapeshifter Rogue", gapCopycrookOracle, nil)
	resolveWithCopyChoice(t, g, bear)
	g.RunStateChecksForTest()
	c := layeredCard(t, g, id)
	if c.Name != "Grizzly Bears" || c.CurrentPower() != 2 {
		t.Fatalf("copy = %q %d/%d, want a Grizzly Bears 2/2", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	unsick(g, id)
	advanceTo(t, g, game.StepDeclareAttackers)
	hand := me.Hand.Size()
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d: the copy did not connive (draw, then discard)", hand, me.Hand.Size())
	}
	if len(g.PendingChoices) == 0 {
		t.Error("connive's discard was not asked")
	}
	// The original Bears has no such ability: it does not connive.
	if layeredCard(t, g, bear).Name != "Grizzly Bears" {
		t.Error("the original changed")
	}
}

func gapSorceryCard(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Green Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Owner: owner, Controller: owner}
}

func gapExileCastable(g *game.Game, me *game.Player, c game.Card) {
	g.WithWriteLock(func() {
		g.Exile.PushTop(c)
		g.GrantCastPermissionOverCardForEffect(c.InstanceID, game.CastPermission{
			Player: me.ID, Zone: game.ZoneExile, Duration: game.WhileInZoneDuration()})
	})
}

func TestGapPartyThrasherGivesConvokeToNoncreatureSpellsFromExileOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Party Thrasher", "Creature — Lizard Wizard", gapPartyThrasherOracle, false)
	elves := cgTwoGreenCreatures(g, me.ID)

	// A noncreature spell from exile: convokable.
	spell := gapSorceryCard(me.ID)
	gapExileCastable(g, me, spell)
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{FromZone: "exile", Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking a noncreature spell from exile: %v", err)
	}
	if !cgTapped(g, elves) {
		t.Error("the convoking creatures were not tapped")
	}
}

func TestGapPartyThrasherGivesNoConvokeToCreatureSpellsOrToSpellsFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Party Thrasher", "Creature — Lizard Wizard", gapPartyThrasherOracle, false)
	elves := cgTwoGreenCreatures(g, me.ID)

	bear := cgBear(me.ID)
	gapExileCastable(g, me, bear)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{FromZone: "exile", Strict: true, TapIDs: elves}); err == nil {
		t.Error("a creature spell cast from exile was convoked")
	}
	spell := gapSorceryCard(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(spell) })
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Error("a noncreature spell cast from hand was convoked")
	}
}

func TestGapPartyThrasherDiscardsToExileTwoAndPlaysTheChosenOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Party Thrasher", "Creature — Lizard Wizard", gapPartyThrasherOracle, 1, 4)
	ids := fireglassLands(me, 3)
	pitch := pushCatalogHandCard(me, "Pitched", "Sorcery", "")
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("pick %+v, want a choice between the two exiled cards", pick)
	}
	if z := zoneOfCard(g, ids[2]); z != game.ZoneLibrary {
		t.Errorf("a third card left the library (%v)", z)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("played the card that was not chosen")
	}
	if err := g.CastSpell(me.ID, ids[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("could not play the chosen card: %v", err)
	}
}

func TestGapPartyThrasherDecliningTheDiscardExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Party Thrasher", "Creature — Lizard Wizard", gapPartyThrasherOracle, 1, 4)
	ids := fireglassLands(me, 3)
	pushCatalogHandCard(me, "Kept", "Sorcery", "")
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) != nil || zoneOfCard(g, ids[0]) != game.ZoneLibrary {
		t.Error("cards were exiled without a discard")
	}
}
