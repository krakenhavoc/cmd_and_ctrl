package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b24_cards_test.go — BNS batch 24 (Foundations Jumpstart): each test
// plays the real catalog entry through the engine.

const (
	bnsB24ScavengedBladeOracle = "2e7b4df3-7c36-469f-8db8-b3081a2c5691"
	bnsB24PickAxeOracle        = "6d2cf8b4-1468-4e00-94ca-65ef0c564b09"
	bnsB24ValkyrieOracle       = "236c4ff1-8c3f-4763-8df9-925021e65726"
	bnsB24SaddleOracle         = "d0eeb863-6c1b-4e84-9199-e77f15724626"
	bnsB24ConstructOracle      = "39363190-7f79-4fae-bcfc-b17ec4e07bfe"
	bnsB24KeeperOracle         = "3ef7e4c3-f87b-442d-8680-39673ede07e3"
	bnsB24JourneymageOracle    = "b400700c-1d82-4721-a166-56f88ba6ad19"
	bnsB24AdvocateOracle       = "1208cb6e-34b4-4a23-9a5b-a01229113e38"
	bnsB24AerialGuideOracle    = "b56c0846-e465-4251-93ea-2e3e7be8dffc"
	bnsB24ArtilleristOracle    = "860d0449-8032-458b-8335-30be238d6689"
	bnsB24AinokGuideOracle     = "e5f96482-4bae-4d1b-97b2-dcc09676eb6e"
	bnsB24EquipType            = "Artifact — Equipment"
)

func TestBnsB24EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB24ScavengedBladeOracle: "Scavenged Blade",
		bnsB24PickAxeOracle:        "Skyclave Pick-Axe",
		bnsB24ValkyrieOracle:       "Valkyrie's Sword",
		bnsB24SaddleOracle:         "Wolfrider's Saddle",
		bnsB24ConstructOracle:      "Eager Construct",
		bnsB24KeeperOracle:         "Treasure Keeper",
		bnsB24JourneymageOracle:    "Academy Journeymage",
		bnsB24AdvocateOracle:       "Advocate of the Beast",
		bnsB24AerialGuideOracle:    "Aerial Guide",
		bnsB24ArtilleristOracle:    "Ainok Artillerist",
		bnsB24AinokGuideOracle:     "Ainok Guide",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBnsB24ScavengedBladeAttachesOnEnteringAndPumpsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blade := castCatalogSpell(t, g, "Scavenged Blade", bnsB24EquipType, bnsB24ScavengedBladeOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 4/2", p, tough)
	}
	if c, _ := battlefieldCard(g, blade); !c.IsAttachedTo(bear) {
		t.Error("the Blade is not attached to the Bear")
	}
}

func TestBnsB24ScavengedBladeStaysUnattachedWithNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	blade := castCatalogSpell(t, g, "Scavenged Blade", bnsB24EquipType, bnsB24ScavengedBladeOracle, nil)
	assistCardSettle(t, g, me.ID, nil, true)
	if !g.Battlefield.Contains(blade) {
		t.Fatal("the Blade is not on the battlefield")
	}
	if host := attachmentHostOf(t, g, blade); host.Kind == game.TargetCard {
		t.Errorf("the Blade attached to %s with no creature to name", host.ID)
	}
}

func TestBnsB24PickAxeAttachesAndPumpsOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	axe := castCatalogSpell(t, g, "Skyclave Pick-Axe", bnsB24EquipType, bnsB24PickAxeOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if c, _ := battlefieldCard(g, axe); !c.IsAttachedTo(bear) {
		t.Fatal("the Pick-Axe is not attached to the Bear")
	}
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the equipped Bear is %d/%d after a land drop, want 4/4", p, tough)
	}
	if p, _ := ptOf(t, g, other); p != 2 {
		t.Errorf("the unequipped Bear has power %d, want 2", p)
	}
	advanceToEndStepOf(t, g, 0)
	advanceToUpkeepOf(t, g, 1)
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("the pump outlived the turn: power %d", p)
	}
}

