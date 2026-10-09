package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_e_test.go — slice fra-creature-e (tracker
// #2795): Reality Fracture creatures, one or more tests each.

const (
	oracleSamutHazoret    = "add02a36-8811-438b-a8e1-f5c8a8146b4e"
	oracleSoulbreaker     = "47e33aa1-2f54-471c-a484-de13bcd2dceb"
	oracleCryomancer      = "c3936db3-d3b6-4a92-9773-6f3c52dc419d"
	oraclePegasus         = "7d6641e7-4d14-4a06-b588-0c752af529bd"
	oracleSimulacrum      = "9e45e2da-7064-4ec0-8d09-55fc7d1aeaa8"
	oracleBattlecarver    = "4fc7c7a0-bfff-4833-a0d0-a38dc29e9dfb"
	oracleSolariumSentry  = "afddba56-f9a0-4917-8167-c9fd17ca9954"
	oracleSphinxFalse     = "9f004382-c57b-4fde-af97-0c85516c3cf4"
	oracleStingcaster     = "056b651e-e0e2-4333-9235-d1ffe8fcca29"
	oracleSureshot        = "5bbc9d7d-91bd-4a58-8d19-f7be650720b7"
	oracleTamiyoUpriser   = "8e0f2e26-1e14-422a-aafe-e1da38ab10a3"
	oracleTethermage      = "571cb06d-55b5-46e4-9678-1f74aca73068"
	oracleTetherTech      = "a9d0657f-2c5e-4b6f-b8ae-d657cc0f7aef"
	oracleTetsukoPursuer  = "708d4567-a5a3-437f-bae4-ae10f1537aa9"
	oracleThaliaSurvivor  = "55ef129e-698e-424f-be3c-3fbba6c2cc3e"
	oracleTheoreticalNecr = "6c5ed756-b999-49f7-9a52-fb8196202f0f"
	oracleTinybonesPocket = "97dcf9fb-3f2f-4aa6-959a-e3c4889c1673"
	oracleTitanbones      = "510bccb1-61cd-49f1-a54a-80352653d1e3"
	oracleTomikSparkmage  = "16604fee-cd8d-41e4-8269-3c7ff756613d"
)

func rfCrEFoe(g *game.Game) *game.Player { return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)] }

// --- Samut, Hazoret's Champion -------------------------------------

func TestSamutHazoretsChampionGivesEveryCreatureYouControlHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	samut := rfCrAPush(g, me.ID, "Samut, Hazoret's Champion", "Legendary Creature — Human Warrior Cleric", oracleSamutHazoret, 2, 2)
	bear := rfCrAPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := rfCrAPush(g, foe.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	for _, id := range []uuid.UUID{samut, bear} {
		if !effectiveAbilitiesContain(t, g, id, "haste") {
			t.Fatalf("%s has no haste", id)
		}
	}
	if effectiveAbilitiesContain(t, g, theirs, "haste") {
		t.Fatal("an opponent's creature got haste")
	}
}

// --- Screeching Soulbreaker ----------------------------------------

func TestScreechingSoulbreakerDrainsOneWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	id := rfCrAPush(g, me.ID, "Screeching Soulbreaker", "Creature — Siren Bard", oracleSoulbreaker, 1, 4)
	myLife, theirLife := me.Life, foe.Life
	declareAttack(t, g, foe.ID, id)
	passPriorityAroundTable(t, g)
	if foe.Life != theirLife-1 || me.Life != myLife+1 {
		t.Fatalf("life me %d->%d, foe %d->%d; want +1 and -1", myLife, me.Life, theirLife, foe.Life)
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && p.Life != theirLife-1 {
			t.Fatalf("opponent %s life %d, want %d", p.ID, p.Life, theirLife-1)
		}
	}
}

// --- Shatterwing Pegasus -------------------------------------------

func TestShatterwingPegasusPumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	toMain(t, g)
	peg := rfCrAPush(g, me.ID, "Shatterwing Pegasus", "Creature — Pegasus", oraclePegasus, 2, 3)
	bear := rfCrAPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := rfCrAPush(g, foe.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	rfCrAAddMana(t, g, me, "{W}{W}{W}{W}{W}")
	b16Activate(t, g, me.ID, peg, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := rfCrAPower(t, g, peg); got != 3 {
		t.Fatalf("Pegasus power %d, want 3", got)
	}
	if got := rfCrAPower(t, g, bear); got != 3 {
		t.Fatalf("Bear power %d, want 3", got)
	}
	if got := rfCrAPower(t, g, theirs); got != 2 {
		t.Fatalf("their Bear power %d, want 2", got)
	}
}

// --- Skilled Battlecarver ------------------------------------------

func TestSkilledBattlecarverHasFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	mine := rfCrAPush(g, me.ID, "Skilled Battlecarver", "Creature — Human Warrior", oracleBattlecarver, 2, 1)
	theirs := rfCrAPush(g, foe.ID, "Skilled Battlecarver", "Creature — Human Warrior", oracleBattlecarver, 2, 1)
	if !effectiveAbilitiesContain(t, g, mine, "first strike") {
		t.Fatal("no first strike on its controller's turn")
	}
	if effectiveAbilitiesContain(t, g, theirs, "first strike") {
		t.Fatal("first strike on an opponent's turn")
	}
}

func TestSkilledBattlecarverPumpsItselfForOneRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	id := rfCrAPush(g, me.ID, "Skilled Battlecarver", "Creature — Human Warrior", oracleBattlecarver, 2, 1)
	rfCrAAddMana(t, g, me, "{R}{R}{R}{R}")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := rfCrAPower(t, g, id); got != 4 {
		t.Fatalf("power %d after two pumps, want 4", got)
	}
}

// --- Solarium Sentry -----------------------------------------------

func TestSolariumSentryGainsLifeOnlyForCheapOpposingSpells(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	owner := rfCrEFoe(g)
	rfCrAPush(g, owner.ID, "Solarium Sentry", "Creature — Cat Soldier", oracleSolariumSentry, 3, 3)
	toMain(t, g)
	life := owner.Life
	cast := func(name, cost string) {
		t.Helper()
		id := uuid.New()
		caster.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: cost, Owner: caster.ID, Controller: caster.ID})
		if err := g.CastSpell(caster.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("CastSpell %s: %v", name, err)
		}
		passPriorityAroundTable(t, g)
	}
	cast("Cheap", "{1}{R}")
	if owner.Life != life+2 {
		t.Fatalf("life %d after a mana value 2 spell, want %d", owner.Life, life+2)
	}
	cast("Pricey", "{2}{R}")
	if owner.Life != life+2 {
		t.Fatalf("life %d after a mana value 3 spell, want it unchanged", owner.Life)
	}
}

func TestSolariumSentryIgnoresItsOwnControllersSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfCrAPush(g, me.ID, "Solarium Sentry", "Creature — Cat Soldier", oracleSolariumSentry, 3, 3)
	toMain(t, g)
	life := me.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Cheap", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// --- Thalia, the Survivor ------------------------------------------

func TestThaliaTheSurvivorTaxesOnlyOpposingNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Thalia, the Survivor", "Legendary Creature — Human Soldier", oracleThaliaSurvivor, false)
	if got := priceInHand(t, g, them, "Swords", "Instant", "{W}"); got != 2 {
		t.Errorf("an opponent's noncreature spell costs %d, want 2", got)
	}
	if got := priceInHand(t, g, them, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("an opponent's creature spell costs %d, want 2 (untaxed)", got)
	}
	if got := priceInHand(t, g, me, "Swords", "Instant", "{W}"); got != 1 {
		t.Errorf("Thalia's controller pays %d for a noncreature spell, want 1", got)
	}
}

// --- Simulacrum Shaper ---------------------------------------------

func TestSimulacrumShaperFetchesABasicTappedAndDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id := rfCrACast(t, g, "Simulacrum Shaper", "Creature — Elf Druid", oracleSimulacrum, 2, 2)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
	land, ok := battlefieldCard(g, ids[0])
	if !ok {
		t.Fatal("the basic land did not reach the battlefield")
	}
	if !land.Tapped {
		t.Fatal("the basic land entered untapped")
	}
	hand := me.Hand.Size()
	rfCrADestroy(t, g, id)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after it died, want %d", me.Hand.Size(), hand+1)
	}
}

func TestSimulacrumShaperCanDeclineTheSearch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"))
	rfCrACast(t, g, "Simulacrum Shaper", "Creature — Elf Druid", oracleSimulacrum, 2, 2)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ids[0]) {
		t.Fatal("the land came out although the search was declined")
	}
}

// --- Stingcaster Mage ----------------------------------------------

