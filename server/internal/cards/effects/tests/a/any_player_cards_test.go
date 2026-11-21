package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_cards_test.go — the "Any player may activate this
// ability" cards of ADR 0106 delivery PR 6 (#1793), beyond Xantcha
// (xantcha_sleeper_agent_test.go). Each test has a player who does NOT
// control the permanent activate it and pay for it (CR 602.1a), and
// pins what the card does and to whom: the activator is "you" (CR
// 109.5), "this creature" is the permanent.

const (
	apaFlailingManticore = "044a169e-32ae-464a-a642-8d3409e82d9f"
	apaFlailingOgre      = "e81fe764-8b26-40b3-8e17-741ab5e231b4"
	apaFlailingSoldier   = "2c1dd6f4-2db5-449d-9ed8-579d98cfcfb8"
	apaFanFavorite       = "9991c9f2-cb6e-4cfe-a3b7-650212454195"
	apaOonasProwler      = "4f87252e-3218-4759-accc-a4f186fa8857"
	apaZerapaMinotaur    = "4520c630-3650-462f-ab01-215eb2bdbbf5"
	apaRibbonSnake       = "98b53af9-af1b-4d5d-ae5a-bd444c1339d2"
	apaVintaraElephant   = "0a73b4da-9b5c-480e-b58b-6a82a52ea71e"
	apaSquallmonger      = "50bf5af6-50d7-4dea-936e-508e24db0a03"
	apaWarmonger         = "bae92332-1a0b-476b-8719-190e8d8cc03a"
	apaIfhBiffEfreet     = "e503a4f2-a785-4e7a-89a7-a9b24fb98831"
	apaSailmonger        = "b15af529-a61a-4228-bf2a-7304c9050720"
	apaScandalmonger     = "2d98d8d8-e277-42c4-bf94-dd7d7f2047c3"
	apaExcavation        = "9f7b2ec6-a828-499b-a2ac-766bdaaf30c1"
	apaWellOfKnowledge   = "99e2b36b-1fab-4b09-b371-d97a6854dfbc"
	apaSaprolingCluster  = "3b64051e-d320-4991-bc9e-6e1435a23c17"
	apaFeralHydra        = "b5d5dd1f-2d31-4849-8cb9-08354dd48d31"
	apaCaseyJones        = "1eaa37a3-8769-4213-bcd5-39bcf4771273"
	apaQuicksilverWall   = "82372776-bedd-4893-92ad-f1272f9c3250"
	apaEndbringersRevel  = "4ca2511d-76f0-470c-b522-b75fb08932d9"
)

// Every card of the batch registers only any-player rows, and only the
// two plain card draws declare what the row buys another player, the
// half of the purpose the bot reads across the table (owner decision
// 2). The three damage sweepers declare their sweep (ADR 0126 §6),
// which buys an activator nothing the bot prices.
func TestAnyPlayerBatchRowsAndPurposes(t *testing.T) {
	cases := []struct {
		oracle string
		rows   int
		draws  int
	}{
		{apaFlailingManticore, 2, 0}, {apaFlailingOgre, 2, 0}, {apaFlailingSoldier, 2, 0},
		{apaFanFavorite, 1, 0}, {apaOonasProwler, 1, 0}, {apaZerapaMinotaur, 1, 0},
		{apaRibbonSnake, 1, 0}, {apaVintaraElephant, 1, 0}, {apaSquallmonger, 1, 0},
		{apaWarmonger, 1, 0}, {apaIfhBiffEfreet, 1, 0}, {apaSailmonger, 1, 0},
		{apaScandalmonger, 1, 0}, {apaExcavation, 1, 1}, {apaWellOfKnowledge, 1, 1},
		{apaSaprolingCluster, 1, 0}, {apaFeralHydra, 1, 0}, {apaCaseyJones, 1, 0},
		{apaQuicksilverWall, 1, 0}, {apaEndbringersRevel, 1, 0},
	}
	for _, tc := range cases {
		rows := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(rows) != tc.rows {
			t.Errorf("%s: %d rows, want %d", tc.oracle, len(rows), tc.rows)
			continue
		}
		for i, r := range rows {
			if !r.AnyPlayer {
				t.Errorf("%s row %d is not an any-player row", tc.oracle, i)
			}
			across := game.Purpose{Draws: r.Purpose.Draws, ControllerLosesLife: r.Purpose.ControllerLosesLife}
			if want := (game.Purpose{Draws: tc.draws}); across != want {
				t.Errorf("%s row %d purpose = %+v, want %+v", tc.oracle, i, r.Purpose, want)
			}
			sweeper := tc.oracle == apaSquallmonger || tc.oracle == apaWarmonger || tc.oracle == apaIfhBiffEfreet
			if r.Purpose.Sweep.IsZero() == sweeper {
				t.Errorf("%s row %d sweep = %+v, want one exactly on the three damage sweepers", tc.oracle, i, r.Purpose.Sweep)
			}
			if rest := r.Purpose; rest.Discards+rest.Lands+rest.Tutors+rest.SelfMillTutor+rest.Tokens != 0 || rest.DeathPayoff {
				t.Errorf("%s row %d declares %+v, more than its draws and its sweep", tc.oracle, i, r.Purpose)
			}
		}
	}
}

