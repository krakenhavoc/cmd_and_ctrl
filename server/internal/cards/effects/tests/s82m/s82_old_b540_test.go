package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b540_test.go — Shadowmoor, batch 540 (#5591).

const (
	oldB540Counterbore = "1198ec53-9050-417d-812f-692e1378d346"
	oldB540Consign     = "624bdbfb-b611-43b6-a3c5-cc8b11dbfaae"
	oldB540Guttural    = "9fb80575-9c8e-4745-9a8f-d48cdcbbb868"
	oldB540Cursecatch  = "a59a23f2-158f-4dea-8593-75fe7b31b726"
	oldB540Syphon      = "f30e9371-2c12-496a-80e1-40c3408defed"
	oldB540Initiate    = "227f95ba-65b1-4399-8b12-39f8b1327d62"
	oldB540Hollowsage  = "4e1efd99-8bfc-4650-a430-a754578facbd"
	oldB540Deepchannel = "53b121df-3040-4ad0-a25d-587a299b4f8b"
	oldB540Bloodmark   = "8c22b218-cf40-47af-8c11-73c581a43390"
	oldB540Roughshod   = "ec43fb64-122d-4fc5-9fe0-b7913373f7a4"
)

func TestOldB540EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB540Counterbore: "Counterbore", oldB540Consign: "Consign to Dream", oldB540Guttural: "Guttural Response",
		oldB540Cursecatch: "Cursecatcher", oldB540Syphon: "Spell Syphon", oldB540Initiate: "Intimidator Initiate",
		oldB540Hollowsage: "Hollowsage", oldB540Deepchannel: "Deepchannel Mentor",
		oldB540Bloodmark: "Bloodmark Mentor", oldB540Roughshod: "Roughshod Mentor",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB540ConsignToDreamTucksRedOrGreenAndBouncesTheRest(t *testing.T) {
	for _, tc := range []struct {
		color string
		tuck  bool
	}{{"R", true}, {"G", true}, {"U", false}, {"B", false}} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		perm := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Thing", TypeLine: "Creature — Bear", Colors: []string{tc.color},
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
		})
		effectstest.CastCatalogSpell(t, g, "Consign to Dream", "Instant", oldB540Consign, effectstest.CardTarget(perm))
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, perm) {
			t.Fatalf("%s: still on the battlefield", tc.color)
		}
		if tc.tuck && !(opp.Library.Size() > 0 && effectstest.LibraryTopIDs(opp, 1)[0] == perm) {
			t.Errorf("%s: not on top of its owner's library", tc.color)
		}
		if !tc.tuck && !effectstest.InHandOf(opp, perm) {
			t.Errorf("%s: not returned to its owner's hand", tc.color)
		}
	}
}

func TestOldB540GutturalResponseCountersOnlyBlueInstants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	blueSorcery := effectstest.AvB1Hand(me, game.Card{Name: "Blue Sorcery", TypeLine: "Sorcery", Colors: []string{"U"}})
	blueInstant := effectstest.AvB1Hand(me, game.Card{Name: "Blue Instant", TypeLine: "Instant", Colors: []string{"U"}})
	redInstant := effectstest.AvB1Hand(me, game.Card{Name: "Red Instant", TypeLine: "Instant", Colors: []string{"R"}})
	for _, id := range []uuid.UUID{blueSorcery, blueInstant, redInstant} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
	}
	for _, id := range []uuid.UUID{blueSorcery, redInstant} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Guttural Response", "Instant", oldB540Guttural, effectstest.CardTarget(id)); err == nil {
			t.Error("a spell that is not a blue instant was a legal target")
		}
	}
	effectstest.CastCatalogSpell(t, g, "Guttural Response", "Instant", oldB540Guttural, effectstest.CardTarget(blueInstant))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InGraveyardOf(g, me.ID, blueInstant) {
		t.Error("the blue instant was not countered")
	}
}

func TestOldB540CursecatcherTaxesAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	catcher := effectstest.PushCatalogPermanent(g, me.ID, "Cursecatcher", "Creature — Merfolk Wizard", oldB540Cursecatch, true)
	bolt := effectstest.CastCatalogSpell(t, g, "Sorcery", "Sorcery", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, catcher, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bolt)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.OnBattlefieldByID(g, catcher) {
		t.Error("the Cursecatcher was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the spell's controller was not asked to pay {1}")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InGraveyardOf(g, me.ID, bolt) {
		t.Error("an unpaid spell was not countered")
	}
}

