package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b847_cards_test.go — Ice Age, batch 847 (#5636): each test plays the
// real catalog entry through the real engine paths.

const (
	b847Malachite = "0be4877d-72ff-4951-aa52-f4f816abc844"
	b847Hematite  = "586601dd-ac73-4451-848d-f21e64feb6a9"
	b847Lapis     = "83447e57-06f7-4e7b-91d4-7b8f71ea0676"
	b847Staff     = "86af21ba-58ad-45e8-a316-5411c870e415"
	b847Golem     = "cd444267-712c-450f-a349-e27faf4f4c7c"
	b847Onyx      = "d19a19ed-33f9-400d-9669-c393d7127652"
	b847Nacre     = "f87aaed9-7f78-4f8b-8af5-44d4be1d6f30"
)

func b847Tapped(g *game.Game, id uuid.UUID) bool {
	c := effectstest.FindBattlefieldCardByID(g, id)
	return c != nil && c.Tapped
}

func TestS82OldB847CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b847Malachite: "Malachite Talisman", b847Hematite: "Hematite Talisman", b847Lapis: "Lapis Lazuli Talisman",
		b847Staff: "Staff of the Ages", b847Golem: "Soldevi Golem", b847Onyx: "Onyx Talisman", b847Nacre: "Nacre Talisman",
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

// b847CastSorcery casts a one-mana sorcery of the given color from my hand.
func b847CastSorcery(t *testing.T, g *game.Game, me *game.Player, color string) {
	t.Helper()
	c := game.NewCard(color+" Spell", me.ID)
	c.TypeLine, c.ManaCost, c.Colors = "Sorcery", "{"+color+"}", []string{color}
	me.Hand.PushTop(c)
	effectstest.B06AddMana(me, color)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the %s spell: %v", color, err)
	}
}

func TestS82OldB847TalismansUntapForTheirColorWhenPaid(t *testing.T) {
	for _, tc := range []struct {
		oracle, name, color, other string
	}{
		{b847Malachite, "Malachite Talisman", "G", "R"},
		{b847Hematite, "Hematite Talisman", "R", "G"},
		{b847Lapis, "Lapis Lazuli Talisman", "U", "R"},
		{b847Onyx, "Onyx Talisman", "B", "G"},
		{b847Nacre, "Nacre Talisman", "W", "R"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Push(g, me.ID, tc.name, "Artifact", tc.oracle, 0, 0)
			land := effectstest.BlbB1Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
			effectstest.AdvanceToMain(t, g)
			effectstest.TapCard(g, land)
			// A spell of another color triggers nothing.
			b847CastSorcery(t, g, me, tc.other)
			if effectstest.LatestPickTarget(g, me.ID) != nil {
				t.Fatalf("a %s spell triggered the %s", tc.other, tc.name)
			}
			effectstest.PassPriorityAroundTable(t, g)
			// Its own color, paid: the land untaps.
			b847CastSorcery(t, g, me, tc.color)
			effectstest.AvB34Pick(t, g, me.ID, land)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.B06AddMana(me, "C", "C", "C")
			effectstest.AnswerPayUnless(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			if b847Tapped(g, land) {
				t.Error("the land stayed tapped though {3} was paid")
			}
			// Declined: nothing untaps.
			effectstest.TapCard(g, land)
			b847CastSorcery(t, g, me, tc.color)
			effectstest.AvB34Pick(t, g, me.ID, land)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerPayUnless(t, g, me.ID, false)
			effectstest.PassPriorityAroundTable(t, g)
			if !b847Tapped(g, land) {
				t.Error("the land untapped though {3} was declined")
			}
		})
	}
}

func TestS82OldB847StaffOfTheAgesLetsLandwalkersBeBlocked(t *testing.T) {
	for _, staff := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		walker := effectstest.Pr7bCreature(g, me.ID, "Swampwalker", 2, "{1}", []string{"B"}, "swampwalk")
		effectstest.BlbB1Push(g, opp.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
		blocker := effectstest.B16Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2, "G")
		if staff {
			// Whoever controls the Staff: it waives the evasion for every attacker.
			effectstest.B12Push(g, opp.ID, "Staff of the Ages", "Artifact", b847Staff, 0, 0)
		}
		effectstest.AdvanceToMain(t, g)
		effectstest.DeclareAttack(t, g, opp.ID, walker)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(blocker, walker)
		if staff && err != nil {
			t.Errorf("with the Staff the swampwalker could not be blocked: %v", err)
		}
		if !staff && err == nil {
			t.Error("without the Staff a swampwalker was blocked by a player with a Swamp")
		}
		if staff && !effectstest.HasKeywordOnBattlefield(t, g, walker, "swampwalk") {
			t.Error("the Staff removed the keyword; it only waives the evasion")
		}
	}
}

func TestS82OldB847SoldeviGolemStaysTappedUntilItsUpkeepUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := effectstest.B12Push(g, me.ID, "Soldevi Golem", "Artifact Creature — Golem", b847Golem, 5, 3)
	tapped := effectstest.B16Creature(g, opp.ID, "Tapped Bear", "Creature — Bear", 2, 2, "G")
	mine := effectstest.B16Creature(g, me.ID, "My Tapped Bear", "Creature — Bear", 2, 2, "G")
	effectstest.TapCard(g, golem)
	effectstest.TapCard(g, mine)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.TapCard(g, tapped)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if !b847Tapped(g, golem) {
		t.Fatal("the Golem untapped in its controller's untap step")
	}
	if b847Tapped(g, mine) {
		t.Fatal("my own creature stayed tapped through my untap step")
	}
	// The trigger can only target an opponent's tapped creature: the one legal
	// target is chosen for me.
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AvB34Pick(t, g, me.ID, tapped)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !b847Tapped(g, tapped) || !b847Tapped(g, golem) {
		t.Fatal("declining untapped something")
	}
	// Next upkeep, accept.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.TapCard(g, tapped)
	effectstest.TapCard(g, mine)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AvB34Pick(t, g, me.ID, tapped)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if b847Tapped(g, tapped) {
		t.Error("the target creature stayed tapped")
	}
	if b847Tapped(g, golem) {
		t.Error("the Golem stayed tapped")
	}
	if b847Tapped(g, mine) {
		t.Error("my own creature was not untapped by my untap step")
	}
}
