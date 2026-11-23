package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b06_cards_test.go — Innistrad: Midnight Hunt, S82 batch 6 (#5497):
// the Vampires that ask whether an opponent lost life this turn.

const (
	midB06BatWhispererOracle = "f20f6da9-d678-486b-99f9-7121669fef81"
	midB06OutlawOracle       = "f202712d-7cd1-47aa-9664-526a884a1be2"
	midB06PitFighterOracle   = "46d6bc30-9a21-4476-947b-87951cd9c882"
	midB06AmbusherOracle     = "897f666e-0dda-4577-a3d7-792e3ac2f460"
	midB06DreadknightOracle  = "b5e42be3-6bfc-4905-9237-cb8d53e2349e"
	midB06SocialiteOracle    = "91a45e96-2a93-4c48-b472-11904d2b856d"
	midB06ForagersOracle     = "e95f22dd-b89a-4836-86ad-5a9a1627af38"
	midB06PerforatorOracle   = "6e991808-fea4-46a8-9e87-7b5917f33ce8"
	midB06StingerOracle      = "5bd66394-6e1e-4b13-8824-e83768fb533e"
	midB06RushOracle         = "721aeedc-1de5-40d0-a04c-a2ebc056d06a"
	midB06SpecialistOracle   = "ab431f93-73bb-49d9-9321-a0712f096ed7"
	midB06CullingOracle      = "0b7fabe8-bc44-47b1-93bb-b895a1bd0aa9"
)

func midB06OpponentLosesLife(t *testing.T, g *game.Game, p uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
}

func midB06CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	return effectstest.BattlefieldNamed(g, controller, name)
}

func TestMidB06EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		midB06BatWhispererOracle: "Bat Whisperer",
		midB06OutlawOracle:       "Arrogant Outlaw",
		midB06PitFighterOracle:   "Falkenrath Pit Fighter",
		midB06AmbusherOracle:     "Voldaren Ambusher",
		midB06DreadknightOracle:  "Mounted Dreadknight",
		midB06SocialiteOracle:    "Vampire Socialite",
		midB06ForagersOracle:     "Famished Foragers",
		midB06PerforatorOracle:   "Falkenrath Perforator",
		midB06StingerOracle:      "Voldaren Stinger",
		midB06RushOracle:         "Neonate's Rush",
		midB06SpecialistOracle:   "Slaughter Specialist",
		midB06CullingOracle:      "Bloodline Culling",
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

func TestMidB06BatWhispererMakesABatOnlyAfterAnOpponentLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		if lost {
			midB06OpponentLosesLife(t, g, opp.ID)
		}
		effectstest.CastHoldCreature(t, g, "Bat Whisperer", "Creature — Vampire", midB06BatWhispererOracle, 4, 2)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if lost {
			want = 1
		}
		if got := midB06CountNamed(g, me.ID, "Bat"); got != want {
			t.Errorf("lost=%v: %d Bats, want %d", lost, got, want)
		}
	}
}

func TestMidB06BatWhispererIgnoresMyOwnLifeLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	midB06OpponentLosesLife(t, g, me.ID)
	effectstest.CastHoldCreature(t, g, "Bat Whisperer", "Creature — Vampire", midB06BatWhispererOracle, 4, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := midB06CountNamed(g, me.ID, "Bat"); got != 0 {
		t.Errorf("my own life loss made %d Bats", got)
	}
}

func TestMidB06ArrogantOutlawDrainsOnlyAfterAnOpponentLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, other := effectstest.BlbB1Seats(g)
		if lost {
			midB06OpponentLosesLife(t, g, opp.ID)
		}
		myLife, oppLife, otherLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, other.ID)
		effectstest.CastHoldCreature(t, g, "Arrogant Outlaw", "Creature — Vampire Noble", midB06OutlawOracle, 3, 2)
		effectstest.PassPriorityAroundTable(t, g)
		wantMe, wantOpp, wantOther := myLife, oppLife, otherLife
		if lost {
			wantMe, wantOpp, wantOther = myLife+2, oppLife-2, otherLife-2
		}
		if effectstest.LifeOf(g, me.ID) != wantMe || effectstest.LifeOf(g, opp.ID) != wantOpp || effectstest.LifeOf(g, other.ID) != wantOther {
			t.Errorf("lost=%v: life %d/%d/%d, want %d/%d/%d", lost,
				effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, other.ID), wantMe, wantOpp, wantOther)
		}
	}
}

