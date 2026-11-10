package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fabrication_foundry_test.go — #5401: "{2}{W}, {T}, Exile one or more
// other artifacts you control with total mana value X: Return target
// artifact card with mana value X or less from your graveyard to the
// battlefield. Activate only as a sorcery." X is announced with the
// activation (CR 601.2b via CR 602.2b), bounds the target (CR 601.2c),
// and the exile at CR 601.2h must add up to exactly X.

const fabricationFoundryOracle = "0bb20c36-689f-4a7c-80b5-a12bda821847"

// foundryTable seats Fabrication Foundry on seat 0 in its main phase,
// with artifacts of mana value 1 and 2 (and a 0-value token) beside it
// and artifact cards of mana value 2, 3 and 4 in the graveyard.
type foundryTable struct {
	g                    *game.Game
	me                   *game.Player
	foundry              uuid.UUID
	token, one, two      uuid.UUID
	gy2, gy3, gy4, gyBea uuid.UUID
}

func newFoundryTable(t *testing.T) foundryTable {
	t.Helper()
	g, me, _ := spendTable(t)
	ft := foundryTable{g: g, me: me}
	ft.foundry = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fabrication Foundry", OracleID: fabricationFoundryOracle,
		TypeLine: "Artifact", ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID,
	})
	art := func(name, cost string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Artifact", ManaCost: cost,
			Owner: me.ID, Controller: me.ID,
		})
	}
	ft.token = art("Treasure", "")
	ft.one = art("Bauble", "{1}")
	ft.two = art("Signet", "{2}")
	gy := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
		return id
	}
	ft.gy2 = gy("Two Drop Relic", "Artifact", "{2}")
	ft.gy3 = gy("Three Drop Relic", "Artifact", "{3}")
	ft.gy4 = gy("Four Drop Relic", "Artifact", "{4}")
	ft.gyBea = gy("Bear", "Creature — Bear", "{1}")
	return ft
}

