package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_trigger_test.go — #763 / ADR 0074: CR 605.1b triggered mana
// abilities, the one trigger kind that never uses the stack.
//
// Every test here is about a rule rather than a card; the cards that
// prove it are in cards/effects (Wild Growth, Overgrowth, Utopia
// Sprawl, Fertile Ground, Mana Flare).

const (
	wildGrowthTestOracle = "test-wild-growth"
	manaFlareTestOracle  = "test-mana-flare"
	forestTestOracle     = "test-forest"
)

// withCatalogManaTriggers installs a CatalogManaTriggers shim for the
// duration of a test, the way withCatalogTriggers does for the stack
// triggers.
func withCatalogManaTriggers(t *testing.T, fn func(key string) []ManaTrigger) {
	t.Helper()
	prev := CatalogManaTriggers
	CatalogManaTriggers = fn
	t.Cleanup(func() { CatalogManaTriggers = prev })
}

// attachedLandTapped is the "whenever enchanted land is tapped for
// mana" condition every Aura in the family shares.
func attachedLandTapped(prod ManaProduced, source *Card, _ *Game) bool {
	return source.IsAttachedTo(prod.SourceID())
}

// wildGrowthTrigger is the printed card as the engine reads it.
func wildGrowthTrigger(produced string) ManaTrigger {
	return ManaTrigger{
		Label:     "Wild Growth — add an additional " + produced,
		AppliesTo: attachedLandTapped,
		Produced: func(ManaProduced, *Card, *Game) string {
			return produced
		},
	}
}

// pushForest seeds a basic Forest under `owner`.
func pushForest(g *Game, owner *Player) uuid.UUID {
	return pushBattlefieldForTest(g, owner.ID, "Forest", "Basic Land — Forest", forestTestOracle)
}

// pushAuraOn seeds an Aura attached to `host`.
func pushAuraOn(g *Game, owner *Player, name, oracle string, host uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Enchantment — Aura",
		OracleID:   oracle,
		Owner:      owner.ID,
		Controller: owner.ID,
		AttachedTo: TargetRef{Kind: TargetCard, ID: host},
	})
	return id
}

// The headline: the trigger resolves as part of tapping the land, with
// no stack item and no priority window (CR 605.4a).
func TestManaTriggerResolvesWithoutTheStack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushForest(g, me)
	pushAuraOn(g, me, "Wild Growth", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}")}
		}
		return nil
	})

	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); got["G"] != 2 {
		t.Errorf("pool = %v, want the Forest's {G} and Wild Growth's", got)
	}
	if len(g.PendingTriggers) != 0 {
		t.Errorf("PendingTriggers = %d, want none: a mana ability does not use the stack", len(g.PendingTriggers))
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("stack = %d items, want none", len(g.StackMeta))
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("PendingChoices = %d, want none for a fixed-colour trigger", len(g.PendingChoices))
	}
}

// The extra mana is attributed to the TRIGGER's permanent: it comes
// from Wild Growth's ability, not from the land.
func TestManaTriggerMintsFromItsOwnSource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushForest(g, me)
	aura := pushAuraOn(g, me, "Overgrowth", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}{G}")}
		}
		return nil
	})

	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	fromAura := 0
	for _, tok := range me.ManaPool {
		if tok.Source == aura {
			fromAura++
		}
	}
	if fromAura != 2 {
		t.Errorf("%d tokens from the Aura, want the two it prints (pool %v)", fromAura, me.ManaPool)
	}
}

// A land that produces nothing was not tapped for mana, and nothing
// fires.
func TestManaTriggerDoesNotFireOnAnEmptyProduction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	empty := pushIntrinsicPermanent(g, me, "Uncharted Haven", "Land", []ManaAbilityShape{{
		TapCost:      true,
		ProducedFunc: func(*Game, uuid.UUID, uuid.UUID) string { return "" },
		Label:        "Add one mana of the chosen color",
	}}, nil)
	pushAuraOn(g, me, "Wild Growth", wildGrowthTestOracle, empty)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}")}
		}
		return nil
	})

	if err := g.ActivateManaAbility(me.ID, empty, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want nothing: the land added no mana, so it was not tapped for mana", me.ManaPool)
	}
}

