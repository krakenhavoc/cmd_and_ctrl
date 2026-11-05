package effects

import (
	"testing"

	"github.com/google/uuid"

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
	return pushBattlefieldCardWithTimestamp(g, c)
}

// trekB42Pick answers the open choose_cards prompt of `chooser`.
func trekB42Pick(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	ch := latestChooseCards(g, chooser)
	if ch == nil {
		t.Fatalf("no choose_cards prompt for %s", chooser)
	}
	if err := g.ResolveChooseCards(ch.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
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
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
}

// --- K'Vort-Class Raider --------------------------------------------------------

func TestTrekB42KVortShieldsYourSideOnYourTurnAndStationsAtFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ship := trekB42Ship(g, me.ID, "K'Vort-Class Raider", "Artifact — Spacecraft", trekB42KVortOracle, 5, 3, 0)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	if !hasEffectiveKeyword(t, g, ship, "hexproof") || !hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("on your turn the Spacecraft and your creature have hexproof")
	}
	if hasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("an opponent's creature got hexproof")
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Fatal("an unstationed Spacecraft is a creature")
	}

	four := b12Creature(g, me.ID, "Four", "Creature — Human", 4, 4)
	station(t, g, me, ship, four)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 5 || c.CurrentToughness() != 3 {
		t.Fatalf("at 4+ it is a 5/3 artifact creature; creature=%v %d/%d", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
	for _, kw := range []string{"flying", "menace", "haste"} {
		if !game.HasKeyword(&c, kw) {
			t.Errorf("at 4+ it lacks %s", kw)
		}
	}

	advanceToMainOf(t, g, 1)
	if hasEffectiveKeyword(t, g, ship, "hexproof") || hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("hexproof lasted into an opponent's turn")
	}
}

// --- Planet Killer --------------------------------------------------------------

func TestTrekB42PlanetKillerMustAttackAndGrowsPerDefendingLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pk := b12Push(g, me.ID, "Planet Killer", "Artifact Creature — Construct", trekB42PlanetOracle, 9, 9)
	for _, n := range []string{"Forest", "Island", "Swamp"} {
		b12Permanent(g, opp.ID, n, "Basic Land — "+n)
	}
	b12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	if !hasEffectiveKeyword(t, g, pk, "flying") || !hasEffectiveKeyword(t, g, pk, "indestructible") {
		t.Error("Planet Killer lacks flying or indestructible")
	}

	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); err == nil {
		t.Fatal("passing with Planet Killer home succeeded, want the attack requirement to refuse it")
	}
	if err := g.DeclareAttacker(pk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != 40-9 {
		t.Fatalf("opponent life %d, want 31", opp.Life)
	}
	if got := countersOn(g, pk, game.CounterPlusOne); got != 3 {
		t.Errorf("Planet Killer has %d +1/+1 counters, want 3 (the damaged player's lands)", got)
	}
}

// --- Queen's Diamond ------------------------------------------------------------

