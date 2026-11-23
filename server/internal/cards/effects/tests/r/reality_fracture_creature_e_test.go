package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Samut, Hazoret's Champion -------------------------------------

func TestSamutHazoretsChampionGivesEveryCreatureYouControlHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	samut := effectstest.RfCrAPush(g, me.ID, "Samut, Hazoret's Champion", "Legendary Creature — Human Warrior Cleric", effectstest.OracleSamutHazoret, 2, 2)
	bear := effectstest.RfCrAPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.RfCrAPush(g, foe.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	for _, id := range []uuid.UUID{samut, bear} {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "haste") {
			t.Fatalf("%s has no haste", id)
		}
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "haste") {
		t.Fatal("an opponent's creature got haste")
	}
}

// --- Screeching Soulbreaker ----------------------------------------

func TestScreechingSoulbreakerDrainsOneWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	id := effectstest.RfCrAPush(g, me.ID, "Screeching Soulbreaker", "Creature — Siren Bard", effectstest.OracleSoulbreaker, 1, 4)
	myLife, theirLife := me.Life, foe.Life
	effectstest.DeclareAttack(t, g, foe.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	effectstest.ToMain(t, g)
	peg := effectstest.RfCrAPush(g, me.ID, "Shatterwing Pegasus", "Creature — Pegasus", effectstest.OraclePegasus, 2, 3)
	bear := effectstest.RfCrAPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.RfCrAPush(g, foe.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.RfCrAAddMana(t, g, me, "{W}{W}{W}{W}{W}")
	effectstest.B16Activate(t, g, me.ID, peg, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	mine := effectstest.RfCrAPush(g, me.ID, "Skilled Battlecarver", "Creature — Human Warrior", effectstest.OracleBattlecarver, 2, 1)
	theirs := effectstest.RfCrAPush(g, foe.ID, "Skilled Battlecarver", "Creature — Human Warrior", effectstest.OracleBattlecarver, 2, 1)
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "first strike") {
		t.Fatal("no first strike on its controller's turn")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "first strike") {
		t.Fatal("first strike on an opponent's turn")
	}
}

func TestSkilledBattlecarverPumpsItselfForOneRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	id := effectstest.RfCrAPush(g, me.ID, "Skilled Battlecarver", "Creature — Human Warrior", effectstest.OracleBattlecarver, 2, 1)
	effectstest.RfCrAAddMana(t, g, me, "{R}{R}{R}{R}")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfCrAPower(t, g, id); got != 4 {
		t.Fatalf("power %d after two pumps, want 4", got)
	}
}

// --- Solarium Sentry -----------------------------------------------

func TestSolariumSentryGainsLifeOnlyForCheapOpposingSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	owner := effectstest.RfCrEFoe(g)
	effectstest.RfCrAPush(g, owner.ID, "Solarium Sentry", "Creature — Cat Soldier", effectstest.OracleSolariumSentry, 3, 3)
	effectstest.ToMain(t, g)
	life := owner.Life
	cast := func(name, cost string) {
		t.Helper()
		id := uuid.New()
		caster.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: cost, Owner: caster.ID, Controller: caster.ID})
		if err := g.CastSpell(caster.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("CastSpell %s: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RfCrAPush(g, me.ID, "Solarium Sentry", "Creature — Cat Soldier", effectstest.OracleSolariumSentry, 3, 3)
	effectstest.ToMain(t, g)
	life := me.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Cheap", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// --- Thalia, the Survivor ------------------------------------------

func TestThaliaTheSurvivorTaxesOnlyOpposingNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Thalia, the Survivor", "Legendary Creature — Human Soldier", effectstest.OracleThaliaSurvivor, false)
	if got := effectstest.PriceInHand(t, g, them, "Swords", "Instant", "{W}"); got != 2 {
		t.Errorf("an opponent's noncreature spell costs %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("an opponent's creature spell costs %d, want 2 (untaxed)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Swords", "Instant", "{W}"); got != 1 {
		t.Errorf("Thalia's controller pays %d for a noncreature spell, want 1", got)
	}
}

// --- Simulacrum Shaper ---------------------------------------------

func TestSimulacrumShaperFetchesABasicTappedAndDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id := effectstest.RfCrACast(t, g, "Simulacrum Shaper", "Creature — Elf Druid", effectstest.OracleSimulacrum, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
	land, ok := effectstest.BattlefieldCard(g, ids[0])
	if !ok {
		t.Fatal("the basic land did not reach the battlefield")
	}
	if !land.Tapped {
		t.Fatal("the basic land entered untapped")
	}
	hand := me.Hand.Size()
	rfCrADestroy(t, g, id)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after it died, want %d", me.Hand.Size(), hand+1)
	}
}

func TestSimulacrumShaperCanDeclineTheSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))
	effectstest.RfCrACast(t, g, "Simulacrum Shaper", "Creature — Elf Druid", effectstest.OracleSimulacrum, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ids[0]) {
		t.Fatal("the land came out although the search was declined")
	}
}

