package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// c20_5513_b1_cards_test.go — Commander 2020 batch 1 (#5513). Every card is
// registered Full, and each one is played through the engine.

const (
	c20TrynnOracle          = "76771556-3eb7-4f87-aaeb-32ea81105aa2"
	c20SilvarOracle         = "09199ebd-2c09-455f-ac10-3dbff8e6d428"
	c20ShabrazOracle        = "7599fe9a-b6a5-491b-a308-aef3bf1e9daf"
	c20BrallinOracle        = "464df46a-079b-47c1-95e7-930160abef2f"
	c20NikaraOracle         = "2a2376b8-6962-4f9e-ba36-fd7de87dc77d"
	c20TitanHunterOracle    = "abd9ed47-e3ef-40f3-9c18-769228e68792"
	c20GlademuseOracle      = "028f2c1d-9c1d-4ac3-adf4-644b5166502b"
	c20JirinaOracle         = "b122bf99-b003-452a-a854-b25597ff7ca2"
	c20KelsienOracle        = "5b3326a5-18c5-4d45-90e1-f6d00ca2bced"
	c20FiendbonderOracle    = "0a746c49-cc44-467a-85b7-3c498559aea7"
	c20MycodraxOracle       = "f852adf6-3e54-4a1d-92e1-7dbc6d2228ce"
	c20TrilobiteOracle      = "28c0fd21-a88a-4a40-853f-2c1810517bc3"
	c20CuriousHerdOracle    = "e6f2e363-e361-4571-af37-6c63da04bdbf"
	c20VitalityOracle       = "d383c589-d8c7-4c4d-8084-37034d14b385"
	c20BarracudaOracle      = "edfc0a46-414e-467d-9f27-e42d08ec127a"
	c20SpecialistOracle     = "51feda69-0ae8-4df3-bb50-a4bb9deb965d"
	c20AltarOracle          = "a127d62a-8709-46bb-9301-96f09e7924a1"
	c20OrnamentOracle       = "634a61a1-47a4-4a23-b138-fb487e6a7cba"
	c20GigantotheriumOracle = "022624eb-c9d7-4110-bcc3-465c8d48fcd7"
	c20SawtuskOracle        = "7f93a368-9568-456c-9bd6-20d5fe5ad2f1"
	c20SnatcherOracle       = "ee8fb1d0-db86-489f-8d23-4238b2e8eba4"
	c20MindleecherOracle    = "4864cc98-4fde-4010-81e9-ea4944347e7e"

	c20BearType  = "Creature — Bear"
	c20HumanType = "Creature — Human Soldier"
)

