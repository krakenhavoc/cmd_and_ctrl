package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pcy_b1_cards_test.go — Prophecy, batch 1 (#5617). Each test plays the
// real catalog entry and asserts what its text says.

const (
	pcyB1GriffinOracle     = "8785f9b4-eea2-4b45-9ae4-194b9a715702"
	pcyB1BoaOracle         = "246f7446-036d-469c-9583-5699ebdcec6d"
	pcyB1RazorbackOracle   = "2bdd82e7-cbbf-4f12-bcb7-b276df021351"
	pcyB1RagerOracle       = "54e7a6d4-d405-4627-93ab-8d8bd945fe02"
	pcyB1SearingWindOracle = "b196045d-ece1-46ac-a647-b65f04225c10"
	pcyB1DespoilOracle     = "77fa18fa-5dba-4e4c-9b24-9bfa9b68779f"
	pcyB1PlagueWindOracle  = "18ec721f-c1ac-4581-a61d-2f0b09d6bf92"
	pcyB1VitalizingOracle  = "98b802f3-4785-41af-adfe-e77d0ae4648c"
	pcyB1SergeantOracle    = "1e25c9bd-2a6a-4856-80e6-775847a2c89f"
	pcyB1ZombieOracle      = "c498f0ff-812a-4106-834b-f43a0a7e0bac"
	pcyB1RaptorOracle      = "92cdcafc-ffeb-436d-b2bd-d6ad895d0e08"
	pcyB1HatchlingOracle   = "8109aeff-1ba7-46c2-89ab-450d6ff99c28"
	pcyB1DrakeOracle       = "6568859d-b96d-4015-84da-1a8960752737"
	pcyB1ScryingOracle     = "7a07971a-bfc3-4d36-ad8a-b9d395cbd6e9"
)

func TestPcyB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		pcyB1GriffinOracle:     "Diving Griffin",
		pcyB1BoaOracle:         "Marsh Boa",
		pcyB1RazorbackOracle:   "Pygmy Razorback",
		pcyB1RagerOracle:       "Ridgeline Rager",
		pcyB1SearingWindOracle: "Searing Wind",
		pcyB1DespoilOracle:     "Despoil",
		pcyB1PlagueWindOracle:  "Plague Wind",
		pcyB1VitalizingOracle:  "Vitalizing Wind",
		pcyB1SergeantOracle:    "Whip Sergeant",
		pcyB1ZombieOracle:      "Whipstitched Zombie",
		pcyB1RaptorOracle:      "Pit Raptor",
		pcyB1HatchlingOracle:   "Spiketail Hatchling",
		pcyB1DrakeOracle:       "Spiketail Drake",
		pcyB1ScryingOracle:     "Rhystic Scrying",
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

// pcyB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func pcyB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- printed keywords --------------------------------------------------------------

func TestPcyB1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		keywords               []string
	}{
		{"Diving Griffin", pcyB1GriffinOracle, "Creature — Griffin", []string{"flying", "vigilance"}},
		{"Marsh Boa", pcyB1BoaOracle, "Creature — Snake", []string{"swampwalk"}},
		{"Pygmy Razorback", pcyB1RazorbackOracle, "Creature — Boar", []string{"trample"}},
		{"Pit Raptor", pcyB1RaptorOracle, "Creature — Bird Mercenary", []string{"flying", "first strike"}},
		{"Spiketail Hatchling", pcyB1HatchlingOracle, "Creature — Drake", []string{"flying"}},
		{"Spiketail Drake", pcyB1DrakeOracle, "Creature — Drake", []string{"flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := pcyB1Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s does not have %s", tc.name, kw)
				}
			}
		})
	}
}

func TestPcyB1MarshBoaIsUnblockableOnlyAgainstASwamp(t *testing.T) {
	for _, swamp := range []bool{false, true} {
		g := newCatalogGame(t)
		me, them := seatsForCombat(g)
		boa := pcyB1Push(g, me.ID, "Marsh Boa", "Creature — Snake", pcyB1BoaOracle, 1, 1)
		bear := seedCreature(g, "Grizzly Bears", them.ID)
		if swamp {
			seedLand(g, them.ID, "Swamp", "Basic Land — Swamp", "")
		} else {
			seedLand(g, them.ID, "Forest", "Basic Land — Forest", "")
		}
		attackIntoBlocks(t, g, them.ID, boa)
		got := blockRefusal(t, g, bear, boa)
		switch {
		case swamp && got != game.BlockReasonLandwalk:
			t.Errorf("against a Swamp: reason %q, want landwalk", got)
		case !swamp && got != "":
			t.Errorf("against a Forest: the block was refused (%q)", got)
		}
	}
}

