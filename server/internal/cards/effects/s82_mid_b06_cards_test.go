package effects

import (
	"testing"

	"github.com/google/uuid"

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
	return battlefieldNamed(g, controller, name)
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

func TestMidB06BatWhispererMakesABatOnlyAfterAnOpponentLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		if lost {
			midB06OpponentLosesLife(t, g, opp.ID)
		}
		castHoldCreature(t, g, "Bat Whisperer", "Creature — Vampire", midB06BatWhispererOracle, 4, 2)
		passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	midB06OpponentLosesLife(t, g, me.ID)
	castHoldCreature(t, g, "Bat Whisperer", "Creature — Vampire", midB06BatWhispererOracle, 4, 2)
	passPriorityAroundTable(t, g)
	if got := midB06CountNamed(g, me.ID, "Bat"); got != 0 {
		t.Errorf("my own life loss made %d Bats", got)
	}
}

func TestMidB06ArrogantOutlawDrainsOnlyAfterAnOpponentLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, other := blbB1Seats(g)
		if lost {
			midB06OpponentLosesLife(t, g, opp.ID)
		}
		myLife, oppLife, otherLife := lifeOf(g, me.ID), lifeOf(g, opp.ID), lifeOf(g, other.ID)
		castHoldCreature(t, g, "Arrogant Outlaw", "Creature — Vampire Noble", midB06OutlawOracle, 3, 2)
		passPriorityAroundTable(t, g)
		wantMe, wantOpp, wantOther := myLife, oppLife, otherLife
		if lost {
			wantMe, wantOpp, wantOther = myLife+2, oppLife-2, otherLife-2
		}
		if lifeOf(g, me.ID) != wantMe || lifeOf(g, opp.ID) != wantOpp || lifeOf(g, other.ID) != wantOther {
			t.Errorf("lost=%v: life %d/%d/%d, want %d/%d/%d", lost,
				lifeOf(g, me.ID), lifeOf(g, opp.ID), lifeOf(g, other.ID), wantMe, wantOpp, wantOther)
		}
	}
}

func TestMidB06FamishedForagersAddsRedOnlyAfterAnOpponentLostLifeAndRummages(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		if lost {
			midB06OpponentLosesLife(t, g, opp.ID)
		}
		castHoldCreature(t, g, "Famished Foragers", "Creature — Vampire", midB06ForagersOracle, 4, 3)
		passPriorityAroundTable(t, g)
		want := 0
		if lost {
			want = 3
		}
		if got := mcPoolCount(me, "R"); got != want {
			t.Errorf("lost=%v: pool has %d {R}, want %d", lost, got, want)
		}
	}
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	f := blbB1Push(g, me.ID, "Famished Foragers", "Creature — Vampire", midB06ForagersOracle, 4, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	socB21SetHand(me, 2)
	pitched := me.Hand.Cards[0].InstanceID
	lib := me.Library.Size()
	floatForTest(g, me, "RRR")
	if err := g.ActivateCatalogAbility(me.ID, f, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitched}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) || me.Hand.Size() != 2 || me.Library.Size() != lib-1 {
		t.Errorf("rummage: hand %d, library %d -> %d", me.Hand.Size(), lib, me.Library.Size())
	}
}

func TestMidB06MountedDreadknightEntersWithACounterOnlyAfterAnOpponentLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := newCatalogGame(t)
		_, opp, _ := blbB1Seats(g)
		if lost {
			midB06OpponentLosesLife(t, g, opp.ID)
		}
		k := castHoldCreature(t, g, "Mounted Dreadknight", "Creature — Vampire Knight", midB06DreadknightOracle, 5, 4)
		want := 5
		if lost {
			want = 6
		}
		if p, _ := ptOf(t, g, k); p != want {
			t.Errorf("lost=%v: power %d, want %d", lost, p, want)
		}
		assertKeywords(t, g, k, "trample")
	}
}

func TestMidB06VampireSocialiteCountersOtherVampiresOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	v := blbB1Push(g, me.ID, "Other Vampire", "Creature — Vampire", "", 2, 2)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Vampire", "Creature — Vampire", "", 2, 2)
	midB06OpponentLosesLife(t, g, opp.ID)
	soc := castHoldCreature(t, g, "Vampire Socialite", "Creature — Vampire Noble", midB06SocialiteOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, v, game.CounterPlusOne); got != 1 {
		t.Errorf("my other Vampire has %d counters, want 1", got)
	}
	if counterOn(g, bear, game.CounterPlusOne) != 0 || counterOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("a non-Vampire or an opponent's Vampire got a counter")
	}
	if got := counterOn(g, soc, game.CounterPlusOne); got != 0 {
		t.Errorf("the Socialite put %d counters on itself", got)
	}
}

