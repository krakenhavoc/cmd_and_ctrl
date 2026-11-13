package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pls_b1_cards_test.go — Planeshift, batch 1 (#5615). Each test plays the
// real catalog entry and asserts what its text says.

const (
	plsB1SilverDrakeOracle   = "ec7eb7d8-92b2-4442-856e-a9790301c3b3"
	plsB1HornedKavuOracle    = "14817050-9f2d-453f-b36b-e6429fa29165"
	plsB1ShivanWurmOracle    = "fc988ce3-0cad-41a5-8f13-656caca732df"
	plsB1PaladinOracle       = "15e0250e-a23b-499a-a8d2-24fc13084c28"
	plsB1PantherOracle       = "d99e8f94-4a8a-42a1-b5d0-2b08cacb8676"
	plsB1LavaZombieOracle    = "2232c38e-49eb-4823-a325-2342a952db72"
	plsB1HarpyOracle         = "81c40bd8-b989-41e8-9527-eae8fb87311d"
	plsB1SparkcasterOracle   = "84c912e8-6a92-47f7-88b6-e777bd0bb424"
	plsB1FlametongueOracle   = "b215196a-3716-451f-bdbd-d16c05ace5b4"
	plsB1ErtaiOracle         = "36934bd0-b275-4222-926c-b5a74cf0967d"
	plsB1CrosisCharmOracle   = "e59d70a2-40ac-45b0-995d-65b9caa290c8"
	plsB1DarigaazCharmOracle = "7ee01801-5e42-4916-a359-64e9975090a7"
	plsB1DromarCharmOracle   = "e9950393-8458-4132-9dc7-01246282a41a"
	plsB1TrevaCharmOracle    = "e9795590-0918-49f9-b37c-1278c8616ab1"
	plsB1ShockOracle         = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
)

func TestPlsB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		plsB1SilverDrakeOracle:   "Silver Drake",
		plsB1HornedKavuOracle:    "Horned Kavu",
		plsB1ShivanWurmOracle:    "Shivan Wurm",
		plsB1PaladinOracle:       "Steel Leaf Paladin",
		plsB1PantherOracle:       "Fleetfoot Panther",
		plsB1LavaZombieOracle:    "Lava Zombie",
		plsB1HarpyOracle:         "Cavern Harpy",
		plsB1SparkcasterOracle:   "Sparkcaster",
		plsB1FlametongueOracle:   "Flametongue Kavu",
		plsB1ErtaiOracle:         "Ertai, the Corrupted",
		plsB1CrosisCharmOracle:   "Crosis's Charm",
		plsB1DarigaazCharmOracle: "Darigaaz's Charm",
		plsB1DromarCharmOracle:   "Dromar's Charm",
		plsB1TrevaCharmOracle:    "Treva's Charm",
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

// plsB1Colored seeds a creature with a mana cost (its colours) on the
// battlefield, untapped and not summoning sick.
func plsB1Colored(g *game.Game, owner uuid.UUID, name, manaCost string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// plsB1Cast casts a catalog card with a mana cost from the active seat's
// hand, without resolving it.
func plsB1Cast(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// plsB1Settle resolves the stack, ordering triggers as offered and
// answering every target pick with pickTarget and every own-permanent
// pick with pickOwn. It returns the own-permanent prompts it answered.
func plsB1Settle(t *testing.T, g *game.Game,
	pickTarget func(c *game.PendingChoice) game.TargetRef,
	pickOwn func(c *game.PendingChoice) []uuid.UUID,
) []game.PendingChoice {
	t.Helper()
	var own []game.PendingChoice
	for i := 0; i < 32; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c == nil {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerOrder:
				if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
			case game.PendingChoicePickTarget:
				if pickTarget == nil {
					t.Fatal("an unexpected target pick")
				}
				if err := g.ResolvePickTarget(c.ID, c.Chooser, pickTarget(c)); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
			case game.PendingChoiceOwnPermanents:
				if pickOwn == nil {
					t.Fatal("an unexpected own-permanent pick")
				}
				own = append(own, *c)
				if err := g.ResolveOwnPermanents(c.ID, c.Chooser, pickOwn(c)); err != nil {
					t.Fatalf("ResolveOwnPermanents: %v", err)
				}
			default:
				continue
			}
			answered = true
			break
		}
		if answered {
			continue
		}
		if stackFullyEmpty(g) {
			return own
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("the table did not settle")
	return own
}

func plsB1Pick(id uuid.UUID) func(*game.PendingChoice) []uuid.UUID {
	return func(*game.PendingChoice) []uuid.UUID { return []uuid.UUID{id} }
}

// --- the gating creatures ----------------------------------------------------

func TestPlsB1GatingCreaturesReturnACreatureOfTheirColoursYouControl(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, cost string
		power, toughness             int
		in, out                      string // a cost of each colour the clause does and does not name
		keywords                     []string
	}{
		{"Silver Drake", "Creature — Drake", plsB1SilverDrakeOracle, "{1}{W}{U}", 3, 3, "{U}", "{G}", []string{"flying"}},
		{"Horned Kavu", "Creature — Kavu", plsB1HornedKavuOracle, "{R}{G}", 3, 4, "{G}", "{W}", nil},
		{"Shivan Wurm", "Creature — Wurm", plsB1ShivanWurmOracle, "{3}{R}{G}", 7, 7, "{R}", "{U}", []string{"trample"}},
		{"Steel Leaf Paladin", "Creature — Elf Knight", plsB1PaladinOracle, "{4}{G}{W}", 4, 4, "{W}", "{B}", []string{"first strike"}},
		{"Fleetfoot Panther", "Creature — Cat", plsB1PantherOracle, "{1}{G}{W}", 3, 4, "{G}", "{R}", []string{"flash"}},
		{"Lava Zombie", "Creature — Zombie", plsB1LavaZombieOracle, "{1}{B}{R}", 4, 3, "{B}", "{G}", nil},
		{"Cavern Harpy", "Creature — Harpy Beast", plsB1HarpyOracle, "{U}{B}", 2, 1, "{B}", "{W}", []string{"flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			in := plsB1Colored(g, me.ID, "In", tc.in)
			out := plsB1Colored(g, me.ID, "Out", tc.out)
			theirs := plsB1Colored(g, opp.ID, "Theirs", tc.in)
			self := plsB1Cast(t, g, tc.name, tc.typeLine, tc.oracle, tc.cost, tc.power, tc.toughness)
			asked := plsB1Settle(t, g, nil, plsB1Pick(in))
			if len(asked) != 1 {
				t.Fatalf("%d return picks, want 1", len(asked))
			}
			p := asked[0]
			if !hasID(p.ChooseCards, in) || !hasID(p.ChooseCards, self) {
				t.Errorf("offered %v: want the %s creature and the card itself", p.ChooseCards, tc.in)
			}
			if hasID(p.ChooseCards, out) || hasID(p.ChooseCards, theirs) {
				t.Error("a creature of another colour, or an opponent's, was offered")
			}
			if p.ChooseMin != 1 {
				t.Error("the return is mandatory")
			}
			if !me.Hand.Contains(in) || g.Battlefield.Contains(in) {
				t.Error("the chosen creature did not return to its owner's hand")
			}
			if !g.Battlefield.Contains(self) {
				t.Errorf("%s left the battlefield", tc.name)
			}
			for _, kw := range tc.keywords {
				if kw == "flash" {
					continue
				}
				if !hasKeywordOnBattlefield(t, g, self, kw) {
					t.Errorf("%s does not have %s", tc.name, kw)
				}
			}
		})
	}
}

func TestPlsB1GatingReturnsItselfWhenItIsTheOnlyCandidate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	plsB1Colored(g, me.ID, "Black", "{B}")
	drake := plsB1Cast(t, g, "Silver Drake", "Creature — Drake", plsB1SilverDrakeOracle, "{1}{W}{U}", 3, 3)
	asked := plsB1Settle(t, g, nil, plsB1Pick(drake))
	if len(asked) != 1 || len(asked[0].ChooseCards) != 1 || asked[0].ChooseCards[0] != drake {
		t.Fatalf("asked %+v, want one pick offering only the Drake", asked)
	}
	if !me.Hand.Contains(drake) {
		t.Error("the Drake did not return itself")
	}
}

func TestPlsB1FleetfootPantherHasFlash(t *testing.T) {
	spec, _ := Lookup(plsB1PantherOracle)
	found := false
	for _, kw := range spec.PrintedKeywords {
		found = found || kw == "flash"
	}
	if !found {
		t.Error("Fleetfoot Panther does not print flash")
	}
}

// --- Lava Zombie ---------------------------------------------------------------------

func TestPlsB1LavaZombiePumpsItsPowerForTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	zombie := invB1Push(g, me.ID, "Lava Zombie", "Creature — Zombie", plsB1LavaZombieOracle, 4, 3)
	for i := 0; i < 2; i++ {
		dcMana(t, g, me.ID, "{C}{C}")
		if err := invB1Activate(g, me.ID, zombie, 0); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if p, tough := ptOf(t, g, zombie); p != 6 || tough != 3 {
		t.Errorf("Lava Zombie is %d/%d after two activations, want 6/3", p, tough)
	}
}

// --- Cavern Harpy --------------------------------------------------------------------

func TestPlsB1CavernHarpyPaysOneLifeToReturnToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	harpy := invB1Push(g, me.ID, "Cavern Harpy", "Creature — Harpy Beast", plsB1HarpyOracle, 2, 1)
	before := me.Life
	if err := invB1Activate(g, me.ID, harpy, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d after paying, want %d", me.Life, before-1)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(harpy) || !me.Hand.Contains(harpy) {
		t.Error("Cavern Harpy did not return to its owner's hand")
	}
}

