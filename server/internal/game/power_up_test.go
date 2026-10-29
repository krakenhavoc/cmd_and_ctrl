package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// power_up_test.go — #4527, CR 702.193. The engine half of power-up:
// "Activate this ability only once" per object, and "if this permanent
// entered this turn, this ability's cost is reduced by this permanent's
// mana cost", coloured symbols included (CR 702.193b, 118.7b–g).
// The cards (Brave Brawler, Hulk, Advancing the Spirit) are in
// internal/cards/effects/power_up_cards_test.go.

const powerUpProbeLabel = "Power-up — {4}{W}: probe"

func powerUpProbe() ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label:   powerUpProbeLabel,
		PowerUp: true,
		Cost:    AbilityCost{Mana: "{4}{W}"},
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "ran-power-up", 1)
		},
	}
}

// pushPowerUpSource puts a {1}{W} creature keyed on `oracle` straight
// onto the battlefield, which is not an entry (no EventETB), so it did
// not enter this turn.
func pushPowerUpSource(g *Game, owner *Player, oracle string) uuid.UUID {
	c := NewCard("Power-up Probe", owner.ID)
	c.TypeLine = "Creature — Human"
	c.ManaCost = "{1}{W}"
	c.Power, c.Toughness = 2, 2
	c.OracleID = oracle
	c.Controller = owner.ID
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func powerUpPrice(t *testing.T, g *Game, who, src uuid.UUID) string {
	t.Helper()
	p, err := g.PriceActivation(who, src, 0, nil)
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	return p.Total.String()
}

// firePowerUp activates the probe with `mana` floated, one token per
// symbol ("W", "C").
func firePowerUp(t *testing.T, g *Game, me *Player, src uuid.UUID, mana ...string) error {
	t.Helper()
	for _, c := range mana {
		me.ManaPool.AddMana(ManaToken{Color: c})
	}
	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{})
	if err != nil {
		me.ManaPool = nil
	}
	return err
}

func TestAPowerUpAbilityCanBeActivatedOnlyOnce(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "power-up-once", powerUpProbe())
	src := pushPowerUpSource(g, me, "power-up-once")

	if err := firePowerUp(t, g, me, src, "C", "C", "C", "C", "W"); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	passBothForTest(g)
	if got := counterOf(g, src, "ran-power-up"); got != 1 {
		t.Fatalf("the first activation resolved %d times, want 1", got)
	}
	for _, c := range []string{"C", "C", "C", "C", "W"} {
		me.ManaPool.AddMana(ManaToken{Color: c})
	}
	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{})
	if !errors.Is(err, ErrPowerUpUsed) {
		t.Fatalf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
	if me.ManaPool.Total() != 5 {
		t.Errorf("pool = %v, want all five floating — a refused activation pays nothing", me.ManaPool)
	}
	var used bool
	g.WithWriteLock(func() { used = g.PowerUpUsed(src, powerUpProbe()) })
	if !used {
		t.Error("PowerUpUsed = false after the activation")
	}
}

// TestAFlickeredCreatureMayPowerUpAgain — CR 400.7: the creature that
// comes back is a new object, so its power-up has not been activated,
// and since it entered this turn it costs {4}{W} less {1}{W}.
func TestAFlickeredCreatureMayPowerUpAgain(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "power-up-flicker", powerUpProbe())
	src := pushPowerUpSource(g, me, "power-up-flicker")

	if err := firePowerUp(t, g, me, src, "C", "C", "C", "C", "W"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passBothForTest(g)
	if err := firePowerUp(t, g, me, src, "C", "C", "C", "C", "W"); !errors.Is(err, ErrPowerUpUsed) {
		t.Fatalf("staying put: err = %v, want ErrPowerUpUsed", err)
	}

	bounceAndReturn(t, g, src, me.ID)
	if got := powerUpPrice(t, g, me.ID, src); got != "{3}" {
		t.Errorf("the returned creature's power-up costs %q, want {3} — it entered this turn", got)
	}
	if err := firePowerUp(t, g, me, src, "C", "C", "C"); err != nil {
		t.Errorf("the returned creature cannot power up: %v", err)
	}
}

// TestThePowerUpCostIsReducedOnlyTheTurnItEntered — CR 702.193a: the
// reduction is the turn the permanent entered, not the turn after, and
// not for a permanent that has been there all along.
func TestThePowerUpCostIsReducedOnlyTheTurnItEntered(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "power-up-turn", powerUpProbe())
	src := pushPowerUpSource(g, me, "power-up-turn")

	if got := powerUpPrice(t, g, me.ID, src); got != "{4}{W}" {
		t.Fatalf("a creature that did not enter this turn: %q, want the printed {4}{W}", got)
	}
	bounceAndReturn(t, g, src, me.ID)
	if got := powerUpPrice(t, g, me.ID, src); got != "{3}" {
		t.Fatalf("the turn it entered: %q, want {3} ({4}{W} less {1}{W})", got)
	}
	passTurnTo(t, g, 1)
	passTurnTo(t, g, 0)
	if got := powerUpPrice(t, g, me.ID, src); got != "{4}{W}" {
		t.Errorf("the next turn: %q, want the printed {4}{W}", got)
	}
}