func TestTrekB42QueensDiamondAddsChargeManaOnlyForYourCommander(t *testing.T) {
	g, me, opp := spendTable(t)
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
	dCard, _ := battlefieldCard(g, diamond)
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(dCard)) {
		t.Error("the mana pays for an activated ability")
	}

	wrong := handSpell(me, "Klingon Warrior", "Creature — Klingon", "{2}{R}")
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	diamond := trekB42Ship(g, me.ID, "Queen's Diamond", "Legendary Artifact — Spacecraft", trekB42DiamondOracle, 5, 5, 0)
	if err := g.ActivateManaAbility(me.ID, diamond, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil && me.ManaPool.Total() != 0 {
		t.Errorf("with no charge counters it added %d mana", me.ManaPool.Total())
	}
	seven := b12Creature(g, me.ID, "Seven", "Creature — Human", 7, 7)
	station(t, g, me, diamond, seven)
	passPriorityAroundTable(t, g)
	if layeredCard(t, g, diamond).IsCreature() {
		t.Fatal("a creature at 7 charge counters")
	}
	one := b12Creature(g, me.ID, "One", "Creature — Human", 1, 1)
	station(t, g, me, diamond, one)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, diamond)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rp := castCreatureCard(t, g, "Reconnaissance Partition", "Artifact Creature — Borg", trekB42PartitionOracle, 3, 3)
	if got := countersOn(g, rp, game.CounterCharge); got != 2 {
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

	stackLibrary(me, "Drawn", "Sorcery", "")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, rp, 0, game.ActivateAbilityParams{CounterKind: game.CounterCharge, Modes: []int{1}}); err != nil {
		t.Fatalf("draw mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
	if got := countersOn(g, rp, game.CounterCharge); got != 1 {
		t.Errorf("%d charge counters left, want 1", got)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rp {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	forest := stackLibrary(me, "Forest", "Basic Land — Forest", "")
	if err := g.ActivateCatalogAbility(me.ID, rp, 0, game.ActivateAbilityParams{CounterKind: game.CounterCharge, Modes: []int{0}}); err != nil {
		t.Fatalf("land mode: %v", err)
	}
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, forest); !ok || !c.Tapped {
		t.Errorf("the basic land is on the battlefield %v, tapped %v; want both", ok, c.Tapped)
	}
	if got := countersOn(g, rp, game.CounterCharge); got != 0 {
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
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	ship := trekB42Ship(g, me.ID, "S.S. Botany Bay", "Legendary Artifact — Spacecraft", trekB42BotanyOracle, 3, 5, 8)
	c1 := trekB42Graveyard(opp1, "Their Bear", "Creature — Bear")
	trekB42Graveyard(opp1, "Their Spell", "Instant")
	c2 := trekB42Graveyard(opp2, "Other Bear", "Creature — Bear")
	khan := trekB42Graveyard(me, "Khan", "Creature — Human")
	trekB42Graveyard(me, "My Spell", "Sorcery")

	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestChooseCards(g, g.Seats[3].ID) != nil {
		t.Error("an opponent with an empty graveyard was asked")
	}
	trekB42Pick(t, g, opp1.ID, c1)
	trekB42Pick(t, g, opp2.ID, c2)
	ch := latestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("expected your may-pick of one card: %+v", ch)
	}
	trekB42Pick(t, g, me.ID, khan)
	passPriorityAroundTable(t, g)
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
	if c := layeredCard(t, g, ship); !c.IsCreature() || !game.HasKeyword(&c, "flying") {
		t.Fatal("at 8+ Botany Bay is a flying creature")
	}
	attackWith(t, g, opp1.ID, ship)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castFromHand(t, g, game.Card{Name: "Starfleet Communicator", TypeLine: "Artifact — Communicator Equipment",
		OracleID: trekB42CommOracle, ManaCost: "{3}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	var comms []uuid.UUID
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Starfleet Communicator" && c.Controller == me.ID {
			comms = append(comms, c.InstanceID)
			if IsToken(c) {
				tokens++
			}
		}
	}
	if len(comms) != 2 || tokens != 1 {
		t.Fatalf("%d Communicators (%d tokens), want the card and one token copy", len(comms), tokens)
	}

	a := b12Creature(g, me.ID, "Attacker", "Creature — Human", 2, 2)
	b := b12Creature(g, me.ID, "Crew", "Creature — Human", 2, 2)
	equipTo(t, g, me.ID, comms[0], a)
	equipTo(t, g, me.ID, comms[1], b)

	stackLibrary(me, "Drawn Land", "Basic Land — Island", "")
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Hand Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	trekB42Attack(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d, want X = 1 card drawn", hand, me.Hand.Size())
	}
	ch := latestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("expected up to one land from hand: %+v", ch)
	}
	trekB42Pick(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, land); !ok || !c.Tapped {
		t.Errorf("the land is on the battlefield %v, tapped %v; want both", ok, c.Tapped)
	}
}

func TestTrekB42CommunicatorCountsOnlyOtherEquippedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	comm := b12Push(g, me.ID, "Starfleet Communicator", "Artifact — Communicator Equipment", trekB42CommOracle, 0, 0)
	a := b12Creature(g, me.ID, "Attacker", "Creature — Human", 2, 2)
	b12Creature(g, me.ID, "Bare", "Creature — Human", 2, 2)
	equipTo(t, g, me.ID, comm, a)
	stackLibrary(me, "Drawn", "Sorcery", "")
	hand := me.Hand.Size()
	trekB42Attack(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d; X is 0 with no other Communicator-equipped creature", hand, me.Hand.Size())
	}
}

