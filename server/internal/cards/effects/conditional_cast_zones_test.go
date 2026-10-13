package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_cast_zones_test.go — #3105. Spec.CastableZonesIf and the
// three cards that use it: Lightwheel Enhancements (max speed),
// Gravecrawler (a Zombie) and Marang River Prowler (a black or green
// permanent). Each test casts the card out of the graveyard through the
// engine's own CastSpell, so the claim is "the cast is accepted /
// refused", not "a helper returns true".

const (
	lightwheelEnhancementsOracle = "17fae94e-558f-449a-b30c-e3c37a068b79"
	gravecrawlerOracle           = "09ff28b1-b6c9-48e6-b12e-2f0e644f709f"
	marangRiverProwlerOracle     = "002e767c-0b30-4ad1-a924-88da400a4cff"
)

// condZoneGraveyardCard puts a catalog card into the seat's graveyard at a
// main phase.
func condZoneGraveyardCard(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	c := game.NewCard(name, owner.ID)
	c.TypeLine = typeLine
	c.ManaCost = cost
	c.OracleID = oracle
	owner.Graveyard.PushTop(c)
	return c.InstanceID
}

func condZoneCastFromGraveyard(g *game.Game, owner *game.Player, id uuid.UUID, targets ...game.TargetRef) error {
	return g.CastSpell(owner.ID, id, game.CastSpellParams{FromZone: "graveyard", Targets: targets})
}

func TestLightwheelEnhancementsPumpsAndGrantsVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Lightwheel Enhancements", lightwheelEnhancementsOracle, bear)

	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("enchanted bear is %d/%d, want 3/3", p, tt)
	}
	if !containsString(effectiveAbilities(t, g, bear), "vigilance") {
		t.Errorf("abilities %v lack vigilance", effectiveAbilities(t, g, bear))
	}
	// Start your engines! (CR 702.179a): the Aura on the battlefield
	// gives its controller speed 1.
	if got := g.SpeedOf(me.ID); got != 1 {
		t.Errorf("speed after the Aura entered = %d, want 1", got)
	}
}

func TestLightwheelEnhancementsEnchantsAVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vehicle := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Wagon", TypeLine: "Artifact — Vehicle",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	aura := auraCast(t, g, "Lightwheel Enhancements", lightwheelEnhancementsOracle, vehicle)
	var host *game.Card
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(aura); ok {
			host = g.AttachedHostOf(&c)
		}
	})
	if host == nil || host.InstanceID != vehicle {
		t.Errorf("the Aura did not attach to an uncrewed Vehicle (host %v)", host)
	}
}

// The proof card: cast from the graveyard only at max speed, and the
// permission follows the speed in both directions.
func TestLightwheelEnhancementsCastableFromTheGraveyardOnlyAtMaxSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	id := condZoneGraveyardCard(t, g, me, "Lightwheel Enhancements", "Enchantment — Aura", "{W}", lightwheelEnhancementsOracle)
	target := game.TargetRef{Kind: game.TargetCard, ID: bear}

	if err := condZoneCastFromGraveyard(g, me, id, target); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("with no speed: got %v, want ErrCastZoneNotAllowed", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed-1)
	if err := condZoneCastFromGraveyard(g, me, id, target); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("at speed %d: got %v, want ErrCastZoneNotAllowed", game.MaxSpeed-1, err)
	}

	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := condZoneCastFromGraveyard(g, me, id, target); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("the recast Aura did not pump: bear power %d, want 3", p)
	}
}

// Max speed is its owner's, not the table's.
func TestLightwheelEnhancementsIgnoresAnOpponentsSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	id := condZoneGraveyardCard(t, g, me, "Lightwheel Enhancements", "Enchantment — Aura", "{W}", lightwheelEnhancementsOracle)

	g.SetSpeedForTest(other.ID, game.MaxSpeed)
	err := condZoneCastFromGraveyard(g, me, id, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("an opponent's max speed opened the graveyard: %v", err)
	}
}

