package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set AVR, batch 435 (#5570): Avacyn Restored, pump and combat tricks,
// instants and sorceries.

const (
	avr435JointAssault    = "0f719e56-5203-4138-bbe9-6a37b8ac2512"
	avr435NarstadScrapper = "16acd858-dba6-401e-bbfb-cabdd6d0120e"
	avr435BuildersBlss    = "33a64de1-dc78-43d3-b89e-b4b13049e174"
	avr435WildDefiance    = "76703951-f42f-4e1d-af51-e1e06a89bd5e"
	avr435WildwoodGeist   = "9fb4aa2d-6414-4176-9070-e5230e100c6a"
	avr435PrimalSurge     = "44b12ccc-81cb-4ec1-891e-4bf706e53aa4"
	avr435ThatcherRevolt  = "83cdff93-531d-4148-bff7-5312b3b53bf0"
	avr435LairDelve       = "8d7ce7b5-2e56-4060-9406-f700ce1fbb7d"
	avr435RiteOfRuin      = "d0dd0586-d89e-40e9-863d-9c390bee117a"
	avr435RushOfBlood     = "f0bacbf1-3c35-436e-b0bf-10613b60b726"
)

func avr435Push(g *game.Game, owner uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestAvrB435RegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avr435JointAssault: "Joint Assault", avr435NarstadScrapper: "Narstad Scrapper",
		avr435BuildersBlss: "Builder's Blessing", avr435WildDefiance: "Wild Defiance",
		avr435WildwoodGeist: "Wildwood Geist", avr435PrimalSurge: "Primal Surge",
		avr435ThatcherRevolt: "Thatcher Revolt", avr435LairDelve: "Lair Delve",
		avr435RiteOfRuin: "Rite of Ruin", avr435RushOfBlood: "Rush of Blood",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestAvrB435BuildersBlessingOnlyUntappedCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Builder's Blessing", "Enchantment", avr435BuildersBlss, false)
	mine := avr435Push(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	theirs := avr435Push(g, opp.ID, "Their Bear", effectstest.TestCreatureTypeLine, 2, 2)
	if _, tough := effectstest.EffectivePTOf(t, g, mine); tough != 4 {
		t.Errorf("an untapped creature has toughness %d, want 4", tough)
	}
	if _, tough := effectstest.EffectivePTOf(t, g, theirs); tough != 2 {
		t.Errorf("an opponent's creature has toughness %d, want 2", tough)
	}
	if err := g.TapCard(mine, true); err != nil {
		t.Fatal(err)
	}
	if _, tough := effectstest.EffectivePTOf(t, g, mine); tough != 2 {
		t.Errorf("a tapped creature has toughness %d, want 2", tough)
	}
}

func TestAvrB435NarstadScrapperPumpsForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scrapper := effectstest.PushCatalogPermanent(g, me.ID, "Narstad Scrapper", "Artifact Creature — Construct", avr435NarstadScrapper, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	before, _ := effectstest.EffectivePTOf(t, g, scrapper)
	if err := g.ActivateCatalogAbility(me.ID, scrapper, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if after, _ := effectstest.EffectivePTOf(t, g, scrapper); after != before+1 {
		t.Errorf("power %d, want %d", after, before+1)
	}
}

func TestAvrB435WildwoodGeistGrowsOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	geist := effectstest.PushCatalogPermanent(g, me.ID, "Wildwood Geist", "Creature — Spirit", avr435WildwoodGeist, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, geist)
		c.Power, c.Toughness = 3, 3
	})
	if p, tough := effectstest.EffectivePTOf(t, g, geist); p != 5 || tough != 5 {
		t.Fatalf("on its controller's turn it is %d/%d, want 5/5", p, tough)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if p, tough := effectstest.EffectivePTOf(t, g, geist); p != 3 || tough != 3 {
		t.Errorf("on an opponent's turn it is %d/%d, want 3/3", p, tough)
	}
}

func TestAvrB435RushOfBloodDoublesPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := avr435Push(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 3, 2)
	effectstest.CastCatalogSpell(t, g, "Rush of Blood", "Instant", avr435RushOfBlood,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 6 || tough != 2 {
		t.Errorf("creature is %d/%d, want 6/2", p, tough)
	}
}

func TestAvrB435ThatcherRevoltMakesThreeHastyHumansThatDieAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Thatcher Revolt", "Sorcery", avr435ThatcherRevolt, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Human"); n != 3 {
		t.Fatalf("Human tokens = %d, want 3", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Human" && !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "haste") {
			t.Error("a Human token lacks haste")
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Human"); n != 0 {
		t.Errorf("Human tokens at the end step = %d, want 0 (sacrificed)", n)
	}
}

