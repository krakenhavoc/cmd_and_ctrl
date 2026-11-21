package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rad_counter_batch1_test.go pins the first batch of the remaining
// rad-counter cards (#2809): The Ghoul, Gunslinger; Vault 12: The
// Necropolis; Acquired Mutation; Nightkin Ambusher; Mariposa Military
// Base; Nuka-Nuke Launcher. The counters' own trigger (CR 728.1) is
// pinned in game/rad_counters_test.go and the earlier cards in
// rad_counter_cards_test.go.

const (
	radGhoulGunslinger  = "ddfa5c7d-bc47-43e0-9607-b3baa8e6a17d"
	radVault12          = "bc9d75e7-1f64-4376-8e19-f6622ffbeb28"
	radAcquiredMutation = "9ecd53b0-4551-4505-b5d2-02ae49d4cbdb"
	radNightkinAmbusher = "2808d652-907d-4071-b386-814507a1549e"
	radMariposaBase     = "f1e03d99-024a-430b-9342-ffd2268bd103"
	radNukaNukeLauncher = "7fb88a5d-9b72-4da9-ade4-d09cadd7e1cb"
)

// radGive gives a player n rad counters outside any card, as a setup.
func radGive(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(p.ID, game.CounterRad, n); err != nil {
			t.Fatalf("give %s rad counters: %v", p.Name, err)
		}
	})
}

// radLandLibraries fills every library with lands, so the rad counters'
// own mill (CR 728.1) costs nobody a life or a counter while a test
// walks through a precombat main phase.
func radLandLibraries(g *game.Game) {
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Library.Cards = nil
			for i := 0; i < 30; i++ {
				p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: p.ID, Controller: p.ID})
			}
		}
	})
}

// radPickPlayer answers the pending pick_target prompt of `chooser`
// with a player.
func radPickPlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	p := effectstest.LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// radAnswerMay answers the "may" prompt of an optional replacement.
func radAnswerMay(t *testing.T, g *game.Game, player uuid.UUID, yes bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement && c.Chooser == player {
			if err := g.ResolveOptionalReplacement(c.ID, player, yes); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			return
		}
	}
	t.Fatalf("no may prompt for %s", player)
}

// radSettleTurnStart settles the triggers of a precombat main phase: a
// player with rad counters has CR 728.1's trigger beside any chapter, so
// the table first asks the controller for an order.
func radSettleTurnStart(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == player {
			order := append([]uuid.UUID(nil), c.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(c.ID, player, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- The Ghoul, Gunslinger ----------------------------------------

// The Ghoul dying aims two rad counters at a player of its controller's
// choice; an opponent gets them and no Treasure comes.
func TestGhoulGunslingerOwnDeathIrradiatesAChosenOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "The Ghoul, Gunslinger", OracleID: radGhoulGunslinger, TypeLine: "Legendary Creature — Zombie Mutant Rogue", Power: 2, Toughness: 3})
	effectstest.RadDestroy(t, g, ghoul)
	radPickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(opp) != 2 || effectstest.RadOf(me) != 0 {
		t.Errorf("rad: opponent %d (want 2), me %d (want 0)", effectstest.RadOf(opp), effectstest.RadOf(me))
	}
	if n := effectstest.CountTreasuresControlledBy(g, me.ID); n != 0 {
		t.Errorf("%d Treasures; the Treasure is only for a trigger aimed at its controller", n)
	}
}

// Aimed at its own controller, the trigger gives them the counters and
// a Treasure.
func TestGhoulGunslingerAimedAtYouMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "The Ghoul, Gunslinger", OracleID: radGhoulGunslinger, TypeLine: "Legendary Creature — Zombie Mutant Rogue", Power: 2, Toughness: 3})
	effectstest.RadDestroy(t, g, ghoul)
	radPickPlayer(t, g, me.ID, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(me) != 2 {
		t.Errorf("rad %d, want 2", effectstest.RadOf(me))
	}
	if n := effectstest.CountTreasuresControlledBy(g, me.ID); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
}

