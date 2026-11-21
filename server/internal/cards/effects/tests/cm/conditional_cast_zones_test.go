package testscm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestLightwheelEnhancementsPumpsAndGrantsVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Lightwheel Enhancements", effectstest.LightwheelEnhancementsOracle, bear)

	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("enchanted bear is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "vigilance") {
		t.Errorf("abilities %v lack vigilance", effectstest.EffectiveAbilities(t, g, bear))
	}
	// Start your engines! (CR 702.179a): the Aura on the battlefield
	// gives its controller speed 1.
	if got := g.SpeedOf(me.ID); got != 1 {
		t.Errorf("speed after the Aura entered = %d, want 1", got)
	}
}

func TestLightwheelEnhancementsEnchantsAVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vehicle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Wagon", TypeLine: "Artifact — Vehicle",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	aura := effectstest.AuraCast(t, g, "Lightwheel Enhancements", effectstest.LightwheelEnhancementsOracle, vehicle)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	id := effectstest.CondZoneGraveyardCard(t, g, me, "Lightwheel Enhancements", "Enchantment — Aura", "{W}", effectstest.LightwheelEnhancementsOracle)
	target := game.TargetRef{Kind: game.TargetCard, ID: bear}

	if err := effectstest.CondZoneCastFromGraveyard(g, me, id, target); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("with no speed: got %v, want ErrCastZoneNotAllowed", err)
	}
	g.SetSpeedForTest(me.ID, game.MaxSpeed-1)
	if err := effectstest.CondZoneCastFromGraveyard(g, me, id, target); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("at speed %d: got %v, want ErrCastZoneNotAllowed", game.MaxSpeed-1, err)
	}

	g.SetSpeedForTest(me.ID, game.MaxSpeed)
	if err := effectstest.CondZoneCastFromGraveyard(g, me, id, target); err != nil {
		t.Fatalf("at max speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("the recast Aura did not pump: bear power %d, want 3", p)
	}
}

// Max speed is its owner's, not the table's.
func TestLightwheelEnhancementsIgnoresAnOpponentsSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, other := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	id := effectstest.CondZoneGraveyardCard(t, g, me, "Lightwheel Enhancements", "Enchantment — Aura", "{W}", effectstest.LightwheelEnhancementsOracle)

	g.SetSpeedForTest(other.ID, game.MaxSpeed)
	err := effectstest.CondZoneCastFromGraveyard(g, me, id, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("an opponent's max speed opened the graveyard: %v", err)
	}
}

func TestGravecrawlerCastFromTheGraveyardNeedsAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CondZoneGraveyardCard(t, g, me, "Gravecrawler", "Creature — Zombie", "{B}", effectstest.GravecrawlerOracle)

	if err := effectstest.CondZoneCastFromGraveyard(g, me, id); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("with no Zombie: got %v, want ErrCastZoneNotAllowed", err)
	}
	// A non-Zombie does not count.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := effectstest.CondZoneCastFromGraveyard(g, me, id); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("with only a Bear: got %v, want ErrCastZoneNotAllowed", err)
	}
	// An opponent's Zombie is not yours.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Zombie", TypeLine: "Creature — Zombie",
		Power: 2, Toughness: 2, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := effectstest.CondZoneCastFromGraveyard(g, me, id); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("with an opponent's Zombie: got %v, want ErrCastZoneNotAllowed", err)
	}

	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Zombie", TypeLine: "Creature — Zombie",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := effectstest.CondZoneCastFromGraveyard(g, me, id); err != nil {
		t.Fatalf("with a Zombie of mine: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if r := effectstest.AuraRestrictions(t, g, id); r&game.CantBlock == 0 {
		t.Errorf("the recast Gravecrawler can block (restrictions %v)", r)
	}
}

func TestMarangRiverProwlerCastFromTheGraveyardNeedsABlackOrGreenPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CondZoneGraveyardCard(t, g, me, "Marang River Prowler", "Creature — Human Rogue", "{2}{U}", effectstest.MarangRiverProwlerOracle)

	blue := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Thing", TypeLine: effectstest.TestCreatureTypeLine, ManaCost: "{1}{U}", Colors: []string{"U"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	_ = blue
	if err := effectstest.CondZoneCastFromGraveyard(g, me, id); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("with only a blue permanent: got %v, want ErrCastZoneNotAllowed", err)
	}

	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Thing", TypeLine: effectstest.TestCreatureTypeLine, ManaCost: "{1}{G}", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := effectstest.CondZoneCastFromGraveyard(g, me, id); err != nil {
		t.Fatalf("with a green permanent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	r := effectstest.AuraRestrictions(t, g, id)
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
			effectstest.MustPanic(t, tc.want, func() {
				effects.Register(effects.Spec{
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
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: "test-conditional-zone-bound-cost",
		Name:     "Conditional Flashback",
		CastableZonesIf: []game.ConditionalCastZone{
			{Zone: game.ZoneGraveyard, Cond: func(*game.Game, uuid.UUID, game.Card) bool { return true }},
		},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{R}")},
	})
	if !effects.Has("test-conditional-zone-bound-cost") {
		t.Error("a flashback cost on a conditional graveyard did not register")
	}
}
