package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b07_test.go — Innistrad: Crimson Vow, batch 7 (#5495). Each test
// plays the real catalog entry and asserts what its text says.

const (
	vowB07Ridgewolf  = "5a7cd356-c07b-47dd-b089-8ecdd0e15581"
	vowB07Pup        = "cd796e12-c422-4869-8d59-794600425381"
	vowB07Runebound  = "955e4c54-14c4-49bb-94e1-c6cc800c6d91"
	vowB07Wolfrider  = "d9c42e56-1818-4b73-9cd6-1c5aaf6a5b56"
	vowB07Packleader = "7e81853c-d0c0-40bc-80fc-13a5964878b2"
	vowB07Vanguard   = "fa0bff4f-af97-4733-b9d3-a0bc9ef1e2ca"
	vowB07Slayer     = "bae41d41-9dbe-4d6b-bf5d-8a7d7041462c"
	vowB07Flame      = "24ac7e7e-b60d-4965-91c1-7ea7e6387737"
	vowB07Aim        = "b191ae0c-7d54-459f-abe3-aab5d792e514"
	vowB07Gift       = "901f0a8f-7462-4970-8d21-265c8c0188e6"
	vowB07Slinger    = "df7397b6-9a37-4012-8677-ae8e3b7c4bc3"
	vowB07Chill      = "2ee4f7ba-7238-44b5-8315-9f50502ed796"
	vowB07Markov     = "3a9ca9b2-0085-41de-9c08-60aaa8122da2"
)

func TestVowB07EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		vowB07Ridgewolf: "Hungry Ridgewolf", vowB07Pup: "Packsong Pup", vowB07Runebound: "Runebound Wolf",
		vowB07Wolfrider: "Kessig Wolfrider", vowB07Packleader: "Ascendant Packleader", vowB07Vanguard: "Hamlet Vanguard",
		vowB07Slayer: "Vampire Slayer", vowB07Flame: "Rending Flame", vowB07Aim: "Aim for the Head",
		vowB07Gift: "Gift of Fangs", vowB07Slinger: "Sawblade Slinger", vowB07Chill: "Chill of the Grave",
		vowB07Markov: "Markov Retribution",
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

// --- fixtures ------------------------------------------------------------------

func vowB07Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// vowB07Perm puts a catalog permanent with real stats on the battlefield.
func vowB07Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// vowB07Hand puts a card in the seat's hand.
func vowB07Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

func vowB07Counters(g *game.Game, id uuid.UUID) int {
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

// --- Hungry Ridgewolf --------------------------------------------------------------

func TestVowB07HungryRidgewolfGainsPowerAndTrampleBesideAnotherWolfOrWerewolf(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := vowB07Seats(g)
	wolf := vowB07Perm(g, me.ID, "Hungry Ridgewolf", "Creature — Wolf", vowB07Ridgewolf, 2, 2)
	if effectivePower(t, g, wolf) != 2 || hasEffectiveKeyword(t, g, wolf, "trample") {
		t.Fatal("alone, the Ridgewolf is a plain 2/2: it is not its own other Wolf")
	}
	pal := b12Creature(g, me.ID, "Pal", "Creature — Wolf", 1, 1)
	if effectivePower(t, g, wolf) != 3 || !hasEffectiveKeyword(t, g, wolf, "trample") {
		t.Fatalf("beside a Wolf: power %d, trample %v", effectivePower(t, g, wolf), hasEffectiveKeyword(t, g, wolf, "trample"))
	}
	if effectiveToughness(t, g, wolf) != 2 {
		t.Error("the bonus is +1/+0, not +1/+1")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pal) })
	g.RunStateChecksForTest()
	if effectivePower(t, g, wolf) != 2 || hasEffectiveKeyword(t, g, wolf, "trample") {
		t.Error("the bonus outlived the other Wolf")
	}
	b12Creature(g, me.ID, "Howler", "Creature — Human Werewolf", 2, 2)
	if effectivePower(t, g, wolf) != 3 {
		t.Error("a Werewolf did not count")
	}
}

