package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b206_test.go — cards missing from sets since AFR, batch 206 (#5948):
// Double Masters 2022 creatures and enchantments. Each test plays the real
// catalog entry and asserts what its text says.

const (
	gap206Abzan    = "3b7d4206-0d88-407b-ace5-d5d28881bf95"
	gap206Splicer  = "9676da93-1e17-4f6c-b610-a242b78c71ab"
	gap206Augustin = "1f8d4d5f-e82f-45f3-823e-1bb6b536eb18"
	gap206Witch    = "38e274e2-bd04-48de-a1df-44f0ee987ba8"
	gap206Tyranny  = "55082cd0-a33c-4f1d-9f46-9578f63e6bb6"
	gap206Strands  = "897c2c66-fedc-4201-b187-db79414583fb"
	gap206Seeker   = "e206842f-139b-490d-bc10-d365705ac720"
)

func TestGap206EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap206Abzan:    "Abzan Ascendancy",
		gap206Splicer:  "Sensor Splicer",
		gap206Augustin: "Grand Arbiter Augustin IV",
		gap206Witch:    "Mistmeadow Witch",
		gap206Tyranny:  "Phyrexian Tyranny",
		gap206Strands:  "Strands of Undeath",
		gap206Seeker:   "Seeker of the Way",
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

// --- Abzan Ascendancy --------------------------------------------------------

func TestGap206AbzanCountersOnEntryAndSpiritsWhenANontokenCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.Ecl6Cast(t, g, me, "Abzan Ascendancy", "Enchantment", gap206Abzan, 0, 0, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, mine); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("my creature has %d counters, want 1", c.Counters[game.CounterPlusOne])
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Counters[game.CounterPlusOne] != 0 {
		t.Error("an opponent's creature got a counter")
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Spirit", me.ID); n != 0 {
		t.Fatalf("%d Spirits after an opponent's creature died", n)
	}
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Spirit", me.ID); n != 1 {
		t.Fatalf("%d Spirits after my nontoken creature died, want 1", n)
	}
	// A token dying makes none.
	var spirit uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" {
			spirit = c.InstanceID
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, spirit, "flying") {
		t.Error("the Spirit token has no flying")
	}
	effectstest.KillCreature(t, g, me.ID, spirit)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Spirit", me.ID); n != 0 {
		t.Errorf("%d Spirits after a token died, want none more", n)
	}
}

// --- Sensor Splicer ----------------------------------------------------------

func TestGap206SensorSplicerMakesAVigilantGolemOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	sp := effectstest.Ecl6Cast(t, g, me, "Sensor Splicer", "Creature — Phyrexian Artificer", gap206Splicer, 1, 1, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var golem uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Golem" {
			golem = c.InstanceID
		}
	}
	if golem == uuid.Nil {
		t.Fatal("no Golem token")
	}
	if p, tt := effectstest.PtOf(t, g, golem); p != 3 || tt != 3 {
		t.Errorf("the Golem is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, golem, "vigilance") {
		t.Error("the Golem has no vigilance")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") || effectstest.HasEffectiveKeyword(t, g, sp, "vigilance") {
		t.Error("a non-Golem got vigilance")
	}
}

// --- Grand Arbiter Augustin IV -----------------------------------------------

