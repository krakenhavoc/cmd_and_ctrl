package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b10_cards_test.go — Secrets of Strixhaven Commander, batch 10
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	socB10ParasiticOracle = "ff3de07d-86c3-49e1-9de0-97c85301ff93"
	socB10PearlEarOracle  = "9bdc1e63-1cf9-4ddb-a7d5-f2a2eb91bdfd"
	socB10InfestOracle    = "1b704798-0c69-4c18-ac7e-42933ce90028"
	socB10RescuerOracle   = "02be0b99-d40b-4287-989b-12eb2701115c"
	socB10PlarggOracle    = "9e769a65-7aa5-484d-9493-badb485b5964"
	socB10PrimoOracle     = "65fb0e20-61f8-4c61-bc08-7f6f718ee9ee"
	socB10PrismariCampus  = "3a3a1b35-ae4d-49d5-ae09-5a1693ad53ce"
	socB10PianistOracle   = "f58117e4-ba85-41b8-8fd5-4245716a84dc"
	socB10ApprenticeOracl = "92116bbd-c7ad-4033-8d98-3ea09fea8ae7"
	socB10QuandrixCampus  = "172f86b6-9580-4eb2-b7dc-2a44277d978b"
)

func TestSocB10EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB10ParasiticOracle: "Parasitic Impetus",
		socB10PearlEarOracle:  "Pearl-Ear, Imperial Advisor",
		socB10InfestOracle:    "Pest Infestation",
		socB10RescuerOracle:   "Pest Rescuer",
		socB10PlarggOracle:    "Plargg and Nassari",
		socB10PrimoOracle:     "Primo, the Unbounded",
		socB10PrismariCampus:  "Prismari Campus",
		socB10PianistOracle:   "Prismari Pianist",
		socB10ApprenticeOracl: "Quandrix Apprentice",
		socB10QuandrixCampus:  "Quandrix Campus",
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

// socB10Named lists the battlefield permanents with the name.
func socB10Named(g *game.Game, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- Parasitic Impetus ---

func TestSocB10ParasiticImpetusPumpsGoadsAndDrainsTheCreaturesController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	bear := auraBear(g, opp.ID)
	auraCast(t, g, "Parasitic Impetus", socB10ParasiticOracle, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("enchanted bear is %d/%d, want 4/4", p, tg)
	}
	if c, ok := g.LookupCardForEffect(bear); !ok || !c.Goaded() {
		t.Error("the enchanted creature is not goaded")
	}
	myLife, oppLife := me.Life, opp.Life
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the creature's controller is at %d, want %d", opp.Life, oppLife-2)
	}
	if me.Life != myLife+2 {
		t.Errorf("the Aura's controller is at %d, want %d", me.Life, myLife+2)
	}
}

// --- Pearl-Ear, Imperial Advisor ---

func TestSocB10PearlEarGivesYourEnchantmentSpellsAffinityForAuras(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	b12Push(g, me.ID, "Pearl-Ear, Imperial Advisor", "Legendary Creature — Fox Advisor", socB10PearlEarOracle, 3, 4)
	b12Permanent(g, me.ID, "Aura A", "Enchantment — Aura")
	b12Permanent(g, me.ID, "Aura B", "Enchantment — Aura")
	b12Permanent(g, me.ID, "Plain Enchantment", "Enchantment")
	spell := handCardFull(me, "Test Enchantment", "Enchantment", "{3}{W}", "", []string{"W"})
	b06AddMana(me, "C", "W")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("two Auras should make {3}{W} cost {1}{W}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want 0", me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)

	// A creature spell gets no reduction.
	creature := handCardFull(me, "Test Creature", "Creature — Bear", "{3}{W}", "", []string{"W"})
	b06AddMana(me, "C", "W")
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature spell was cheaper with two Auras out")
	}
}

func TestSocB10PearlEarDrawsWhenAnAuraTargetsAModifiedPermanentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	b12Push(g, me.ID, "Pearl-Ear, Imperial Advisor", "Legendary Creature — Fox Advisor", socB10PearlEarOracle, 3, 4)
	plain := auraBear(g, me.ID)
	counter := auraBear(g, me.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == counter {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Test Aura", auraTypeLine, "", []game.TargetRef{{Kind: game.TargetCard, ID: plain}})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("an Aura on an unmodified creature drew: hand %d, want %d", got, hand)
	}
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Test Aura", auraTypeLine, "", []game.TargetRef{{Kind: game.TargetCard, ID: counter}})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("an Aura on a creature with a counter: hand %d, want %d", got, hand+1)
	}
}