// c20Power is a battlefield permanent's power with the layers caught up.
func c20Power(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c := apaLive(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.PowerForComparison()
}

// c20PickIfAsked answers an open target pick with `card`, if one is open,
// then passes priority so the item resolves.
func c20PickIfAsked(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	if latestPickTarget(g, chooser) == nil {
		passPriorityAroundTable(t, g)
	}
	if latestPickTarget(g, chooser) != nil {
		pickTriggerTarget(t, g, chooser, card)
	}
	passPriorityAroundTable(t, g)
}

// c20Mutate casts a catalog mutate card from `owner`'s hand onto `host`,
// over it, and resolves the spell.
func c20Mutate(t *testing.T, g *game.Game, owner, host uuid.UUID, c game.Card) uuid.UUID {
	t.Helper()
	id := rprMutatorInHand(t, g, owner, c)
	if err := castMutating(g, owner, id, host, false); err != nil {
		t.Fatalf("cast %s mutating: %v", c.Name, err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, owner, 0)
	return id
}

// --- every card is in the catalog ----------------------------------------

func TestC205513EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		c20TrynnOracle, c20SilvarOracle, c20ShabrazOracle, c20BrallinOracle, c20NikaraOracle,
		c20TitanHunterOracle, c20GlademuseOracle, c20JirinaOracle, c20KelsienOracle, c20FiendbonderOracle,
		c20MycodraxOracle, c20TrilobiteOracle, c20CuriousHerdOracle, c20VitalityOracle, c20BarracudaOracle,
		c20SpecialistOracle, c20AltarOracle, c20OrnamentOracle, c20GigantotheriumOracle, c20SawtuskOracle,
		c20SnatcherOracle, c20MindleecherOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// The four partners each carry the partner-with entry search naming the
// other half of their pair.
func TestC205513PartnersSearchForTheirPartner(t *testing.T) {
	for oracle, partner := range map[string]string{
		c20TrynnOracle:   "Silvar, Devourer of the Free",
		c20SilvarOracle:  "Trynn, Champion of Freedom",
		c20ShabrazOracle: "Brallin, Skyshark Rider",
		c20BrallinOracle: "Shabraz, the Skyshark",
		c20NikaraOracle:  "Yannik, Scavenging Sentinel",
	} {
		spec, _ := Lookup(oracle)
		found := false
		for _, tr := range spec.Triggered {
			if IsPartnerWithRow(tr) && tr.Key == spec.Name+partnerWithKeyMark+partner {
				found = true
			}
		}
		if !found {
			t.Errorf("%s has no partner-with row for %s", spec.Name, partner)
		}
	}
}

func TestC205513PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Silvar, Devourer of the Free", "Legendary Creature — Cat Nightmare", c20SilvarOracle, 4, 2, []string{"menace"}},
		{"Shabraz, the Skyshark", "Legendary Creature — Shark Bird", c20ShabrazOracle, 3, 3, []string{"flying"}},
		{"Nikara, Lair Scavenger", "Legendary Creature — Human Cleric", c20NikaraOracle, 2, 2, []string{"menace"}},
		{"Kelsien, the Plague", "Legendary Creature — Human Assassin", c20KelsienOracle, 2, 2, []string{"vigilance", "haste"}},
		{"Daring Fiendbonder", "Creature — Human Warlock", c20FiendbonderOracle, 5, 1, []string{"haste"}},
		{"Vitality Hunter", "Creature — Nightmare", c20VitalityOracle, 3, 4, []string{"lifelink"}},
		{"Sawtusk Demolisher", "Creature — Beast", c20SawtuskOracle, 6, 6, []string{"trample"}},
		{"Souvenir Snatcher", "Creature — Bird", c20SnatcherOracle, 4, 4, []string{"flying"}},
		{"Mindleecher", "Creature — Nightmare", c20MindleecherOracle, 5, 5, []string{"flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			assertKeywords(t, g, id, tc.keywords...)
		})
	}
}

// --- Trynn, Champion of Freedom --------------------------------------------

func TestC205513TrynnMakesASoldierOnlyAfterAnAttack(t *testing.T) {
	t.Run("attacked", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b12Push(g, me.ID, "Trynn, Champion of Freedom", "Legendary Creature — Human Soldier", c20TrynnOracle, 3, 3)
		bear := b12Creature(g, me.ID, "Bear", c20BearType, 2, 2)
		brAttack(t, g, bear)
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		if n := cardsNamed(g, "Human Soldier", me.ID); n != 1 {
			t.Errorf("%d Human Soldier tokens after attacking, want 1", n)
		}
	})
	t.Run("did not attack", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b12Push(g, me.ID, "Trynn, Champion of Freedom", "Legendary Creature — Human Soldier", c20TrynnOracle, 3, 3)
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		if n := cardsNamed(g, "Human Soldier", me.ID); n != 0 {
			t.Errorf("%d Human Soldier tokens without an attack, want 0", n)
		}
	})
}

// --- Silvar, Devourer of the Free ------------------------------------------

