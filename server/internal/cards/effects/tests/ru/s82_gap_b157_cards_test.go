package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b157_cards_test.go — GAP batch 157 (#5948): seven Jumpstart 2022
// red creatures that were not in the catalog. Each test plays the real
// catalog entry through the engine and asserts what its text says.

const (
	gapB157ChannelerOracle  = "382267da-c63c-432e-8a11-f803cf0e3a3a"
	gapB157HeartPiercerOrcl = "e4b15cba-9cd7-422d-9ca6-39243bf67dca"
	gapB157TorchrunnerOracl = "045b2b84-96bd-46fb-8aae-a33968e741c7"
	gapB157MuxusOracle      = "07c82cde-f00d-4454-9df4-9286e344c9bb"
	gapB157PredatorOracle   = "dfd260f2-90e6-464b-91e2-a66bb3877918"
	gapB157MarmosetOracle   = "a1ef3b31-b996-48ea-967a-2b56fb2fb22b"
	gapB157CyclopsOracle    = "cbb6139f-827f-490b-b9b7-e8c3ef714d44"
)

func TestGapB157EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB157ChannelerOracle:  "Magmatic Channeler",
		gapB157HeartPiercerOrcl: "Mardu Heart-Piercer",
		gapB157TorchrunnerOracl: "Mudbutton Torchrunner",
		gapB157MuxusOracle:      "Muxus, Goblin Grandee",
		gapB157PredatorOracle:   "Plundering Predator",
		gapB157MarmosetOracle:   "Prickly Marmoset",
		gapB157CyclopsOracle:    "Quakefoot Cyclops",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
		}
	}
}

func TestGapB157MarduHeartPiercerDealsTwoOnlyAfterAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AvB16CastCreature(t, g, "Mardu Heart-Piercer", "Creature — Human Archer", gapB157HeartPiercerOrcl, 2, 3)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the raid trigger fired without an attack this turn")
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.AvB16CastCreature(t, g, "Mardu Heart-Piercer", "Creature — Human Archer", gapB157HeartPiercerOrcl, 2, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived 2 damage")
	}
}

func TestGapB157MudbuttonTorchrunnerDealsThreeWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := effectstest.B12Push(g, me.ID, "Mudbutton Torchrunner", "Creature — Goblin Warrior", gapB157TorchrunnerOracl, 1, 1)
	life := opp.Life
	effectstest.B25Destroy(g, runner)
	effectstest.Vow1Pick(t, g, me, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, opp.Life)
	}
}

func TestGapB157PlunderingPredatorFliesAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastHoldCreature(t, g, "Plundering Predator", "Creature — Dragon", gapB157PredatorOracle, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("the Dragon has no flying")
	}
	hand, graves := me.Hand.Size(), me.Graveyard.Size()
	pitch := me.Hand.Cards[0].InstanceID
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d (discard one, draw one)", got, hand)
	}
	if got := me.Graveyard.Size(); got != graves+1 || !me.Graveyard.Contains(pitch) {
		t.Errorf("graveyard %d, want %d with the discarded card", got, graves+1)
	}
}

func TestGapB157PricklyMarmosetGetsTwoPowerPerCycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mar := effectstest.B12Push(g, me.ID, "Prickly Marmoset", "Creature — Monkey", gapB157MarmosetOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mar, "first strike") {
		t.Error("the Marmoset has no first strike")
	}
	effectstest.CycleFromHand(t, g, "Scrounging Skyray", "Creature — Bird", effectstest.ScroungingSkyrayOracle, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mar); got != 4 {
		t.Errorf("power %d after one cycle, want 4", got)
	}
}

func TestGapB157QuakefootCyclopsEntersAndCyclesForCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Quakefoot Cyclops", "Creature — Cyclops", gapB157CyclopsOracle, 5, 4)
	effectstest.Vow1Pick(t, g, me, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	cantBlock := func(id uuid.UUID) bool {
		card, _ := g.LookupCardForEffect(id)
		return card.Effective().Restrictions&game.CantBlock != 0
	}
	if !cantBlock(a) || !cantBlock(b) || cantBlock(c) {
		t.Errorf("can't block: a=%v b=%v c=%v, want the two targets only", cantBlock(a), cantBlock(b), cantBlock(c))
	}
}

func TestGapB157QuakefootCyclopsCyclingTargetsOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id, me := effectstest.CycleFromHand(t, g, "Quakefoot Cyclops", "Creature — Cyclops", gapB157CyclopsOracle, "{C}{R}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Cyclops is not in the graveyard")
	}
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	card, _ := g.LookupCardForEffect(bear)
	if card.Effective().Restrictions&game.CantBlock == 0 {
		t.Error("the cycle trigger did not stop the creature blocking")
	}
}

func TestGapB157QuakefootCyclopsCyclingIsRefusedWithoutMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Quakefoot Cyclops", "Creature — Cyclops", gapB157CyclopsOracle)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{1}{R} must be paid to cycle")
	}
	if !me.Hand.Contains(id) {
		t.Error("the Cyclops was discarded despite the unpaid cost")
	}
}

