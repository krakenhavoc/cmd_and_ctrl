package testsru

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB91Astrid  = "fa55177c-ba65-4113-9c05-2d63318cf7c5"
	gapB91Atraxi  = "5f31a2d6-649a-40f4-89ce-95c635340e9e"
	gapB91Barbara = "0e61a062-f13a-4958-91de-909650c662a8"
	gapB91Ian     = "32c9ce8c-a19d-49e4-bc6d-c7182fa00b51"
	gapB91Jo      = "65ea3638-1766-4096-88a8-662a0d6faf0c"
	gapB91Peri    = "2fadf4aa-1b7e-4e72-ad20-809a59ddccf8"
	gapB91Romana  = "3346a585-d35a-4f8b-b9ab-7e453ebd15f4"
	gapB91Rose    = "2f56de72-80d7-48c6-9b57-84fbd252c699"
	gapB91Sarah   = "da268892-b190-4eee-9a2b-b0b9bb763033"
	gapB91Tegan   = "ac46f435-5b25-4b46-b89f-175bca81b9ad"
	gapB91Wilfred = "5a970742-687f-49ed-8822-e2953f4530b6"
)

var gapB91Names = map[string]string{
	gapB91Astrid:  "Astrid Peth",
	gapB91Atraxi:  "Atraxi Warden",
	gapB91Barbara: "Barbara Wright",
	gapB91Jo:      "Jo Grant",
	gapB91Peri:    "Peri Brown",
	gapB91Romana:  "Romana II",
	gapB91Rose:    "Rose Tyler",
	gapB91Sarah:   "Sarah Jane Smith",
	gapB91Tegan:   "Tegan Jovanka",
	gapB91Wilfred: "Wilfred Mott",
}

func TestGapB91CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range gapB91Names {
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

func gapB91Count(g *game.Game, owner uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

func TestGapB91AtraxiWardenExilesATappedCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(gapB91Atraxi)
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Counters != 5 || spec.SpecialActions[0].Cost != "{1}{W}" {
		t.Errorf("Atraxi Warden's suspend is %+v, want suspend 5 for {1}{W}", spec.SpecialActions)
	}
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped Foe", "Creature — Bear", 2, 2)
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped Foe", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffectBy(tapped, opp.ID) })
	id := effectstest.AvB1Hand(me, game.Card{Name: "Atraxi Warden", TypeLine: "Creature — Alien Eye", OracleID: gapB91Atraxi, Power: 6, Toughness: 6})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// An untapped creature is not a legal target.
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: untapped}); err == nil {
			t.Error("an untapped creature was a legal target")
		}
		effectstest.PickTriggerTarget(t, g, me.ID, tapped)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(tapped) {
		t.Error("the tapped creature was not exiled")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature was exiled")
	}
}

func TestGapB91SarahJaneSmithInvestigatesOnceEachTurnOnAHistoricSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Sarah Jane Smith", "Legendary Creature — Human Detective", gapB91Sarah, 2, 1)
	plain := effectstest.AvB1Hand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if gapB91Count(g, me.ID, "Clue") != 0 {
		t.Fatal("a nonhistoric spell investigated")
	}
	for i, tl := range []string{"Artifact", "Legendary Creature — Human", "Enchantment — Saga"} {
		id := effectstest.AvB1Hand(me, game.Card{Name: "Historic " + tl, TypeLine: tl, Power: 1, Toughness: 1})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := gapB91Count(g, me.ID, "Clue"); got != 1 {
			t.Fatalf("after historic spell %d clues %d, want 1 (once each turn)", i+1, got)
		}
	}
}

