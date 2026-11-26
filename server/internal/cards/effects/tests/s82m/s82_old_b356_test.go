package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b356_test.go — S82 older sets, batch 356: Magic 2015 token makers
// and enters/dies triggers (#5556).

const (
	oldB356GoblinKaboomist = "5209e1da-aeed-45b3-9640-c2db121d6612"
	oldB356BroodKeeper     = "c123f45d-a8db-405f-93ef-7227e826790c"
	oldB356Twingrove       = "cbaf8e87-da5b-4120-afec-8fe2db55b835"
	oldB356CoralBarrier    = "f03474ce-b687-489f-add3-a82d77e18355"
	oldB356SpiritBonds     = "f41ea1b5-2c31-43fa-b142-244add1efd35"
	oldB356Phytotitan      = "2541e0cd-2cc1-4e47-94e3-030c1cb1a6bd"
	oldB356Skirmisher      = "6b142a4a-769b-4e1b-aae6-ba6e11f35836"
	oldB356InvasiveSpecies = "9c4c478c-6f0c-430a-91eb-4125a73d03a7"
	oldB356FoundryDenizen  = "e8148616-9f0b-4d4e-8d6d-18251a84e5e8"
	oldB356InvisibilityAu  = "de26b0c6-dfb7-45a8-9d7f-f8d45522d675"
)

func TestOldB356EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB356GoblinKaboomist: "Goblin Kaboomist", oldB356BroodKeeper: "Brood Keeper",
		oldB356Twingrove: "Kalonian Twingrove", oldB356CoralBarrier: "Coral Barrier",
		oldB356SpiritBonds: "Spirit Bonds", oldB356Phytotitan: "Phytotitan",
		oldB356Skirmisher: "Kinsbaile Skirmisher", oldB356InvasiveSpecies: "Invasive Species",
		oldB356FoundryDenizen: "Foundry Street Denizen",
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

// oldB356Cast puts a creature with the given colours in hand and casts it.
func oldB356Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tough int, colors ...string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Power: p, Toughness: tough, Colors: colors, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB356FoundryStreetDenizenGrowsOnlyForOtherRedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	denizen := effectstest.B12Push(g, me.ID, "Foundry Street Denizen", "Creature — Goblin Warrior", oldB356FoundryDenizen, 1, 1)
	effectstest.AdvanceToMain(t, g)
	oldB356Cast(t, g, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, denizen); p != 1 {
		t.Errorf("power %d after a green creature entered, want 1", p)
	}
	oldB356Cast(t, g, "Red Bear", "Creature — Bear", "", 2, 2, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, denizen); p != 2 {
		t.Errorf("power %d after a red creature entered, want 2", p)
	}
	oldB356Cast(t, g, "Another Red", "Creature — Bear", "", 2, 2, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, denizen); p != 3 {
		t.Errorf("power %d after a second red creature entered, want 3", p)
	}
}

func TestOldB356KinsbaileSkirmisherPumpsATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	oldB356Cast(t, g, "Kinsbaile Skirmisher", "Creature — Kithkin Soldier", oldB356Skirmisher, 2, 2, "W")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, friend)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, friend); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
	_ = opp
}

func TestOldB356CoralBarrierMakesASquidWithIslandwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB356Cast(t, g, "Coral Barrier", "Creature — Wall", oldB356CoralBarrier, 1, 3, "U")
	effectstest.PassPriorityAroundTable(t, g)
	squid := effectstest.FindBattlefieldByName(g, "Squid")
	if squid == uuid.Nil {
		t.Fatal("no Squid token")
	}
	if p, tough := effectstest.PtOf(t, g, squid); p != 1 || tough != 1 {
		t.Errorf("the Squid is %d/%d, want 1/1", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, squid, "islandwalk") {
		t.Error("the Squid has no islandwalk")
	}
}

