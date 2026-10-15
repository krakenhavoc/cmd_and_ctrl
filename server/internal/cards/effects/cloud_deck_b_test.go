package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cloud_deck_b_test.go — the second slice of the Cloud, Ex-SOLDIER deck
// request (#3656): Summon: Kujata, Ultimate Magic: Meteor, Lifestream's
// Blessing, Heidegger, Cid, Bronze Guardian, Unfinished Business,
// Aerith, Yuffie and Cait Sith.

const (
	summonKujataOracle      = "9fe77f70-85cf-4a18-9451-2fad4c341f07"
	ultimateMeteorOracle    = "31cc5d9d-9384-4a9b-ba7f-6a4edbcb7a35"
	lifestreamsBlessingOrcl = "7cbdba72-ed33-40c6-a9c0-4d65c280a4d0"
	heideggerOracle         = "215024ba-811e-4011-bbc8-8dd14a34f973"
	cidFreeflierOracle      = "d3bdfb25-dfb2-4df4-b502-ef0909a39f86"
	bronzeGuardianOracle    = "922a030c-857b-4c61-93df-b70ebb32abd3"
	unfinishedBusinessOracl = "800cfd70-6a97-40eb-8dc0-40b762e4db76"
	aerithLastAncientOracle = "d6cba688-faff-4d9f-8148-37d389fa5fb3"
	yuffieMateriaOracle     = "acdcbb49-9d71-4c36-8ccb-650fb11c3a58"
	caitSithOracle          = "31851cb6-0caf-4575-aa5e-41aeadccafc1"
)

func cloudDeckPickTargets(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	pick := latestPickTarget(g, chooser)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	var refs []game.TargetRef
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

func cloudDeckKujata(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return tdCastFromHand(t, g, game.Card{Name: "Summon: Kujata", TypeLine: "Enchantment Creature — Saga Ox",
		OracleID: summonKujataOracle, ManaCost: "{5}{R}", Power: 7, Toughness: 5}, game.CastSpellParams{})
}

func TestCloudDeckKujataChapterIDealsThreeToTwoCreatures(t *testing.T) {
	cloudDeckWantFull(t, summonKujataOracle)
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	a := tdCreature(g, opp.ID, "A", 3, 3)
	b := tdCreature(g, opp.ID, "B", 2, 3)
	c := tdCreature(g, opp.ID, "C", 4, 4)

	cloudDeckKujata(t, g)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	cloudDeckPickTargets(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("chapter I did not kill both 3-toughness creatures")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("the unchosen creature died")
	}
}

func TestCloudDeckKujataChapterIIIRummagesAndDamagesEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	kujata := cloudDeckKujata(t, g)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	cloudDeckPickTargets(t, g, me.ID, kujata)
	passPriorityAroundTable(t, g)

	advanceToPrecombatMainOf(t, g, seat) // chapter II
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	cloudDeckPickTargets(t, g, me.ID, kujata)
	passPriorityAroundTable(t, g)

	fodder := me.Hand.PushTop
	id := uuid.New()
	fodder(game.Card{InstanceID: id, Name: "Fodder", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	before := me.Hand.Size()

	advanceToPrecombatMainOf(t, g, seat) // chapter III
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, id)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("hand grew by %d, want 2 (the turn's draw, discard one, draw two)", got)
	}
	if got := lives[opp.ID] - opp.Life; got != 3 {
		t.Errorf("opponent lost %d, want 3 (the discarded card's mana value)", got)
	}
	if got := lives[me.ID] - me.Life; got != 0 {
		t.Errorf("the controller lost %d life", got)
	}
}

func cloudDeckForetellAndCast(t *testing.T, g *game.Game, name, typeLine, oracle, foretellCost string, pay []game.ManaToken, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	card := handCardForTest(active, name, typeLine, oracle)
	active.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(active.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell %s: %v", name, err)
	}
	g.WithWriteLock(func() { g.Turn.Seq++ })
	active.ManaPool.AddMana(pay...)
	if err := g.CastSpell(active.ID, card, game.CastSpellParams{
		Strict:          true,
		FromZone:        "exile",
		AlternativeCost: game.AltCostKeyForetell,
		Targets:         targets,
	}); err != nil {
		t.Fatalf("cast the foretold %s: %v", name, err)
	}
}

func cloudDeckMana(color string, n int) []game.ManaToken {
	out := make([]game.ManaToken, n)
	for i := range out {
		out[i] = game.ManaToken{Color: color}
	}
	return out
}

func TestCloudDeckMeteorForetoldAlsoDestroysAnArtifactOrLandPerOpponent(t *testing.T) {
	cloudDeckWantFull(t, ultimateMeteorOracle)
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Test", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Test", 3, 3)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	other := b12Permanent(g, opp.ID, "Other Rock", "Artifact")

	pay := append(cloudDeckMana("C", 5), game.ManaToken{Color: "R"})
	cloudDeckForetellAndCast(t, g, "Ultimate Magic: Meteor", "Sorcery", ultimateMeteorOracle, "{5}{R}", pay, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt for the artifact or land")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("7 damage to each creature left a creature alive")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the chosen artifact survived")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("an unchosen artifact was destroyed")
	}
}