func TestMidB06FamishedForagersAddsRedOnlyAfterAnOpponentLostLifeAndRummages(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		if lost {
			midB06OpponentLosesLife(t, g, opp.ID)
		}
		effectstest.CastHoldCreature(t, g, "Famished Foragers", "Creature — Vampire", midB06ForagersOracle, 4, 3)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if lost {
			want = 3
		}
		if got := effectstest.McPoolCount(me, "R"); got != want {
			t.Errorf("lost=%v: pool has %d {R}, want %d", lost, got, want)
		}
	}
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	f := effectstest.BlbB1Push(g, me.ID, "Famished Foragers", "Creature — Vampire", midB06ForagersOracle, 4, 3)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.SocB21SetHand(me, 2)
	pitched := me.Hand.Cards[0].InstanceID
	lib := me.Library.Size()
	effectstest.FloatForTest(g, me, "RRR")
	if err := g.ActivateCatalogAbility(me.ID, f, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitched}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) || me.Hand.Size() != 2 || me.Library.Size() != lib-1 {
		t.Errorf("rummage: hand %d, library %d -> %d", me.Hand.Size(), lib, me.Library.Size())
	}
}

func TestMidB06MountedDreadknightEntersWithACounterOnlyAfterAnOpponentLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		_, opp, _ := effectstest.BlbB1Seats(g)
		if lost {
			midB06OpponentLosesLife(t, g, opp.ID)
		}
		k := effectstest.CastHoldCreature(t, g, "Mounted Dreadknight", "Creature — Vampire Knight", midB06DreadknightOracle, 5, 4)
		want := 5
		if lost {
			want = 6
		}
		if p, _ := effectstest.PtOf(t, g, k); p != want {
			t.Errorf("lost=%v: power %d, want %d", lost, p, want)
		}
		effectstest.AssertKeywords(t, g, k, "trample")
	}
}

func TestMidB06VampireSocialiteCountersOtherVampiresOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	v := effectstest.BlbB1Push(g, me.ID, "Other Vampire", "Creature — Vampire", "", 2, 2)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Vampire", "Creature — Vampire", "", 2, 2)
	midB06OpponentLosesLife(t, g, opp.ID)
	soc := effectstest.CastHoldCreature(t, g, "Vampire Socialite", "Creature — Vampire Noble", midB06SocialiteOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, v, game.CounterPlusOne); got != 1 {
		t.Errorf("my other Vampire has %d counters, want 1", got)
	}
	if effectstest.CounterOn(g, bear, game.CounterPlusOne) != 0 || effectstest.CounterOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("a non-Vampire or an opponent's Vampire got a counter")
	}
	if got := effectstest.CounterOn(g, soc, game.CounterPlusOne); got != 0 {
		t.Errorf("the Socialite put %d counters on itself", got)
	}
}

func TestMidB06VampireSocialiteDoesNothingWithoutLifeLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	v := effectstest.BlbB1Push(g, me.ID, "Other Vampire", "Creature — Vampire", "", 2, 2)
	effectstest.CastHoldCreature(t, g, "Vampire Socialite", "Creature — Vampire Noble", midB06SocialiteOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, v, game.CounterPlusOne); got != 0 {
		t.Errorf("the Vampire got %d counters with no life lost", got)
	}
	late := effectstest.CastHoldCreature(t, g, "Late Vampire", "Creature — Vampire", "", 2, 2)
	if got := effectstest.CounterOn(g, late, game.CounterPlusOne); got != 0 {
		t.Errorf("a Vampire entered with %d counters with no life lost", got)
	}
}

func TestMidB06VampireSocialiteGivesEnteringVampiresACounterWhileAnOpponentLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Vampire Socialite", "Creature — Vampire Noble", midB06SocialiteOracle, 2, 2)
	midB06OpponentLosesLife(t, g, opp.ID)
	vamp := effectstest.CastHoldCreature(t, g, "Late Vampire", "Creature — Vampire", "", 2, 2)
	if got := effectstest.CounterOn(g, vamp, game.CounterPlusOne); got != 1 {
		t.Errorf("an entering Vampire has %d counters, want 1", got)
	}
	bear := effectstest.CastHoldCreature(t, g, "Late Bear", "Creature — Bear", "", 2, 2)
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("an entering non-Vampire has %d counters, want 0", got)
	}
}