// Another creature of yours dying triggers it only when it is a
// nontoken Zombie or Mutant.
func TestGhoulGunslingerWatchesNontokenZombiesAndMutantsYouControl(t *testing.T) {
	cases := []struct {
		name     string
		typeLine string
		theirs   bool
		fires    bool
	}{
		{"a nontoken Zombie", "Creature — Zombie", false, true},
		{"a nontoken Mutant", "Creature — Mutant Warrior", false, true},
		{"a nontoken Human", "Creature — Human", false, false},
		{"a Zombie token", "Token Creature — Zombie", false, false},
		{"a Mutant token", "Token Creature — Zombie Mutant", false, false},
		{"an opponent's Zombie", "Creature — Zombie", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "The Ghoul, Gunslinger", OracleID: radGhoulGunslinger, TypeLine: "Legendary Creature — Zombie Mutant Rogue", Power: 2, Toughness: 3})
			owner := me
			if tc.theirs {
				owner = opp
			}
			fodder := effectstest.ApaPush(g, owner.ID, owner.ID, game.Card{Name: "Fodder", TypeLine: tc.typeLine, Power: 1, Toughness: 1})
			effectstest.RadDestroy(t, g, fodder)
			fired := effectstest.LatestPickTarget(g, me.ID) != nil
			if fired != tc.fires {
				t.Fatalf("trigger fired = %v, want %v", fired, tc.fires)
			}
			if fired {
				radPickPlayer(t, g, me.ID, opp.ID)
				effectstest.PassPriorityAroundTable(t, g)
				if effectstest.RadOf(opp) != 2 {
					t.Errorf("rad %d, want 2", effectstest.RadOf(opp))
				}
			}
		})
	}
}

// The "you control" and the creature types are read as the creature
// last existed (CR 603.10a): a Zombie only through a lord's grant still
// counts, and the Ghoul's controller is the one who is asked.
func TestGhoulGunslingerUsesTheTypesTheCreatureDiedWith(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "The Ghoul, Gunslinger", OracleID: radGhoulGunslinger, TypeLine: "Legendary Creature — Zombie Mutant Rogue", Power: 2, Toughness: 3})
	// A stolen Zombie dies under the thief's control, and counts for the thief.
	stolen := effectstest.ApaPush(g, opp.ID, me.ID, game.Card{Name: "Stolen Zombie", TypeLine: "Creature — Zombie", Power: 1, Toughness: 1})
	effectstest.RadDestroy(t, g, stolen)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("a Zombie that died under my control did not trigger")
	}
}

// --- Acquired Mutation --------------------------------------------

func TestAcquiredMutationPumpsGoadsAndIrradiatesTheDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := effectstest.B12Push(g, me.ID, "Acquired Mutation", "Enchantment — Aura", radAcquiredMutation, 0, 0)
	bear := effectstest.BrCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BrAttach(t, g, aura, bear)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4 (+2/+2)", p)
	}
	if c := effectstest.GaLayered(t, g, bear); !c.Goaded() {
		t.Error("the enchanted creature is not goaded")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(opp) != 2 {
		t.Errorf("defending player has %d rad counters, want 2", effectstest.RadOf(opp))
	}
	for _, p := range g.Seats {
		if p != opp && effectstest.RadOf(p) != 0 {
			t.Errorf("%s has %d rad counters, want 0", p.Name, effectstest.RadOf(p))
		}
	}
}

// The Aura is the source: when an opponent controls the enchanted
// creature, they are goaded away from the Aura's controller, and the
// defending player is whoever they pick instead.
func TestAcquiredMutationOnAnOpponentsCreatureIrradiatesWhoeverItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	aura := effectstest.B12Push(g, me.ID, "Acquired Mutation", "Enchantment — Aura", radAcquiredMutation, 0, 0)
	bear := effectstest.BrCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BrAttach(t, g, aura, bear)
	// The goad forbids attacking the Aura's controller while another
	// player can be attacked.
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(bear, me.ID); err == nil {
		t.Fatal("a goaded creature attacked the Aura's controller")
	}
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(third) != 2 || effectstest.RadOf(me) != 0 || effectstest.RadOf(opp) != 0 {
		t.Errorf("rad: attacked %d (want 2), Aura's controller %d, creature's controller %d (want 0, 0)", effectstest.RadOf(third), effectstest.RadOf(me), effectstest.RadOf(opp))
	}
}

// --- Nightkin Ambusher --------------------------------------------

func TestNightkinAmbusherEntersAndGivesAPlayerFourRadCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.EnterFromHand(t, g, me.ID, "Nightkin Ambusher", "Creature — Mutant Warrior", radNightkinAmbusher)
	radPickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(opp) != 4 {
		t.Errorf("rad %d, want 4", effectstest.RadOf(opp))
	}
}

