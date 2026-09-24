package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// paused_exit_tap_cost_test.go — #1427, the non-moving half of #1445.
//
// A commander that an effect has destroyed sits on the battlefield
// while its owner answers CR 903.9, but as far as the rules are
// concerned it has already left. #1451 refused every cost that would
// MOVE it; this file covers the costs that only TAP it — its own {T}
// (activated and mana), crew, tap-N-untapped-permanents on both
// ability kinds, and an activated ability's waterbend. None of them
// pays the same cost twice; each spends an object the destroy has
// already taken. The cast path's convoke and the proof cards are in
// cards/effects/paused_exit_tap_cost_test.go.

// tapMarkAbility is "{T}: mark".
func tapMarkAbility() ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label: "{T}: mark",
		Cost:  AbilityCost{Tap: true},
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}
}

// tapForGreen is Birds of Paradise without the pick: "{T}: Add {G}."
func tapForGreen() []ManaAbilityShape {
	return []ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "{T}: Add {G}"}}
}

// withTapAbilities gives a creature "{T}: mark" and "{T}: Add {G}".
func withTapAbilities(c *Card) {
	c.ActivatedAbilities = []ActivatedAbilityShape{tapMarkAbility()}
	c.ManaAbilities = tapForGreen()
}

// pushCrewVehicle seats a Vehicle with "Crew 2: mark".
func pushCrewVehicle(g *Game, owner *Player) uuid.UUID {
	c := NewCard("Test Vehicle", owner.ID)
	c.TypeLine = "Artifact — Vehicle"
	c.Controller = owner.ID
	c.Power, c.Toughness = 3, 3
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "Crew 2",
		Cost:  AbilityCost{Crew: 2},
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// tapCostTable is the board every case below starts from: a destroyed
// commander with a {T} ability and a {T} mana ability, its owner's
// CR 903.9 prompt open, and one source per tap-another shape to name
// it to.
type tapCostTable struct {
	g        *Game
	me       *Player
	cmdr     uuid.UUID
	station  uuid.UUID
	bender   uuid.UUID
	vehicle  uuid.UUID
	drum     uuid.UUID
	prompt   *PendingChoice
	fixtures []uuid.UUID
}

func newTapCostTable(t *testing.T) *tapCostTable {
	t.Helper()
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	tb := &tapCostTable{g: g, me: me}
	tb.station = pushTapOthersAbilitySource(g, me, AbilityCost{TapOthers: tapOthersStationCost()})
	tb.bender = pushTapOthersAbilitySource(g, me, waterbendAbilityCost("{1}"))
	tb.vehicle = pushCrewVehicle(g, me)
	tb.drum = pushIntrinsicPermanent(g, me, "Drum", "Artifact", []ManaAbilityShape{{
		TapCost:   true,
		TapOthers: tapOthersStationCost(),
		Produced:  "{C}",
		Label:     "{T}, Tap an untapped creature you control: Add {C}",
	}}, nil)
	tb.cmdr = seatCommander(t, g.Battlefield, me)
	editBattlefieldCard(g, tb.cmdr, withTapAbilities)
	tb.fixtures = []uuid.UUID{tb.cmdr, tb.station, tb.bender, tb.vehicle, tb.drum}
	tb.prompt = destroyCommanderPaused(t, g, me, tb.cmdr)
	return tb
}

// A destroyed commander, its owner still deciding, cannot be TAPPED to
// pay anything: not its own {T} ability or {T} mana ability, not a
// crew, not a tap-another cost on either ability kind, not a
// waterbend. Each refusal taps nothing, pays nothing and leaves the
// destroy's prompt alone. The same payment still goes through with an
// ordinary creature in the same slot, prompt open — the guard is about
// the card, not the table — and after the answer the commander is
// where the destroy sent it.
func TestADestroyedCommanderCannotBeTappedForACostWhileAsked(t *testing.T) {
	for _, tc := range []struct {
		name string
		try  func(tb *tapCostTable, id uuid.UUID) error
	}{
		{"its own {T} ability", func(tb *tapCostTable, id uuid.UUID) error {
			return tb.g.ActivateCatalogAbility(tb.me.ID, id, 0, ActivateAbilityParams{})
		}},
		{"its own {T} mana ability", func(tb *tapCostTable, id uuid.UUID) error {
			return tb.g.ActivateManaAbility(tb.me.ID, id, 0, ManaAbilityParams{})
		}},
		{"a crew", func(tb *tapCostTable, id uuid.UUID) error {
			return tb.g.ActivateCatalogAbility(tb.me.ID, tb.vehicle, 0, ActivateAbilityParams{CrewIDs: []uuid.UUID{id}})
		}},
		{"a tap-another activated cost", func(tb *tapCostTable, id uuid.UUID) error {
			return tb.g.ActivateCatalogAbility(tb.me.ID, tb.station, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{id}})
		}},
		{"a tap-another mana cost", func(tb *tapCostTable, id uuid.UUID) error {
			return tb.g.ActivateManaAbility(tb.me.ID, tb.drum, 0, ManaAbilityParams{TapIDs: []uuid.UUID{id}})
		}},
		{"a waterbend", func(tb *tapCostTable, id uuid.UUID) error {
			return tb.g.ActivateCatalogAbility(tb.me.ID, tb.bender, 0, ActivateAbilityParams{WaterbendIDs: []uuid.UUID{id}, Strict: true})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tb := newTapCostTable(t)
			g, me := tb.g, tb.me
			if err := tc.try(tb, tb.cmdr); !errors.Is(err, ErrChoicePending) {
				t.Errorf("naming the destroyed commander: err = %v, want ErrChoicePending", err)
			}
			for _, id := range tb.fixtures {
				if tappedIn(g, id) {
					t.Errorf("the refused payment tapped %v", id)
				}
			}
			if len(g.StackMeta) != 0 || len(me.ManaPool) != 0 {
				t.Errorf("stack %d, pool %v — the refused payment paid something", len(g.StackMeta), me.ManaPool)
			}
			if len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != tb.prompt.ID {
				t.Fatalf("the refusal disturbed the destroy's prompt (%d pending)", len(g.PendingChoices))
			}

			bear := pushTapOthersPermanent(g, me, "Bear", "Creature — Bear", false)
			editBattlefieldCard(g, bear, withTapAbilities)
			if err := tc.try(tb, bear); err != nil {
				t.Errorf("naming an ordinary creature while the prompt is open: %v", err)
			}
			if !tappedIn(g, bear) {
				t.Error("the ordinary creature did not tap")
			}

			if err := g.ResolveOptionalReplacement(tb.prompt.ID, me.ID, false); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			assertOnlyIn(t, tb.cmdr, me.Graveyard, g.Battlefield, me.Command)
		})
	}
}