// TestPowerUpNeedsTheBitToBeReducedOrLimited — the keyword is
// the bit. The same cost on a plain ability is repeatable and full price.
func TestPowerUpNeedsTheBitToBeReducedOrLimited(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	plain := powerUpProbe()
	plain.PowerUp = false
	plain.Label = "{4}{W}: plain probe"
	exhaustCatalog(t, "power-up-plain", plain)
	src := pushPowerUpSource(g, me, "power-up-plain")
	bounceAndReturn(t, g, src, me.ID)
	if got := powerUpPrice(t, g, me.ID, src); got != "{4}{W}" {
		t.Errorf("a plain ability the turn its source entered: %q, want {4}{W}", got)
	}
	for i := 0; i < 2; i++ {
		if err := firePowerUp(t, g, me, src, "C", "C", "C", "C", "W"); err != nil {
			t.Fatalf("plain activation %d: %v", i+1, err)
		}
		passBothForTest(g)
	}
}

// TestPowerUpReductionFollowsCR702_193b pins the reduction symbol by
// symbol: generic reduces generic, a coloured or colourless symbol
// reduces mana of its type and its excess reduces generic (118.7b–d),
// hybrid takes a matching half (118.7e), Phyrexian is its colour
// (118.7f), snow is generic (118.7g), and nothing goes below {0}.
func TestPowerUpReductionFollowsCR702_193b(t *testing.T) {
	for _, tc := range []struct {
		name, cost, by, want string
	}{
		{"generic and colour (Brave Brawler)", "{4}{W}", "{1}{W}", "{3}"},
		{"two colours (Hulk)", "{6}{R}{G}", "{3}{R}{G}", "{3}"},
		{"excess colour reduces generic (118.7c)", "{5}{R}", "{R}{R}", "{4}"},
		{"colour the cost lacks reduces generic (118.7b)", "{6}", "{2}{W}", "{3}"},
		{"colourless takes colourless", "{C}{W}{U}{B}{R}{G}", "{R}{W}{B}", "{C}{U}{G}"},
		{"excess colourless reduces generic (118.7d)", "{2}{U}", "{C}{C}", "{U}"},
		{"generic never takes a colour (118.7a)", "{1}{W}{W}", "{5}", "{W}{W}"},
		{"hybrid cost by hybrid (Brawn)", "{4}{G/U}", "{1}{G/U}", "{3}"},
		{"hybrid by hybrid, two hybrids (Abomination)", "{5}{R/G}{R/G}", "{3}{R/G}", "{2}{R/G}"},
		{"a colour takes the exact symbol before a hybrid", "{R/G}{R}", "{R}", "{R/G}"},
		{"hybrid with no match reduces one generic", "{4}{B}", "{W/U}", "{3}{B}"},
		{"Phyrexian is its colour (118.7f)", "{3}{W}", "{W/P}", "{3}"},
		{"snow is generic (118.7g)", "{3}{G}", "{S}", "{2}{G}"},
		{"X in the mana cost is zero", "{4}{W}", "{X}{W}", "{4}"},
		{"down to nothing", "{U}", "{2}{U}{U}", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cost, err := ParseCost(tc.cost)
			if err != nil {
				t.Fatal(err)
			}
			by, err := ParseCost(tc.by)
			if err != nil {
				t.Fatal(err)
			}
			before := cost.String()
			if got := reduceByManaCost(cost, by).String(); got != tc.want {
				t.Errorf("%s reduced by %s = %q, want %q", tc.cost, tc.by, got, tc.want)
			}
			if after := cost.String(); after != before {
				t.Errorf("the reduction changed its input: %q, was %q", after, before)
			}
		})
	}
}

// TestThePowerUpReductionReachesTheManaAnnouncedForX — Stature's
// "{X}{U}{U}" under its own {U}: the coloured symbol goes, and nothing
// generic is left over to take off the X.
func TestThePowerUpReductionReachesTheManaAnnouncedForX(t *testing.T) {
	cost, _ := ParseCost("{X}{U}{U}")
	by, _ := ParseCost("{U}")
	got := reduceByManaCost(cost, by)
	if got.String() != "{X}{U}" || got.XReduced != 0 {
		t.Errorf("{X}{U}{U} reduced by {U} = %q (XReduced %d), want {X}{U} and nothing off X", got.String(), got.XReduced)
	}
	by, _ = ParseCost("{2}{U}")
	got = reduceByManaCost(cost, by)
	if got.String() != "{X}{U}" || got.XReduced != 2 {
		t.Errorf("{X}{U}{U} reduced by {2}{U} = %q (XReduced %d), want {X}{U} with 2 off X", got.String(), got.XReduced)
	}
}

