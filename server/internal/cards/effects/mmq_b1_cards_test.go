package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mmq_b1_cards_test.go — Mercadian Masques, batch 1 (#5619). Each test
// plays the real catalog entry and asserts what its text says.

const (
	mmqB1CrashOracle       = "562af5da-1080-46d2-8a38-71f06b7e8457"
	mmqB1ThunderclapOracle = "02f5ceb4-8b9c-4107-972b-6eec365c4a6c"
	mmqB1PulverizeOracle   = "d403d1f3-5a92-4f4c-aff2-cff6a0b3b1ec"
	mmqB1RouseOracle       = "475099d9-4743-420a-ada8-3932f5ef9e82"
	mmqB1CaveInOracle      = "7d937265-02f5-4f2c-9fef-c5930e13b3be"
	mmqB1VineDryadOracle   = "990cd7de-66a2-4def-bf00-390b2682e995"
	mmqB1TidalBoreOracle   = "513ad32a-2887-4825-adad-86a87cf1d2cb"
	mmqB1SaprazzanOracle   = "d20b62fc-e248-4c84-87ac-7319b92531d4"
	mmqB1KyrenOracle       = "41b46aa5-6091-4bc4-8d17-d2c6b5fe1c5e"
	mmqB1ChoArrimOracle    = "780c9cb2-9d98-4789-87e0-cb81db477204"
	mmqB1DeepwoodOracle    = "414eb407-dd39-41d2-bf25-a7a77ff35cb4"
	mmqB1KrisMageOracle    = "d53623b0-f7d8-401f-a8f1-80034f6f20df"
	mmqB1SergeantOracle    = "9b0a14b9-977d-4e25-a519-eaddb0f94b6a"
	mmqB1LieutenantOracle  = "ddd2a87e-4020-4408-b4e6-a2120d485855"
)

func TestMmqB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mmqB1CrashOracle:       "Crash",
		mmqB1ThunderclapOracle: "Thunderclap",
		mmqB1PulverizeOracle:   "Pulverize",
		mmqB1RouseOracle:       "Rouse",
		mmqB1CaveInOracle:      "Cave-In",
		mmqB1VineDryadOracle:   "Vine Dryad",
		mmqB1TidalBoreOracle:   "Tidal Bore",
		mmqB1SaprazzanOracle:   "Saprazzan Legate",
		mmqB1KyrenOracle:       "Kyren Legate",
		mmqB1ChoArrimOracle:    "Cho-Arrim Legate",
		mmqB1DeepwoodOracle:    "Deepwood Legate",
		mmqB1KrisMageOracle:    "Kris Mage",
		mmqB1SergeantOracle:    "Ramosian Sergeant",
		mmqB1LieutenantOracle:  "Ramosian Lieutenant",
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

// mmqB1Push seeds a permanent on the battlefield, untapped and not
// summoning sick.
func mmqB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

// mmqB1Cast casts a hand card for its alternative cost `key` under the
// strict mana gate, with an empty pool, so nothing is paid but the
// alternative cost.
func mmqB1Cast(g *game.Game, caster, id uuid.UUID, key string, pay []uuid.UUID, targets []game.TargetRef) error {
	return g.CastSpell(caster, id, game.CastSpellParams{
		Strict: true, AlternativeCost: key, AltCostIDs: pay, Targets: targets,
	})
}

// --- Crash, Thunderclap and Pulverize: sacrifice Mountains instead ------------

func TestMmqB1CrashSacrificesAMountainToDestroyAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	crash := handCardFull(me, "Crash", "Instant", "{2}{R}", mmqB1CrashOracle, []string{"R"})
	rock := mmqB1Push(g, them.ID, "Mana Rock", "Artifact", "", 0, 0)
	island := seedLand(g, me.ID, "Island", "Basic Land — Island", "")
	at := cardRefs(rock)

	if err := mmqB1Cast(g, me.ID, crash, "sacrifice", []uuid.UUID{island}, at); err == nil {
		t.Fatal("an Island paid \"sacrifice a Mountain\"")
	}
	mountain := seedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	if err := mmqB1Cast(g, me.ID, crash, "sacrifice", []uuid.UUID{mountain}, at); err != nil {
		t.Fatalf("Crash for a Mountain: %v", err)
	}
	if g.Battlefield.Contains(mountain) || !g.Battlefield.Contains(island) {
		t.Fatal("the cost took the wrong land")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived Crash")
	}
}

