package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foundations Jumpstart bonus-sheet cards, batch 34 (#5929). Masked
// Blackguard is Ravnica: Clue Edition batch 13's.

const (
	bnsB34Leonin     = "3c692dd9-3986-46de-ac10-cc917b7c41b7"
	bnsB34Light      = "79fc0d30-9b9b-400c-a614-c3a1f90a5380"
	bnsB34Lightwalk  = "6f4848d4-a5dc-4c71-955c-ad2b1cbb7ede"
	bnsB34Lord       = "cc60181a-5eeb-4d66-9876-3194bac1e7a7"
	bnsB34Rider      = "66412a7a-4ec0-41f0-80ad-eb483213ff81"
	bnsB34Blademast  = "774e474a-f2a1-4b62-8cfe-40ad509f9062"
	bnsB34Warlord    = "2279005d-5434-49ce-b16a-212759c5775a"
	bnsB34Branchwalk = "24b5dbe3-578a-4897-8215-e723f00a8f1a"
	bnsB34Pupil      = "bb244878-6440-4dfe-b96b-c9e18d0abe45"
	bnsB34Unicorn    = "5101ff9e-4680-4d38-80a5-b4c21bc67356"
)

func TestBnsB34CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB34Leonin: "Leonin of the Lost Pride", bnsB34Light: "Light of the Legion", bnsB34Lightwalk: "Lightwalker",
		bnsB34Lord: "Lord of the Unreal", bnsB34Rider: "Maraleaf Rider", bnsB34Blademast: "Markov Blademaster",
		bnsB34Warlord: "Markov Warlord", bnsB34Branchwalk: "Merfolk Branchwalker", bnsB34Pupil: "Merfolk Pupil",
		bnsB34Unicorn: "Mesa Unicorn",
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

func bnsB34Counters(g *game.Game, id uuid.UUID) int {
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

func TestBnsB34MesaUnicornHasLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Mesa Unicorn", "Creature — Unicorn", bnsB34Unicorn, 2, 2)
	if !hasKeywordOnBattlefield(t, g, id, "lifelink") {
		t.Error("no lifelink")
	}
}

func TestBnsB34LightwalkerFliesOnlyWithAPlusOneCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Lightwalker", "Creature — Human Warrior", bnsB34Lightwalk, 2, 1)
	if hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Fatal("flies without a counter")
	}
	g.WithWriteLock(func() {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(&Context{Game: g, Item: &game.StackItem{Controller: me.ID, SourceCardID: id}}); err != nil {
			t.Fatal(err)
		}
	})
	if !hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("does not fly with a +1/+1 counter")
	}
}

func TestBnsB34LordPumpsAndShieldsOnlyYourIllusions(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := b12Push(g, me.ID, "Lord of the Unreal", "Creature — Human Wizard", bnsB34Lord, 2, 2)
	mine := b12Creature(g, me.ID, "My Illusion", "Creature — Illusion", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Illusion", "Creature — Illusion", 1, 1)
	for id, want := range map[uuid.UUID]int{lord: 2, mine: 2, bear: 2, theirs: 1} {
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	for id, want := range map[uuid.UUID]bool{lord: false, mine: true, bear: false, theirs: false} {
		if got := hasKeywordOnBattlefield(t, g, id, "hexproof"); got != want {
			t.Errorf("hexproof %v, want %v", got, want)
		}
	}
}

func TestBnsB34BlademasterGrowsOnCombatDamageToAPlayerOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b12Push(g, me.ID, "Markov Blademaster", "Creature — Vampire Warrior", bnsB34Blademast, 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: id, Target: opp.ID, Amount: 1})
	})
	passPriorityAroundTable(t, g)
	if got := bnsB34Counters(g, id); got != 0 {
		t.Fatalf("noncombat damage left %d counters", got)
	}
	dealCombatDamageToPlayer(g, id, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if got := bnsB34Counters(g, id); got != 1 {
		t.Fatalf("%d counters after combat damage, want 1", got)
	}
	dealCombatDamageToPlayer(g, id, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if got := bnsB34Counters(g, id); got != 2 {
		t.Errorf("%d counters after the second hit, want 2", got)
	}
}

func TestBnsB34WarlordStopsUpToTwoCreaturesBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 4)
	b := b12Creature(g, me.ID, "Wall B", "Creature — Wall", 0, 4)
	free := b12Creature(g, opp.ID, "Wall C", "Creature — Wall", 0, 4)
	castHoldCreature(t, g, "Markov Warlord", "Creature — Vampire Warrior", bnsB34Warlord, 4, 4)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{eoeCDCard(a), eoeCDCard(b)}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"first": a, "second": b} {
		if auraRestrictions(t, g, id)&game.CantBlock == 0 {
			t.Errorf("the %s target can still block", name)
		}
	}
	if auraRestrictions(t, g, free)&game.CantBlock != 0 {
		t.Error("an untargeted creature can't block")
	}
}

func TestBnsB34BranchwalkerExploresIntoAHandOrACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := putOnTopForTest(me, "Needle", "Land")
	id := castHoldCreature(t, g, "Merfolk Branchwalker", "Creature — Merfolk Scout", bnsB34Branchwalk, 2, 1)
	if !me.Hand.Contains(land) {
		t.Error("the revealed land did not go to hand")
	}
	if got := bnsB34Counters(g, id); got != 0 {
		t.Errorf("%d counters after a land, want 0", got)
	}

	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	spell := putOnTopForTest(me, "Needle", "Sorcery")
	id = castHoldCreature(t, g, "Merfolk Branchwalker", "Creature — Merfolk Scout", bnsB34Branchwalk, 2, 1)
	answerExploreQuestion(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := bnsB34Counters(g, id); got != 1 {
		t.Errorf("%d counters after a nonland, want 1", got)
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the nonland card was not put into the graveyard")
	}
}

func TestBnsB34PupilLootsOnEntryAndFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	putOnTopForTest(me, "Drawn", "Land")
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	castHoldCreature(t, g, "Merfolk Pupil", "Creature — Merfolk Wizard", bnsB34Pupil, 1, 1)
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Fatal("the entry loot did not discard")
	}

	// From the graveyard: needs {1}{U}, exiles itself, loots once.
	gy := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: gy, Name: "Merfolk Pupil", TypeLine: "Creature — Merfolk Wizard", OracleID: bnsB34Pupil, Owner: me.ID, Controller: me.ID})
	putOnTopForTest(me, "Drawn 2", "Land")
	if err := g.ActivateCatalogAbility(me.ID, gy, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	fillPoolColored(me, "U", 2)
	if err := g.ActivateCatalogAbility(me.ID, gy, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	if !g.Exile.Contains(gy) {
		t.Error("the Pupil was not exiled as a cost")
	}
	hand := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Fatalf("hand %d, want %d after the draw", len(me.Hand.Cards), hand+1)
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand %d, want %d after the discard", len(me.Hand.Cards), hand)
	}
}

func TestBnsB34RiderNeedsAFoodAndMakesTheTargetBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rider := b12Push(g, me.ID, "Maraleaf Rider", "Creature — Elf Knight", bnsB34Rider, 3, 1)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	params := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{eoeCDCard(bear)}}
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, params); err == nil {
		t.Fatal("activated with no Food")
	}
	food := pushToken(g, me.ID, FoodToken())
	params.SacrificeIDs = []uuid.UUID{food}
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, food) != nil {
		t.Error("the Food was not sacrificed")
	}
	recs := blocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != rider {
		t.Fatalf("records = %+v, want one naming the Rider", recs)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rider, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if got := mustBlockNow(g)[bear]; got != rider {
		t.Errorf("the target must block %v, want the Rider", got)
	}
}

func TestBnsB34LeoninExilesAnOpponentsGraveyardCardWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leonin := b12Push(g, me.ID, "Leonin of the Lost Pride", "Creature — Cat Warrior", bnsB34Leonin, 3, 1)
	theirs := s82b10Grave(opp, "Their Card", "Sorcery", "{1}")
	mine := s82b10Grave(me, "My Card", "Sorcery", "{1}")
	s82b10Kill(g, leonin)
	s82b10Settle(t, g, me.ID, s82b10Target(theirs))
	if !g.Exile.Contains(theirs) || opp.Graveyard.Contains(theirs) {
		t.Error("the opponent's card was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard card was exiled")
	}
}

func TestBnsB34LightOfTheLegionBlessesEachWhiteCreatureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	light := b12Push(g, me.ID, "Light of the Legion", "Creature — Angel", bnsB34Light, 5, 5)
	white := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "White Bear", TypeLine: "Creature — Bear",
		Colors: []string{"W"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their White Bear", TypeLine: "Creature — Bear",
		Colors: []string{"W"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	s82b10Kill(g, light)
	s82b10Settle(t, g, me.ID, nil)
	for id, want := range map[uuid.UUID]int{white: 1, green: 0, theirs: 0} {
		if got := bnsB34Counters(g, id); got != want {
			t.Errorf("counters %d, want %d", got, want)
		}
	}
}

func TestBnsB34LightOfTheLegionMentorsALesserAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	light := b12Push(g, me.ID, "Light of the Legion", "Creature — Angel", bnsB34Light, 5, 5)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 6, 6)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{light, small, big} {
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if got := bnsB34Counters(g, small); got != 1 {
		t.Errorf("the lesser attacker has %d counters, want 1", got)
	}
	if got := bnsB34Counters(g, big); got != 0 {
		t.Errorf("the greater attacker has %d counters, want 0", got)
	}
}
