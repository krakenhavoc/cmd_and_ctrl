package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m1_cards_test.go — Marvel Super Heroes Commander, the M–Z half,
// first batch (#4455). Each test plays the real catalog entry and asserts
// what its text says.

const (
	mscM1RaysOracle       = "053a90b8-61b1-4390-afc0-c5eaef0138a9"
	mscM1MeleeOracle      = "5ee9940a-0569-45b6-b28c-0ef3914013a9"
	mscM1HelixOracle      = "82e52a09-ce6f-417d-8460-8806df63f0bf"
	mscM1RendezvousOracle = "8ca585b5-25c4-41cc-a0b9-dc66ae803d04"
	mscM1StrikeOracle     = "8147a759-e387-48f6-9c1e-bb39622ed553"
	mscM1SpreeOracle      = "f0371d3d-ca11-432a-bebd-684caad5781f"
	mscM1ThwipOracle      = "f7d11254-b48a-4016-a035-c0d6642fdfae"
	mscM1SoldierOnOracle  = "f522e161-f249-4add-8326-b6ab0de8ba57"
	mscM1StanceOracle     = "1f938353-9081-4dc1-b8e7-d18ece038bd4"
	mscM1TacticsOracle    = "15948fdf-3f9e-4eed-b85b-c5a554ff0571"
	mscM1ShieldGuardOracl = "a9268e61-2697-46ff-b219-4f2be646cbe4"
	mscM1PietroOracle     = "c5c1fb0f-b9cf-48a0-a4ea-e995e222e4b4"
	mscM1WhizzerOracle    = "aa86bce1-9cde-44d8-9455-24502221f31d"
	mscM1WarriorsOracle   = "7abbbe18-a2e5-47ce-9c65-8008e5153a4d"
	mscM1FrogManOracle    = "b0c2938d-8370-4786-8ddd-126c61538bdd"
	mscM1WandaOracle      = "488591d4-3dfb-4e1e-b0c4-1950ff92b848"
)

func TestMscM1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mscM1RaysOracle:       "Repulsor Rays",
		mscM1MeleeOracle:      "Marvelous Melee",
		mscM1HelixOracle:      "Void Helix",
		mscM1RendezvousOracle: "Romantic Rendezvous",
		mscM1StrikeOracle:     "Sudden Strike",
		mscM1SpreeOracle:      "Smashing Spree",
		mscM1ThwipOracle:      "Thwip!",
		mscM1SoldierOnOracle:  "Soldier On",
		mscM1StanceOracle:     "Valorous Stance",
		mscM1TacticsOracle:    "Spectacular Tactics",
		mscM1ShieldGuardOracl: "Wakandan Shield Guard",
		mscM1PietroOracle:     "Quicksilver, Pietro Maximoff",
		mscM1WhizzerOracle:    "The Whizzer, Classic Speedster",
		mscM1WarriorsOracle:   "Warriors of Wakanda",
		mscM1FrogManOracle:    "The Fabulous Frog-Man",
		mscM1WandaOracle:      "Scarlet Witch, Wanda Maximoff",
	}
	for oracle, name := range want {
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

// mscM1Attacking marks a battlefield creature as attacking a defender.
func mscM1Attacking(g *game.Game, attacker, defender uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].AttackingTarget = defender
			}
		}
	})
}

func TestMscM1RepulsorRaysDealsThree(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Three Toughness", "Creature — Wall", 0, 3)
	big := b12Creature(g, opp.ID, "Four Toughness", "Creature — Wall", 0, 4)
	castCatalogSpell(t, g, "Repulsor Rays", "Sorcery", mscM1RaysOracle, b16TargetCard(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a creature with toughness 3 survived 3 damage")
	}
	castCatalogSpell(t, g, "Repulsor Rays", "Sorcery", mscM1RaysOracle, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, big)
	if !ok || card.DamageMarked != 3 {
		t.Errorf("toughness 4 creature: on battlefield %v, damage %d, want 3", ok, card.DamageMarked)
	}
}

func TestMscM1RepulsorRaysCannotTargetAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	err := castCatalogSpellErr(t, g, "Repulsor Rays", "Sorcery", mscM1RaysOracle, b16TargetPlayer(g.Seats[1].ID))
	if err == nil {
		t.Error("a player was accepted as the target of a creature-only spell")
	}
}