func TestBnsB24WolfridersSaddleMakesAWolfAndWearsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	saddle := castCatalogSpell(t, g, "Wolfrider's Saddle", bnsB24EquipType, bnsB24SaddleOracle, nil)
	passPriorityAroundTable(t, g)
	host := attachmentHostOf(t, g, saddle)
	if host.Kind != game.TargetCard {
		t.Fatal("the Saddle is attached to nothing")
	}
	if got := tmtB1Count(g, me.ID, "Wolf"); got != 1 {
		t.Fatalf("%d Wolves, want 1", got)
	}
	if p, tough := ptOf(t, g, host.ID); p != 3 || tough != 3 {
		t.Errorf("the Wolf is %d/%d, want 3/3", p, tough)
	}
	sizes, colors := fdnDGTokens(g, me.ID, "Wolf")
	if len(sizes) != 1 || sizes[0] != [2]int{3, 3} || len(colors[0]) != 1 || colors[0][0] != "G" {
		t.Errorf("tokens %v %v, want one green Wolf", sizes, colors)
	}
}

func TestBnsB24WolfridersSaddleLimitsBlockersToOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	saddle := seedEquipment(g, me.ID, "Wolfrider's Saddle", bnsB24SaddleOracle)
	floatMana(t, g, me, "{C}{C}{C}")
	equipTo(t, g, me.ID, saddle, bear)
	b1 := b12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 4)
	b2 := b12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(b1, bear); err != nil {
		t.Fatalf("the first blocker: %v", err)
	}
	if err := g.DeclareBlocker(b2, bear); err == nil {
		t.Error("a second creature blocked the equipped creature")
	}
}

func TestBnsB24ValkyriesSwordMayPayForAnAngel(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		sword := seedEquipment(g, me.ID, "Valkyrie's Sword", bnsB24ValkyrieOracle)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: sword})
		})
		passPriorityAroundTable(t, g)
		if !hasPayUnlessFor(g, me.ID) {
			t.Fatal("no {4}{W} offer")
		}
		if pay {
			floatMana(t, g, me, "{C}{C}{C}{C}{W}")
		}
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		host := attachmentHostOf(t, g, sword)
		if !pay {
			if host.Kind == game.TargetCard {
				t.Errorf("declined, but the Sword is attached to %s", host.ID)
			}
			continue
		}
		if host.Kind != game.TargetCard {
			t.Fatal("paid, but the Sword is attached to nothing")
		}
		if p, tough := ptOf(t, g, host.ID); p != 6 || tough != 5 {
			t.Errorf("the Angel Warrior is %d/%d, want 6/5", p, tough)
		}
		if !hasEffectiveKeyword(t, g, host.ID, "flying") || !hasEffectiveKeyword(t, g, host.ID, "vigilance") {
			t.Error("the token lacks flying or vigilance")
		}
		if st := effectiveSubtypes(t, g, host.ID); !containsString(st, "Angel") || !containsString(st, "Warrior") {
			t.Errorf("the token is %v, want an Angel Warrior", st)
		}
	}
}

func TestBnsB24AcademyJourneymageCostsLessWithAWizardAndBounces(t *testing.T) {
	cast := func(t *testing.T, wizard bool) (*game.Game, uuid.UUID, uuid.UUID, error) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		if wizard {
			b12Creature(g, me.ID, "Wizard", "Creature — Human Wizard", 1, 1)
		}
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		floatMana(t, g, me, "{C}{C}{C}{U}")
		id, err := castFromHandErr(g, me, "Academy Journeymage", "Creature — Human Wizard", "{4}{U}", bnsB24JourneymageOracle,
			game.CastSpellParams{Strict: true})
		return g, id, theirs, err
	}
	t.Run("with a Wizard", func(t *testing.T) {
		g, _, theirs, err := cast(t, true)
		if err != nil {
			t.Fatalf("cast for {3}{U} with a Wizard: %v", err)
		}
		passPriorityAroundTable(t, g)
		answerPickTargetIfOpen(t, g, theirs)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) || !g.Seats[1].Hand.Contains(theirs) {
			t.Error("the opponent's creature was not returned to its owner's hand")
		}
	})
	t.Run("without a Wizard", func(t *testing.T) {
		if _, _, _, err := cast(t, false); err == nil {
			t.Error("cast for {3}{U} with no Wizard")
		}
	})
}