func TestGap206AugustinDiscountsWhiteAndBlueAndTaxesOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Grand Arbiter Augustin IV", "Legendary Creature — Human Advisor", gap206Augustin, false)
	if got := effectstest.PriceInHand(t, g, me, "White Thing", "Creature — Human", "{2}{W}"); got != 2 {
		t.Errorf("a white spell costs %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Blue Thing", "Instant", "{1}{U}"); got != 1 {
		t.Errorf("a blue spell costs %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Azorius Thing", "Instant", "{2}{W}{U}"); got != 2 {
		t.Errorf("a white-blue spell costs %d, want both discounts (2)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Red Thing", "Instant", "{1}{R}"); got != 2 {
		t.Errorf("a red spell costs %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Their Red Thing", "Instant", "{1}{R}"); got != 3 {
		t.Errorf("an opponent's spell costs %d, want 3", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Their White Thing", "Instant", "{1}{W}"); got != 3 {
		t.Errorf("an opponent's white spell costs %d, want 3 (the discount is only yours)", got)
	}
}

// --- Mistmeadow Witch --------------------------------------------------------

func TestGap206WitchExilesAnyCreatureAndReturnsItAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	w := effectstest.PushCatalogPermanent(g, me.ID, "Mistmeadow Witch", "Creature — Kithkin Wizard", gap206Witch, false)
	foe := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.B06AddMana(me, "C", "C", "W", "U")
	effectstest.B16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(foe)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Wolf", opp.ID); n != 0 {
		t.Fatalf("%d Wolves still on the battlefield, want it exiled", n)
	}
	effectstest.Ecl6NextTurn(t, g)
	if n := effectstest.CountOnBattlefield(g, "Wolf", opp.ID); n != 1 {
		t.Errorf("%d Wolves back under its owner's control, want 1", n)
	}
}

// --- Seeker of the Way -------------------------------------------------------

func TestGap206SeekerGainsLifelinkOnlyOnANoncreatureCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	s := effectstest.PushCatalogPermanent(g, me.ID, "Seeker of the Way", "Creature — Human Warrior", gap206Seeker, false)
	if effectstest.HasEffectiveKeyword(t, g, s, "lifelink") {
		t.Fatal("lifelink before any spell")
	}
	base := effectstest.EffectivePower(t, g, s)
	effectstest.CastCatalogSpell(t, g, "Filler", "Instant", "", nil)
	effectstest.SettleProwess(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, s, "lifelink") {
		t.Error("no lifelink after a noncreature spell")
	}
	if p := effectstest.EffectivePower(t, g, s); p != base+1 {
		t.Errorf("power %d, want prowess's +1", p)
	}
}

// --- Phyrexian Tyranny -------------------------------------------------------

func TestGap206TyrannyMakesAnyDrawerPayTwoOrLoseTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Tyranny", "Enchantment", gap206Tyranny, false)
	for _, who := range []*game.Player{me, opp} {
		life := who.Life
		if err := g.DrawCard(who.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, who.ID) {
			t.Fatalf("%s was not asked to pay", who.ID)
		}
		effectstest.AnswerPayUnless(t, g, who.ID, false)
		if who.Life != life-2 {
			t.Errorf("a drawer who declined is at %d, want %d", who.Life, life-2)
		}
	}
	// Paying keeps the life.
	life := opp.Life
	effectstest.B06AddMana(opp, "C", "C")
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	if opp.Life != life {
		t.Errorf("a drawer who paid is at %d, want %d", opp.Life, life)
	}
}

// --- Strands of Undeath ------------------------------------------------------

func TestGap206StrandsDiscardsTwoAndRegeneratesTheHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	a := effectstest.PutInHand(opp, game.Card{InstanceID: uuid.New(), Name: "A", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	b := effectstest.PutInHand(opp, game.Card{InstanceID: uuid.New(), Name: "B", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	aura := effectstest.Ecl6Cast(t, g, me, "Strands of Undeath", "Enchantment — Aura", gap206Strands, 0, 0,
		game.CastSpellParams{Targets: effectstest.B16TargetCard(bear)})
	effectstest.Ecl6ResolveSpell(t, g)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, opp.ID) != 2 {
		t.Fatalf("the opponent owes %d discards, want 2", effectstest.DiscardOwed(g, opp.ID))
	}
	effectstest.AnswerDiscard(t, g, opp.ID, a, b)
	if effectstest.InHandOf(opp, a) || effectstest.InHandOf(opp, b) {
		t.Error("the discards did not happen")
	}
	effectstest.B06AddMana(me, "B")
	effectstest.B16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DestroyIt(t, g, bear)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the enchanted creature was destroyed despite the regeneration shield")
	}
}
