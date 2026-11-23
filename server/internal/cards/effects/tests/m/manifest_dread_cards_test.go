package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest_dread_cards_test.go — #2570 / ADR 0082's 2026-10-07
// amendment through the catalog: the keyword action's seam is tested
// in game/manifest_dread_test.go; here each card is driven through its
// own text. Every test names the SECOND card of the pair or a card that
// is not on top where the choice matters, so a build that took the top
// card would fail it.

const (
	mdManifestDreadOracle   = "df5a66d1-61f6-44d9-b464-3ee739e78dce"
	mdUnsettlingTwinsOracle = "b0b2ca8c-cb5a-45f8-ae7c-63ad9ebb4afd"
	mdInnocuousRatOracle    = "114725ef-ac81-4d13-9153-0a303c3ce9c4"
	mdBashfulBeastieOracle  = "e7aeabd9-a72c-47ee-b275-b025d7142766"
	mdUnnervingGraspOracle  = "7a335332-87ff-42dc-a547-c84295dd8138"
	mdUnderTheSkinOracle    = "8e121178-e9e3-43a5-a174-ecbe47d759f9"
	mdAnalystOracle         = "8cbe71d3-b77b-49a7-9c51-ed308011b8b0"
	mdPipesOracle           = "1b8870fd-2d8a-42af-8173-b747d8d4c04f"
	mdGrowingDreadOracle    = "aa1d8bfa-729a-4e10-8069-57a1d4f4392b"
	mdThreatsOracle         = "5c104ac8-738a-4a9d-a68b-5706018d2061"
	mdKillersMaskOracle     = "93a04ce0-9b8e-4d69-82bf-d4c89a9856a9"
	mdMacheteOracle         = "513dbf8a-a250-4fd9-a5ae-2fb9c4810ecc"
	mdWindbreakerOracle     = "8a265910-2e10-4e97-88ab-c0cb1a712eb8"
	mdToolsOracle           = "50de5faa-d22e-4909-a1ff-3810cfcfcb1a"
	mdBreakDownOracle       = "a11d8372-87fb-40a2-b638-314c6a5af03b"
	mdTwistRealityOracle    = "55b0518e-9c87-4f0e-81c2-9c984f760be1"
	mdOnslaughtOracle       = "663d9559-6619-46fe-8ff4-655f1d5e0e4c"
	mdTurnInsideOutOracle   = "be0a3925-8e0d-4ef6-85cb-c9f5eef6b4bb"
	mdGlitchOracle          = "cf2c507a-41b4-457b-bbe8-edcdfcc681ac"
	mdCuratorOracle         = "6cf8a502-6620-460f-8fab-b4ede28beb11"
	mdMirthOracle           = "7010bfbb-8aea-473b-a9a5-27488de13cc5"
	mdStayHiddenOracle      = "266561cc-d014-48cc-ba5d-fa72f3a91736"
	mdWeightRoomOracle      = "34d0d070-fcaf-410e-a002-012cbe104fb7"
	mdTicketBoothOracle     = "4d01b62b-b924-4da5-8ff5-b2f29d7f19b2"
	mdUnderwaterOracle      = "2b46394e-d337-4b1a-88e6-7fdac2ee4ac4"
	mdExpLabOracle          = "8869df0c-fe84-4964-8d38-a8ecefa9c252"
)

