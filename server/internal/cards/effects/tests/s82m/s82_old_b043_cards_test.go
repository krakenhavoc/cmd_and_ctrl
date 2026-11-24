package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b043_cards_test.go — Strixhaven: School of Mages, batch 043 (S82,
// #5502). Each test plays the real catalog entry through the engine.

const (
	oldB43PracticalOracle = "9110480b-a5d6-4cf8-a72c-dc1a50f11fac"
	oldB43BloodmageOracle = "96c12c12-140a-462d-bff0-0554d192b1f4"
	oldB43PlarggOracle    = "a0e85762-e433-4a23-bbee-69daeebb5898"
	oldB43MentorOracle    = "ba961e01-315a-4195-a6d5-694a97a70b5a"
	oldB43FerventOracle   = "cae6d903-3fd6-422d-b21b-92201b78fa1b"
	oldB43TalentsOracle   = "d808262d-43ca-46f8-ab8b-b584773f9c58"
	oldB43SilencerOracle  = "df9ab83c-850a-4be7-89cd-5122353081f6"
)

func TestOldB43EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		oldB43PracticalOracle:     "Practical Research",
		oldB43BloodmageOracle:     "Callous Bloodmage",
		oldB43PlarggOracle:        "Plargg, Dean of Chaos",
		oldB43PlarggOracle + "#1": "Augusta, Dean of Order",
		oldB43MentorOracle:        "Mentor's Guidance",
		oldB43FerventOracle:       "Fervent Mastery",
		oldB43TalentsOracle:       "Test of Talents",
		oldB43SilencerOracle:      "Silverquill Silencer",
	}
	for oracle, name := range want {
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

// oldB43Cast casts a catalog spell from the hand, paying `mana` exactly.
func oldB43Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, mana)
	params.Strict = true
	id := effectstest.CastFromHandForTest(t, g, me, name, typeLine, cost, oracle, params)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func oldB43ClearHand(p *game.Player) { p.Hand.Cards = nil }

func TestOldB43PracticalResearchDrawsFourThenDiscardsOneInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB43ClearHand(me)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	oldB43Cast(t, g, me, "Practical Research", "Instant", "{3}{U}{R}", oldB43PracticalOracle, "CCCUR", game.CastSpellParams{})
	if me.Hand.Size() != 4 {
		t.Fatalf("hand %d, want the four drawn cards", me.Hand.Size())
	}
	// One instant or sorcery card is enough (the seeded cards are sorceries).
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 3 {
		t.Errorf("hand %d after discarding one sorcery, want 3", me.Hand.Size())
	}
}

func TestOldB43PracticalResearchNeedsTwoCardsWhenNoneIsAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB43ClearHand(me)
	for i := 0; i < 6; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	oldB43Cast(t, g, me, "Practical Research", "Instant", "{3}{U}{R}", oldB43PracticalOracle, "CCCUR", game.CastSpellParams{})
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{me.Hand.Cards[0].InstanceID}); err == nil {
		t.Fatal("discarding one creature card was accepted")
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 2 {
		t.Errorf("hand %d after discarding two, want 2", me.Hand.Size())
	}
}