func TestMscM1MarvelousMeleeDealsSix(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Six Toughness", "Creature — Wall", 0, 6)
	big := b12Creature(g, opp.ID, "Seven Toughness", "Creature — Wall", 0, 7)
	castCatalogSpell(t, g, "Marvelous Melee", "Instant", mscM1MeleeOracle, b16TargetCard(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a creature with toughness 6 survived 6 damage")
	}
	castCatalogSpell(t, g, "Marvelous Melee", "Instant", mscM1MeleeOracle, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, big)
	if !ok || card.DamageMarked != 6 {
		t.Errorf("toughness 7 creature: on battlefield %v, damage %d, want 6", ok, card.DamageMarked)
	}
}

func TestMscM1MarvelousMeleeBasicLandcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := cycleFromHand(t, g, "Marvelous Melee", "Instant", mscM1MeleeOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the landcycling search did not ask")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Fatal("no basic land reached the hand")
	}
}

func TestMscM1VoidHelixDamagesAnyTargetAndGainsFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Void Helix", "Instant", mscM1HelixOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-5 || me.Life != life+5 {
		t.Errorf("life: opponent %d -> %d, me %d -> %d, want -5 and +5", oppLife, opp.Life, life, me.Life)
	}
	wall := b12Creature(g, opp.ID, "Five Toughness", "Creature — Wall", 0, 5)
	castCatalogSpell(t, g, "Void Helix", "Instant", mscM1HelixOracle, b16TargetCard(wall))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("a creature with toughness 5 survived 5 damage")
	}
	if me.Life != life+10 {
		t.Errorf("life %d after two Helixes, want %d", me.Life, life+10)
	}
}

func TestMscM1RomanticRendezvousDiscardsThenDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three")
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitched", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Romantic Rendezvous", "Sorcery", mscM1RendezvousOracle, nil)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the chosen card was not discarded")
	}
	if got := lib - me.Library.Size(); got != 2 {
		t.Errorf("library fell by %d, want 2", got)
	}
	if got := me.Hand.Size(); got != hand-1+2 {
		t.Errorf("hand %d -> %d, want one card discarded and two drawn", hand, got)
	}
}

func TestMscM1SuddenStrikeDestroysAnAttackerOrBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	crtBlock(g, blocker, attacker, me.ID)
	if err := castCatalogSpellErr(t, g, "Sudden Strike", "Instant", mscM1StrikeOracle, b16TargetCard(idle)); err == nil {
		t.Error("a creature that is neither attacking nor blocking was accepted")
	}
	castCatalogSpell(t, g, "Sudden Strike", "Instant", mscM1StrikeOracle, b16TargetCard(attacker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking creature survived")
	}
	castCatalogSpell(t, g, "Sudden Strike", "Instant", mscM1StrikeOracle, b16TargetCard(blocker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Error("the blocking creature survived")
	}
}

func TestMscM1SmashingSpreePumpsAnAttackerWithTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	mscM1Attacking(g, bear, opp.ID)
	if err := castCatalogSpellErr(t, g, "Smashing Spree", "Instant", mscM1SpreeOracle, b16TargetCard(idle)); err == nil {
		t.Error("a creature that is not attacking was accepted")
	}
	castCatalogSpell(t, g, "Smashing Spree", "Instant", mscM1SpreeOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Errorf("attacker is %d/%d, want 5/5", p, tg)
	}
	assertKeywords(t, g, bear, "trample")
}

