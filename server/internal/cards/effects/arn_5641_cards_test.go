package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// arn_5641_cards_test.go — a batch of Arabian Nights cards (#5641).
// Every test plays the real catalog entry through the engine.

const (
	arnBirdMaidenOracle          = "bddd6b93-7c69-427a-97ee-0c1f9f950f09"
	arnFlyingMenOracle           = "b5666c68-059d-4e32-9b04-548b3058430d"
	arnDancingScimitarOracle     = "82b13601-d460-45c5-94a1-07e146d463a9"
	arnMoorishCavalryOracle      = "bccdc42f-9e55-4f8e-a29e-ef39f38add6a"
	arnRepentantBlacksmithOracle = "a83073a2-e63d-4105-8bad-9612e411fc85"
	arnJuzamDjinnOracle          = "4e81596c-9225-43d1-bd35-798212144f2c"
	arnJununEfreetOracle         = "afda663e-c5f7-4182-86f7-d95d71793717"
	arnBrassManOracle            = "64f56228-7874-4465-ba58-1049083ea02f"
	arnAladdinsRingOracle        = "7810915f-df4d-445a-8ba8-fe6ed3abf1ae"
	arnDesertTwisterOracle       = "6f880348-6dc8-4cf8-9313-9893c41a70a3"
	arnArmyOfAllahOracle         = "3483946d-8645-4c22-b0ba-a65a44456324"
	arnPietyOracle               = "0c017406-7fc3-4701-93ec-ddb02044c12a"
	arnSandstormOracle           = "7bb41690-f8ec-462a-ba29-be453eb86fca"
	arnFlyingCarpetOracle        = "453524fb-2ffa-421f-883c-b77ae2850c52"
	arnJandorsSaddlebagsOracle   = "3aa0e73f-ac88-47a5-9cc5-0c941a939eae"
	arnWyluliWolfOracle          = "00185f3b-2777-4417-a8e0-4691f41c0ec1"
	arnKingSuleimanOracle        = "97f548b1-6398-4dc7-b5eb-da5a3f24ddb7"
	arnAliBabaOracle             = "ef3fdfea-330d-4948-94ec-67e344b48086"
	arnElephantGraveyardOracle   = "8ada7388-fd8b-434c-a17a-bce19cf3e615"
	arnDiamondValleyOracle       = "84cef34a-c3e1-4059-b4cd-c481938a53a5"
	arnKirdApeOracle             = "fbbc3acb-c917-44ff-ac6f-9dd6ebe3f4ad"
	arnGiantTortoiseOracle       = "2dd50d7f-941f-4deb-a15c-ee2357844c35"
	arnSorceressQueenOracle      = "3e4cb1b2-e2cc-4925-a226-6c6f1501d9c1"
	arnSingingTreeOracle         = "23f9bee4-ac7e-4828-82c0-576fee0d29b7"
	arnIslandOfWakWakOracle      = "d427e61d-5b30-4d2a-bad2-2e7f016036ca"
	arnElHajjajOracle            = "f92c9a5d-853f-4157-89ba-8d8c8033c533"
	arnUnstableMutationOracle    = "278b237e-9699-43eb-a03e-0b68eccc08b3"
	arnFishliverOilOracle        = "75457fe5-4ab6-42c4-98e5-8ed6e8bf122c"
)

// arnSetCombat marks `id` as attacking `attacking` and/or blocking
// `blocking` without walking the combat steps, for the spells and
// abilities that read who is in combat as they resolve.
func arnSetCombat(g *game.Game, id, attacking, blocking uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].AttackingTarget = attacking
				g.Battlefield.Cards[i].BlockingTarget = blocking
			}
		}
	})
}

// arnTap taps a permanent the way an effect would.
func arnTap(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.TapTargetForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
}

func arnTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// --- the keyword cards ------------------------------------------------

func TestArn5641KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Bird Maiden", "Creature — Human Bird", arnBirdMaidenOracle, "flying"},
		{"Flying Men", "Creature — Human", arnFlyingMenOracle, "flying"},
		{"Dancing Scimitar", "Artifact Creature — Spirit", arnDancingScimitarOracle, "flying"},
		{"Moorish Cavalry", "Creature — Human Knight", arnMoorishCavalryOracle, "trample"},
		{"Repentant Blacksmith", "Creature — Human", arnRepentantBlacksmithOracle, "protection from red"},
		{"Junún Efreet", "Creature — Efreet", arnJununEfreetOracle, "flying"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			if !hasAbility(effectiveAbilities(t, g, id), tc.keyword) {
				t.Errorf("%s has no %s", tc.name, tc.keyword)
			}
		})
	}
}