// The Flailing pair: two other players pump and shrink the controller's
// Ogre, each from their own pool.
func TestFlailingOgreAnyPlayerPumpsAndShrinks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob, cat := g.Seats[0], g.Seats[1], g.Seats[2]
	ogre := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Flailing Ogre", apaFlailingOgre, 3, 3))

	effectstest.ApaMana(bob, "C")
	effectstest.ApaActivate(t, g, bob, ogre, 1, game.ActivateAbilityParams{}) // -1/-1
	if c := effectstest.ApaLive(g, ogre); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Fatalf("after Bob's -1/-1: %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.ApaMana(cat, "C")
	effectstest.ApaActivate(t, g, cat, ogre, 0, game.ActivateAbilityParams{}) // +1/+1
	if c := effectstest.ApaLive(g, ogre); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("after Cat's +1/+1: %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.ApaLive(g, ogre); c.Controller != me.ID {
		t.Errorf("the Ogre changed controller")
	}
}

// Shrinking a Flailing Soldier to 0 toughness kills it (CR 704.5f).
func TestFlailingSoldierShrunkToZeroDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	soldier := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Flailing Soldier", apaFlailingSoldier, 2, 2))
	for i := 0; i < 2; i++ {
		effectstest.ApaMana(bob, "C")
		effectstest.ApaActivate(t, g, bob, soldier, 1, game.ActivateAbilityParams{})
	}
	if effects.OnBattlefield(g, soldier) {
		t.Fatalf("a 0/0 Flailing Soldier is still on the battlefield")
	}
	if me.Graveyard.Size() != 1 {
		t.Errorf("owner's graveyard holds %d, want the Soldier", me.Graveyard.Size())
	}
}

func TestFlailingManticoreKeepsItsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	m := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Flailing Manticore", apaFlailingManticore, 3, 3))
	c := effectstest.ApaLive(g, m)
	if !game.HasKeyword(c, "flying") || !game.HasKeyword(c, "first strike") {
		t.Errorf("Flailing Manticore abilities = %v", c.Effective().Abilities)
	}
}

func TestFanFavoriteAnyPlayerPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	fan := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Fan Favorite", apaFanFavorite, 2, 2))
	effectstest.ApaMana(bob, "C", "C")
	effectstest.ApaActivate(t, g, bob, fan, 0, game.ActivateAbilityParams{})
	if c := effectstest.ApaLive(g, fan); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Fan Favorite %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
}