// --- Sparkcaster ---------------------------------------------------------------------

func TestPlsB1SparkcasterReturnsACreatureAndPingsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	kavu := plsB1Cast(t, g, "Sparkcaster", "Creature — Kavu", plsB1SparkcasterOracle, "{2}{R}{G}", 5, 3)
	plsB1Settle(t, g,
		func(c *game.PendingChoice) game.TargetRef {
			if !hasID(c.PickTargetPlayers, opp.ID) || !hasID(c.PickTargetPlayers, me.ID) {
				t.Error("any player is a legal target")
			}
			if hasID(c.PickTargetCards, kavu) {
				t.Error("a creature was offered as the ping's target")
			}
			return playerTarget(opp.ID)
		},
		plsB1Pick(kavu))
	if got := before - opp.Life; got != 1 {
		t.Errorf("the target lost %d life, want 1", got)
	}
	if !me.Hand.Contains(kavu) {
		t.Error("Sparkcaster did not return itself")
	}
}

// --- Flametongue Kavu ----------------------------------------------------------------

func TestPlsB1FlametongueKavuDealsFourToTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ogre := invB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 4, 4)
	kavu := plsB1Cast(t, g, "Flametongue Kavu", "Creature — Kavu", plsB1FlametongueOracle, "{3}{R}", 4, 2)
	plsB1Settle(t, g, func(c *game.PendingChoice) game.TargetRef {
		if len(c.PickTargetPlayers) != 0 {
			t.Error("a player was offered as the target")
		}
		return game.TargetRef{Kind: game.TargetCard, ID: ogre}
	}, nil)
	if g.Battlefield.Contains(ogre) {
		t.Error("the 4/4 survived 4 damage")
	}
	if !g.Battlefield.Contains(kavu) {
		t.Error("Flametongue Kavu left the battlefield")
	}
}