// "Can't be blocked as long as defending player has a rad counter" is
// judged on the player this attack is aimed at, when blocks are made.
func TestNightkinAmbusherCantBeBlockedWhileTheDefenderHasARadCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ambusher := effectstest.B12Push(g, me.ID, "Nightkin Ambusher", "Creature — Mutant Warrior", radNightkinAmbusher, 4, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	radGive(t, g, opp, 1)
	effectstest.BrAttack(t, g, ambusher)

	if effectstest.BrOffers(effectstest.BrOffered(t, g, opp.ID), blocker, ambusher) {
		t.Error("the enumerator offers a block on an Ambusher whose defender has a rad counter")
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlocker(blocker, ambusher), game.BlockReasonCantBeBlocked)
	if refused.Source != ambusher {
		t.Errorf("the refusal names %s, want the Ambusher", refused.Source)
	}

}

// A rad counter on somebody else does not matter, and the defender
// losing theirs before blockers are declared makes the Ambusher blockable.
func TestNightkinAmbusherIsBlockableOnceTheDefenderLosesTheirRadCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	ambusher := effectstest.B12Push(g, me.ID, "Nightkin Ambusher", "Creature — Mutant Warrior", radNightkinAmbusher, 4, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	radGive(t, g, opp, 1)
	radGive(t, g, bystander, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ambusher, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	radGive(t, g, opp, -1)
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if !effectstest.BrOffers(effectstest.BrOffered(t, g, opp.ID), blocker, ambusher) {
		t.Error("the enumerator hides a legal block")
	}
	if err := g.DeclareBlocker(blocker, ambusher); err != nil {
		t.Fatalf("the defender has no rad counter now: %v", err)
	}
}

// Attacking a player without rad counters leaves it blockable even when
// another player has some.
func TestNightkinAmbusherIsBlockableByAnUnirradiatedDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	ambusher := effectstest.B12Push(g, me.ID, "Nightkin Ambusher", "Creature — Mutant Warrior", radNightkinAmbusher, 4, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	radGive(t, g, other, 5)
	effectstest.BrAttack(t, g, ambusher)
	if !effectstest.BrOffers(effectstest.BrOffered(t, g, opp.ID), blocker, ambusher) {
		t.Error("the enumerator hides a legal block")
	}
	if err := g.DeclareBlocker(blocker, ambusher); err != nil {
		t.Fatalf("block: %v", err)
	}
}

// --- Vault 12: The Necropolis -------------------------------------

func radVaultChapterI(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	saga := effectstest.CastCatalogSpell(t, g, "Vault 12: The Necropolis", "Enchantment — Saga", radVault12, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return saga
}

func TestVault12ChaptersGiveCountTokensThenCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	radLandLibraries(g)

	saga := radVaultChapterI(t, g)
	radEach(t, g, 3)

	// A nontoken Zombie, a Mutant and a Human of mine, and an opponent's
	// Zombie, for chapter III to sort.
	zombie := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Walker", TypeLine: "Creature — Zombie", Power: 1, Toughness: 1})
	mutant := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Brute", TypeLine: "Creature — Mutant", Power: 1, Toughness: 1})
	human := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Settler", TypeLine: "Creature — Human", Power: 1, Toughness: 1})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Walker", TypeLine: "Creature — Zombie", Power: 1, Toughness: 1})

	// Chapter II: X is the 12 rad counters among the four players.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	radSettleTurnStart(t, g, me.ID)
	if got := effectstest.LoreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters %d, want 2", got)
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie Mutant"); n != 12 {
		t.Fatalf("%d Zombie Mutant tokens, want 12", n)
	}
	var token game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Zombie Mutant" {
				token = c
				return
			}
		}
	})
	if !token.IsToken() || !token.HasSubtype("Zombie") || !token.HasSubtype("Mutant") || token.Power != 2 || token.Toughness != 2 || !token.HasColor("B") {
		t.Errorf("the token is %q %q %d/%d %v, want a 2/2 black Zombie Mutant token", token.Name, token.TypeLine, token.Power, token.Toughness, token.Colors)
	}

	// Chapter III: two +1/+1 counters on each Zombie or Mutant I
	// control, tokens included, and nothing else; then the Saga goes.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	radSettleTurnStart(t, g, me.ID)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"my Zombie", zombie, 2}, {"my Mutant", mutant, 2}, {"my Human", human, 0}, {"their Zombie", theirs, 0}} {
		if n := effectstest.Pr8Counters(g, tc.id, game.CounterPlusOne); n != tc.want {
			t.Errorf("%s has %d +1/+1 counters, want %d", tc.name, n, tc.want)
		}
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Zombie Mutant" && c.Counters[game.CounterPlusOne] != 2 {
				t.Errorf("a token has %d +1/+1 counters, want 2", c.Counters[game.CounterPlusOne])
			}
		}
	})
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// X counts the rad counters as chapter II resolves: none among players
// is no tokens, and not an error.
func TestVault12ChapterTwoWithNoRadCountersMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	radLandLibraries(g)
	radVaultChapterI(t, g)
	for _, p := range g.Seats {
		radGive(t, g, p, -3)
	}
	radEach(t, g, 0)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Zombie Mutant"); n != 0 {
		t.Errorf("%d tokens, want 0", n)
	}
}

