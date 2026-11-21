package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b126_cards_test.go — set GAP, batch 126 (#5948). Each test plays
// the real catalog entry through the engine and asserts what its text says.

const (
	gapB126ButterburOracle   = "ec136f6f-015b-4200-97b6-a2c1bf39e135"
	gapB126ErkenbrandOracle  = "8cac919f-3bb8-49cb-b4e4-4dd0dcd7746c"
	gapB126GimliOracle       = "96ebcd8b-ae3c-47b0-b895-bfc9848713e5"
	gapB126GloinOracle       = "29375a7a-af4e-46aa-8d1c-e8d22d3c0ea7"
	gapB126GrishnakhOracle   = "1f139012-fa71-4f5f-8c57-6847bff95678"
	gapB126SpearmasterOracle = "4da49285-e577-487f-9997-ba49f2b5753e"
	gapB126MoriaOracle       = "9f0027e3-815b-4965-9ef4-3c4814c62cf3"
	gapB126RidersOracle      = "c7031047-5109-482f-b2e9-14800b9279bf"
	gapB126WarbeastOracle    = "7ca52e15-5a55-4d79-b0bf-0e2dc515ae6a"
	gapB126EomerOracle       = "c07340bb-04b6-4a80-9999-727b8ab3534d"
	gapB126ShadowfaxOracle   = "aa630f3d-9443-46fe-a54f-8b3c9fcb872e"
)

func TestGapB126EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB126ButterburOracle:   "Butterbur, Bree Innkeeper",
		gapB126ErkenbrandOracle:  "Erkenbrand, Lord of Westfold",
		gapB126GimliOracle:       "Gimli, Counter of Kills",
		gapB126GloinOracle:       "Glóin, Dwarf Emissary",
		gapB126GrishnakhOracle:   "Grishnákh, Brash Instigator",
		gapB126SpearmasterOracle: "Haradrim Spearmaster",
		gapB126MoriaOracle:       "Moria Marauder",
		gapB126RidersOracle:      "Riders of the Mark",
		gapB126WarbeastOracle:    "Warbeast of Gorgoroth",
		gapB126EomerOracle:       "Éomer of the Riddermark",
		gapB126ShadowfaxOracle:   "Shadowfax, Lord of Horses",
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

// gapB126Main is a fresh table at the active seat's main phase, with that
// seat and the next.
func gapB126Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// gapB126SettleWithPick passes priority until the stack is empty, answering
// `chooser`'s target picks with `ref` and saying `may` to their "you may"
// prompts on the way. Any other prompt stops it, for the test to answer.
func gapB126SettleWithPick(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef, may bool) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if c := effectstest.LatestTriggerPrompt(g, chooser); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, chooser, may); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		if len(g.PendingTriggers) > 0 {
			g.RunStateChecksForTest()
			if effectstest.LatestPickTarget(g, chooser) != nil || effectstest.LatestTriggerPrompt(g, chooser) != nil {
				continue
			}
		}
		if len(g.PendingChoices) > 0 || effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				continue
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack did not settle")
}

// gapB126ChoosableCards is the card list of the open choose-cards prompt.
func gapB126ChoosableCards(g *game.Game, who uuid.UUID) []uuid.UUID {
	if c := effectstest.ChooseCardsChoiceFor(g, who); c != nil {
		return c.ChooseCards
	}
	return nil
}

func gapB126Ref(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func gapB126Foods(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.HasSubtype("Food") {
			n++
		}
	}
	return n
}

func gapB126Die(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
}

func gapB126Library(p *game.Player, cards ...game.Card) {
	for i := len(cards) - 1; i >= 0; i-- {
		cards[i].InstanceID = uuid.New()
		cards[i].Owner, cards[i].Controller = p.ID, p.ID
		p.Library.PushTop(cards[i])
	}
}

// --- Butterbur ---------------------------------------------------------------