func TestMidB06VampireSocialiteDoesNotCounterItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	midB06OpponentLosesLife(t, g, opp.ID)
	soc := effectstest.CastHoldCreature(t, g, "Vampire Socialite", "Creature — Vampire Noble", midB06SocialiteOracle, 2, 2)
	if got := effectstest.CounterOn(g, soc, game.CounterPlusOne); got != 0 {
		t.Errorf("the Socialite entered with %d counters, want 0 (\"other\")", got)
	}
}

func TestMidB06FalkenrathPerforatorPingsTheDefendingPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	p := effectstest.BlbB1Push(g, me.ID, "Falkenrath Perforator", "Creature — Vampire", midB06PerforatorOracle, 2, 1)
	oppLife, otherLife := effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, other.ID)
	effectstest.DeclareAttack(t, g, opp.ID, p)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("defending player at %d, want %d", got, oppLife-1)
	}
	if got := effectstest.LifeOf(g, other.ID); got != otherLife {
		t.Errorf("a bystander lost life: %d", got)
	}
}

func TestMidB06FalkenrathPerforatorIgnoresOtherAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Falkenrath Perforator", "Creature — Vampire", midB06PerforatorOracle, 2, 1)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	oppLife := effectstest.LifeOf(g, opp.ID)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife {
		t.Errorf("another creature's attack pinged: %d, want %d", got, oppLife)
	}
}

func TestMidB06VoldarenStingerHasFirstStrikeOnlyWhileAttackingAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	s := effectstest.BlbB1Push(g, me.ID, "Voldaren Stinger", "Creature — Vampire Warrior", midB06StingerOracle, 1, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, s, "first strike") {
		t.Error("the Stinger has first strike while not attacking")
	}
	effectstest.DeclareAttack(t, g, opp.ID, s)
	if !effectstest.HasKeywordOnBattlefield(t, g, s, "first strike") {
		t.Error("the attacking Stinger lacks first strike")
	}
	effectstest.FloatForTest(g, me, "RRR")
	effectstest.B16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, s); p != 3 || tough != 1 {
		t.Errorf("the pumped Stinger is %d/%d, want 3/1", p, tough)
	}
}

func TestMidB06VoldarenAmbusherDamagesByVampireCountAfterLifeLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Other Vampire", "Creature — Vampire", "", 2, 2)
	small := effectstest.BlbB1Push(g, opp.ID, "Small", "Creature — Bear", "", 2, 2)
	big := effectstest.BlbB1Push(g, opp.ID, "Big", "Creature — Bear", "", 3, 3)
	midB06OpponentLosesLife(t, g, opp.ID)
	effectstest.CastHoldCreature(t, g, "Voldaren Ambusher", "Creature — Vampire Archer", midB06AmbusherOracle, 2, 2)
	effectstest.AnswerPickTarget(t, g, small)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("X=2 did not kill the 2/2")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the untargeted creature died")
	}
}

func TestMidB06VoldarenAmbusherDoesNothingWithoutLifeLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	small := effectstest.BlbB1Push(g, opp.ID, "Small", "Creature — Bear", "", 1, 1)
	effectstest.CastHoldCreature(t, g, "Voldaren Ambusher", "Creature — Vampire Archer", midB06AmbusherOracle, 2, 2)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("the Ambusher asked for a target with no life lost")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the Ambusher dealt damage with no life lost")
	}
}

func TestMidB06FalkenrathPitFighterDrawsTwoOnlyAfterLifeLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	f := effectstest.BlbB1Push(g, me.ID, "Falkenrath Pit Fighter", "Creature — Vampire Warrior", midB06PitFighterOracle, 2, 1)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.SocB21SetHand(me, 2)
	pitched := me.Hand.Cards[0].InstanceID
	effectstest.FloatForTest(g, me, "RR")
	params := game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitched}, SacrificeIDs: []uuid.UUID{f}}
	if err := g.ActivateCatalogAbility(me.ID, f, 0, params); err == nil {
		t.Fatal("the Pit Fighter activated with no opponent having lost life")
	}
	midB06OpponentLosesLife(t, g, opp.ID)
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, f, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(f) {
		t.Error("the Vampire was not sacrificed")
	}
	if me.Library.Size() != lib-2 || !me.Graveyard.Contains(pitched) {
		t.Errorf("library %d -> %d, want two drawn and the card discarded", lib, me.Library.Size())
	}
}

