package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mrd_batch1_cards_test.go — Mirrodin, batch 1 (#5606, S82): the cards
// whose names start with A to C.

const (
	mrdAlphaMyrOracle          = "4caa8e68-e599-416f-b839-3d009569ff29"
	mrdAltarsLightOracle       = "fa9b6be2-b88c-4302-b7e2-faf25a60bcb9"
	mrdAssertAuthorityOracle   = "085bc7be-bd44-40af-8e8f-a1a8006fe22c"
	mrdAuriokBladewardenOracle = "2884e332-707c-4a97-adc1-e1317a513ff4"
	mrdAuriokTransfixerOracle  = "0ad5f4c5-f6d2-4e96-af65-5488d9711113"
	mrdBattlegrowthOracle      = "650bf82e-7f83-470c-bf4a-34281bfe9341"
	mrdBlinkmothWellOracle     = "dd3fd7cd-d8a1-48e2-861d-c7c40086dc75"
	mrdBoshOracle              = "2fb6f65f-a0fb-4d45-be1b-d6405e61d47a"
	mrdChimneyImpOracle        = "3901bf30-b7c1-4977-a7b1-fcdafcc266cd"
	mrdClockworkBeetleOracle   = "7b9ddff4-82f5-4068-8105-116b8a6180d5"
	mrdClockworkCondorOracle   = "3eff8f21-9a1b-4afc-8a7b-75a740d0a4ea"
	mrdClockworkDragonOracle   = "3361cf88-171a-4ddd-ac56-a31d42613c80"
	mrdClockworkVorracOracle   = "e1a64bba-eaff-4f59-b85d-7df0b13e2e6b"
	mrdCobaltGolemOracle       = "1e50af57-e49d-4b94-a370-44846d9f6b33"
)

func TestMrdBatch1CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Alpha Myr": mrdAlphaMyrOracle, "Altar's Light": mrdAltarsLightOracle,
		"Assert Authority": mrdAssertAuthorityOracle, "Auriok Bladewarden": mrdAuriokBladewardenOracle,
		"Auriok Transfixer": mrdAuriokTransfixerOracle, "Battlegrowth": mrdBattlegrowthOracle,
		"Blinkmoth Well": mrdBlinkmothWellOracle, "Bosh, Iron Golem": mrdBoshOracle,
		"Chimney Imp": mrdChimneyImpOracle, "Clockwork Beetle": mrdClockworkBeetleOracle,
		"Clockwork Condor": mrdClockworkCondorOracle, "Clockwork Dragon": mrdClockworkDragonOracle,
		"Clockwork Vorrac": mrdClockworkVorracOracle, "Cobalt Golem": mrdCobaltGolemOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered as %q, %v; want its name and Full", name, spec.Name, spec.Completeness)
		}
	}
}

// mrdPermanent puts a plain permanent on the battlefield.
func mrdPermanent(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// mrdEnter puts a catalog card onto the battlefield from its owner's hand,
// so its entry replacements apply, and lets it act this turn.
func mrdEnter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := pr8Enter(t, g, owner, name, typeLine, oracle)
	findBattlefieldCardForTest(g, id).SummonedThisTurn = false
	return id
}

func mrdPlusOneCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return layeredCard(t, g, id).Counters[game.CounterPlusOne]
}

// --- Altar's Light ---------------------------------------------------------

func TestMrdAltarsLightExilesAnArtifactOrAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := mrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	aura := mrdPermanent(g, opp.ID, "Their Enchantment", "Enchantment", "{1}{W}", 0, 0)
	bear := seedBear(g, opp.ID)

	if err := castCatalogSpellErr(t, g, "Altar's Light", "Instant", mrdAltarsLightOracle, cardRefs(bear)); err == nil {
		t.Error("a creature that is neither an artifact nor an enchantment was accepted as the target")
	}
	castCatalogSpell(t, g, "Altar's Light", "Instant", mrdAltarsLightOracle, cardRefs(rock))
	passPriorityAroundTable(t, g)
	if !inExile(g, rock) {
		t.Error("the artifact was not exiled")
	}
	castCatalogSpell(t, g, "Altar's Light", "Instant", mrdAltarsLightOracle, cardRefs(aura))
	passPriorityAroundTable(t, g)
	if !inExile(g, aura) {
		t.Error("the enchantment was not exiled")
	}
}