func TestSocB10PearlEarDoesNotDrawForAnOpponentsModifiedPermanentOrANonAura(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	b12Push(g, me.ID, "Pearl-Ear, Imperial Advisor", "Legendary Creature — Fox Advisor", socB10PearlEarOracle, 3, 4)
	mine := auraBear(g, me.ID)
	theirs := auraBear(g, opp.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			id := g.Battlefield.Cards[i].InstanceID
			if id == mine || id == theirs {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Test Aura", auraTypeLine, "", []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("an Aura on an opponent's modified creature drew: hand %d, want %d", got, hand)
	}
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Test Instant", "Instant", "", []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("a non-Aura spell drew: hand %d, want %d", got, hand)
	}
}

// --- Pest Infestation ---

func TestSocB10PestInfestationDestroysUpToXAndMakesTwiceXPests(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Permanent(g, opp.ID, "Rock", "Artifact")
	e := b12Permanent(g, opp.ID, "Shrine", "Enchantment")
	keep := b12Permanent(g, opp.ID, "Spare Rock", "Artifact")
	castXSpell(t, g, "Pest Infestation", "Sorcery", socB10InfestOracle, "{X}{X}{G}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: e}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(e) {
		t.Error("a targeted artifact or enchantment survived")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("an untargeted artifact was destroyed")
	}
	pests := socB10Named(g, "Pest")
	if len(pests) != 4 {
		t.Fatalf("%d Pests, want twice X = 4", len(pests))
	}
	for _, id := range pests {
		if c, _ := g.LookupCardForEffect(id); c.Controller != me.ID {
			t.Errorf("a Pest is under %v, want the caster", c.Controller)
		}
	}
	life := me.Life
	killCreature(t, g, me.ID, pests[0])
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("a Pest dying gained %d, want 1", me.Life-life)
	}
}

func TestSocB10PestInfestationWithNoTargetsStillMakesPests(t *testing.T) {
	g := newCatalogGame(t)
	castXSpell(t, g, "Pest Infestation", "Sorcery", socB10InfestOracle, "{X}{X}{G}", 3, nil)
	passPriorityAroundTable(t, g)
	if n := len(socB10Named(g, "Pest")); n != 6 {
		t.Errorf("%d Pests, want 6 for X = 3 and no targets", n)
	}
}

func TestSocB10PestInfestationRefusesTooManyTargetsAndACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Permanent(g, opp.ID, "Rock", "Artifact")
	b := b12Permanent(g, opp.ID, "Rock Two", "Artifact")
	bear := auraBear(g, opp.ID)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Pest Infestation", TypeLine: "Sorcery",
		OracleID: socB10InfestOracle, ManaCost: "{X}{X}{G}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 1, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}}); err == nil {
		t.Error("two targets were accepted on X = 1")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 1, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear},
	}}); err == nil {
		t.Error("a creature was accepted as a target")
	}
}

// --- Pest Rescuer ---

func TestSocB10PestRescuerMakesAPestOnEachUpkeepWhenYouHaveNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Pest Rescuer", "Creature — Dryad Druid", socB10RescuerOracle, 2, 2)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	pests := socB10Named(g, "Pest")
	if len(pests) != 1 {
		t.Fatalf("%d Pests after an opponent's upkeep, want 1", len(pests))
	}
	if c, _ := g.LookupCardForEffect(pests[0]); c.Controller != me.ID {
		t.Error("the Pest is not yours")
	}
	// With a Pest already out, the next upkeep makes none.
	advanceToUpkeepOf(t, g, 2)
	passPriorityAroundTable(t, g)
	if n := len(socB10Named(g, "Pest")); n != 1 {
		t.Errorf("%d Pests, want still 1 (you control one)", n)
	}
	// Lose it and the next upkeep, anyone's, makes another.
	killCreature(t, g, me.ID, pests[0])
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 3)
	passPriorityAroundTable(t, g)
	if n := len(socB10Named(g, "Pest")); n != 1 {
		t.Errorf("%d Pests, want a new one after the first died", n)
	}
}