func TestPlsB1FlametongueKavuMustTargetItselfAlone(t *testing.T) {
	g := newCatalogGame(t)
	kavu := plsB1Cast(t, g, "Flametongue Kavu", "Creature — Kavu", plsB1FlametongueOracle, "{3}{R}", 4, 2)
	plsB1Settle(t, g, func(c *game.PendingChoice) game.TargetRef {
		if len(c.PickTargetCards) != 1 || c.PickTargetCards[0] != kavu {
			t.Errorf("offered %v, want only the Kavu", c.PickTargetCards)
		}
		return game.TargetRef{Kind: game.TargetCard, ID: kavu}
	}, nil)
	if g.Battlefield.Contains(kavu) {
		t.Error("the Kavu survived its own 4 damage")
	}
}

// --- Ertai, the Corrupted ------------------------------------------------------------

func TestPlsB1ErtaiSacrificesACreatureOrEnchantmentToCounterASpell(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	ertai := invB1Push(g, me.ID, "Ertai, the Corrupted", "Legendary Creature — Phyrexian Human Wizard", plsB1ErtaiOracle, 3, 4)
	land := invB1Push(g, me.ID, "Plains", "Basic Land — Plains", "", 0, 0)
	ench := invB1Push(g, me.ID, "Fodder", "Enchantment", "", 0, 0)
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant", OracleID: plsB1ShockOracle,
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast Shock: %v", err)
	}
	dcMana(t, g, me.ID, "{U}")
	activate := func(sac uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, ertai, 0, game.ActivateAbilityParams{
			Strict: true, Targets: cardTarget(shock), SacrificeIDs: []uuid.UUID{sac},
		})
	}
	if err := activate(land); err == nil {
		t.Fatal("a land was accepted as the sacrifice")
	}
	if err := activate(ench); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment was not sacrificed as the cost")
	}
	before := me.Life
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(shock) {
		t.Error("the countered Shock is not in its owner's graveyard")
	}
	if me.Life != before {
		t.Errorf("the countered Shock still dealt damage: life %d, want %d", me.Life, before)
	}
	if !findCardAnywhere(t, g, ertai).Tapped {
		t.Error("Ertai did not tap")
	}
}