func TestStingcasterMageGivesAnInstantFlashbackAtItsManaCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := seedGraveyardCard(t, g, "Filler Bolt", "Instant", "test-stingcaster-bolt")
	rfCrACast(t, g, "Stingcaster Mage", "Creature — Human Wizard", oracleStingcaster, 2, 1)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	perm := grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || perm.Cost != "" {
		t.Fatalf("granted permission %+v, want flashback at its mana cost", perm)
	}
}

func TestStingcasterMageRefusesACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := seedGraveyardCard(t, g, "Gy Bear", "Creature — Bear", "test-stingcaster-bear")
	rfCrACast(t, g, "Stingcaster Mage", "Creature — Human Wizard", oracleStingcaster, 2, 1)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Fatal("a creature card was accepted as the target")
		}
	}
	if grantedPermissionOn(g, me.ID, bear, game.ZoneGraveyard).Granted() {
		t.Fatal("a creature card was granted flashback")
	}
}

// --- Sphinx of False Conclusions -----------------------------------

func TestSphinxOfFalseConclusionsLeavesOneTokenCopyWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrACast(t, g, "Sphinx of False Conclusions", "Creature — Sphinx Illusion", oracleSphinxFalse, 4, 2)
	passPriorityAroundTable(t, g)
	rfCrADestroy(t, g, id)
	passPriorityAroundTable(t, g)
	var tok uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Sphinx of False Conclusions" && c.Controller == me.ID {
			tok = c.InstanceID
			if !IsToken(c) {
				t.Fatal("the copy is not a token")
			}
		}
	}
	if tok == uuid.Nil {
		t.Fatal("no copy appeared")
	}
	rfCrADestroy(t, g, tok)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Sphinx of False Conclusions"); n != 0 {
		t.Fatalf("%d Sphinxes after the token died, want none", n)
	}
}

func TestSphinxOfFalseConclusionsLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	id := rfCrAPush(g, me.ID, "Sphinx of False Conclusions", "Creature — Sphinx Illusion", oracleSphinxFalse, 4, 2)
	hand := me.Hand.Size()
	declareAttack(t, g, foe.ID, id)
	passPriorityAroundTable(t, g)
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("no discard owed after the draw")
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d after looting, want %d", me.Hand.Size(), hand)
	}
}

// --- Sureshot Sower ------------------------------------------------

func TestSureshotSowerDiscardsItselfToDestroyAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	toMain(t, g)
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: foe.ID, Controller: foe.ID,
	})
	ground := rfCrAPush(g, foe.ID, "Bear", "Creature — Bear", "", 2, 2)
	sower := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: sower, Name: "Sureshot Sower", TypeLine: "Creature — Human Archer",
		OracleID: oracleSureshot, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
	rfCrAAddMana(t, g, me, "{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, sower, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ground}},
	}); err == nil {
		t.Fatal("a creature without flying was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, sower, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: flyer}},
	}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) {
		t.Fatal("the flyer survived")
	}
	if !g.Battlefield.Contains(ground) {
		t.Fatal("the ground creature was destroyed")
	}
	if !me.Graveyard.Contains(sower) {
		t.Fatal("the Sower was not discarded")
	}
}

// --- Tamiyo, Upriser Crowned ---------------------------------------

func TestTamiyoUpriserCrownedMakesYouTheMonarchWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfCrACast(t, g, "Tamiyo, Upriser Crowned", "Legendary Creature — Moonfolk Warrior", oracleTamiyoUpriser, 3, 5)
	passPriorityAroundTable(t, g)
	if !YoureTheMonarch(g, me.ID) {
		t.Fatal("not the monarch after Tamiyo entered")
	}
}

func TestTamiyoUpriserCrownedStunsCreaturesThatHitTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	rfCrAPush(g, me.ID, "Tamiyo, Upriser Crowned", "Legendary Creature — Moonfolk Warrior", oracleTamiyoUpriser, 3, 5)
	g.WithWriteLock(func() { _ = g.SetMonarchForEffect(me.ID) })
	hitter := rfCrAPush(g, foe.ID, "Hitter", "Creature — Bear", "", 2, 2)
	dealCombatDamageToPlayer(g, hitter, me.ID, 2)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, hitter)
	if !ok || !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Fatalf("hitter %+v (ok %v), want tapped with one stun counter", c, ok)
	}
}