func TestVowB07HungryRidgewolfIgnoresAnOpponentsWolf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	wolf := vowB07Perm(g, me.ID, "Hungry Ridgewolf", "Creature — Wolf", vowB07Ridgewolf, 2, 2)
	b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	if effectivePower(t, g, wolf) != 2 {
		t.Error("an opponent's Wolf gave the bonus")
	}
}

// --- Packsong Pup ------------------------------------------------------------------

func TestVowB07PacksongPupGrowsAtCombatOnlyBesideAnotherWolf(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := vowB07Seats(g)
	pup := vowB07Perm(g, me.ID, "Packsong Pup", "Creature — Wolf", vowB07Pup, 1, 1)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if n := vowB07Counters(g, pup); n != 0 {
		t.Fatalf("a lone Pup got %d counters", n)
	}

	g2 := newCatalogGame(t)
	me2, _ := vowB07Seats(g2)
	pup2 := vowB07Perm(g2, me2.ID, "Packsong Pup", "Creature — Wolf", vowB07Pup, 1, 1)
	b12Creature(g2, me2.ID, "Howler", "Creature — Werewolf", 2, 2)
	advanceTo(t, g2, game.StepBeginCombat)
	passPriorityAroundTable(t, g2)
	if n := vowB07Counters(g2, pup2); n != 1 {
		t.Fatalf("a Pup beside a Werewolf got %d counters, want 1", n)
	}
}

func TestVowB07PacksongPupGainsLifeEqualToItsPowerWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := vowB07Seats(g)
	pup := vowB07Perm(g, me.ID, "Packsong Pup", "Creature — Wolf", vowB07Pup, 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(pup, game.CounterPlusOne, 2) })
	before := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pup) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life %d, want %d (its power, counters included)", me.Life, before+3)
	}
}

// --- Runebound Wolf ----------------------------------------------------------------

func TestVowB07RuneboundWolfDealsDamageForEachWolfAndWerewolf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	wolf := vowB07Perm(g, me.ID, "Runebound Wolf", "Creature — Wolf", vowB07Runebound, 2, 2)
	b12Creature(g, me.ID, "Pal", "Creature — Wolf", 1, 1)
	b12Creature(g, me.ID, "Howler", "Creature — Human Werewolf", 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	b12Creature(g, opp.ID, "Theirs", "Creature — Wolf", 1, 1)
	advanceToMain(t, g)
	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("activated without {R}")
	}
	floatForTest(g, me, "R")
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life %d, want %d (three Wolves and Werewolves of yours)", opp.Life, before-3)
	}
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Error("targeted yourself")
	}
}

// --- Kessig Wolfrider --------------------------------------------------------------

func TestVowB07KessigWolfriderMakesAWolfForThreeGraveyardCards(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := vowB07Seats(g)
	rider := vowB07Perm(g, me.ID, "Kessig Wolfrider", "Creature — Human Knight", vowB07Wolfrider, 1, 2)
	advanceToMain(t, g)
	var yard []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		yard = append(yard, id)
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	floatForTest(g, me, "CCR")
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: yard}); err == nil {
		t.Fatal("activated with only two graveyard cards")
	}
	id := uuid.New()
	yard = append(yard, id)
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: yard}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, y := range yard {
		if !g.Exile.Contains(y) {
			t.Errorf("card %s was not exiled as a cost", y)
		}
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Wolf" {
			found = true
			eff := c.Effective()
			if eff.Power != 3 || eff.Toughness != 2 || !c.HasColor("R") || !c.IsToken() {
				t.Errorf("the token is %d/%d, colours %v", eff.Power, eff.Toughness, c.Colors)
			}
		}
	}
	if !found {
		t.Error("no Wolf token")
	}
}

// --- Ascendant Packleader ----------------------------------------------------------

