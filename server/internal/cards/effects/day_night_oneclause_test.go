package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// day_night_oneclause_test.go — the single-clause day/night cards
// (#2561, ADR 0132). The Celestus's own tests are in
// day_night_cards_test.go, the werewolves' in werewolf_cards_test.go.

const (
	brimstoneVandalOracle     = "7af23925-6d32-407d-9a99-a57b2ebb63ac"
	firmamentSageOracle       = "57163ade-fe89-4ba0-98a9-daeeed5badb9"
	sunriseCavalierOracle     = "0718ea67-ced3-46bc-9164-8841e6bd73ae"
	moonragersSlashOracle     = "40a41774-fe50-438d-85e4-7f2beb29fb82"
	oliviasAmbushOracle       = "a3b8a08b-5409-4d5f-9bab-8a7a6376a5b9"
	obsessiveAstronomerOracle = "fde06097-1cf1-4774-b6ca-7cf15dfe4837"
	intoTheNightOracle        = "58c131f3-c623-4324-b061-9535018676b6"
)

// dnFlip flips the designation through the engine's own write, so the
// flip trigger fires.
func dnFlip(g *game.Game) {
	g.WithWriteLock(func() { g.ToggleDayNightForEffect() })
}

func TestBrimstoneVandalPingsOnAFlipAndNotOnTheFirstDesignation(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dnCast(t, g, "Brimstone Vandal", "Creature — Devil", brimstoneVandalOracle)
	if !g.IsDay() {
		t.Fatalf("entering from neither: %q, want day", dnDesignation(g))
	}
	before := opp.Life
	monSettle(t, g)
	if opp.Life != before {
		t.Fatalf("the first designation pinged: opponent life %d -> %d", before, opp.Life)
	}
	dnFlip(g)
	monSettle(t, g)
	if got := opp.Life; got != before-1 {
		t.Errorf("opponent life = %d, want %d after a flip", got, before-1)
	}
	if me.Life != before {
		t.Errorf("the controller lost life: %d", me.Life)
	}
	for _, o := range g.Seats[2:] {
		if o.Life != before-1 {
			t.Errorf("%s life = %d, want %d: each opponent takes 1", o.Name, o.Life, before-1)
		}
	}
}

func TestFirmamentSageDrawsOnAFlip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dnCast(t, g, "Firmament Sage", "Creature — Human Wizard", firmamentSageOracle)
	hand := me.Hand.Size()
	dnFlip(g)
	monSettle(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d after the flip", got, hand+1)
	}
}

func TestSunriseCavalierPutsACounterOnAChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	dnCast(t, g, "Sunrise Cavalier", "Creature — Human Knight", sunriseCavalierOracle)
	dnFlip(g)
	monSettle(t, g)
	answerPickTarget(t, g, bear)
	monSettle(t, g)
	c, _ := battlefieldCard(g, bear)
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("Bear counters = %v, want one +1/+1", c.Counters)
	}
}

// With only {R} floating, the spell is castable at night ({R} after the
// discount) and not by day ({2}{R}): the strict mana gate is what proves
// the discount is really charged.
func TestMoonragersSlashCostsTwoLessAtNight(t *testing.T) {
	castable := func(night bool) bool {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me, opp := g.Seats[0], g.Seats[1]
		if night {
			g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
		}
		id := handCardFull(me, "Moonrager's Slash", "Instant", "{2}{R}", moonragersSlashOracle, []string{"R"})
		floatMana(t, g, me, "{R}")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
		return err == nil
	}
	if !castable(true) {
		t.Error("Moonrager's Slash could not be cast for {R} at night")
	}
	if castable(false) {
		t.Error("Moonrager's Slash was cast for {R} by day")
	}
}

func TestOliviasMidnightAmbushShrinksByTwoByDayAndThirteenAtNight(t *testing.T) {
	for _, tc := range []struct {
		name  string
		night bool
		want  int
	}{{"day", false, 2}, {"night", true, 13}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			bear := pushCatalogPermanent(g, g.Seats[1].ID, "Big Bear", "Creature — Bear", "", false)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == bear {
						g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 20, 20
					}
				}
				if tc.night {
					g.DayNight.Designation = game.DesignationNight
				}
			})
			castCatalogSpell(t, g, "Olivia's Midnight Ambush", "Instant", oliviasAmbushOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
			passPriorityAroundTable(t, g)
			if got := effectivePower(t, g, bear); got != 20-tc.want {
				t.Errorf("power = %d, want %d", got, 20-tc.want)
			}
		})
	}
}

func TestIntoTheNightMakesItNightAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationDay })
	castCatalogSpell(t, g, "Into the Night", "Sorcery", intoTheNightOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.IsNight() {
		t.Fatalf("designation = %q, want night", dnDesignation(g))
	}
	hand := me.Hand.Size()
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt for Into the Night")
	}
	ids := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	if err := g.ResolveChooseCards(c.ID, me.ID, ids); err != nil {
		t.Fatalf("answer the discard: %v", err)
	}
	monSettle(t, g)
	if got := me.Hand.Size(); got != hand-2+3 {
		t.Errorf("hand = %d, want %d: discarded 2, drew 2+1", got, hand-2+3)
	}
}

func TestObsessiveAstronomerLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dnCast(t, g, "Obsessive Astronomer", "Creature — Human Wizard", obsessiveAstronomerOracle)
	dnFlip(g)
	monSettle(t, g)
	hand := me.Hand.Size()
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	ids := []uuid.UUID{me.Hand.Cards[0].InstanceID}
	if err := g.ResolveChooseCards(c.ID, me.ID, ids); err != nil {
		t.Fatalf("answer the discard: %v", err)
	}
	monSettle(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand = %d, want %d: discard one, draw one", got, hand)
	}
}