func TestMmqB1ThunderclapDealsThreeToACreatureForAMountain(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	clap := handCardFull(me, "Thunderclap", "Instant", "{2}{R}", mmqB1ThunderclapOracle, []string{"R"})
	big := mmqB1Push(g, them.ID, "Ogre", "Creature — Ogre", "", 3, 3, "R")
	mountain := seedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	if err := mmqB1Cast(g, me.ID, clap, "sacrifice", []uuid.UUID{mountain}, cardRefs(big)); err != nil {
		t.Fatalf("Thunderclap for a Mountain: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 3/3 survived 3 damage")
	}
	if !me.Graveyard.Contains(clap) {
		t.Error("Thunderclap did not go to the graveyard")
	}
}

func TestMmqB1PulverizeSacrificesTwoMountainsToDestroyAllArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	pulverize := handCardFull(me, "Pulverize", "Sorcery", "{4}{R}{R}", mmqB1PulverizeOracle, []string{"R"})
	mine := mmqB1Push(g, me.ID, "My Rock", "Artifact", "", 0, 0)
	theirs := mmqB1Push(g, them.ID, "Their Rock", "Artifact", "", 0, 0)
	bear := mmqB1Push(g, them.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	m1 := seedLand(g, me.ID, "Mountain A", "Basic Land — Mountain", "")

	if err := mmqB1Cast(g, me.ID, pulverize, "sacrifice", []uuid.UUID{m1}, nil); err == nil {
		t.Fatal("one Mountain paid \"sacrifice two Mountains\"")
	}
	m2 := seedLand(g, me.ID, "Mountain B", "Basic Land — Mountain", "")
	if err := mmqB1Cast(g, me.ID, pulverize, "sacrifice", []uuid.UUID{m1, m2}, nil); err != nil {
		t.Fatalf("Pulverize for two Mountains: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an artifact survived Pulverize")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("Pulverize destroyed a creature")
	}
}

// --- Rouse: 2 life instead, with a Swamp ------------------------------------

func TestMmqB1RousePaysTwoLifeWithASwamp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	rouse := handCardFull(me, "Rouse", "Instant", "{1}{B}", mmqB1RouseOracle, []string{"B"})
	bear := mmqB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	at := cardRefs(bear)

	if err := mmqB1Cast(g, me.ID, rouse, "pay_life", nil, at); err == nil {
		t.Fatal("Rouse paid 2 life with no Swamp")
	}
	seedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	before := me.Life
	if err := mmqB1Cast(g, me.ID, rouse, "pay_life", nil, at); err != nil {
		t.Fatalf("Rouse for 2 life with a Swamp: %v", err)
	}
	if me.Life != before-2 {
		t.Errorf("life = %d, want %d", me.Life, before-2)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 4/2", p, tough)
	}
}

// --- Cave-In and Vine Dryad: exile a card of their colour instead ----------

func TestMmqB1CaveInPitchesARedCardAndHitsEveryCreatureAndPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	cave := handCardFull(me, "Cave-In", "Sorcery", "{3}{R}{R}", mmqB1CaveInOracle, []string{"R"})
	blue := handCardFull(me, "Blue Card", "Instant", "{U}", "", []string{"U"})
	red := handCardFull(me, "Red Card", "Instant", "{R}", "", []string{"R"})
	mine := mmqB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2, "G")
	theirs := mmqB1Push(g, them.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")
	wall := mmqB1Push(g, them.ID, "Wall", "Creature — Wall", "", 0, 4, "W")

	if err := mmqB1Cast(g, me.ID, cave, "pitch", []uuid.UUID{blue}, nil); err == nil {
		t.Fatal("a blue card paid \"exile a red card\"")
	}
	myLife, theirLife := me.Life, them.Life
	if err := mmqB1Cast(g, me.ID, cave, "pitch", []uuid.UUID{red}, nil); err != nil {
		t.Fatalf("Cave-In for a red card: %v", err)
	}
	if me.Hand.Contains(red) || !me.Hand.Contains(blue) {
		t.Fatal("the cost took the wrong card")
	}
	passPriorityAroundTable(t, g)
	if me.Life != myLife-2 || them.Life != theirLife-2 {
		t.Errorf("life lost: mine %d, theirs %d; want 2 and 2", myLife-me.Life, theirLife-them.Life)
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a 2/2 survived Cave-In")
	}
	if !g.Battlefield.Contains(wall) {
		t.Error("a 0/4 died to 2 damage")
	}
}