func TestC205513SilvarSacrificesAHumanForACounterAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	silvar := b12Push(g, me.ID, "Silvar, Devourer of the Free", "Legendary Creature — Cat Nightmare", c20SilvarOracle, 4, 2)
	bear := b12Creature(g, me.ID, "Bear", c20BearType, 2, 2)
	human := b12Creature(g, me.ID, "Villager", c20HumanType, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, silvar, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a Bear paid \"Sacrifice a Human\"")
	}
	apaActivate(t, g, me, silvar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{human}})
	if g.Battlefield.Contains(human) {
		t.Error("the Human was not sacrificed")
	}
	if n := plusOnes(g, silvar); n != 1 {
		t.Errorf("%d +1/+1 counters on Silvar, want 1", n)
	}
	if !hasAbility(effectiveAbilities(t, g, silvar), "indestructible") {
		t.Error("Silvar did not gain indestructible")
	}
}

// --- Shabraz, the Skyshark -------------------------------------------------

func TestC205513ShabrazGrowsAndGainsOnEachDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shabraz := b12Push(g, me.ID, "Shabraz, the Skyshark", "Legendary Creature — Shark Bird", c20ShabrazOracle, 3, 3)
	advanceToMain(t, g)
	life := me.Life
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, shabraz); n != 1 {
		t.Errorf("%d +1/+1 counters after a draw, want 1", n)
	}
	if me.Life != life+1 {
		t.Errorf("life %d after a draw, want %d", me.Life, life+1)
	}
}

func TestC205513ShabrazGivesAHumanFlying(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	shabraz := b12Push(g, me.ID, "Shabraz, the Skyshark", "Legendary Creature — Shark Bird", c20ShabrazOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", c20BearType, 2, 2)
	human := b12Creature(g, me.ID, "Villager", c20HumanType, 1, 1)
	apaMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, shabraz, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Fatal("Shabraz targeted a creature that isn't a Human")
	}
	apaActivate(t, g, me, shabraz, 0, game.ActivateAbilityParams{Targets: cardTarget(human)})
	if !hasAbility(effectiveAbilities(t, g, human), "flying") {
		t.Error("the Human did not gain flying")
	}
}

// --- Brallin, Skyshark Rider -----------------------------------------------

func TestC205513BrallinGrowsAndPingsOnYourDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	brallin := b12Push(g, me.ID, "Brallin, Skyshark Rider", "Legendary Creature — Human Shaman", c20BrallinOracle, 3, 3)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	var err error
	g.WithWriteLock(func() { err = g.DiscardRandomForEffect(me.ID, 1) })
	if err != nil {
		t.Fatalf("discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, brallin); n != 1 {
		t.Errorf("%d +1/+1 counters after a discard, want 1", n)
	}
	for _, p := range g.Seats {
		want := lives[p.ID] - 1
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("%s: life %d, want %d", p.Name, p.Life, want)
		}
	}
}

func TestC205513BrallinGivesASharkTrample(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	brallin := b12Push(g, me.ID, "Brallin, Skyshark Rider", "Legendary Creature — Human Shaman", c20BrallinOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", c20BearType, 2, 2)
	shark := b12Creature(g, me.ID, "Shark", "Creature — Shark", 2, 2)
	apaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, brallin, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Fatal("Brallin targeted a creature that isn't a Shark")
	}
	apaActivate(t, g, me, brallin, 0, game.ActivateAbilityParams{Targets: cardTarget(shark)})
	if !hasAbility(effectiveAbilities(t, g, shark), "trample") {
		t.Error("the Shark did not gain trample")
	}
}

// --- Nikara, Lair Scavenger ------------------------------------------------

func TestC205513NikaraDrawsWhenACreatureWithCountersLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Nikara, Lair Scavenger", "Legendary Creature — Human Cleric", c20NikaraOracle, 2, 2)
	grown := b12Creature(g, me.ID, "Grown Bear", c20BearType, 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(grown, game.CounterPlusOne, 1) })
	plain := b12Creature(g, me.ID, "Plain Bear", c20BearType, 2, 2)
	hand, life := me.Hand.Size(), me.Life
	b27Kill(g, plain)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatalf("a creature with no counters triggered Nikara: hand %d life %d", me.Hand.Size(), me.Life)
	}
	b27Kill(g, grown)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

