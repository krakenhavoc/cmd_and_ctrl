package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCloudDeckKujataChapterIDealsThreeToTwoCreatures(t *testing.T) {
	cloudDeckWantFull(t, effectstest.SummonKujataOracle)
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	a := effectstest.TdCreature(g, opp.ID, "A", 3, 3)
	b := effectstest.TdCreature(g, opp.ID, "B", 2, 3)
	c := effectstest.TdCreature(g, opp.ID, "C", 4, 4)

	effectstest.CloudDeckKujata(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.CloudDeckPickTargets(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("chapter I did not kill both 3-toughness creatures")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("the unchosen creature died")
	}
}

func TestCloudDeckKujataChapterIIIRummagesAndDamagesEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	kujata := effectstest.CloudDeckKujata(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.CloudDeckPickTargets(t, g, me.ID, kujata)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // chapter II
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.CloudDeckPickTargets(t, g, me.ID, kujata)
	effectstest.PassPriorityAroundTable(t, g)

	fodder := me.Hand.PushTop
	id := uuid.New()
	fodder(game.Card{InstanceID: id, Name: "Fodder", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	before := me.Hand.Size()

	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // chapter III
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)

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

func TestCloudDeckMeteorForetoldAlsoDestroysAnArtifactOrLandPerOpponent(t *testing.T) {
	cloudDeckWantFull(t, effectstest.UltimateMeteorOracle)
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Test", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Test", 3, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	other := effectstest.B12Permanent(g, opp.ID, "Other Rock", "Artifact")

	pay := append(effectstest.CloudDeckMana("C", 5), game.ManaToken{Color: "R"})
	effectstest.CloudDeckForetellAndCast(t, g, "Ultimate Magic: Meteor", "Sorcery", effectstest.UltimateMeteorOracle, "{5}{R}", pay, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt for the artifact or land")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Test", 3, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	card := effectstest.HandCardForTest(me, "Ultimate Magic: Meteor", "Sorcery", effectstest.UltimateMeteorOracle)
	me.ManaPool.AddMana(append(effectstest.CloudDeckMana("C", 5), game.ManaToken{Color: "R"})...)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the creature survived 7 damage")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed without the exile cast")
	}
}

func TestCloudDeckLifestreamsBlessingDrawsGreatestPowerAndGainsDoubleFromExile(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.LifestreamsBlessingOrcl)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || len(spec.CastTimeValues) != 1 {
		t.Fatalf("Lifestream's Blessing registered %v, %s, %d caveats, %d cast-time values; want Full with one cast-time value",
			ok, spec.Completeness, len(spec.Caveats), len(spec.CastTimeValues))
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Test", 4, 4)
	effectstest.B12Creature(g, me.ID, "Small", "Creature — Test", 1, 1)
	life, hand := me.Life, me.Hand.Size()

	pay := append(effectstest.CloudDeckMana("C", 4), game.ManaToken{Color: "G"})
	effectstest.CloudDeckForetellAndCast(t, g, "Lifestream's Blessing", "Instant", effectstest.LifestreamsBlessingOrcl, "{4}{G}", pay, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4 (the greatest power)", got)
	}
	if got := me.Life - life; got != 8 {
		t.Errorf("gained %d, want 8 (twice 4, cast from exile)", got)
	}
}

// #3781, CR 601.2i: X is fixed as the spell is cast. What happens to the
// board while the spell sits on the stack does not change it.
func TestCloudDeckLifestreamsBlessingXIsFixedAsYouCast(t *testing.T) {
	for _, tc := range []struct {
		name     string
		respond  func(t *testing.T, g *game.Game, me *game.Player, big uuid.UUID)
		wantDraw int
	}{
		{"the biggest creature dies in response", func(t *testing.T, g *game.Game, _ *game.Player, big uuid.UUID) {
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(big); err != nil {
					t.Fatalf("destroy: %v", err)
				}
			})
		}, 4},
		{"a bigger creature arrives in response", func(t *testing.T, g *game.Game, me *game.Player, _ uuid.UUID) {
			effectstest.B12Creature(g, me.ID, "Late", "Creature — Test", 7, 7)
		}, 4},
		{"the biggest creature grows in response", func(t *testing.T, g *game.Game, _ *game.Player, big uuid.UUID) {
			if err := g.AddCounter(big, "+1/+1", 3); err != nil {
				t.Fatalf("counter: %v", err)
			}
		}, 4},
		{"nothing happens", func(*testing.T, *game.Game, *game.Player, uuid.UUID) {}, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Test", 4, 4)
			effectstest.B12Creature(g, me.ID, "Small", "Creature — Test", 1, 1)
			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			card := effectstest.HandCardForTest(me, "Lifestream's Blessing", "Instant", effectstest.LifestreamsBlessingOrcl)
			hand := me.Hand.Size() - 1 // the cast card leaves the hand
			me.ManaPool.AddMana(append(effectstest.CloudDeckMana("C", 4), effectstest.CloudDeckMana("G", 2)...)...)
			if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			var recorded int
			var ok bool
			g.ReadSnapshot(func() { recorded, ok = g.StackItemForEffect(card).CastValue(cardskr.LifestreamsBlessingX) })
			if !ok || recorded != 4 {
				t.Fatalf("cast-time value = %d, %v; want 4 recorded as the spell was cast", recorded, ok)
			}
			tc.respond(t, g, me, big)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.wantDraw {
				t.Errorf("drew %d, want %d (the greatest power as it was cast)", got, tc.wantDraw)
			}
		})
	}
}