// ADR 0150 §4: Fan Favorite's assist, from the catalog's printed
// keyword alone. Bob pays its {3}; the caster pays the {B}.
func TestFanFavoriteAssistLetsAnotherPlayerPayTheGeneric(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bob := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	fan := game.NewCard("Fan Favorite", me.ID)
	fan.OracleID = apaFanFavorite
	fan.TypeLine = "Creature — Human Rogue"
	fan.ManaCost = "{3}{B}"
	me.Hand.PushTop(fan)
	effectstest.ApaMana(me, "B")
	effectstest.ApaMana(bob, "C", "C", "C")
	if err := g.CastSpell(me.ID, fan.InstanceID, game.CastSpellParams{Strict: true, AssistPlayer: bob.ID}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Assist == nil {
		t.Fatalf("the cast did not wait for Bob: %d prompts", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Chooser != bob.ID || c.PayAmount.Max != 3 || c.Assist.Generic != 3 {
		t.Fatalf("prompt to %v, max %d, generic %d", c.Chooser, c.PayAmount.Max, c.Assist.Generic)
	}
	if err := g.ResolvePayAmount(c.ID, bob.ID, 3); err != nil {
		t.Fatalf("Bob helps: %v", err)
	}
	item := g.StackMeta[fan.InstanceID]
	if item == nil || item.Paid.AssistedBy != bob.ID || item.Paid.AssistPaid != 3 {
		t.Fatalf("stack item = %+v", item)
	}
	if me.ManaPool.Total() != 0 || bob.ManaPool.Total() != 0 {
		t.Errorf("pools left %d and %d, want both spent", me.ManaPool.Total(), bob.ManaPool.Total())
	}
}

// The discard is the activator's: a card from THEIR hand pays.
func TestOonasProwlerActivatorDiscardsFromTheirOwnHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	prowler := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Oona's Prowler", apaOonasProwler, 3, 1))
	myHand, bobHand := me.Hand.Size(), bob.Hand.Size()
	pitch := bob.Hand.Cards[0].InstanceID

	// The controller's card cannot pay for Bob's activation.
	err := g.ActivateCatalogAbility(bob.ID, prowler, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{me.Hand.Cards[0].InstanceID}})
	if err == nil {
		t.Fatalf("Bob paid Oona's Prowler's discard with the controller's card")
	}
	effectstest.ApaActivate(t, g, bob, prowler, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}})
	if top, err := bob.Graveyard.Top(); bob.Hand.Size() != bobHand-1 || err != nil || top.InstanceID != pitch {
		t.Errorf("Bob's hand %d (want %d), graveyard top is not the pitched card", bob.Hand.Size(), bobHand-1)
	}
	if me.Hand.Size() != myHand {
		t.Errorf("the controller's hand changed: %d, want %d", me.Hand.Size(), myHand)
	}
	c := effectstest.ApaLive(g, prowler)
	if c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("Oona's Prowler %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
	if !game.HasKeyword(c, "flying") {
		t.Errorf("Oona's Prowler lost flying")
	}
}

// "Loses <keyword> until end of turn", on the three cards that print it.
func TestAnyPlayerKeywordLoss(t *testing.T) {
	cases := []struct {
		name, oracle, keyword string
		cost                  []string
	}{
		{"Zerapa Minotaur", apaZerapaMinotaur, "first strike", []string{"C", "C"}},
		{"Ribbon Snake", apaRibbonSnake, "flying", []string{"C", "C"}},
		{"Vintara Elephant", apaVintaraElephant, "trample", []string{"C", "C", "C"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, bob := g.Seats[0], g.Seats[1]
			id := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature(tc.name, tc.oracle, 3, 3))
			if !game.HasKeyword(effectstest.ApaLive(g, id), tc.keyword) {
				t.Fatalf("%s does not have %s printed", tc.name, tc.keyword)
			}
			effectstest.ApaMana(bob, tc.cost...)
			effectstest.ApaActivate(t, g, bob, id, 0, game.ActivateAbilityParams{})
			if game.HasKeyword(effectstest.ApaLive(g, id), tc.keyword) {
				t.Errorf("%s still has %s after Bob's activation", tc.name, tc.keyword)
			}
			effectstest.AdvancePastCleanupForTest(t, g)
			if !game.HasKeyword(effectstest.ApaLive(g, id), tc.keyword) {
				t.Errorf("%s did not get %s back at end of turn", tc.name, tc.keyword)
			}
		})
	}
}