// --- Ridgeline Rager ------------------------------------------------------------------

func TestPcyB1RidgelineRagerPumpsPlusOnePlusZero(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rager := pcyB1Push(g, me.ID, "Ridgeline Rager", "Creature — Beast", pcyB1RagerOracle, 1, 2)
	for i := 1; i <= 2; i++ {
		dcMana(t, g, me.ID, "{R}")
		if err := g.ActivateCatalogAbility(me.ID, rager, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		if p, tough := ptOf(t, g, rager); p != 1+i || tough != 2 {
			t.Fatalf("Ridgeline Rager is %d/%d after %d activations, want %d/2", p, tough, i, 1+i)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, rager, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the pump was activated with an empty mana pool")
	}
}

// --- Searing Wind ---------------------------------------------------------------------

func TestPcyB1SearingWindDealsTenDamage(t *testing.T) {
	g := newCatalogGame(t)
	_, victim := seatsForCombat(g)
	life := victim.Life
	castCatalogSpell(t, g, "Searing Wind", "Instant", pcyB1SearingWindOracle, []game.TargetRef{playerTarget(victim.ID)})
	passPriorityAroundTable(t, g)
	if got := life - victim.Life; got != 10 {
		t.Errorf("the target lost %d life, want 10", got)
	}
}

// --- Despoil --------------------------------------------------------------------------

func TestPcyB1DespoilDestroysALandAndItsControllerLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := seatsForCombat(g)
	land := seedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")
	casterLife, oppLife := caster.Life, opp.Life
	castCatalogSpell(t, g, "Despoil", "Sorcery", pcyB1DespoilOracle, cardTarget(land))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if opp.Life != oppLife-2 || caster.Life != casterLife {
		t.Errorf("life: land's controller %d (want %d), caster %d (want %d)", opp.Life, oppLife-2, caster.Life, casterLife)
	}
}

func TestPcyB1DespoilRefusesANonlandTarget(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	bear := seedCreature(g, "Grizzly Bears", opp.ID)
	if err := castCatalogSpellErr(t, g, "Despoil", "Sorcery", pcyB1DespoilOracle, cardTarget(bear)); err == nil {
		t.Error("a creature was accepted as Despoil's target")
	}
}

// --- Plague Wind ----------------------------------------------------------------------

func TestPcyB1PlagueWindDestroysOnlyCreaturesYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	mine := seedCreature(g, "My Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)
	shielded := seedCreature(g, "Their Shielded Bear", opp.ID)
	g.WithWriteLock(func() {
		if err := (Regenerate{Target: shielded}).Apply(NewContext(g, &game.StackItem{Controller: opp.ID, SourceCardID: uuid.New()})); err != nil {
			t.Fatalf("regenerate: %v", err)
		}
	})
	castCatalogSpell(t, g, "Plague Wind", "Sorcery", pcyB1PlagueWindOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's own creature was destroyed")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("an opponent's creature survived")
	}
	if g.Battlefield.Contains(shielded) {
		t.Error("a regeneration shield saved an opponent's creature")
	}
}

// --- Vitalizing Wind ------------------------------------------------------------------

func TestPcyB1VitalizingWindPumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	mine := seedCreature(g, "My Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)
	castCatalogSpell(t, g, "Vitalizing Wind", "Instant", pcyB1VitalizingOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 9 || tough != 9 {
		t.Errorf("the caster's creature is %d/%d, want 9/9", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

// --- Whip Sergeant --------------------------------------------------------------------

func TestPcyB1WhipSergeantGivesHaste(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sergeant := pcyB1Push(g, me.ID, "Whip Sergeant", "Creature — Human Soldier", pcyB1SergeantOracle, 2, 1)
	bear := seedCreature(g, "Grizzly Bears", me.ID)
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Fatal("the bear has haste before the ability")
	}
	dcMana(t, g, me.ID, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, sergeant, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the target did not gain haste")
	}
}

// --- the upkeep payments --------------------------------------------------------------

func TestPcyB1UpkeepIsPaidOrTheCreatureIsSacrificed(t *testing.T) {
	for _, tc := range []struct {
		// cost is mana that funds the printed upkeep: {B}{B}{B}{B} pays {2}{B}{B}.
		name, oracle, typeLine, cost string
	}{
		{"Whipstitched Zombie", pcyB1ZombieOracle, "Creature — Zombie", "{B}"},
		{"Pit Raptor", pcyB1RaptorOracle, "Creature — Bird Mercenary", "{B}{B}{B}{B}"},
	} {
		for _, pay := range []bool{true, false} {
			label := tc.name + " declined"
			if pay {
				label = tc.name + " paid"
			}
			t.Run(label, func(t *testing.T) {
				g := newCatalogGame(t)
				me := g.Seats[0]
				id := pcyB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
				walkToUpkeepOfSeat0(t, g)
				passPriorityAroundTable(t, g)
				if pay {
					dcMana(t, g, me.ID, tc.cost)
				}
				answerPayUnless(t, g, me.ID, pay)
				if g.Battlefield.Contains(id) != pay {
					t.Fatalf("paid=%v: %s on the battlefield = %v", pay, tc.name, g.Battlefield.Contains(id))
				}
				if pay && me.ManaPool.Total() != 0 {
					t.Errorf("%d mana left in the pool after paying %s", me.ManaPool.Total(), tc.cost)
				}
				if !pay && !me.Graveyard.Contains(id) {
					t.Errorf("%s is not in its owner's graveyard", tc.name)
				}
			})
		}
	}
}

// --- the Spiketail Drakes -------------------------------------------------------------

func TestPcyB1SpiketailsCounterUnlessTheTaxIsPaid(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		tax          int
	}{
		{"Spiketail Hatchling", pcyB1HatchlingOracle, 1},
		{"Spiketail Drake", pcyB1DrakeOracle, 3},
	} {
		for _, pay := range []bool{false, true} {
			label := tc.name + " declined"
			if pay {
				label = tc.name + " paid"
			}
			t.Run(label, func(t *testing.T) {
				g := newCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				drake := pcyB1Push(g, me.ID, tc.name, "Creature — Drake", tc.oracle, 1, 1)
				spell := castCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
				if err := g.ActivateCatalogAbility(me.ID, drake, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(spell)}); err != nil {
					t.Fatalf("activate: %v", err)
				}
				if g.Battlefield.Contains(drake) {
					t.Errorf("%s was not sacrificed as part of the cost", tc.name)
				}
				passUntilTaxed(t, g, me.ID)
				if pay {
					fillPool(me, tc.tax)
				}
				answerPayUnless(t, g, me.ID, pay)
				if g.Stack.Contains(spell) != pay {
					t.Errorf("paid=%v: the spell is still on the stack = %v", pay, g.Stack.Contains(spell))
				}
			})
		}
	}
}

// --- Rhystic Scrying ------------------------------------------------------------------

func TestPcyB1RhysticScryingDiscardsThreeOnlyWhenAPlayerPays(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		castCatalogSpell(t, g, "Rhystic Scrying", "Sorcery", pcyB1ScryingOracle, nil)
		hand := me.Hand.Size()
		appMana(opp, 2)
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != hand+3 {
			t.Fatalf("paid=%v: hand %d → %d while the offer is open, want three cards drawn", paid, hand, got)
		}
		var payers []uuid.UUID
		if paid {
			payers = append(payers, opp.ID)
		}
		asked := appAnswerOffer(t, g, payers...)
		if len(asked) == 0 || asked[0] != me.ID {
			t.Errorf("paid=%v: asked %v, want the active player first", paid, asked)
		}
		owed := discardOwed(g, me.ID)
		if paid {
			if owed != 3 {
				t.Errorf("paid: the caster owes %d discards, want 3", owed)
			}
			if len(opp.ManaPool) != 0 {
				t.Errorf("paid: the payer kept %d mana", len(opp.ManaPool))
			}
			continue
		}
		if owed != 0 {
			t.Errorf("unpaid: the caster owes %d discards, want none", owed)
		}
		if got := me.Hand.Size(); got != hand+3 {
			t.Errorf("unpaid: hand %d → %d, want the three cards kept", hand, got)
		}
	}
}