// --- Titan Hunter ----------------------------------------------------------

func TestC205513TitanHunterHitsTheActivePlayerOnAQuietTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Titan Hunter", "Creature — Human Warrior", c20TitanHunterOracle, 4, 5)
	life := me.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("the active player's life %d, want %d", me.Life, life-4)
	}
}

func TestC205513TitanHunterIsQuietAfterADeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Titan Hunter", "Creature — Human Warrior", c20TitanHunterOracle, 4, 5)
	bear := b12Creature(g, me.ID, "Bear", c20BearType, 2, 2)
	advanceToMain(t, g)
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	life := me.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a creature died and the Hunter still dealt damage: life %d, want %d", me.Life, life)
	}
}

func TestC205513TitanHunterSacrificesForLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hunter := b12Push(g, me.ID, "Titan Hunter", "Creature — Human Warrior", c20TitanHunterOracle, 4, 5)
	bear := b12Creature(g, me.ID, "Bear", c20BearType, 2, 2)
	life := me.Life
	apaMana(me, "C", "B")
	apaActivate(t, g, me, hunter, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

// --- Glademuse -------------------------------------------------------------

func TestC205513GlademuseDrawsForAnOffTurnCaster(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Glademuse", "Creature — Beast", c20GlademuseOracle, 2, 4)
	advanceToMain(t, g)
	theirs := opp.Hand.Size()
	b29CastAs(t, g, opp, "Opt", "Instant", "", "{U}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != theirs+1 {
		t.Errorf("an opponent casting on my turn: hand %d, want %d", got, theirs+1)
	}
	mine := me.Hand.Size()
	b29CastAs(t, g, me, "Shock", "Instant", "", "{R}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != mine {
		t.Errorf("casting on my own turn drew: hand %d, want %d", got, mine)
	}
}

// --- Jirina Kudro ----------------------------------------------------------

func TestC205513JirinaMakesASoldierPerCommanderCastAndPumpsHumans(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.CommanderCasts = map[uuid.UUID]int{uuid.New(): 2}
	bear := b12Creature(g, me.ID, "Bear", c20BearType, 2, 2)
	jirina := game.Card{InstanceID: uuid.New(), Name: "Jirina Kudro", TypeLine: "Legendary Creature — Human Soldier",
		OracleID: c20JirinaOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID}
	castFromHandAndResolve(t, g, me, jirina)
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Human Soldier", me.ID); n != 2 {
		t.Fatalf("%d Human Soldier tokens after two commander casts, want 2", n)
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Human Soldier" && c.Controller == me.ID {
			if p := c20Power(t, g, c.InstanceID); p != 3 {
				t.Errorf("a Human Soldier token has power %d, want 3 (1 + 2)", p)
			}
		}
	}
	if p := c20Power(t, g, jirina.InstanceID); p != 3 {
		t.Errorf("Jirina pumps herself: power %d, want 3", p)
	}
	if p := c20Power(t, g, bear); p != 2 {
		t.Errorf("a Bear is pumped: power %d, want 2", p)
	}
}

// --- Kelsien, the Plague ---------------------------------------------------

func TestC205513KelsienPingsAndGainsExperienceWhenItKills(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kelsien := b12Push(g, me.ID, "Kelsien, the Plague", "Legendary Creature — Human Assassin", c20KelsienOracle, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", c20BearType, 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", c20BearType, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, kelsien, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(mine)}); err == nil {
		t.Fatal("Kelsien targeted a creature you control")
	}
	apaActivate(t, g, me, kelsien, 0, game.ActivateAbilityParams{Targets: cardTarget(victim)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the 1/1 survived 1 damage")
	}
	if n := me.Counters[game.CounterExperience]; n != 1 {
		t.Errorf("%d experience counters, want 1", n)
	}
	if p := c20Power(t, g, kelsien); p != 3 {
		t.Errorf("Kelsien's power %d with one experience counter, want 3", p)
	}
}

