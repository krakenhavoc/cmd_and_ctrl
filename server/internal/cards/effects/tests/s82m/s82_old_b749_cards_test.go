package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b749_cards_test.go — Stronghold, instants and sorceries and creatures
// with keywords, batch 749 (#5627, S82 OLD). Every test plays the real catalog
// entry through the engine.

const (
	oldB749Scapegoat   = "04eb4fc0-0cc5-44f1-af5a-44579125d5c2"
	oldB749Verdant     = "68845200-5346-473d-becd-27315879443d"
	oldB749ElvenRite   = "755dbf29-7484-444f-b5df-29ec4b4bbc82"
	oldB749Crossbow    = "c10fd7a6-360f-43d6-a4ef-d990c97cb2f6"
	oldB749Razors      = "19ed5429-5631-4a36-a69c-ab35860160b6"
	oldB749Drake       = "53599f89-1598-4d3d-8aa7-75b397b258eb"
	oldB749WallOfTears = "5825d718-6a32-4f4c-9b56-1a50051cfe28"
	oldB749Falcon      = "877ab494-da90-4c26-a184-c6ba64ec30b5"
	oldB749Deceiver    = "c304777a-39a9-4a2a-8cec-8116afb34c5b"
)

func TestOldB749CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB749Scapegoat: "Scapegoat", oldB749Verdant: "Verdant Touch", oldB749ElvenRite: "Elven Rite",
		oldB749Crossbow: "Crossbow Ambush", oldB749Razors: "Wall of Razors", oldB749Drake: "Spindrift Drake",
		oldB749WallOfTears: "Wall of Tears", oldB749Falcon: "Skyshroud Falcon", oldB749Deceiver: "Thalakos Deceiver",
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

func TestOldB749PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Wall of Razors", "Creature — Wall", oldB749Razors, []string{"defender", "first strike"}},
		{"Skyshroud Falcon", "Creature — Bird", oldB749Falcon, []string{"flying", "vigilance"}},
		{"Wall of Tears", "Creature — Wall", oldB749WallOfTears, []string{"defender"}},
		{"Spindrift Drake", "Creature — Drake", oldB749Drake, []string{"flying"}},
		{"Thalakos Deceiver", "Creature — Thalakos Wizard", oldB749Deceiver, []string{"shadow"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			for _, kw := range tc.keywords {
				if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestOldB749CrossbowAmbushGivesYourCreaturesReachUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Crossbow Ambush", "Instant", oldB749Crossbow, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "reach") {
		t.Error("my creature has no reach")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "reach") {
		t.Error("an opponent's creature gained reach")
	}
	late := effectstest.B12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if effectstest.HasKeywordOnBattlefield(t, g, late, "reach") {
		t.Error("a creature that arrived after the spell resolved gained reach")
	}
}

func TestOldB749ScapegoatSacrificesOneAndSavesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Scapegoat", "Instant", oldB749Scapegoat, effectstest.CardRefs(theirs), nil, []uuid.UUID{fodder}); err == nil {
		t.Fatal("an opponent's creature was accepted as a target")
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Scapegoat", "Instant", oldB749Scapegoat, effectstest.CardRefs(a, b), nil, nil); err == nil {
		t.Fatal("cast without sacrificing a creature")
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Scapegoat", "Instant", oldB749Scapegoat, effectstest.CardRefs(a, b), nil, []uuid.UUID{fodder}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || !me.Graveyard.Contains(fodder) {
		t.Error("the sacrificed creature did not go to the graveyard")
	}
	for name, id := range map[string]uuid.UUID{"A": a, "B": b} {
		if g.Battlefield.Contains(id) || !effectstest.HandHasID(me, id) {
			t.Errorf("Bear %s was not returned to its owner's hand", name)
		}
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("an opponent's creature was returned")
	}
}

func TestOldB749ScapegoatCanTargetNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	keep := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Scapegoat", "Instant", oldB749Scapegoat, nil, nil, []uuid.UUID{fodder}); err != nil {
		t.Fatalf("cast with no targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(keep) || g.Battlefield.Contains(fodder) {
		t.Error("only the sacrificed creature should have left")
	}
}

