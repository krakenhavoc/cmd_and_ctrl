package effects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s36_cards_test.go — batch 36, morph payoffs. Each test plays the
// real catalog entry.

const (
	s36AinokOracle        = "7a5dae5f-29ab-4996-b236-c754de8769bc"
	s36DragonslayerOracle = "862af5c4-cb34-4bfa-a7ff-209e103defd6"
	s36DenProtectorOracle = "59324efc-65b7-4b31-89c5-17aed5c7eeb1"
	s36AmbushersOracle    = "2373e035-25f7-473a-9386-2093f77d3f94"
)

// s36CastFaceDown casts a card from the active seat's hand face down by
// its alternative cost key ("morph" or "megamorph") and resolves it.
func s36CastFaceDown(t *testing.T, g *game.Game, name, typeLine, oracle, key string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: key}); err != nil {
		t.Fatalf("cast %s face down: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if c := findCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatalf("%s did not enter face down", name)
	}
	return id
}

func s36TurnUp(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana ...string) {
	t.Helper()
	mkmMPTurnFaceUp(t, g, p, id, mana...)
}

func TestCmdr2S36MorphCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s36AinokOracle: "Ainok Survivalist", s36DragonslayerOracle: "Hidden Dragonslayer",
		s36DenProtectorOracle: "Den Protector", s36AmbushersOracle: "Salt Road Ambushers",
	} {
		def, ok := Lookup(oracle)
		if !ok || def.Completeness != CompletenessFull {
			t.Errorf("%s: registered=%v completeness=%v", name, ok, def.Completeness)
		}
		alt := game.FaceDownCastFor(oracle)
		if alt == nil || alt.FaceDown == nil || !alt.FaceDown.FaceUpCounter {
			t.Errorf("%s has no megamorph declaration", name)
		}
	}
}

func TestCmdr2S36AinokSurvivalistDestroysAnOpponentsArtifactWhenTurnedUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	id := s36CastFaceDown(t, g, "Ainok Survivalist", "Creature — Dog Shaman", s36AinokOracle, "megamorph", 2, 1)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("entering face down triggered the Survivalist")
	}
	s36TurnUp(t, g, me, id, "G", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the opponent's artifact survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own artifact was destroyed")
	}
	if n := counterCount(g, id, game.CounterPlusOne); n != 1 {
		t.Errorf("megamorph counters = %d, want 1", n)
	}
}

func TestCmdr2S36HiddenDragonslayerDestroysAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	id := s36CastFaceDown(t, g, "Hidden Dragonslayer", "Creature — Human Warrior", s36DragonslayerOracle, "megamorph", 2, 1)
	s36TurnUp(t, g, me, id, "W", "C", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the power-4 creature survived")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the power-3 creature was destroyed")
	}
}

func TestCmdr2S36HiddenDragonslayerCannotTargetAPowerThreeCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	id := s36CastFaceDown(t, g, "Hidden Dragonslayer", "Creature — Human Warrior", s36DragonslayerOracle, "megamorph", 2, 1)
	s36TurnUp(t, g, me, id, "W", "C", "C")
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("a power-3 creature was destroyed")
	}
}

func TestCmdr2S36DenProtectorReturnsACardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	card := b17GraveyardCard(me, "Old Spell", "Sorcery", "{1}")
	id := s36CastFaceDown(t, g, "Den Protector", "Creature — Human Warrior", s36DenProtectorOracle, "megamorph", 2, 1)
	s36TurnUp(t, g, me, id, "G", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, card)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the card did not return to hand")
	}
}

func TestCmdr2S36DenProtectorSmallerCreaturesCannotBlockIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	protector := b12Push(g, me.ID, "Den Protector", "Creature — Human Warrior", s36DenProtectorOracle, 3, 1)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	small := b12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 2, 2)
	equal := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	brAttack(t, g, protector, bear)
	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, small, protector) {
		t.Error("a power-2 creature may block the power-3 Den Protector")
	}
	if !brOffers(offered, small, bear) {
		t.Error("Den Protector's rule also stops blocks of another attacker")
	}
	if !brOffers(offered, equal, protector) {
		t.Error("a creature with equal power may not block (less than is strict)")
	}
}

func TestCmdr2S36SaltRoadAmbushersGivesTurnedUpCreaturesTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Salt Road Ambushers", "Creature — Dog Warrior", s36AmbushersOracle, 3, 3)
	id := s36CastFaceDown(t, g, "Ainok Survivalist", "Creature — Dog Shaman", s36AinokOracle, "megamorph", 2, 1)
	if n := counterCount(g, id, game.CounterPlusOne); n != 0 {
		t.Fatalf("entering face down got %d counters", n)
	}
	s36TurnUp(t, g, me, id, "G", "C")
	passPriorityAroundTable(t, g)
	// one megamorph counter plus the Ambushers' two
	if n := counterCount(g, id, game.CounterPlusOne); n != 3 {
		t.Errorf("counters = %d, want 3", n)
	}
}

