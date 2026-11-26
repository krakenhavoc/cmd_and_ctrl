package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 Dragons of Tarkir batch 331 (#5551): dash and megamorph.

const (
	dtk331Warbringer  = "0636090b-49e0-45b4-a68d-48b1cdc0c54d"
	dtk331Forerunners = "07b6fbd3-03f9-4aef-929e-514c799d75b9"
	dtk331Zurgo       = "46808926-d95b-4313-89d1-6bdfe13fcb77"
	dtk331Ambuscade   = "96e1d890-4ebb-4c79-8df9-24e3adaf4d9b"
	dtk331Warbrute    = "e680f3e9-5c76-41e6-a4a5-662ee05ac3af"
	dtk331Stormsinger = "0790c8bd-a094-49b5-8843-e8b02cba145b"
	dtk331Loremaster  = "1c8ef715-3c3c-4d04-9f30-941b870b11d3"
	dtk331Shieldbear  = "2bd8a6a4-d0ac-4da3-be1d-5f404ae4b614"
	dtk331Shorecrash  = "5eb0d8dc-be4b-4a20-b480-f70514ffb299"
	dtk331IreShaman   = "7f691e0d-dac5-4d7e-8727-dd2cab9a7863"
)

func TestDtk331CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dtk331Warbringer: "Warbringer", dtk331Forerunners: "Kolaghan Forerunners", dtk331Zurgo: "Zurgo Bellstriker",
		dtk331Ambuscade: "Ambuscade Shaman", dtk331Warbrute: "Sprinting Warbrute", dtk331Stormsinger: "Kolaghan Stormsinger",
		dtk331Loremaster: "Monastery Loremaster", dtk331Shieldbear: "Guardian Shield-Bearer",
		dtk331Shorecrash: "Shorecrasher Elemental", dtk331IreShaman: "Ire Shaman",
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

// dtk331HandWarbrute puts a Sprinting Warbrute ({4}{R}, dash {3}{R}) in hand.
func dtk331HandWarbrute(me *game.Player) uuid.UUID {
	return effectstest.HandCardFull(me, "Sprinting Warbrute", "Creature — Ogre Berserker", "{4}{R}", dtk331Warbrute, []string{"R"})
}

func TestDtk331WarbringerReducesOnlyDashCosts(t *testing.T) {
	cast := func(t *testing.T, withWarbringer bool, alt, pool string) error {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.ToMainForCost(t, g)
		if withWarbringer {
			effectstest.B12Push(g, me.ID, "Warbringer", "Creature — Orc Berserker", dtk331Warbringer, 3, 3)
		}
		id := dtk331HandWarbrute(me)
		effectstest.FloatForTest(g, me, pool)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: alt})
	}
	if err := cast(t, true, "dash", "RC"); err != nil {
		t.Errorf("a {3}{R} dash cost with Warbringer out is {1}{R}, but {R}{C} was refused: %v", err)
	}
	if err := cast(t, false, "dash", "RC"); err == nil {
		t.Error("without Warbringer the {3}{R} dash cost was paid with two mana")
	}
	if err := cast(t, true, "dash", "R"); err == nil {
		t.Error("with Warbringer the dash cost lost a coloured symbol too")
	}
	if err := cast(t, true, "", "RCCC"); err == nil {
		t.Error("Warbringer reduced the {4}{R} mana cost of a creature cast the normal way")
	}
	if err := cast(t, true, "", "RCCCC"); err != nil {
		t.Errorf("a normal cast with the full {4}{R}: %v", err)
	}
}

func TestDtk331WarbringerReducesOnlyItsControllersDashCosts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMainForCost(t, g)
	effectstest.B12Push(g, opp.ID, "Warbringer", "Creature — Orc Berserker", dtk331Warbringer, 3, 3)
	id := dtk331HandWarbrute(me)
	effectstest.FloatForTest(g, me, "RC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "dash"}); err == nil {
		t.Error("an opponent's Warbringer reduced my dash cost")
	}
}

func TestDtk331WarbringerItselfCanBeDashedForTwoR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Warbringer", "Creature — Orc Berserker", "{3}{R}", dtk331Warbringer, []string{"R"})
	effectstest.FloatForTest(g, me, "RCC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "dash"}); err != nil {
		t.Fatalf("dash {2}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Error("the dashed Warbringer has no haste")
	}
}

func TestDtk331ForerunnersPowerIsTheNumberOfCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	runner := effectstest.B12Push(g, me.ID, "Kolaghan Forerunners", "Creature — Human Berserker", dtk331Forerunners, 0, 3)
	if p, tough := effectstest.EffectivePTOf(t, g, runner); p != 1 || tough != 3 {
		t.Errorf("alone it is %d/%d, want 1/3", p, tough)
	}
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, runner); p != 3 {
		t.Errorf("with two other creatures of mine its power is %d, want 3 (an opponent's creature does not count)", p)
	}
	if !effectstest.HasKeywordOf(g, runner, "trample") {
		t.Error("Kolaghan Forerunners lacks trample")
	}
}

func TestDtk331ZurgoCantBlockCreaturesWithPowerTwoOrGreater(t *testing.T) {
	for _, tc := range []struct {
		power   int
		refused bool
	}{{1, false}, {2, true}, {3, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", tc.power, 4)
		zurgo := effectstest.B12Push(g, opp.ID, "Zurgo Bellstriker", "Legendary Creature — Orc Warrior", dtk331Zurgo, 2, 2)
		effectstest.AdvanceToMain(t, g)
		refused := effectstest.Msc13BlockRefused(t, g, opp.ID, zurgo, attacker)
		if refused[attacker] != tc.refused {
			t.Errorf("power %d: block refused = %v, want %v", tc.power, refused[attacker], tc.refused)
		}
	}
}

func TestDtk331AmbuscadeShamanPumpsItselfAndOtherCreaturesYouControlOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	shaman := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Ambuscade Shaman", TypeLine: "Creature — Orc Shaman", OracleID: dtk331Ambuscade, Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, shaman); p != 4 || tough != 4 {
		t.Errorf("the Shaman itself is %d/%d after entering, want 4/4", p, tough)
	}
	bear := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("my entering creature is %d/%d, want 4/4", p, tough)
	}
	theirs := effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's entering creature is %d/%d, want 2/2", p, tough)
	}
	land := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, land); p != 0 {
		t.Errorf("a land got power %d", p)
	}
}

func TestDtk331SprintingWarbruteAttacksEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	brute := effectstest.B12Push(g, me.ID, "Sprinting Warbrute", "Creature — Ogre Berserker", dtk331Warbrute, 5, 4)
	effectstest.Unsick(g, brute)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != brute {
		t.Fatalf("pass with the Warbrute home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(brute, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

// dtk331FaceDown casts the card from hand face down for its megamorph cost.
func dtk331FaceDown(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	return effectstest.S36CastFaceDown(t, g, name, typeLine, oracle, "megamorph", p, tough)
}

func TestDtk331KolaghanStormsingerGivesATargetCreatureHasteWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := dtk331FaceDown(t, g, "Kolaghan Stormsinger", "Creature — Human Shaman", dtk331Stormsinger, 1, 1)
	if effectstest.HasKeywordOf(g, bear, "haste") {
		t.Fatal("the Bear already has haste")
	}
	effectstest.S36TurnUp(t, g, me, id, "R")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, "haste") {
		t.Error("the targeted creature did not gain haste")
	}
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the megamorphed creature, want 1", got)
	}
}

func TestDtk331MonasteryLoremasterReturnsANoncreatureNonlandCardWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	put := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return id
	}
	sorcery := put("Old Sorcery", "Sorcery")
	creature := put("Old Bear", "Creature — Bear")
	land := put("Old Forest", "Basic Land — Forest")
	id := dtk331FaceDown(t, g, "Monastery Loremaster", "Creature — Djinn Wizard", dtk331Loremaster, 3, 2)
	effectstest.S36TurnUp(t, g, me, id, "U", "C", "C", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, sorcery)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(sorcery) {
		t.Error("the noncreature, nonland card did not return to hand")
	}
	if me.Hand.Contains(creature) || me.Hand.Contains(land) {
		t.Error("a creature or land card returned to hand")
	}
}

func TestDtk331GuardianShieldBearerPutsACounterOnAnotherCreatureWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	id := dtk331FaceDown(t, g, "Guardian Shield-Bearer", "Creature — Human Soldier", dtk331Shieldbear, 2, 1)
	effectstest.S36TurnUp(t, g, me, id, "G", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature has %d counters, want 0", got)
	}
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("the Shield-Bearer has %d counters, want only the one megamorph put on it", got)
	}
}

func TestDtk331IreShamanExilesTheTopCardWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := effectstest.PushLibraryCard(me, game.Card{Name: "Top Card", TypeLine: "Sorcery"})
	id := dtk331FaceDown(t, g, "Ire Shaman", "Creature — Orc Shaman", dtk331IreShaman, 2, 1)
	if !me.Library.Contains(top) {
		t.Fatal("setup: the card left the library before the Shaman turned up")
	}
	effectstest.S36TurnUp(t, g, me, id, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Contains(top) || !g.Exile.Contains(top) {
		t.Error("the top card of the library was not exiled")
	}
}

func dtk331FaceDownPermanent(g *game.Game, owner uuid.UUID) (uuid.UUID, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.FaceDownIsPermanent() {
			return c.InstanceID, true
		}
	}
	return uuid.Nil, false
}

func TestDtk331ShorecrasherElementalReturnsItselfFaceDownForU(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	elemental := effectstest.B12Push(g, me.ID, "Shorecrasher Elemental", "Creature — Elemental", dtk331Shorecrash, 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, elemental, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the blink activated without {U}")
	}
	effectstest.FloatForTest(g, me, "U")
	if err := g.ActivateCatalogAbility(me.ID, elemental, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elemental) {
		t.Error("the original object is still on the battlefield")
	}
	down, ok := dtk331FaceDownPermanent(g, me.ID)
	if !ok {
		t.Fatal("no face-down permanent came back")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, down); p != 2 || tough != 2 {
		t.Errorf("the face-down creature is %d/%d, want 2/2", p, tough)
	}
	// It prints megamorph, so it can be turned face up again for {4}{U}.
	effectstest.S36TurnUp(t, g, me, down, "U", "C", "C", "C", "C")
	c := effectstest.FaceDownCardForTest(t, g, down)
	if c.FaceDownIsPermanent() {
		t.Error("the Elemental did not turn face up for its megamorph cost")
	}
}

func TestDtk331ShorecrasherElementalPumpsEitherWay(t *testing.T) {
	for _, tc := range []struct {
		yes     bool
		p, tou  int
		heading string
	}{{true, 4, 2, "+1/-1"}, {false, 2, 4, "-1/+1"}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		elemental := effectstest.B12Push(g, me.ID, "Shorecrasher Elemental", "Creature — Elemental", dtk331Shorecrash, 3, 3)
		effectstest.FloatForTest(g, me, "C")
		if err := g.ActivateCatalogAbility(me.ID, elemental, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerConfirmFor(t, g, me, tc.yes)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.EffectivePTOf(t, g, elemental); p != tc.p || tough != tc.tou {
			t.Errorf("%s: %d/%d, want %d/%d", tc.heading, p, tough, tc.p, tc.tou)
		}
	}
}