func TestOldB356InvasiveSpeciesMustReturnAnotherPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	species := oldB356Cast(t, g, "Invasive Species", "Creature — Insect", oldB356InvasiveSpecies, 3, 3, "G")
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no permanent choice")
	}
	if effectstest.HasID(pick.ChooseCards, species) {
		t.Error("the Species itself was offered")
	}
	if !effectstest.HasID(pick.ChooseCards, rock) || !effectstest.HasID(pick.ChooseCards, bear) {
		t.Errorf("offered %v, want the artifact and the creature", pick.ChooseCards)
	}
	// Declining is not an option: the pick needs one.
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err == nil {
		t.Error("the return was declined, but it isn't a may")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Error("the picked creature was not returned to hand")
	}
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(species) {
		t.Error("something that was not picked left the battlefield")
	}
}

func TestOldB356InvasiveSpeciesAloneReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	species := oldB356Cast(t, g, "Invasive Species", "Creature — Insect", oldB356InvasiveSpecies, 3, 3, "G")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(species) {
		t.Error("the Species returned itself")
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, g.Seats[g.Turn.ActiveSeat].ID) != nil {
		t.Error("a pick was offered with nothing to return")
	}
}

func TestOldB356KalonianTwingroveCountsForestsAndMakesAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 3; i++ {
		effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	}
	effectstest.B31Push(g, opp.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.B31Push(g, me.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	grove := oldB356Cast(t, g, "Kalonian Twingrove", "Creature — Treefolk Warrior", oldB356Twingrove, 0, 0, "G")
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, grove); p != 3 || tough != 3 {
		t.Errorf("Twingrove is %d/%d, want 3/3 (my three Forests)", p, tough)
	}
	token := effectstest.FindBattlefieldByName(g, "Treefolk Warrior")
	if token == uuid.Nil {
		t.Fatal("no Treefolk Warrior token")
	}
	if p, tough := effectstest.PtOf(t, g, token); p != 3 || tough != 3 {
		t.Errorf("the token is %d/%d, want 3/3", p, tough)
	}
	effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	if p, _ := effectstest.PtOf(t, g, grove); p != 4 {
		t.Errorf("Twingrove power %d after a fourth Forest, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, token); p != 4 {
		t.Errorf("the token's power %d after a fourth Forest, want 4", p)
	}
}

func TestOldB356BroodKeeperMakesADragonForEachAuraAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	keeper := effectstest.B12Push(g, me.ID, "Brood Keeper", "Creature — Human Shaman", oldB356BroodKeeper, 2, 3)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	countDragons := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Dragon" && c.Controller == me.ID {
				n++
			}
		}
		return n
	}
	// An Aura attached to another creature does nothing.
	oldB355AttachB356(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := countDragons(); got != 0 {
		t.Fatalf("%d Dragons after an Aura went on another creature, want 0", got)
	}
	aura := oldB355AttachB356(t, g, me.ID, keeper)
	effectstest.PassPriorityAroundTable(t, g)
	if got := countDragons(); got != 1 {
		t.Fatalf("%d Dragons after an Aura went on the Keeper, want 1", got)
	}
	dragon := effectstest.FindBattlefieldByName(g, "Dragon")
	if p, tough := effectstest.PtOf(t, g, dragon); p != 2 || tough != 2 {
		t.Errorf("the Dragon is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("the Dragon has no flying")
	}
	// Its {R} ability: +1/+0.
	effectstest.FloatForTest(g, me, "R")
	effectstest.B16Activate(t, g, me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, dragon); p != 3 {
		t.Errorf("the Dragon's power %d after {R}, want 3", p)
	}
	_ = aura
}

// oldB355AttachB356 puts an Invisibility on host, as an Aura becoming
// attached to it.
func oldB355AttachB356(t *testing.T, g *game.Game, owner, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.PushCatalogPermanent(g, owner, "Invisibility", effectstest.AuraTypeLine, oldB356InvisibilityAu, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	return aura
}

func TestOldB356PhytotitanReturnsTappedAtItsOwnersNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	titan := effectstest.B12Push(g, me.ID, "Phytotitan", "Creature — Plant Elemental", oldB356Phytotitan, 7, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, titan)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(titan) {
		t.Fatal("the Titan is still on the battlefield")
	}
	// The next upkeep is the owner's, a full round later; not the next player's.
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	effectstest.AdvanceToUpkeepOfSeat(t, g, next)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(titan) {
		t.Error("the Titan came back at another player's upkeep")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, g.Turn.ActiveSeat+0)
	for i := 0; i < len(g.Seats) && !g.Battlefield.Contains(titan); i++ {
		effectstest.AdvanceToUpkeepOfSeat(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(titan) {
		t.Fatal("the Titan never came back")
	}
	if !effectstest.IsTapped(g, titan) {
		t.Error("the Titan did not return tapped")
	}
}

func TestOldB356SpiritBondsPaysWForASpiritAndSacrificesOneForIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bonds := effectstest.B12Push(g, me.ID, "Spirit Bonds", "Enchantment", oldB356SpiritBonds, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "W")
	oldB356Cast(t, g, "Bear", "Creature — Bear", "", 2, 2, "G")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	spirit := effectstest.FindBattlefieldByName(g, "Spirit")
	if spirit == uuid.Nil {
		t.Fatal("no Spirit token after paying {W}")
	}
	if !effectstest.HasEffectiveKeyword(t, g, spirit, "flying") {
		t.Error("the Spirit has no flying")
	}
	// A token entering does not trigger it; declining the payment makes none.
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("a prompt is open after the token entered: %d", len(g.PendingChoices))
	}
	effectstest.FloatForTest(g, me, "W")
	oldB356Cast(t, g, "Second Bear", "Creature — Bear", "", 2, 2, "G")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	spirits := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" {
			spirits++
		}
	}
	if spirits != 1 {
		t.Errorf("%d Spirits after declining the second payment, want 1", spirits)
	}
	// {1}{W}, sacrifice a Spirit: target non-Spirit creature gains indestructible.
	bear := effectstest.B12Creature(g, me.ID, "Target", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CW")
	if err := g.ActivateCatalogAbility(me.ID, bonds, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(spirit), SacrificeIDs: []uuid.UUID{spirit}}); err == nil {
		t.Error("a Spirit was targeted")
	}
	effectstest.B16Activate(t, g, me.ID, bonds, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear), SacrificeIDs: []uuid.UUID{spirit}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spirit) {
		t.Error("the Spirit was not sacrificed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the target did not gain indestructible")
	}
}

