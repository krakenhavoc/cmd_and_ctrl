package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b398_cards_test.go — Duel Decks: Heroes vs. Monsters, S82 older sets
// batch 398 (#5562): Ghor-Clan Savage, Skarrgan Firebird, Crowned Ceratok,
// Somberwald Vigilante, Shower of Sparks, Conquering Manticore, Skarrg, the
// Rage Pits and Deus of Calamity, through the real cast, activation, combat
// and trigger paths.

const (
	ddl398Savage    = "28f35fa1-befa-40b4-95ab-30bb53d339ab"
	ddl398Firebird  = "42a96276-fddc-435e-bd47-04f9c6001ec7"
	ddl398Ceratok   = "5cdeb1fe-9ce3-4088-a35c-dcfb762f76d8"
	ddl398Vigilante = "31c77036-8627-476a-996b-2a6b005c9e12"
	ddl398Shower    = "a43500ed-4aa6-49de-8a6e-a1781301b108"
	ddl398Manticore = "140b58b9-7e82-4feb-99d7-5e8b707b845c"
	ddl398Skarrg    = "92bac34e-2045-4331-842f-185711c1ac56"
	ddl398Deus      = "f263f9d0-2ca9-4d3f-9b41-c67bdc73281a"
)

func TestS82OldB398EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ddl398Savage: "Ghor-Clan Savage", ddl398Firebird: "Skarrgan Firebird", ddl398Ceratok: "Crowned Ceratok",
		ddl398Vigilante: "Somberwald Vigilante", ddl398Shower: "Shower of Sparks",
		ddl398Manticore: "Conquering Manticore", ddl398Skarrg: "Skarrg, the Rage Pits", ddl398Deus: "Deus of Calamity",
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

// --- Ghor-Clan Savage, Skarrgan Firebird: bloodthirst 3 ---

func TestS82OldB398BloodthirstThreeNeedsAnOpponentDealtDamage(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
	}{
		{"Ghor-Clan Savage", "Creature — Centaur Berserker", ddl398Savage, 2, 3},
		{"Skarrgan Firebird", "Creature — Phoenix", ddl398Firebird, 3, 3},
	} {
		t.Run(tc.name+" with damage", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			_, opp := effectstest.GptB1Seats(g)
			effectstest.GptB1DamagePlayer(g, opp.ID, 1)
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 3 {
				t.Errorf("%d +1/+1 counters, want 3", n)
			}
			if p, tough := effectstest.PtOf(t, g, id); p != tc.p+3 || tough != tc.tough+3 {
				t.Errorf("%d/%d, want %d/%d", p, tough, tc.p+3, tc.tough+3)
			}
		})
		t.Run(tc.name+" without damage", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 0 {
				t.Errorf("%d +1/+1 counters with no opponent damaged, want 0", n)
			}
		})
	}
}

func TestS82OldB398BloodthirstIgnoresDamageToItsOwnController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	effectstest.GptB1DamagePlayer(g, me.ID, 3)
	id := effectstest.CastHoldCreature(t, g, "Ghor-Clan Savage", "Creature — Centaur Berserker", ddl398Savage, 2, 3)
	if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 0 {
		t.Errorf("damage to its own controller gave %d counters", n)
	}
}

// --- Skarrgan Firebird: the graveyard return ---

func firebird398InGraveyard(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Skarrgan Firebird", TypeLine: "Creature — Phoenix", OracleID: ddl398Firebird,
		ManaCost: "{4}{R}{R}", Colors: []string{"R"}, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	return id
}

func TestS82OldB398FirebirdReturnsFromTheGraveyardOnlyAfterAnOpponentWasDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	id := firebird398InGraveyard(me)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RRR")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated before any opponent was dealt damage")
	}
	effectstest.GptB1DamagePlayer(g, opp.ID, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate after the damage: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Error("the Firebird did not return to hand")
	}
}

func TestS82OldB398FirebirdNeedsRedManaToReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	id := firebird398InGraveyard(me)
	effectstest.AdvanceToMain(t, g)
	effectstest.GptB1DamagePlayer(g, opp.ID, 1)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with two red mana")
	}
}

// --- Crowned Ceratok ---

func TestS82OldB398CeratokGivesTrampleToCreaturesWithAPlusOneCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Crowned Ceratok", "Creature — Rhino", ddl398Ceratok, 4, 3)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	counted := effectstest.B12Creature(g, me.ID, "Counted", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(counted, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1)
	})
	if effectstest.HasEffectiveKeyword(t, g, plain, "trample") {
		t.Error("a creature with no counter has trample")
	}
	if !effectstest.HasEffectiveKeyword(t, g, counted, "trample") {
		t.Error("a creature with a +1/+1 counter has no trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature with a counter has trample")
	}
}

// --- Somberwald Vigilante ---