// A source whose colour is a PICK fires exactly once, from
// ResolveManaChoice — the only place the produced colour is known —
// and not a second time at activation.
func TestManaTriggerFiresOnceThroughTheColourPick(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	dual := pushIntrinsicPermanent(g, me, "Test Dual", "Land",
		[]ManaAbilityShape{{TapCost: true, Produced: "{U|B}", Label: "Add {U} or {B}"}}, nil)
	pushAuraOn(g, me, "Wild Growth", wildGrowthTestOracle, dual)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}")}
		}
		return nil
	})

	if err := g.ActivateManaAbility(me.ID, dual, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 0 {
		t.Fatalf("pool = %v before the pick is answered, want nothing", got)
	}
	pick := choiceByKind(g, PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana pick queued for the dual land")
	}
	if !pick.ManaTapped {
		t.Error("the pick does not record that a permanent was tapped for mana")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	got := poolColors(me)
	if got["U"] != 1 || got["G"] != 1 || len(got) != 2 {
		t.Errorf("pool = %v, want exactly {U} and one Wild Growth {G}", got)
	}
}

// The produced COLOUR reaches the trigger: "one mana of any type that
// land produced" (Mana Flare, Mirari's Wake).
func TestManaTriggerReadsTheProducedColour(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	island := pushBattlefieldForTest(g, me.ID, "Island", "Basic Land — Island", "")
	pushBattlefieldForTest(g, me.ID, "Mana Flare", "Enchantment", manaFlareTestOracle)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key != manaFlareTestOracle {
			return nil
		}
		return []ManaTrigger{{
			Label: "Mana Flare",
			AppliesTo: func(prod ManaProduced, _ *Card, _ *Game) bool {
				return prod.Source.IsLand()
			},
			Produced: func(prod ManaProduced, _ *Card, _ *Game) string {
				return prod.ProducedColorPipe()
			},
		}}
	})

	if err := g.ActivateManaAbility(me.ID, island, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); got["U"] != 2 || len(got) != 1 {
		t.Errorf("pool = %v, want two blue: the Island's and the type it produced again", got)
	}
	// One type produced is one option, so no prompt is owed.
	if len(g.PendingChoices) != 0 {
		t.Errorf("PendingChoices = %d, want none: one produced type is not a choice", len(g.PendingChoices))
	}
	// And the colour rides the event, which is the other half of the
	// seam this closes.
	sawColor := false
	for _, ev := range g.Events {
		if ev.Kind == EventManaAdded && len(ev.Colors) == 1 && ev.Colors[0] == "U" {
			sawColor = true
		}
	}
	if !sawColor {
		t.Error("no EventManaAdded carried the colour it added")
	}
}

// A trigger whose output is a colour CHOICE (Fertile Ground's "any
// color") prompts on a hand-clicked activation.
func TestManaTriggerWithAColourChoicePromptsByHand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushForest(g, me)
	pushAuraOn(g, me, "Fertile Ground", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{W|U|B|R|G}")}
		}
		return nil
	})

	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := choiceByKind(g, PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana pick for the trigger's 'any color'")
	}
	if pick.ManaTapped {
		t.Error("the trigger's own pick claims a permanent was tapped for it — that would re-trigger")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	got := poolColors(me)
	if got["G"] != 1 || got["R"] != 1 {
		t.Errorf("pool = %v, want the Forest's {G} and the chosen {R}", got)
	}
	// Answering the trigger's own pick must not fire it again.
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %v, want exactly two mana — the trigger must not re-trigger", me.ManaPool)
	}
}

// …and picks greedily against the cast inside the auto-tapper, which
// may not leave a prompt open mid-cast.
func TestManaTriggerWithAColourChoicePicksGreedilyUnderAutoTap(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushForest(g, me)
	pushAuraOn(g, me, "Fertile Ground", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{W|U|B|R|G}")}
		}
		return nil
	})

	// The planner knows only about the Forest's own {G} (ADR 0074 §7),
	// so it plans the Forest for the {G} and the trigger's mana pays
	// the {U} that arrives with it.
	cost, _ := ParseCost("{G}")
	g.WithWriteLock(func() {
		plan, ok := g.autoTapLocked(me.ID, cost, 0, nil)
		if !ok {
			t.Fatalf("no plan for {G} off a Forest")
		}
		g.materializePlanLocked(me, plan, cost)
	})
	if len(g.PendingChoices) != 0 {
		t.Fatalf("PendingChoices = %d, want none: the auto-tapper asks no questions", len(g.PendingChoices))
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %v, want the Forest's {G} and the trigger's extra", me.ManaPool)
	}
}

