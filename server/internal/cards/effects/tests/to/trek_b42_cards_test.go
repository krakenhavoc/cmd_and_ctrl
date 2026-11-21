package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b42_cards_test.go — Star Trek Commander, batch 42 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB42KVortOracle      = "a52d7ab2-a750-4c3a-9f0c-999121aa81af"
	trekB42PlanetOracle     = "4c2c8cc6-bb90-468a-867b-db56f4eb7e59"
	trekB42DiamondOracle    = "44837c9b-971a-44b6-8e55-301cebc3a361"
	trekB42PartitionOracle  = "c52e4704-6edb-477f-b80c-ea8146d4385f"
	trekB42BotanyOracle     = "8284c6ae-d86e-4f72-81da-7d3f9861fd12"
	trekB42CommOracle       = "d5560aaa-7459-498f-b3c5-e096d8c75b51"
	trekB42EnterpriseOracle = "9b627fa4-af4c-482b-b1cf-2ad15938602c"
	trekB42KahlessOracle    = "3e301803-49ab-4fce-82ae-945a39c458b5"
	trekB42MigrantOracle    = "c7c529c4-08eb-480b-a699-0d783215419b"
	trekB42ToughOracle      = "87b8fef3-f40e-44a9-afd5-31d3917b1d59"
)

func TestTrekB42EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB42KVortOracle:      "K'Vort-Class Raider",
		trekB42PlanetOracle:     "Planet Killer",
		trekB42DiamondOracle:    "Queen's Diamond",
		trekB42PartitionOracle:  "Reconnaissance Partition",
		trekB42BotanyOracle:     "S.S. Botany Bay",
		trekB42CommOracle:       "Starfleet Communicator",
		trekB42EnterpriseOracle: "Starship Enterprise",
		trekB42KahlessOracle:    "The Sword of Kahless",
		trekB42MigrantOracle:    "U.S.S. Defiant, Mirror Migrant",
		trekB42ToughOracle:      "U.S.S. Defiant, Tough Little Ship",
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

// --- fixtures -------------------------------------------------------------------

// trekB42Ship seeds a permanent ready to act with `charge` charge counters.
func trekB42Ship(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness, charge int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	}
	if charge > 0 {
		c.Counters = map[string]int{game.CounterCharge: charge}
	}
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

// trekB42Graveyard puts a card in p's graveyard.
func trekB42Graveyard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// trekB42Attack declares the attackers at `defender` and locks the attack
// in, so "whenever this attacks" triggers are on the stack.
func trekB42Attack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
}

// --- K'Vort-Class Raider --------------------------------------------------------

func TestTrekB42KVortShieldsYourSideOnYourTurnAndStationsAtFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ship := trekB42Ship(g, me.ID, "K'Vort-Class Raider", "Artifact — Spacecraft", trekB42KVortOracle, 5, 3, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	if !effectstest.HasEffectiveKeyword(t, g, ship, "hexproof") || !effectstest.HasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("on your turn the Spacecraft and your creature have hexproof")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("an opponent's creature got hexproof")
	}
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Fatal("an unstationed Spacecraft is a creature")
	}

	four := effectstest.B12Creature(g, me.ID, "Four", "Creature — Human", 4, 4)
	effectstest.Station(t, g, me, ship, four)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 5 || c.CurrentToughness() != 3 {
		t.Fatalf("at 4+ it is a 5/3 artifact creature; creature=%v %d/%d", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
	for _, kw := range []string{"flying", "menace", "haste"} {
		if !game.HasKeyword(&c, kw) {
			t.Errorf("at 4+ it lacks %s", kw)
		}
	}

	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, ship, "hexproof") || effectstest.HasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("hexproof lasted into an opponent's turn")
	}
}

// --- Planet Killer --------------------------------------------------------------

func TestTrekB42PlanetKillerMustAttackAndGrowsPerDefendingLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pk := effectstest.B12Push(g, me.ID, "Planet Killer", "Artifact Creature — Construct", trekB42PlanetOracle, 9, 9)
	for _, n := range []string{"Forest", "Island", "Swamp"} {
		effectstest.B12Permanent(g, opp.ID, n, "Basic Land — "+n)
	}
	effectstest.B12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	if !effectstest.HasEffectiveKeyword(t, g, pk, "flying") || !effectstest.HasEffectiveKeyword(t, g, pk, "indestructible") {
		t.Error("Planet Killer lacks flying or indestructible")
	}

	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); err == nil {
		t.Fatal("passing with Planet Killer home succeeded, want the attack requirement to refuse it")
	}
	if err := g.DeclareAttacker(pk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40-9 {
		t.Fatalf("opponent life %d, want 31", opp.Life)
	}
	if got := effectstest.CountersOn(g, pk, game.CounterPlusOne); got != 3 {
		t.Errorf("Planet Killer has %d +1/+1 counters, want 3 (the damaged player's lands)", got)
	}
}