func TestCmdr2S36SaltRoadAmbushersIgnoresAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, opp.ID, "Salt Road Ambushers", "Creature — Dog Warrior", s36AmbushersOracle, 3, 3)
	id := s36CastFaceDown(t, g, "Ainok Survivalist", "Creature — Dog Shaman", s36AinokOracle, "megamorph", 2, 1)
	s36TurnUp(t, g, me, id, "G", "C")
	passPriorityAroundTable(t, g)
	if n := counterCount(g, id, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d, want only the megamorph one", n)
	}
}

const (
	s36HydraOracle   = "637ddd65-112d-40f7-9bab-a61724d9165c"
	s36PhoenixOracle = "fb2d7c3d-b68f-410e-9a73-cfcbfe702ade"
)

func TestCmdr2S36HoodedHydraTurnsUpWithFiveCountersAndLeavesSnakes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := s36CastFaceDown(t, g, "Hooded Hydra", "Creature — Snake Hydra", s36HydraOracle, "morph", 0, 0)
	if n := counterCount(g, id, game.CounterPlusOne); n != 0 {
		t.Fatalf("a face-down Hydra has %d counters", n)
	}
	s36TurnUp(t, g, me, id, "G", "G", "C", "C", "C")
	if n := counterCount(g, id, game.CounterPlusOne); n != 5 {
		t.Fatalf("counters = %d, want 5", n)
	}
	before := tokensNamedControlledBy(g, "Snake", me.ID)
	killCreature(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Snake", me.ID) - before; got != 5 {
		t.Errorf("Snakes = %d, want one per counter (5)", got)
	}
}

func TestCmdr2S36HoodedHydraCastFaceUpEntersWithXCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	active := me
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Hooded Hydra", TypeLine: "Creature — Snake Hydra", OracleID: s36HydraOracle,
		ManaCost: "{X}{G}{G}", Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	g.WithWriteLock(func() { active.ManaPool.AddMana(manaTokens("G", "G", "C", "C", "C")...) })
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, id, game.CounterPlusOne); n != 3 {
		t.Errorf("counters = %d, want X=3", n)
	}
}

func TestCmdr2S36AshcloudPhoenixReturnsFaceDownAndTurnsUpForDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b12Push(g, me.ID, "Ashcloud Phoenix", "Creature — Phoenix", s36PhoenixOracle, 4, 1)
	killCreature(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	c := findCardAnywhere(t, g, id)
	if !g.Battlefield.Contains(id) || !c.FaceDownIsPermanent() {
		t.Fatalf("the Phoenix did not return face down (battlefield=%v faceDown=%v)", g.Battlefield.Contains(id), c.FaceDownIsPermanent())
	}
	if c.Controller != me.ID {
		t.Error("it did not return under my control")
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	s36TurnUp(t, g, me, id, "R", "R", "C", "C", "C", "C")
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-2 {
			t.Errorf("seat life %d → %d, want 2 damage", lives[p.ID], p.Life)
		}
	}
}

const s36RaptorOracle = "a7d4f5aa-26d4-4363-aa5a-a26d6f791b82"

// s36RaptorInGraveyard seeds a Raptor in my graveyard and turns a
// megamorph creature of mine face up.
func s36RaptorTrigger(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	raptor := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: raptor, Name: "Deathmist Raptor", TypeLine: "Creature — Dinosaur Beast", OracleID: s36RaptorOracle,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	id := s36CastFaceDown(t, g, "Ainok Survivalist", "Creature — Dog Shaman", s36AinokOracle, "megamorph", 2, 1)
	s36TurnUp(t, g, me, id, "G", "C")
	return raptor
}

func TestCmdr2S36DeathmistRaptorReturnsFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	raptor := s36RaptorTrigger(t, g, me)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(raptor) {
		t.Fatal("the Raptor did not return")
	}
	if c := findCardAnywhere(t, g, raptor); c.FaceDownIsPermanent() {
		t.Error("the Raptor returned face down")
	}
}

func TestCmdr2S36DeathmistRaptorReturnsFaceDownAndMegamorphsUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	raptor := s36RaptorTrigger(t, g, me)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(raptor) || !findCardAnywhere(t, g, raptor).FaceDownIsPermanent() {
		t.Fatal("the Raptor did not return face down")
	}
	s36TurnUp(t, g, me, raptor, "G", "C", "C", "C", "C")
	if n := counterCount(g, raptor, game.CounterPlusOne); n != 1 {
		t.Errorf("megamorph counters = %d, want 1", n)
	}
}