// TestAPowerUpStaysUsedAcrossASnapshotRoundTrip — the record is game
// state a player can lose a game over, so a restored game refuses the
// second activation exactly as the live one does.
func TestAPowerUpStaysUsedAcrossASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "power-up-snapshot", powerUpProbe())
	src := pushPowerUpSource(g, me, "power-up-snapshot")
	if err := firePowerUp(t, g, me, src, "C", "C", "C", "C", "W"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passBothForTest(g)

	_, restored := roundTrip(t, g)
	var used bool
	restored.WithWriteLock(func() { used = restored.PowerUpUsed(src, powerUpProbe()) })
	if !used {
		t.Fatal("the restored game reads the power-up as unused")
	}
	if err := firePowerUp(t, restored, restored.Seats[0], src, "C", "C", "C", "C", "W"); !errors.Is(err, ErrPowerUpUsed) {
		t.Errorf("the restored game: err = %v, want ErrPowerUpUsed", err)
	}
	if got := restored.TurnTallyFor(me.ID).PowerUpActivations; got != 1 {
		t.Errorf("the restored game counts %d power-up activations this turn, want 1", got)
	}
}

// TestAPowerUpActivationIsAnnouncedAsOne — Event.PowerUp on the
// announcement (Marvel Boy, Noh-Varr's "whenever you activate a
// power-up ability") and the per-turn count Advancing the Spirit reads,
// both written whether or not the ability resolves.
func TestAPowerUpActivationIsAnnouncedAsOne(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "power-up-event", powerUpProbe())
	src := pushPowerUpSource(g, me, "power-up-event")
	if err := firePowerUp(t, g, me, src, "C", "C", "C", "C", "W"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	var stamped bool
	for _, ev := range g.Events {
		if ev.Kind == EventActivateAbility && ev.Source == src {
			stamped = ev.PowerUp
		}
	}
	if !stamped {
		t.Error("EventActivateAbility does not carry PowerUp")
	}
	var n int
	g.WithWriteLock(func() { n = g.PowerUpActivationsThisTurn(me.ID) })
	if n != 1 {
		t.Errorf("PowerUpActivationsThisTurn = %d before resolution, want 1", n)
	}
}

// TestPowerUpZeroAlternativeReplacesTheCostButNeverAnXCost —
// CostZeroAlternative (Advancing the Spirit's shape) prices a power-up
// from {0}, and is not applied to a cost with {X}: taking it would hold
// X to 0 (CR 107.3b through 602.2b), and the engine takes it without
// asking. A tax on activations still applies to the alternative
// (CR 118.9d).
func TestPowerUpZeroAlternativeReplacesTheCostButNeverAnXCost(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	costModifierCatalog(t, "zero-alt", CostModifier{
		Kind: CostZeroAlternative, Label: "zero", Activations: true,
		AppliesTo: func(q CostQuery) bool { return q.Ability != nil && q.Ability.PowerUp },
	})
	alt := NewCard("Zero Alternative", me.ID)
	alt.TypeLine = "Enchantment"
	alt.OracleID = "zero-alt"
	alt.Controller = me.ID
	g.Battlefield.PushTop(alt)

	exhaustCatalog(t, "power-up-plain-cost", powerUpProbe())
	xProbe := powerUpProbe()
	xProbe.Label = "Power-up — {X}{U}{U}: probe"
	xProbe.Cost = AbilityCost{Mana: "{X}{U}{U}"}
	exhaustCatalog(t, "power-up-x-cost", xProbe)
	plain := pushPowerUpSource(g, me, "power-up-plain-cost")
	withX := pushPowerUpSource(g, me, "power-up-x-cost")

	if got := powerUpPrice(t, g, me.ID, plain); got != "" {
		t.Errorf("a power-up under the zero alternative: %q, want {0}", got)
	}
	if got := powerUpPrice(t, g, me.ID, withX); got != "{X}{U}{U}" {
		t.Errorf("an {X} power-up under the zero alternative: %q, want the printed {X}{U}{U}", got)
	}

	costModifierCatalog(t, "tax", CostModifier{Kind: CostIncrease, Label: "tax", Activations: true,
		Amount: func(CostQuery) int { return 2 }})
	tax := NewCard("Tax", me.ID)
	tax.TypeLine = "Enchantment"
	tax.OracleID = "tax"
	tax.Controller = me.ID
	g.Battlefield.PushTop(tax)
	if got := powerUpPrice(t, g, me.ID, plain); got != "{2}" {
		t.Errorf("the alternative under a {2} tax: %q, want {2} — increases apply to the alternative", got)
	}
}