func (ft foundryTable) activate(x int, exiled []uuid.UUID, target uuid.UUID) error {
	return ft.g.ActivateCatalogAbility(ft.me.ID, ft.foundry, 0, game.ActivateAbilityParams{
		Strict:            true,
		XValue:            x,
		ExilePermanentIDs: exiled,
		Targets:           []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
}

func foundryOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

func TestFabricationFoundryIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(fabricationFoundryOracle)
	if !ok {
		t.Fatal("Fabrication Foundry is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	if len(spec.Activated) != 1 || !spec.Activated[0].Cost.DemandsX() || !spec.Activated[0].SorcerySpeed {
		t.Error("the return ability does not announce an X or is not sorcery-speed")
	}
}

// Exiling artifacts worth 1 + 2 announces X = 3 and returns the
// three-drop; the exiled artifacts are in exile and the cost is paid.
func TestFabricationFoundryReturnsAnArtifactWithManaValueUpToTheExiledTotal(t *testing.T) {
	ft := newFoundryTable(t)
	floatMana(t, ft.g, ft.me, "{C}{C}{W}")
	b16Activate(t, ft.g, ft.me.ID, ft.foundry, 0, game.ActivateAbilityParams{
		Strict:            true,
		XValue:            3,
		ExilePermanentIDs: []uuid.UUID{ft.one, ft.two},
		Targets:           []game.TargetRef{{Kind: game.TargetCard, ID: ft.gy3}},
	})
	if !foundryOnBattlefield(ft.g, ft.gy3) {
		t.Error("the three-drop did not return to the battlefield")
	}
	if !ft.g.Exile.Contains(ft.one) || !ft.g.Exile.Contains(ft.two) {
		t.Error("the artifacts paid for the cost were not exiled")
	}
	if !b16Tapped(t, ft.g, ft.foundry) {
		t.Error("the Foundry was not tapped")
	}
}

// "Mana value X or less": X = 3 may return the two-drop too.
func TestFabricationFoundryReturnsLessThanX(t *testing.T) {
	ft := newFoundryTable(t)
	floatMana(t, ft.g, ft.me, "{C}{C}{W}")
	b16Activate(t, ft.g, ft.me.ID, ft.foundry, 0, game.ActivateAbilityParams{
		Strict: true, XValue: 3,
		ExilePermanentIDs: []uuid.UUID{ft.one, ft.two},
		Targets:           []game.TargetRef{{Kind: game.TargetCard, ID: ft.gy2}},
	})
	if !foundryOnBattlefield(ft.g, ft.gy2) {
		t.Error("the two-drop did not return at X = 3")
	}
}

// The target is bounded by the announced X: a four-drop at X = 3 is
// not a legal target, and the activation is refused with nothing paid.
// A creature card is not an artifact card.
func TestFabricationFoundryTargetIsBoundedByX(t *testing.T) {
	ft := newFoundryTable(t)
	floatMana(t, ft.g, ft.me, "{C}{C}{W}")
	if err := ft.activate(3, []uuid.UUID{ft.one, ft.two}, ft.gy4); err == nil {
		t.Fatal("a four-drop was a legal target at X = 3")
	}
	if err := ft.activate(3, []uuid.UUID{ft.one, ft.two}, ft.gyBea); err == nil {
		t.Fatal("a creature card was a legal target")
	}
	if !foundryOnBattlefield(ft.g, ft.one) || !foundryOnBattlefield(ft.g, ft.two) || b16Tapped(t, ft.g, ft.foundry) {
		t.Error("a refused activation paid part of its cost")
	}
	if ft.me.ManaPool.Total() != 3 {
		t.Errorf("a refused activation spent mana: %d left", ft.me.ManaPool.Total())
	}
}

// The exile must total exactly the announced X: 1 + 2 is not 4 (the
// activator announcing one number and paying another), and 2 is not 3.
func TestFabricationFoundryExileMustTotalExactlyX(t *testing.T) {
	ft := newFoundryTable(t)
	floatMana(t, ft.g, ft.me, "{C}{C}{W}")
	for _, c := range []struct {
		name   string
		x      int
		exiled []uuid.UUID
		target uuid.UUID
	}{
		{"total 3 announced as 4", 4, []uuid.UUID{ft.one, ft.two}, ft.gy4},
		{"total 2 announced as 3", 3, []uuid.UUID{ft.two}, ft.gy3},
		{"total 1 announced as 0", 0, []uuid.UUID{ft.one}, ft.gy2},
	} {
		err := ft.activate(c.x, c.exiled, c.target)
		if !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("%s: %v, want ErrInvalidParam", c.name, err)
		}
	}
	if !foundryOnBattlefield(ft.g, ft.one) || !foundryOnBattlefield(ft.g, ft.two) {
		t.Error("a refused activation exiled an artifact")
	}
}

// With no set of other artifacts totalling X there is no payment for X
// (CR 118.3): the Foundry itself is never one ("other"), an opponent's
// artifact is not "you control", and nothing at all is refused too.
func TestFabricationFoundryRefusedWhenNoArtifactsTotalX(t *testing.T) {
	ft := newFoundryTable(t)
	floatMana(t, ft.g, ft.me, "{C}{C}{W}")
	opp := ft.g.Seats[1]
	theirs := pushBattlefieldCardWithTimestamp(ft.g, game.Card{
		InstanceID: uuid.New(), Name: "Their Relic", TypeLine: "Artifact", ManaCost: "{4}",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := ft.activate(2, []uuid.UUID{ft.foundry}, ft.gy2); err == nil {
		t.Error("the Foundry paid its own cost")
	}
	if err := ft.activate(4, []uuid.UUID{theirs}, ft.gy4); err == nil {
		t.Error("an opponent's artifact paid the cost")
	}
	if err := ft.activate(3, nil, ft.gy3); err == nil {
		t.Error("the ability was activated exiling nothing")
	}
	// Only artifacts worth 0, 1 and 2 are mine, so no exile adds up to 4.
	if err := ft.activate(4, []uuid.UUID{ft.token, ft.one, ft.two}, ft.gy4); err == nil {
		t.Error("artifacts totalling 3 paid for X = 4")
	}
	if b16Tapped(t, ft.g, ft.foundry) {
		t.Error("a refused activation tapped the Foundry")
	}
}

// A token with no mana cost is worth zero: X = 0 returns nothing of
// mana value 1, and the zero-value exile adds nothing to a total.
func TestFabricationFoundryCountsATokenAsZero(t *testing.T) {
	ft := newFoundryTable(t)
	floatMana(t, ft.g, ft.me, "{C}{C}{W}")
	if err := ft.activate(1, []uuid.UUID{ft.token}, ft.gy2); err == nil {
		t.Error("a token counted as mana value 1")
	}
	b16Activate(t, ft.g, ft.me.ID, ft.foundry, 0, game.ActivateAbilityParams{
		Strict: true, XValue: 2,
		ExilePermanentIDs: []uuid.UUID{ft.token, ft.two},
		Targets:           []game.TargetRef{{Kind: game.TargetCard, ID: ft.gy2}},
	})
	if !foundryOnBattlefield(ft.g, ft.gy2) || foundryOnBattlefield(ft.g, ft.token) || foundryOnBattlefield(ft.g, ft.two) {
		t.Error("the token and the two-drop did not pay for X = 2")
	}
}

// Activate only as a sorcery (CR 602.5d): not in combat.
func TestFabricationFoundryIsSorcerySpeed(t *testing.T) {
	ft := newFoundryTable(t)
	advanceTo(t, ft.g, game.StepBeginCombat)
	floatMana(t, ft.g, ft.me, "{C}{C}{W}")
	if err := ft.activate(3, []uuid.UUID{ft.one, ft.two}, ft.gy3); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("activated in combat: %v, want ErrSorcerySpeedRequired", err)
	}
}

// The {W} pays for an artifact spell and an artifact's ability — the
// Foundry's own return among them — and not for anything else.
func TestFabricationFoundryManaPaysArtifactsOnly(t *testing.T) {
	ft := newFoundryTable(t)
	if err := ft.g.ActivateManaAbility(ft.me.ID, ft.foundry, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Foundry for mana: %v", err)
	}
	cost, _ := game.ParseCost("{W}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"}
	if !ft.me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) {
		t.Error("refused an artifact spell")
	}
	if !ft.me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("refused an ability of an artifact source")
	}
	if ft.me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("paid for a creature spell")
	}
	if ft.me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("paid for a creature's ability")
	}
	// A real cast: the restricted {W} pays for a white artifact spell
	// and is refused for a white creature spell.
	relic := handSpell(ft.me, "White Relic", "Artifact", "{W}")
	knight := handSpell(ft.me, "White Knight", "Creature — Knight", "{W}")
	refusedForMana(t, ft.g.CastSpell(ft.me.ID, knight, game.CastSpellParams{Strict: true}), "a creature spell with the Foundry's {W}")
	if err := ft.g.CastSpell(ft.me.ID, relic, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the Foundry's {W} did not pay for an artifact spell: %v", err)
	}
}