func TestS82OldB398VigilantePingsEachCreatureThatBlocksIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vig := effectstest.B12Push(g, me.ID, "Somberwald Vigilante", "Creature — Human Warrior", ddl398Vigilante, 1, 1)
	chump := effectstest.B12Creature(g, opp.ID, "Chump", "Creature — Bear", 1, 1)
	sturdy := effectstest.B12Creature(g, opp.ID, "Sturdy", "Creature — Bear", 3, 3)
	effectstest.BrAttack(t, g, vig)
	for _, b := range []uuid.UUID{chump, sturdy} {
		if err := g.DeclareBlocker(b, vig); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(chump) {
		t.Error("the 1/1 blocker survived the Vigilante's trigger")
	}
	if !g.Battlefield.Contains(sturdy) {
		t.Error("the 3/3 blocker died to a 1-damage trigger")
	}
}

func TestS82OldB398VigilanteDoesNothingWhenUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vig := effectstest.B12Push(g, me.ID, "Somberwald Vigilante", "Creature — Human Warrior", ddl398Vigilante, 1, 1)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 1, 1)
	effectstest.BrAttack(t, g, vig)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature that did not block was damaged")
	}
}

// --- Shower of Sparks ---

func TestS82OldB398ShowerOfSparksHitsTheCreatureAndThePlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Shower of Sparks", "Instant", ddl398Shower,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, effectstest.PbPlayer(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AppInZone(g, bear, game.ZoneBattlefield) {
		t.Error("the 1/1 survived")
	}
	if opp.Life != before-1 {
		t.Errorf("life %d → %d, want 1 damage", before, opp.Life)
	}
}

func TestS82OldB398ShowerOfSparksStillHitsThePlayerWhenTheCreatureIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Shower of Sparks", "Instant", ddl398Shower,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, effectstest.PbPlayer(opp.ID)})
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d → %d, want 1 damage", before, opp.Life)
	}
}

// --- Conquering Manticore ---

func TestS82OldB398ManticoreStealsAnOpposingCreatureUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.B16Tap(g, ogre)
	effectstest.CastHoldCreatureNoResolve(t, g, "Conquering Manticore", "Creature — Manticore", ddl398Manticore, 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, ogre) != me.ID {
		t.Fatal("control of the Ogre did not change")
	}
	if c, _ := effectstest.CardOnBattlefield(g, ogre); c.Tapped {
		t.Error("the stolen Ogre is still tapped")
	}
	if !effectstest.P1g2HasKeyword(g, ogre, "haste") {
		t.Error("the stolen Ogre has no haste")
	}
	for g.Turn.Step != game.StepCleanup {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, ogre) != opp.ID {
		t.Error("the Ogre did not go home at end of turn")
	}
}

func TestS82OldB398ManticoreCannotTargetYourOwnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastHoldCreatureNoResolve(t, g, "Conquering Manticore", "Creature — Manticore", ddl398Manticore, 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		// One legal target is chosen for the player: it can only be the Ogre.
		if effectstest.ControllerOf(t, g, mine) != me.ID {
			t.Error("control of my creature changed")
		}
		return
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("my own creature was a legal target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: ogre}); err != nil {
		t.Errorf("the opposing creature was refused: %v", err)
	}
}

// --- Skarrg, the Rage Pits ---

func TestS82OldB398SkarrgTapsForColorlessAndPumpsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Skarrg, the Rage Pits", "Land", ddl398Skarrg, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.ActivateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})
	if got := effectstest.PoolCount(me, "C"); got != 1 {
		t.Errorf("%d colorless mana, want 1", got)
	}
	effectstest.UntapForTest(g, land)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "RG")
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the {T} cost was not paid")
	}
}

func TestS82OldB398SkarrgNeedsBothRedAndGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Skarrg, the Rage Pits", "Land", ddl398Skarrg, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("activated without green mana")
	}
}

// --- Deus of Calamity ---

func deus398(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return effectstest.B27Push(g, me.ID, "Deus of Calamity", "Creature — Spirit Avatar", ddl398Deus, "{R/G}{R/G}{R/G}{R/G}{R/G}", 6, 6, "R", "G")
}

func TestS82OldB398DeusDestroysALandOfTheOpponentItDamagedForSixOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	deus := deus398(t, g, me)
	oppLand := effectstest.PushLand(g, opp.ID, "Forest")
	myLand := effectstest.PushLand(g, me.ID, "Mountain")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(deus, opp.ID, 6) })
	effectstest.PickTriggerTarget(t, g, me.ID, oppLand)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(oppLand) {
		t.Error("the opponent's land survived")
	}
	if !g.Battlefield.Contains(myLand) {
		t.Error("my own land was destroyed")
	}
}

func TestS82OldB398DeusDoesNothingForFiveDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	deus := deus398(t, g, me)
	oppLand := effectstest.PushLand(g, opp.ID, "Forest")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(deus, opp.ID, 5) })
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("five damage raised the trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(oppLand) {
		t.Error("five damage destroyed a land")
	}
}

func TestS82OldB398DeusIgnoresSixDamageDealtToItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.GptB1Seats(g)
	deus := deus398(t, g, me)
	myLand := effectstest.PushLand(g, me.ID, "Mountain")
	oppLand := effectstest.PushLand(g, opp.ID, "Forest")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(deus, me.ID, 6) })
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("damage to its own controller raised the trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(myLand) || !g.Battlefield.Contains(oppLand) {
		t.Error("damage to its own controller destroyed a land")
	}
}
