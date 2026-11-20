package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mom_b06_test.go — S82, March of the Machine, batch 6: convoke, plain
// creatures and spells (#5477).

const (
	momHaloHopperOracle   = "f48d30ee-bba6-4449-8c50-caf0d7f06528"
	momBranchBurnerOracle = "bf7f5e9d-de3d-4a29-af25-f0b56336cecb"
	momThunderheadOracle  = "d74d6f6a-aff0-4a36-9a99-8b2281f5e29d"
	momRefereeOracle      = "e14532e9-ecc2-4028-b99e-3629533962bc"
	momMascotOracle       = "74d9331b-832b-420b-9249-f2a58253a361"
	momWingspanOracle     = "a054b6ed-f6ca-4436-a433-89771ffdd2d6"
	momAerialBoostOracle  = "d7fa5c42-7cd5-42e7-a42e-724c1a3df96a"
	momSunsFuryOracle     = "9c0073d7-a12a-45ba-8ac3-9e2230e2fb06"
	momNightmareOracle    = "fa7b7db2-c830-405c-a5e6-e5431d573ed3"
	momMeetingOracle      = "3b91dc3e-46ef-4a68-829f-ef7d79598cf7"
	momMessageOracle      = "1f7f7765-aeb9-4e47-a6b9-ce823bb29e33"
	momShatterOracle      = "d8b3bcf2-1be1-46ed-9531-ca4ce98e3b3c"
	momGreatswordOracle   = "8c53a177-c8c8-4b9f-a221-f205da67daa5"
)

func TestMomB06EveryCardIsRegisteredFullWithConvoke(t *testing.T) {
	for oracle, name := range map[string]string{
		momHaloHopperOracle: "Halo Hopper", momBranchBurnerOracle: "Shivan Branch-Burner",
		momThunderheadOracle: "Thunderhead Squadron", momRefereeOracle: "Referee Squad",
		momMascotOracle: "Interdisciplinary Mascot", momWingspanOracle: "Astral Wingspan",
		momAerialBoostOracle: "Aerial Boost", momSunsFuryOracle: "Burning Sun's Fury",
		momNightmareOracle: "Collective Nightmare", momMeetingOracle: "Meeting of Minds",
		momMessageOracle: "Transcendent Message", momShatterOracle: "Shatter the Source",
		momGreatswordOracle: "Ramosian Greatsword",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
		if spec.TapCost == nil || spec.TapCost.Key != "convoke" {
			t.Errorf("%s has no convoke", name)
		}
	}
}

// momFloat adds coloured mana to the active seat's pool.
func momFloat(t *testing.T, g *game.Game, me *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
}

func momCast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	params.Strict = true
	return effectstest.CastFromHandForTest(t, g, me, name, typeLine, cost, oracle, params)
}

func momCastErr(g *game.Game, me *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) error {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: oracle, Owner: me.ID, Controller: me.ID})
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func TestMomB06HaloHopperIsPaidByConvokeAndNotByTooFewCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 3)
	if err := momCastErr(g, me, "Halo Hopper", "Artifact Creature — Frog", "{3}", momHaloHopperOracle,
		game.CastSpellParams{TapIDs: tappers[:2]}); err == nil {
		t.Fatal("two creatures paid for {3}")
	}
	hopper := momCast(t, g, me, "Halo Hopper", "Artifact Creature — Frog", "{3}", momHaloHopperOracle,
		game.CastSpellParams{TapIDs: tappers})
	for _, s := range tappers {
		if !effectstest.TapCostTapped(g, s) {
			t.Error("a convoking creature is untapped")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hopper) {
		t.Error("Halo Hopper did not enter")
	}
}

func TestMomB06ShivanBranchBurnerAndThunderheadSquadronHaveTheirKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 10)
	momFloat(t, g, me, "{R}{R}")
	dragon := momCast(t, g, me, "Shivan Branch-Burner", "Creature — Dragon", "{5}{R}{R}", momBranchBurnerOracle,
		game.CastSpellParams{TapIDs: tappers[:5]})
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, dragon, kw) {
			t.Errorf("Shivan Branch-Burner lacks %s", kw)
		}
	}
	momFloat(t, g, me, "{U}")
	knight := momCast(t, g, me, "Thunderhead Squadron", "Creature — Human Knight", "{5}{U}", momThunderheadOracle,
		game.CastSpellParams{TapIDs: tappers[5:10]})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, knight, "flying") {
		t.Error("Thunderhead Squadron lacks flying")
	}
}