// --- Daring Fiendbonder ----------------------------------------------------

func TestC205513FiendbonderGivesAnIndestructibleCounterFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", c20BearType, 2, 2)
	id, _ := activateFromGraveyard(t, g, "Daring Fiendbonder", "Creature — Human Warlock", c20FiendbonderOracle, 5, 1, "",
		game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !g.Exile.Contains(id) {
		t.Fatal("the cost did not exile the card")
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterIndestructible); n != 1 {
		t.Errorf("%d indestructible counters, want 1", n)
	}
}

// --- Boneyard Mycodrax -----------------------------------------------------

func TestC205513MycodraxCountsCreatureCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wthGraveyardCard(me, "Dead Bear", c20BearType)
	wthGraveyardCard(me, "Dead Elf", "Creature — Elf")
	wthGraveyardCard(me, "Spent Shock", "Instant")
	myco := b12Push(g, me.ID, "Boneyard Mycodrax", "Creature — Fungus", c20MycodraxOracle, 0, 0)
	if p := c20Power(t, g, myco); p != 2 {
		t.Errorf("power %d with two creature cards in the graveyard, want 2", p)
	}
}

func TestC205513MycodraxScavengesItsGraveyardCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wthGraveyardCard(me, "Dead Bear", c20BearType)
	wthGraveyardCard(me, "Dead Elf", "Creature — Elf")
	target := b12Creature(g, me.ID, "Bear", c20BearType, 2, 2)
	activateFromGraveyard(t, g, "Boneyard Mycodrax", "Creature — Fungus", c20MycodraxOracle, 0, 0, "",
		game.ActivateAbilityParams{Targets: cardTarget(target)})
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, target); n != 2 {
		t.Errorf("%d +1/+1 counters from scavenge, want 2", n)
	}
}

// --- Cryptic Trilobite -----------------------------------------------------

func TestC205513TrilobiteEntersWithXAndMakesColorlessFromACounter(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Cryptic Trilobite", TypeLine: "Creature — Trilobite",
		OracleID: c20TrilobiteOracle, ManaCost: "{X}{X}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, id); n != 2 {
		t.Fatalf("%d +1/+1 counters with X=2, want 2", n)
	}
	pool := me.ManaPool.Total()
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := me.ManaPool.Total(); got != pool+2 {
		t.Errorf("pool %d after the mana ability, want %d", got, pool+2)
	}
	if n := plusOnes(g, id); n != 1 {
		t.Errorf("%d +1/+1 counters after paying one, want 1", n)
	}
}

func TestC205513TrilobiteTapsForACounter(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cryptic Trilobite",
		TypeLine: "Creature — Trilobite", OracleID: c20TrilobiteOracle, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1}})
	apaMana(me, "C")
	apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if n := plusOnes(g, id); n != 2 {
		t.Errorf("%d +1/+1 counters after the tap ability, want 2", n)
	}
	if !hmlTapped(t, g, id) {
		t.Error("the Trilobite did not tap")
	}
}

// --- Curious Herd ----------------------------------------------------------

func TestC205513CuriousHerdCountsTheTargetOpponentsArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	b12Permanent(g, opp.ID, "Rock", "Artifact")
	b12Permanent(g, opp.ID, "Another Rock", "Artifact")
	b12Permanent(g, other.ID, "Not Theirs", "Artifact")
	castCatalogSpell(t, g, "Curious Herd", "Instant", c20CuriousHerdOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Beast", me.ID); n != 2 {
		t.Errorf("%d Beasts, want 2 (the target opponent's artifacts)", n)
	}
}

// --- Vitality Hunter -------------------------------------------------------