func TestOldB43CallousBloodmageModes(t *testing.T) {
	// Each mode is chosen as the enters trigger goes on the stack.
	cases := []struct {
		name  string
		mode  int
		check func(t *testing.T, g *game.Game, me, opp *game.Player, before [4]int)
	}{
		{"pest", 0, func(t *testing.T, g *game.Game, me, _ *game.Player, _ [4]int) {
			n := 0
			for _, c := range g.Battlefield.Cards {
				if c.IsToken() && c.Name == "Pest" && c.Controller == me.ID {
					n++
				}
			}
			if n != 1 {
				t.Errorf("%d Pests, want 1", n)
			}
		}},
		{"draw and lose life", 1, func(t *testing.T, g *game.Game, me, _ *game.Player, before [4]int) {
			if me.Hand.Size() != before[0]+1 {
				t.Errorf("hand %d, want %d", me.Hand.Size(), before[0]+1)
			}
			if me.Life != before[1]-1 {
				t.Errorf("life %d, want %d", me.Life, before[1]-1)
			}
		}},
		{"exile a graveyard", 2, func(t *testing.T, g *game.Game, _, opp *game.Player, _ [4]int) {
			if opp.Graveyard.Size() != 0 {
				t.Errorf("the target's graveyard still holds %d cards", opp.Graveyard.Size())
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.SeedLibrary(me, "A", "B", "C")
			effectstest.P1g2Graveyard(opp, "Sorcery", 3)
			effectstest.AdvanceToMain(t, g)
			handBefore := me.Hand.Size()
			lifeBefore := me.Life
			effectstest.CastAndResolveCreature(t, g, "Callous Bloodmage", "Creature — Vampire Warlock", oldB43BloodmageOracle)
			pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
			if pick == nil {
				t.Fatal("no mode_pick prompt")
			}
			if err := g.ResolveModePick(pick.ID, me.ID, []int{tc.mode}); err != nil {
				t.Fatalf("ResolveModePick(%d): %v", tc.mode, err)
			}
			if tc.mode == 2 {
				effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
			}
			effectstest.PassPriorityAroundTable(t, g)
			tc.check(t, g, me, opp, [4]int{handBefore, lifeBefore, 0, 0})
		})
	}
}

func oldB43Wizard(g *game.Game, owner uuid.UUID, typeLine string) uuid.UUID {
	if typeLine == "planeswalker" {
		return effectstest.P1g2Planeswalker(g, owner, 3)
	}
	return effectstest.B12Creature(g, owner, "Mentor", typeLine, 1, 1)
}

func oldB43MentorsGuidance(t *testing.T, g *game.Game, me *game.Player) int {
	t.Helper()
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	effectstest.AdvanceToMain(t, g)
	oldB43ClearHand(me)
	effectstest.FloatForTest(g, me, "CCU")
	effectstest.CastFromHandForTest(t, g, me, "Mentor's Guidance", "Sorcery", "{2}{U}", oldB43MentorOracle, game.CastSpellParams{Strict: true})
	// The cast trigger resolves first, then the spell and its copy: each scrys
	// then draws.
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.ScryChoiceFor(g, me.ID) == nil {
			break
		}
		effectstest.AnswerScryKeepAll(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return me.Hand.Size()
}

func TestOldB43MentorsGuidanceCopiesOnlyWithAMentor(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		want           int
	}{
		{"no mentor", "Creature — Bear", 1},
		{"cleric", "Creature — Human Cleric", 2},
		{"wizard", "Creature — Human Wizard", 2},
		{"druid", "Creature — Elf Druid", 2},
		{"shaman", "Creature — Orc Shaman", 2},
		{"warlock", "Creature — Human Warlock", 2},
		{"planeswalker", "planeswalker", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			oldB43Wizard(g, me.ID, tc.typeLine)
			if got := oldB43MentorsGuidance(t, g, me); got != tc.want {
				t.Errorf("drew %d cards, want %d", got, tc.want)
			}
		})
	}
}

func TestOldB43MentorsGuidanceIgnoresAnOpponentsMentor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB43Wizard(g, opp.ID, "Creature — Human Wizard")
	if got := oldB43MentorsGuidance(t, g, me); got != 1 {
		t.Errorf("drew %d cards, want 1: an opponent's Wizard does not copy it", got)
	}
}

func oldB43LibraryOf(me *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i, n := range names {
		ids[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: ids[i], Name: n, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	return ids
}

func TestOldB43FerventMasteryPlainCastSearchesThreeThenDiscardsThreeAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB43ClearHand(me)
	ids := oldB43LibraryOf(me, "One", "Two", "Three", "Four")
	oldB43Cast(t, g, me, "Fervent Mastery", "Sorcery", "{3}{R}{R}", oldB43FerventOracle, "CCCRR", game.CastSpellParams{})
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0], ids[1], ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Errorf("hand %d, want 0: three tutored, three discarded at random", me.Hand.Size())
	}
	if me.Graveyard.Size() < 3 {
		t.Errorf("graveyard %d, want at least the three discards", me.Graveyard.Size())
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts still pending", len(g.PendingChoices))
	}
}

func TestOldB43FerventMasteryAlternativeCostMakesAnOpponentWheel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB43ClearHand(me)
	for _, s := range g.Seats[2:] {
		oldB43ClearHand(s)
	}
	oldB43ClearHand(opp)
	oldB42Fill := func(p *game.Player, n int) []uuid.UUID {
		var out []uuid.UUID
		for i := 0; i < n; i++ {
			id := uuid.New()
			p.Hand.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
			out = append(out, id)
		}
		return out
	}
	theirs := oldB42Fill(opp, 3)
	effectstest.SeedLibrary(opp, "X", "Y", "Z", "W")
	ids := oldB43LibraryOf(me, "One", "Two", "Three")
	effectstest.FloatForTest(g, me, "CCRR")
	effectstest.CastFromHandForTest(t, g, me, "Fervent Mastery", "Sorcery", "{3}{R}{R}", oldB43FerventOracle,
		game.CastSpellParams{Strict: true, AlternativeCost: "fervent-mastery"})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.AnswerDiscard(t, g, opp.ID, theirs[0], theirs[1])
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 3 {
		t.Errorf("the opponent holds %d cards, want 3 (two discarded, two drawn, one kept)", opp.Hand.Size())
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids...)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Errorf("hand %d, want 0 after three discards at random", me.Hand.Size())
	}
}