// Protection from red: a red spell can't target the Blacksmith.
func TestArn5641RepentantBlacksmithCantBeTargetedByRed(t *testing.T) {
	g := newCatalogGame(t)
	smith := b12Push(g, g.Seats[1].ID, "Repentant Blacksmith", "Creature — Human", arnRepentantBlacksmithOracle, 1, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == smith {
				g.Battlefield.Cards[i].Colors = []string{"W"}
			}
		}
	})
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle,
		ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: cardTarget(smith)}); err == nil {
		t.Error("Lightning Bolt targeted a creature with protection from red")
	}
}

// --- Juzám Djinn ------------------------------------------------------

func TestArn5641JuzamDjinnDealsOneToItsControllerEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner, other := g.Seats[1], g.Seats[2]
	djinn := b12Push(g, owner.ID, "Juzám Djinn", "Creature — Djinn", arnJuzamDjinnOracle, 5, 5)
	advanceToUpkeepOf(t, g, 2)
	if triggerOnStack(g, djinn) != nil {
		t.Fatal("the Djinn triggered on another player's upkeep")
	}
	life, otherLife := owner.Life, other.Life
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, djinn) == nil {
		t.Fatal("no upkeep trigger on its controller's upkeep")
	}
	passPriorityAroundTable(t, g)
	if owner.Life != life-1 {
		t.Errorf("controller life %d -> %d, want -1", life, owner.Life)
	}
	if other.Life != otherLife {
		t.Error("the damage reached another player")
	}
}

// --- Junún Efreet -----------------------------------------------------

func TestArn5641JununEfreetPaysOrIsSacrificed(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		owner := g.Seats[1]
		efreet := b12Push(g, owner.ID, "Junún Efreet", "Creature — Efreet", arnJununEfreetOracle, 3, 3)
		advanceToUpkeepOf(t, g, 1)
		if triggerOnStack(g, efreet) == nil {
			t.Fatal("no upkeep trigger")
		}
		passPriorityAroundTable(t, g)
		if c := upkeepPromptFor(t, g, owner.ID); c.PayCost != "{B}{B}" {
			t.Fatalf("prompt cost %q, want {B}{B}", c.PayCost)
		}
		if pay {
			apaMana(owner, "B", "B")
		}
		answerPayUnless(t, g, owner.ID, pay)
		if kept := g.Battlefield.Contains(efreet); kept != pay {
			t.Errorf("pay=%v: Efreet kept = %v", pay, kept)
		}
	}
}

// --- Brass Man --------------------------------------------------------

func TestArn5641BrassManStaysTappedUnlessItsControllerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		owner := g.Seats[1]
		brass := b12Push(g, owner.ID, "Brass Man", "Artifact Creature — Construct", arnBrassManOracle, 1, 3)
		arnTap(t, g, brass)
		advanceToUpkeepOf(t, g, 1)
		if !arnTapped(t, g, brass) {
			t.Fatal("Brass Man untapped in its controller's untap step")
		}
		if triggerOnStack(g, brass) == nil {
			t.Fatal("no upkeep trigger")
		}
		passPriorityAroundTable(t, g)
		if pay {
			apaMana(owner, "C")
		}
		answerPayUnless(t, g, owner.ID, pay)
		if untapped := !arnTapped(t, g, brass); untapped != pay {
			t.Errorf("pay=%v: untapped = %v", pay, untapped)
		}
	}
}

// --- Aladdin's Ring ---------------------------------------------------

func TestArn5641AladdinsRingDealsFour(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	ring := pushCatalogPermanent(g, me.ID, "Aladdin's Ring", "Artifact", arnAladdinsRingOracle, false)
	life := opp.Life
	apaMana(me, "C", "C", "C", "C", "C", "C", "C", "C")
	apaActivate(t, g, me, ring, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want -4", life, opp.Life)
	}
	if !arnTapped(t, g, ring) {
		t.Error("the Ring was not tapped to pay the cost")
	}
}

// --- Desert Twister ---------------------------------------------------