func TestGravecrawlerCastFromTheGraveyardNeedsAZombie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := condZoneGraveyardCard(t, g, me, "Gravecrawler", "Creature — Zombie", "{B}", gravecrawlerOracle)

	if err := condZoneCastFromGraveyard(g, me, id); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("with no Zombie: got %v, want ErrCastZoneNotAllowed", err)
	}
	// A non-Zombie does not count.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := condZoneCastFromGraveyard(g, me, id); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("with only a Bear: got %v, want ErrCastZoneNotAllowed", err)
	}
	// An opponent's Zombie is not yours.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Zombie", TypeLine: "Creature — Zombie",
		Power: 2, Toughness: 2, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := condZoneCastFromGraveyard(g, me, id); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("with an opponent's Zombie: got %v, want ErrCastZoneNotAllowed", err)
	}

	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Zombie", TypeLine: "Creature — Zombie",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := condZoneCastFromGraveyard(g, me, id); err != nil {
		t.Fatalf("with a Zombie of mine: %v", err)
	}
	passPriorityAroundTable(t, g)
	if r := auraRestrictions(t, g, id); r&game.CantBlock == 0 {
		t.Errorf("the recast Gravecrawler can block (restrictions %v)", r)
	}
}

func TestMarangRiverProwlerCastFromTheGraveyardNeedsABlackOrGreenPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := condZoneGraveyardCard(t, g, me, "Marang River Prowler", "Creature — Human Rogue", "{2}{U}", marangRiverProwlerOracle)

	blue := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Thing", TypeLine: testCreatureTypeLine, ManaCost: "{1}{U}", Colors: []string{"U"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	_ = blue
	if err := condZoneCastFromGraveyard(g, me, id); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("with only a blue permanent: got %v, want ErrCastZoneNotAllowed", err)
	}

	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Thing", TypeLine: testCreatureTypeLine, ManaCost: "{1}{G}", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := condZoneCastFromGraveyard(g, me, id); err != nil {
		t.Fatalf("with a green permanent: %v", err)
	}
	passPriorityAroundTable(t, g)
	r := auraRestrictions(t, g, id)
	if r&game.CantBlock == 0 || r&game.CantBeBlocked == 0 {
		t.Errorf("the recast Prowler has restrictions %v, want can't block and can't be blocked", r)
	}
}

// The registry refuses the shapes that compile and then mean something
// else: a conditional zone that is not a graveyard, and one with no
// condition (an unconditional zone written the long way round).
func TestRegisterRefusesAMalformedConditionalCastZone(t *testing.T) {
	always := func(*game.Game, uuid.UUID, game.Card) bool { return true }
	for _, tc := range []struct {
		name string
		zone game.ZoneKind
		cond func(*game.Game, uuid.UUID, game.Card) bool
		want string
	}{
		{"exile", game.ZoneExile, always, "only a graveyard"},
		{"hand", game.ZoneHand, always, "only a graveyard"},
		{"nil condition", game.ZoneGraveyard, nil, "no condition"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mustPanic(t, tc.want, func() {
				Register(Spec{
					OracleID: "test-conditional-zone-" + tc.name,
					Name:     "Malformed " + tc.name,
					CastableZonesIf: []game.ConditionalCastZone{
						{Zone: tc.zone, Cond: tc.cond},
					},
				})
			})
		})
	}
}

// A zone-bound alternative cost may name a conditional zone, just as it
// may name an unconditional one.
func TestRegisterAcceptsAZoneBoundCostOnAConditionalZone(t *testing.T) {
	registerForTest(t, Spec{
		OracleID: "test-conditional-zone-bound-cost",
		Name:     "Conditional Flashback",
		CastableZonesIf: []game.ConditionalCastZone{
			{Zone: game.ZoneGraveyard, Cond: func(*game.Game, uuid.UUID, game.Card) bool { return true }},
		},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}")},
	})
	if !Has("test-conditional-zone-bound-cost") {
		t.Error("a flashback cost on a conditional graveyard did not register")
	}
}
