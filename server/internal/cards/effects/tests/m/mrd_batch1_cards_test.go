package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMrdBatch1CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Alpha Myr": effectstest.MrdAlphaMyrOracle, "Altar's Light": effectstest.MrdAltarsLightOracle,
		"Assert Authority": effectstest.MrdAssertAuthorityOracle, "Auriok Bladewarden": effectstest.MrdAuriokBladewardenOracle,
		"Auriok Transfixer": effectstest.MrdAuriokTransfixerOracle, "Battlegrowth": effectstest.MrdBattlegrowthOracle,
		"Blinkmoth Well": effectstest.MrdBlinkmothWellOracle, "Bosh, Iron Golem": effectstest.MrdBoshOracle,
		"Chimney Imp": effectstest.MrdChimneyImpOracle, "Clockwork Beetle": effectstest.MrdClockworkBeetleOracle,
		"Clockwork Condor": effectstest.MrdClockworkCondorOracle, "Clockwork Dragon": effectstest.MrdClockworkDragonOracle,
		"Clockwork Vorrac": effectstest.MrdClockworkVorracOracle, "Cobalt Golem": effectstest.MrdCobaltGolemOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: registered as %q, %v; want its name and Full", name, spec.Name, spec.Completeness)
		}
	}
}

func mrdPlusOneCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.LayeredCard(t, g, id).Counters[game.CounterPlusOne]
}

// --- Altar's Light ---------------------------------------------------------

func TestMrdAltarsLightExilesAnArtifactOrAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.MrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	aura := effectstest.MrdPermanent(g, opp.ID, "Their Enchantment", "Enchantment", "{1}{W}", 0, 0)
	bear := effectstest.SeedBear(g, opp.ID)

	if err := effectstest.CastCatalogSpellErr(t, g, "Altar's Light", "Instant", effectstest.MrdAltarsLightOracle, effectstest.CardRefs(bear)); err == nil {
		t.Error("a creature that is neither an artifact nor an enchantment was accepted as the target")
	}
	effectstest.CastCatalogSpell(t, g, "Altar's Light", "Instant", effectstest.MrdAltarsLightOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, rock) {
		t.Error("the artifact was not exiled")
	}
	effectstest.CastCatalogSpell(t, g, "Altar's Light", "Instant", effectstest.MrdAltarsLightOracle, effectstest.CardRefs(aura))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, aura) {
		t.Error("the enchantment was not exiled")
	}
}

// --- Assert Authority ------------------------------------------------------

func TestMrdAssertAuthorityHasAffinityForArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"})
	}
	effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Signet", TypeLine: "Artifact"})
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.MrdAssertAuthorityOracle, "Instant", "{5}{U}{U}"); got != 4 {
		t.Errorf("with three artifacts it costs mana value %d, want 4 ({2}{U}{U})", got)
	}
	for i := 0; i < 4; i++ {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.MrdAssertAuthorityOracle, "Instant", "{5}{U}{U}"); got != 2 {
		t.Errorf("with seven artifacts it costs mana value %d, want 2 ({U}{U} is never reduced)", got)
	}
}

func TestMrdAssertAuthorityCountersAndExilesTheSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant",
		OracleID: effectstest.B5ShockOracle, Owner: opp.ID, Controller: opp.ID,
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
		OracleID: effectstest.MrdAssertAuthorityOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, authority, game.CastSpellParams{Targets: effectstest.CardRefs(shock)}); err != nil {
		t.Fatalf("cast Assert Authority: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warden := effectstest.PushCatalogPermanent(g, me.ID, "Auriok Bladewarden", "Creature — Human Soldier", effectstest.MrdAuriokBladewardenOracle, false)
	effectstest.FindBattlefieldCardForTest(g, warden).Power = 3
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("bear is %d/%d, want 5/5 (+3/+3 from a 3-power Bladewarden)", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.TappedForTest(t, g, warden) {
		t.Error("the Bladewarden should be tapped to pay the cost")
	}
}

// --- Auriok Transfixer and Blinkmoth Well ----------------------------------

func TestMrdAuriokTransfixerTapsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fixer := effectstest.PushCatalogPermanent(g, me.ID, "Auriok Transfixer", "Creature — Human Scout", effectstest.MrdAuriokTransfixerOracle, false)
	golem := effectstest.MrdPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "{4}", 3, 3)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceToMain(t, g)

	effectstest.B06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, fixer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Fatal("a nonartifact creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, fixer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(golem), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, golem) {
		t.Error("the artifact creature was not tapped")
	}
}