// The Mongers' symmetric damage: the creatures the card names and every
// player, whoever activated it, with the permanent as the source.
func TestAnyPlayerDamageToEachCreatureAndEachPlayer(t *testing.T) {
	cases := []struct {
		name, oracle string
		cost         []string
		flier        bool // the source itself has flying
		hitsFliers   bool
	}{
		{"Squallmonger", apaSquallmonger, []string{"C", "C"}, false, true},
		{"Warmonger", apaWarmonger, []string{"C", "C"}, false, false},
		{"Ifh-Bíff Efreet", apaIfhBiffEfreet, []string{"G"}, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, bob := g.Seats[0], g.Seats[1]
			src := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature(tc.name, tc.oracle, 3, 3))
			flier := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.ApaCreature("Ribbon Snake", apaRibbonSnake, 2, 3))
			ground := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.ApaCreature("Ground Bear", "", 2, 2))
			lives := map[uuid.UUID]int{}
			for _, p := range g.Seats {
				lives[p.ID] = p.Life
			}
			effectstest.ApaMana(bob, tc.cost...)
			effectstest.ApaActivate(t, g, bob, src, 0, game.ActivateAbilityParams{})
			for _, p := range g.Seats {
				if p.Life != lives[p.ID]-1 {
					t.Errorf("%s life %d, want %d", p.Name, p.Life, lives[p.ID]-1)
				}
			}
			want := func(hit bool) int {
				if hit {
					return 1
				}
				return 0
			}
			if got := effectstest.ApaLive(g, flier).DamageMarked; got != want(tc.hitsFliers) {
				t.Errorf("flier damage %d, want %d", got, want(tc.hitsFliers))
			}
			if got := effectstest.ApaLive(g, ground).DamageMarked; got != want(!tc.hitsFliers) {
				t.Errorf("ground creature damage %d, want %d", got, want(!tc.hitsFliers))
			}
			if got := effectstest.ApaLive(g, src).DamageMarked; got != want(tc.flier == tc.hitsFliers) {
				t.Errorf("source damage %d, want %d", got, want(tc.flier == tc.hitsFliers))
			}
		})
	}
}

// Sailmonger: the activator chooses the target, here their own creature.
func TestSailmongerActivatorChoosesTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	sail := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Sailmonger", apaSailmonger, 3, 3))
	bear := effectstest.ApaPush(g, bob.ID, bob.ID, effectstest.ApaCreature("Ground Bear", "", 2, 2))
	effectstest.ApaMana(bob, "C", "C")
	effectstest.ApaActivate(t, g, bob, sail, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if !game.HasKeyword(effectstest.ApaLive(g, bear), "flying") {
		t.Errorf("Bob's creature did not gain flying")
	}
	if game.HasKeyword(effectstest.ApaLive(g, sail), "flying") {
		t.Errorf("Sailmonger itself gained flying")
	}
}

// Scandalmonger: "only as a sorcery" is the ACTIVATOR's sorcery timing.
func TestScandalmongerOnlyAsASorceryForTheActivator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	sm := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Scandalmonger", apaScandalmonger, 3, 3))
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}

	// Seat 0's main phase: not Bob's sorcery window.
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.ApaMana(bob, "C", "C")
	if err := g.ActivateCatalogAbility(bob.ID, sm, 0, game.ActivateAbilityParams{Targets: target}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("Bob activated Scandalmonger during another player's turn: err = %v", err)
	}
	bob.ManaPool = nil

	// Bob's own main phase, empty stack: his window, not the controller's.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sm, 0, game.ActivateAbilityParams{Targets: target}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("the controller activated Scandalmonger during Bob's turn: err = %v", err)
	}
	me.ManaPool = nil
	effectstest.ApaMana(bob, "C", "C")
	if err := g.ActivateCatalogAbility(bob.ID, sm, 0, game.ActivateAbilityParams{Targets: target, Strict: true}); err != nil {
		t.Fatalf("Bob activates Scandalmonger in his main phase: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, me.ID); got != 1 {
		t.Fatalf("the targeted controller owes %d discards, want 1", got)
	}
	if got := effectstest.DiscardOwed(g, bob.ID); got != 0 {
		t.Errorf("the activator owes %d discards, want 0", got)
	}
	hand := me.Hand.Size()
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand-1 {
		t.Errorf("controller hand %d after the discard, want %d", me.Hand.Size(), hand-1)
	}
}