func TestMidB06VampireSocialiteDoesNothingWithoutLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	v := blbB1Push(g, me.ID, "Other Vampire", "Creature — Vampire", "", 2, 2)
	castHoldCreature(t, g, "Vampire Socialite", "Creature — Vampire Noble", midB06SocialiteOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, v, game.CounterPlusOne); got != 0 {
		t.Errorf("the Vampire got %d counters with no life lost", got)
	}
	late := castHoldCreature(t, g, "Late Vampire", "Creature — Vampire", "", 2, 2)
	if got := counterOn(g, late, game.CounterPlusOne); got != 0 {
		t.Errorf("a Vampire entered with %d counters with no life lost", got)
	}
}

func TestMidB06VampireSocialiteGivesEnteringVampiresACounterWhileAnOpponentLostLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Vampire Socialite", "Creature — Vampire Noble", midB06SocialiteOracle, 2, 2)
	midB06OpponentLosesLife(t, g, opp.ID)
	vamp := castHoldCreature(t, g, "Late Vampire", "Creature — Vampire", "", 2, 2)
	if got := counterOn(g, vamp, game.CounterPlusOne); got != 1 {
		t.Errorf("an entering Vampire has %d counters, want 1", got)
	}
	bear := castHoldCreature(t, g, "Late Bear", "Creature — Bear", "", 2, 2)
	if got := counterOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("an entering non-Vampire has %d counters, want 0", got)
	}
}

func TestMidB06VampireSocialiteDoesNotCounterItself(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	midB06OpponentLosesLife(t, g, opp.ID)
	soc := castHoldCreature(t, g, "Vampire Socialite", "Creature — Vampire Noble", midB06SocialiteOracle, 2, 2)
	if got := counterOn(g, soc, game.CounterPlusOne); got != 0 {
		t.Errorf("the Socialite entered with %d counters, want 0 (\"other\")", got)
	}
}

func TestMidB06FalkenrathPerforatorPingsTheDefendingPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	p := blbB1Push(g, me.ID, "Falkenrath Perforator", "Creature — Vampire", midB06PerforatorOracle, 2, 1)
	oppLife, otherLife := lifeOf(g, opp.ID), lifeOf(g, other.ID)
	declareAttack(t, g, opp.ID, p)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("defending player at %d, want %d", got, oppLife-1)
	}
	if got := lifeOf(g, other.ID); got != otherLife {
		t.Errorf("a bystander lost life: %d", got)
	}
}

func TestMidB06FalkenrathPerforatorIgnoresOtherAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Falkenrath Perforator", "Creature — Vampire", midB06PerforatorOracle, 2, 1)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	oppLife := lifeOf(g, opp.ID)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != oppLife {
		t.Errorf("another creature's attack pinged: %d, want %d", got, oppLife)
	}
}

func TestMidB06VoldarenStingerHasFirstStrikeOnlyWhileAttackingAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	s := blbB1Push(g, me.ID, "Voldaren Stinger", "Creature — Vampire Warrior", midB06StingerOracle, 1, 1)
	if hasKeywordOnBattlefield(t, g, s, "first strike") {
		t.Error("the Stinger has first strike while not attacking")
	}
	declareAttack(t, g, opp.ID, s)
	if !hasKeywordOnBattlefield(t, g, s, "first strike") {
		t.Error("the attacking Stinger lacks first strike")
	}
	floatForTest(g, me, "RRR")
	b16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, s); p != 3 || tough != 1 {
		t.Errorf("the pumped Stinger is %d/%d, want 3/1", p, tough)
	}
}

func TestMidB06VoldarenAmbusherDamagesByVampireCountAfterLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Other Vampire", "Creature — Vampire", "", 2, 2)
	small := blbB1Push(g, opp.ID, "Small", "Creature — Bear", "", 2, 2)
	big := blbB1Push(g, opp.ID, "Big", "Creature — Bear", "", 3, 3)
	midB06OpponentLosesLife(t, g, opp.ID)
	castHoldCreature(t, g, "Voldaren Ambusher", "Creature — Vampire Archer", midB06AmbusherOracle, 2, 2)
	answerPickTarget(t, g, small)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("X=2 did not kill the 2/2")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the untargeted creature died")
	}
}

func TestMidB06VoldarenAmbusherDoesNothingWithoutLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	small := blbB1Push(g, opp.ID, "Small", "Creature — Bear", "", 1, 1)
	castHoldCreature(t, g, "Voldaren Ambusher", "Creature — Vampire Archer", midB06AmbusherOracle, 2, 2)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("the Ambusher asked for a target with no life lost")
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the Ambusher dealt damage with no life lost")
	}
}