func TestOldB356GoblinKaboomistMakesALandMineAndMayHurtItself(t *testing.T) {
	sawWin, sawLoss := false, false
	for i := 0; i < 64 && !(sawWin && sawLoss); i++ {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.SeatsForCombat(g)
		kaboom := effectstest.B12Push(g, me.ID, "Goblin Kaboomist", "Creature — Goblin Warrior", oldB356GoblinKaboomist, 1, 2)
		effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
		effectstest.PassPriorityAroundTable(t, g)
		flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil {
			t.Fatal("no coin call")
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
			t.Fatalf("coin: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		mine := effectstest.FindBattlefieldByName(g, "Land Mine")
		if mine == uuid.Nil {
			t.Fatal("no Land Mine token")
		}
		flips := effectstest.RandomEvents(g, game.EventFlipCoin)
		won := len(flips) > 0 && flips[len(flips)-1].Won
		if won {
			sawWin = true
			if !g.Battlefield.Contains(kaboom) {
				t.Error("won the flip, but the Kaboomist died")
			}
		} else {
			sawLoss = true
			if g.Battlefield.Contains(kaboom) {
				t.Error("lost the flip, but the Kaboomist survived 2 damage on a 1/2")
			}
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("never saw both outcomes of the flip (win %v, loss %v)", sawWin, sawLoss)
	}
}

func TestOldB356LandMineDamagesAnAttackingCreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Land Mine", TypeLine: "Token Artifact", TokenKey: game.TokenKey("land-mine"),
		Owner: opp.ID, Controller: opp.ID,
	})
	walker := effectstest.B12Creature(g, me.ID, "Walker", "Creature — Bear", 2, 3)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Keywords: []string{"flying"}, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{walker, flier} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.FloatForTest(g, opp, "R")
	if err := g.ActivateCatalogAbility(opp.ID, mine, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(flier)}); err == nil {
		t.Error("the Land Mine targeted a creature with flying")
	}
	effectstest.B16Activate(t, g, opp.ID, mine, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(walker)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the Land Mine was not sacrificed")
	}
	if got := effectstest.DamageOn(t, g, walker); got != 2 {
		t.Errorf("the attacker has %d damage, want 2", got)
	}
}