func TestC205513VitalityHunterGivesLifelinkCountersUpToX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := b12Push(g, me.ID, "Vitality Hunter", "Creature — Nightmare", c20VitalityOracle, 3, 4)
	a := b12Creature(g, me.ID, "A", c20BearType, 2, 2)
	b := b12Creature(g, opp.ID, "B", c20BearType, 2, 2)
	activateMonstrosity(t, g, me.ID, hunter, 2)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, hunter); n != 2 {
		t.Errorf("%d +1/+1 counters on the Hunter, want 2", n)
	}
	for _, id := range []uuid.UUID{a, b} {
		if n := countersOn(g, id, game.CounterLifelink); n != 1 {
			t.Errorf("%s: %d lifelink counters, want 1", id, n)
		}
	}
}

// --- Tidal Barracuda -------------------------------------------------------

// c20CastSorcery puts a vanilla sorcery in the caster's hand and casts it.
func c20CastSorcery(g *game.Game, caster *game.Player) error {
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Sorcery", TypeLine: "Sorcery",
		Owner: caster.ID, Controller: caster.ID})
	return g.CastSpell(caster.ID, id, game.CastSpellParams{})
}

func TestC205513TidalBarracudaGivesEveryoneFlash(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepBeginCombat)
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := c20CastSorcery(g, active); err == nil {
		t.Fatal("a sorcery was cast at the beginning of combat with no Barracuda")
	}
	seedEnchantment(g, c20BarracudaOracle, "Tidal Barracuda", "Creature — Fish", owner.ID)
	if err := c20CastSorcery(g, active); err != nil {
		t.Errorf("the Barracuda's flash did not reach another player: %v", err)
	}
}

func TestC205513TidalBarracudaLocksOpponentsOutOfYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	active, opp := g.Seats[0], g.Seats[1]
	seedEnchantment(g, c20BarracudaOracle, "Tidal Barracuda", "Creature — Fish", active.ID)
	e7Main(t, g)
	assertCantCast(t, e7CastInstant(g, opp), "during your turn")
	if err := e7CastInstant(g, active); err != nil {
		t.Errorf("the Barracuda's controller was refused: %v", err)
	}
}

// --- Species Specialist ----------------------------------------------------

func TestC205513SpeciesSpecialistDrawsWhenTheChosenTypeDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Species Specialist",
		TypeLine: "Creature — Human Warrior", OracleID: c20SpecialistOracle, Power: 2, Toughness: 3,
		Owner: me.ID, Controller: me.ID, NamedTribe: "Goblin"})
	goblin := b12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	elf := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	hand := me.Hand.Size()
	b27Kill(g, elf)
	passPriorityAroundTable(t, g)
	if b19HasTriggerPromptFor(g, me.ID) || me.Hand.Size() != hand {
		t.Fatal("a creature of another type triggered the Specialist")
	}
	b27Kill(g, goblin)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d after a Goblin died, want %d", me.Hand.Size(), hand+1)
	}
}

// --- Netherborn Altar ------------------------------------------------------

func TestC205513NetherbornAltarReturnsYourCommanderForLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	cmdr := uuid.New()
	me.Command.PushTop(game.Card{InstanceID: cmdr, Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Owner: me.ID, Controller: me.ID, IsCommander: true})
	altar := pushCatalogPermanent(g, me.ID, "Netherborn Altar", "Artifact", c20AltarOracle, false)
	life := me.Life
	apaActivate(t, g, me, altar, 0, game.ActivateAbilityParams{})
	if !me.Hand.Contains(cmdr) || me.Command.Contains(cmdr) {
		t.Error("the commander did not move to your hand")
	}
	if n := countersOn(g, altar, c20SoulCounter); n != 1 {
		t.Errorf("%d soul counters, want 1", n)
	}
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
}

// --- Bonder's Ornament -----------------------------------------------------