func TestSocB10PestRescuerChecksThePestAgainOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Pest Rescuer", "Creature — Dryad Druid", socB10RescuerOracle, 2, 2)
	advanceToUpkeepOf(t, g, 1)
	if n := len(socB10Named(g, "Pest")); n != 0 {
		t.Fatalf("%d Pests before the trigger resolved", n)
	}
	// A Pest arrives in response.
	pest := PestToken()
	pest.InstanceID, pest.Owner, pest.Controller = uuid.New(), me.ID, me.ID
	pushBattlefieldCardWithTimestamp(g, pest)
	passPriorityAroundTable(t, g)
	if n := len(socB10Named(g, "Pest")); n != 1 {
		t.Errorf("%d Pests, want 1: the trigger's intervening if is rechecked on resolution", n)
	}
}

func TestSocB10PestRescuerAddsOneToYourLifeGainOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Pest Rescuer", "Creature — Dryad Druid", socB10RescuerOracle, 2, 2)
	life, oppLife := me.Life, opp.Life
	gainLifeNow(t, g, uuid.Nil, me, 3)
	if me.Life != life+4 {
		t.Errorf("gaining 3 left you at %d, want %d", me.Life, life+4)
	}
	gainLifeNow(t, g, uuid.Nil, opp, 3)
	if opp.Life != oppLife+3 {
		t.Errorf("an opponent's gain of 3 left them at %d, want %d", opp.Life, oppLife+3)
	}
}

// --- Primo, the Unbounded ---

func TestSocB10PrimoEntersWithTwiceXCounters(t *testing.T) {
	g := newCatalogGame(t)
	castXSpell(t, g, "Primo, the Unbounded", "Legendary Creature — Fractal Wolf", socB10PrimoOracle, "{X}{G}{G}{U}", 3, nil)
	passPriorityAroundTable(t, g)
	ids := socB10Named(g, "Primo, the Unbounded")
	if len(ids) != 1 {
		t.Fatal("Primo did not enter")
	}
	if n := countersOn(g, ids[0], game.CounterPlusOne); n != 6 {
		t.Errorf("Primo has %d counters, want twice X = 6", n)
	}
	if !hasEffectiveKeyword(t, g, ids[0], "trample") {
		t.Error("Primo lacks trample")
	}
}

func TestSocB10PrimoMakesAFractalWithTheDamageDealtByBasePowerZeroCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	primo := b12Push(g, me.ID, "Primo, the Unbounded", "Legendary Creature — Fractal Wolf", socB10PrimoOracle, 0, 0)
	other := b12Creature(g, me.ID, "Zero Wall", "Creature — Wall", 0, 4)
	big := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{primo, other} {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2}
				}
			}
		})
	}
	life := opp.Life
	declareAttack(t, g, opp.ID, primo, other, big)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-(2+2+2) {
		t.Fatalf("opponent took %d, want 6", life-opp.Life)
	}
	fractals := socB10Named(g, "Fractal")
	if len(fractals) != 1 {
		t.Fatalf("%d Fractals, want one for the batch", len(fractals))
	}
	// Primo and the Wall (4 damage together); the 2/2 Bear has base power 2.
	if n := countersOn(g, fractals[0], game.CounterPlusOne); n != 4 {
		t.Errorf("the Fractal has %d counters, want the 4 damage Primo and the Wall dealt", n)
	}
	if p := effectivePowerWithCounters(t, g, fractals[0]); p != 4 {
		t.Errorf("the Fractal has power %d, want 4", p)
	}
}

func TestSocB10PrimoIgnoresCreaturesWithBasePowerAboveZero(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Primo, the Unbounded", "Legendary Creature — Fractal Wolf", socB10PrimoOracle, 0, 0)
	big := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, big)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if n := len(socB10Named(g, "Fractal")); n != 0 {
		t.Errorf("%d Fractals, want none: a 2/2 does not have base power 0", n)
	}
}

// --- Prismari Pianist ---