// --- Assert Authority ------------------------------------------------------

func TestMrdAssertAuthorityHasAffinityForArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"})
	}
	pushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	if got := selfPricedMV(t, g, me, mrdAssertAuthorityOracle, "Instant", "{5}{U}{U}"); got != 4 {
		t.Errorf("with three artifacts it costs mana value %d, want 4 ({2}{U}{U})", got)
	}
	for i := 0; i < 4; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	if got := selfPricedMV(t, g, me, mrdAssertAuthorityOracle, "Instant", "{5}{U}{U}"); got != 2 {
		t.Errorf("with seven artifacts it costs mana value %d, want 2 ({U}{U} is never reduced)", got)
	}
}

func TestMrdAssertAuthorityCountersAndExilesTheSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant",
		OracleID: b5ShockOracle, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("cast Shock: %v", err)
	}
	life := me.Life
	authority := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: authority, Name: "Assert Authority", TypeLine: "Instant",
		OracleID: mrdAssertAuthorityOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, authority, game.CastSpellParams{Targets: cardRefs(shock)}); err != nil {
		t.Fatalf("cast Assert Authority: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(shock) || opp.Graveyard.Contains(shock) {
		t.Error("the countered spell should be exiled instead of going to its owner's graveyard")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d: the countered Shock dealt no damage", me.Life, life)
	}
	if !me.Graveyard.Contains(authority) {
		t.Error("Assert Authority itself should be in its owner's graveyard")
	}
}

// --- Auriok Bladewarden ----------------------------------------------------

func TestMrdAuriokBladewardenPumpsByItsPowerAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	warden := pushCatalogPermanent(g, me.ID, "Auriok Bladewarden", "Creature — Human Soldier", mrdAuriokBladewardenOracle, false)
	findBattlefieldCardForTest(g, warden).Power = 3
	bear := seedBear(g, me.ID)
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, bear); c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("bear is %d/%d, want 5/5 (+3/+3 from a 3-power Bladewarden)", c.CurrentPower(), c.CurrentToughness())
	}
	if !tappedForTest(t, g, warden) {
		t.Error("the Bladewarden should be tapped to pay the cost")
	}
}

// --- Auriok Transfixer and Blinkmoth Well ----------------------------------

func TestMrdAuriokTransfixerTapsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fixer := pushCatalogPermanent(g, me.ID, "Auriok Transfixer", "Creature — Human Scout", mrdAuriokTransfixerOracle, false)
	golem := mrdPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "{4}", 3, 3)
	bear := seedBear(g, opp.ID)
	advanceToMain(t, g)

	b06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, fixer, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("a nonartifact creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, fixer, 0, game.ActivateAbilityParams{Targets: cardRefs(golem), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, golem) {
		t.Error("the artifact creature was not tapped")
	}
}

func TestMrdBlinkmothWellTapsOnlyANoncreatureArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	well := pushCatalogPermanent(g, me.ID, "Blinkmoth Well", "Land", mrdBlinkmothWellOracle, false)
	golem := mrdPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "{4}", 3, 3)
	rock := mrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	advanceToMain(t, g)

	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, well, 0, game.ActivateAbilityParams{Targets: cardRefs(golem), Strict: true}); err == nil {
		t.Fatal("an artifact creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, well, 0, game.ActivateAbilityParams{Targets: cardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, rock) || !tappedForTest(t, g, well) {
		t.Error("the noncreature artifact and the Well should both be tapped")
	}
}

// --- Battlegrowth ----------------------------------------------------------

func TestMrdBattlegrowthPutsACounterOnTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := seedBear(g, opp.ID)
	castCatalogSpell(t, g, "Battlegrowth", "Instant", mrdBattlegrowthOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if n := mrdPlusOneCounters(t, g, bear); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1 (any creature, an opponent's too)", n)
	}
}

// --- Bosh, Iron Golem ------------------------------------------------------