func TestMidB06FalkenrathPitFighterCanSacrificeAnotherVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	f := effectstest.BlbB1Push(g, me.ID, "Falkenrath Pit Fighter", "Creature — Vampire Warrior", midB06PitFighterOracle, 2, 1)
	other := effectstest.BlbB1Push(g, me.ID, "Other Vampire", "Creature — Vampire", "", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.SocB21SetHand(me, 2)
	midB06OpponentLosesLife(t, g, opp.ID)
	effectstest.FloatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, f, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{me.Hand.Cards[0].InstanceID}, SacrificeIDs: []uuid.UUID{other},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) || !g.Battlefield.Contains(f) {
		t.Error("the other Vampire should be sacrificed and the Pit Fighter stay")
	}
}

func TestMidB06NeonatesRushHitsTheCreatureAndItsControllerAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	victim := effectstest.BlbB1Push(g, opp.ID, "Victim", "Creature — Bear", "", 1, 1)
	oppLife := effectstest.LifeOf(g, opp.ID)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Neonate's Rush", "Instant", midB06RushOracle, effectstest.B16TargetCard(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 1/1 survived 1 damage")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("the creature's controller is at %d, want %d", got, oppLife-1)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d: a card should have been drawn", hand, got)
	}
}

func TestMidB06NeonatesRushCostsOneLessWithAVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	if got := effectstest.SelfPricedMV(t, g, me, midB06RushOracle, "Instant", "{2}{R}"); got != 3 {
		t.Errorf("no Vampire: %d, want 3", got)
	}
	effectstest.BlbB1Push(g, opp.ID, "Their Vampire", "Creature — Vampire", "", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, midB06RushOracle, "Instant", "{2}{R}"); got != 3 {
		t.Errorf("an opponent's Vampire: %d, want 3", got)
	}
	effectstest.BlbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, midB06RushOracle, "Instant", "{2}{R}"); got != 3 {
		t.Errorf("a non-Vampire: %d, want 3", got)
	}
	effectstest.BlbB1Push(g, me.ID, "My Vampire", "Creature — Vampire", "", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, midB06RushOracle, "Instant", "{2}{R}"); got != 2 {
		t.Errorf("my Vampire: %d, want 2", got)
	}
}

func TestMidB06SlaughterSpecialistGivesEachOpponentAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	effectstest.CastHoldCreature(t, g, "Slaughter Specialist", "Creature — Vampire Warrior", midB06SpecialistOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range []*game.Player{opp, other} {
		if got := midB06CountNamed(g, p.ID, "Human"); got != 1 {
			t.Errorf("opponent %s has %d Humans, want 1", p.ID, got)
		}
	}
	if got := midB06CountNamed(g, me.ID, "Human"); got != 0 {
		t.Errorf("I got %d Humans", got)
	}
}

func TestMidB06SlaughterSpecialistGrowsWhenAnOpponentsCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	s := effectstest.BlbB1Push(g, me.ID, "Slaughter Specialist", "Creature — Vampire Warrior", midB06SpecialistOracle, 3, 3)
	mine := effectstest.BlbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 1, 1)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 1, 1)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, theirs); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, s, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after an opponent's creature died, want 1", got)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, mine); err != nil {
		t.Fatalf("destroy mine: %v", err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, s, game.CounterPlusOne); got != 1 {
		t.Errorf("my own creature dying grew it: %d counters", got)
	}
}

func midB06CastCulling(t *testing.T, g *game.Game, mode int, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Bloodline Culling", TypeLine: "Instant", OracleID: midB06CullingOracle,
		Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets, Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestMidB06BloodlineCullingShrinksATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	big := effectstest.BlbB1Push(g, opp.ID, "Big", "Creature — Beast", "", 6, 6)
	midB06CastCulling(t, g, 0, effectstest.B16TargetCard(big))
	if p, tough := effectstest.PtOf(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
	small := effectstest.BlbB1Push(g, opp.ID, "Small", "Creature — Beast", "", 2, 2)
	midB06CastCulling(t, g, 0, effectstest.B16TargetCard(small))
	if g.Battlefield.Contains(small) {
		t.Error("a 2/2 survived -5/-5")
	}
}

func TestMidB06BloodlineCullingTokenModeHitsOnlyTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	tok := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 2, Toughness: 2})
	mine := effectstest.PushPermanent(g, me.ID, game.Card{Name: "My Token", TypeLine: "Token Creature — Goblin", Power: 3, Toughness: 3})
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	midB06CastCulling(t, g, 1, nil)
	if g.Battlefield.Contains(tok) {
		t.Error("a 2/2 token survived -2/-2")
	}
	if p, tough := effectstest.PtOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("my 3/3 token is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a nontoken got smaller: %d/%d", p, tough)
	}
}