func oldB43OppCastsBolt(t *testing.T, x effectstest.ExtractionTable) uuid.UUID {
	t.Helper()
	bolt := uuid.New()
	x.Opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", Owner: x.Opp.ID, Controller: x.Opp.ID})
	effectstest.AdvanceToMain(t, x.G)
	if err := x.G.CastSpell(x.Opp.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Bolt: %v", err)
	}
	return bolt
}

func TestOldB43TestOfTalentsCountersAndExtractsAndTheVictimDraws(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	bolt := oldB43OppCastsBolt(t, x)
	effectstest.CastCatalogSpell(t, x.G, "Test of Talents", "Instant", oldB43TalentsOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, x.G)
	if !x.Opp.Graveyard.Contains(bolt) && !x.G.Exile.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	handBefore := x.Opp.Hand.Size()
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["hand"], x.Cards["lib"], x.Cards["gy"], x.Cards["gy 2"], bolt)
	effectstest.PassPriorityAroundTable(t, x.G)
	x.Exiled(t, "hand", "lib", "gy", "gy 2")
	x.NotExiled(t, "gy other", "hand other", "lib other")
	if !x.G.Exile.Contains(bolt) {
		t.Error("the countered Bolt was not exiled with the rest")
	}
	// One of the exiled cards came from the hand, so the victim draws one.
	if got := x.Opp.Hand.Size(); got != handBefore {
		t.Errorf("hand %d, want %d (one exiled from hand, one drawn)", got, handBefore)
	}
}

func TestOldB43TestOfTalentsRefusesACreatureSpell(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Grizzly Bears", "Creature — Bear")
	id := effectstest.TdCastFromHand(t, x.G, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Test of Talents", "Instant", oldB43TalentsOracle, []game.TargetRef{{Kind: game.TargetCard, ID: id}}); err == nil {
		t.Error("a creature spell was a legal target")
	}
}

func oldB43Silencer(t *testing.T, g *game.Game, me *game.Player, name string) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Silverquill Silencer", "Creature — Human Cleric", oldB43SilencerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, name)
	return id
}

func TestOldB43SilencerPunishesAnOpponentCastingTheChosenName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	oldB43Silencer(t, g, me, "Shock")
	life, hand := them.Life, me.Hand.Size()
	effectstest.CastFromHandForTest(t, g, them, "Shock", "Instant", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != life-3 {
		t.Errorf("the caster is at %d, want %d", them.Life, life-3)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB43SilencerIgnoresOtherNamesAndYourOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	oldB43Silencer(t, g, me, "Shock")
	life, hand := them.Life, me.Hand.Size()
	effectstest.CastFromHandForTest(t, g, them, "Opt", "Instant", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Shock", "Instant", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != life {
		t.Errorf("an unnamed spell cost %d life", life-them.Life)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: no draw for another name or your own spell", me.Hand.Size(), hand)
	}
}

// --- Plargg, Dean of Chaos // Augusta, Dean of Order ---------------------------

func oldB43Dean(owner uuid.UUID, face int) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oldB43PlarggOracle, Layout: game.LayoutModalDFC,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Plargg, Dean of Chaos", TypeLine: "Legendary Creature — Orc Shaman", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: 2, Toughness: 2},
			{Name: "Augusta, Dean of Order", TypeLine: "Legendary Creature — Human Cleric", ManaCost: "{2}{W}", Colors: []string{"W"}, Power: 1, Toughness: 3},
		},
	}
	c.SetFace(face)
	return c
}

func TestOldB43PlarggLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	plargg := effectstest.PushBattlefieldCardWithTimestamp(g, oldB43Dean(me.ID, 0))
	discard := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: discard, Name: "Pitch", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	effectstest.ApaActivate(t, g, me, plargg, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discard}})
	if !me.Graveyard.Contains(discard) {
		t.Error("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (one out, one in)", me.Hand.Size(), hand)
	}
}