// --- Starship Enterprise --------------------------------------------------------

func TestTrekB42EnterpriseLandfallChargesAndDrawsOnDamageAtSixteen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := trekB42Ship(g, me.ID, "Starship Enterprise", "Legendary Artifact — Spacecraft", trekB42EnterpriseOracle, 4, 6, 12)
	if layeredCard(t, g, ship).IsCreature() {
		t.Fatal("a creature at 12 charge counters")
	}
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if got := countersOn(g, ship, game.CounterCharge); got != 16 {
		t.Fatalf("%d charge counters after landfall, want 16", got)
	}
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 4 || c.CurrentToughness() != 6 || !game.HasKeyword(&c, "flying") {
		t.Fatalf("at 16+ a 4/6 flier; creature=%v %d/%d flying=%v", c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"))
	}
	for i := 0; i < 5; i++ {
		trekB23Library(me, game.Card{Name: "Card", TypeLine: "Sorcery"})
	}
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, ship)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+4 {
		t.Errorf("hand %d → %d, want four cards drawn (the damage dealt)", hand, me.Hand.Size())
	}
}

// --- The Sword of Kahless -------------------------------------------------------

func TestTrekB42SwordOfKahlessPumpsAndPutsAKlingonFromTheTopFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sword := b12Push(g, me.ID, "The Sword of Kahless", "Legendary Artifact — Equipment", trekB42KahlessOracle, 0, 0)
	bearer := b12Creature(g, me.ID, "Bearer", "Creature — Human", 2, 2)
	equipTo(t, g, me.ID, sword, bearer)
	if c := layeredCard(t, g, bearer); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
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
	passPriorityAroundTable(t, g)
	ch := latestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("expected a may-pick of one card: %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0]}); err == nil {
		t.Error("a creature that is neither Klingon nor Warrior was accepted")
	}
	trekB42Pick(t, g, me.ID, lib[1])
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(lib[1]) && !trekB42OnBattlefieldNamed(g, me.ID, "Worf") {
		t.Fatal("the Klingon was not put onto the battlefield")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != lib[5] {
		t.Errorf("the library top is %v, want the sixth card (the rest went to the bottom)", top.Name)
	}
}

// --- U.S.S. Defiant, Mirror Migrant ---------------------------------------------

func TestTrekB42MirrorMigrantPhasesOutWithACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ship := castHoldCreature(t, g, "U.S.S. Defiant, Mirror Migrant", "Legendary Artifact — Spacecraft", trekB42MigrantOracle, 5, 5)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("an opponent's creature was accepted")
	}
	answerTriggerTargets(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if !g.PhasedOut.Contains(ship) || !g.PhasedOut.Contains(mine) {
		t.Errorf("phased out: Defiant %v, creature %v; want both", g.PhasedOut.Contains(ship), g.PhasedOut.Contains(mine))
	}
	if spec, _ := Lookup(trekB42MigrantOracle); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Error("the Defiant does not print flash")
	}
}

// --- U.S.S. Defiant, Tough Little Ship ------------------------------------------

func TestTrekB42ToughLittleShipDestroysAnArtifactAndStationsAtFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ship := castHoldCreature(t, g, "U.S.S. Defiant, Tough Little Ship", "Legendary Artifact — Spacecraft", trekB42ToughOracle, 7, 4)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("a creature was accepted as an artifact or enchantment")
	}
	answerTriggerTargets(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}

	five := b12Creature(g, me.ID, "Five", "Creature — Human", 5, 5)
	station(t, g, me, ship, five)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 7 || c.CurrentToughness() != 4 || !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "hexproof") {
		t.Errorf("at 5+ a 7/4 with flying and hexproof; creature=%v %d/%d", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
}