func TestTamiyoUpriserCrownedLeavesCreaturesAloneWhenYouAreNotTheMonarchOrTheDamageIsNoncombat(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	rfCrAPush(g, me.ID, "Tamiyo, Upriser Crowned", "Legendary Creature — Moonfolk Warrior", oracleTamiyoUpriser, 3, 5)
	hitter := rfCrAPush(g, foe.ID, "Hitter", "Creature — Bear", "", 2, 2)
	dealCombatDamageToPlayer(g, hitter, me.ID, 2)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, hitter); c.Tapped || c.Counters[game.CounterStun] != 0 {
		t.Fatal("stunned although Tamiyo's controller is not the monarch")
	}
	g.WithWriteLock(func() { _ = g.SetMonarchForEffect(me.ID) })
	emitNoncombatDamage(g, hitter, me.ID, 2)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, hitter); c.Tapped || c.Counters[game.CounterStun] != 0 {
		t.Fatal("stunned by noncombat damage")
	}
}

// --- Tenured Tethermage --------------------------------------------

func TestTenuredTethermageSacrificesALandForTwoTappedHeartwoods(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := rfCrAPush(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	rfCrACast(t, g, "Tenured Tethermage", "Creature — Human Artificer", oracleTethermage, 1, 1)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Fatal("the land was not sacrificed")
	}
	if n := b16CountNamed(g, "Heartwood"); n != 2 {
		t.Fatalf("%d Heartwood tokens, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Heartwood" && !c.Tapped {
			t.Fatal("a Heartwood entered untapped")
		}
	}
}

func TestTenuredTethermageMakesNothingWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := rfCrAPush(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	rfCrACast(t, g, "Tenured Tethermage", "Creature — Human Artificer", oracleTethermage, 1, 1)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || b16CountNamed(g, "Heartwood") != 0 {
		t.Fatal("declining still cost a land or made tokens")
	}
}

func TestTenuredTethermageTapsTwoArtifactsForTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	mage := rfCrAPush(g, me.ID, "Tenured Tethermage", "Creature — Human Artificer", oracleTethermage, 1, 1)
	a := rfCrAPush(g, me.ID, "Rock", "Artifact", "", 0, 0)
	b := rfCrAPush(g, me.ID, "Rock", "Artifact", "", 0, 0)
	b16Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}})
	passPriorityAroundTable(t, g)
	if got := rfCrACounters(g, mage, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d counters, want 2", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := battlefieldCard(g, id); !c.Tapped {
			t.Fatal("an artifact was not tapped to pay")
		}
	}
}

// --- Tether Technician ---------------------------------------------

func TestTetherTechnicianDiscardsACardToDealTwoDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	rfCrACast(t, g, "Tether Technician", "Creature — Minotaur Artificer", oracleTetherTech, 4, 5)
	passPriorityAroundTable(t, g)
	life := foe.Life
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	pickPlayerTarget(t, g, me.ID, foe.ID)
	passPriorityAroundTable(t, g)
	if foe.Life != life-2 {
		t.Fatalf("life %d, want %d", foe.Life, life-2)
	}
}

func TestTetherTechnicianDiscardingNothingDealsNoDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	rfCrACast(t, g, "Tether Technician", "Creature — Minotaur Artificer", oracleTetherTech, 4, 5)
	passPriorityAroundTable(t, g)
	life := foe.Life
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || foe.Life != life {
		t.Fatal("a target was asked for or damage dealt without a discard")
	}
}

// --- Tetsuko Umezawa, Pursuer --------------------------------------

func TestTetsukoPingsTheControllerOfAWeakBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	rfCrAPush(g, me.ID, "Tetsuko Umezawa, Pursuer", "Legendary Creature — Human Mercenary", oracleTetsukoPursuer, 2, 4)
	weak := rfCrAPush(g, foe.ID, "Weak", "Creature — Bear", "", 1, 3)
	frail := rfCrAPush(g, foe.ID, "Frail", "Creature — Bear", "", 3, 1)
	big := rfCrAPush(g, foe.ID, "Big", "Creature — Bear", "", 3, 3)
	attacker := rfCrAPush(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2)
	life := foe.Life
	block := func(blocker uuid.UUID) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: foe.ID, CardID: blocker, Target: attacker, Amount: 1})
		})
		passPriorityAroundTable(t, g)
	}
	block(weak)
	block(frail)
	if foe.Life != life-2 {
		t.Fatalf("life %d after two weak blockers, want %d", foe.Life, life-2)
	}
	block(big)
	if foe.Life != life-2 {
		t.Fatalf("a 3/3 blocker cost %d life", life-2-foe.Life)
	}
}