func TestArn5641DesertTwisterDestroysAnyPermanent(t *testing.T) {
	g := newCatalogGame(t)
	land := b12Permanent(g, g.Seats[1].ID, "Their Land", "Land")
	castCatalogSpell(t, g, "Desert Twister", "Sorcery", arnDesertTwisterOracle, cardTarget(land))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the targeted land survived Desert Twister")
	}
}

// --- Army of Allah, Piety, Sandstorm ----------------------------------

func TestArn5641ArmyOfAllahPumpsEveryAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	home := b12Creature(g, me.ID, "Home", "Creature — Bear", 2, 2)
	arnSetCombat(g, mine, opp.ID, uuid.Nil)
	arnSetCombat(g, theirs, me.ID, uuid.Nil)
	castCatalogSpell(t, g, "Army of Allah", "Instant", arnArmyOfAllahOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tt := ptOf(t, g, id); p != 4 || tt != 2 {
			t.Errorf("attacker %d/%d, want 4/2", p, tt)
		}
	}
	if p, tt := ptOf(t, g, home); p != 2 || tt != 2 {
		t.Errorf("a creature at home %d/%d, want 2/2", p, tt)
	}
}

func TestArn5641PietyPumpsEveryBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	idle := b12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	arnSetCombat(g, attacker, opp.ID, uuid.Nil)
	arnSetCombat(g, blocker, uuid.Nil, attacker)
	castCatalogSpell(t, g, "Piety", "Instant", arnPietyOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, blocker); p != 2 || tt != 5 {
		t.Errorf("blocker %d/%d, want 2/5", p, tt)
	}
	for _, id := range []uuid.UUID{attacker, idle} {
		if p, tt := ptOf(t, g, id); p != 2 || tt != 2 {
			t.Errorf("a creature not blocking %d/%d, want 2/2", p, tt)
		}
	}
}

func TestArn5641SandstormDamagesEachAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	home := b12Creature(g, opp.ID, "Home", "Creature — Bear", 1, 1)
	arnSetCombat(g, small, me.ID, uuid.Nil)
	arnSetCombat(g, big, me.ID, uuid.Nil)
	life := me.Life
	castCatalogSpell(t, g, "Sandstorm", "Instant", arnSandstormOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 1/1 attacker survived 1 damage")
	}
	if pr6Marked(g, big) != 1 {
		t.Errorf("the 3/3 attacker has %d damage, want 1", pr6Marked(g, big))
	}
	if !g.Battlefield.Contains(home) || pr6Marked(g, home) != 0 {
		t.Error("a creature not attacking was damaged")
	}
	if me.Life != life {
		t.Error("Sandstorm damaged a player")
	}
}

// --- the activated abilities ------------------------------------------

func TestArn5641FlyingCarpetGrantsFlyingForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	carpet := pushCatalogPermanent(g, me.ID, "Flying Carpet", "Artifact", arnFlyingCarpetOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	apaMana(me, "C", "C")
	apaActivate(t, g, me, carpet, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target has no flying")
	}
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("flying outlasted the turn")
	}
}

func TestArn5641JandorsSaddlebagsUntapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bags := pushCatalogPermanent(g, me.ID, "Jandor's Saddlebags", "Artifact", arnJandorsSaddlebagsOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	arnTap(t, g, bear)
	apaMana(me, "C", "C", "C")
	apaActivate(t, g, me, bags, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if arnTapped(t, g, bear) {
		t.Error("the target is still tapped")
	}
	if !arnTapped(t, g, bags) {
		t.Error("the Saddlebags were not tapped to pay the cost")
	}
}

func TestArn5641WyluliWolfPumpsATarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	wolf := pushCatalogPermanent(g, me.ID, "Wyluli Wolf", "Creature — Wolf", arnWyluliWolfOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	apaActivate(t, g, me, wolf, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("target %d/%d, want 3/3", p, tt)
	}
}

func TestArn5641KingSuleimanDestroysOnlyDjinnAndEfreet(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		legal    bool
	}{
		{"Creature — Djinn", true},
		{"Creature — Efreet", true},
		{"Creature — Bear", false},
	} {
		t.Run(tc.typeLine, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, opp := top100ActiveSeat(g), top100Opponent(g)
			king := pushCatalogPermanent(g, me.ID, "King Suleiman", "Creature — Human Noble", arnKingSuleimanOracle, false)
			victim := b12Creature(g, opp.ID, "Victim", tc.typeLine, 5, 5)
			err := g.ActivateCatalogAbility(me.ID, king, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(victim)})
			if !tc.legal {
				if err == nil {
					t.Fatal("King Suleiman targeted a creature that is neither a Djinn nor an Efreet")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) {
				t.Error("the target survived")
			}
		})
	}
}

