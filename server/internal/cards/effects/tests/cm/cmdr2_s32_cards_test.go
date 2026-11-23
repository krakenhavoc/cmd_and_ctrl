package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s32_cards_test.go — Commander deck cards, batch 32: Clues as
// objects, detectives and exile-until-leaves (#5368).

const (
	cmdr2S32RansomNoteOracle   = "49c0aac5-d650-4ada-a194-3bf77f104bc1"
	cmdr2S32KelpOracle         = "1d7d9404-cad6-4c1a-bd47-7a31ebb2791c"
	cmdr2S32ArmedOracle        = "b401a490-d25f-4029-bb91-8999e140c6a1"
	cmdr2S32SophiaOracle       = "eaf8027a-017a-4471-ae7b-a372fd8e0ff3"
	cmdr2S32SereneSleuthOracle = "b0d639cb-71d4-454f-9dd6-89cd7a59b702"
	cmdr2S32SteamboatOracle    = "d97a0e41-9dd7-4466-afb7-3d8bd6cd0eed"
	cmdr2S32EscortOracle       = "c0ebc1df-51ea-413b-bb88-7666a53ad783"
	cmdr2S32ForagersOracle     = "0f5f6f3e-1b3c-43e2-b11d-bd454366845c"
)

func TestCmdr2S32EveryBuiltCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2S32RansomNoteOracle:   "Ransom Note",
		cmdr2S32KelpOracle:         "Tangletrove Kelp",
		cmdr2S32ArmedOracle:        "Armed with Proof",
		cmdr2S32SophiaOracle:       "Sophia, Dogged Detective",
		cmdr2S32SereneSleuthOracle: "Serene Sleuth",
		cmdr2S32SteamboatOracle:    "Foreboding Steamboat",
		cmdr2S32EscortOracle:       "Otherworldly Escort",
		cmdr2S32ForagersOracle:     "Bronzebeak Foragers",
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

// --- Ransom Note --------------------------------------------------------

func TestCmdr2S32RansomNoteSurveilsOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.CastHoldCreature(t, g, "Ransom Note", "Artifact — Clue", cmdr2S32RansomNoteOracle, 0, 0)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("entering did not surveil")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
}

func cmdr2S32Note(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	note := effectstest.B12Push(g, me.ID, "Ransom Note", "Artifact — Clue", cmdr2S32RansomNoteOracle, 0, 0)
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.MkmAC2AddMana(t, g, me, "{C}{C}")
	return g, me, opp, note
}