func TestGapB91AstridPethMakesFoodAndExploresOnASacrificedClueOrFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	astrid := effectstest.Msc17Play(t, g, me, game.Card{Name: "Astrid Peth", TypeLine: "Legendary Creature — Human", OracleID: gapB91Astrid, Power: 2, Toughness: 2})
	if got := gapB91Count(g, me.ID, "Food"); got != 1 {
		t.Fatalf("Food after entering %d, want 1", got)
	}
	var food uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Food") {
			food = c.InstanceID
		}
	}
	effectstest.PutOnTopForTest(me, "Some Sorcery", "Sorcery")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(food) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerExploreQuestion(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, astrid, game.CounterPlusOne); got != 1 {
		t.Errorf("Astrid has %d counters after exploring a nonland, want 1", got)
	}
	// A nontoken-nonFood sacrifice does nothing.
	goat := effectstest.B12Creature(g, me.ID, "Goat", "Creature — Goat", 1, 1)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(goat) })
	effectstest.PassPriorityAroundTable(t, g)
	// Attacking makes another Food.
	_ = astrid
	second := effectstest.B12Push(g, me.ID, "Astrid Peth", "Creature — Human", gapB91Astrid, 2, 2) // not legendary, so the legend rule leaves it alone
	effectstest.StdB42Attack(t, g, opp.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB91Count(g, me.ID, "Food"); got != 1 {
		t.Errorf("Food after attacking %d, want 1", got)
	}
}

func TestGapB91RomanaCopiesATokenThatEnteredThisTurnTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	romana := effectstest.B12Push(g, me.ID, "Romana II", "Legendary Creature — Time Lord Scientist", gapB91Romana, 3, 3)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1}, 1)
	})
	old := effectstest.B12Permanent(g, me.ID, "Old Token", "Token Artifact — Treasure")
	_ = old
	var tok uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" {
			tok = c.InstanceID
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, romana, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(old)}); err == nil {
		t.Fatal("copied a token that did not enter this turn")
	}
	effectstest.ApaMana(me, "C")
	effectstest.ApaActivate(t, g, me, romana, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(tok)})
	goblins, tapped := 0, 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" {
			goblins++
			if c.Tapped {
				tapped++
			}
		}
	}
	if goblins != 2 || tapped != 1 {
		t.Errorf("goblins %d, tapped %d; want 2 and 1 tapped", goblins, tapped)
	}
}

func TestGapB91BarbaraWrightGivesSagasReadAhead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Barbara Wright", "Legendary Creature — Human Advisor", gapB91Barbara, 1, 3)
	saga := effectstest.B12Permanent(g, me.ID, "Test Saga", "Enchantment — Saga")
	other := effectstest.B12Permanent(g, g.Seats[1].ID, "Their Saga", "Enchantment — Saga")
	enchant := effectstest.B12Permanent(g, me.ID, "Plain Aura", "Enchantment")
	has := func(id uuid.UUID) bool {
		for _, a := range effectstest.LayeredCard(t, g, id).Effective().Abilities {
			if a == game.KeywordReadAhead {
				return true
			}
		}
		return false
	}
	if !has(saga) {
		t.Error("my Saga lacks read ahead")
	}
	if has(other) || has(enchant) {
		t.Error("a Saga I don't control, or a non-Saga, has read ahead")
	}
}

func TestGapB91PeriBrownGivesTheFirstHistoricSpellEachTurnConvoke(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Peri Brown", "Legendary Creature — Human", gapB91Peri, 2, 3)
	elves := []uuid.UUID{effectstest.B12Creature(g, me.ID, "Elf One", "Creature — Elf", 1, 1), effectstest.B12Creature(g, me.ID, "Elf Two", "Creature — Elf", 1, 1)}
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Fatal("a nonhistoric spell was convoked")
	}
	rock := game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(rock) })
	if err := g.CastSpell(me.ID, rock.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking the first historic spell: %v", err)
	}
	if !effectstest.CgTapped(g, elves) || !effectstest.CpHadConvoke(g, rock.InstanceID) {
		t.Error("the first historic spell did not use convoke")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, e := range elves {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(e) })
	}
	rock2 := game.Card{InstanceID: uuid.New(), Name: "Rock Two", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(rock2) })
	if err := g.CastSpell(me.ID, rock2.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Error("the second historic spell was convoked")
	}
}

func TestGapB91RoseTylerGrowsWithTimeCountersAndAddsMoreWhenAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rose := effectstest.B12Push(g, me.ID, "Rose Tyler", "Legendary Creature — Human", gapB91Rose, 2, 2)
	clock := effectstest.B12Permanent(g, me.ID, "Clock", "Artifact")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(clock, game.CounterTime, 3) })
	suspended := game.Card{InstanceID: uuid.New(), Name: "Suspended Spell", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() {
		g.Exile.PushTop(suspended)
		_ = g.GainSuspendForEffect(me.ID, suspended.InstanceID, 2)
	})
	effectstest.StdB42Attack(t, g, opp.ID, rose)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, rose, game.CounterTime); got != 2 {
		t.Errorf("Rose has %d time counters, want 2 (one suspended card, one other permanent)", got)
	}
	if p := effectstest.EffectivePower(t, g, rose); p != 4 {
		t.Errorf("Rose has power %d, want 4 (2 plus two time counters)", p)
	}
}