func TestArn5641AliBabaTapsAWall(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	ali := pushCatalogPermanent(g, me.ID, "Ali Baba", "Creature — Human Rogue", arnAliBabaOracle, true)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, ali, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Error("Ali Baba targeted a creature that is not a Wall")
	}
	// No {T} in the cost: the turn he arrives is fine.
	apaMana(me, "R")
	apaActivate(t, g, me, ali, 0, game.ActivateAbilityParams{Targets: cardTarget(wall)})
	if !arnTapped(t, g, wall) {
		t.Error("the Wall is untapped")
	}
}

func TestArn5641ElephantGraveyardRegeneratesAnElephant(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	yard := pushCatalogPermanent(g, me.ID, "Elephant Graveyard", "Land", arnElephantGraveyardOracle, false)
	elephant := b12Creature(g, me.ID, "Elephant", "Creature — Elephant", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, yard, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Error("Elephant Graveyard targeted a creature that is not an Elephant")
	}
	apaActivate(t, g, me, yard, 0, game.ActivateAbilityParams{Targets: cardTarget(elephant)})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(elephant); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(elephant) {
		t.Fatal("the shielded Elephant was destroyed")
	}
	if !arnTapped(t, g, elephant) {
		t.Error("regeneration did not tap the Elephant")
	}
}

func TestArn5641ElephantGraveyardTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	yard := pushCatalogPermanent(g, me.ID, "Elephant Graveyard", "Land", arnElephantGraveyardOracle, false)
	if err := g.ActivateManaAbility(me.ID, yard, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if n := me.ManaPool.Total(); n != 1 {
		t.Errorf("pool holds %d, want 1", n)
	}
}

func TestArn5641DiamondValleyGainsTheToughness(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	valley := pushCatalogPermanent(g, me.ID, "Diamond Valley", "Land", arnDiamondValleyOracle, false)
	ox := b12Creature(g, me.ID, "Ox", "Creature — Ox", 1, 4)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, valley, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ox}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ox) {
		t.Error("the Ox was not sacrificed")
	}
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d", me.Life, life+4)
	}
}

func TestArn5641SorceressQueenMakesAnotherCreatureZeroTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	queen := pushCatalogPermanent(g, me.ID, "Sorceress Queen", "Creature — Human Wizard Sorcerer", arnSorceressQueenOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, queen, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(queen)}); err == nil {
		t.Fatal("Sorceress Queen targeted herself")
	}
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
	apaActivate(t, g, me, queen, 0, game.ActivateAbilityParams{Targets: cardTarget(giant)})
	if p, tt := ptOf(t, g, giant); p != 0 || tt != 2 {
		t.Errorf("target %d/%d, want 0/2", p, tt)
	}
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, tt := ptOf(t, g, giant); p != 6 || tt != 6 {
		t.Errorf("next turn %d/%d, want 6/6", p, tt)
	}
}

func TestArn5641SingingTreeZeroesAnAttackersPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	tree := pushCatalogPermanent(g, me.ID, "Singing Tree", "Creature — Plant", arnSingingTreeOracle, false)
	attacker := b12Creature(g, opp.ID, "Attacker", "Creature — Giant", 5, 5)
	idle := b12Creature(g, opp.ID, "Idle", "Creature — Giant", 5, 5)
	if err := g.ActivateCatalogAbility(me.ID, tree, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(idle)}); err == nil {
		t.Fatal("Singing Tree targeted a creature that is not attacking")
	}
	arnSetCombat(g, attacker, me.ID, uuid.Nil)
	apaActivate(t, g, me, tree, 0, game.ActivateAbilityParams{Targets: cardTarget(attacker)})
	if p, tt := ptOf(t, g, attacker); p != 0 || tt != 5 {
		t.Errorf("attacker %d/%d, want 0/5", p, tt)
	}
}