// A restore point taken between the chapters carries the Saga and the
// counters, and chapter II still counts them.
func TestVault12SurvivesARestoreBetweenChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	radLandLibraries(g)
	saga := radVaultChapterI(t, g)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	radEach(t, restored, 3)
	effectstest.AdvanceToPrecombatMainOf(t, restored, seat)
	radSettleTurnStart(t, restored, restored.Seats[seat].ID)
	if got := effectstest.LoreCountersOn(restored, saga); got != 2 {
		t.Fatalf("restored: lore counters %d, want 2", got)
	}
	if n := effectstest.OnBattlefieldNamed(restored, "Zombie Mutant"); n != 12 {
		t.Errorf("restored: %d tokens, want 12", n)
	}
}

// --- Mariposa Military Base ---------------------------------------

func TestMariposaMilitaryBaseEntersTappedForTwoRadCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PlayLandFromHand(t, g, "Mariposa Military Base", radMariposaBase)
	radAnswerMay(t, g, me.ID, true)
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the land should have entered tapped")
	}
	if effectstest.RadOf(me) != 2 {
		t.Errorf("rad %d, want 2", effectstest.RadOf(me))
	}
}

func TestMariposaMilitaryBaseDeclinedEntersUntappedWithNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PlayLandFromHand(t, g, "Mariposa Military Base", radMariposaBase)
	radAnswerMay(t, g, me.ID, false)
	if effectstest.B16Tapped(t, g, land) {
		t.Error("declined, the land should have entered untapped")
	}
	if effectstest.RadOf(me) != 0 {
		t.Errorf("rad %d, want 0", effectstest.RadOf(me))
	}
}

// {5},{T}: Draw a card costs {1} less for each rad counter its
// controller has, down to just the tap.
func TestMariposaMilitaryBaseDrawCostsLessPerRadCounter(t *testing.T) {
	// Six mana float; what is left after the activation is what the
	// ability did not cost.
	for _, tc := range []struct{ rad, spent int }{
		{0, 5}, {1, 4}, {2, 3}, {4, 1}, {5, 0}, {9, 0},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Mariposa Military Base", TypeLine: "Land", OracleID: radMariposaBase,
			Owner: me.ID, Controller: me.ID,
		})
		if tc.rad > 0 {
			radGive(t, g, me, tc.rad)
		}
		effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
		hand := me.Hand.Size()
		effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		var left int
		g.ReadSnapshot(func() { left = me.ManaPool.Total() })
		if left != 6-tc.spent {
			t.Errorf("%d rad counters: %d mana left of 6, want %d (the ability costs %d)", tc.rad, left, 6-tc.spent, tc.spent)
		}
		if got := me.Hand.Size() - hand; got != 1 {
			t.Errorf("%d rad counters: drew %d, want 1", tc.rad, got)
		}
		if !effectstest.B16Tapped(t, g, land) {
			t.Errorf("%d rad counters: the land was not tapped for the cost", tc.rad)
		}
		for _, ev := range g.Events {
			if ev.Kind == game.EventCostWarning {
				t.Errorf("%d rad counters: the cost warned", tc.rad)
			}
		}
	}
}

// --- Nuka-Nuke Launcher -------------------------------------------

// radLauncherAttack equips a Bear of the active player with a
// Nuka-Nuke Launcher and attacks `target` (a player, or a permanent),
// settling the trigger. The libraries are lands, so the counters' own
// mill (CR 728.1) costs nobody anything when a test reaches their turn.
func radLauncherAttack(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	radLandLibraries(g)
	me := g.Seats[g.Turn.ActiveSeat]
	launcher := effectstest.B12Push(g, me.ID, "Nuka-Nuke Launcher", "Artifact — Equipment", radNukaNukeLauncher, 0, 0)
	bear := effectstest.BrCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BrAttach(t, g, launcher, bear)
	effectstest.DeclareAttack(t, g, target, bear)
	effectstest.PassPriorityAroundTable(t, g)
}