func TestTetsukoIgnoresItsOwnControllersBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	rfCrAPush(g, me.ID, "Tetsuko Umezawa, Pursuer", "Legendary Creature — Human Mercenary", oracleTetsukoPursuer, 2, 4)
	mine := rfCrAPush(g, me.ID, "Weak", "Creature — Bear", "", 1, 1)
	attacker := rfCrAPush(g, foe.ID, "Attacker", "Creature — Bear", "", 2, 2)
	life := me.Life
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: mine, Target: attacker, Amount: 1})
	})
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// --- Theoretical Necromancer ---------------------------------------

func TestTheoreticalNecromancerExilesItselfToReturnAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	necro := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: necro, Name: "Theoretical Necromancer", TypeLine: "Creature — Vampire Warlock",
		OracleID: oracleTheoreticalNecr, Power: 4, Toughness: 1, Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Gy Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	rfCrAAddMana(t, g, me, "{B}{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: necro}},
	}); err == nil {
		t.Fatal("the Necromancer targeted itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate from graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Fatal("the other creature did not return to hand")
	}
	if !g.Exile.Contains(necro) {
		t.Fatal("the Necromancer was not exiled")
	}
}

// --- Tinybones, Pocket Nuisance ------------------------------------

func TestTinybonesPocketNuisanceTriggersOncePerDiscardingPlayerPerBatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfCrACast(t, g, "Tinybones, Pocket Nuisance", "Legendary Creature — Skeleton Rogue", oracleTinybonesPocket, 2, 1)
	passPriorityAroundTable(t, g)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
		if p.ID != me.ID {
			discardFromHand(t, g, p.ID)
		}
	}
	passPriorityAroundTable(t, g)
	// Three opponents each discarded one card: three triggers, each pinging
	// each of the three opponents once.
	for _, p := range g.Seats {
		if p.ID != me.ID && p.Life != lives[p.ID]-3 {
			t.Fatalf("opponent %s life %d, want %d", p.ID, p.Life, lives[p.ID]-3)
		}
	}
	if me.Life != lives[me.ID] {
		t.Fatalf("own life changed to %d", me.Life)
	}
	// One player discarding two cards at once is one trigger.
	foe := rfCrEFoe(g)
	before := foe.Life
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 2) })
	passPriorityAroundTable(t, g)
	if before-foe.Life != 1 {
		t.Fatalf("a two-card discard cost the opponent %d life, want 1", before-foe.Life)
	}
}

// --- Titanbones, Towering Heart ------------------------------------

func TestTitanbonesGrowsByTwoWheneverYouGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfCrAPush(g, me.ID, "Titanbones, Towering Heart", "Legendary Creature — Skeleton Druid", oracleTitanbones, 4, 3)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := rfCrACounters(g, id, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d counters, want 2", got)
	}
}

func TestTitanbonesGainsThreeLifeWhenDiscarded(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Titanbones, Towering Heart", TypeLine: "Legendary Creature — Skeleton Druid",
		OracleID: oracleTitanbones, Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID})
	me.Hand.Cards = me.Hand.Cards[len(me.Hand.Cards)-1:]
	life := me.Life
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Fatalf("life %d, want %d", me.Life, life+3)
	}
}

// --- Tomik, Izzet Sparkmage ----------------------------------------

func TestTomikAddsOneToNoncombatDamageAgainstOpponentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	tomik := rfCrAPush(g, me.ID, "Tomik, Izzet Sparkmage", "Legendary Creature — Human Wizard", oracleTomikSparkmage, 1, 2)
	src := rfCrAPush(g, me.ID, "Pinger", "Creature — Bear", "", 1, 1)
	bear := rfCrAPush(g, foe.ID, "Their Bear", "Creature — Bear", "", 5, 5)
	life := foe.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, foe.ID, 2) })
	if foe.Life != life-3 {
		t.Fatalf("life %d after 2 damage, want %d", foe.Life, life-3)
	}
	mine := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 2) })
	if me.Life != mine-2 {
		t.Fatalf("own life %d, want %d (no bonus)", me.Life, mine-2)
	}
	var dmg int
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(src, bear, 2)
		c, _ := g.LookupCardForEffect(bear)
		dmg = c.DamageMarked
	})
	if dmg != 3 {
		t.Fatalf("damage marked on their creature %d, want 3", dmg)
	}
	// A source controlled by the opponent is not boosted.
	theirSrc := rfCrAPush(g, foe.ID, "Their Pinger", "Creature — Bear", "", 1, 1)
	pre := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirSrc, me.ID, 2) })
	if me.Life != pre-2 {
		t.Fatalf("an opposing source dealt %d, want 2", pre-me.Life)
	}
	_ = tomik
}

func TestTomikDoesNotBoostCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	rfCrAPush(g, me.ID, "Tomik, Izzet Sparkmage", "Legendary Creature — Human Wizard", oracleTomikSparkmage, 1, 2)
	atk := rfCrAPush(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2)
	life := foe.Life
	declareAttack(t, g, foe.ID, atk)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if foe.Life != life-2 {
		t.Fatalf("combat damage took %d life, want 2", life-foe.Life)
	}
}

// --- Seasoned Cryomancer -------------------------------------------

func TestSeasonedCryomancerLootsTwoThenStunsOnePerNonlandDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	me.Hand.Cards = nil
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	victim := rfCrAPush(g, foe.ID, "Victim", "Creature — Bear", "", 2, 2)
	other := rfCrAPush(g, foe.ID, "Other", "Creature — Bear", "", 2, 2)
	rfCrACast(t, g, "Seasoned Cryomancer", "Creature — Human Wizard", oracleCryomancer, 2, 2)
	passPriorityAroundTable(t, g)
	// Hand: the Island and one Spell. Pitch both: one nonland, so one target.
	spellInHand := uuid.Nil
	for _, c := range me.Hand.Cards {
		if c.Name == "Spell" {
			spellInHand = c.InstanceID
		}
	}
	if me.Hand.Size() != 2 || spellInHand == uuid.Nil {
		t.Fatalf("hand %d after drawing two, want 2 including a Spell", me.Hand.Size())
	}
	answerDiscard(t, g, me.ID, land, spellInHand)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	v, _ := battlefieldCard(g, victim)
	o, _ := battlefieldCard(g, other)
	if !v.Tapped || v.Counters[game.CounterStun] != 1 {
		t.Fatalf("the chosen creature: tapped %v, stun %d", v.Tapped, v.Counters[game.CounterStun])
	}
	if o.Tapped || o.Counters[game.CounterStun] != 0 {
		t.Fatal("a creature that was not chosen was affected")
	}
}

func TestSeasonedCryomancerDiscardingOnlyLandsMakesNoTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	me.Hand.Cards = nil
	for i := 0; i < 2; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	}
	victim := rfCrAPush(g, foe.ID, "Victim", "Creature — Bear", "", 2, 2)
	rfCrACast(t, g, "Seasoned Cryomancer", "Creature — Human Wizard", oracleCryomancer, 2, 2)
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a target prompt opened after discarding only lands")
	}
	if v, _ := battlefieldCard(g, victim); v.Tapped {
		t.Fatal("a creature was tapped")
	}
}

func TestSeasonedCryomancerDrawsTwoFromTheGraveyardAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Seasoned Cryomancer", TypeLine: "Creature — Human Wizard",
		OracleID: oracleCryomancer, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	rfCrAAddMana(t, g, me, "{U}{U}{U}{U}{U}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate from graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || !g.Exile.Contains(id) {
		t.Fatalf("hand %d (want %d), exiled %v", me.Hand.Size(), hand+2, g.Exile.Contains(id))
	}
}

// Real combat: two creatures hit the monarch in the same damage step. The
// CR 725.2 monarch steal runs as the first one connects, but the damage is
// simultaneous, so both must be stunned.
func TestTamiyoUpriserCrownedStunsEveryCreatureOfASimultaneousHit(t *testing.T) {
	g := newCatalogGame(t)
	atkr, def := g.Seats[g.Turn.ActiveSeat], rfCrEFoe(g)
	rfCrAPush(g, def.ID, "Tamiyo, Upriser Crowned", "Legendary Creature — Moonfolk Warrior", oracleTamiyoUpriser, 3, 5)
	g.WithWriteLock(func() { _ = g.SetMonarchForEffect(def.ID) })
	a := rfCrAPush(g, atkr.ID, "A", "Creature — Bear", "", 2, 2)
	b := rfCrAPush(g, atkr.ID, "B", "Creature — Bear", "", 2, 2)
	attackWith(t, g, def.ID, a, b)
	annSettle(t, g, nil)
	for _, id := range []uuid.UUID{a, b} {
		c, ok := battlefieldCard(g, id)
		if !ok || c.Counters[game.CounterStun] != 1 {
			t.Fatalf("creature %s: ok %v, stun %d, want 1", id, ok, c.Counters[game.CounterStun])
		}
	}
}