// The auto-tapper never activates through ActivateManaAbility, so the
// announcement gate does not reach it: it plans a destroyed commander's
// "{T}: Add {G}" unless it asks the same question itself. Checked
// through the executor handed a stale plan, the planner, and a real
// auto-tapped activation, which must fail short rather than tap it.
func TestAutoTapSkipsADestroyedCommandersTapAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	cmdr := seatCommander(t, g.Battlefield, me)
	editBattlefieldCard(g, cmdr, func(c *Card) { c.ManaAbilities = tapForGreen() })
	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0); !ok || !containsID(plan, cmdr) {
		t.Fatalf("premise: the commander's {T} ability should pay {1} before anything happens (plan %v, ok %v)", plan, ok)
	}

	prompt := destroyCommanderPaused(t, g, me, cmdr)

	// A plan made before the destroy arrives stale; the executor must
	// drop the source rather than tap it.
	g.WithWriteLock(func() { g.materializePlanLocked(me, tapPlan{{CardID: cmdr}}, costFor(t, "{1}")) })
	if len(me.ManaPool) != 0 || tappedIn(g, cmdr) {
		t.Fatalf("a stale plan tapped the destroyed commander (pool %v)", me.ManaPool)
	}
	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0); ok {
		t.Errorf("auto-tap planned %v with the destroyed commander as the only source", plan)
	}
	src := pushReturnCostSource(g, me, AbilityCost{Mana: "{1}"})
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{AutoTap: true}); err == nil {
		t.Error("an auto-tapped {1} activation was paid by tapping the destroyed commander")
	}
	if tappedIn(g, cmdr) || len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != prompt.ID {
		t.Fatal("the auto-tapper tapped the commander or disturbed its prompt")
	}

	// An ordinary tapper is still a source with the prompt open.
	elf := pushIntrinsicPermanent(g, me, "Elf", "Creature — Elf", tapForGreen(), nil)
	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{1}"), 0); !ok || !containsID(plan, elf) || containsID(plan, cmdr) {
		t.Errorf("plan %v (ok %v), want the Elf and not the commander", plan, ok)
	}
}
