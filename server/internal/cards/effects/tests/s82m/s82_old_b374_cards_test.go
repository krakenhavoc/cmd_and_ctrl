package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o374Radiance      = "0870e9f8-9e02-4997-9554-41a4e8eec4cb"
	o374Grisly        = "0bb359bb-a1db-4535-8245-097ee989f4f5"
	o374Sunbond       = "163ece8a-0479-4bb2-b559-581985d1898b"
	o374Claim         = "557dda97-c9e6-4d3d-95d8-8180ab9d87b3"
	o374Storm         = "833fe5f4-cfab-46ce-a4fe-250673896350"
	o374Temper        = "85e9bc3e-e0da-4ad4-9631-472f75f91663"
	o374Wolves        = "a9e364e6-6202-4859-9f65-69c9951959ac"
	o374Might         = "d08117ec-e6ec-4c63-91d4-1bda59446c54"
	o374Enlightenment = "da848568-1c66-4ce1-87f0-a35c317f2c61"
	o374Intellect     = "e766b774-eaf3-41ce-bc23-fd29fd22bc39"
)

func TestO374CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o374Radiance: "Ephara's Radiance", o374Grisly: "Grisly Transformation", o374Sunbond: "Sunbond",
		o374Claim: "Claim of Erebos", o374Storm: "Epiphany Storm", o374Temper: "Fearsome Temper",
		o374Wolves: "Raised by Wolves", o374Might: "Thunderous Might", o374Enlightenment: "Ephara's Enlightenment",
		o374Intellect: "Evanescent Intellect",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// o374Setup is a table at main phase with a vanilla bear under the active seat.
func o374Setup(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	return g, me, effectstest.AuraBear(g, me.ID)
}

func o374Activate(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID, targets []game.TargetRef) error {
	t.Helper()
	return g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

func TestO374EpharasRadianceGivesAGainThreeAbility(t *testing.T) {
	g, me, bear := o374Setup(t)
	effectstest.AuraCast(t, g, "Ephara's Radiance", o374Radiance, bear)
	life := me.Life
	if err := o374Activate(t, g, me, bear, nil); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.ApaMana(me, "W", "W")
	if err := o374Activate(t, g, me, bear, nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Fatalf("life %d, want %d", me.Life, life+3)
	}
	effectstest.ApaMana(me, "W", "W")
	if err := o374Activate(t, g, me, bear, nil); err == nil {
		t.Fatal("activated twice: the creature is tapped")
	}
}

func TestO374GrislyTransformationDrawsAndGivesIntimidate(t *testing.T) {
	g, me, bear := o374Setup(t)
	hand := len(me.Hand.Cards)
	effectstest.AuraSeedLibrary(me, 3)
	effectstest.AuraCast(t, g, "Grisly Transformation", o374Grisly, bear)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, len(me.Hand.Cards))
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "intimidate") {
		t.Error("the enchanted creature has no intimidate")
	}
}

func TestO374SunbondPutsThatManyCountersWhenYouGainLife(t *testing.T) {
	g, me, bear := o374Setup(t)
	effectstest.AuraCast(t, g, "Sunbond", o374Sunbond, bear)
	if err := g.ChangePlayerLifeForEffect(bear, me.ID, 3); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCardByID(g, bear)
	if n := c.Counters[game.CounterPlusOne]; n != 3 {
		t.Fatalf("%d +1/+1 counters, want 3", n)
	}
	// An opponent's life gain is not "you".
	if err := g.ChangePlayerLifeForEffect(bear, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, 5); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ = effectstest.BattlefieldCardByID(g, bear)
	if n := c.Counters[game.CounterPlusOne]; n != 3 {
		t.Fatalf("%d counters after an opponent gained life, want 3", n)
	}
}