func TestMrdBlinkmothWellTapsOnlyANoncreatureArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	well := effectstest.PushCatalogPermanent(g, me.ID, "Blinkmoth Well", "Land", effectstest.MrdBlinkmothWellOracle, false)
	golem := effectstest.MrdPermanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "{4}", 3, 3)
	rock := effectstest.MrdPermanent(g, opp.ID, "Their Rock", "Artifact", "{2}", 0, 0)
	effectstest.AdvanceToMain(t, g)

	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, well, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(golem), Strict: true}); err == nil {
		t.Fatal("an artifact creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, well, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, rock) || !effectstest.TappedForTest(t, g, well) {
		t.Error("the noncreature artifact and the Well should both be tapped")
	}
}

// --- Battlegrowth ----------------------------------------------------------

func TestMrdBattlegrowthPutsACounterOnTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Battlegrowth", "Instant", effectstest.MrdBattlegrowthOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if n := mrdPlusOneCounters(t, g, bear); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1 (any creature, an opponent's too)", n)
	}
}

// --- Bosh, Iron Golem ------------------------------------------------------

func TestMrdBoshDealsTheSacrificedArtifactsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bosh := effectstest.B12Push(g, me.ID, "Bosh, Iron Golem", "Legendary Artifact Creature — Golem", effectstest.MrdBoshOracle, 6, 7)
	rock := effectstest.MrdPermanent(g, me.ID, "Mind Rock", "Artifact", "{3}", 0, 0)
	if !effectstest.HasEffectiveKeyword(t, g, bosh, "trample") {
		t.Error("Bosh has no trample")
	}
	effectstest.AdvanceToMain(t, g)
	life := opp.Life

	effectstest.B06AddMana(me, "C", "C", "C", "R")
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
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent at %d, want %d (the sacrificed artifact's mana value)", opp.Life, life-3)
	}
}

func TestMrdBoshCanSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bosh := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bosh, Iron Golem", TypeLine: "Legendary Artifact Creature — Golem",
		OracleID: effectstest.MrdBoshOracle, ManaCost: "{8}", Power: 6, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life

	effectstest.B06AddMana(me, "C", "C", "C", "R")
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
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bosh) || !me.Graveyard.Contains(bosh) {
		t.Error("Bosh should be in its owner's graveyard")
	}
	if opp.Life != life-8 {
		t.Errorf("opponent at %d, want %d (Bosh's own mana value, from its last known information)", opp.Life, life-8)
	}
}

// --- Chimney Imp -----------------------------------------------------------

func TestMrdChimneyImpMakesTheTargetOpponentPutACardBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imp := effectstest.PushCatalogPermanent(g, me.ID, "Chimney Imp", "Creature — Imp", effectstest.MrdChimneyImpOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, imp, "flying") {
		t.Error("Chimney Imp has no flying")
	}
	secret := effectstest.HandCardForTest(opp, "Secret Plan", "Sorcery", "")
	hand := effectstest.HandSize(opp)
	effectstest.AdvanceToMain(t, g)

	effectstest.KillCreature(t, g, me.ID, imp)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the dies trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("the Imp's controller was accepted as \"target opponent\"")
	}
	effectstest.AnswerTriggerTargets2(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, secret)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.HandSize(opp); got != hand-1 {
		t.Errorf("opponent's hand %d, want %d", got, hand-1)
	}
	if top, err := opp.Library.Top(); err != nil || top.InstanceID != secret {
		t.Error("the chosen card is not on top of the opponent's library")
	}
}

// --- The Clockwork creatures -----------------------------------------------