// castDreadModal casts an instant with announced modes and targets.
func castDreadModal(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, modes []int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func TestManifestDreadCardManifestsTheChosenCard(t *testing.T) {
	g, me, top, second := effectstest.DreadTable(t)
	effectstest.CastCatalogSpell(t, g, "Manifest Dread", "Sorcery", mdManifestDreadOracle, nil)
	effectstest.AnswerDread(t, g, me, second)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RequireManifested(t, g, me, second)
	if !effectstest.MdInGraveyard(me, top) {
		t.Error("the card that was not chosen is not in the graveyard")
	}
}

func TestUnsettlingTwinsManifestsDreadWhenItEnters(t *testing.T) {
	g, me, top, second := effectstest.DreadTable(t)
	effectstest.CastCatalogSpell(t, g, "Unsettling Twins", "Creature — Human", mdUnsettlingTwinsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDread(t, g, me, second)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RequireManifested(t, g, me, second)
	if !effectstest.MdInGraveyard(me, top) {
		t.Error("the other card is not in the graveyard")
	}
}

func TestInnocuousRatAndBashfulBeastieManifestDreadWhenTheyDie(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Innocuous Rat", mdInnocuousRatOracle},
		{"Bashful Beastie", mdBashfulBeastieOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, top, second := effectstest.DreadTable(t)
			dier := effectstest.PushPermanentForTest(g, me.ID, tc.name, tc.oracle, "Creature — Test")
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dier) })
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerDread(t, g, me, second)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.RequireManifested(t, g, me, second)
			if !effectstest.MdInGraveyard(me, top) || !effectstest.MdInGraveyard(me, dier) {
				t.Error("the dead creature and the other looked-at card should both be in the graveyard")
			}
		})
	}
}

func TestInnocuousRatDoesNotManifestWhenExiled(t *testing.T) {
	g, me, _, _ := effectstest.DreadTable(t)
	rat := effectstest.PushPermanentForTest(g, me.ID, "Innocuous Rat", mdInnocuousRatOracle, "Creature — Rat")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(rat) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("an exiled Rat manifested dread; only a death does")
	}
}

func TestParanormalAnalystTakesTheCardPutIntoTheGraveyard(t *testing.T) {
	g, me, top, second := effectstest.DreadTable(t)
	effectstest.PushPermanentForTest(g, me.ID, "Paranormal Analyst", mdAnalystOracle, "Creature — Human Detective")
	effectstest.CastCatalogSpell(t, g, "Manifest Dread", "Sorcery", mdManifestDreadOracle, nil)
	effectstest.AnswerDread(t, g, me, second)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Fatal("the card put into the graveyard this way is not in hand")
	}
	if effectstest.MdInGraveyard(me, top) {
		t.Error("the card is still in the graveyard")
	}
	effectstest.RequireManifested(t, g, me, second)
}

func TestParanormalAnalystIgnoresAnOpponentsManifestDread(t *testing.T) {
	g, me, _, _ := effectstest.DreadTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushPermanentForTest(g, me.ID, "Paranormal Analyst", mdAnalystOracle, "Creature — Human Detective")
	handBefore := me.Hand.Size()
	_, oppSecond := effectstest.DreadPair(opp)
	g.WithWriteLock(func() {
		if err := g.ManifestDreadKeyedForEffect(opp.ID, uuid.Nil, game.KeyedTail{}); err != nil {
			t.Fatalf("ManifestDreadKeyedForEffect: %v", err)
		}
	})
	effectstest.AnswerChooseCards(t, g, opp.ID, oppSecond)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Error("the Analyst's controller took a card from an opponent's manifest dread")
	}
}

func TestUnderTheSkinCanReturnTheCardJustPutIntoTheGraveyard(t *testing.T) {
	g, me, top, second := effectstest.DreadTable(t)
	effectstest.CastCatalogSpell(t, g, "Under the Skin", "Sorcery", mdUnderTheSkinOracle, nil)
	effectstest.AnswerDread(t, g, me, second)
	// The second prompt is the "you may return a permanent card" pick.
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no return prompt after the manifest")
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("return prompt bounds %d..%d, want a may-pick of one", prompt.ChooseMin, prompt.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("the creature card was not returned to hand")
	}
}

func TestUnderTheSkinMayDecline(t *testing.T) {
	g, me, top, second := effectstest.DreadTable(t)
	effectstest.CastCatalogSpell(t, g, "Under the Skin", "Sorcery", mdUnderTheSkinOracle, nil)
	effectstest.AnswerDread(t, g, me, second)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(top) || !effectstest.MdInGraveyard(me, top) {
		t.Error("declining the may-return still moved the card")
	}
}