func TestMrdBoshDealsTheSacrificedArtifactsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bosh := b12Push(g, me.ID, "Bosh, Iron Golem", "Legendary Artifact Creature — Golem", mrdBoshOracle, 6, 7)
	rock := mrdPermanent(g, me.ID, "Mind Rock", "Artifact", "{3}", 0, 0)
	if !hasEffectiveKeyword(t, g, bosh, "trample") {
		t.Error("Bosh has no trample")
	}
	advanceToMain(t, g)
	life := opp.Life

	b06AddMana(me, "C", "C", "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, bosh, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		SacrificeIDs: []uuid.UUID{rock},
		Strict:       true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact should be sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent at %d, want %d (the sacrificed artifact's mana value)", opp.Life, life-3)
	}
}

func TestMrdBoshCanSacrificeItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bosh := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bosh, Iron Golem", TypeLine: "Legendary Artifact Creature — Golem",
		OracleID: mrdBoshOracle, ManaCost: "{8}", Power: 6, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	bear := seedBear(g, me.ID)
	advanceToMain(t, g)
	life := opp.Life

	b06AddMana(me, "C", "C", "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, bosh, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		SacrificeIDs: []uuid.UUID{bear},
		Strict:       true,
	}); err == nil {
		t.Fatal("a creature that is not an artifact paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, bosh, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		SacrificeIDs: []uuid.UUID{bosh},
		Strict:       true,
	}); err != nil {
		t.Fatalf("activate sacrificing Bosh: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bosh) || !me.Graveyard.Contains(bosh) {
		t.Error("Bosh should be in its owner's graveyard")
	}
	if opp.Life != life-8 {
		t.Errorf("opponent at %d, want %d (Bosh's own mana value, from its last known information)", opp.Life, life-8)
	}
}

// --- Chimney Imp -----------------------------------------------------------

func TestMrdChimneyImpMakesTheTargetOpponentPutACardBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imp := pushCatalogPermanent(g, me.ID, "Chimney Imp", "Creature — Imp", mrdChimneyImpOracle, false)
	if !hasEffectiveKeyword(t, g, imp, "flying") {
		t.Error("Chimney Imp has no flying")
	}
	secret := handCardForTest(opp, "Secret Plan", "Sorcery", "")
	hand := handSize(opp)
	advanceToMain(t, g)

	killCreature(t, g, me.ID, imp)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the dies trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("the Imp's controller was accepted as \"target opponent\"")
	}
	answerTriggerTargets2(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, secret)
	passPriorityAroundTable(t, g)

	if got := handSize(opp); got != hand-1 {
		t.Errorf("opponent's hand %d, want %d", got, hand-1)
	}
	if top, err := opp.Library.Top(); err != nil || top.InstanceID != secret {
		t.Error("the chosen card is not on top of the opponent's library")
	}
}

// --- The Clockwork creatures -----------------------------------------------

func TestMrdClockworkBeetleLosesACounterAtEndOfCombatAfterAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beetle := mrdEnter(t, g, me, "Clockwork Beetle", "Artifact Creature — Insect", mrdClockworkBeetleOracle)
	if n := mrdPlusOneCounters(t, g, beetle); n != 2 {
		t.Fatalf("Clockwork Beetle entered with %d +1/+1 counters, want 2", n)
	}
	life := opp.Life

	declareAttack(t, g, opp.ID, beetle)
	if n := triggersOnStackFrom(g, beetle); n != 1 {
		t.Fatalf("attack trigger: %d on the stack, want 1", n)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("defender at %d, want %d: the counter comes off only at end of combat", opp.Life, life-2)
	}
	if n := mrdPlusOneCounters(t, g, beetle); n != 2 {
		t.Fatalf("%d counters before end of combat, want 2", n)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if n := mrdPlusOneCounters(t, g, beetle); n != 1 {
		t.Errorf("%d counters after end of combat, want 1", n)
	}
}

