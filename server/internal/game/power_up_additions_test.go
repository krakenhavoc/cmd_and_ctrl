package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// power_up_additions_test.go — #4571, the engine half of the three
// power-up additions (CR 702.193). The cards (Wonder Man, Kang the
// Conqueror, Quinjet Technician) are in
// internal/cards/effects/power_up_additions_test.go.

// "Spend this mana only to activate power-up abilities" (CR 106.6): the
// tag admits an activation of a power-up ability and nothing else.
func TestManaRestrictPowerUpAdmitsOnlyAPowerUpActivation(t *testing.T) {
	tags := []string{ManaRestrictPowerUp}
	for _, tc := range []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"a power-up activation", ManaSpendContext{Purpose: SpendPurposeActivate, PowerUp: true}, true},
		{"an ordinary activation", ManaSpendContext{Purpose: SpendPurposeActivate}, false},
		{"a spell", ManaSpendContext{Purpose: SpendPurposeCast}, false},
		{"a spell that wrongly claims the bit", ManaSpendContext{Purpose: SpendPurposeCast, PowerUp: true}, false},
		{"an unlock", ManaSpendContext{Purpose: SpendPurposeUnlock, PowerUp: true}, false},
		{"a payment that says nothing", ManaSpendContext{}, false},
		{"a payment that says nothing but the bit", ManaSpendContext{PowerUp: true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ctx.allows(tags); got != tc.want {
				t.Errorf("allows = %v, want %v", got, tc.want)
			}
		})
	}
}

// The tag rides ManaRestrictAnyOf like every other purpose tag, and a
// token carrying it pays only through ManaPool.CanPayFor's context.
func TestPowerUpOnlyManaIsSpentThroughTheSpendContext(t *testing.T) {
	cost, err := ParseCost("{R}{R}")
	if err != nil {
		t.Fatal(err)
	}
	pool := ManaPool{
		{Color: "R", Restrictions: []string{ManaRestrictPowerUp}},
		{Color: "R", Restrictions: []string{ManaRestrictPowerUp}},
	}
	if !pool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeActivate, PowerUp: true}) {
		t.Error("power-up-only mana does not pay a power-up activation")
	}
	if pool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeActivate}) {
		t.Error("power-up-only mana pays an ordinary activation")
	}
	if pool.CanPayFor(cost, 0, ManaSpendContext{}) {
		t.Error("power-up-only mana pays a payment that names no purpose")
	}

	either := ManaRestrictAnyOf([]string{ManaRestrictCast}, []string{ManaRestrictPowerUp})
	if !(ManaSpendContext{Purpose: SpendPurposeActivate, PowerUp: true}).allows([]string{either}) ||
		!(ManaSpendContext{Purpose: SpendPurposeCast}).allows([]string{either}) ||
		(ManaSpendContext{Purpose: SpendPurposeActivate}).allows([]string{either}) {
		t.Error("ManaRestrictAnyOf does not combine the power-up tag with a cast clause")
	}
}

func registerPowerUpBan(t *testing.T, g *Game) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeGame, uuid.Nil,
			[]Mod{{Kind: ModCantActivatePowerUp}}, g.UntilEndOfTurnDuration(), "Kang the Conqueror")
	})
	if !ok {
		t.Fatal("the power-up ban was not registered")
	}
}

// "Power-up abilities can't be activated": every power-up ability of
// every player, no other ability, only while the effect stands.
func TestAPowerUpBanRefusesOnlyPowerUpActivations(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "power-up-ban", powerUpProbe())
	src := pushPowerUpSource(g, me, "power-up-ban")
	card := *g.findCardByIDLocked(src)
	powerUp := ActivationAbility{Label: powerUpProbeLabel, PowerUp: true}
	plain := ActivationAbility{Label: "Not a power-up"}

	g.WithWriteLock(func() {
		if g.AnyActivationRestrictionsForEffect() {
			t.Error("AnyActivationRestrictionsForEffect before the ban")
		}
		if err := g.ActivationGateLocked(me.ID, card, ZoneBattlefield, powerUp); err != nil {
			t.Errorf("gate before the ban = %v", err)
		}
	})
	registerPowerUpBan(t, g)
	g.WithWriteLock(func() {
		if !g.AnyActivationRestrictionsForEffect() {
			t.Error("the fast negative misses the ban: the view and the enumerator would never ask the gate")
		}
		for _, who := range g.Seats {
			err := g.ActivationGateLocked(who.ID, card, ZoneBattlefield, powerUp)
			if !errors.Is(err, ErrCantActivate) {
				t.Errorf("gate for %s's power-up = %v, want ErrCantActivate", who.Name, err)
			}
		}
		if err := g.ActivationGateLocked(me.ID, card, ZoneBattlefield, plain); err != nil {
			t.Errorf("gate for an ordinary ability = %v, want nil", err)
		}
		if got := g.CantActivateReasonLocked(me.ID, card, ZoneBattlefield, powerUp); got == "" {
			t.Error("the view's reason is empty under the ban")
		}
	})

	// The ability itself is refused, and pays nothing.
	for _, c := range []string{"C", "C", "C", "C", "W"} {
		me.ManaPool.AddMana(ManaToken{Color: c})
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); !errors.Is(err, ErrCantActivate) {
		t.Fatalf("activation under the ban = %v, want ErrCantActivate", err)
	}
	if me.ManaPool.Total() != 5 {
		t.Errorf("pool = %v, want all five floating — a refused activation pays nothing", me.ManaPool)
	}
}

// The ban is an until-end-of-turn rules effect, so it ends at cleanup.
func TestAPowerUpBanEndsWithTheTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	exhaustCatalog(t, "power-up-ban-ends", powerUpProbe())
	src := pushPowerUpSource(g, me, "power-up-ban-ends")
	registerPowerUpBan(t, g)

	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	card := *g.findCardByIDLocked(src)
	g.WithWriteLock(func() {
		if err := g.ActivationGateLocked(me.ID, card, ZoneBattlefield, ActivationAbility{Label: powerUpProbeLabel, PowerUp: true}); err != nil {
			t.Errorf("gate the turn after = %v, want nil", err)
		}
		if g.AnyActivationRestrictionsForEffect() {
			t.Error("the ban outlived the turn")
		}
	})
}

// ActivationAbilityOf derives the bit from the shape, so no card file can
// forget it.
func TestActivationAbilityOfCarriesThePowerUpBit(t *testing.T) {
	if !ActivationAbilityOf(powerUpProbe()).PowerUp {
		t.Error("a power-up shape lost its bit on the way to the gate")
	}
	if ActivationAbilityOf(ActivatedAbilityShape{Label: "x"}).PowerUp {
		t.Error("an ordinary shape gained the power-up bit")
	}
}
