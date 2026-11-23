package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB166Wayfarer = "8790c7be-c14a-4293-bc4a-3d0efe42ed79"
	gapB166Preston  = "2e00ccc9-6d5f-4de3-bbec-48b49d363244"
	gapB166Steward  = "559463f8-396f-40e0-9515-a1be6397ca31"
	gapB166Prowler  = "5e0d13b0-d5d0-4c5f-945e-5cda7d980670"
)

func TestGapB166EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB166Wayfarer: "Pious Wayfarer", gapB166Preston: "Preston, the Vanisher",
		gapB166Steward: "Serene Steward", gapB166Prowler: "Skyhunter Prowler",
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

func TestGapB166SkyhunterProwlerHasFlyingAndVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Skyhunter Prowler", "Creature — Cat Knight", gapB166Prowler, 1, 3)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestGapB166PiousWayfarerPumpsOnYourEnchantmentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Pious Wayfarer", "Creature — Human Scout", gapB166Wayfarer, 1, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Ench", TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Fatalf("an opponent's enchantment pumped the bear to %d power", p)
	}
	effectstest.CastCatalogSpell(t, g, "My Ench", "Enchantment", "", nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", p, tough)
	}
}

func TestGapB166SereneStewardPutsACounterOnlyWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		steward := effectstest.B12Push(g, me.ID, "Serene Steward", "Creature — Human Cleric Ally", gapB166Steward, 2, 2)
		bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		effectstest.B06AddMana(me, "W")
		effectstest.GainLifeNow(t, g, steward, me, 3)
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 1
		}
		if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != want {
			t.Errorf("pay=%v: %d counters, want %d", pay, got, want)
		}
	}
}

func gapB166Tokens(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

func TestGapB166PrestonCopiesOnlyNontokenCreaturesThatWereNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Preston, the Vanisher", "Legendary Creature — Rabbit Wizard", gapB166Preston, 2, 5)
	// Cast: no copy.
	effectstest.CastHoldCreature(t, g, "Cast Ogre", "Creature — Ogre", "", 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapB166Tokens(g, me.ID, "Cast Ogre")); n != 0 {
		t.Fatalf("a cast creature was copied %d times", n)
	}
	// Entered without being cast: one 0/1 white Illusion copy, and the copy
	// (a token) is not copied in turn.
	orig := game.NewCard("Fetched Ogre", me.ID)
	orig.TypeLine, orig.Power, orig.Toughness = "Creature — Ogre Warrior", 4, 4
	me.Hand.PushTop(orig)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, orig.InstanceID); err != nil {
		t.Fatalf("enter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	toks := gapB166Tokens(g, me.ID, "Fetched Ogre")
	if len(toks) != 1 {
		t.Fatalf("%d token copies, want 1", len(toks))
	}
	id := toks[0].InstanceID
	if p, tough := effectstest.PtOf(t, g, id); p != 0 || tough != 1 {
		t.Errorf("copy is %d/%d, want 0/1", p, tough)
	}
	if cols := effectstest.EffectiveColors(t, g, id); !slices.Equal(cols, []string{"W"}) {
		t.Errorf("copy colours %v, want [W]", cols)
	}
	if !toks[0].HasSubtype("Illusion") || toks[0].HasSubtype("Ogre") || toks[0].HasSubtype("Warrior") {
		t.Error("the copy should be an Illusion and no longer an Ogre")
	}
	if p, tough := effectstest.PtOf(t, g, orig.InstanceID); p != 4 || tough != 4 {
		t.Errorf("the original changed to %d/%d", p, tough)
	}
}

func TestGapB166PrestonSacrificesFiveIllusionsToExileANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	preston := effectstest.B12Push(g, me.ID, "Preston, the Vanisher", "Legendary Creature — Rabbit Wizard", gapB166Preston, 2, 5)
	var ill []uuid.UUID
	for i := 0; i < 5; i++ {
		ill = append(ill, effectstest.B12Creature(g, me.ID, "Illusion", "Creature — Illusion", 0, 1))
	}
	target := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, preston, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: ill[:4], Targets: effectstest.CardRefs(target),
	}); err == nil {
		t.Fatal("activated with four Illusions")
	}
	if err := g.ActivateCatalogAbility(me.ID, preston, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: ill, Targets: effectstest.CardRefs(target),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Error("the target is still on the battlefield")
	}
	if !g.Exile.Contains(target) {
		t.Error("the target was not exiled")
	}
	for _, id := range ill {
		if g.Battlefield.Contains(id) {
			t.Error("an Illusion was not sacrificed")
		}
	}
}