// --- Stingcaster Mage ----------------------------------------------

func TestStingcasterMageGivesAnInstantFlashbackAtItsManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := effectstest.SeedGraveyardCard(t, g, "Filler Bolt", "Instant", "test-stingcaster-bolt")
	effectstest.RfCrACast(t, g, "Stingcaster Mage", "Creature — Human Wizard", effectstest.OracleStingcaster, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	perm := effectstest.GrantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || perm.Cost != "" {
		t.Fatalf("granted permission %+v, want flashback at its mana cost", perm)
	}
}

func TestStingcasterMageRefusesACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.SeedGraveyardCard(t, g, "Gy Bear", "Creature — Bear", "test-stingcaster-bear")
	effectstest.RfCrACast(t, g, "Stingcaster Mage", "Creature — Human Wizard", effectstest.OracleStingcaster, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Fatal("a creature card was accepted as the target")
		}
	}
	if effectstest.GrantedPermissionOn(g, me.ID, bear, game.ZoneGraveyard).Granted() {
		t.Fatal("a creature card was granted flashback")
	}
}

// --- Sphinx of False Conclusions -----------------------------------

func TestSphinxOfFalseConclusionsLeavesOneTokenCopyWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.RfCrACast(t, g, "Sphinx of False Conclusions", "Creature — Sphinx Illusion", effectstest.OracleSphinxFalse, 4, 2)
	effectstest.PassPriorityAroundTable(t, g)
	rfCrADestroy(t, g, id)
	effectstest.PassPriorityAroundTable(t, g)
	var tok uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Sphinx of False Conclusions" && c.Controller == me.ID {
			tok = c.InstanceID
			if !effects.IsToken(c) {
				t.Fatal("the copy is not a token")
			}
		}
	}
	if tok == uuid.Nil {
		t.Fatal("no copy appeared")
	}
	rfCrADestroy(t, g, tok)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Sphinx of False Conclusions"); n != 0 {
		t.Fatalf("%d Sphinxes after the token died, want none", n)
	}
}

func TestSphinxOfFalseConclusionsLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	id := effectstest.RfCrAPush(g, me.ID, "Sphinx of False Conclusions", "Creature — Sphinx Illusion", effectstest.OracleSphinxFalse, 4, 2)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, foe.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Fatalf("no discard owed after the draw")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d after looting, want %d", me.Hand.Size(), hand)
	}
}

// --- Sureshot Sower ------------------------------------------------

func TestSureshotSowerDiscardsItselfToDestroyAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	effectstest.ToMain(t, g)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: foe.ID, Controller: foe.ID,
	})
	ground := effectstest.RfCrAPush(g, foe.ID, "Bear", "Creature — Bear", "", 2, 2)
	sower := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: sower, Name: "Sureshot Sower", TypeLine: "Creature — Human Archer",
		OracleID: effectstest.OracleSureshot, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.RfCrAAddMana(t, g, me, "{G}{G}{G}{G}")
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
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RfCrACast(t, g, "Tamiyo, Upriser Crowned", "Legendary Creature — Moonfolk Warrior", effectstest.OracleTamiyoUpriser, 3, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.YoureTheMonarch(g, me.ID) {
		t.Fatal("not the monarch after Tamiyo entered")
	}
}