func TestMrdClockworkCondorLosesACounterAfterBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := mrdPermanent(g, me.ID, "Squire", "Creature — Human Soldier", "{1}{W}", 1, 1)
	condor := mrdEnter(t, g, opp, "Clockwork Condor", "Artifact Creature — Bird", mrdClockworkCondorOracle)
	if n := mrdPlusOneCounters(t, g, condor); n != 3 {
		t.Fatalf("Clockwork Condor entered with %d +1/+1 counters, want 3", n)
	}
	if !hasEffectiveKeyword(t, g, condor, "flying") {
		t.Error("Clockwork Condor has no flying")
	}

	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(condor, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if n := mrdPlusOneCounters(t, g, condor); n != 2 {
		t.Errorf("a blocking Condor has %d counters after end of combat, want 2", n)
	}
	if g.Battlefield.Contains(attacker) {
		t.Error("the 1/1 attacker should have died to the 3/3 Condor")
	}
}

// The counter comes off after combat damage, so a Condor that blocked a
// 2/2 is a 2/2 with 2 damage once it is removed, and dies (CR 704.5g).
func TestMrdClockworkCondorDiesWhenItShrinksAfterTakingDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := seedBear(g, me.ID)
	condor := mrdEnter(t, g, opp, "Clockwork Condor", "Artifact Creature — Bird", mrdClockworkCondorOracle)

	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(condor, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(condor) {
		t.Fatal("the 3/3 Condor should survive 2 combat damage")
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(condor) || !opp.Graveyard.Contains(condor) {
		t.Error("a 2/2 Condor with 2 damage should die once the counter is removed")
	}
}

func TestMrdClockworkCounterIsNotRemovedFromANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beetle := mrdEnter(t, g, me, "Clockwork Beetle", "Artifact Creature — Insect", mrdClockworkBeetleOracle)
	declareAttack(t, g, opp.ID, beetle)
	passPriorityAroundTable(t, g)

	// The Beetle leaves combat and the battlefield before end of combat:
	// the delayed trigger has nothing left to act on (CR 603.7c).
	killCreature(t, g, me.ID, beetle)
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(beetle) || !me.Graveyard.Contains(beetle) {
		t.Error("the Beetle should be in its owner's graveyard")
	}
}

func TestMrdClockworkDragonWindsItselfUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dragon := mrdEnter(t, g, me, "Clockwork Dragon", "Artifact Creature — Dragon", mrdClockworkDragonOracle)
	if n := mrdPlusOneCounters(t, g, dragon); n != 6 {
		t.Fatalf("Clockwork Dragon entered with %d +1/+1 counters, want 6", n)
	}
	advanceToMain(t, g)
	for i := 0; i < 2; i++ {
		b06AddMana(me, "C", "C", "C")
		if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	if n := mrdPlusOneCounters(t, g, dragon); n != 8 {
		t.Errorf("%d counters after two activations, want 8", n)
	}
	if tappedForTest(t, g, dragon) {
		t.Error("the {3} ability does not tap the Dragon")
	}
}

func TestMrdClockworkVorracTapsToWindUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vorrac := mrdEnter(t, g, me, "Clockwork Vorrac", "Artifact Creature — Boar Beast", mrdClockworkVorracOracle)
	if n := mrdPlusOneCounters(t, g, vorrac); n != 4 {
		t.Fatalf("Clockwork Vorrac entered with %d +1/+1 counters, want 4", n)
	}
	if !hasEffectiveKeyword(t, g, vorrac, "trample") {
		t.Error("Clockwork Vorrac has no trample")
	}
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, vorrac, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := mrdPlusOneCounters(t, g, vorrac); n != 5 || !tappedForTest(t, g, vorrac) {
		t.Errorf("%d counters (tapped %v), want 5 and tapped", n, tappedForTest(t, g, vorrac))
	}
	if err := g.ActivateCatalogAbility(me.ID, vorrac, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a tapped Vorrac activated its {T} ability again")
	}
}

// --- Cobalt Golem ----------------------------------------------------------

func TestMrdCobaltGolemGainsFlyingUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	golem := pushCatalogPermanent(g, me.ID, "Cobalt Golem", "Artifact Creature — Golem", mrdCobaltGolemOracle, false)
	advanceToMain(t, g)
	if hasEffectiveKeyword(t, g, golem, "flying") {
		t.Fatal("Cobalt Golem flies before the ability resolves")
	}
	b06AddMana(me, "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, golem, "flying") {
		t.Error("Cobalt Golem did not gain flying")
	}
}