func TestSocB10PrismariPianistMakesOneElementalOrThreeForABigSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	b12Push(g, me.ID, "Prismari Pianist", "Creature — Djinn Bard", socB10PianistOracle, 2, 1)

	small := handCardFull(me, "Cheap Instant", "Instant", "{1}{R}", "", []string{"R"})
	b06AddMana(me, "C", "R")
	if err := g.CastSpell(me.ID, small, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(socB10Named(g, "Elemental")); n != 1 {
		t.Fatalf("%d Elementals after a mana value 2 instant, want 1", n)
	}

	big := handCardFull(me, "Big Sorcery", "Sorcery", "{4}{R}", "", []string{"R"})
	b06AddMana(me, "C", "C", "C", "C", "R")
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the sorcery: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(socB10Named(g, "Elemental")); n != 4 {
		t.Fatalf("%d Elementals after a mana value 5 sorcery, want 1 + 3", n)
	}
	for _, id := range socB10Named(g, "Elemental") {
		c, _ := g.LookupCardForEffect(id)
		if eff := c.Effective(); eff.Power != 1 || eff.Toughness != 1 || len(eff.Colors) != 2 {
			t.Errorf("an Elemental is %d/%d with colours %v, want a 1/1 blue and red", eff.Power, eff.Toughness, eff.Colors)
		}
	}

	// A creature spell does not trigger it.
	bear := handCardFull(me, "Bear Spell", "Creature — Bear", "{1}{R}", "", []string{"R"})
	b06AddMana(me, "C", "R")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(socB10Named(g, "Elemental")); n != 4 {
		t.Errorf("%d Elementals after a creature spell, want still 4", n)
	}
}

// --- Quandrix Apprentice ---

func TestSocB10QuandrixApprenticeTakesALandAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	b12Push(g, me.ID, "Quandrix Apprentice", "Creature — Human Wizard", socB10ApprenticeOracl, 2, 2)
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	ids := seedLibrary(me, "A", "B")
	hand := me.Hand.Size()

	spell := handCardFull(me, "Cheap Instant", "Instant", "{U}", "", []string{"U"})
	b06AddMana(me, "U")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	// The trigger asks whether to use the ability (optional "you may").
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no take-one prompt: %+v", g.PendingChoices)
	}
	for _, c := range pick.ChooseCards {
		if c == ids[0] || c == ids[1] {
			t.Errorf("a nonland card %v is offered", c)
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("take the land: %v", err)
	}
	if !me.Hand.Contains(land) {
		t.Fatal("the land is not in hand")
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom {
		t.Fatalf("the rest are not ordered on the bottom: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[1], ids[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := libraryBottomIDs(me, 2); !sameIDs(got, []uuid.UUID{ids[1], ids[0]}) {
		t.Errorf("bottom = %v, want the player's order", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: the land joined it", me.Hand.Size(), hand+1)
	}
}

// --- Plargg and Nassari ---

// latestChooseCardsChooser is who owes the open choose-cards prompt.
func latestChooseCardsChooser(g *game.Game) uuid.UUID {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			return c.Chooser
		}
	}
	return uuid.Nil
}

// socB10Library tops the player's library with cards, first named on top.
func socB10Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

func TestSocB10PlarggExilesUntilNonlandAnOpponentKeepsOneAndYouCastTwoOfTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2, o3 := g.Seats[1], g.Seats[0], g.Seats[2], g.Seats[3]
	b12Push(g, me.ID, "Plargg and Nassari", "Legendary Creature — Orc Efreet", socB10PlarggOracle, 5, 4)
	land := game.Card{Name: "Island", TypeLine: "Basic Land — Island"}
	mine := socB10Library(me, land, land, game.Card{Name: "My Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}"})
	first := socB10Library(o1, game.Card{Name: "First Spell", TypeLine: "Instant", ManaCost: "{U}"})
	second := socB10Library(o2, land, game.Card{Name: "Second Spell", TypeLine: "Sorcery", ManaCost: "{4}"})
	third := socB10Library(o3, game.Card{Name: "Third Spell", TypeLine: "Instant", ManaCost: "{B}"})
	libBefore := me.Library.Size()

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	// Several opponents: the controller names the one who chooses.
	answerOptionPick(t, g, me.ID, 1)
	chooserID := latestChooseCardsChooser(g)
	if chooserID == uuid.Nil {
		t.Fatalf("no choose-cards prompt for an opponent: %+v", g.PendingChoices)
	}
	if chooserID == me.ID {
		t.Fatal("the controller was asked to choose, want an opponent")
	}
	pick := chooseCardsChoiceFor(g, chooserID)
	if pick == nil || len(pick.ChooseCards) != 4 {
		t.Fatalf("the choice should name the four nonland cards: %+v", pick)
	}
	if got := libBefore - me.Library.Size(); got != 3 {
		t.Errorf("Plargg's controller exiled %d cards, want 3 (two lands and a spell)", got)
	}
	// The opponent keeps the controller's own spell out of reach.
	if err := g.ResolveChooseCards(pick.ID, chooserID, []uuid.UUID{mine[2]}); err != nil {
		t.Fatalf("choose a card: %v", err)
	}
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, first[0], second[1], third[0])
	if slices.Contains(offer.DiscardOptions, mine[2]) {
		t.Error("the chosen card is offered for casting")
	}
	for _, l := range []uuid.UUID{mine[0], mine[1], second[0]} {
		if slices.Contains(offer.DiscardOptions, l) {
			t.Error("a land is offered for casting")
		}
	}
	if err := castThroughResolutionCast(t, g, me.ID, first[0], game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the first spell free: %v", err)
	}
	if err := castThroughResolutionCast(t, g, me.ID, third[0], game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the second spell free: %v", err)
	}
	// Two casts were spent: the offer closes with the third card unused.
	if resolutionCastOfferFor(g, me.ID) != nil {
		if err := castThroughResolutionCast(t, g, me.ID, second[1], game.CastSpellParams{Strict: true}); err == nil {
			t.Error("a third spell was cast")
		}
		declineResolutionCast(t, g, me.ID)
	}
	if !g.Exile.Contains(second[1]) || !g.Exile.Contains(mine[2]) {
		t.Error("the uncast and the chosen card should stay in exile")
	}
}