func TestTamiyoUpriserCrownedStunsCreaturesThatHitTheMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	effectstest.RfCrAPush(g, me.ID, "Tamiyo, Upriser Crowned", "Legendary Creature — Moonfolk Warrior", effectstest.OracleTamiyoUpriser, 3, 5)
	g.WithWriteLock(func() { _ = g.SetMonarchForEffect(me.ID) })
	hitter := effectstest.RfCrAPush(g, foe.ID, "Hitter", "Creature — Bear", "", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, hitter, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, hitter)
	if !ok || !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Fatalf("hitter %+v (ok %v), want tapped with one stun counter", c, ok)
	}
}

func TestTamiyoUpriserCrownedLeavesCreaturesAloneWhenYouAreNotTheMonarchOrTheDamageIsNoncombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	effectstest.RfCrAPush(g, me.ID, "Tamiyo, Upriser Crowned", "Legendary Creature — Moonfolk Warrior", effectstest.OracleTamiyoUpriser, 3, 5)
	hitter := effectstest.RfCrAPush(g, foe.ID, "Hitter", "Creature — Bear", "", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, hitter, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, hitter); c.Tapped || c.Counters[game.CounterStun] != 0 {
		t.Fatal("stunned although Tamiyo's controller is not the monarch")
	}
	g.WithWriteLock(func() { _ = g.SetMonarchForEffect(me.ID) })
	effectstest.EmitNoncombatDamage(g, hitter, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, hitter); c.Tapped || c.Counters[game.CounterStun] != 0 {
		t.Fatal("stunned by noncombat damage")
	}
}

// --- Tenured Tethermage --------------------------------------------

func TestTenuredTethermageSacrificesALandForTwoTappedHeartwoods(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.RfCrAPush(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.RfCrACast(t, g, "Tenured Tethermage", "Creature — Human Artificer", effectstest.OracleTethermage, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Fatal("the land was not sacrificed")
	}
	if n := effectstest.B16CountNamed(g, "Heartwood"); n != 2 {
		t.Fatalf("%d Heartwood tokens, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Heartwood" && !c.Tapped {
			t.Fatal("a Heartwood entered untapped")
		}
	}
}

func TestTenuredTethermageMakesNothingWhenYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.RfCrAPush(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.RfCrACast(t, g, "Tenured Tethermage", "Creature — Human Artificer", effectstest.OracleTethermage, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || effectstest.B16CountNamed(g, "Heartwood") != 0 {
		t.Fatal("declining still cost a land or made tokens")
	}
}

func TestTenuredTethermageTapsTwoArtifactsForTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	mage := effectstest.RfCrAPush(g, me.ID, "Tenured Tethermage", "Creature — Human Artificer", effectstest.OracleTethermage, 1, 1)
	a := effectstest.RfCrAPush(g, me.ID, "Rock", "Artifact", "", 0, 0)
	b := effectstest.RfCrAPush(g, me.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.B16Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfCrACounters(g, mage, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d counters, want 2", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := effectstest.BattlefieldCard(g, id); !c.Tapped {
			t.Fatal("an artifact was not tapped to pay")
		}
	}
}

// --- Tether Technician ---------------------------------------------

func TestTetherTechnicianDiscardsACardToDealTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	effectstest.RfCrACast(t, g, "Tether Technician", "Creature — Minotaur Artificer", effectstest.OracleTetherTech, 4, 5)
	effectstest.PassPriorityAroundTable(t, g)
	life := foe.Life
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayerTarget(t, g, me.ID, foe.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if foe.Life != life-2 {
		t.Fatalf("life %d, want %d", foe.Life, life-2)
	}
}

func TestTetherTechnicianDiscardingNothingDealsNoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	effectstest.RfCrACast(t, g, "Tether Technician", "Creature — Minotaur Artificer", effectstest.OracleTetherTech, 4, 5)
	effectstest.PassPriorityAroundTable(t, g)
	life := foe.Life
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || foe.Life != life {
		t.Fatal("a target was asked for or damage dealt without a discard")
	}
}

// --- Tetsuko Umezawa, Pursuer --------------------------------------

func TestTetsukoPingsTheControllerOfAWeakBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	effectstest.RfCrAPush(g, me.ID, "Tetsuko Umezawa, Pursuer", "Legendary Creature — Human Mercenary", effectstest.OracleTetsukoPursuer, 2, 4)
	weak := effectstest.RfCrAPush(g, foe.ID, "Weak", "Creature — Bear", "", 1, 3)
	frail := effectstest.RfCrAPush(g, foe.ID, "Frail", "Creature — Bear", "", 3, 1)
	big := effectstest.RfCrAPush(g, foe.ID, "Big", "Creature — Bear", "", 3, 3)
	attacker := effectstest.RfCrAPush(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2)
	life := foe.Life
	block := func(blocker uuid.UUID) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: foe.ID, CardID: blocker, Target: attacker, Amount: 1})
		})
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	effectstest.RfCrAPush(g, me.ID, "Tetsuko Umezawa, Pursuer", "Legendary Creature — Human Mercenary", effectstest.OracleTetsukoPursuer, 2, 4)
	mine := effectstest.RfCrAPush(g, me.ID, "Weak", "Creature — Bear", "", 1, 1)
	attacker := effectstest.RfCrAPush(g, foe.ID, "Attacker", "Creature — Bear", "", 2, 2)
	life := me.Life
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: mine, Target: attacker, Amount: 1})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// --- Theoretical Necromancer ---------------------------------------