func TestUnnervingGraspBouncesThenManifests(t *testing.T) {
	g, me, top, second := effectstest.DreadTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := effectstest.PushPermanentForTest(g, opp.ID, "Opposing Relic", "", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Unnerving Grasp", "Sorcery", mdUnnervingGraspOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: relic}})
	effectstest.AnswerDread(t, g, me, second)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, relic) != nil || !opp.Hand.Contains(relic) {
		t.Error("the targeted permanent was not returned to its owner's hand")
	}
	effectstest.RequireManifested(t, g, me, second)
	if !effectstest.MdInGraveyard(me, top) {
		t.Error("the other card is not in the graveyard")
	}
}

func TestUnnervingGraspManifestsWithNoTarget(t *testing.T) {
	g, me, _, second := effectstest.DreadTable(t)
	effectstest.CastCatalogSpell(t, g, "Unnerving Grasp", "Sorcery", mdUnnervingGraspOracle, nil)
	effectstest.AnswerDread(t, g, me, second)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RequireManifested(t, g, me, second)
}

func TestBreakDownTheDoorModes(t *testing.T) {
	t.Run("manifest dread asks for no target", func(t *testing.T) {
		g, me, top, second := effectstest.DreadTable(t)
		castDreadModal(t, g, me, "Break Down the Door", "Instant", mdBreakDownOracle, []int{2}, nil)
		effectstest.AnswerDread(t, g, me, second)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.RequireManifested(t, g, me, second)
		if !effectstest.MdInGraveyard(me, top) {
			t.Error("the other card is not in the graveyard")
		}
	})
	t.Run("exile target artifact", func(t *testing.T) {
		g, me, _, _ := effectstest.DreadTable(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		relic := effectstest.PushPermanentForTest(g, opp.ID, "Opposing Relic", "", "Artifact")
		castDreadModal(t, g, me, "Break Down the Door", "Instant", mdBreakDownOracle, []int{0},
			[]game.TargetRef{{Kind: game.TargetCard, ID: relic}})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.FindBattlefieldCardForTest(g, relic) != nil {
			t.Error("the artifact was not exiled")
		}
		if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
			t.Error("the exile mode also manifested dread")
		}
	})
	t.Run("exile target enchantment", func(t *testing.T) {
		g, me, _, _ := effectstest.DreadTable(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		aura := effectstest.PushPermanentForTest(g, opp.ID, "Opposing Aura", "", "Enchantment")
		castDreadModal(t, g, me, "Break Down the Door", "Instant", mdBreakDownOracle, []int{1},
			[]game.TargetRef{{Kind: game.TargetCard, ID: aura}})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.FindBattlefieldCardForTest(g, aura) != nil {
			t.Error("the enchantment was not exiled")
		}
	})
}

func TestTwistRealityModes(t *testing.T) {
	t.Run("manifest dread", func(t *testing.T) {
		g, me, top, second := effectstest.DreadTable(t)
		castDreadModal(t, g, me, "Twist Reality", "Instant", mdTwistRealityOracle, []int{1}, nil)
		effectstest.AnswerDread(t, g, me, second)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.RequireManifested(t, g, me, second)
		if !effectstest.MdInGraveyard(me, top) {
			t.Error("the other card is not in the graveyard")
		}
	})
	t.Run("counter target spell", func(t *testing.T) {
		g, me, _, _ := effectstest.DreadTable(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bolt := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
			OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Owner: opp.ID, Controller: opp.ID})
		life := me.Life
		if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
			t.Fatalf("opponent's Bolt: %v", err)
		}
		castDreadModal(t, g, me, "Twist Reality", "Instant", mdTwistRealityOracle, []int{0},
			[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life {
			t.Errorf("life %d -> %d: the Bolt resolved", life, me.Life)
		}
		if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
			t.Error("the counter mode also manifested dread")
		}
	})
}

func TestTheyCameFromThePipesManifestsTwiceAndDrawsForEach(t *testing.T) {
	g, me, _, _ := effectstest.DreadTable(t)
	// Two pairs: the second look must see the library the first left.
	firstTop, firstSecond := effectstest.DreadPair(me)
	_ = firstTop
	secondTop, secondSecond := effectstest.DreadPair(me)
	_ = secondTop
	// dreadTable already stacked one pair beneath; stack order, top
	// first, is now: secondTop secondSecond firstTop firstSecond + the
	// table's own pair.
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "They Came from the Pipes", "Enchantment", mdPipesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, secondSecond)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, firstSecond)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RequireManifested(t, g, me, secondSecond)
	effectstest.RequireManifested(t, g, me, firstSecond)
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand grew by %d, want one draw per face-down creature that entered (2)", got)
	}
}