func TestMidB06FalkenrathPitFighterDrawsTwoOnlyAfterLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	f := blbB1Push(g, me.ID, "Falkenrath Pit Fighter", "Creature — Vampire Warrior", midB06PitFighterOracle, 2, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	socB21SetHand(me, 2)
	pitched := me.Hand.Cards[0].InstanceID
	floatForTest(g, me, "RR")
	params := game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitched}, SacrificeIDs: []uuid.UUID{f}}
	if err := g.ActivateCatalogAbility(me.ID, f, 0, params); err == nil {
		t.Fatal("the Pit Fighter activated with no opponent having lost life")
	}
	midB06OpponentLosesLife(t, g, opp.ID)
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, f, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(f) {
		t.Error("the Vampire was not sacrificed")
	}
	if me.Library.Size() != lib-2 || !me.Graveyard.Contains(pitched) {
		t.Errorf("library %d -> %d, want two drawn and the card discarded", lib, me.Library.Size())
	}
}

func TestMidB06FalkenrathPitFighterCanSacrificeAnotherVampire(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	f := blbB1Push(g, me.ID, "Falkenrath Pit Fighter", "Creature — Vampire Warrior", midB06PitFighterOracle, 2, 1)
	other := blbB1Push(g, me.ID, "Other Vampire", "Creature — Vampire", "", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	socB21SetHand(me, 2)
	midB06OpponentLosesLife(t, g, opp.ID)
	floatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, f, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{me.Hand.Cards[0].InstanceID}, SacrificeIDs: []uuid.UUID{other},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) || !g.Battlefield.Contains(f) {
		t.Error("the other Vampire should be sacrificed and the Pit Fighter stay")
	}
}

func TestMidB06NeonatesRushHitsTheCreatureAndItsControllerAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	victim := blbB1Push(g, opp.ID, "Victim", "Creature — Bear", "", 1, 1)
	oppLife := lifeOf(g, opp.ID)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Neonate's Rush", "Instant", midB06RushOracle, b16TargetCard(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 1/1 survived 1 damage")
	}
	if got := lifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("the creature's controller is at %d, want %d", got, oppLife-1)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d: a card should have been drawn", hand, got)
	}
}

func TestMidB06NeonatesRushCostsOneLessWithAVampire(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	if got := selfPricedMV(t, g, me, midB06RushOracle, "Instant", "{2}{R}"); got != 3 {
		t.Errorf("no Vampire: %d, want 3", got)
	}
	blbB1Push(g, opp.ID, "Their Vampire", "Creature — Vampire", "", 2, 2)
	if got := selfPricedMV(t, g, me, midB06RushOracle, "Instant", "{2}{R}"); got != 3 {
		t.Errorf("an opponent's Vampire: %d, want 3", got)
	}
	blbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	if got := selfPricedMV(t, g, me, midB06RushOracle, "Instant", "{2}{R}"); got != 3 {
		t.Errorf("a non-Vampire: %d, want 3", got)
	}
	blbB1Push(g, me.ID, "My Vampire", "Creature — Vampire", "", 2, 2)
	if got := selfPricedMV(t, g, me, midB06RushOracle, "Instant", "{2}{R}"); got != 2 {
		t.Errorf("my Vampire: %d, want 2", got)
	}
}

func TestMidB06SlaughterSpecialistGivesEachOpponentAHuman(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	castHoldCreature(t, g, "Slaughter Specialist", "Creature — Vampire Warrior", midB06SpecialistOracle, 3, 3)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	s := blbB1Push(g, me.ID, "Slaughter Specialist", "Creature — Vampire Warrior", midB06SpecialistOracle, 3, 3)
	mine := blbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 1, 1)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 1, 1)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, theirs); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := counterOn(g, s, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after an opponent's creature died, want 1", got)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, mine); err != nil {
		t.Fatalf("destroy mine: %v", err)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := counterOn(g, s, game.CounterPlusOne); got != 1 {
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
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets, Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestMidB06BloodlineCullingShrinksATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	big := blbB1Push(g, opp.ID, "Big", "Creature — Beast", "", 6, 6)
	midB06CastCulling(t, g, 0, b16TargetCard(big))
	if p, tough := ptOf(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
	small := blbB1Push(g, opp.ID, "Small", "Creature — Beast", "", 2, 2)
	midB06CastCulling(t, g, 0, b16TargetCard(small))
	if g.Battlefield.Contains(small) {
		t.Error("a 2/2 survived -5/-5")
	}
}

func TestMidB06BloodlineCullingTokenModeHitsOnlyTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	tok := pushPermanent(g, opp.ID, game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 2, Toughness: 2})
	mine := pushPermanent(g, me.ID, game.Card{Name: "My Token", TypeLine: "Token Creature — Goblin", Power: 3, Toughness: 3})
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	midB06CastCulling(t, g, 1, nil)
	if g.Battlefield.Contains(tok) {
		t.Error("a 2/2 token survived -2/-2")
	}
	if p, tough := ptOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("my 3/3 token is %d/%d, want 1/1", p, tough)
	}
	if p, tough := ptOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a nontoken got smaller: %d/%d", p, tough)
	}
}