func TestAvrB435JointAssaultPumpsTheSoulbondPartnerToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := avr435Push(g, me.ID, "A", effectstest.TestCreatureTypeLine, 2, 2)
	b := avr435Push(g, me.ID, "B", effectstest.TestCreatureTypeLine, 1, 1)
	c := avr435Push(g, me.ID, "C", effectstest.TestCreatureTypeLine, 1, 1)
	g.WithWriteLock(func() {
		if !g.PairForEffect(me.ID, a, b) {
			t.Fatal("could not pair the creatures")
		}
	})
	effectstest.CastCatalogSpell(t, g, "Joint Assault", "Instant", avr435JointAssault,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, a); p != 4 {
		t.Errorf("target power %d, want 4", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, b); p != 3 {
		t.Errorf("paired creature power %d, want 3", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, c); p != 1 {
		t.Errorf("unpaired creature power %d, want 1", p)
	}
}

func TestAvrB435JointAssaultUnpairedOnlyPumpsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := avr435Push(g, me.ID, "A", effectstest.TestCreatureTypeLine, 2, 2)
	b := avr435Push(g, me.ID, "B", effectstest.TestCreatureTypeLine, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Joint Assault", "Instant", avr435JointAssault,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, a); p != 4 || tough != 4 {
		t.Errorf("target is %d/%d, want 4/4", p, tough)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, b); p != 1 {
		t.Errorf("another creature power %d, want 1", p)
	}
}

func TestAvrB435WildDefianceSavesACreatureTargetedByAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Wild Defiance", "Enchantment", avr435WildDefiance, false)
	bear := avr435Push(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the trigger's +3/+3 should resolve before the Bolt and let the creature live")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("creature is %d/%d, want 5/5 until end of turn", p, tough)
	}
}

func TestAvrB435WildDefianceIgnoresAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Wild Defiance", "Enchantment", avr435WildDefiance, false)
	theirs := avr435Push(g, opp.ID, "Their Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("an opponent's creature is not pumped and dies to the Bolt")
	}
}

func TestAvrB435LairDelveTakesCreaturesAndLandsAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibraryTop(me, "Filler", "Sorcery", "")
	spell := effectstest.SeedLibraryTop(me, "Some Instant", "Instant", "{U}")
	land := effectstest.SeedLibraryTop(me, "Some Forest", "Basic Land — Forest", "")
	hand, libSize := me.Hand.Size(), me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Lair Delve", "Sorcery", avr435LairDelve, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the revealed land should be in hand")
	}
	if me.Hand.Contains(spell) {
		t.Error("the revealed instant should not be in hand")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
	if me.Library.Size() != libSize-1 {
		t.Errorf("library = %d, want %d", me.Library.Size(), libSize-1)
	}
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[0].InstanceID != spell {
		t.Error("the instant should be on the bottom of the library")
	}
}