func TestGrowingDreadPutsACounterOnAPermanentTurnedFaceUp(t *testing.T) {
	g, me, top, _ := effectstest.DreadTable(t)
	effectstest.CastCatalogSpell(t, g, "Growing Dread", "Enchantment", mdGrowingDreadOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Manifest the creature card so it can be turned face up.
	effectstest.AnswerChooseCards(t, g, me.ID, top)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RequireManifested(t, g, me, top)
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("G", "C")...) })
	if err := g.PerformSpecialAction(me.ID, top, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, top)
	if c == nil || c.FaceDown {
		t.Fatal("the permanent is not face up on the battlefield")
	}
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", c.Counters[game.CounterPlusOne])
	}
}

func TestThreatsAroundEveryCornerFetchesALandForEachFaceDownPermanent(t *testing.T) {
	g, me, _, second := effectstest.DreadTable(t)
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Threats Around Every Corner", "Enchantment", mdThreatsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.RequireManifested(t, g, me, second)
	f := effectstest.FindBattlefieldCardForTest(g, forest)
	if f == nil || !f.Tapped {
		t.Fatalf("the basic land did not enter tapped: %+v", f)
	}
}

func TestManifestDreadEquipmentAttachesToTheManifestedCreature(t *testing.T) {
	cases := []struct {
		name, oracle string
		check        func(t *testing.T, g *game.Game, host uuid.UUID)
	}{
		{"Killer's Mask", mdKillersMaskOracle, func(t *testing.T, g *game.Game, host uuid.UUID) {
			if !effectstest.EffectiveAbilitiesContain(t, g, host, "menace") {
				t.Error("the host has no menace")
			}
		}},
		{"Cursed Windbreaker", mdWindbreakerOracle, func(t *testing.T, g *game.Game, host uuid.UUID) {
			if !effectstest.EffectiveAbilitiesContain(t, g, host, "flying") {
				t.Error("the host has no flying")
			}
		}},
		{"Conductive Machete", mdMacheteOracle, func(t *testing.T, g *game.Game, host uuid.UUID) {
			if p, tough := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 4 || tough != 3 {
				t.Errorf("host is %d/%d, want 4/3", p, tough)
			}
		}},
		{"Dissection Tools", mdToolsOracle, func(t *testing.T, g *game.Game, host uuid.UUID) {
			if p, tough := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 4 || tough != 4 {
				t.Errorf("host is %d/%d, want 4/4", p, tough)
			}
			if !effectstest.EffectiveAbilitiesContain(t, g, host, "deathtouch") || !effectstest.EffectiveAbilitiesContain(t, g, host, "lifelink") {
				t.Error("the host lacks deathtouch or lifelink")
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, me, top, second := effectstest.DreadTable(t)
			equipment := effectstest.CastCatalogSpell(t, g, tc.name, "Artifact — Equipment", tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerChooseCards(t, g, me.ID, second)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.RequireManifested(t, g, me, second)
			eq := effectstest.FindBattlefieldCardForTest(g, equipment)
			if eq == nil || eq.AttachedTo.ID != second {
				t.Fatalf("the Equipment is not attached to the manifested creature: %+v", eq)
			}
			tc.check(t, g, second)
			if !effectstest.MdInGraveyard(me, top) {
				t.Error("the other card is not in the graveyard")
			}
		})
	}
}

func TestDissectionToolsEquipSacrificesACreature(t *testing.T) {
	g, me, _, _ := effectstest.DreadTable(t)
	tools := effectstest.PushPermanentForTest(g, me.ID, "Dissection Tools", mdToolsOracle, "Artifact — Equipment")
	host := effectstest.PushVanillaCreature(g, me.ID, "Host", 2, 2)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, tools, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: host}},
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, fodder) != nil {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if eq := effectstest.FindBattlefieldCardForTest(g, tools); eq == nil || eq.AttachedTo.ID != host {
		t.Errorf("the Equipment did not attach to the target: %+v", eq)
	}
}