// The form is an activated ability's open count, and it claims the
// announced X: a mana ability announces no X (CR 605.3b), a fixed count
// is not "one or more", and the cost cannot also pay an {X}.
func TestRegisterRefusesTheExileTotalWhereItCannotWork(t *testing.T) {
	ret := func(*game.Game, *game.StackItem) error { return nil }
	target := func() *game.TargetSpec {
		return TargetCardInGraveyard("target artifact card with mana value X or less from your graveyard", Artifact(), YouOwn()).
			WithManaValueAtMostX()
	}
	cases := map[string]func() Spec{
		"on a mana ability": func() Spec {
			return Spec{OracleID: "00000000-0000-0000-0000-0000000ff001", Name: "Ff Mana",
				ManaAbilities: []ManaAbility{{
					Cost:     ManaAbilityCost{ExilePermanents: ExileArtifactsWithTotalManaValueX("x").ExilePermanents},
					Produced: "{C}",
				}}}
		},
		"beside {X} in the mana cost": func() Spec {
			return Spec{OracleID: "00000000-0000-0000-0000-0000000ff002", Name: "Ff Mana X",
				Activated: []ActivatedAbility{{
					Label:   "{X}, Exile artifacts with total mana value X: Return.",
					Cost:    Plus(ManaCost("{X}"), ExileArtifactsWithTotalManaValueX("x")),
					Targets: target(),
					Effect:  ret,
				}}}
		},
		"without an open count": func() Spec {
			cost := ExileArtifactsWithTotalManaValueX("x")
			cost.ExilePermanents.OrMore = false
			return Spec{OracleID: "00000000-0000-0000-0000-0000000ff003", Name: "Ff Fixed",
				Activated: []ActivatedAbility{{
					Label: "Exile an artifact with total mana value X: Return.", Cost: cost, Targets: target(), Effect: ret,
				}}}
		},
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("Register accepted a declaration it must refuse")
				}
			}()
			Register(build())
		})
	}
}