// radCastFree casts a spell with no cost and no effect as `p`.
func radCastFree(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	effectstest.CastFromHandForTest(t, g, p, "Test Spell", "Instant", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestNukaNukeLauncherPumpsAndGivesIntimidate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	launcher := effectstest.B12Push(g, me.ID, "Nuka-Nuke Launcher", "Artifact — Equipment", radNukaNukeLauncher, 0, 0)
	bear := effectstest.BrCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BrAttach(t, g, launcher, bear)
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Errorf("power %d, want 5 (+3/+0)", p)
	}
	equipped := effectstest.GaLayered(t, g, bear)
	if !game.HasKeyword(&equipped, "intimidate") {
		t.Error("the equipped creature should have intimidate")
	}
}

// The attack aims the effect at the defending player alone, for the
// rest of their next turn: every spell they cast gives them two rad
// counters, and nobody else's does.
func TestNukaNukeLauncherIrradiatesTheDefendersSpellsUntilTheEndOfTheirNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	radLauncherAttack(t, g, opp.ID)

	// Their spells cast right now, in my turn, count.
	radCastFree(t, g, opp)
	if effectstest.RadOf(opp) != 2 {
		t.Fatalf("rad %d after one spell, want 2", effectstest.RadOf(opp))
	}
	radCastFree(t, g, opp)
	if effectstest.RadOf(opp) != 4 {
		t.Fatalf("rad %d after two spells, want 4", effectstest.RadOf(opp))
	}
	// Mine, and a third player's, do not.
	radCastFree(t, g, me)
	radCastFree(t, g, other)
	if effectstest.RadOf(me) != 0 || effectstest.RadOf(other) != 0 {
		t.Errorf("rad: me %d, other %d, want 0 and 0", effectstest.RadOf(me), effectstest.RadOf(other))
	}

	// It lasts through the defender's own next turn, and ends at that
	// turn's cleanup step.
	for g.Turn.ActiveSeat != 1 {
		effectstest.AdvancePastCleanupForTest(t, g)
	}
	effectstest.AdvanceToMain(t, g)
	radCastFree(t, g, opp)
	if effectstest.RadOf(opp) != 6 {
		t.Errorf("rad %d after a spell in their own turn, want 6", effectstest.RadOf(opp))
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	effectstest.AdvanceToMain(t, g)
	radCastFree(t, g, opp)
	if effectstest.RadOf(opp) != 6 {
		t.Errorf("rad %d after a spell once their turn was over, want 6", effectstest.RadOf(opp))
	}
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("%d delayed triggers outlived the duration", n)
	}
}

// The waiting effect is data: a restore point taken while it is live
// still gives the counters, and still ends.
func TestNukaNukeLauncherSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	radLauncherAttack(t, g, g.Seats[1].ID)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	opp := restored.Seats[1]
	radCastFree(t, restored, opp)
	if effectstest.RadOf(opp) != 2 {
		t.Fatalf("restored: rad %d after a spell, want 2", effectstest.RadOf(opp))
	}
	for restored.Turn.ActiveSeat != 1 {
		effectstest.AdvancePastCleanupForTest(t, restored)
	}
	effectstest.AdvancePastCleanupForTest(t, restored)
	effectstest.AdvanceToMain(t, restored)
	radCastFree(t, restored, opp)
	if effectstest.RadOf(opp) != 2 {
		t.Errorf("restored: rad %d after their turn ended, want it to stay 2", effectstest.RadOf(opp))
	}
}

// "Defending player" is the player behind the attack (CR 508.5): the
// controller of the planeswalker the equipped creature attacks.
func TestNukaNukeLauncherAimsAtThePlaneswalkersController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, third := g.Seats[1], g.Seats[2]
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 4}, Owner: third.ID, Controller: third.ID,
	})
	radLauncherAttack(t, g, walker)
	radCastFree(t, g, opp)
	radCastFree(t, g, third)
	if effectstest.RadOf(opp) != 0 || effectstest.RadOf(third) != 2 {
		t.Errorf("rad: other opponent %d (want 0), the planeswalker's controller %d (want 2)", effectstest.RadOf(opp), effectstest.RadOf(third))
	}
}