// --- Queen's Diamond ------------------------------------------------------------

func TestTrekB42QueensDiamondAddsChargeManaOnlyForYourCommander(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	diamond := trekB42Ship(g, me.ID, "Queen's Diamond", "Legendary Artifact — Spacecraft", trekB42DiamondOracle, 5, 5, 3)
	if err := g.ActivateManaAbility(me.ID, diamond, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tap Queen's Diamond: %v", err)
	}
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool %d, want three mana (one per charge counter)", me.ManaPool.Total())
	}
	cost, _ := game.ParseCost("{2}{R}")

	other := game.Card{InstanceID: uuid.New(), Name: "Not A Commander", TypeLine: "Creature — Klingon", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(other)) {
		t.Error("the mana pays for a spell that is not a commander")
	}
	theirs := game.Card{InstanceID: uuid.New(), Name: "Their Commander", TypeLine: "Legendary Creature — Klingon", ManaCost: "{2}{R}", IsCommander: true, Owner: opp.ID, Controller: me.ID}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(theirs)) {
		t.Error("the mana pays for a commander you don't own")
	}
	dCard, _ := effectstest.BattlefieldCard(g, diamond)
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(dCard)) {
		t.Error("the mana pays for an activated ability")
	}

	wrong := effectstest.HandSpell(me, "Klingon Warrior", "Creature — Klingon", "{2}{R}")
	if err := g.CastSpell(me.ID, wrong, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a non-commander spell was paid with the Diamond's mana")
	}
	cmdr := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cmdr, Name: "Kor", TypeLine: "Legendary Creature — Klingon", ManaCost: "{2}{R}",
		IsCommander: true, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, cmdr, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("your commander with the Diamond's mana: %v", err)
	}
}

func TestTrekB42QueensDiamondStationsToAFlyingVigilantHaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	diamond := trekB42Ship(g, me.ID, "Queen's Diamond", "Legendary Artifact — Spacecraft", trekB42DiamondOracle, 5, 5, 0)
	if err := g.ActivateManaAbility(me.ID, diamond, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil && me.ManaPool.Total() != 0 {
		t.Errorf("with no charge counters it added %d mana", me.ManaPool.Total())
	}
	seven := effectstest.B12Creature(g, me.ID, "Seven", "Creature — Human", 7, 7)
	effectstest.Station(t, g, me, diamond, seven)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, diamond).IsCreature() {
		t.Fatal("a creature at 7 charge counters")
	}
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Human", 1, 1)
	effectstest.Station(t, g, me, diamond, one)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, diamond)
	if !c.IsCreature() || c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Fatalf("at 8+ a 5/5 creature; creature=%v %d/%d", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
	for _, kw := range []string{"flying", "vigilance", "haste"} {
		if !game.HasKeyword(&c, kw) {
			t.Errorf("at 8+ it lacks %s", kw)
		}
	}
}

// --- Reconnaissance Partition ---------------------------------------------------