func vowB07CastPackleader(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := vowB07Hand(me, game.Card{Name: "Ascendant Packleader", TypeLine: "Creature — Wolf", OracleID: vowB07Packleader,
		ManaCost: "{G}", Colors: []string{"G"}, Power: 2, Toughness: 1})
	advanceToMain(t, g)
	floatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestVowB07AscendantPackleaderEntersWithACounterOnlyBesideAManaValueFourPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := vowB07Seats(g)
	id := vowB07CastPackleader(t, g, me)
	if n := vowB07Counters(g, id); n != 0 {
		t.Fatalf("entered with %d counters and no big permanent", n)
	}

	g2 := newCatalogGame(t)
	me2, _ := vowB07Seats(g2)
	pushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant",
		ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me2.ID, Controller: me2.ID})
	id2 := vowB07CastPackleader(t, g2, me2)
	if n := vowB07Counters(g2, id2); n != 1 {
		t.Fatalf("entered with %d counters beside a mana value 4 permanent, want 1", n)
	}
	if p := effectivePowerWithCounters(t, g2, id2); p != 3 {
		t.Errorf("a 2/1 with a counter has power %d, want 3", p)
	}
}

func TestVowB07AscendantPackleaderIgnoresAnOpponentsBigPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant",
		ManaCost: "{5}", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
	id := vowB07CastPackleader(t, g, me)
	if n := vowB07Counters(g, id); n != 0 {
		t.Errorf("an opponent's permanent gave %d counters", n)
	}
}

func TestVowB07AscendantPackleaderGrowsOnAManaValueFourSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := vowB07Seats(g)
	leader := vowB07Perm(g, me.ID, "Ascendant Packleader", "Creature — Wolf", vowB07Packleader, 2, 1)
	small := vowB07Hand(me, game.Card{Name: "Small", TypeLine: "Sorcery", ManaCost: "{1}"})
	big := vowB07Hand(me, game.Card{Name: "Big", TypeLine: "Sorcery", ManaCost: "{4}"})
	advanceToMain(t, g)
	floatForTest(g, me, "CCCCC")
	if err := g.CastSpell(me.ID, small, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast small: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := vowB07Counters(g, leader); n != 0 {
		t.Fatalf("a mana value 1 spell gave %d counters", n)
	}
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast big: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := vowB07Counters(g, leader); n != 1 {
		t.Errorf("a mana value 4 spell gave %d counters, want 1", n)
	}
}

// --- Hamlet Vanguard ---------------------------------------------------------------

func TestVowB07HamletVanguardEntersWithTwoCountersPerOtherNontokenHuman(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	b12Creature(g, me.ID, "Human A", "Creature — Human Soldier", 1, 1)
	b12Creature(g, me.ID, "Human B", "Creature — Human Wizard", 1, 1)
	b12Creature(g, me.ID, "Human Token", "Token Creature — Human", 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	b12Creature(g, opp.ID, "Their Human", "Creature — Human", 1, 1)
	id := vowB07Hand(me, game.Card{Name: "Hamlet Vanguard", TypeLine: "Creature — Human Warrior", OracleID: vowB07Vanguard,
		ManaCost: "{2}{G}", Colors: []string{"G"}, Power: 1, Toughness: 1})
	advanceToMain(t, g)
	floatForTest(g, me, "CCG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := vowB07Counters(g, id); n != 4 {
		t.Errorf("entered with %d counters, want 4 (two nontoken Humans, two each)", n)
	}
	if !hasEffectiveKeyword(t, g, id, "ward") {
		t.Error("no ward")
	}
}

func TestVowB07HamletVanguardAloneEntersWithNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := vowB07Seats(g)
	id := vowB07Hand(me, game.Card{Name: "Hamlet Vanguard", TypeLine: "Creature — Human Warrior", OracleID: vowB07Vanguard,
		ManaCost: "{2}{G}", Colors: []string{"G"}, Power: 1, Toughness: 1})
	advanceToMain(t, g)
	floatForTest(g, me, "CCG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := vowB07Counters(g, id); n != 0 {
		t.Errorf("entered with %d counters; it is not its own other Human", n)
	}
}