func TestMmqB1VineDryadPitchesAGreenCardAndHasFlashAndForestwalk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	dryad := handCardFull(me, "Vine Dryad", "Creature — Dryad", "{3}{G}", mmqB1VineDryadOracle, []string{"G"})
	red := handCardFull(me, "Red Card", "Instant", "{R}", "", []string{"R"})
	green := handCardFull(me, "Green Card", "Instant", "{G}", "", []string{"G"})

	if err := mmqB1Cast(g, me.ID, dryad, "pitch", []uuid.UUID{red}, nil); err == nil {
		t.Fatal("a red card paid \"exile a green card\"")
	}
	if err := mmqB1Cast(g, me.ID, dryad, "pitch", []uuid.UUID{green}, nil); err != nil {
		t.Fatalf("Vine Dryad for a green card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dryad) {
		t.Fatal("Vine Dryad did not enter")
	}
	for _, kw := range []string{"flash", "forestwalk"} {
		if !hasEffectiveKeyword(t, g, dryad, kw) {
			t.Errorf("Vine Dryad does not have %s", kw)
		}
	}
}

// --- Tidal Bore: return an Island instead -----------------------------------

func TestMmqB1TidalBoreReturnsAnIslandAndUntapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	bore := handCardFull(me, "Tidal Bore", "Instant", "{1}{U}", mmqB1TidalBoreOracle, []string{"U"})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID,
	})
	mountain := seedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	if err := mmqB1Cast(g, me.ID, bore, "return", []uuid.UUID{mountain}, cardRefs(bear)); err == nil {
		t.Fatal("a Mountain paid \"return an Island\"")
	}
	island := seedLand(g, me.ID, "Island", "Basic Land — Island", "")
	if err := mmqB1Cast(g, me.ID, bore, "return", []uuid.UUID{island}, cardRefs(bear)); err != nil {
		t.Fatalf("Tidal Bore for an Island: %v", err)
	}
	if !me.Hand.Contains(island) {
		t.Fatal("the Island did not return to hand as the cost")
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 1) // "Untap it"
	if b16Tapped(t, g, bear) {
		t.Error("the bear is still tapped")
	}
}

// --- the Legates: free if an opponent controls one land and you another ----

var mmqB1Legates = []struct {
	name, oracle, typeLine, theirs, yours, keyword string
}{
	{"Saprazzan Legate", mmqB1SaprazzanOracle, "Creature — Merfolk Soldier", "Mountain", "Island", "flying"},
	{"Kyren Legate", mmqB1KyrenOracle, "Creature — Goblin", "Plains", "Mountain", "haste"},
	{"Cho-Arrim Legate", mmqB1ChoArrimOracle, "Creature — Human Soldier", "Swamp", "Plains", "protection from black"},
	{"Deepwood Legate", mmqB1DeepwoodOracle, "Creature — Shade", "Forest", "Swamp", ""},
}