func TestTrekB42ReconnaissancePartitionSpendsCountersOnALandOrACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rp := effectstest.CastCreatureCard(t, g, "Reconnaissance Partition", "Artifact Creature — Borg", trekB42PartitionOracle, 3, 3)
	if got := effectstest.CountersOn(g, rp, game.CounterCharge); got != 2 {
		t.Fatalf("entered with %d charge counters, want 2", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, rp, 0, game.ActivateAbilityParams{CounterKind: game.CounterCharge, Modes: []int{1}}); err == nil {
		t.Fatal("a summoning-sick creature paid {T}")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rp {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})

	effectstest.StackLibrary(me, "Drawn", "Sorcery", "")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, rp, 0, game.ActivateAbilityParams{CounterKind: game.CounterCharge, Modes: []int{1}}); err != nil {
		t.Fatalf("draw mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
	if got := effectstest.CountersOn(g, rp, game.CounterCharge); got != 1 {
		t.Errorf("%d charge counters left, want 1", got)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rp {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	forest := effectstest.StackLibrary(me, "Forest", "Basic Land — Forest", "")
	if err := g.ActivateCatalogAbility(me.ID, rp, 0, game.ActivateAbilityParams{CounterKind: game.CounterCharge, Modes: []int{0}}); err != nil {
		t.Fatalf("land mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, forest); !ok || !c.Tapped {
		t.Errorf("the basic land is on the battlefield %v, tapped %v; want both", ok, c.Tapped)
	}
	if got := effectstest.CountersOn(g, rp, game.CounterCharge); got != 0 {
		t.Errorf("%d charge counters left, want 0", got)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rp {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, rp, 0, game.ActivateAbilityParams{CounterKind: game.CounterCharge, Modes: []int{1}}); err == nil {
		t.Error("activated with no counter to remove")
	}
}

// --- S.S. Botany Bay ------------------------------------------------------------

func TestTrekB42BotanyBayExilesFromGraveyardsAndReturnsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	ship := trekB42Ship(g, me.ID, "S.S. Botany Bay", "Legendary Artifact — Spacecraft", trekB42BotanyOracle, 3, 5, 8)
	c1 := trekB42Graveyard(opp1, "Their Bear", "Creature — Bear")
	trekB42Graveyard(opp1, "Their Spell", "Instant")
	c2 := trekB42Graveyard(opp2, "Other Bear", "Creature — Bear")
	khan := trekB42Graveyard(me, "Khan", "Creature — Human")
	trekB42Graveyard(me, "My Spell", "Sorcery")

	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, g.Seats[3].ID) != nil {
		t.Error("an opponent with an empty graveyard was asked")
	}
	effectstest.TrekB42Pick(t, g, opp1.ID, c1)
	effectstest.TrekB42Pick(t, g, opp2.ID, c2)
	ch := effectstest.LatestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("expected your may-pick of one card: %+v", ch)
	}
	effectstest.TrekB42Pick(t, g, me.ID, khan)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{c1, c2, khan} {
		if !g.Exile.Contains(id) {
			t.Fatalf("%s was not exiled", id)
		}
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ship {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if c := effectstest.LayeredCard(t, g, ship); !c.IsCreature() || !game.HasKeyword(&c, "flying") {
		t.Fatal("at 8+ Botany Bay is a flying creature")
	}
	effectstest.AttackWith(t, g, opp1.ID, ship)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(khan) && !trekB42OnBattlefieldNamed(g, me.ID, "Khan") {
		t.Error("your creature card exiled with it did not return")
	}
	if !g.Exile.Contains(c1) || !g.Exile.Contains(c2) {
		t.Error("an opponent's exiled card left exile")
	}
}

func trekB42OnBattlefieldNamed(g *game.Game, controller uuid.UUID, name string) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			return true
		}
	}
	return false
}

// --- Starfleet Communicator -----------------------------------------------------

func TestTrekB42CommunicatorCopiesItselfAndDrawsPerOtherCommunicator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastFromHand(t, g, game.Card{Name: "Starfleet Communicator", TypeLine: "Artifact — Communicator Equipment",
		OracleID: trekB42CommOracle, ManaCost: "{3}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	var comms []uuid.UUID
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Starfleet Communicator" && c.Controller == me.ID {
			comms = append(comms, c.InstanceID)
			if effects.IsToken(c) {
				tokens++
			}
		}
	}
	if len(comms) != 2 || tokens != 1 {
		t.Fatalf("%d Communicators (%d tokens), want the card and one token copy", len(comms), tokens)
	}

	a := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Human", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Crew", "Creature — Human", 2, 2)
	effectstest.EquipTo(t, g, me.ID, comms[0], a)
	effectstest.EquipTo(t, g, me.ID, comms[1], b)

	effectstest.StackLibrary(me, "Drawn Land", "Basic Land — Island", "")
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Hand Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	trekB42Attack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d, want X = 1 card drawn", hand, me.Hand.Size())
	}
	ch := effectstest.LatestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("expected up to one land from hand: %+v", ch)
	}
	effectstest.TrekB42Pick(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, land); !ok || !c.Tapped {
		t.Errorf("the land is on the battlefield %v, tapped %v; want both", ok, c.Tapped)
	}
}

func TestTrekB42CommunicatorCountsOnlyOtherEquippedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	comm := effectstest.B12Push(g, me.ID, "Starfleet Communicator", "Artifact — Communicator Equipment", trekB42CommOracle, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Human", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bare", "Creature — Human", 2, 2)
	effectstest.EquipTo(t, g, me.ID, comm, a)
	effectstest.StackLibrary(me, "Drawn", "Sorcery", "")
	hand := me.Hand.Size()
	trekB42Attack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d; X is 0 with no other Communicator-equipped creature", hand, me.Hand.Size())
	}
}