// --- Vampire Slayer ----------------------------------------------------------------

func TestVowB07VampireSlayerDestroysAVampireItDamages(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	slayer := vowB07Perm(g, me.ID, "Vampire Slayer", "Creature — Human Soldier", vowB07Slayer, 2, 2)
	vamp := vowB07Perm(g, opp.ID, "Big Vampire", "Creature — Vampire", "", 1, 9)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(slayer, vamp, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vamp) {
		t.Error("the Vampire survived the Slayer's damage")
	}
}

func TestVowB07VampireSlayerLeavesANonVampireAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	slayer := vowB07Perm(g, me.ID, "Vampire Slayer", "Creature — Human Soldier", vowB07Slayer, 2, 2)
	bear := vowB07Perm(g, opp.ID, "Big Bear", "Creature — Bear", "", 1, 9)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(slayer, bear, 2) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a Bear was destroyed")
	}
}

// --- Rending Flame -----------------------------------------------------------------

func vowB07CastFlame(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	id := vowB07Hand(me, game.Card{Name: "Rending Flame", TypeLine: "Instant", OracleID: vowB07Flame,
		ManaCost: "{2}{R}", Colors: []string{"R"}})
	advanceToMain(t, g)
	floatForTest(g, me, "CCR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(target)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestVowB07RendingFlameAlsoHitsTheControllerOfASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	ghost := b12Creature(g, opp.ID, "Ghost", "Creature — Spirit", 2, 5)
	before := opp.Life
	vowB07CastFlame(t, g, me, ghost)
	if g.Battlefield.Contains(ghost) {
		t.Error("the Spirit survived 5 damage")
	}
	if opp.Life != before-2 {
		t.Errorf("life %d, want %d: 2 to the Spirit's controller even though it died", opp.Life, before-2)
	}
}

func TestVowB07RendingFlameSparesTheControllerOfANonSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 5)
	before := opp.Life
	vowB07CastFlame(t, g, me, bear)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived 5 damage")
	}
	if opp.Life != before {
		t.Errorf("life %d, want %d", opp.Life, before)
	}
}

func TestVowB07RendingFlameRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	id := vowB07Hand(me, game.Card{Name: "Rending Flame", TypeLine: "Instant", OracleID: vowB07Flame, ManaCost: "{2}{R}"})
	advanceToMain(t, g)
	floatForTest(g, me, "CCR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(land)}); err == nil {
		t.Error("targeted a land")
	}
}

// --- Aim for the Head --------------------------------------------------------------

func TestVowB07AimForTheHeadExilesAZombie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	zombie := b12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	id := vowB07Hand(me, game.Card{Name: "Aim for the Head", TypeLine: "Sorcery", OracleID: vowB07Aim, ManaCost: "{2}{B}"})
	advanceToMain(t, g)
	floatForTest(g, me, "CCB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0}}}); err == nil {
		t.Fatal("targeted a Bear with the Zombie bullet")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zombie, Mode: 0}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(zombie) {
		t.Error("the Zombie was not exiled")
	}
}

func TestVowB07AimForTheHeadMakesAnOpponentExileTwoFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	keep := handCardFull(opp, "Keep", "Instant", "", "", nil)
	a := handCardFull(opp, "A", "Instant", "", "", nil)
	b := handCardFull(opp, "B", "Instant", "", "", nil)
	id := vowB07Hand(me, game.Card{Name: "Aim for the Head", TypeLine: "Sorcery", OracleID: vowB07Aim, ManaCost: "{2}{B}"})
	advanceToMain(t, g)
	floatForTest(g, me, "CCB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, opp.ID)
	if c == nil || c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Fatalf("the opponent's prompt is %+v, want exactly two cards", c)
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) || !opp.Hand.Contains(keep) {
		t.Error("the two chosen cards were not the ones exiled")
	}
}