func TestCmdr2S32RansomNoteDrawMode(t *testing.T) {
	g, me, _, note := cmdr2S32Note(t)
	hand := me.Hand.Size()
	effectstest.P7Activate(t, g, me, note, 0, game.ActivateAbilityParams{Modes: []int{2}})
	if g.Battlefield.Contains(note) {
		t.Error("Ransom Note was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestCmdr2S32RansomNoteCloakMode(t *testing.T) {
	g, me, _, note := cmdr2S32Note(t)
	lib, bf := me.Library.Size(), len(g.Battlefield.Cards)
	effectstest.P7Activate(t, g, me, note, 0, game.ActivateAbilityParams{Modes: []int{0}})
	if got := lib - me.Library.Size(); got != 1 {
		t.Errorf("library lost %d cards, want 1", got)
	}
	// The Note left and the cloaked card entered.
	if got := len(g.Battlefield.Cards); got != bf {
		t.Errorf("battlefield has %d cards, want %d", got, bf)
	}
}

func TestCmdr2S32RansomNoteGoadMode(t *testing.T) {
	g, me, opp, note := cmdr2S32Note(t)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.P7Activate(t, g, me, note, 0, game.ActivateAbilityParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if c, ok := g.LookupCardForEffect(bear); !ok || c.LatestGoader() != me.ID {
		t.Fatal("the bear was not goaded")
	}
	if g.Battlefield.Contains(note) {
		t.Error("Ransom Note was not sacrificed")
	}
}

func TestCmdr2S32RansomNoteGoadNeedsACreature(t *testing.T) {
	g, me, _, note := cmdr2S32Note(t)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, note, 0, game.ActivateAbilityParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err == nil {
		t.Fatal("a non-creature was accepted as the goad target")
	}
	if !g.Battlefield.Contains(note) {
		t.Error("a refused activation sacrificed the Note")
	}
}

// --- Tangletrove Kelp ---------------------------------------------------

func TestCmdr2S32KelpAnimatesOtherCluesAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kelp := effectstest.B12Push(g, me.ID, "Tangletrove Kelp", "Artifact Creature — Clue Plant", cmdr2S32KelpOracle, 6, 6)
	mine := effectstest.PushToken(g, me.ID, effects.ClueToken())
	theirs := effectstest.PushToken(g, opp.ID, effects.ClueToken())
	if c := effectstest.TmtB1Card(t, g, mine); c.IsCreature() {
		t.Fatal("a Clue is a creature before combat")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.TmtB1Card(t, g, mine)
	if !c.IsCreature() || !c.HasSubtype("Plant") || !c.HasSubtype("Clue") {
		t.Errorf("my Clue: creature %v, types %v", c.IsCreature(), c.TypeLine)
	}
	if p, tough := effectstest.MkmB1PT(t, g, mine); p != 6 || tough != 6 {
		t.Errorf("my Clue is %d/%d, want 6/6", p, tough)
	}
	if effectstest.TmtB1Card(t, g, theirs).IsCreature() {
		t.Error("an opponent's Clue became a creature")
	}
	if p, _ := effectstest.MkmB1PT(t, g, kelp); p != 6 {
		t.Errorf("the Kelp is %d power, want 6", p)
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepCleanup)
	if effectstest.TmtB1Card(t, g, mine).IsCreature() {
		t.Error("the Clue is still a creature after the turn ended")
	}
}

func TestCmdr2S32KelpSacrificeDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AuraSeedLibrary(me, 3)
	kelp := effectstest.B12Push(g, me.ID, "Tangletrove Kelp", "Artifact Creature — Clue Plant", cmdr2S32KelpOracle, 6, 6)
	effectstest.MkmAC2AddMana(t, g, me, "{C}{C}")
	hand := me.Hand.Size()
	effectstest.P7Activate(t, g, me, kelp, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(kelp) {
		t.Error("the Kelp was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Armed with Proof ---------------------------------------------------

func TestCmdr2S32ArmedWithProofInvestigatesTwiceAndEquips(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Armed with Proof", "Enchantment", cmdr2S32ArmedOracle, 0, 0)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 2 {
		t.Fatalf("%d Clues on entering, want 2", got)
	}
	var clue uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Clue" {
			clue = c.InstanceID
			break
		}
	}
	if !effectstest.TmtB1Card(t, g, clue).HasSubtype("Equipment") {
		t.Fatal("a Clue is not an Equipment")
	}
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.BeEquip(t, g, me.ID, clue, bear, "Equip {2}")
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("equipped bear's power is %d, want 4", p)
	}
	// The enchantment leaves: the Clue stops being Equipment and falls off.
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == cmdr2S32ArmedOracle {
				if err := g.SacrificePermanentForEffect(c.InstanceID); err != nil {
					t.Fatalf("sacrifice: %v", err)
				}
				break
			}
		}
	})
	g.RunStateChecksForTest()
	if effectstest.TmtB1Card(t, g, clue).HasSubtype("Equipment") {
		t.Error("the Clue is still Equipment without Armed with Proof")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the bear's power is %d, want 2", p)
	}
}

// --- Sophia, Dogged Detective -------------------------------------------

func TestCmdr2S32SophiaCreatesTinyAndFeedsOnDogDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastHoldCreature(t, g, "Sophia, Dogged Detective", "Legendary Creature — Human Detective", cmdr2S32SophiaOracle, 3, 4)
	var tiny uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Tiny" && c.Controller == me.ID {
			tiny = c.InstanceID
		}
	}
	if tiny == uuid.Nil {
		t.Fatal("Sophia did not create Tiny")
	}
	c := effectstest.TmtB1Card(t, g, tiny)
	if !c.IsLegendary() || !c.HasSubtype("Dog") || !c.HasSubtype("Detective") {
		t.Errorf("Tiny is %q , want a legendary Dog Detective", c.TypeLine)
	}
	effectstest.AssertKeywords(t, g, tiny, "trample")
	if p, tough := effectstest.MkmB1PT(t, g, tiny); p != 2 || tough != 2 {
		t.Errorf("Tiny is %d/%d, want 2/2", p, tough)
	}

	// A Dog connecting makes a Food, then a Clue; a non-Dog does nothing.
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Count(g, me.ID, "Food") != 0 || effectstest.TmtB1Count(g, me.ID, "Clue") != 0 {
		t.Fatal("a Bear triggered Sophia")
	}
	effectstest.DealCombatDamageToPlayer(g, tiny, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Foods, want 1", got)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestCmdr2S32SophiaSacrificesAnArtifactTokenForDogCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sophia := effectstest.B12Push(g, me.ID, "Sophia, Dogged Detective", "Legendary Creature — Human Detective", cmdr2S32SophiaOracle, 3, 4)
	dog := effectstest.B12Creature(g, me.ID, "Good Dog", "Creature — Dog", 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	clue := effectstest.PushToken(g, me.ID, effects.ClueToken())
	effectstest.MkmAC2AddMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, sophia, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("a nontoken artifact paid the cost")
	}
	effectstest.P7Activate(t, g, me, sophia, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{clue}})
	if g.Battlefield.Contains(clue) {
		t.Error("the Clue was not sacrificed")
	}
	if p, _ := effectstest.MkmB1PT(t, g, dog); p != 2 {
		t.Errorf("the Dog's power is %d, want 2", p)
	}
	if p, _ := effectstest.MkmB1PT(t, g, sophia); p != 3 {
		t.Errorf("Sophia (not a Dog) has power %d, want 3", p)
	}
}