func TestOldB749ElvenRiteDistributesTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, g.Seats[1].ID, "Bear B", "Creature — Bear", 2, 2)
	hand := effectstest.HandCardOf(g, me, "Elven Rite", "Sorcery", "{1}{G}", oldB749ElvenRite)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	}); err != nil {
		t.Fatalf("Elven Rite: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"A": a, "B": b} {
		if got := effectstest.LayeredCard(t, g, id).Counters[game.CounterPlusOne]; got != 1 {
			t.Errorf("creature %s has %d +1/+1 counters, want 1", name, got)
		}
	}
}

func TestOldB749ElvenRiteCanPutBothOnOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	hand := effectstest.HandCardOf(g, me, "Elven Rite", "Sorcery", "{1}{G}", oldB749ElvenRite)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{
		Targets: effectstest.CardRefs(a), Distribution: map[uuid.UUID]int{a: 2},
	}); err != nil {
		t.Fatalf("Elven Rite: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, a).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("creature has %d +1/+1 counters, want 2", got)
	}
}

func TestOldB749VerdantTouchMakesAPermanent22CreatureLand(t *testing.T) {
	for _, buyback := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
		var opt []int
		if buyback {
			opt = []int{0}
		}
		id, err := effectstest.CastWithOptionalCosts(t, g, "Verdant Touch", "Sorcery", oldB749Verdant, effectstest.CardTarget(land), opt, nil)
		if err != nil {
			t.Fatalf("buyback=%v: cast: %v", buyback, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.LayeredCard(t, g, land)
		if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
			t.Errorf("buyback=%v: land is creature=%v land=%v %d/%d, want a 2/2 creature land", buyback, c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
		}
		effectstest.TmpB1AssertBuyback(t, me, id, buyback)
		// It lasts indefinitely: still a creature in a later turn.
		effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		if !effectstest.LayeredCard(t, g, land).IsCreature() {
			t.Errorf("buyback=%v: the land stopped being a creature", buyback)
		}
	}
}

func TestOldB749SpindriftDrakeIsSacrificedUnlessYouPayU(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		drake := effectstest.B12Push(g, me.ID, "Spindrift Drake", "Creature — Drake", oldB749Drake, 2, 1)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		if pay {
			effectstest.B06AddMana(me, "U")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("no pay prompt for the upkeep trigger")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(drake) != pay {
			t.Errorf("pay=%v: Drake on the battlefield = %v", pay, g.Battlefield.Contains(drake))
		}
	}
}

func TestOldB749WallOfTearsReturnsWhatItBlocksAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.HmlCreature(g, me.ID, "Attacker", "", 1, 5)
	wall := effectstest.HmlCreature(g, opp.ID, "Wall of Tears", oldB749WallOfTears, 0, 4)
	bystander := effectstest.HmlCreature(g, me.ID, "Bystander", "", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(attacker) {
		t.Fatal("the attacker left before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) || !effectstest.HandHasID(me, attacker) {
		t.Error("the blocked creature was not returned to its owner's hand at end of combat")
	}
	if !g.Battlefield.Contains(wall) || !g.Battlefield.Contains(bystander) {
		t.Error("the Wall or a bystander was returned")
	}
}

func TestOldB749ThalakosDeceiverTradesItselfForACreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		deceiver := effectstest.NcdAttacker(g, "Thalakos Deceiver", oldB749Deceiver, 1, 1)
		prize := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
		effectstest.NcdPlay(t, g, deceiver, yes, prize)
		c := effectstest.LayeredCard(t, g, prize)
		if yes {
			if c.Controller != me.ID {
				t.Errorf("yes: the target is controlled by %v, want me", c.Controller)
			}
			if g.Battlefield.Contains(deceiver) || !me.Graveyard.Contains(deceiver) {
				t.Error("yes: the Deceiver was not sacrificed")
			}
		} else {
			if c.Controller != opp.ID {
				t.Error("no: the control of the creature changed")
			}
			if !g.Battlefield.Contains(deceiver) {
				t.Error("no: the Deceiver was sacrificed")
			}
		}
	}
}