func TestMomB06RefereeSquadTapsAndStunsAnOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 2)
	momFloat(t, g, me, "{U}")
	squad := momCast(t, g, me, "Referee Squad", "Creature — Homunculus", "{2}{U}", momRefereeOracle,
		game.CastSpellParams{TapIDs: tappers})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, theirs) || effectstest.CounterCount(g, theirs, game.CounterStun) != 1 {
		t.Errorf("their bear tapped %v, stun %d; want tapped with one stun counter",
			effectstest.TappedForTest(t, g, theirs), effectstest.CounterCount(g, theirs, game.CounterStun))
	}
	if effectstest.TappedForTest(t, g, mine) || effectstest.CounterCount(g, mine, game.CounterStun) != 0 {
		t.Error("my own creature was touched")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, squad, "vigilance") {
		t.Error("Referee Squad lacks vigilance")
	}
}

func TestMomB06InterdisciplinaryMascotLooksAtFourTakesOneAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	effectstest.AdvanceToMain(t, g)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 6)
	momFloat(t, g, me, "{U}{U}")
	hand, lib := me.Hand.Size(), me.Library.Size()
	mascot := momCast(t, g, me, "Interdisciplinary Mascot", "Creature — Elemental Fractal", "{6}{U}{U}", momMascotOracle,
		game.CastSpellParams{TapIDs: tappers})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mascot) {
		t.Fatal("the Mascot did not enter")
	}
	if !me.Hand.Contains(ids[2]) || me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1 (the pick; the spell was pushed and left)", hand, me.Hand.Size())
	}
	if got := effectstest.LibraryTopNames(me, 2); got[0] != "E" || got[1] != "F" {
		t.Errorf("library top is %v, want E F (the fifth and sixth cards untouched)", got)
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d → %d, want one card fewer", lib, me.Library.Size())
	}
}

func TestMomB06InterdisciplinaryMascotHasWardThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mascot := effectstest.PushCatalogPermanent(g, me.ID, "Interdisciplinary Mascot", "Creature — Elemental Fractal", momMascotOracle, false)
	effectstest.CastAtWardedCreature(t, g, opp, mascot)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the opponent's spell")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mascot) {
		t.Error("a declined ward must counter the spell")
	}
}

func TestMomB06AstralWingspanDrawsAndGivesPlusTwoAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 4)
	momFloat(t, g, me, "{U}")
	hand := me.Hand.Size()
	momCast(t, g, me, "Astral Wingspan", "Enchantment — Aura", "{4}{U}", momWingspanOracle,
		game.CastSpellParams{TapIDs: tappers, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Errorf("Bear = %d/%d, want a 4/4 flier", c.CurrentPower(), c.CurrentToughness())
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1 (one card drawn)", hand, me.Hand.Size())
	}
}

func TestMomB06AerialBoostPumpsAndGrantsFlyingForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 2)
	momCast(t, g, me, "Aerial Boost", "Instant", "{1}{W}", momAerialBoostOracle,
		game.CastSpellParams{TapIDs: tappers, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 4 || !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Errorf("Bear = %d/%d, want a 4/4 flier", p, tt)
	}
	effectstest.EndTurn(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 2 || tt != 2 || effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Errorf("after the turn Bear = %d/%d, want a plain 2/2", p, tt)
	}
}

func TestMomB06BurningSunsFuryPumpsAndHastesUpToTwoAndRefusesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	c := effectstest.PushVanillaCreature(g, me.ID, "Bear C", 2, 2)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 2)
	ref := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, id := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		return out
	}
	momFloat(t, g, me, "{R}")
	if err := momCastErr(g, me, "Burning Sun's Fury", "Instant", "{1}{R}", momSunsFuryOracle,
		game.CastSpellParams{TapIDs: tappers[:1], Targets: ref(a, b, c)}); err == nil {
		t.Fatal("three targets were accepted for up to two")
	}
	momCast(t, g, me, "Burning Sun's Fury", "Instant", "{1}{R}", momSunsFuryOracle,
		game.CastSpellParams{TapIDs: tappers[:1], Targets: ref(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tt := effectstest.PtOf(t, g, id); p != 4 || tt != 2 || !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
			t.Errorf("target = %d/%d, want 4/2 with haste", p, tt)
		}
	}
	if p, _ := effectstest.PtOf(t, g, c); p != 2 || effectstest.HasKeywordOnBattlefield(t, g, c, "haste") {
		t.Error("the untargeted creature changed")
	}
}