func TestCmdr2S36DeathmistRaptorMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	raptor := s36RaptorTrigger(t, g, me)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raptor) {
		t.Error("the Raptor returned although I declined")
	}
}

const (
	s36KheruOracle       = "c01411e0-77b2-4e65-a369-5dbe13745769"
	s36SpelljackerOracle = "b23c28ec-dc4d-4072-8fd8-6640a104335a"
)

func TestCmdr2S36KheruSpellsnatcherExilesASpellAndLetsYouCastItFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := s36CastFaceDown(t, g, "Kheru Spellsnatcher", "Creature — Snake Wizard", s36KheruOracle, "morph", 3, 3)
	life := opp.Life
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	s36TurnUp(t, g, me, id, "U", "U", "C", "C", "C", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bolt) {
		t.Fatal("the Bolt was not exiled")
	}
	if me.Graveyard.Contains(bolt) {
		t.Error("the Bolt went to the graveyard")
	}
	if opp.Life != life {
		t.Fatalf("the countered Bolt dealt damage: %d → %d", life, opp.Life)
	}
	// "for as long as it remains exiled": cast later, for no mana.
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{playerTarget(opp.ID)}, Strict: true, FromZone: "exile",
	}); err != nil {
		t.Fatalf("cast the exiled Bolt for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-3)
	}
}

func TestCmdr2S36KheruSpellsnatcherDoesNothingFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s36CastFaceDown(t, g, "Kheru Spellsnatcher", "Creature — Snake Wizard", s36KheruOracle, "morph", 3, 3)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("entering face down triggered the Spellsnatcher")
	}
}

// s36SpelljackerOverTheirBolt: the opponent is the active player and
// casts a Bolt at me; I flash in the Spelljacker and target it.
func s36SpelljackerOverTheirBolt(t *testing.T, g *game.Game) (spelljacker, bolt uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 })
	bolt = castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(me.ID)})
	g.WithWriteLock(func() { g.Turn.PriorityHolder = 0 })
	spelljacker = castFromSeat(t, g, me, "Smirking Spelljacker", "Creature — Djinn Wizard Rogue", s36SpelljackerOracle, nil)
	passUntilOnBattlefield(t, g, spelljacker)
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	_ = opp
	return spelljacker, bolt
}

func TestCmdr2S36SmirkingSpelljackerExilesTheirSpellLinkedToIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	spelljacker, bolt := s36SpelljackerOverTheirBolt(t, g)
	if !g.Exile.Contains(bolt) {
		t.Fatal("the Bolt was not exiled")
	}
	if me.Life != life {
		t.Errorf("the exiled Bolt still resolved: life %d → %d", life, me.Life)
	}
	src, _ := g.LookupCardForEffect(spelljacker)
	linked := cmdr2S36CardsExiledWithThis(g, &src)
	if len(linked) != 1 || linked[0].InstanceID != bolt {
		t.Errorf("cards exiled with the Spelljacker = %v, want the Bolt", linked)
	}
}

func TestCmdr2S36SmirkingSpelljackerCastsTheExiledCardWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spelljacker, bolt := s36SpelljackerOverTheirBolt(t, g)
	// It is their turn and the Spelljacker is new: play on to my next
	// turn's declare-attackers step.
	for i := 0; i < 80 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepDeclareAttackers); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("never reached my declare-attackers step (seat %d, %s)", g.Turn.ActiveSeat, g.Turn.Step)
	}
	oppLife := opp.Life
	if err := g.DeclareAttacker(spelljacker, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	lockInAttacks(t, g)
	for i := 0; i < 8 && resolutionCastOfferFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, bolt)
	if err := castThroughResolutionCast(t, g, me.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{playerTarget(opp.ID)}, Strict: true,
	}); err != nil {
		t.Fatalf("free cast of the exiled Bolt: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life > oppLife-3 {
		t.Errorf("opponent at %d, want at most %d", opp.Life, oppLife-3)
	}
}

func TestCmdr2S36SmirkingSpelljackerWithNothingExiledDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sj := b12Push(g, me.ID, "Smirking Spelljacker", "Creature — Djinn Wizard Rogue", s36SpelljackerOracle, 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	triggers := func() (n int) {
		for _, ev := range g.Events {
			if ev.Kind == game.EventTrigger {
				n++
			}
		}
		return n
	}
	before := triggers()
	if err := g.DeclareAttacker(sj, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := triggers(); got != before {
		t.Errorf("the attack trigger fired with nothing exiled with it (%d new)", got-before)
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a cast offer opened with nothing exiled")
	}
}