func TestValgavothsOnslaughtManifestsXTimesAndCountersEachOfThem(t *testing.T) {
	g, me, _, _ := effectstest.DreadTable(t)
	_, firstSecond := effectstest.DreadPair(me)
	_, secondSecond := effectstest.DreadPair(me)
	effectstest.CastXSpell(t, g, "Valgavoth's Onslaught", "Sorcery", mdOnslaughtOracle, "{X}{X}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, secondSecond)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, firstSecond)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{firstSecond, secondSecond} {
		c := effectstest.RequireManifested(t, g, me, id)
		if c.Counters[game.CounterPlusOne] != 2 {
			t.Errorf("manifested creature has %d +1/+1 counters, want X = 2", c.Counters[game.CounterPlusOne])
		}
	}
}

func TestValgavothsOnslaughtWithXZeroDoesNothing(t *testing.T) {
	g, me, _, _ := effectstest.DreadTable(t)
	effectstest.CastXSpell(t, g, "Valgavoth's Onslaught", "Sorcery", mdOnslaughtOracle, "{X}{X}{G}", 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("X = 0 still looked at the library")
	}
}

func TestTurnInsideOutManifestsDreadWhenTheCreatureDiesThisTurn(t *testing.T) {
	g, me, top, second := effectstest.DreadTable(t)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Turn Inside Out", "Instant", mdTurnInsideOutOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Fatalf("power = %d, want 2 + 3", p)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("manifest dread happened before the creature died")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RequireManifested(t, g, me, second)
	if !effectstest.MdInGraveyard(me, top) {
		t.Error("the other card is not in the graveyard")
	}
}

func TestTurnInsideOutDoesNothingIfTheCreatureSurvives(t *testing.T) {
	g, me, _, _ := effectstest.DreadTable(t)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Turn Inside Out", "Instant", mdTurnInsideOutOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a death after the turn ended still manifested dread")
	}
}

func TestGlitchInterpreterBouncesItselfOnlyWithNoFaceDownPermanent(t *testing.T) {
	t.Run("no face-down permanent", func(t *testing.T) {
		g, me, top, second := effectstest.DreadTable(t)
		id := effectstest.CastCatalogSpell(t, g, "Glitch Interpreter", "Creature — Human Wizard", mdGlitchOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerDread(t, g, me, second)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(id) || effectstest.FindBattlefieldCardForTest(g, id) != nil {
			t.Error("the Interpreter was not returned to its owner's hand")
		}
		effectstest.RequireManifested(t, g, me, second)
		if !effectstest.MdInGraveyard(me, top) {
			t.Error("the other card is not in the graveyard")
		}
	})
	t.Run("a face-down permanent already out", func(t *testing.T) {
		g, me, _, _ := effectstest.DreadTable(t)
		g.WithWriteLock(func() {
			if _, err := g.ManifestForEffect(me.ID); err != nil {
				t.Fatalf("manifest: %v", err)
			}
		})
		id := effectstest.CastCatalogSpell(t, g, "Glitch Interpreter", "Creature — Human Wizard", mdGlitchOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.FindBattlefieldCardForTest(g, id) == nil {
			t.Error("the Interpreter left the battlefield with a face-down permanent out")
		}
		if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
			t.Error("the Interpreter manifested dread with a face-down permanent out")
		}
	})
}

func TestGlitchInterpreterDrawsWhenColorlessCreaturesConnect(t *testing.T) {
	g, me, _, _ := effectstest.DreadTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushPermanentForTest(g, me.ID, "Glitch Interpreter", mdGlitchOracle, "Creature — Human Wizard")
	colorless := effectstest.PushDiesCreatureForTest(g, me.ID, "Colorless Bear", "", "Artifact Creature — Construct", 2, 2)
	green := effectstest.PushDiesCreatureForTest(g, me.ID, "Green Bear", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, green).Colors = []string{"G"} })
	handBefore := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, colorless, green)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("drew %d cards, want 1 for the colorless creature that connected", got)
	}
}