func TestMomB06CollectiveNightmareGivesMinusThreeAndKillsAToughnessThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 3)
	momFloat(t, g, me, "{B}")
	momCast(t, g, me, "Collective Nightmare", "Instant", "{2}{B}", momNightmareOracle,
		game.CastSpellParams{TapIDs: tappers[:2], Targets: effectstest.B16TargetCard(big)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, big); p != 2 || tt != 2 {
		t.Errorf("Big = %d/%d, want 2/2", p, tt)
	}
	momFloat(t, g, me, "{B}{B}")
	momCast(t, g, me, "Collective Nightmare", "Instant", "{2}{B}", momNightmareOracle,
		game.CastSpellParams{TapIDs: tappers[2:], Targets: effectstest.B16TargetCard(small)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 3/3 survived -3/-3")
	}
}

func TestMomB06MeetingOfMindsDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 3)
	momFloat(t, g, me, "{U}")
	hand := me.Hand.Size()
	momCast(t, g, me, "Meeting of Minds", "Instant", "{3}{U}", momMeetingOracle, game.CastSpellParams{TapIDs: tappers})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want +2 (the spell left, two drawn)", hand, me.Hand.Size())
	}
}

func TestMomB06TranscendentMessageDrawsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 3)
	momFloat(t, g, me, "{U}{U}{U}{U}")
	hand := me.Hand.Size()
	momCast(t, g, me, "Transcendent Message", "Instant", "{X}{U}{U}{U}{U}", momMessageOracle,
		game.CastSpellParams{XValue: 3, TapIDs: tappers})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want +3 for X=3 (the spell left, three drawn)", hand, me.Hand.Size())
	}
}

func TestMomB06ShatterTheSourceDealsSixToACreaturePlaneswalkerOrBattleAndDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 6)
	cast := func(taps []uuid.UUID, mode int, target uuid.UUID) {
		t.Helper()
		momFloat(t, g, me, "{R}")
		if taps == nil {
			momFloat(t, g, me, "{R}{R}{R}{R}{R}")
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Shatter the Source", TypeLine: "Instant",
			ManaCost: "{5}{R}", OracleID: momShatterOracle, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: taps,
			Modes: []int{mode}, Targets: effectstest.B16TargetCard(target)}); err != nil {
			t.Fatalf("cast mode %d: %v", mode, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	// A creature can't be the target of the artifact mode.
	momFloat(t, g, me, "{R}")
	bad := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bad, Name: "Shatter the Source", TypeLine: "Instant",
		ManaCost: "{5}{R}", OracleID: momShatterOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bad, game.CastSpellParams{Strict: true, TapIDs: tappers[:5],
		Modes: []int{1}, Targets: effectstest.B16TargetCard(giant)}); err == nil {
		t.Fatal("the destroy-artifact mode took a non-artifact creature")
	}
	g.WithWriteLock(func() { me.Hand.Remove(bad) })

	cast(tappers[:5], 0, giant)
	if g.Battlefield.Contains(giant) {
		t.Error("six damage left a 6/6 alive")
	}
	if effectstest.TapCostTapped(g, tappers[5]) {
		t.Error("an unused creature was tapped")
	}
	cast(nil, 1, rock)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived the destroy mode")
	}
}

func TestMomB06ShatterTheSourceHitsAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 4},
	})
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 5)
	momFloat(t, g, me, "{R}")
	momCast(t, g, me, "Shatter the Source", "Instant", "{5}{R}", momShatterOracle,
		game.CastSpellParams{TapIDs: tappers, Modes: []int{0}, Targets: effectstest.B16TargetCard(walker)})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(walker) {
		t.Error("six damage left a four-loyalty planeswalker on the battlefield")
	}
}

func TestMomB06RamosianGreatswordGivesPlusThreePlusOneAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 4)
	momFloat(t, g, me, "{R}")
	sword := momCast(t, g, me, "Ramosian Greatsword", "Artifact — Equipment", "{4}{R}", momGreatswordOracle,
		game.CastSpellParams{TapIDs: tappers})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EquipTo(t, g, me.ID, sword, bear)
	if p, tt := effectstest.PtOf(t, g, bear); p != 5 || tt != 3 {
		t.Errorf("equipped Bear = %d/%d, want 5/3", p, tt)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("equipped creature lacks trample")
	}
}
