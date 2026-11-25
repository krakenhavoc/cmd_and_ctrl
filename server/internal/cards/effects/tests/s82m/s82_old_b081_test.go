package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 081 (#5507): Commander Legends, enters and dies triggers.

const (
	oldB081PromiseOracle  = "26c22894-10e5-48f1-a32a-a7f1f3f7e12e"
	oldB081DenizenOracle  = "34312179-f3f9-47ca-bdf4-5abbd493876e"
	oldB081SphinxOracle   = "4b4f268d-34ae-416d-a412-66f4fe37659c"
	oldB081DoomOracle     = "5fba553e-81b9-419b-aa20-b2a7cfcd042b"
	oldB081DruidOracle    = "6c2e6447-c32b-4b92-a3e6-4d770e3dd250"
	oldB081SlinnOracle    = "753d98d3-d534-42c1-b728-a4f14c7f29ec"
	oldB081HydraOracle    = "9b157b9a-92da-4054-8441-04fb4e415d77"
	oldB081HeraldOracle   = "ea9b2e3d-2de4-4cde-b233-6d0ba5797c20"
	oldB081CastTypeLine   = "Creature — Bear"
	oldB081CommanderTypes = "Legendary Creature — Bear"
)

func TestOldB081EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB081PromiseOracle: "Promise of Tomorrow",
		oldB081DenizenOracle: "Court Street Denizen",
		oldB081SphinxOracle:  "Sphinx of Uthuun",
		oldB081DoomOracle:    "Elvish Doomsayer",
		oldB081DruidOracle:   "Rootweaver Druid",
		oldB081SlinnOracle:   "Slinn Voda, the Rising Deep",
		oldB081HydraOracle:   "Stumpsquall Hydra",
		oldB081HeraldOracle:  "Brass Herald",
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

// oldB081Cast casts a creature spell with the given colours from the active
// seat's hand.
func oldB081Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, colors []string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.HandCardFull(active, name, typeLine, cost, oracle, colors)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func TestOldB081ElvishDoomsayerMakesEachOpponentDiscardWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	doom := effectstest.B12Push(g, me.ID, "Elvish Doomsayer", "Creature — Elf Shaman", oldB081DoomOracle, 1, 1)
	for _, p := range g.Seats {
		p.Hand.Cards = nil
		effectstest.PushCatalogHandCard(p, "Only", "Instant", "")
	}
	effectstest.KillCreature(t, g, me.ID, doom)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		effectstest.DiscardFromHand(t, g, p.ID)
	}
	if me.Hand.Size() != 1 {
		t.Errorf("the controller's hand is %d, want 1 untouched", me.Hand.Size())
	}
	for i, p := range g.Seats[1:] {
		if p.Hand.Size() != 0 {
			t.Errorf("opponent %d still holds %d cards", i+1, p.Hand.Size())
		}
	}
}

func TestOldB081CourtStreetDenizenTapsForAnotherWhiteCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Court Street Denizen", "Creature — Human Soldier", oldB081DenizenOracle, 2, 2, "W")
	mine := effectstest.B21Push(g, me.ID, "Mine", oldB081CastTypeLine, "", 2, 2, "G")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB081CastTypeLine, 2, 2)
	// A green creature entering does nothing.
	oldB081Cast(t, g, "Green Bear", oldB081CastTypeLine, "{G}", "", []string{"G"})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a green creature triggered the Denizen")
	}
	// A white one does, and only opponents' creatures are offered.
	white := oldB081Cast(t, g, "White Bear", oldB081CastTypeLine, "{W}", "", []string{"W"})
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, theirs) || effectstest.HasID(p.PickTargetCards, mine) || effectstest.HasID(p.PickTargetCards, white) {
		t.Fatalf("target creature an opponent controls: offered %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, theirs); c == nil || !c.Tapped {
		t.Error("the opponent's creature was not tapped")
	}
}

func TestOldB081CourtStreetDenizenIgnoresAnOpponentsWhiteCreatureAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, g.Seats[2].ID, "Theirs", oldB081CastTypeLine, 2, 2)
	oldB081Cast(t, g, "Court Street Denizen", "Creature — Human Soldier", "{2}{W}", oldB081DenizenOracle, []string{"W"})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the Denizen triggered on its own entering")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	oldB081Cast(t, g, "Their White Bear", oldB081CastTypeLine, "{W}", "", []string{"W"})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's white creature triggered the Denizen")
	}
}