// No creature at cast time: X is zero even if one arrives in response.
func TestCloudDeckLifestreamsBlessingWithNoCreatureDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	card := effectstest.HandCardForTest(me, "Lifestream's Blessing", "Instant", effectstest.LifestreamsBlessingOrcl)
	hand := me.Hand.Size() - 1
	me.ManaPool.AddMana(append(effectstest.CloudDeckMana("C", 4), effectstest.CloudDeckMana("G", 2)...)...)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.B12Creature(g, me.ID, "Late", "Creature — Test", 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d, want 0 (no creature as it was cast)", got)
	}
}

func TestCloudDeckLifestreamsBlessingFromHandGainsNoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Test", 3, 3)
	life, hand := me.Life, me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	card := effectstest.HandCardForTest(me, "Lifestream's Blessing", "Instant", effectstest.LifestreamsBlessingOrcl)
	me.ManaPool.AddMana(append(effectstest.CloudDeckMana("C", 4), effectstest.CloudDeckMana("G", 2)...)...)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand grew by %d, want 3 (three drawn; the cast card is added and leaves again)", got)
	}
	if me.Life != life {
		t.Errorf("gained %d life from a cast out of hand", me.Life-life)
	}
}

func TestCloudDeckHeideggerBoostsByTheSoldierCountAndMakesSoldiers(t *testing.T) {
	cloudDeckWantFull(t, effectstest.HeideggerOracle)
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	heidegger := effectstest.PushCatalogPermanent(g, me.ID, "Heidegger, Shinra Executive", "Legendary Creature — Human Soldier", effectstest.HeideggerOracle, false)
	effectstest.B12Creature(g, me.ID, "Trooper", "Creature — Human Soldier", 1, 1)
	bear := effectstest.SeedBear(g, me.ID)
	for i := 0; i < 4; i++ {
		effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Test", 1, 1)
	}

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.CloudDeckPickTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("Bear power %d, want 4 (+2 for two Soldiers)", got)
	}

	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("Soldier tokens = %d, want 1 (one opponent has more creatures)", n)
	}
	_ = heidegger
}

func TestCloudDeckCidReducesEquipmentAndHasJumpOnYourTurn(t *testing.T) {
	cloudDeckWantFull(t, effectstest.CidFreeflierOracle)
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMain(t, g)
	cid := effectstest.PushCatalogPermanent(g, me.ID, "Cid, Freeflier Pilot", "Legendary Creature — Human Warrior Pilot", effectstest.CidFreeflierOracle, false)
	if !cloudDeckHasKeyword(t, g, cid, "flying") {
		t.Error("Cid lacks flying on his controller's turn")
	}
	card := effectstest.HandCardForTest(me, "Test Blade", effectstest.EquipTypeLine, "")
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
	cloudDeckWantFull(t, effectstest.BronzeGuardianOracle)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	guardian := effectstest.PushCatalogPermanent(g, me.ID, "Bronze Guardian", "Artifact Creature — Golem", effectstest.BronzeGuardianOracle, false)
	if got := effectstest.EffectivePower(t, g, guardian); got != 1 {
		t.Errorf("power %d, want 1 (itself)", got)
	}
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Other Rock", "Artifact")
	if got := effectstest.EffectivePower(t, g, guardian); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	spec, _ := effects.Lookup(effectstest.BronzeGuardianOracle)
	if len(spec.Triggered) != 2 {
		t.Errorf("Bronze Guardian has %d ward triggers, want its own and the grant", len(spec.Triggered))
	}
}

