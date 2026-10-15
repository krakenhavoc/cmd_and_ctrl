package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tidus_c_test.go — the Tidus, Yuna's Guardian deck request (#3246),
// slice C: Yuna, the Summons and the rest of the spells.

const (
	tdIxionOracle         = "1ccb8122-9135-4b8d-b1d1-18997fd7e9d2"
	tdValeforOracle       = "67d29e56-9ca8-46bc-9ee0-20184e867ce4"
	tdYunasDecisionOracle = "77dd6ce0-91fc-4e71-99ff-47cdf3359067"
	tdYunasWhistleOracle  = "39bf633c-41a6-496d-8aef-1ab7befb503d"
	tdEndlessDetourOracle = "d65728cf-e9ce-46d8-a890-e1295b18cfbe"
	tdFightRiggingOracle  = "117d89b2-b4c0-4421-a631-37810c0e0969"
	tdBlitzballOracle     = "a4b66fb3-16ea-4bf1-bce3-6df872e8883b"
	tdScholarOracle       = "3c7ea603-c985-4107-806b-a467b5fcca36"
	tdYunaOracle          = "ba72d4a9-6f36-4796-bd71-2342fa9b4a30"
	tdAlteredEgoOracle    = "7c35f3fd-c64e-4944-a4d5-37ce916d23c3"
	tdKimahriOracle       = "6d42e81a-84d3-4c90-9c69-c0d562c66cac"
)

func TestTidusCSliceIsRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		tdIxionOracle:         "Summon: Ixion",
		tdValeforOracle:       "Summon: Valefor",
		tdYunasDecisionOracle: "Yuna's Decision",
		tdYunasWhistleOracle:  "Yuna's Whistle",
		tdEndlessDetourOracle: "Endless Detour",
		tdFightRiggingOracle:  "Fight Rigging",
		tdBlitzballOracle:     "Blitzball Stadium",
		tdScholarOracle:       "Scholar of New Horizons",
		tdYunaOracle:          "Yuna, Grand Summoner",
		tdAlteredEgoOracle:    "Altered Ego",
		tdKimahriOracle:       "Kimahri, Valiant Guardian",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

// tdCastFromHand puts a card into the active seat's hand and casts it in
// a main phase.
func tdCastFromHand(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return c.InstanceID
}

func TestTidusCSummonIxionExilesThenCountersAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	threat := tdCreature(g, opp.ID, "Threat", 4, 4)
	bear := tdCreature(g, me.ID, "Bear", 2, 2)

	tdCastFromHand(t, g, game.Card{Name: "Summon: Ixion", TypeLine: "Enchantment Creature — Saga Unicorn",
		OracleID: tdIxionOracle, ManaCost: "{2}{W}", Power: 3, Toughness: 3}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, threat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(threat) {
		t.Fatal("chapter I did not exile the creature")
	}

	advanceToPrecombatMainOf(t, g, seat)
	life := me.Life
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, tdCard(bear)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("chapter II put %d counters, want 1", got)
	}
	if me.Life != life+2 {
		t.Errorf("chapter II gained %d life, want 2", me.Life-life)
	}
}

func TestTidusCSummonValeforBouncesTheBiggestAndStuns(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	small := paPermanent(g, opp.ID, "Small", "Creature — Test", "{1}", 1, 1)
	big := paPermanent(g, opp.ID, "Big", "Creature — Test", "{4}{G}", 5, 5)

	tdCastFromHand(t, g, game.Card{Name: "Summon: Valefor", TypeLine: "Enchantment Creature — Saga Drake",
		OracleID: tdValeforOracle, ManaCost: "{4}{U}", Power: 5, Toughness: 4}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
	if c == nil {
		t.Fatalf("the opponent was not asked to choose: %+v", g.PendingChoices)
	}
	if err := g.ResolveOwnPermanents(c.ID, opp.ID, []uuid.UUID{small}); err == nil {
		t.Error("a creature without the greatest mana value was accepted")
	}
	if err := g.ResolveOwnPermanents(c.ID, opp.ID, []uuid.UUID{big}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(big) {
		t.Fatal("the chosen creature did not return to its owner's hand")
	}

	advanceToPrecombatMainOf(t, g, seat)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, tdCard(small)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, small, game.CounterStun); got != 1 {
		t.Errorf("chapter II put %d stun counters, want 1", got)
	}
}

func TestTidusCYunasDecisionReturnsTwoPermanentCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushGraveyardCreature(g, me.ID, "Dead A", "{2}")
	b := pushGraveyardCreature(g, me.ID, "Dead B", "{3}")
	castModal(t, g, "Yuna's Decision", "Sorcery", tdYunasDecisionOracle, []int{1}, append(tdCard(a), tdCard(b)...))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two permanent cards did not return to hand")
	}
}