func TestMmqB1LegatesAreFreeOnlyWhenAnOpponentHasTheirLandAndYouHaveYours(t *testing.T) {
	for _, tc := range mmqB1Legates {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := g.Seats[0], g.Seats[1]
			toMainForCost(t, g)
			legate := handCardFull(me, tc.name, tc.typeLine, "{3}", tc.oracle, nil)
			seedLand(g, me.ID, tc.yours, "Basic Land — "+tc.yours, "")
			if err := mmqB1Cast(g, me.ID, legate, "free", nil, nil); err == nil {
				t.Fatalf("cast free with no opponent's %s", tc.theirs)
			}
			// Your own land of their type is not an opponent's.
			seedLand(g, me.ID, tc.theirs, "Basic Land — "+tc.theirs, "")
			if err := mmqB1Cast(g, me.ID, legate, "free", nil, nil); err == nil {
				t.Fatalf("cast free off your own %s", tc.theirs)
			}
			seedLand(g, them.ID, tc.theirs, "Basic Land — "+tc.theirs, "")
			if err := mmqB1Cast(g, me.ID, legate, "free", nil, nil); err != nil {
				t.Fatalf("free cast with an opponent's %s and your %s: %v", tc.theirs, tc.yours, err)
			}
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(legate) {
				t.Fatal("the Legate did not enter")
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, legate, tc.keyword) {
				t.Errorf("%s does not have %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestMmqB1LegatesNeedYourLandToo(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	legate := handCardFull(me, "Kyren Legate", "Creature — Goblin", "{1}{R}", mmqB1KyrenOracle, []string{"R"})
	seedLand(g, them.ID, "Plains", "Basic Land — Plains", "")
	if err := mmqB1Cast(g, me.ID, legate, "free", nil, nil); err == nil {
		t.Error("cast free with an opponent's Plains but no Mountain of your own")
	}
}

func TestMmqB1DeepwoodLegatePumpsForB(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mmqB1Push(g, me.ID, "Deepwood Legate", "Creature — Shade", mmqB1DeepwoodOracle, 1, 1, "B")
	for i := 1; i <= 2; i++ {
		dcMana(t, g, me.ID, "{B}")
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		if p, tough := ptOf(t, g, id); p != 1+i || tough != 1+i {
			t.Fatalf("the Legate is %d/%d after %d activations, want %d/%d", p, tough, i, 1+i, 1+i)
		}
	}
}

// --- Kris Mage: {R}, {T}, discard a card ------------------------------------

func TestMmqB1KrisMageDiscardsACardToDealOneDamage(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mage := mmqB1Push(g, me.ID, "Kris Mage", "Creature — Human Spellshaper", mmqB1KrisMageOracle, 1, 1, "R")
	at := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}

	dcMana(t, g, me.ID, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true, Targets: at}); err == nil {
		t.Fatal("Kris Mage was activated without discarding a card")
	}
	fodder := handCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{
		Strict: true, Targets: at, DiscardIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("Kris Mage: %v", err)
	}
	if me.Hand.Contains(fodder) || !me.Graveyard.Contains(fodder) {
		t.Fatal("the card was not discarded as the cost")
	}
	if !b16Tapped(t, g, mage) {
		t.Error("Kris Mage did not tap")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("the opponent lost %d, want 1", before-opp.Life)
	}
}

// --- Ramosian Sergeant and Lieutenant: the rebel search ---------------------

func TestMmqB1RamosianRecruitersFindARebelWithinTheirManaValue(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, pay string
		max               int
	}{
		{"Ramosian Sergeant", mmqB1SergeantOracle, "{C}{C}{C}", 2},
		{"Ramosian Lieutenant", mmqB1LieutenantOracle, "{C}{C}{C}{C}", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			seedSearchLibrary(me,
				game.Card{Name: "Small Rebel", TypeLine: "Creature — Human Rebel", ManaCost: "{1}{W}"},
				game.Card{Name: "Edge Rebel", TypeLine: "Creature — Human Rebel", ManaCost: "{2}{W}"},
				game.Card{Name: "Rebel Sorcery", TypeLine: "Kindred Sorcery — Rebel", ManaCost: "{W}"},
				game.Card{Name: "Small Soldier", TypeLine: "Creature — Human Soldier", ManaCost: "{W}"},
			)
			recruiter := mmqB1Push(g, me.ID, tc.name, "Creature — Human Rebel", tc.oracle, 1, 1, "W")
			dcMana(t, g, me.ID, tc.pay)
			if err := g.ActivateCatalogAbility(me.ID, recruiter, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			passPriorityAroundTable(t, g)
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("the search names a quality, so it asks")
			}
			if searchOptionNamed(g, c, "Rebel Sorcery") != uuid.Nil {
				t.Error("a Rebel sorcery is not a permanent card")
			}
			if searchOptionNamed(g, c, "Small Soldier") != uuid.Nil {
				t.Error("a non-Rebel was offered")
			}
			edge := searchOptionNamed(g, c, "Edge Rebel")
			if tc.max >= 3 && edge == uuid.Nil {
				t.Error("a mana value 3 Rebel was not offered at 3 or less")
			}
			if tc.max < 3 && edge != uuid.Nil {
				t.Errorf("a mana value 3 Rebel was offered at %d or less", tc.max)
			}
			answerSearchNamed(t, g, me.ID, "Small Rebel")
			if findBattlefieldByName(g, "Small Rebel") == uuid.Nil {
				t.Error("the Rebel did not enter the battlefield")
			}
		})
	}
}