func TestGapB91TeganJovankaPumpsAnAttackingHistoricCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Tegan Jovanka", "Legendary Creature — Human", gapB91Tegan, 2, 2)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.StdB42Attack(t, g, opp.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, legend); p != 3 {
		t.Errorf("the target has power %d, want 3", p)
	}
	found := false
	for _, a := range effectstest.LayeredCard(t, g, legend).Effective().Abilities {
		if a == "indestructible" {
			found = true
		}
	}
	if !found {
		t.Error("the target is not indestructible")
	}
}

func TestGapB91JoGrantGivesHistoricCardsInHandCyclingAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	jo := effectstest.B12Push(g, me.ID, "Jo Grant", "Legendary Creature — Human Soldier", gapB91Jo, 3, 2)
	historic := game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", OracleID: "zone-grant-fixture-vanilla", Owner: me.ID, Controller: me.ID}
	plain := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", OracleID: "zone-grant-fixture-vanilla", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(historic); me.Hand.PushTop(plain) })
	if rows, _ := effectstest.ZgRows(g, historic.InstanceID); len(rows) != 1 || !strings.HasPrefix(rows[0].Label, "Cycling {2}{W}") {
		t.Fatalf("historic rows = %v, want Cycling {2}{W}", effectstest.ZgLabels(rows))
	}
	if rows, _ := effectstest.ZgRows(g, plain.InstanceID); len(rows) != 0 {
		t.Errorf("a nonhistoric card has rows %v", effectstest.ZgLabels(rows))
	}
	effectstest.ApaMana(me, "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, historic.InstanceID, 0, game.ActivateAbilityParams{
		Ref: "grant:jo-grant/cycling:0:0",
	}); err != nil {
		t.Fatalf("cycle the artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, jo, game.CounterPlusOne); got != 1 {
		t.Errorf("Jo has %d counters after a cycle, want 1", got)
	}
}

func TestGapB91WilfredMottGrowsAndDigsForCheapPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	next := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wilfred := effectstest.B12Push(g, next.ID, "Wilfred Mott", "Creature — Human Soldier", gapB91Wilfred, 2, 4)
	g.WithWriteLock(func() {
		for next.Library.Size() > 0 {
			_, _ = next.Library.PopTop()
		}
	})
	dear := effectstest.SeedLibraryTop(next, "Dear Rock", "Artifact", "{5}")
	land := effectstest.SeedLibraryTop(next, "Forest", "Basic Land — Forest", "")
	cheap := effectstest.SeedLibraryTop(next, "Cheap Rock", "Artifact", "{2}")
	_ = land
	got := effectstest.ToNextUpkeep(t, g)
	if got.ID != next.ID {
		t.Fatalf("the upkeep is %s's, want %s's", got.Name, next.Name)
	}
	effectstest.PassPriorityAroundTable(t, g)
	answerChooseCardsOrPickFor(t, g, next.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.CountersOn(g, wilfred, game.CounterTime); c != 1 {
		t.Errorf("Wilfred has %d time counters, want 1", c)
	}
	if !g.Battlefield.Contains(cheap) {
		t.Error("the cheap permanent was not put onto the battlefield")
	}
	if g.Battlefield.Contains(dear) {
		t.Error("the dear one was put onto the battlefield")
	}
}

// answerChooseCardsOrPickFor answers whichever of the two pick prompts the
// dig raises with `id`.
func answerChooseCardsOrPickFor(t *testing.T, g *game.Game, chooser, id uuid.UUID) {
	t.Helper()
	if effectstest.ChooseCardsChoiceFor(g, chooser) != nil {
		effectstest.AnswerChooseCards(t, g, chooser, id)
		return
	}
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, chooser); c != nil {
		if err := g.ResolveSearchLibrary(c.ID, chooser, []uuid.UUID{id}); err != nil {
			t.Fatal(err)
		}
		g.SettleResolution()
		return
	}
	t.Fatalf("no pick prompt for %s: %+v", chooser, g.PendingChoices)
}