// --- Serene Sleuth ------------------------------------------------------

func TestCmdr2S32SereneSleuthInvestigatesOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Serene Sleuth", "Creature — Human Detective", cmdr2S32SereneSleuthOracle, 2, 2)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues on entering, want 1", got)
	}
}

func TestCmdr2S32SereneSleuthInvestigatesPerGoadedCreatureAndUngoads(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Serene Sleuth", "Creature — Human Detective", cmdr2S32SereneSleuthOracle, 2, 2)
	a := effectstest.B12Creature(g, me.ID, "Goaded A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Goaded B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Free C", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goaded", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.GoadForEffect(a, opp.ID)
		g.GoadForEffect(b, g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)].ID)
		g.GoadForEffect(theirs, me.ID)
	})
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 2 {
		t.Errorf("%d Clues, want 2 (one per goaded creature I control)", got)
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if effectstest.TmtB1Card(t, g, id).IsGoaded() {
			t.Errorf("%s is still goaded", id)
		}
	}
	if !effectstest.TmtB1Card(t, g, theirs).IsGoaded() {
		t.Error("an opponent's goaded creature stopped being goaded")
	}
}

func TestCmdr2S32EscortReturnsOnceAsASpiritWithFourCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	escort := effectstest.B12Push(g, me.ID, "Otherworldly Escort", "Creature — Human Detective", cmdr2S32EscortOracle, 4, 3)
	effectstest.Cmdr2S32Sacrifice(t, g, escort)
	if !g.Battlefield.Contains(escort) {
		t.Fatal("the Escort did not return")
	}
	c := effectstest.TmtB1Card(t, g, escort)
	if c.HasSubtype("Human") || !c.HasSubtype("Spirit") || !c.HasSubtype("Detective") || !c.IsCreature() {
		t.Errorf("the Escort came back as %q, want a Spirit Detective creature and no Human", c.TypeLine)
	}
	if got := effects.CounterCountOn(g, escort, "charge"); got != 4 {
		t.Errorf("%d charge counters, want 4", got)
	}
	// A Spirit that dies stays dead.
	effectstest.Cmdr2S32Sacrifice(t, g, escort)
	if g.Battlefield.Contains(escort) {
		t.Error("a Spirit Escort returned a second time")
	}
}

func TestCmdr2S32EscortDestroysACreatureThatDamagedYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	escort := effectstest.B12Push(g, me.ID, "Otherworldly Escort", "Creature — Human Detective", cmdr2S32EscortOracle, 4, 3)
	g.WithWriteLock(func() {
		g.AddCounterForEffect(escort, "charge", 2)
	})
	hitter := effectstest.B12Creature(g, opp.ID, "Their Hitter", "Creature — Ogre", 3, 3)
	idle := effectstest.B12Creature(g, opp.ID, "Their Idler", "Creature — Ogre", 3, 3)
	effectstest.MkmAC2AddMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, escort, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hitter}}}); err == nil {
		t.Fatal("a creature that dealt no damage to me was a legal target")
	}
	effectstest.DealCombatDamageToPlayer(g, hitter, me.ID, 3)
	_ = idle
	effectstest.P7Activate(t, g, me, escort, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hitter}}})
	if g.Battlefield.Contains(hitter) {
		t.Error("the creature that dealt damage to me survived")
	}
	if got := effects.CounterCountOn(g, escort, "charge"); got != 1 {
		t.Errorf("%d charge counters left, want 1", got)
	}
}

// --- Bronzebeak Foragers ------------------------------------------------