func TestO374ClaimOfErebosDrainsTwo(t *testing.T) {
	g, me, bear := o374Setup(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AuraCast(t, g, "Claim of Erebos", o374Claim, bear)
	effectstest.ApaMana(me, "B", "B")
	life := opp.Life
	if err := o374Activate(t, g, me, bear, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("life %d, want %d", opp.Life, life-2)
	}
}

func TestO374EpiphanyStormLoots(t *testing.T) {
	g, me, bear := o374Setup(t)
	effectstest.AuraCast(t, g, "Epiphany Storm", o374Storm, bear)
	effectstest.AuraSeedLibrary(me, 3)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.ApaMana(me, "R")
	if err := o374Activate(t, g, me, bear, nil); err == nil {
		t.Fatal("activated with no card to discard")
	}
	card := effectstest.HandCardFull(me, "Test Card", "Sorcery", "{4}{R}", "", []string{"R"})
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(card) || len(me.Hand.Cards) != 1 {
		t.Fatalf("hand %d cards, discarded card still there %v: want the card swapped for a draw", len(me.Hand.Cards), me.Hand.Contains(card))
	}
}

func TestO374FearsomeTemperPumpsAndBarsABlocker(t *testing.T) {
	g, me, bear := o374Setup(t)
	theirs := effectstest.PushVanillaCreature(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Bear", 2, 2)
	effectstest.AuraCast(t, g, "Fearsome Temper", o374Temper, bear)
	if p, tough := effectstest.B38PT(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("%d/%d, want 4/4", p, tough)
	}
	effectstest.ApaMana(me, "R", "R", "R")
	if err := o374Activate(t, g, me, bear, effectstest.CardRefs(theirs)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if r := effectstest.AuraRestrictions(t, g, theirs); r&game.CantBlock != 0 {
		t.Error("the target can't block anything: the bar is only on blocking the enchanted creature")
	}
}

func TestO374RaisedByWolvesMakesTwoWolvesAndPumpsPerWolf(t *testing.T) {
	g, me, bear := o374Setup(t)
	effectstest.AuraCast(t, g, "Raised by Wolves", o374Wolves, bear)
	wolves := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Wolf" && c.Controller == me.ID {
			wolves++
		}
	}
	if wolves != 2 {
		t.Fatalf("%d Wolf tokens, want 2", wolves)
	}
	if p, tough := effectstest.B38PT(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("%d/%d, want 4/4 (+2/+2 for two Wolves)", p, tough)
	}
}

func TestO374ThunderousMightAddsDevotionToRedWhenItAttacks(t *testing.T) {
	g, me, bear := o374Setup(t)
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Red Thing", TypeLine: "Creature — Test", ManaCost: "{R}{R}", Power: 1, Toughness: 1})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Red Thing Two", TypeLine: "Creature — Test", ManaCost: "{1}{R}", Power: 1, Toughness: 1})
	effectstest.AuraCast(t, g, "Thunderous Might", o374Might, bear)
	effectstest.B64ReadyToAttack(g, bear)
	effectstest.LciB1Attack(t, g, bear, g.Seats[(g.Turn.ActiveSeat+1)%4].ID)
	effectstest.PassPriorityAroundTable(t, g)
	// Devotion to red: Red Thing {R}{R} and Red Thing Two {1}{R}.
	if p, _ := effectstest.B38PT(t, g, bear); p != 5 {
		t.Fatalf("power %d, want 5 (2 + devotion 3)", p)
	}
}

func TestO374EpharasEnlightenmentCountersFliesAndMayReturn(t *testing.T) {
	g, me, bear := o374Setup(t)
	aura := effectstest.AuraCast(t, g, "Ephara's Enlightenment", o374Enlightenment, bear)
	c, _ := effectstest.BattlefieldCardByID(g, bear)
	if n := c.Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("%d counters, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("no flying")
	}
	effectstest.PushVanillaCreature(g, me.ID, "Newcomer", 1, 1)
	g.RunStateChecksForTest()
	// A creature entering through the door triggers the "may"; push a second one by casting.
	effectstest.EnterFromHand(t, g, me.ID, "Another Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Fatal("the Aura did not return to its owner's hand")
	}
}

func TestO374EvanescentIntellectMillsThree(t *testing.T) {
	g, me, bear := o374Setup(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AuraCast(t, g, "Evanescent Intellect", o374Intellect, bear)
	effectstest.AuraSeedLibrary(opp, 6)
	lib := len(opp.Library.Cards)
	effectstest.ApaMana(me, "U", "U")
	if err := o374Activate(t, g, me, bear, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Library.Cards) != lib-3 {
		t.Fatalf("library %d -> %d, want three milled", lib, len(opp.Library.Cards))
	}
}