func TestOldB081SlinnVodaBouncesOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.B12Creature(g, me.ID, "Bear", oldB081CastTypeLine, 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB081CastTypeLine, 2, 2)
		fish := effectstest.B12Creature(g, opp.ID, "Fish", "Creature — Merfolk", 1, 1)
		kraken := effectstest.B12Creature(g, opp.ID, "Kraken", "Creature — Kraken", 5, 5)
		eel := effectstest.B12Creature(g, opp.ID, "Eel", "Creature — Serpent", 3, 3)
		effectstest.AdvanceToMain(t, g)
		var opt []int
		if kicked {
			opt = []int{0}
		}
		id, err := effectstest.CastWithOptionalCosts(t, g, "Slinn Voda, the Rising Deep", "Legendary Creature — Leviathan", oldB081SlinnOracle, nil, opt, nil)
		if err != nil {
			t.Fatalf("kicked=%v: CastSpell: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(id) {
			t.Fatalf("kicked=%v: Slinn Voda is not on the battlefield", kicked)
		}
		for _, stays := range []uuid.UUID{fish, kraken, eel} {
			if !g.Battlefield.Contains(stays) {
				t.Errorf("kicked=%v: a protected creature was bounced", kicked)
			}
		}
		for _, bounced := range []uuid.UUID{bear, theirs} {
			if g.Battlefield.Contains(bounced) == kicked {
				t.Errorf("kicked=%v: creature on battlefield = %v", kicked, g.Battlefield.Contains(bounced))
			}
		}
	}
}

func TestOldB081StumpsquallHydraDistributesAmongItselfAndCommanders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Commander", oldB081CommanderTypes, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Commander", oldB081CommanderTypes, 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", oldB081CastTypeLine, 2, 2)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == mine || g.Battlefield.Cards[i].InstanceID == theirs {
			g.Battlefield.Cards[i].IsCommander = true
		}
	}
	hydra := effectstest.CastXSpell(t, g, "Stumpsquall Hydra", "Creature — Hydra", oldB081HydraOracle, "{X}{G}{G}{G}", 5, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Commanders are asked in turn, the Hydra keeps the rest.
	for _, step := range []int{2, 1} {
		effectstest.AnswerNumber(t, g, me.ID, step)
		effectstest.PassPriorityAroundTable(t, g)
	}
	counters := map[string]int{
		"hydra":  effectstest.CounterCount(g, hydra, game.CounterPlusOne),
		"mine":   effectstest.CounterCount(g, mine, game.CounterPlusOne),
		"theirs": effectstest.CounterCount(g, theirs, game.CounterPlusOne),
		"plain":  effectstest.CounterCount(g, plain, game.CounterPlusOne),
	}
	total := counters["hydra"] + counters["mine"] + counters["theirs"]
	if total != 5 || counters["hydra"] < 1 || counters["plain"] != 0 {
		t.Errorf("counters = %v, want five in all with at least one on the Hydra and none on a non-commander", counters)
	}
	if counters["mine"]+counters["theirs"] != 3 {
		t.Errorf("commanders got %d, want the 3 chosen: %v", counters["mine"]+counters["theirs"], counters)
	}
}