func TestMrdClockworkBeetleLosesACounterAtEndOfCombatAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beetle := effectstest.MrdEnter(t, g, me, "Clockwork Beetle", "Artifact Creature — Insect", effectstest.MrdClockworkBeetleOracle)
	if n := mrdPlusOneCounters(t, g, beetle); n != 2 {
		t.Fatalf("Clockwork Beetle entered with %d +1/+1 counters, want 2", n)
	}
	life := opp.Life

	effectstest.DeclareAttack(t, g, opp.ID, beetle)
	if n := effectstest.TriggersOnStackFrom(g, beetle); n != 1 {
		t.Fatalf("attack trigger: %d on the stack, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("defender at %d, want %d: the counter comes off only at end of combat", opp.Life, life-2)
	}
	if n := mrdPlusOneCounters(t, g, beetle); n != 2 {
		t.Fatalf("%d counters before end of combat, want 2", n)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := mrdPlusOneCounters(t, g, beetle); n != 1 {
		t.Errorf("%d counters after end of combat, want 1", n)
	}
}

func TestMrdClockworkCondorLosesACounterAfterBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.MrdPermanent(g, me.ID, "Squire", "Creature — Human Soldier", "{1}{W}", 1, 1)
	condor := effectstest.MrdEnter(t, g, opp, "Clockwork Condor", "Artifact Creature — Bird", effectstest.MrdClockworkCondorOracle)
	if n := mrdPlusOneCounters(t, g, condor); n != 3 {
		t.Fatalf("Clockwork Condor entered with %d +1/+1 counters, want 3", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, condor, "flying") {
		t.Error("Clockwork Condor has no flying")
	}

	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(condor, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.SeedBear(g, me.ID)
	condor := effectstest.MrdEnter(t, g, opp, "Clockwork Condor", "Artifact Creature — Bird", effectstest.MrdClockworkCondorOracle)

	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(condor, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(condor) {
		t.Fatal("the 3/3 Condor should survive 2 combat damage")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(condor) || !opp.Graveyard.Contains(condor) {
		t.Error("a 2/2 Condor with 2 damage should die once the counter is removed")
	}
}

func TestMrdClockworkCounterIsNotRemovedFromANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beetle := effectstest.MrdEnter(t, g, me, "Clockwork Beetle", "Artifact Creature — Insect", effectstest.MrdClockworkBeetleOracle)
	effectstest.DeclareAttack(t, g, opp.ID, beetle)
	effectstest.PassPriorityAroundTable(t, g)

	// The Beetle leaves combat and the battlefield before end of combat:
	// the delayed trigger has nothing left to act on (CR 603.7c).
	effectstest.KillCreature(t, g, me.ID, beetle)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(beetle) || !me.Graveyard.Contains(beetle) {
		t.Error("the Beetle should be in its owner's graveyard")
	}
}

func TestMrdClockworkDragonWindsItselfUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.MrdEnter(t, g, me, "Clockwork Dragon", "Artifact Creature — Dragon", effectstest.MrdClockworkDragonOracle)
	if n := mrdPlusOneCounters(t, g, dragon); n != 6 {
		t.Fatalf("Clockwork Dragon entered with %d +1/+1 counters, want 6", n)
	}
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		effectstest.B06AddMana(me, "C", "C", "C")
		if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", i+1, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := mrdPlusOneCounters(t, g, dragon); n != 8 {
		t.Errorf("%d counters after two activations, want 8", n)
	}
	if effectstest.TappedForTest(t, g, dragon) {
		t.Error("the {3} ability does not tap the Dragon")
	}
}

func TestMrdClockworkVorracTapsToWindUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vorrac := effectstest.MrdEnter(t, g, me, "Clockwork Vorrac", "Artifact Creature — Boar Beast", effectstest.MrdClockworkVorracOracle)
	if n := mrdPlusOneCounters(t, g, vorrac); n != 4 {
		t.Fatalf("Clockwork Vorrac entered with %d +1/+1 counters, want 4", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, vorrac, "trample") {
		t.Error("Clockwork Vorrac has no trample")
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, vorrac, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := mrdPlusOneCounters(t, g, vorrac); n != 5 || !effectstest.TappedForTest(t, g, vorrac) {
		t.Errorf("%d counters (tapped %v), want 5 and tapped", n, effectstest.TappedForTest(t, g, vorrac))
	}
	if err := g.ActivateCatalogAbility(me.ID, vorrac, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a tapped Vorrac activated its {T} ability again")
	}
}

// --- Cobalt Golem ----------------------------------------------------------

func TestMrdCobaltGolemGainsFlyingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	golem := effectstest.PushCatalogPermanent(g, me.ID, "Cobalt Golem", "Artifact Creature — Golem", effectstest.MrdCobaltGolemOracle, false)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, golem, "flying") {
		t.Fatal("Cobalt Golem flies before the ability resolves")
	}
	effectstest.B06AddMana(me, "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, golem, "flying") {
		t.Error("Cobalt Golem did not gain flying")
	}
}