// --- Starship Enterprise --------------------------------------------------------

func TestTrekB42EnterpriseLandfallChargesAndDrawsOnDamageAtSixteen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := trekB42Ship(g, me.ID, "Starship Enterprise", "Legendary Artifact — Spacecraft", trekB42EnterpriseOracle, 4, 6, 12)
	if effectstest.LayeredCard(t, g, ship).IsCreature() {
		t.Fatal("a creature at 12 charge counters")
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ship, game.CounterCharge); got != 16 {
		t.Fatalf("%d charge counters after landfall, want 16", got)
	}
	c := effectstest.LayeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 4 || c.CurrentToughness() != 6 || !game.HasKeyword(&c, "flying") {
		t.Fatalf("at 16+ a 4/6 flier; creature=%v %d/%d flying=%v", c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"))
	}
	for i := 0; i < 5; i++ {
		trekB23Library(me, game.Card{Name: "Card", TypeLine: "Sorcery"})
	}
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, ship)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+4 {
		t.Errorf("hand %d → %d, want four cards drawn (the damage dealt)", hand, me.Hand.Size())
	}
}

// --- The Sword of Kahless -------------------------------------------------------

func TestTrekB42SwordOfKahlessPumpsAndPutsAKlingonFromTheTopFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sword := effectstest.B12Push(g, me.ID, "The Sword of Kahless", "Legendary Artifact — Equipment", trekB42KahlessOracle, 0, 0)
	bearer := effectstest.B12Creature(g, me.ID, "Bearer", "Creature — Human", 2, 2)
	effectstest.EquipTo(t, g, me.ID, sword, bearer)
	if c := effectstest.LayeredCard(t, g, bearer); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Fatalf("equipped creature is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	lib := trekB23Library(me,
		game.Card{Name: "Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1},
		game.Card{Name: "Worf", TypeLine: "Creature — Klingon", Power: 3, Toughness: 3},
		game.Card{Name: "Instant", TypeLine: "Instant"},
		game.Card{Name: "Warrior", TypeLine: "Creature — Human Warrior", Power: 2, Toughness: 2},
		game.Card{Name: "Land", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Sixth Klingon", TypeLine: "Creature — Klingon", Power: 3, Toughness: 3},
	)
	trekB42Attack(t, g, opp.ID, bearer)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("expected a may-pick of one card: %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0]}); err == nil {
		t.Error("a creature that is neither Klingon nor Warrior was accepted")
	}
	effectstest.TrekB42Pick(t, g, me.ID, lib[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(lib[1]) && !trekB42OnBattlefieldNamed(g, me.ID, "Worf") {
		t.Fatal("the Klingon was not put onto the battlefield")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != lib[5] {
		t.Errorf("the library top is %v, want the sixth card (the rest went to the bottom)", top.Name)
	}
}

// --- U.S.S. Defiant, Mirror Migrant ---------------------------------------------

func TestTrekB42MirrorMigrantPhasesOutWithACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ship := effectstest.CastHoldCreature(t, g, "U.S.S. Defiant, Mirror Migrant", "Legendary Artifact — Spacecraft", trekB42MigrantOracle, 5, 5)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("an opponent's creature was accepted")
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.PhasedOut.Contains(ship) || !g.PhasedOut.Contains(mine) {
		t.Errorf("phased out: Defiant %v, creature %v; want both", g.PhasedOut.Contains(ship), g.PhasedOut.Contains(mine))
	}
	if spec, _ := effects.Lookup(trekB42MigrantOracle); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Error("the Defiant does not print flash")
	}
}

// --- U.S.S. Defiant, Tough Little Ship ------------------------------------------

func TestTrekB42ToughLittleShipDestroysAnArtifactAndStationsAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ship := effectstest.CastHoldCreature(t, g, "U.S.S. Defiant, Tough Little Ship", "Legendary Artifact — Spacecraft", trekB42ToughOracle, 7, 4)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("a creature was accepted as an artifact or enchantment")
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}

	five := effectstest.B12Creature(g, me.ID, "Five", "Creature — Human", 5, 5)
	effectstest.Station(t, g, me, ship, five)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 7 || c.CurrentToughness() != 4 || !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "hexproof") {
		t.Errorf("at 5+ a 7/4 with flying and hexproof; creature=%v %d/%d", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
}