func TestGapB126ButterburMakesAFoodAtYourEndStepOnlyWithoutOne(t *testing.T) {
	g, me, _ := gapB126Main(t)
	effectstest.B12Push(g, me.ID, "Butterbur, Bree Innkeeper", "Legendary Creature — Human Peasant", gapB126ButterburOracle, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB126Foods(g, me.ID); got != 1 {
		t.Fatalf("Foods = %d, want 1 at the end step with none", got)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB126Foods(g, me.ID); got != 1 {
		t.Errorf("Foods = %d, want still 1: Butterbur does not make a second one", got)
	}
}

// --- Erkenbrand --------------------------------------------------------------

func TestGapB126ErkenbrandPumpsEveryCreatureWhenAHumanEnters(t *testing.T) {
	g, me, opp := gapB126Main(t)
	effectstest.B12Push(g, me.ID, "Erkenbrand, Lord of Westfold", "Legendary Creature — Human Soldier", gapB126ErkenbrandOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Peasant", "Creature — Human Peasant", "", 1, 1)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 3 {
		t.Errorf("my Bear has %d power, want 3", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("their Bear has %d power, want 2", p)
	}
}

func TestGapB126ErkenbrandIgnoresANonHuman(t *testing.T) {
	g, me, _ := gapB126Main(t)
	effectstest.B12Push(g, me.ID, "Erkenbrand, Lord of Westfold", "Legendary Creature — Human Soldier", gapB126ErkenbrandOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Wolf", "Creature — Wolf", "", 1, 1)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("my Bear has %d power, want 2: a Wolf is not a Human", p)
	}
}

// --- Gimli, Counter of Kills ---------------------------------------------------

func TestGapB126GimliPingsTheDeadCreaturesController(t *testing.T) {
	g, me, opp := gapB126Main(t)
	effectstest.B12Push(g, me.ID, "Gimli, Counter of Kills", "Legendary Creature — Dwarf Warrior", gapB126GimliOracle, 4, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	oppLife, myLife := opp.Life, me.Life
	gapB126Die(t, g, mine)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife || me.Life != myLife {
		t.Errorf("my own creature dying changed life (me %d -> %d, opp %d -> %d)", myLife, me.Life, oppLife, opp.Life)
	}
	gapB126Die(t, g, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife {
		t.Errorf("life after an opponent's creature died: me %d -> %d, opp %d -> %d; want only the opponent -1", myLife, me.Life, oppLife, opp.Life)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, g.Battlefield.Cards[0].InstanceID, "trample") && !gapB126HasTrample(t, g) {
		t.Error("Gimli lacks trample")
	}
}

func gapB126HasTrample(t *testing.T, g *game.Game) bool {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Gimli, Counter of Kills" {
			return effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "trample")
		}
	}
	return false
}

// --- Glóin ---------------------------------------------------------------------

func TestGapB126GloinMakesATreasureOnTheFirstHistoricCastEachTurn(t *testing.T) {
	g, me, _ := gapB126Main(t)
	effectstest.B12Push(g, me.ID, "Glóin, Dwarf Emissary", "Legendary Creature — Dwarf Advisor", gapB126GloinOracle, 3, 3)
	treasures := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.HasSubtype("Treasure") {
				n++
			}
		}
		return n
	}
	effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 1, 1)
	if treasures() != 0 {
		t.Error("a nonhistoric spell made a Treasure")
	}
	effectstest.CastHoldCreature(t, g, "Hero", "Legendary Creature — Human", "", 1, 1)
	if treasures() != 1 {
		t.Fatalf("Treasures = %d after a legendary spell, want 1", treasures())
	}
	effectstest.CastHoldCreature(t, g, "Hero Two", "Legendary Creature — Human", "", 1, 1)
	if treasures() != 1 {
		t.Errorf("Treasures = %d after a second historic spell, want still 1 (once each turn)", treasures())
	}
}