func TestBnsB24AdvocateOfTheBeastGrowsABeastAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Advocate of the Beast", "Creature — Elf Shaman", bnsB24AdvocateOracle, false)
	beast := b12Creature(g, me.ID, "Beast", "Creature — Beast", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Fatal("a non-Beast was accepted as the target")
		}
	}
	answerPickTargetIfOpen(t, g, beast)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, beast, game.CounterPlusOne); n != 1 {
		t.Errorf("the Beast has %d +1/+1 counters, want 1", n)
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("the Bear has %d +1/+1 counters, want 0", n)
	}
}

func TestBnsB24AerialGuideGivesAnotherAttackerFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	guide := pushCatalogPermanent(g, me.ID, "Aerial Guide", "Creature — Drake", bnsB24AerialGuideOracle, false)
	bear := seedBear(g, me.ID)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	declareAttack(t, g, opp.ID, guide, bear)
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the other attacker should gain flying")
	}
}

func TestBnsB24AinokArtilleristHasReachOnlyWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Ainok Artillerist", "Creature — Dog Archer", bnsB24ArtilleristOracle, false)
	if hasEffectiveKeyword(t, g, id, "reach") {
		t.Error("reach without a +1/+1 counter")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterPlusOne, 1) })
	if !hasEffectiveKeyword(t, g, id, "reach") {
		t.Error("no reach with a +1/+1 counter")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterPlusOne, -1) })
	if hasEffectiveKeyword(t, g, id, "reach") {
		t.Error("reach after the counter was removed")
	}
}

func TestBnsB24AinokGuideModes(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := castRingCreature(t, g, "Ainok Guide", "Creature — Dog Scout", bnsB24AinokGuideOracle, 1, 1)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		passPriorityAroundTable(t, g)
		if p, tough := ptOf(t, g, id); p != 2 || tough != 2 {
			t.Errorf("the Guide is %d/%d, want 2/2", p, tough)
		}
	})
	t.Run("search", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		seedSearchLibrary(me,
			searchTestLand("Forest", "Basic Land — Forest"),
			searchTestLand("Grizzly Bears", "Creature — Bear"),
			searchTestLand("Mystic Sanctuary", "Land — Island"),
		)
		castRingCreature(t, g, "Ainok Guide", "Creature — Dog Scout", bnsB24AinokGuideOracle, 1, 1)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		passPriorityAroundTable(t, g)
		c := searchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no search offered")
		}
		if searchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil || searchOptionNamed(g, c, "Mystic Sanctuary") != uuid.Nil {
			t.Error("only basic land cards are legal finds")
		}
		answerSearchNamed(t, g, me.ID, "Forest")
		passPriorityAroundTable(t, g)
		if top := me.Library.Cards[len(me.Library.Cards)-1]; top.Name != "Forest" {
			t.Errorf("the top card is %q, want Forest", top.Name)
		}
	})
}