func TestSocB10PlarggWithOneOpponentAsksThemWithoutNamingAndDeclineLeavesCardsInExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	b12Push(g, me.ID, "Plargg and Nassari", "Legendary Creature — Orc Efreet", socB10PlarggOracle, 5, 4)
	// Eliminate two seats so one opponent is left.
	g.WithWriteLock(func() {
		g.Seats[2].Eliminated = true
		g.Seats[3].Eliminated = true
	})
	opp := g.Seats[0]
	mine := socB10Library(me, game.Card{Name: "My Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}"})
	theirs := socB10Library(opp, game.Card{Name: "Their Spell", TypeLine: "Instant", ManaCost: "{U}"})
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the lone opponent was not asked to choose: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{theirs[0]}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, mine[0])
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(mine[0]) || !g.Exile.Contains(theirs[0]) {
		t.Error("the declined cards should remain in exile")
	}
}

// --- the Campuses ---

func socB10CampusRoundTrip(t *testing.T, name, oracle string, colours [2]string) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")
	id := playLandFromHand(t, g, name, oracle)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	untapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the land made {B}")
	}
	for _, colour := range colours {
		untapForTest(g, id)
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("{T}: Add {%s}: %v", colour, err)
		}
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	}

	untapForTest(g, id)
	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("scry activated with three mana")
	}
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{4},{T}: scry 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("scry looked at %d cards, want 1", len(c.ScryCards))
	}
	top := c.ScryCards[0]
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := libraryBottomIDs(me, 1); !sameIDs(got, []uuid.UUID{top}) {
		t.Error("the scried card did not go to the bottom")
	}
	if card, _ := battlefieldCard(g, id); !card.Tapped {
		t.Error("the {T} cost did not tap the land")
	}
}

func TestSocB10PrismariCampusEntersTappedMakesUOrRAndScrysForFour(t *testing.T) {
	socB10CampusRoundTrip(t, "Prismari Campus", socB10PrismariCampus, [2]string{"U", "R"})
}

func TestSocB10QuandrixCampusEntersTappedMakesGOrUAndScrysForFour(t *testing.T) {
	socB10CampusRoundTrip(t, "Quandrix Campus", socB10QuandrixCampus, [2]string{"G", "U"})
}