// gapB157Library seeds a card on top of a library.
func gapB157Library(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

func TestGapB157MuxusPutsSmallGoblinsOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	seventh := gapB157Library(me, "Seventh Card", "Instant", "{R}")
	sixth := gapB157Library(me, "Sixth Card", "Instant", "{R}")
	elf := gapB157Library(me, "Plain Elf", "Creature — Elf", "{G}")
	big := gapB157Library(me, "Big Goblin", "Creature — Goblin", "{4}{R}{R}")
	small := gapB157Library(me, "Small Goblin", "Creature — Goblin", "{1}{R}")
	five := gapB157Library(me, "Five Goblin", "Creature — Goblin", "{3}{R}{R}")
	spell := gapB157Library(me, "Goblin Spell", "Sorcery — Goblin", "{R}")
	effectstest.CastHoldCreature(t, g, "Muxus, Goblin Grandee", "Legendary Creature — Goblin Noble", gapB157MuxusOracle, 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) || !g.Battlefield.Contains(five) {
		t.Error("a Goblin creature card with mana value 5 or less stayed in the library")
	}
	for name, id := range map[string]uuid.UUID{"mana value 6 Goblin": big, "non-Goblin": elf, "non-creature Goblin": spell, "sixth card": sixth} {
		if g.Battlefield.Contains(id) || !me.Library.Contains(id) {
			t.Errorf("the %s should be in the library, not the battlefield", name)
		}
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != seventh {
		t.Error("the seventh card should be the top of the library now")
	}
	// The four revealed cards that were not put went to the bottom.
	bottom := me.Library.Cards[:4]
	for _, c := range bottom {
		if c.InstanceID != big && c.InstanceID != elf && c.InstanceID != spell && c.InstanceID != sixth {
			t.Errorf("bottom of the library holds %s, want only the unpicked revealed cards", c.Name)
		}
	}
}

func TestGapB157MuxusGetsPlusOnePlusOnePerOtherGoblinWhenAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	muxus := effectstest.B12Push(g, me.ID, "Muxus, Goblin Grandee", "Legendary Creature — Goblin Noble", gapB157MuxusOracle, 6, 6)
	effectstest.B12Creature(g, me.ID, "Goblin One", "Creature — Goblin", 1, 1)
	effectstest.B12Creature(g, me.ID, "Goblin Two", "Creature — Goblin Warrior", 1, 1)
	effectstest.B12Creature(g, me.ID, "Not A Goblin", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, muxus)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, muxus); got != 8 {
		t.Errorf("power %d, want 8 (6 + two other Goblins)", got)
	}
}

func TestGapB157MagmaticChannelerNeedsFourInstantsOrSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ch := effectstest.B12Push(g, me.ID, "Magmatic Channeler", "Creature — Human Wizard", gapB157ChannelerOracle, 2, 2)
	for i := 0; i < 3; i++ {
		gapB157CastBlank(t, g, me, "Instant")
	}
	effectstest.KillCreature(t, g, me.ID, effectstest.B12Creature(g, me.ID, "Dead Bear", "Creature — Bear", 2, 2))
	if got := effectstest.EffectivePower(t, g, ch); got != 2 {
		t.Fatalf("power %d with three instants and a creature, want 2", got)
	}
	gapB157CastBlank(t, g, me, "Sorcery")
	if got, tough := effectstest.EffectivePower(t, g, ch), effectstest.EffectiveToughness(t, g, ch); got != 5 || tough != 3 {
		t.Errorf("%d/%d with four instants and sorceries, want 5/3", got, tough)
	}
}

// gapB157CastBlank casts and resolves an uncatalogued spell with the given
// type line, which leaves it in its owner's graveyard.
func gapB157CastBlank(t *testing.T, g *game.Game, p *game.Player, typeLine string) {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(p, "Blank "+typeLine, typeLine, "", "", nil)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", typeLine, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !p.Graveyard.Contains(id) {
		t.Fatalf("the %s is not in the graveyard", typeLine)
	}
}

func TestGapB157MagmaticChannelerExilesTwoAndPlaysTheChosenOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ch := effectstest.B12Push(g, me.ID, "Magmatic Channeler", "Creature — Human Wizard", gapB157ChannelerOracle, 2, 2)
	third := gapB157Library(me, "Third", "Basic Land — Swamp", "")
	second := gapB157Library(me, "Second", "Basic Land — Swamp", "")
	first := gapB157Library(me, "First", "Basic Land — Swamp", "")
	pitch := me.Hand.Cards[0].InstanceID
	if err := g.ActivateCatalogAbility(me.ID, ch, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discard was not paid as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("pick %+v, want a choice between the two exiled cards", pick)
	}
	if z := effectstest.ZoneOfCard(g, third); z != game.ZoneLibrary {
		t.Errorf("a third card left the library (%v)", z)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("played the card that was not chosen")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("could not play the chosen card: %v", err)
	}
}