// --- Crosis's Charm ------------------------------------------------------------------

func TestPlsB1CrosissCharmBouncesToTheOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen", TypeLine: "Artifact", Owner: opp.ID, Controller: me.ID,
	})
	castModal(t, g, "Crosis's Charm", "Instant", plsB1CrosisCharmOracle, []int{0}, cardTarget(stolen))
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(stolen) {
		t.Error("the permanent did not return to its owner's hand")
	}
}

func TestPlsB1CrosissCharmDestroysANonblackCreatureThroughRegeneration(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	black := plsB1Colored(g, opp.ID, "Black", "{1}{B}")
	green := plsB1Colored(g, opp.ID, "Green", "{1}{G}")
	g.WithWriteLock(func() {
		if c := battlefieldCardFor(g, green); c != nil {
			c.RegenerationShields = 1
		}
	})
	if _, err := castModalErr(t, g, "Crosis's Charm", "Instant", plsB1CrosisCharmOracle, []int{1}, cardTarget(black)); err == nil {
		t.Fatal("a black creature was accepted as the target")
	}
	castModal(t, g, "Crosis's Charm", "Instant", plsB1CrosisCharmOracle, []int{1}, cardTarget(green))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("a regeneration shield saved the creature")
	}
}

func TestPlsB1CrosissCharmDestroysTargetArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	relic := invB1Push(g, opp.ID, "Relic", "Artifact", "", 0, 0)
	bear := seedBear(g, opp.ID)
	if _, err := castModalErr(t, g, "Crosis's Charm", "Instant", plsB1CrosisCharmOracle, []int{2}, cardTarget(bear)); err == nil {
		t.Fatal("a creature was accepted as the artifact target")
	}
	castModal(t, g, "Crosis's Charm", "Instant", plsB1CrosisCharmOracle, []int{2}, cardTarget(relic))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived")
	}
}