func TestCloudDeckMeteorFromHandOnlyDamagesCreatures(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Test", 3, 3)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	advanceTo(t, g, game.StepPrecombatMain)
	card := handCardForTest(me, "Ultimate Magic: Meteor", "Sorcery", ultimateMeteorOracle)
	me.ManaPool.AddMana(append(cloudDeckMana("C", 5), game.ManaToken{Color: "R"})...)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the creature survived 7 damage")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed without the exile cast")
	}
}

func TestCloudDeckLifestreamsBlessingDrawsGreatestPowerAndGainsDoubleFromExile(t *testing.T) {
	spec, ok := Lookup(lifestreamsBlessingOrcl)
	if !ok || spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Fatalf("Lifestream's Blessing registered %v, %s, %d caveats; want one caveat", ok, spec.Completeness, len(spec.Caveats))
	}
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Big", "Creature — Test", 4, 4)
	b12Creature(g, me.ID, "Small", "Creature — Test", 1, 1)
	life, hand := me.Life, me.Hand.Size()

	pay := append(cloudDeckMana("C", 4), game.ManaToken{Color: "G"})
	cloudDeckForetellAndCast(t, g, "Lifestream's Blessing", "Instant", lifestreamsBlessingOrcl, "{4}{G}", pay, nil)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4 (the greatest power)", got)
	}
	if got := me.Life - life; got != 8 {
		t.Errorf("gained %d, want 8 (twice 4, cast from exile)", got)
	}
}

func TestCloudDeckLifestreamsBlessingFromHandGainsNoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Big", "Creature — Test", 3, 3)
	life, hand := me.Life, me.Hand.Size()
	advanceTo(t, g, game.StepPrecombatMain)
	card := handCardForTest(me, "Lifestream's Blessing", "Instant", lifestreamsBlessingOrcl)
	me.ManaPool.AddMana(append(cloudDeckMana("C", 4), cloudDeckMana("G", 2)...)...)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand grew by %d, want 3 (three drawn; the cast card is added and leaves again)", got)
	}
	if me.Life != life {
		t.Errorf("gained %d life from a cast out of hand", me.Life-life)
	}
}

func TestCloudDeckHeideggerBoostsByTheSoldierCountAndMakesSoldiers(t *testing.T) {
	cloudDeckWantFull(t, heideggerOracle)
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	advanceToMain(t, g)
	heidegger := pushCatalogPermanent(g, me.ID, "Heidegger, Shinra Executive", "Legendary Creature — Human Soldier", heideggerOracle, false)
	b12Creature(g, me.ID, "Trooper", "Creature — Human Soldier", 1, 1)
	bear := seedBear(g, me.ID)
	for i := 0; i < 4; i++ {
		b12Creature(g, opp.ID, "Theirs", "Creature — Test", 1, 1)
	}

	advanceTo(t, g, game.StepBeginCombat)
	cloudDeckPickTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("Bear power %d, want 4 (+2 for two Soldiers)", got)
	}

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("Soldier tokens = %d, want 1 (one opponent has more creatures)", n)
	}
	_ = heidegger
}

func TestCloudDeckCidReducesEquipmentAndHasJumpOnYourTurn(t *testing.T) {
	cloudDeckWantFull(t, cidFreeflierOracle)
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	advanceToMain(t, g)
	cid := pushCatalogPermanent(g, me.ID, "Cid, Freeflier Pilot", "Legendary Creature — Human Warrior Pilot", cidFreeflierOracle, false)
	if !cloudDeckHasKeyword(t, g, cid, "flying") {
		t.Error("Cid lacks flying on his controller's turn")
	}
	card := handCardForTest(me, "Test Blade", equipTypeLine, "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == card {
				me.Hand.Cards[i].ManaCost = "{2}"
			}
		}
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("an Equipment costing {2} was not cheaper by {1}: %v", err)
	}

	for g.Turn.ActiveSeat == seat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if cloudDeckHasKeyword(t, g, cid, "flying") {
		t.Error("Cid has flying on an opponent's turn")
	}
}

func TestCloudDeckBronzeGuardianCountsArtifactsAndWardsOtherArtifacts(t *testing.T) {
	cloudDeckWantFull(t, bronzeGuardianOracle)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	guardian := pushCatalogPermanent(g, me.ID, "Bronze Guardian", "Artifact Creature — Golem", bronzeGuardianOracle, false)
	if got := effectivePower(t, g, guardian); got != 1 {
		t.Errorf("power %d, want 1 (itself)", got)
	}
	b12Permanent(g, me.ID, "Rock", "Artifact")
	b12Permanent(g, me.ID, "Other Rock", "Artifact")
	if got := effectivePower(t, g, guardian); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	spec, _ := Lookup(bronzeGuardianOracle)
	if len(spec.Triggered) != 2 {
		t.Errorf("Bronze Guardian has %d ward triggers, want its own and the grant", len(spec.Triggered))
	}
}