func TestOldB43PlarggRevealsUntilACheapNonlegendaryNonlandAndCastsItFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	plargg := effectstest.PushBattlefieldCardWithTimestamp(g, oldB43Dean(me.ID, 0))
	mk := func(name, tl, cost string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: tl, ManaCost: cost, Owner: me.ID, Controller: me.ID}
	}
	hit := mk("Hit", "Creature — Bear", "{2}{G}")
	skip := []game.Card{
		mk("Land", "Basic Land — Forest", ""),
		mk("Legend", "Legendary Creature — Elf", "{1}{G}"),
		mk("Big", "Creature — Bear", "{4}{G}"),
	}
	deeper := mk("Deeper", "Creature — Bear", "{1}")
	me.Library.PushTop(deeper)
	me.Library.PushTop(hit)
	for _, c := range skip {
		me.Library.PushTop(c)
	}
	effectstest.FloatForTest(g, me, "CCCCR")
	effectstest.ApaActivate(t, g, me, plargg, 1, game.ActivateAbilityParams{})
	effectstest.AnswerMayCastPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, hit.InstanceID); !ok {
		t.Fatal("the free cast did not put the hit onto the battlefield")
	}
	for _, c := range skip {
		if z := g.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneLibrary {
			t.Errorf("%s is not back in the library", c.Name)
		}
	}
	if len(me.Library.Cards) < 4 {
		t.Fatalf("library has %d cards", len(me.Library.Cards))
	}
	// The revealed cards that were not cast went to the bottom; the deeper
	// card was never revealed and is now above them.
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:3] {
		bottom[c.InstanceID] = true
	}
	for _, c := range skip {
		if !bottom[c.InstanceID] {
			t.Errorf("%s is not among the bottom cards", c.Name)
		}
	}
}

func TestOldB43PlarggDeclinedHitGoesToTheBottomToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	plargg := effectstest.PushBattlefieldCardWithTimestamp(g, oldB43Dean(me.ID, 0))
	hit := game.Card{InstanceID: uuid.New(), Name: "Hit", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Deeper", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(hit)
	effectstest.FloatForTest(g, me, "CCCCR")
	effectstest.ApaActivate(t, g, me, plargg, 1, game.ActivateAbilityParams{})
	effectstest.AnswerMayCastPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, hit.InstanceID); ok {
		t.Fatal("the hit was cast though the player declined")
	}
	if me.Library.Cards[0].InstanceID != hit.InstanceID {
		t.Errorf("the declined hit is not on the bottom of the library")
	}
}

func TestOldB43AugustaBuffsTappedAndUntappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	augusta := effectstest.PushBattlefieldCardWithTimestamp(g, oldB43Dean(me.ID, 1))
	untapped := effectstest.B12Creature(g, me.ID, "Untapped", "Creature — Bear", 2, 2)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true })
	if p, tough := effectstest.EffectivePower(t, g, untapped), effectstest.EffectiveToughness(t, g, untapped); p != 2 || tough != 3 {
		t.Errorf("an untapped creature is %d/%d, want 2/3", p, tough)
	}
	if p, tough := effectstest.EffectivePower(t, g, tapped), effectstest.EffectiveToughness(t, g, tapped); p != 3 || tough != 2 {
		t.Errorf("a tapped creature is %d/%d, want 3/2", p, tough)
	}
	if p, tough := effectstest.EffectivePower(t, g, augusta), effectstest.EffectiveToughness(t, g, augusta); p != 1 || tough != 3 {
		t.Errorf("Augusta is %d/%d, want her printed 1/3 (it says other)", p, tough)
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(untapped) })
	if p, tough := effectstest.EffectivePower(t, g, untapped), effectstest.EffectiveToughness(t, g, untapped); p != 3 || tough != 2 {
		t.Errorf("after tapping it is %d/%d, want 3/2", p, tough)
	}
}

func TestOldB43AugustaUntapsEverythingThenTapsYourPicks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushBattlefieldCardWithTimestamp(g, oldB43Dean(me.ID, 1))
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	tappedAlready := effectstest.B12Creature(g, me.ID, "Tapped", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, tappedAlready).Tapped = true })
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("declare attacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no tap pick offered")
	}
	for _, id := range []uuid.UUID{attacker, other, tappedAlready} {
		if card, _ := effectstest.BattlefieldCard(g, id); card.Tapped {
			t.Errorf("%s is still tapped before the pick; the trigger untaps each creature first", card.Name)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if card, _ := effectstest.BattlefieldCard(g, attacker); !card.Tapped {
		t.Error("the picked creature is not tapped")
	}
	if card, _ := effectstest.BattlefieldCard(g, other); card.Tapped {
		t.Error("an unpicked creature was tapped")
	}
}