func TestMscM1ThwipPumpsAndFliesAndASpiderGainsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	spider := b12Creature(g, me.ID, "My Spider", "Creature — Spider Human Hero", 1, 1)
	life := me.Life
	castCatalogSpell(t, g, "Thwip!", "Instant", mscM1ThwipOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("bear is %d/%d, want 4/4", p, tg)
	}
	assertKeywords(t, g, bear, "flying")
	if me.Life != life {
		t.Errorf("life %d -> %d for a non-Spider, want no change", life, me.Life)
	}
	castCatalogSpell(t, g, "Thwip!", "Instant", mscM1ThwipOracle, b16TargetCard(spider))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, spider), effectiveToughness(t, g, spider); p != 3 || tg != 3 {
		t.Errorf("Spider is %d/%d, want 3/3", p, tg)
	}
	assertKeywords(t, g, spider, "flying")
	if me.Life != life+2 {
		t.Errorf("life %d -> %d for a Spider, want +2", life, me.Life)
	}
}

func TestMscM1SoldierOnAddsTwoCountersAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	if !b16Tapped(t, g, bear) {
		t.Fatal("the creature did not tap")
	}
	castCatalogSpell(t, g, "Soldier On", "Instant", mscM1SoldierOnOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, bear); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
	if b16Tapped(t, g, bear) {
		t.Error("the creature is still tapped")
	}
}

func TestMscM1ValorousStanceBothModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castModal(t, g, "Valorous Stance", "Instant", mscM1StanceOracle, []int{0}, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, bear, "indestructible")

	small := b12Creature(g, opp.ID, "Three Toughness", "Creature — Wall", 0, 3)
	big := b12Creature(g, opp.ID, "Four Toughness", "Creature — Wall", 0, 4)
	if _, err := mscM1TryModal(t, g, "Valorous Stance", "Instant", mscM1StanceOracle, []int{1}, b16TargetCard(small)); err == nil {
		t.Error("a toughness 3 creature was accepted for the destroy mode")
	}
	castModal(t, g, "Valorous Stance", "Instant", mscM1StanceOracle, []int{1}, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the toughness 4 creature survived")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the toughness 3 creature was destroyed")
	}
}

// mscM1TryModal is castModal for a cast that may be refused.
func mscM1TryModal(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

func TestMscM1SpectacularTacticsBothModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := mscM1TryModal(t, g, "Spectacular Tactics", "Instant", mscM1TacticsOracle, []int{0}, b16TargetCard(theirs)); err == nil {
		t.Error("an opponent's creature was accepted for the counter mode")
	}
	castModal(t, g, "Spectacular Tactics", "Instant", mscM1TacticsOracle, []int{0}, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, bear); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	assertKeywords(t, g, bear, "hexproof")

	small := b12Creature(g, opp.ID, "Three Power", "Creature — Wall", 3, 5)
	big := b12Creature(g, opp.ID, "Four Power", "Creature — Wall", 4, 1)
	if _, err := mscM1TryModal(t, g, "Spectacular Tactics", "Instant", mscM1TacticsOracle, []int{1}, b16TargetCard(small)); err == nil {
		t.Error("a power 3 creature was accepted for the destroy mode")
	}
	castModal(t, g, "Spectacular Tactics", "Instant", mscM1TacticsOracle, []int{1}, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the power 4 creature survived")
	}
}

func TestMscM1WakandanShieldGuardMakesASoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Wakandan Shield Guard", "Creature — Human Soldier", mscM1ShieldGuardOracl, 1, 1)
	if n := mshB1Tokens(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("%d Soldier tokens, want 1", n)
	}
}

func TestMscM1KeywordBodies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Quicksilver, Pietro Maximoff", "Legendary Creature — Mutant Hero", mscM1PietroOracle, 3, 2, []string{"haste"}},
		{"The Whizzer, Classic Speedster", "Legendary Creature — Human Hero", mscM1WhizzerOracle, 3, 3, []string{"first strike", "haste"}},
		{"Warriors of Wakanda", "Creature — Human Warrior", mscM1WarriorsOracle, 4, 4, []string{"trample"}},
		{"The Fabulous Frog-Man", "Legendary Creature — Human Hero", mscM1FrogManOracle, 3, 3, []string{"reach"}},
		{"Scarlet Witch, Wanda Maximoff", "Legendary Creature — Mutant Warlock Hero", mscM1WandaOracle, 2, 3, []string{"menace"}},
	} {
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
		assertKeywords(t, g, id, tc.keywords...)
	}
}