func TestOldB540SpellSyphonTaxesOneForEachBluePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Blue Thing", TypeLine: "Creature — Merfolk", Colors: []string{"U"},
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	}
	effectstest.B12Creature(g, me.ID, "Colorless", "Creature — Bear", 1, 1)
	spell := effectstest.CastCatalogSpell(t, g, "Sorcery", "Sorcery", "", nil)
	effectstest.CastCatalogSpell(t, g, "Spell Syphon", "Instant", oldB540Syphon, effectstest.CardTarget(spell))
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PendingOfKind(g, game.PendingChoicePayUnless)
	if prompt == nil {
		t.Fatal("no pay-unless prompt")
	}
	if prompt.PayCost != "{2}" {
		t.Errorf("the tax is %q, want {2} for the two blue permanents", prompt.PayCost)
	}
}

func TestOldB540CounterboreExilesEveryCardWithTheSpellsName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spell := effectstest.CastCatalogSpell(t, g, "Doomed Bolt", "Sorcery", "", nil)
	other := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Doomed Bolt", TypeLine: "Sorcery"})
	keep := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Other Card", TypeLine: "Sorcery"})
	_ = opp
	effectstest.CastCatalogSpell(t, g, "Counterbore", "Instant", oldB540Counterbore, effectstest.CardTarget(spell))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) {
		t.Error("the countered spell's graveyard copy was not exiled")
	}
	if !g.Exile.Contains(other) {
		t.Error("the same-named card in the library was not exiled (the test takes every match)")
	}
	if g.Exile.Contains(keep) {
		t.Error("a card with another name was exiled")
	}
}

func TestOldB540IntimidatorInitiatePaidStopsTheTargetBlocking(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.B12Push(g, me.ID, "Intimidator Initiate", "Creature — Goblin Shaman", oldB540Initiate, 1, 1)
		victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		red := effectstest.AvB1Hand(me, game.Card{Name: "Red Sorcery", TypeLine: "Sorcery", Colors: []string{"R"}})
		if err := g.CastSpell(me.ID, red, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickCard(t, g, me.ID, victim)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.FillPoolColored(me, "R", 1)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.RestrictionsOf(t, g, victim).Has(game.CantBlock); got != pay {
			t.Errorf("paid=%v: can't block = %v", pay, got)
		}
	}
}

func TestOldB540HollowsageMakesTheTargetPlayerDiscardWhenItUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	me := g.Seats[0]
	sage := effectstest.B12Push(g, me.ID, "Hollowsage", "Creature — Merfolk Wizard", oldB540Hollowsage, 2, 2)
	effectstest.ArnTap(t, g, sage)
	hand := opp.Hand.Size()
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(sage) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, opp.ID) != 1 {
		t.Fatal("the target player was not asked to discard")
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), hand-1)
	}
}

func TestOldB540MentorsGrantTheirKeywordOrEvasionToTheRightColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mk := func(name, color string, oracle string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Goblin", Colors: []string{color}, OracleID: oracle,
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
	}
	bloodmark := mk("Bloodmark Mentor", "R", oldB540Bloodmark)
	roughshod := mk("Roughshod Mentor", "G", oldB540Roughshod)
	red := mk("Red Guy", "R", "")
	green := mk("Green Guy", "G", "")
	blue := mk("Blue Guy", "U", "")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest(); g.RecomputeLayersIfStaleLocked() })
	for _, tc := range []struct {
		id   uuid.UUID
		kw   string
		want bool
		name string
	}{
		{red, "first strike", true, "red creature"}, {bloodmark, "first strike", true, "Bloodmark itself"},
		{green, "first strike", false, "green creature"}, {green, "trample", true, "green creature"},
		{roughshod, "trample", true, "Roughshod itself"}, {red, "trample", false, "red creature"},
		{blue, "trample", false, "blue creature"}, {blue, "first strike", false, "blue creature"},
	} {
		if got := effectstest.HasEffectiveKeyword(t, g, tc.id, tc.kw); got != tc.want {
			t.Errorf("%s has %s = %v, want %v", tc.name, tc.kw, got, tc.want)
		}
	}
}

func TestOldB540DeepchannelMentorMakesBlueCreaturesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Deepchannel Mentor", TypeLine: "Creature — Merfolk Rogue", Colors: []string{"U"}, OracleID: oldB540Deepchannel,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	blue := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Guy", TypeLine: "Creature — Merfolk", Colors: []string{"U"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	red := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Guy", TypeLine: "Creature — Goblin", Colors: []string{"R"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{blue, red} {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, blue); err == nil {
		t.Error("a blue creature was blocked")
	}
	if err := g.DeclareBlocker(blocker, red); err != nil {
		t.Errorf("a red creature could not be blocked: %v", err)
	}
}