func TestCmdr2S32ForagersExilesPerOpponentUntilItLeavesAndBinsForX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp1 := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp2 := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		ManaCost: "{3}", Owner: opp1.ID, Controller: opp1.ID})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp2.ID, Controller: opp2.ID})
	land := effectstest.B12Permanent(g, opp1.ID, "Their Land", "Land")
	foragers := effectstest.CastHoldCreature(t, g, "Bronzebeak Foragers", "Creature — Dinosaur", cmdr2S32ForagersOracle, 3, 4)
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(bear) {
		t.Fatal("the Foragers did not exile the chosen permanents")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("an unchosen land was exiled")
	}
	effectstest.MkmAC2AddMana(t, g, me, "{C}{C}{W}")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, foragers, 0, game.ActivateAbilityParams{XValue: 2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err == nil {
		t.Fatal("a mana value 3 card was a legal target for X = 2")
	}
	effectstest.P7Activate(t, g, me, foragers, 0, game.ActivateAbilityParams{XValue: 2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if !opp2.Graveyard.Contains(bear) {
		t.Error("the Bear is not in its owner's graveyard")
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	// The Foragers leave: the card still held comes back, the binned one does not.
	effectstest.Cmdr2S32Sacrifice(t, g, foragers)
	if got := effectstest.TmtB1Count(g, opp1.ID, "Their Rock"); got != 1 {
		t.Errorf("%d Rocks on the battlefield after the Foragers left, want 1", got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear returned from the graveyard")
	}
}

// --- Foreboding Steamboat -----------------------------------------------

func TestCmdr2S32SteamboatExilesTwoPerPlayerAndBinsOneOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var mine, theirs []uuid.UUID
	for i := 0; i < 3; i++ {
		mine = append(mine, effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2))
	}
	tokenBear := effectstest.PushToken(g, me.ID, effects.TokenCard("2/2 white Soldier"))
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 2; i++ {
		theirs = append(theirs, effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2))
	}
	boat := effectstest.CastHoldCreature(t, g, "Foreboding Steamboat", "Artifact — Vehicle", cmdr2S32SteamboatOracle, 5, 7)
	// each player was asked in turn; answer each with the first two candidates.
	answered := 0
	for i := 0; i < 8; i++ {
		var c *game.PendingChoice
		g.ReadSnapshot(func() {
			for _, pc := range g.PendingChoices {
				if pc != nil && pc.Kind == game.PendingChoiceOwnPermanents {
					c = pc
					return
				}
			}
		})
		if c == nil {
			break
		}
		pool := mine
		if c.Chooser == opp.ID {
			pool = theirs
		}
		if err := g.ResolveOwnPermanents(c.ID, c.Chooser, pool[:2]); err != nil {
			t.Fatalf("answering: %v", err)
		}
		answered++
		effectstest.PassPriorityAroundTable(t, g)
	}
	if answered < 2 {
		t.Fatalf("only %d players were asked", answered)
	}
	for _, id := range append(append([]uuid.UUID(nil), mine[:2]...), theirs...) {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if !g.Battlefield.Contains(mine[2]) || !g.Battlefield.Contains(tokenBear) {
		t.Error("an unchosen creature or a token was exiled")
	}
	// Crew it with the creature left, then attack: a card goes to its
	// owner's graveyard and I investigate.
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == boat {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, boat, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{mine[2]}}); err != nil {
		t.Fatalf("Crew 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, boat)
	effectstest.PassPriorityAroundTable(t, g)
	inYards := 0
	for _, id := range append(append([]uuid.UUID(nil), mine[:2]...), theirs...) {
		if me.Graveyard.Contains(id) || opp.Graveyard.Contains(id) {
			inYards++
		}
	}
	if len(g.PendingChoices) > 0 {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards && len(c.ChooseCards) > 0 {
				if err := g.ResolveChooseCards(c.ID, me.ID, c.ChooseCards[:1]); err != nil {
					t.Fatalf("ResolveChooseCards: %v", err)
				}
				break
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		inYards = 0
		for _, id := range append(append([]uuid.UUID(nil), mine[:2]...), theirs...) {
			if me.Graveyard.Contains(id) || opp.Graveyard.Contains(id) {
				inYards++
			}
		}
	}
	if inYards != 1 {
		t.Errorf("%d exiled cards in graveyards, want 1", inYards)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
	// The Steamboat leaves: the three still held return, the binned one does not.
	effectstest.Cmdr2S32Sacrifice(t, g, boat)
	back := effectstest.TmtB1Count(g, me.ID, "Mine") - 1 + effectstest.TmtB1Count(g, opp.ID, "Theirs")
	if back != 3 {
		t.Errorf("%d creatures returned, want 3", back)
	}
}