func TestBnsB24EagerConstructAsksEachPlayerInTurn(t *testing.T) {
	g := newCatalogGame(t)
	seats := make([]*game.Player, 0, len(g.Seats))
	for i := range g.Seats {
		seats = append(seats, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)])
	}
	me := seats[0]
	topOf := func(p *game.Player) game.Card { return p.Library.Cards[len(p.Library.Cards)-1] }
	myTop := topOf(me)
	castHoldCreature(t, g, "Eager Construct", "Artifact Creature — Construct", bnsB24ConstructOracle, 2, 2)
	for i, p := range seats {
		c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, p.ID)
		if c == nil {
			t.Fatalf("%s (seat %d in turn order) was not asked to scry", p.Name, i)
		}
		for j, later := range seats[i+1:] {
			if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, later.ID) != nil {
				t.Fatalf("%s was asked before %s answered (%d)", later.Name, p.Name, j)
			}
		}
		yes := i == 0
		if err := g.ResolveConfirm(c.ID, p.ID, yes); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		if yes {
			s := latestChoiceOfKindFor(g, game.PendingChoiceScry, p.ID)
			if s == nil || len(s.ScryCards) != 1 {
				t.Fatalf("no scry 1 open after yes: %+v", g.PendingChoices)
			}
			if err := g.ResolveScry(s.ID, p.ID, s.ScryCards, nil); err != nil {
				t.Fatalf("ResolveScry: %v", err)
			}
		}
	}
	if me.Library.Cards[0].InstanceID != myTop.InstanceID {
		t.Error("the card scried to the bottom is not at the bottom of the library")
	}
}

func TestBnsB24TreasureKeeperCastsTheHitFreeAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	keeper := b12Push(g, me.ID, "Treasure Keeper", "Artifact Creature — Construct", bnsB24KeeperOracle, 3, 3)
	island := hobB10Card(me, "Island", "Basic Land — Island", "")
	big := hobB10Card(me, "Big Bear", "Creature — Bear", "{3}{G}")
	hit := hobB10Card(me, "Divination", "Sorcery", "{2}{U}")
	deep := hobB10Card(me, "Deep Card", "Creature — Bear", "{1}{G}")
	hobB10Stack(me, island, big, hit, deep)
	b27Kill(g, keeper)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, hit.InstanceID)
	if err := castThroughResolutionCast(t, g, me.ID, hit.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the revealed card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana in the pool: the cast was not free", me.ManaPool.Total())
	}
	if !me.Graveyard.Contains(hit.InstanceID) {
		t.Error("the cast spell did not resolve into the graveyard")
	}
	for _, id := range []uuid.UUID{island.InstanceID, big.InstanceID} {
		if !me.Library.Contains(id) {
			t.Errorf("revealed card %s is not back in the library", id)
		}
	}
	if me.Library.Size() != 3 || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != deep.InstanceID {
		t.Errorf("library %d cards: the unrevealed card should still be on top of the two bottomed", me.Library.Size())
	}
}

func TestBnsB24TreasureKeeperDecliningBottomsTheHitToo(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	keeper := b12Push(g, me.ID, "Treasure Keeper", "Artifact Creature — Construct", bnsB24KeeperOracle, 3, 3)
	land := hobB10Card(me, "Island", "Basic Land — Island", "")
	hit := hobB10Card(me, "Divination", "Sorcery", "{2}{U}")
	deep := hobB10Card(me, "Deep Card", "Creature — Bear", "{1}{G}")
	hobB10Stack(me, land, hit, deep)
	b27Kill(g, keeper)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, hit.InstanceID)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(hit.InstanceID) {
		t.Error("the declined card was cast")
	}
	for _, id := range []uuid.UUID{land.InstanceID, hit.InstanceID} {
		if !me.Library.Contains(id) {
			t.Errorf("card %s is not back in the library", id)
		}
	}
	if me.Library.Size() != 3 || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != deep.InstanceID {
		t.Error("the revealed cards were not put beneath the unrevealed one")
	}
}

func TestBnsB24TreasureKeeperWithNoHitBottomsEverything(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	keeper := b12Push(g, me.ID, "Treasure Keeper", "Artifact Creature — Construct", bnsB24KeeperOracle, 3, 3)
	a := hobB10Card(me, "Island", "Basic Land — Island", "")
	b := hobB10Card(me, "Big Bear", "Creature — Bear", "{4}{G}")
	hobB10Stack(me, a, b)
	b27Kill(g, keeper)
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a cast was offered with no legal card revealed")
	}
	if me.Library.Size() != 2 || !me.Library.Contains(a.InstanceID) || !me.Library.Contains(b.InstanceID) {
		t.Errorf("library %d cards, want both revealed cards back", me.Library.Size())
	}
}