func TestArn5641IslandOfWakWakZeroesAFliersPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	island := pushCatalogPermanent(g, me.ID, "Island of Wak-Wak", "Land", arnIslandOfWakWakOracle, false)
	flier := b12Push(g, opp.ID, "Flying Men", "Creature — Human", arnFlyingMenOracle, 3, 3)
	walker := b12Creature(g, opp.ID, "Walker", "Creature — Bear", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, island, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(walker)}); err == nil {
		t.Fatal("Island of Wak-Wak targeted a creature without flying")
	}
	apaActivate(t, g, me, island, 0, game.ActivateAbilityParams{Targets: cardTarget(flier)})
	if p, tt := ptOf(t, g, flier); p != 0 || tt != 3 {
		t.Errorf("flier %d/%d, want 0/3", p, tt)
	}
}

// --- the statics -----------------------------------------------------

func TestArn5641KirdApeGrowsWithAForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ape := b12Push(g, me.ID, "Kird Ape", "Creature — Ape", arnKirdApeOracle, 1, 1)
	b12Permanent(g, g.Seats[1].ID, "Their Forest", "Basic Land — Forest")
	if p, tt := ptOf(t, g, ape); p != 1 || tt != 1 {
		t.Errorf("with no Forest of its controller's: %d/%d, want 1/1", p, tt)
	}
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if p, tt := ptOf(t, g, ape); p != 2 || tt != 3 {
		t.Errorf("with a Forest: %d/%d, want 2/3", p, tt)
	}
	killCreature(t, g, me.ID, forest)
	if p, tt := ptOf(t, g, ape); p != 1 || tt != 1 {
		t.Errorf("after the Forest left: %d/%d, want 1/1", p, tt)
	}
}

func TestArn5641GiantTortoiseIsTougherUntapped(t *testing.T) {
	g := newCatalogGame(t)
	tortoise := b12Push(g, g.Seats[0].ID, "Giant Tortoise", "Creature — Turtle", arnGiantTortoiseOracle, 1, 1)
	if p, tt := ptOf(t, g, tortoise); p != 1 || tt != 4 {
		t.Errorf("untapped %d/%d, want 1/4", p, tt)
	}
	arnTap(t, g, tortoise)
	if p, tt := ptOf(t, g, tortoise); p != 1 || tt != 1 {
		t.Errorf("tapped %d/%d, want 1/1", p, tt)
	}
}

// --- El-Hajjâj ---------------------------------------------------------

func TestArn5641ElHajjajGainsTheDamageItDeals(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hajjaj := b12Push(g, me.ID, "El-Hajjâj", "Creature — Human Wizard", arnElHajjajOracle, 1, 4)
	target := b12Creature(g, opp.ID, "Target", "Creature — Giant", 5, 5)
	life := me.Life
	pr6Damage(t, g, hajjaj, target, 3)
	if triggerOnStack(g, hajjaj) == nil {
		t.Fatal("no trigger for the damage it dealt")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d", me.Life, life+3)
	}
	// Damage dealt TO it is not damage it deals.
	pr6Damage(t, g, target, hajjaj, 1)
	if triggerOnStack(g, hajjaj) != nil {
		t.Error("triggered on damage it did not deal")
	}
}

// --- the Auras ---------------------------------------------------------

func TestArn5641UnstableMutationPumpsThenShrinksOnItsHostsUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	auraCast(t, g, "Unstable Mutation", arnUnstableMutationOracle, bear)
	if p, tt := ptOf(t, g, bear); p != 5 || tt != 5 {
		t.Fatalf("enchanted %d/%d, want 5/5", p, tt)
	}
	// The enchanted creature's controller's upkeep.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := minusOneOn(t, g, bear); n != 1 {
		t.Fatalf("-1/-1 counters = %d, want 1", n)
	}
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("after one upkeep %d/%d, want 4/4", p, tt)
	}
	// The Aura's controller's upkeep, and the others', are not the one.
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := minusOneOn(t, g, bear); n != 1 {
		t.Errorf("-1/-1 counters after the other upkeeps = %d, want 1", n)
	}
	_ = me
}

func TestArn5641FishliverOilGrantsIslandwalk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	auraCast(t, g, "Fishliver Oil", arnFishliverOilOracle, bear)
	if !hasEffectiveKeyword(t, g, bear, "islandwalk") {
		t.Error("the enchanted creature has no islandwalk")
	}
}