// --- Darigaaz's Charm ----------------------------------------------------------------

func TestPlsB1DarigaazsCharmReturnsACreatureCardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	relic := pushGraveyardCardTyped(me, "Relic", "Artifact")
	if _, err := castModalErr(t, g, "Darigaaz's Charm", "Instant", plsB1DarigaazCharmOracle, []int{0}, cardTarget(relic)); err == nil {
		t.Fatal("a noncreature card was accepted as the target")
	}
	castModal(t, g, "Darigaaz's Charm", "Instant", plsB1DarigaazCharmOracle, []int{0}, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature card did not come back to hand")
	}
}

func TestPlsB1DarigaazsCharmDealsThreeToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	castModal(t, g, "Darigaaz's Charm", "Instant", plsB1DarigaazCharmOracle, []int{1}, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 3 {
		t.Errorf("the target lost %d life, want 3", got)
	}
}

func TestPlsB1DarigaazsCharmPumpsThreeThree(t *testing.T) {
	g := newCatalogGame(t)
	bear := seedBear(g, g.Seats[0].ID)
	castModal(t, g, "Darigaaz's Charm", "Instant", plsB1DarigaazCharmOracle, []int{2}, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the target is %d/%d, want 5/5", p, tough)
	}
}

// --- Dromar's Charm ------------------------------------------------------------------

func TestPlsB1DromarsCharmGainsFiveLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	castModal(t, g, "Dromar's Charm", "Instant", plsB1DromarCharmOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 5 {
		t.Errorf("gained %d life, want 5", got)
	}
}

func TestPlsB1DromarsCharmCountersTargetSpell(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant", OracleID: plsB1ShockOracle,
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast Shock: %v", err)
	}
	before := me.Life
	castModal(t, g, "Dromar's Charm", "Instant", plsB1DromarCharmOracle, []int{1}, cardTarget(shock))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(shock) || me.Life != before {
		t.Errorf("Shock was not countered: in graveyard %v, life %d, want %d", opp.Graveyard.Contains(shock), me.Life, before)
	}
}

func TestPlsB1DromarsCharmShrinksTwoTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ogre := invB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	castModal(t, g, "Dromar's Charm", "Instant", plsB1DromarCharmOracle, []int{2}, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
}

// --- Treva's Charm -------------------------------------------------------------------

func TestPlsB1TrevasCharmDestroysTargetEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ench := invB1Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	castModal(t, g, "Treva's Charm", "Instant", plsB1TrevaCharmOracle, []int{0}, cardTarget(ench))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
}

func TestPlsB1TrevasCharmExilesOnlyAnAttackingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	attacker := seedBear(g, me.ID)
	idle := seedBear(g, me.ID)
	declareAttack(t, g, opp.ID, attacker)
	cast := func(target uuid.UUID) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Treva's Charm", TypeLine: "Instant", OracleID: plsB1TrevaCharmOracle,
			Owner: me.ID, Controller: me.ID,
		})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: cardTarget(target)})
	}
	if err := cast(idle); err == nil {
		t.Fatal("a creature that is not attacking was accepted as the target")
	}
	if err := cast(attacker); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(attacker) {
		t.Error("the attacking creature was not exiled")
	}
}

func TestPlsB1TrevasCharmDrawsThenDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drawn := seedLibrary(me, "Top")
	castModal(t, g, "Treva's Charm", "Instant", plsB1TrevaCharmOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(drawn[0]) {
		t.Fatal("the top card was not drawn")
	}
	answerDiscard(t, g, me.ID, drawn[0])
	if !me.Graveyard.Contains(drawn[0]) {
		t.Error("the drawn card could not be the discard")
	}
}