func TestC205513BondersOrnamentDrawsForEachOrnamentController(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	mine := pushCatalogPermanent(g, me.ID, "Bonder's Ornament", "Artifact", c20OrnamentOracle, false)
	pushCatalogPermanent(g, opp.ID, "Bonder's Ornament", "Artifact", c20OrnamentOracle, false)
	hands := map[uuid.UUID]int{me.ID: me.Hand.Size(), opp.ID: opp.Hand.Size(), other.ID: other.Hand.Size()}
	apaMana(me, "C", "C", "C", "C")
	apaActivate(t, g, me, mine, 0, game.ActivateAbilityParams{})
	for p, want := range map[*game.Player]int{me: hands[me.ID] + 1, opp: hands[opp.ID] + 1, other: hands[other.ID]} {
		if got := p.Hand.Size(); got != want {
			t.Errorf("%s: hand %d, want %d", p.Name, got, want)
		}
	}
}

// --- Ravenous Gigantotherium -----------------------------------------------

func TestC205513GigantotheriumDevoursThenDividesItsPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	fodder := pushDevourFodder(g, me.ID, "Goblin")
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 2, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 1, 30)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ravenous Gigantotherium", TypeLine: "Creature — Beast",
		OracleID: c20GigantotheriumOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, fodder)
	if latestPickTarget(g, me.ID) == nil {
		passPriorityAroundTable(t, g)
	}
	if n := plusOnes(g, id); n != 3 {
		t.Fatalf("%d +1/+1 counters after devouring one, want 3", n)
	}
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 2, b: 4}, a, b)
	passPriorityAroundTable(t, g)
	if d1, d2 := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; d1 != 2 || d2 != 4 {
		t.Errorf("division landed %d/%d, want 2/4", d1, d2)
	}
	if d := e2Card(t, g, id).DamageMarked; d != 3 {
		t.Errorf("the Gigantotherium took %d, want 3 (powers 2 + 1)", d)
	}
}

// --- mutate triggers ---------------------------------------------------------

func TestC205513SawtuskDemolisherDestroysAndPaysABeast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elves := llanowarOn(g, me.ID)
	charm := b12Permanent(g, opp.ID, "Their Charm", "Enchantment")
	c20Mutate(t, g, me.ID, elves, game.Card{Name: "Sawtusk Demolisher", TypeLine: "Creature — Beast",
		OracleID: c20SawtuskOracle, ManaCost: "{4}{G}{G}", Colors: []string{"G"}, Power: 6, Toughness: 6})
	c20PickIfAsked(t, g, me.ID, charm)
	if g.Battlefield.Contains(charm) {
		t.Error("the enchantment was not destroyed")
	}
	if n := cardsNamed(g, "Beast", opp.ID); n != 1 {
		t.Errorf("its controller has %d Beasts, want 1", n)
	}
}

func TestC205513SouvenirSnatcherStealsANoncreatureArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elves := llanowarOn(g, me.ID)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	c20Mutate(t, g, me.ID, elves, game.Card{Name: "Souvenir Snatcher", TypeLine: "Creature — Bird",
		OracleID: c20SnatcherOracle, ManaCost: "{4}{U}", Colors: []string{"U"}, Power: 4, Toughness: 4})
	c20PickIfAsked(t, g, me.ID, rock)
	if c := findBattlefieldCardByID(g, rock); c == nil || c.Controller != me.ID {
		t.Errorf("the artifact is not under your control: %+v", c)
	}
}

func TestC205513MindleecherExilesEachOpponentsTopCardForYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := llanowarOn(g, me.ID)
	advanceToMain(t, g)
	tops := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			tops[p.ID] = p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		}
	}
	mine := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	c20Mutate(t, g, me.ID, elves, game.Card{Name: "Mindleecher", TypeLine: "Creature — Nightmare",
		OracleID: c20MindleecherOracle, ManaCost: "{4}{B}{B}", Colors: []string{"B"}, Power: 5, Toughness: 5})
	passPriorityAroundTable(t, g)
	for owner, id := range tops {
		if !g.Exile.Contains(id) {
			t.Errorf("%s's top card was not exiled", owner)
			continue
		}
		if grantedPermissionOn(g, me.ID, id, game.ZoneExile) == nil {
			t.Errorf("you may not play %s's exiled card", owner)
		}
	}
	if g.Exile.Contains(mine) {
		t.Error("your own top card was exiled")
	}
}