func TestTidusCYunasDecisionSacrificesDrawsAndPutsACreatureAndALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := tdCreature(g, me.ID, "Fodder", 1, 1)
	land := game.NewCard("Forest", me.ID)
	land.TypeLine = "Basic Land — Forest"
	me.Hand.PushTop(land)
	beast := game.Card{InstanceID: uuid.New(), Name: "Beast", TypeLine: "Creature — Beast", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(beast)
	land2 := game.NewCard("Plains", me.ID)
	land2.TypeLine = "Basic Land — Plains"
	me.Hand.PushTop(land2)

	castModal(t, g, "Yuna's Decision", "Sorcery", tdYunasDecisionOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, fodder)
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("no put-from-hand prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land.InstanceID, land2.InstanceID}); err == nil {
		t.Error("two lands were accepted for a creature card and/or a land card")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land.InstanceID, beast.InstanceID}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land.InstanceID) || !g.Battlefield.Contains(beast.InstanceID) {
		t.Error("the creature and the land did not enter")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
}

func TestTidusCYunasWhistleFindsACreatureAndCountersByItsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := tdCreature(g, me.ID, "Bear", 2, 2)
	hit := game.NewCard("Big Find", me.ID)
	hit.TypeLine, hit.ManaCost = "Creature — Beast", "{3}{G}"
	me.Library.PushTop(hit)
	putOnTopForTest(me, "Above", "Sorcery")

	castCatalogSpell(t, g, "Yuna's Whistle", "Instant", tdYunasWhistleOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(hit.InstanceID) {
		t.Error("the creature card did not go to hand")
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 4 {
		t.Errorf("the target has %d counters, want 4 (the mana value)", got)
	}
}

func TestTidusCEndlessDetourTucksAPermanentAndAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	relic := b12Push(g, opp.ID, "Their Relic", "Artifact", "", 0, 0)
	castCatalogSpell(t, g, "Endless Detour", "Instant", tdEndlessDetourOracle, tdCard(relic))
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{relic}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if !opp.Library.Contains(relic) {
		t.Error("the permanent is not in its owner's library")
	}

	dead := pushGraveyardCreature(g, opp.ID, "Dead", "{2}")
	castCatalogSpell(t, g, "Endless Detour", "Instant", tdEndlessDetourOracle, tdCard(dead))
	passPriorityAroundTable(t, g)
	c = latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatal("the owner was not asked about the graveyard card")
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, nil, []uuid.UUID{dead}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if !opp.Library.Contains(dead) {
		t.Error("the graveyard card is not in its owner's library")
	}
}

func TestTidusCEndlessDetourTucksASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := castCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	detour := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: detour, Name: "Endless Detour", TypeLine: "Instant",
		OracleID: tdEndlessDetourOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, detour, game.CastSpellParams{Targets: tdCard(spell)}); err != nil {
		t.Fatalf("cast Endless Detour at a spell: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	for i := 0; i < 8 && latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if c == nil {
		t.Fatal("the spell's owner was not asked")
	}
	if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, []uuid.UUID{spell}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if !me.Library.Contains(spell) {
		t.Error("the spell is not in its owner's library")
	}
	if g.StackItemForEffect(spell) != nil {
		t.Error("the spell is still on the stack")
	}
}

func TestTidusCFightRiggingCountersAndOffersTheHiddenCardAtSevenPower(t *testing.T) {
	for _, tc := range []struct {
		name      string
		power     int
		wantGrant bool
	}{
		{"six power after the counter", 5, false},
		{"seven power after the counter", 6, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceTo(t, g, game.StepPrecombatMain)
			top := stackTopLibrary(me, "A", "B", "C", "D", "E")
			rig := game.NewCard("Fight Rigging", me.ID)
			rig.OracleID = tdFightRiggingOracle
			rig.TypeLine = "Enchantment"
			rig.ManaCost = "{2}{G}"
			enterWithHideaway(t, g, rig, top[0])

			big := tdCreature(g, me.ID, "Big", tc.power, tc.power)
			advanceTo(t, g, game.StepBeginCombat)
			b04WaitForPick(t, g, me.ID)
			pickCard(t, g, me.ID, big)
			passPriorityAroundTable(t, g)
			if got := countersOn(g, big, game.CounterPlusOne); got != 1 {
				t.Errorf("the creature has %d counters, want 1", got)
			}
			if got := playHiddenGrant(g, top[0]) != nil; got != tc.wantGrant {
				t.Errorf("grant over the hidden card = %v, want %v", got, tc.wantGrant)
			}
		})
	}
}

func TestTidusCBlitzballStadiumSupportsXAndGrantsTheGoal(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	a := tdCreature(g, me.ID, "A", 2, 2)
	b := tdCreature(g, me.ID, "B", 2, 2)
	stadium := tdCastFromHand(t, g, game.Card{Name: "Blitzball Stadium", TypeLine: "Artifact",
		OracleID: tdBlitzballOracle, ManaCost: "{X}{U}"}, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no support prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, append(tdCard(a), tdCard(b)...)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countersOn(g, a, game.CounterPlusOne) != 1 || countersOn(g, b, game.CounterPlusOne) != 1 {
		t.Fatal("support X did not counter both creatures")
	}

	fillPoolColored(me, "C", 3)
	if err := g.ActivateCatalogAbility(me.ID, stadium, 0, game.ActivateAbilityParams{Targets: tdCard(b)}); err != nil {
		t.Fatalf("activate the Stadium: %v", err)
	}
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, b)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d for one kind of counter, want 1", me.Hand.Size()-hand)
	}
}