// CR 106.12a: mana added by a resolving spell was not "tapped for
// mana", and nothing triggers off it.
func TestManaTriggerDoesNotFireOnAddManaForEffect(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushForest(g, me)
	pushAuraOn(g, me, "Wild Growth", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{{
				Label:     "Wild Growth",
				AppliesTo: func(ManaProduced, *Card, *Game) bool { return true },
				Produced:  func(ManaProduced, *Card, *Game) string { return "{G}" },
			}}
		}
		return nil
	})

	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(me.ID, forest, "{B}{B}{B}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	})
	if got := poolColors(me); got["B"] != 3 || len(got) != 1 {
		t.Errorf("pool = %v, want Dark Ritual's three black and nothing else", got)
	}
}

// A mana ability with no {T} is not "tapped for mana" either (Lotus
// Petal's sacrifice, Ashnod's Altar).
func TestManaTriggerDoesNotFireWithoutATap(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	petal := pushIntrinsicPermanent(g, me, "Lotus Petal", "Artifact", []ManaAbilityShape{{
		SacrificeCost: true,
		Produced:      "{G}",
		Label:         "Sacrifice: Add {G}",
	}}, nil)
	pushAuraOn(g, me, "Wild Growth", wildGrowthTestOracle, petal)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}")}
		}
		return nil
	})

	if err := g.ActivateManaAbility(me.ID, petal, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %v, want only the Petal's own {G}", me.ManaPool)
	}
}

// Two copies of the same Aura are two triggers; neither sees the
// other's mana, because nothing was tapped for it.
func TestTriggeredManaDoesNotRetrigger(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushForest(g, me)
	pushAuraOn(g, me, "Wild Growth", wildGrowthTestOracle, forest)
	pushAuraOn(g, me, "Wild Growth", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}")}
		}
		return nil
	})

	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); got["G"] != 3 {
		t.Errorf("pool = %v, want three green: the land and one per Aura", got)
	}
}

// CR 613.1f: an Aura whose abilities were removed has no mana trigger.
func TestManaTriggerGoesQuietWhenAbilitiesAreRemoved(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushForest(g, me)
	aura := pushAuraOn(g, me, "Wild Growth", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}")}
		}
		return nil
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != aura {
				continue
			}
			eff := c.printedCharacteristic()
			eff.AbilitiesRemoved = true
			c.effective = &eff
		}
	})

	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); got["G"] != 1 {
		t.Errorf("pool = %v, want only the Forest's {G}", got)
	}
}

// ADR 0071's gate applies here exactly as it does to a stack trigger:
// a gated-off mana trigger is never matched.
func TestManaTriggerRespectsItsDesignationGate(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushForest(g, me)
	aura := pushAuraOn(g, me, "Gated Growth", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key != wildGrowthTestOracle {
			return nil
		}
		gated := wildGrowthTrigger("{G}")
		gated.ActiveWhen = ClassLevel(2)
		return []ManaTrigger{gated}
	})

	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); got["G"] != 1 {
		t.Fatalf("pool = %v at level 1, want only the Forest's {G}", got)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == aura {
				g.Battlefield.Cards[i].ClassLevel = 2
			}
			if g.Battlefield.Cards[i].InstanceID == forest {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility (level 2): %v", err)
	}
	if got := poolColors(me); got["G"] != 3 {
		t.Errorf("pool = %v at level 2, want the gate open: two Forest {G} and one from the Aura", got)
	}
}

// The auto-tap executor is the third production site, and the trigger
// fires there too — otherwise the same board pays differently
// depending on whether the player clicked the land or pressed
// "Auto-tap & cast".
func TestManaTriggerFiresFromTheAutoTapExecutor(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushForest(g, me)
	pushAuraOn(g, me, "Overgrowth", wildGrowthTestOracle, forest)
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}{G}")}
		}
		return nil
	})

	cost, _ := ParseCost("{G}")
	g.WithWriteLock(func() {
		plan, ok := g.autoTapLocked(me.ID, cost, 0, nil)
		if !ok {
			t.Fatalf("no plan for {G} off a Forest")
		}
		g.materializePlanLocked(me, plan, cost)
	})
	if got := poolColors(me); got["G"] != 3 {
		t.Errorf("pool = %v, want the Forest's {G} and Overgrowth's two", got)
	}
}