func TestTheoreticalNecromancerExilesItselfToReturnAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	necro := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: necro, Name: "Theoretical Necromancer", TypeLine: "Creature — Vampire Warlock",
		OracleID: effectstest.OracleTheoreticalNecr, Power: 4, Toughness: 1, Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Gy Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.RfCrAAddMana(t, g, me, "{B}{B}{B}{B}")
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
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Fatal("the other creature did not return to hand")
	}
	if !g.Exile.Contains(necro) {
		t.Fatal("the Necromancer was not exiled")
	}
}

// --- Tinybones, Pocket Nuisance ------------------------------------

func TestTinybonesPocketNuisanceTriggersOncePerDiscardingPlayerPerBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RfCrACast(t, g, "Tinybones, Pocket Nuisance", "Legendary Creature — Skeleton Rogue", effectstest.OracleTinybonesPocket, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
		if p.ID != me.ID {
			effectstest.DiscardFromHand(t, g, p.ID)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	foe := effectstest.RfCrEFoe(g)
	before := foe.Life
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if before-foe.Life != 1 {
		t.Fatalf("a two-card discard cost the opponent %d life, want 1", before-foe.Life)
	}
}

// --- Titanbones, Towering Heart ------------------------------------

func TestTitanbonesGrowsByTwoWheneverYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.RfCrAPush(g, me.ID, "Titanbones, Towering Heart", "Legendary Creature — Skeleton Druid", effectstest.OracleTitanbones, 4, 3)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := rfCrACounters(g, id, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d counters, want 2", got)
	}
}

func TestTitanbonesGainsThreeLifeWhenDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Titanbones, Towering Heart", TypeLine: "Legendary Creature — Skeleton Druid",
		OracleID: effectstest.OracleTitanbones, Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID})
	me.Hand.Cards = me.Hand.Cards[len(me.Hand.Cards)-1:]
	life := me.Life
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Fatalf("life %d, want %d", me.Life, life+3)
	}
}

// --- Tomik, Izzet Sparkmage ----------------------------------------

func TestTomikAddsOneToNoncombatDamageAgainstOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	tomik := effectstest.RfCrAPush(g, me.ID, "Tomik, Izzet Sparkmage", "Legendary Creature — Human Wizard", effectstest.OracleTomikSparkmage, 1, 2)
	src := effectstest.RfCrAPush(g, me.ID, "Pinger", "Creature — Bear", "", 1, 1)
	bear := effectstest.RfCrAPush(g, foe.ID, "Their Bear", "Creature — Bear", "", 5, 5)
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
	theirSrc := effectstest.RfCrAPush(g, foe.ID, "Their Pinger", "Creature — Bear", "", 1, 1)
	pre := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirSrc, me.ID, 2) })
	if me.Life != pre-2 {
		t.Fatalf("an opposing source dealt %d, want 2", pre-me.Life)
	}
	_ = tomik
}

func TestTomikDoesNotBoostCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	effectstest.RfCrAPush(g, me.ID, "Tomik, Izzet Sparkmage", "Legendary Creature — Human Wizard", effectstest.OracleTomikSparkmage, 1, 2)
	atk := effectstest.RfCrAPush(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2)
	life := foe.Life
	effectstest.DeclareAttack(t, g, foe.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if foe.Life != life-2 {
		t.Fatalf("combat damage took %d life, want 2", life-foe.Life)
	}
}

// --- Seasoned Cryomancer -------------------------------------------

func TestSeasonedCryomancerLootsTwoThenStunsOnePerNonlandDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	me.Hand.Cards = nil
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	victim := effectstest.RfCrAPush(g, foe.ID, "Victim", "Creature — Bear", "", 2, 2)
	other := effectstest.RfCrAPush(g, foe.ID, "Other", "Creature — Bear", "", 2, 2)
	effectstest.RfCrACast(t, g, "Seasoned Cryomancer", "Creature — Human Wizard", effectstest.OracleCryomancer, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.AnswerDiscard(t, g, me.ID, land, spellInHand)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	v, _ := effectstest.BattlefieldCard(g, victim)
	o, _ := effectstest.BattlefieldCard(g, other)
	if !v.Tapped || v.Counters[game.CounterStun] != 1 {
		t.Fatalf("the chosen creature: tapped %v, stun %d", v.Tapped, v.Counters[game.CounterStun])
	}
	if o.Tapped || o.Counters[game.CounterStun] != 0 {
		t.Fatal("a creature that was not chosen was affected")
	}
}

func TestSeasonedCryomancerDiscardingOnlyLandsMakesNoTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	me.Hand.Cards = nil
	for i := 0; i < 2; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	}
	victim := effectstest.RfCrAPush(g, foe.ID, "Victim", "Creature — Bear", "", 2, 2)
	effectstest.RfCrACast(t, g, "Seasoned Cryomancer", "Creature — Human Wizard", effectstest.OracleCryomancer, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target prompt opened after discarding only lands")
	}
	if v, _ := effectstest.BattlefieldCard(g, victim); v.Tapped {
		t.Fatal("a creature was tapped")
	}
}

func TestSeasonedCryomancerDrawsTwoFromTheGraveyardAtInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Seasoned Cryomancer", TypeLine: "Creature — Human Wizard",
		OracleID: effectstest.OracleCryomancer, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.RfCrAAddMana(t, g, me, "{U}{U}{U}{U}{U}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate from graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || !g.Exile.Contains(id) {
		t.Fatalf("hand %d (want %d), exiled %v", me.Hand.Size(), hand+2, g.Exile.Contains(id))
	}
}

// Real combat: two creatures hit the monarch in the same damage step. The
// CR 725.2 monarch steal runs as the first one connects, but the damage is
// simultaneous, so both must be stunned.
func TestTamiyoUpriserCrownedStunsEveryCreatureOfASimultaneousHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	atkr, def := g.Seats[g.Turn.ActiveSeat], effectstest.RfCrEFoe(g)
	effectstest.RfCrAPush(g, def.ID, "Tamiyo, Upriser Crowned", "Legendary Creature — Moonfolk Warrior", effectstest.OracleTamiyoUpriser, 3, 5)
	g.WithWriteLock(func() { _ = g.SetMonarchForEffect(def.ID) })
	a := effectstest.RfCrAPush(g, atkr.ID, "A", "Creature — Bear", "", 2, 2)
	b := effectstest.RfCrAPush(g, atkr.ID, "B", "Creature — Bear", "", 2, 2)
	effectstest.AttackWith(t, g, def.ID, a, b)
	effectstest.AnnSettle(t, g, nil)
	for _, id := range []uuid.UUID{a, b} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || c.Counters[game.CounterStun] != 1 {
			t.Fatalf("creature %s: ok %v, stun %d, want 1", id, ok, c.Counters[game.CounterStun])
		}
	}
}