func TestTidusCScholarOfNewHorizonsFetchesAPlains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	plains := putOnTopForTest(me, "Plains", "Basic Land — Plains")
	scholar := b12Push(g, me.ID, "Scholar of New Horizons", "Creature — Human Scout", tdScholarOracle, 1, 1)
	tdAddCounter(g, me.ID, scholar, game.CounterPlusOne, 1)
	if err := g.ActivateCatalogAbility(me.ID, scholar, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{scholar}, CounterKind: game.CounterPlusOne,
	}); err != nil {
		t.Fatalf("activate the Scholar: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID); p != nil {
		if err := g.ResolveSearchLibrary(p.ID, me.ID, []uuid.UUID{plains}); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(plains) {
		t.Error("with no opponent ahead on lands the Plains goes to hand")
	}
}

func TestTidusCScholarOfNewHorizonsMayPutThePlainsOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	advanceToMain(t, g)
	b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	plains := putOnTopForTest(me, "Plains", "Basic Land — Plains")
	scholar := b12Push(g, me.ID, "Scholar of New Horizons", "Creature — Human Scout", tdScholarOracle, 1, 1)
	tdAddCounter(g, me.ID, scholar, game.CounterPlusOne, 1)
	if err := g.ActivateCatalogAbility(me.ID, scholar, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{scholar}, CounterKind: game.CounterPlusOne,
	}); err != nil {
		t.Fatalf("activate the Scholar: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID); p != nil {
		if err := g.ResolveSearchLibrary(p.ID, me.ID, []uuid.UUID{plains}); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("not offered the battlefield: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{plains}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(plains) {
		t.Fatal("the Plains did not enter")
	}
	if !b13Tapped(t, g, plains) {
		t.Error("the Plains entered untapped")
	}
}

func TestTidusCYunaGrandSummonerBoostsTheNextCreatureAndRecyclesCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	yuna := b12Push(g, me.ID, "Yuna, Grand Summoner", "Legendary Creature — Human Cleric", tdYunaOracle, 1, 5)
	if err := g.ActivateManaAbility(me.ID, yuna, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	beast := tdCastCreature(t, g, "Beast", 2, 2)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, beast, game.CounterPlusOne); got != 2 {
		t.Fatalf("the next creature entered with %d counters, want 2", got)
	}
	second := tdCastCreature(t, g, "Second", 2, 2)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, second, game.CounterPlusOne); got != 0 {
		t.Errorf("a second creature got %d counters, want 0", got)
	}

	killCreature(t, g, me.ID, beast)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, second)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, second, game.CounterPlusOne); got != 2 {
		t.Errorf("the target got %d counters from a two-counter death, want 2", got)
	}
}

func TestTidusCAlteredEgoCopiesWithXCounters(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	giant := tdCreature(g, opp.ID, "Their Giant", 5, 5)
	ego := tdCastFromHand(t, g, game.Card{Name: "Altered Ego", TypeLine: "Creature — Shapeshifter",
		OracleID: tdAlteredEgoOracle, ManaCost: "{X}{2}{G}{U}"}, game.CastSpellParams{XValue: 3})
	resolveWithCopyChoice(t, g, giant)
	if !g.Battlefield.Contains(ego) {
		t.Fatal("Altered Ego is not on the battlefield")
	}
	if got := countersOn(g, ego, game.CounterPlusOne); got != 3 {
		t.Errorf("it entered with %d counters, want 3", got)
	}
}

func TestTidusCKimahriTapsAndMayBecomeACopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kimahri := b12Push(g, me.ID, "Kimahri, Valiant Guardian", "Legendary Creature — Cat Warrior", tdKimahriOracle, 3, 3)
	giant := tdCreature(g, opp.ID, "Their Giant", 6, 6)
	advanceTo(t, g, game.StepBeginCombat)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, giant)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, giant) {
		t.Error("the target was not tapped")
	}
	if got := countersOn(g, kimahri, game.CounterPlusOne); got != 1 {
		t.Errorf("Kimahri has %d counters, want 1", got)
	}
	c := cardByID(g, kimahri)
	if c.Name != "Kimahri, Valiant Guardian" {
		t.Errorf("the copy is named %q", c.Name)
	}
	if c.CurrentPower() != 7 {
		t.Errorf("the copy's power is %d, want 6 plus the counter", c.CurrentPower())
	}
	if !containsString(effectiveAbilities(t, g, kimahri), "vigilance") {
		t.Error("the copy has no vigilance")
	}
}