func TestCloudDeckUnfinishedBusinessReturnsACreatureWithItsEquipment(t *testing.T) {
	cloudDeckWantFull(t, effectstest.UnfinishedBusinessOracl)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature := effectstest.PushGraveyardPermanent(me, "Fallen", "Creature — Test", "{2}")
	blade := effectstest.PushGraveyardPermanent(me, "Test Blade", effectstest.EquipTypeLine, "{2}")

	effectstest.CastCatalogSpell(t, g, "Unfinished Business", "Sorcery", effectstest.UnfinishedBusinessOracl, []game.TargetRef{
		{Kind: game.TargetCard, ID: creature, Slot: 0},
		{Kind: game.TargetCard, ID: blade, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(creature) || !g.Battlefield.Contains(blade) {
		t.Fatalf("creature on battlefield %v, Equipment %v", g.Battlefield.Contains(creature), g.Battlefield.Contains(blade))
	}
	if host := effectstest.AttachmentHostOf(t, g, blade); host.ID != creature {
		t.Errorf("Equipment host = %+v, want the returned creature", host)
	}
}

func TestCloudDeckAerithRaiseReturnsToHandOrBattlefieldBySevenLife(t *testing.T) {
	cloudDeckWantFull(t, effectstest.AerithLastAncientOracle)
	for _, tc := range []struct {
		life        int
		wantInHand  bool
		wantOnField bool
	}{{3, true, false}, {7, false, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		effectstest.PushCatalogPermanent(g, me.ID, "Aerith, Last Ancient", "Legendary Creature — Human Cleric Druid", effectstest.AerithLastAncientOracle, false)
		dead := effectstest.PushGraveyardPermanent(me, "Fallen", "Creature — Test", "{2}")
		g.WithWriteLock(func() {
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, tc.life); err != nil {
				t.Fatalf("gain life: %v", err)
			}
		})

		effectstest.AdvanceToEndStep(t, g)
		if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
			effectstest.PickCard(t, g, me.ID, dead)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Contains(dead); got != tc.wantInHand {
			t.Errorf("life %d: in hand = %v, want %v", tc.life, got, tc.wantInHand)
		}
		if got := g.Battlefield.Contains(dead); got != tc.wantOnField {
			t.Errorf("life %d: on battlefield = %v, want %v", tc.life, got, tc.wantOnField)
		}
	}
}

func TestCloudDeckAerithDoesNothingWithoutLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Aerith, Last Ancient", "Legendary Creature — Human Cleric Druid", effectstest.AerithLastAncientOracle, false)
	dead := effectstest.PushGraveyardPermanent(me, "Fallen", "Creature — Test", "{2}")

	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(dead) || g.Battlefield.Contains(dead) {
		t.Error("Raise triggered without life gained")
	}
}

func TestCloudDeckYuffieStealsAnArtifactAndAttachesAnEquipment(t *testing.T) {
	cloudDeckWantFull(t, effectstest.YuffieMateriaOracle)
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	blade := effectstest.SeedEquipment(g, me.ID, "Test Blade", "")

	yuffie := effectstest.CastAndResolveCreature(t, g, "Yuffie, Materia Hunter", "Legendary Creature — Human Ninja", effectstest.YuffieMateriaOracle)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, rock)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if card, ok := effectstest.LookupBattlefield(g, rock); !ok || card.Controller != me.ID {
		t.Errorf("the artifact is not under Yuffie's controller: %+v %v", card.Controller, ok)
	}
	if prompt := effectstest.LatestChooseCardsFor(g, me.ID); prompt != nil {
		if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{blade}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	} else {
		t.Fatal("no attach offer")
	}
	if host := effectstest.AttachmentHostOf(t, g, blade); host.ID != yuffie {
		t.Errorf("Equipment host = %+v, want Yuffie", host)
	}
}

func TestCloudDeckCaitSithExilesTheTopCardAndBoostsByItsManaValue(t *testing.T) {
	cloudDeckWantFull(t, effectstest.CaitSithOracle)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Cait Sith, Fortune Teller", "Legendary Artifact Creature — Cat Moogle", effectstest.CaitSithOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	top := effectstest.SeedLibraryTop(me, "Top Spell", "Sorcery", "{3}{R}")

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if scry := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID); scry != nil {
		if err := g.ResolveScry(scry.ID, me.ID, nil, []uuid.UUID{top}); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
	}
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(top) {
		t.Error("the scried top card was not exiled")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 6 {
		t.Errorf("Bear power %d, want 6 (+4 for mana value 4)", got)
	}
}