func TestOldB081StumpsquallHydraAtXZeroPutsNothingAndWithNoCommandersTakesAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	hydra := effectstest.CastXSpell(t, g, "Stumpsquall Hydra", "Creature — Hydra", oldB081HydraOracle, "{X}{G}{G}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hydra, game.CounterPlusOne); got != 3 {
		t.Errorf("a Hydra with no commander in play has %d counters, want all 3", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	zero := effectstest.CastXSpell(t, g2, "Stumpsquall Hydra", "Creature — Hydra", oldB081HydraOracle, "{X}{G}{G}{G}", 0, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CounterCount(g2, zero, game.CounterPlusOne); got != 0 {
		t.Errorf("an X = 0 Hydra has %d counters", got)
	}
}

// oldB081Piles answers Fact or Fiction's two prompts: the opponent separates
// pile one from the rest and the controller takes the option at pick.
func oldB081Piles(t *testing.T, g *game.Game, me, opp *game.Player, pileOne []uuid.UUID, pick int) {
	t.Helper()
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	}
	split := effectstest.LatestRevealPickFor(g, opp.ID)
	if split == nil {
		t.Fatalf("the opponent was not asked to separate: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, opp.ID, pileOne); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	p := effectstest.LatestOptionPickFor(g, me.ID)
	if p == nil {
		t.Fatalf("no pile pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveOptionPick(p.ID, me.ID, pick); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
}

func TestOldB081SphinxOfUthuunRevealsFiveAndSplitsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lib := effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Forest(), effectstest.C2b15Spell(), effectstest.C2b15Bear(), effectstest.C2b15Forest(), effectstest.C2b15Spell())
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.EnterFromHand(t, g, me.ID, "Sphinx of Uthuun", "Creature — Sphinx", oldB081SphinxOracle)
	effectstest.PassPriorityAroundTable(t, g)
	split := effectstest.LatestOptionPickFor(g, me.ID)
	if split == nil && effectstest.LatestRevealPickFor(g, opp.ID) == nil {
		t.Fatalf("no prompt after the enters trigger: %+v", g.PendingChoices)
	}
	oldB081Piles(t, g, me, opp, []uuid.UUID{lib[0], lib[1]}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	inHand, inYard := 0, 0
	for _, id := range lib[:5] {
		switch {
		case me.Hand.Contains(id):
			inHand++
		case me.Graveyard.Contains(id):
			inYard++
		}
	}
	if inHand+inYard != 5 || inHand == 0 || inYard == 0 {
		t.Errorf("hand %d, graveyard %d of the five revealed, want a split of all five", inHand, inYard)
	}
	if me.Hand.Size() != hand+inHand {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+inHand)
	}
	if !me.Library.Contains(lib[5]) {
		t.Error("the sixth card left the library")
	}
}

func TestOldB081BrassHeraldNamesATypeTakesItsCreaturesAndPumpsEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB081CastTypeLine, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	lib := effectstest.DftMPLibraryTop(me,
		game.Card{Name: "Bear A", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf", Power: 2, Toughness: 2},
		game.Card{Name: "Bear Spell", TypeLine: "Instant — Bear"},
		game.Card{Name: "Bear B", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Deep Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
	)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	herald := effectstest.EnterFromHand(t, g, me.ID, "Brass Herald", "Artifact Creature — Golem", oldB081HeraldOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Bear")
	effectstest.PassPriorityAroundTable(t, g)
	// Two Bear creature cards among the top four: both come to hand.
	for _, id := range []uuid.UUID{lib[0], lib[3]} {
		if !me.Hand.Contains(id) {
			t.Errorf("a revealed Bear creature card did not reach the hand")
		}
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+2)
	}
	// The rest go to the bottom in an order the controller chooses.
	if !me.Library.Contains(lib[4]) || me.Hand.Contains(lib[1]) || me.Hand.Contains(lib[2]) {
		t.Error("a non-matching card was taken, or the fifth card left the library")
	}
	// Every Bear gets +1/+1, whoever controls it, and nothing else does.
	if p := effectstest.EffectivePower(t, g, theirs); p != 3 {
		t.Errorf("the opponent's Bear has power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, other); p != 2 {
		t.Errorf("a Wolf has power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, herald); p != 2 {
		t.Errorf("the Golem Herald has power %d, want its printed 2", p)
	}
}

func TestOldB081RootweaverDruidFirstLandGoesToTheDruidsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats[1:] {
		p.Library.Cards = nil
	}
	var lands []uuid.UUID
	for i := 0; i < 4; i++ {
		lands = append(lands, effectstest.DftMPLibraryTop(opp, effectstest.C2b15Forest())...)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.EnterFromHand(t, g, me.ID, "Rootweaver Druid", "Creature — Elf Druid", oldB081DruidOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, opp.ID) == nil {
		t.Fatalf("the opponent was not offered a search: %+v", g.PendingChoices)
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, lands[0], lands[1], lands[2])
	effectstest.PassPriorityAroundTable(t, g)
	var mine, theirs int
	for _, id := range lands[:3] {
		c := effectstest.FindBattlefieldCardByID(g, id)
		if c == nil {
			t.Fatalf("a chosen land did not enter the battlefield")
		}
		if !c.Tapped {
			t.Errorf("a land entered untapped")
		}
		switch c.Controller {
		case me.ID:
			mine++
		case opp.ID:
			theirs++
		}
	}
	if mine != 1 || theirs != 2 {
		t.Errorf("under the Druid's controller %d, under the searcher %d; want 1 and 2", mine, theirs)
	}
	if g.Battlefield.Contains(lands[3]) {
		t.Error("a fourth land entered")
	}
}

func TestOldB081RootweaverDruidAnOpponentMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.DftMPLibraryTop(opp, effectstest.C2b15Forest())[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.EnterFromHand(t, g, me.ID, "Rootweaver Druid", "Creature — Elf Druid", oldB081DruidOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchFailToFind(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("a land entered although the opponent found nothing")
	}
}

func TestOldB081PromiseOfTomorrowExilesYourDeadCreaturesAndReturnsThemWhenYouHaveNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	promise := effectstest.PushCatalogPermanent(g, me.ID, "Promise of Tomorrow", "Enchantment", oldB081PromiseOracle, false)
	a := effectstest.B12Push(g, me.ID, "Court Street Denizen", "Creature — Human Soldier", oldB081DenizenOracle, 2, 2)
	b := effectstest.B12Push(g, me.ID, "Court Street Denizen", "Creature — Human Soldier", oldB081DenizenOracle, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB081CastTypeLine, 2, 2)
	effectstest.KillCreature(t, g, me.ID, a)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) {
		t.Fatal("the dying creature was not exiled")
	}
	// An opponent's creature dying is not "you control".
	effectstest.KillCreature(t, g, opp.ID, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(theirs) || !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's creature was exiled")
	}
	// With a creature left, the end step does nothing.
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassThroughEndStepTriggers(t, g)
	if !g.Battlefield.Contains(promise) || !g.Exile.Contains(a) {
		t.Fatal("the Promise acted while its controller still had a creature")
	}
	effectstest.KillCreature(t, g, me.ID, b)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassThroughEndStepTriggers(t, g)
	if g.Battlefield.Contains(promise) {
		t.Error("the Promise was not sacrificed")
	}
	// A card returned from exile is a new object, so count them by name.
	back := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Court Street Denizen" && c.Controller == me.ID {
			back++
		}
	}
	if back != 2 {
		t.Errorf("%d exiled creatures came back under the Promise's controller, want 2", back)
	}
	if g.Exile.Contains(a) || g.Exile.Contains(b) {
		t.Error("an exiled creature is still in exile")
	}
}