func TestCuratorBeastieMakesItsOwnManifestFourByFour(t *testing.T) {
	g, me, _, second := effectstest.DreadTable(t)
	effectstest.CastCatalogSpell(t, g, "Curator Beastie", "Creature — Beast", mdCuratorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.RequireManifested(t, g, me, second)
	if c.Counters[game.CounterPlusOne] != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2", c.Counters[game.CounterPlusOne])
	}
	// Counters are read through CurrentPower / CurrentToughness, not the
	// layered characteristic (the 2/2 face-down body is layer 1).
	if p, tough := c.CurrentPower(), c.CurrentToughness(); p != 4 || tough != 4 {
		t.Errorf("manifested creature is %d/%d, want 4/4", p, tough)
	}
}

func TestCuratorBeastieLeavesColoredCreaturesAlone(t *testing.T) {
	g, me, _, _ := effectstest.DreadTable(t)
	effectstest.PushPermanentForTest(g, me.ID, "Curator Beastie", mdCuratorOracle, "Creature — Beast")
	bear := game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bear)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, bear.InstanceID); c == nil || c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("a green creature got counters: %+v", c)
	}
}

func TestDisturbingMirthSacrificesForCardsAndManifestsWhenItself(t *testing.T) {
	t.Run("sacrifice another permanent: draw two", func(t *testing.T) {
		g, me, _, _ := effectstest.DreadTable(t)
		fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
		handBefore := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Disturbing Mirth", "Enchantment", mdMirthOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, fodder)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.FindBattlefieldCardForTest(g, fodder) != nil {
			t.Error("the creature was not sacrificed")
		}
		if got := me.Hand.Size() - handBefore; got != 2 {
			t.Errorf("drew %d, want 2", got)
		}
	})
	t.Run("declining draws nothing", func(t *testing.T) {
		g, me, _, _ := effectstest.DreadTable(t)
		fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
		handBefore := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Disturbing Mirth", "Enchantment", mdMirthOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.FindBattlefieldCardForTest(g, fodder) == nil || me.Hand.Size() != handBefore {
			t.Error("declining still sacrificed or drew")
		}
	})
	t.Run("sacrificing the enchantment itself manifests dread", func(t *testing.T) {
		g, me, top, second := effectstest.DreadTable(t)
		mirth := effectstest.PushPermanentForTest(g, me.ID, "Disturbing Mirth", mdMirthOracle, "Enchantment")
		g.WithWriteLock(func() {
			if err := g.SacrificeAllKeyedForEffect(uuid.Nil, []uuid.UUID{mirth}, game.KeyedTail{}); err != nil {
				t.Fatal(err)
			}
		})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerDread(t, g, me, second)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.RequireManifested(t, g, me, second)
		if !effectstest.MdInGraveyard(me, top) {
			t.Error("the other card is not in the graveyard")
		}
	})
	t.Run("destroying it does not", func(t *testing.T) {
		g, me, _, _ := effectstest.DreadTable(t)
		mirth := effectstest.PushPermanentForTest(g, me.ID, "Disturbing Mirth", mdMirthOracle, "Enchantment")
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mirth) })
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
			t.Error("an enchantment that was destroyed manifested dread; only a sacrifice does")
		}
	})
}

func TestStayHiddenStaySilentShufflesTheCreatureInAndManifests(t *testing.T) {
	g, me, _, second := effectstest.DreadTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	aura := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: aura, Name: "Stay Hidden, Stay Silent", TypeLine: "Enchantment — Aura",
		OracleID: mdStayHiddenOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast the Aura: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	v := effectstest.FindBattlefieldCardForTest(g, victim)
	if v == nil || !v.Tapped {
		t.Fatalf("the enchanted creature was not tapped: %+v", v)
	}
	libBefore := opp.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, victim) != nil {
		t.Error("the enchanted creature is still on the battlefield")
	}
	if opp.Library.Size() != libBefore+1 {
		t.Errorf("owner's library %d -> %d, want the creature shuffled in", libBefore, opp.Library.Size())
	}
	effectstest.AnswerChooseCards(t, g, me.ID, effectstest.ChooseCardsChoiceFor(g, me.ID).ChooseCards[0])
	effectstest.PassPriorityAroundTable(t, g)
	_ = second
	manifested := 0
	for _, c := range g.Battlefield.Cards {
		if c.FaceDown && c.FaceDownKind == game.FaceDownManifested {
			manifested++
		}
	}
	if manifested != 1 {
		t.Errorf("manifested creatures on the battlefield = %d, want 1", manifested)
	}
}