func TestCloudDeckUnfinishedBusinessReturnsACreatureWithItsEquipment(t *testing.T) {
	cloudDeckWantFull(t, unfinishedBusinessOracl)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature := pushGraveyardPermanent(me, "Fallen", "Creature — Test", "{2}")
	blade := pushGraveyardPermanent(me, "Test Blade", equipTypeLine, "{2}")

	castCatalogSpell(t, g, "Unfinished Business", "Sorcery", unfinishedBusinessOracl, []game.TargetRef{
		{Kind: game.TargetCard, ID: creature, Slot: 0},
		{Kind: game.TargetCard, ID: blade, Slot: 1},
	})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(creature) || !g.Battlefield.Contains(blade) {
		t.Fatalf("creature on battlefield %v, Equipment %v", g.Battlefield.Contains(creature), g.Battlefield.Contains(blade))
	}
	if host := attachmentHostOf(t, g, blade); host.ID != creature {
		t.Errorf("Equipment host = %+v, want the returned creature", host)
	}
}

func TestCloudDeckAerithRaiseReturnsToHandOrBattlefieldBySevenLife(t *testing.T) {
	cloudDeckWantFull(t, aerithLastAncientOracle)
	for _, tc := range []struct {
		life        int
		wantInHand  bool
		wantOnField bool
	}{{3, true, false}, {7, false, true}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		pushCatalogPermanent(g, me.ID, "Aerith, Last Ancient", "Legendary Creature — Human Cleric Druid", aerithLastAncientOracle, false)
		dead := pushGraveyardPermanent(me, "Fallen", "Creature — Test", "{2}")
		g.WithWriteLock(func() {
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, tc.life); err != nil {
				t.Fatalf("gain life: %v", err)
			}
		})

		advanceToEndStep(t, g)
		if pick := latestPickTarget(g, me.ID); pick != nil {
			pickCard(t, g, me.ID, dead)
		}
		passPriorityAroundTable(t, g)
		if got := me.Hand.Contains(dead); got != tc.wantInHand {
			t.Errorf("life %d: in hand = %v, want %v", tc.life, got, tc.wantInHand)
		}
		if got := g.Battlefield.Contains(dead); got != tc.wantOnField {
			t.Errorf("life %d: on battlefield = %v, want %v", tc.life, got, tc.wantOnField)
		}
	}
}

func TestCloudDeckAerithDoesNothingWithoutLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Aerith, Last Ancient", "Legendary Creature — Human Cleric Druid", aerithLastAncientOracle, false)
	dead := pushGraveyardPermanent(me, "Fallen", "Creature — Test", "{2}")

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(dead) || g.Battlefield.Contains(dead) {
		t.Error("Raise triggered without life gained")
	}
}

func TestCloudDeckYuffieStealsAnArtifactAndAttachesAnEquipment(t *testing.T) {
	cloudDeckWantFull(t, yuffieMateriaOracle)
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	blade := seedEquipment(g, me.ID, "Test Blade", "")

	yuffie := castAndResolveCreature(t, g, "Yuffie, Materia Hunter", "Legendary Creature — Human Ninja", yuffieMateriaOracle)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickCard(t, g, me.ID, rock)
	}
	passPriorityAroundTable(t, g)
	if card, ok := lookupBattlefield(g, rock); !ok || card.Controller != me.ID {
		t.Errorf("the artifact is not under Yuffie's controller: %+v %v", card.Controller, ok)
	}
	if prompt := latestChooseCardsFor(g, me.ID); prompt != nil {
		if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{blade}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	} else {
		t.Fatal("no attach offer")
	}
	if host := attachmentHostOf(t, g, blade); host.ID != yuffie {
		t.Errorf("Equipment host = %+v, want Yuffie", host)
	}
}

func lookupBattlefield(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, ok = c, true
			}
		}
	})
	return out, ok
}

func TestCloudDeckCaitSithExilesTheTopCardAndBoostsByItsManaValue(t *testing.T) {
	cloudDeckWantFull(t, caitSithOracle)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Cait Sith, Fortune Teller", "Legendary Artifact Creature — Cat Moogle", caitSithOracle, false)
	bear := seedBear(g, me.ID)
	top := seedLibraryTop(me, "Top Spell", "Sorcery", "{3}{R}")

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if scry := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID); scry != nil {
		if err := g.ResolveScry(scry.ID, me.ID, nil, []uuid.UUID{top}); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
	}
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(top) {
		t.Error("the scried top card was not exiled")
	}
	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("Bear power %d, want 6 (+4 for mana value 4)", got)
	}
}