func TestGapB126GloinGoadsByTappingAndSacrificingATreasure(t *testing.T) {
	g, me, opp := gapB126Main(t)
	gloin := effectstest.B12Push(g, me.ID, "Glóin, Dwarf Emissary", "Legendary Creature — Dwarf Advisor", gapB126GloinOracle, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	if err := effectstest.ConB1Activate(t, g, me, gloin, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Fatal("activated with no Treasure to sacrifice")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	var treasure uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.HasSubtype("Treasure") {
			treasure = c.InstanceID
		}
	}
	if err := effectstest.ConB1Activate(t, g, me, gloin, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim), SacrificeIDs: []uuid.UUID{treasure}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(treasure) {
		t.Error("the Treasure was not sacrificed")
	}
	if !effectstest.HmlTapped(t, g, gloin) {
		t.Error("Glóin did not tap")
	}
	c := effectstest.GaLayered(t, g, victim)
	if len(c.Goads) == 0 {
		t.Error("the target was not goaded")
	}
}

// --- Grishnákh -------------------------------------------------------------------

func TestGapB126GrishnakhStealsASmallNonlegendaryCreatureUntilEndOfTurn(t *testing.T) {
	g, me, opp := gapB126Main(t)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	legend := effectstest.B12Creature(g, opp.ID, "Legend", "Legendary Creature — Human", 1, 1)
	effectstest.CastHoldCreature(t, g, "Grishnákh, Brash Instigator", "Legendary Creature — Goblin Soldier", gapB126GrishnakhOracle, 1, 1)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB126Ref(big)); err == nil {
		t.Error("a creature with power above the Army's was accepted")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB126Ref(legend)); err == nil {
		t.Error("a legendary creature was accepted")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB126Ref(small)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, small)
	if c.Controller != me.ID {
		t.Error("the creature was not stolen")
	}
	if !effectstest.HasEffectiveKeyword(t, g, small, "haste") {
		t.Error("the stolen creature lacks haste")
	}
	if army := effectstest.ArmyOf(t, g, me.ID); effectstest.CounterCount(g, army.InstanceID, game.CounterPlusOne) != 2 {
		t.Error("the Army did not get two +1/+1 counters")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if c := effectstest.GaLayered(t, g, small); c.Controller != opp.ID {
		t.Error("the creature was not returned at end of turn")
	}
}

// --- Haradrim Spearmaster ----------------------------------------------------------

func TestGapB126SpearmasterPumpsAnotherCreatureAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spear := effectstest.B12Push(g, me.ID, "Haradrim Spearmaster", "Creature — Human Warrior", gapB126SpearmasterOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB126Ref(spear)); err == nil {
		t.Error("the Spearmaster targeted itself")
	}
	gapB126SettleWithPick(t, g, me.ID, gapB126Ref(bear), false)
	if pw, _ := effectstest.EffectivePTOf(t, g, bear); pw != 3 {
		t.Errorf("the Bear has %d power, want 3", pw)
	}
	if pw, _ := effectstest.EffectivePTOf(t, g, spear); pw != 2 {
		t.Errorf("the Spearmaster has %d power, want 2", pw)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, spear, "reach") {
		t.Error("the Spearmaster lacks reach")
	}
}

// --- Moria Marauder ------------------------------------------------------------------

func TestGapB126MoriaMarauderExilesTheTopCardWhenAGoblinOrOrcConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marauder := effectstest.B12Push(g, me.ID, "Moria Marauder", "Creature — Goblin Warrior", gapB126MoriaOracle, 1, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, marauder, "double strike") {
		t.Fatal("the Marauder lacks double strike")
	}
	gapB126Library(me, game.Card{Name: "Top One"}, game.Card{Name: "Top Two"}, game.Card{Name: "Top Three"})
	library := len(me.Library.Cards)
	effectstest.DeclareAttack(t, g, opp.ID, marauder)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if exiled := library - len(me.Library.Cards); exiled != 2 {
		t.Errorf("%d cards exiled, want 2 (double strike deals damage twice)", exiled)
	}
}

func TestGapB126MoriaMarauderIgnoresOtherTribes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Moria Marauder", "Creature — Goblin Warrior", gapB126MoriaOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	gapB126Library(me, game.Card{Name: "Top One"}, game.Card{Name: "Top Two"})
	library := len(me.Library.Cards)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Library.Cards) != library {
		t.Error("a Bear connecting exiled a card")
	}
}

// --- Riders of the Mark ----------------------------------------------------------------

func TestGapB126RidersHasAffinityForHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, gapB126RidersOracle, "Creature — Human Knight", "{6}{R}"); got != 7 {
		t.Errorf("no Humans: costs %d, want 7", got)
	}
	effectstest.B12Creature(g, me.ID, "H1", "Creature — Human", 1, 1)
	effectstest.B12Creature(g, me.ID, "H2", "Creature — Human Soldier", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	if got := effectstest.SelfPricedMV(t, g, me, gapB126RidersOracle, "Creature — Human Knight", "{6}{R}"); got != 5 {
		t.Errorf("two Humans: costs %d, want 5", got)
	}
}

func TestGapB126RidersReturnsAfterAttackingAndMakesTokensEqualToToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	riders := effectstest.B12Push(g, me.ID, "Riders of the Mark", "Creature — Human Knight", gapB126RidersOracle, 7, 4)
	for _, kw := range []string{"trample", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, riders, kw) {
			t.Errorf("Riders lack %s", kw)
		}
	}
	effectstest.DeclareAttack(t, g, opp.ID, riders)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(riders) {
		t.Fatal("Riders were not returned to hand")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Human Soldier"); n != 4 {
		t.Errorf("%d Human Soldier tokens, want 4 (Riders' toughness)", n)
	}
}

func TestGapB126RidersStayIfTheyDidNotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	riders := effectstest.B12Push(g, me.ID, "Riders of the Mark", "Creature — Human Knight", gapB126RidersOracle, 7, 4)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(riders) {
		t.Error("Riders returned without having attacked")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Human Soldier"); n != 0 {
		t.Errorf("%d tokens, want 0", n)
	}
}

// --- Warbeast of Gorgoroth ----------------------------------------------------------------

func TestGapB126WarbeastAmassesWhenAPowerFourCreatureOfYoursDies(t *testing.T) {
	g, me, opp := gapB126Main(t)
	beast := effectstest.B12Push(g, me.ID, "Warbeast of Gorgoroth", "Creature — Beast", gapB126WarbeastOracle, 5, 4)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 6, 6)
	armyCount := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && game.IsArmy(c) {
				n += c.Counters[game.CounterPlusOne]
			}
		}
		return n
	}
	gapB126Die(t, g, small)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	gapB126Die(t, g, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if armyCount() != 0 {
		t.Fatalf("an Army has %d counters after a small creature and an opponent's creature died, want none", armyCount())
	}
	gapB126Die(t, g, big)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if armyCount() != 2 {
		t.Errorf("Army counters = %d after a power-4 creature died, want 2", armyCount())
	}
	gapB126Die(t, g, beast)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if armyCount() != 4 {
		t.Errorf("Army counters = %d after Warbeast itself died, want 4", armyCount())
	}
}

// --- Éomer of the Riddermark -----------------------------------------------------------------

func TestGapB126EomerMakesATokenOnlyWithTheGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eomer := effectstest.B12Push(g, me.ID, "Éomer of the Riddermark", "Legendary Creature — Human Knight", gapB126EomerOracle, 5, 4)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Wurm", 9, 9)
	effectstest.DeclareAttack(t, g, opp.ID, eomer)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Human Soldier"); n != 0 {
		t.Errorf("%d tokens with a bigger creature on the table, want 0", n)
	}
	_ = big
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	eomer2 := effectstest.B12Push(g2, me2.ID, "Éomer of the Riddermark", "Legendary Creature — Human Knight", gapB126EomerOracle, 5, 4)
	effectstest.B12Creature(g2, opp2.ID, "Tied", "Creature — Ogre", 5, 5)
	if !effectstest.HasKeywordOnBattlefield(t, g2, eomer2, "haste") {
		t.Fatal("Éomer lacks haste")
	}
	effectstest.DeclareAttack(t, g2, opp2.ID, eomer2)
	effectstest.PassPriorityAroundTable(t, g2)
	if n, _ := effectstest.CountTokensNamed(g2, me2.ID, "Human Soldier"); n != 1 {
		t.Errorf("%d tokens when tied for the greatest power, want 1", n)
	}
}

// --- Shadowfax ---------------------------------------------------------------------------------

func TestGapB126ShadowfaxGivesHorsesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Shadowfax, Lord of Horses", "Legendary Creature — Horse", gapB126ShadowfaxOracle, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mare", "Creature — Horse", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Horse", "Creature — Horse", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("my Horse lacks haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "haste") || effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("a non-Horse or an opponent's Horse has haste")
	}
}

func TestGapB126ShadowfaxPutsALesserPowerCreatureFromHandAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shadowfax := effectstest.B12Push(g, me.ID, "Shadowfax, Lord of Horses", "Legendary Creature — Horse", gapB126ShadowfaxOracle, 4, 4)
	small, equal := uuid.New(), uuid.New()
	for _, c := range []game.Card{
		{InstanceID: small, Name: "Small", TypeLine: "Creature — Bear", Power: 3, Toughness: 3},
		{InstanceID: equal, Name: "Equal", TypeLine: "Creature — Bear", Power: 4, Toughness: 4},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Hand.PushTop(c)
	}
	effectstest.DeclareAttack(t, g, opp.ID, shadowfax)
	effectstest.PassPriorityAroundTable(t, g)
	offered := gapB126ChoosableCards(g, me.ID)
	for _, id := range offered {
		if id == equal {
			t.Error("a creature with equal power was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, small)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID); p != nil {
		if err := g.ResolveOptionPick(p.ID, me.ID, 0); err != nil {
			t.Fatalf("choose what it attacks: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Fatal("the lesser-power creature did not enter")
	}
	c := effectstest.GaLayered(t, g, small)
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("the creature entered tapped=%v attacking-target=%v, want tapped and attacking", c.Tapped, c.AttackingTarget)
	}
	if !me.Hand.Contains(equal) {
		t.Error("the equal-power creature left the hand")
	}
}