// Excavation: the land is the activator's, and the activator draws.
func TestExcavationActivatorSacrificesTheirLandAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	ex := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Excavation", OracleID: apaExcavation, TypeLine: "Enchantment"})
	myLand := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	bobLand := effectstest.ApaPush(g, bob.ID, bob.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	effectstest.ApaMana(bob, "C")
	if err := g.ActivateCatalogAbility(bob.ID, ex, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{myLand}}); err == nil {
		t.Fatalf("Bob sacrificed the controller's land (CR 701.21a)")
	}
	bob.ManaPool = nil
	myHand, bobHand := me.Hand.Size(), bob.Hand.Size()
	effectstest.ApaMana(bob, "C")
	effectstest.ApaActivate(t, g, bob, ex, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bobLand}})
	if effects.OnBattlefield(g, bobLand) || !effects.OnBattlefield(g, myLand) {
		t.Errorf("wrong land sacrificed: Bob's on battlefield %v, controller's %v", effects.OnBattlefield(g, bobLand), effects.OnBattlefield(g, myLand))
	}
	if bob.Hand.Size() != bobHand+1 {
		t.Errorf("Bob's hand %d, want %d", bob.Hand.Size(), bobHand+1)
	}
	if me.Hand.Size() != myHand {
		t.Errorf("the controller drew: hand %d, want %d", me.Hand.Size(), myHand)
	}
}

// Well of Knowledge: "their draw step" is the activator's own.
func TestWellOfKnowledgeOnlyDuringTheActivatorsDrawStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t) // seat 0's draw step
	me, bob := g.Seats[0], g.Seats[1]
	well := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Well of Knowledge", OracleID: apaWellOfKnowledge, TypeLine: "Artifact"})

	effectstest.ApaMana(bob, "C", "C")
	if err := g.ActivateCatalogAbility(bob.ID, well, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("Bob activated Well of Knowledge in the controller's draw step: err = %v", err)
	}
	bob.ManaPool = nil

	effectstest.AdvanceToDrawStepOf(t, g, 1)
	if err := g.ActivateCatalogAbility(me.ID, well, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("the controller activated Well of Knowledge in Bob's draw step: err = %v", err)
	}
	myHand, bobHand := me.Hand.Size(), bob.Hand.Size()
	effectstest.ApaMana(bob, "C", "C")
	effectstest.ApaActivate(t, g, bob, well, 0, game.ActivateAbilityParams{})
	if bob.Hand.Size() != bobHand+1 {
		t.Errorf("Bob's hand %d, want %d", bob.Hand.Size(), bobHand+1)
	}
	if me.Hand.Size() != myHand {
		t.Errorf("the controller drew: hand %d, want %d", me.Hand.Size(), myHand)
	}
}

// Saproling Cluster: the activator discards and gets the token.
func TestSaprolingClusterTokenIsTheActivators(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	sc := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Saproling Cluster", OracleID: apaSaprolingCluster, TypeLine: "Enchantment"})
	bobHand := bob.Hand.Size()
	effectstest.ApaMana(bob, "C")
	effectstest.ApaActivate(t, g, bob, sc, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{bob.Hand.Cards[0].InstanceID}})
	if bob.Hand.Size() != bobHand-1 {
		t.Errorf("Bob's hand %d, want %d", bob.Hand.Size(), bobHand-1)
	}
	var saps []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Saproling" {
			saps = append(saps, c)
		}
	}
	if len(saps) != 1 || saps[0].Controller != bob.ID || saps[0].Owner != bob.ID {
		t.Fatalf("Saprolings = %+v, want one under Bob's control", saps)
	}
}

func TestFeralHydraAnyPlayerAddsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, cat := g.Seats[0], g.Seats[2]
	h := effectstest.ApaCreature("Feral Hydra", apaFeralHydra, 0, 0)
	h.Counters = map[string]int{game.CounterPlusOne: 2}
	hydra := effectstest.ApaPush(g, me.ID, me.ID, h)
	effectstest.ApaMana(cat, "C", "C", "C")
	effectstest.ApaActivate(t, g, cat, hydra, 0, game.ActivateAbilityParams{})
	c := effectstest.ApaLive(g, hydra)
	if c.Counters[game.CounterPlusOne] != 3 || c.CurrentPower() != 3 {
		t.Errorf("Feral Hydra counters %d power %d, want 3 and 3", c.Counters[game.CounterPlusOne], c.CurrentPower())
	}
}