// --- Gift of Fangs -----------------------------------------------------------------

func vowB07Enchant(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID) {
	t.Helper()
	id := vowB07Hand(me, game.Card{Name: "Gift of Fangs", TypeLine: "Enchantment — Aura", OracleID: vowB07Gift,
		ManaCost: "{B}", Colors: []string{"B"}})
	advanceToMain(t, g)
	floatForTest(g, me, "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(host)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestVowB07GiftOfFangsPumpsAVampire(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := vowB07Seats(g)
	vamp := b12Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2)
	vowB07Enchant(t, g, me, vamp)
	if effectivePower(t, g, vamp) != 4 || effectiveToughness(t, g, vamp) != 4 {
		t.Errorf("a Vampire is %d/%d, want 4/4", effectivePower(t, g, vamp), effectiveToughness(t, g, vamp))
	}
}

func TestVowB07GiftOfFangsShrinksANonVampire(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := vowB07Seats(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	vowB07Enchant(t, g, me, bear)
	if effectivePower(t, g, bear) != 1 || effectiveToughness(t, g, bear) != 1 {
		t.Errorf("a Bear is %d/%d, want 1/1", effectivePower(t, g, bear), effectiveToughness(t, g, bear))
	}
}

func TestVowB07GiftOfFangsKillsAToughnessTwoNonVampire(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := vowB07Seats(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	vowB07Enchant(t, g, me, bear)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 Bear survived -2/-2")
	}
}

// --- Sawblade Slinger --------------------------------------------------------------

func vowB07CastSlinger(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := vowB07Hand(me, game.Card{Name: "Sawblade Slinger", TypeLine: "Creature — Human Archer", OracleID: vowB07Slinger,
		ManaCost: "{3}{G}", Colors: []string{"G"}, Power: 4, Toughness: 3})
	advanceToMain(t, g)
	floatForTest(g, me, "CCCG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// vowB07PickMode answers the enters trigger's mode prompt with the modes.
func vowB07PickMode(t *testing.T, g *game.Game, me *game.Player, modes ...int) {
	t.Helper()
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mode.ID, me.ID, modes); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestVowB07SawbladeSlingerCanChooseNeitherMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	zombie := b12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	art := b12Permanent(g, opp.ID, "Rock", "Artifact")
	id := vowB07CastSlinger(t, g, me)
	vowB07PickMode(t, g, me)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || !g.Battlefield.Contains(zombie) || !g.Battlefield.Contains(art) {
		t.Error("choosing no mode changed the board")
	}
}

func TestVowB07SawbladeSlingerDestroysAnOpponentsArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	vowB07CastSlinger(t, g, me)
	vowB07PickMode(t, g, me, 0)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("accepted my own artifact")
		}
	}
	avB34Pick(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the artifact survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my artifact was destroyed")
	}
}

func TestVowB07SawbladeSlingerFightsAnOpponentsZombie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	zombie := b12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 4)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	slinger := vowB07CastSlinger(t, g, me)
	vowB07PickMode(t, g, me, 1)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Error("accepted a Bear for the Zombie clause")
		}
	}
	avB34Pick(t, g, me.ID, zombie)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(zombie) {
		t.Error("the 2/4 Zombie survived a 4-power fight")
	}
	c := findBattlefieldCardByID(g, slinger)
	if c == nil || c.DamageMarked != 2 {
		t.Errorf("the Slinger should have taken 2 damage: %+v", c)
	}
}

// --- Chill of the Grave ------------------------------------------------------------

func vowB07CastChill(t *testing.T, g *game.Game, me *game.Player, mana string, target uuid.UUID) error {
	t.Helper()
	id := vowB07Hand(me, game.Card{Name: "Chill of the Grave", TypeLine: "Instant", OracleID: vowB07Chill,
		ManaCost: "{2}{U}", Colors: []string{"U"}})
	advanceToMain(t, g)
	floatForTest(g, me, mana)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(target)})
}