func TestStayHiddenStaySilentIsSorcerySpeed(t *testing.T) {
	g, me, _, _ := effectstest.DreadTable(t)
	victim := effectstest.PushVanillaCreature(g, me.ID, "Victim", 3, 3)
	aura := effectstest.PushPermanentForTest(g, me.ID, "Stay Hidden, Stay Silent", mdStayHiddenOracle, "Enchantment — Aura")
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: victim}
	})
	// A different player's turn: sorcery speed is refused.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.ActivateCatalogAbility(opp.ID, aura, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("an opponent activated a sorcery-speed ability of a permanent they do not control")
	}
}

// The four Room doors that shipped empty in PR #2565 (#2555).

func TestWeightRoomPutsThreeCountersOnWhatItManifested(t *testing.T) {
	g, me, top, second := effectstest.DreadTable(t)
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, mdWeightRoomOracle, "Moldering Gym", "{2}{G}", "Weight Room", "{5}{G}"), 1)
	effectstest.RoomsBSettle(t, g, me, second)
	c := effectstest.RequireManifested(t, g, me, second)
	if c.Counters[game.CounterPlusOne] != 3 {
		t.Errorf("counters = %d, want 3", c.Counters[game.CounterPlusOne])
	}
	if !effectstest.MdInGraveyard(me, top) {
		t.Error("the other card is not in the graveyard")
	}
}

func TestSlimyAquariumPutsACounterOnWhatItManifested(t *testing.T) {
	g, me, _, second := effectstest.DreadTable(t)
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, mdUnderwaterOracle, "Underwater Tunnel", "{U}", "Slimy Aquarium", "{3}{U}"), 1)
	effectstest.RoomsBSettle(t, g, me, second)
	if c := effectstest.RequireManifested(t, g, me, second); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("counters = %d, want 1", c.Counters[game.CounterPlusOne])
	}
}

func TestTicketBoothManifestsDread(t *testing.T) {
	g, me, top, second := effectstest.DreadTable(t)
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, mdTicketBoothOracle, "Ticket Booth", "{2}{R}", "Tunnel of Hate", "{4}{R}{R}"), 0)
	effectstest.RoomsBSettle(t, g, me, second)
	effectstest.RequireManifested(t, g, me, second)
	if !effectstest.MdInGraveyard(me, top) {
		t.Error("the other card is not in the graveyard")
	}
}

func TestExperimentalLabPutsTwoCountersAndATrampleCounterOnWhatItManifested(t *testing.T) {
	g, me, _, second := effectstest.DreadTable(t)
	effectstest.RoomsBCast(t, g, me, effectstest.RoomsBCard(me.ID, mdExpLabOracle, "Experimental Lab", "{3}{G}", "Staff Room", "{2}{G}"), 0)
	effectstest.RoomsBSettle(t, g, me, second)
	c := effectstest.RequireManifested(t, g, me, second)
	if c.Counters[game.CounterPlusOne] != 2 || c.Counters[game.CounterTrample] != 1 {
		t.Errorf("counters = %v, want two +1/+1 and one trample", c.Counters)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, second, "trample") {
		t.Error("the trample counter grants no trample")
	}
}

// The doors are complete, and since #2590 so is Staff Room's: it turns a
// face-down creature face up as an effect.
func TestManifestDreadRoomsAreComplete(t *testing.T) {
	for _, id := range []string{mdWeightRoomOracle, mdTicketBoothOracle, mdUnderwaterOracle, mdExpLabOracle} {
		spec, _ := effects.Lookup(id)
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s should be complete now that manifest dread exists", spec.Name)
		}
	}
}