// CR 701.10b: Casey gets +X/+0 for its power as the ability resolves,
// counters included; CR 701.10c: a negative power doubles downwards.
func TestCaseyJonesDoublesItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	cj := effectstest.ApaCreature("Casey Jones, Asphalt Hooligan", apaCaseyJones, 2, 2)
	cj.Counters = map[string]int{game.CounterPlusOne: 1}
	casey := effectstest.ApaPush(g, me.ID, me.ID, cj)
	if !game.HasKeyword(effectstest.ApaLive(g, casey), "double strike") {
		t.Errorf("Casey Jones lacks double strike")
	}
	effectstest.ApaMana(bob, "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, bob, casey, 0, game.ActivateAbilityParams{})
	if c := effectstest.ApaLive(g, casey); c.CurrentPower() != 6 || c.CurrentToughness() != 3 {
		t.Errorf("Casey %d/%d, want 6/3", c.CurrentPower(), c.CurrentToughness())
	}

	// Under another controller, so the legend rule leaves both alone.
	cat := g.Seats[2]
	neg := effectstest.ApaPush(g, cat.ID, cat.ID, effectstest.ApaCreature("Casey Jones, Asphalt Hooligan", apaCaseyJones, -1, 2))
	effectstest.ApaMana(bob, "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, bob, neg, 0, game.ActivateAbilityParams{})
	if got := effectstest.ApaLive(g, neg).PowerForComparison(); got != -2 {
		t.Errorf("a -1 power Casey doubled to %d, want -2", got)
	}
}

// Quicksilver Wall goes to its OWNER's hand, not the activator's or the
// controller's.
func TestQuicksilverWallReturnsToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	controller, owner, cat := g.Seats[0], g.Seats[1], g.Seats[2]
	wall := effectstest.ApaPush(g, owner.ID, controller.ID, effectstest.ApaCreature("Quicksilver Wall", apaQuicksilverWall, 1, 6))
	if !game.HasKeyword(effectstest.ApaLive(g, wall), "defender") {
		t.Errorf("Quicksilver Wall lacks defender")
	}
	ownerHand, catHand, ctrlHand := owner.Hand.Size(), cat.Hand.Size(), controller.Hand.Size()
	effectstest.ApaMana(cat, "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, cat, wall, 0, game.ActivateAbilityParams{})
	if effects.OnBattlefield(g, wall) {
		t.Fatalf("Quicksilver Wall is still on the battlefield")
	}
	if owner.Hand.Size() != ownerHand+1 || cat.Hand.Size() != catHand || controller.Hand.Size() != ctrlHand {
		t.Errorf("hands owner %d/%d activator %d/%d controller %d/%d",
			owner.Hand.Size(), ownerHand+1, cat.Hand.Size(), catHand, controller.Hand.Size(), ctrlHand)
	}
}

// Endbringer's Revel: any graveyard, the card's owner's hand, the
// activator's sorcery timing.
func TestEndbringersRevelReturnsToTheCardsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	revel := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Endbringer's Revel", OracleID: apaEndbringersRevel, TypeLine: "Enchantment"})
	dead := effectstest.PushGraveyardCardForTest(me, "Fallen Bear")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: dead}}

	effectstest.ApaMana(bob, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(bob.ID, revel, 0, game.ActivateAbilityParams{Targets: target}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("Bob activated Endbringer's Revel outside his main phase: err = %v", err)
	}
	bob.ManaPool = nil

	effectstest.AdvanceToMainOf(t, g, 1)
	myHand, bobHand := me.Hand.Size(), bob.Hand.Size()
	effectstest.ApaMana(bob, "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, bob, revel, 0, game.ActivateAbilityParams{Targets: target})
	if me.Hand.Size() != myHand+1 || bob.Hand.Size() != bobHand {
		t.Errorf("hands: owner %d (want %d), activator %d (want %d)", me.Hand.Size(), myHand+1, bob.Hand.Size(), bobHand)
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("the creature card is still in the graveyard")
	}
}

// A non-controller cannot activate the ability with too little mana in
// their own pool, whatever the controller has (CR 602.1a).
func TestAnyPlayerBatchCostIsTheActivators(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	elephant := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Vintara Elephant", apaVintaraElephant, 4, 3))
	effectstest.ApaMana(me, "C", "C", "C")
	effectstest.ApaMana(bob, "C", "C")
	err := g.ActivateCatalogAbility(bob.ID, elephant, 0, game.ActivateAbilityParams{Strict: true})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("Bob with {2} for a {3} cost: err = %v, want InsufficientManaError", err)
	}
	if me.ManaPool.Total() != 3 {
		t.Errorf("the controller's pool paid for Bob's activation")
	}
}