func TestVowB07ChillOfTheGraveCostsOneLessBesideAZombie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	victim := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := vowB07CastChill(t, g, me, "CU", victim); err == nil {
		t.Fatal("cast for {1}{U} with no Zombie")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := vowB07Seats(g2)
	victim2 := b12Creature(g2, opp2.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g2, me2.ID, "Zombie", "Creature — Zombie", 1, 1)
	handBefore := me2.Hand.Size()
	if err := vowB07CastChill(t, g2, me2, "CU", victim2); err != nil {
		t.Fatalf("cast for {1}{U} beside a Zombie: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d, want %d: the spell left the hand and a card was drawn", me2.Hand.Size(), handBefore+1)
	}
	c := findBattlefieldCardByID(g2, victim2)
	if c == nil || !c.Tapped {
		t.Error("the creature was not tapped")
	}
}

func TestVowB07ChillOfTheGraveKeepsTheCreatureTappedThroughItsNextUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	victim := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := vowB07CastChill(t, g, me, "CCU", victim); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepUpkeep)
	if c := findBattlefieldCardByID(g, victim); c == nil || !c.Tapped {
		t.Error("the creature untapped during its controller's next untap step")
	}
}

// --- Markov Retribution ------------------------------------------------------------

func vowB07CastMarkov(t *testing.T, g *game.Game, me *game.Player, modes []int, targets []game.TargetRef) error {
	t.Helper()
	id := vowB07Hand(me, game.Card{Name: "Markov Retribution", TypeLine: "Sorcery", OracleID: vowB07Markov,
		ManaCost: "{2}{R}", Colors: []string{"R"}})
	advanceToMain(t, g)
	floatForTest(g, me, "CCR")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: modes, Targets: targets})
}

func TestVowB07MarkovRetributionPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := vowB07CastMarkov(t, g, me, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, mine) != 3 || effectivePower(t, g, theirs) != 2 {
		t.Errorf("powers %d and %d, want 3 and 2", effectivePower(t, g, mine), effectivePower(t, g, theirs))
	}
}

func TestVowB07MarkovRetributionVampireBitesAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	vamp := b12Creature(g, me.ID, "Vampire", "Creature — Vampire", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 4)
	bite := func(biter, prey uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: biter, Mode: 0, Slot: 0}, {Kind: game.TargetCard, ID: prey, Mode: 0, Slot: 1}}
	}
	if err := vowB07CastMarkov(t, g, me, []int{1}, bite(bear, victim)); err == nil {
		t.Fatal("a Bear was accepted as the Vampire")
	}
	if err := vowB07CastMarkov(t, g, me, []int{1}, bite(vamp, vamp)); err == nil {
		t.Fatal("the Vampire was accepted as the other creature")
	}
	if err := vowB07CastMarkov(t, g, me, []int{1}, bite(vamp, victim)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, victim); c == nil || c.DamageMarked != 3 {
		t.Errorf("the victim should have 3 damage: %+v", c)
	}
	if c := findBattlefieldCardByID(g, vamp); c == nil || c.DamageMarked != 0 {
		t.Errorf("the bite is one-sided: %+v", c)
	}
}

func TestVowB07MarkovRetributionBothModesBiteWithThePumpedPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := vowB07Seats(g)
	vamp := b12Creature(g, me.ID, "Vampire", "Creature — Vampire", 3, 3)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 9)
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: vamp, Mode: 1, Slot: 0},
		{Kind: game.TargetCard, ID: victim, Mode: 1, Slot: 1},
	}
	if err := vowB07CastMarkov(t, g, me, []int{0, 1}, targets); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, victim); c == nil || c.DamageMarked != 4 {
		t.Errorf("the victim should have 4 damage (3 power + the pump): %+v", c)
	}
}