func TestAvrB435PrimalSurgeKeepsGoingWhileYouSayYes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	instant := effectstest.SeedLibraryTop(me, "Some Instant", "Instant", "{U}")
	land2 := effectstest.SeedLibraryTop(me, "Second Forest", "Basic Land — Forest", "")
	land1 := effectstest.SeedLibraryTop(me, "First Forest", "Basic Land — Forest", "")
	effectstest.CastCatalogSpell(t, g, "Primal Surge", "Sorcery", avr435PrimalSurge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	_, _ = land1, land2
	if effectstest.CountBattlefieldByName(g, "First Forest") != 1 || effectstest.CountBattlefieldByName(g, "Second Forest") != 1 {
		t.Fatal("both lands should be on the battlefield")
	}
	if g.Battlefield.Contains(instant) {
		t.Error("an instant can't be put onto the battlefield")
	}
	if z := g.FindCardZoneForEffect(instant); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the instant should have been exiled, it is in %v", z)
	}
}

func TestAvrB435PrimalSurgeStopsOnADecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land2 := effectstest.SeedLibraryTop(me, "Second Forest", "Basic Land — Forest", "")
	land1 := effectstest.SeedLibraryTop(me, "First Forest", "Basic Land — Forest", "")
	effectstest.CastCatalogSpell(t, g, "Primal Surge", "Sorcery", avr435PrimalSurge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldByName(g, "First Forest") != 0 {
		t.Error("the declined land should stay in exile")
	}
	if z := g.FindCardZoneForEffect(land1); z == nil || z.Kind != game.ZoneExile {
		t.Error("the declined land should be in exile")
	}
	if effectstest.CountBattlefieldByName(g, "Second Forest") != 0 || !me.Library.Contains(land2) {
		t.Error("the process stops: the second land is still in the library")
	}
}

func avr435RiteBoard(g *game.Game, owner uuid.UUID) {
	for i := 0; i < 4; i++ {
		avr435Push(g, owner, "Artifact", "Artifact", 0, 0)
		avr435Push(g, owner, "Forest", "Basic Land — Forest", 0, 0)
	}
	for i := 0; i < 3; i++ {
		avr435Push(g, owner, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	}
}

func avr435Count(g *game.Game, owner uuid.UUID, typ string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.HasCardType(typ) {
			n++
		}
	}
	return n
}

func avr435AnswerSacrifices(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 40; i++ {
		answered := false
		for _, s := range g.Seats {
			if c := effectstest.SacrificeChoiceFor(g, s.ID); c != nil {
				effectstest.AnswerSacrifice(t, g, s.ID, c.ChooseCards[:c.ChooseMin]...)
				answered = true
			}
		}
		if !answered {
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.SacrificeChoiceFor(g, g.Seats[0].ID) == nil && effectstest.SacrificeChoiceFor(g, g.Seats[1].ID) == nil &&
				effectstest.SacrificeChoiceFor(g, g.Seats[2].ID) == nil && effectstest.SacrificeChoiceFor(g, g.Seats[3].ID) == nil {
				return
			}
		}
	}
}

func TestAvrB435RiteOfRuinSacrificesInTheChosenOrder(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		option                     int
		artifacts, creatures, land int
	}{
		{"artifacts, creatures, lands", 0, 3, 1, 1},
		{"lands, creatures, artifacts", 5, 1, 1, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			for _, s := range g.Seats {
				avr435RiteBoard(g, s.ID)
			}
			effectstest.CastCatalogSpell(t, g, "Rite of Ruin", "Sorcery", avr435RiteOfRuin, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerOptionPick(t, g, g.Seats[0].ID, tc.option)
			avr435AnswerSacrifices(t, g)
			for _, s := range g.Seats {
				if a, c, l := avr435Count(g, s.ID, "artifact"), avr435Count(g, s.ID, "creature"), avr435Count(g, s.ID, "land"); a != tc.artifacts || c != tc.creatures || l != tc.land {
					t.Errorf("%s has artifacts/creatures/lands %d/%d/%d, want %d/%d/%d", s.ID, a, c, l, tc.artifacts, tc.creatures, tc.land)
				}
			}
		})
	}
}
