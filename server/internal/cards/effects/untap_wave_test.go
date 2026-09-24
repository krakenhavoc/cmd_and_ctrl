package effects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestUntapWaveCardsAreRegisteredWithTheDeclaredSeam(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		static       bool
		full         bool
	}{
		{"Tangle", "f627e125-15af-4e53-b34e-82b60e4ec87b", false, true},
		{"Mana Vault", "736892cb-a34b-4bb9-b56c-e26e3db207a2", true, true}, // full since #1418
		{"Basalt Monolith", "6b8cf2a0-b045-4d91-9d91-c602d40c6237", true, true},
		{"Grim Monolith", "229d6627-1292-4ae1-8849-b0f956fa6540", true, true},
		{"Goblin Sharpshooter", "d81285b7-a718-411a-8be3-ecc0cfe0bcb0", true, true},
		{"Traxos, Scourge of Kroog", "c1c78144-b335-4d22-a668-9173ab6a0d04", true, true},
		{"Meekstone", "5ba73182-30a7-4bad-9cb6-c0feecc2db33", true, true},
		{"Back to Basics", "05c2dec2-d2f7-4036-b91f-4fccba10a8bb", true, true},
		{"Intruder Alarm", "1e943e04-e213-4781-b1a7-935aad8790e1", true, true},
		{"Claustrophobia", "62d8c8c8-bc24-42f2-9e2e-9efd08e47bb1", true, true},
		{"Wall of Frost", "741e4f32-0587-40fa-a73d-5bcf66b52348", false, true},
		{"Kefnet's Monument", "b6294891-79e6-4f2a-a82d-6cffce968356", false, true},
		{"Frost Breath", "382097b3-f753-493c-bde4-101c0538feb4", false, true},
		{"Sleep", "9b93ff69-f195-4d72-8e1d-574c3e53bca8", false, true},
		{"Alchemax Slayer-Bots", "cdc7356c-9621-4420-9fda-c8b91df0ec7c", false, true},
		{"Dreamdew Entrancer", "b6b2d63f-5b8c-47ee-8712-58596e0e9e94", false, true},
		{"Cryogen Relic", "106e9c0d-67a2-4db7-97d9-03e1ea1b40b5", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, ok := Lookup(tc.oracle)
			if !ok || spec.Name != tc.name || (spec.Completeness == CompletenessFull) != tc.full {
				t.Fatalf("Lookup(%q) = %#v, %v", tc.oracle, spec, ok)
			}
			if tc.static != (len(spec.UntapStepRestrictions) != 0) {
				t.Errorf("static restriction presence = %v, want %v", len(spec.UntapStepRestrictions) != 0, tc.static)
			}
		})
	}
}

func TestSleepActuallyTapsAndMarksOnlyTheTargetPlayerCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushPermanentForTest(g, me.ID, "Mine", "", "Creature — Bear")
	theirs := pushPermanentForTest(g, opp.ID, "Theirs", "", "Creature — Bear")
	castCatalogSpell(t, g, "Sleep", "Sorcery", "9b93ff69-f195-4d72-8e1d-574c3e53bca8", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id     uuid.UUID
		tapped bool
		marked bool
	}{{mine, false, false}, {theirs, true, true}} {
		c, _ := battlefieldCard(g, tc.id)
		if c.Tapped != tc.tapped || (len(c.NextUntapSkips) != 0) != tc.marked {
			t.Errorf("%s after Sleep = tapped %v markers %d", c.Name, c.Tapped, len(c.NextUntapSkips))
		}
	}
}

func TestTangleMarksAttackersButNotBlockersAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushPermanentForTest(g, me.ID, "Attacker", "", "Creature — Bear")
	blocker := pushPermanentForTest(g, opp.ID, "Blocker", "", "Creature — Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].AttackingTarget = opp.ID
			}
			if g.Battlefield.Cards[i].InstanceID == blocker {
				g.Battlefield.Cards[i].BlockingTarget = attacker
			}
		}
	})
	castCatalogSpell(t, g, "Tangle", "Instant", "f627e125-15af-4e53-b34e-82b60e4ec87b", nil)
	passPriorityAroundTable(t, g)
	a, _ := battlefieldCard(g, attacker)
	b, _ := battlefieldCard(g, blocker)
	if len(a.NextUntapSkips) != 1 {
		t.Errorf("attacker markers = %d, want 1", len(a.NextUntapSkips))
	}
	if len(b.NextUntapSkips) != 0 {
		t.Errorf("blocker markers = %d, want 0", len(b.NextUntapSkips))
	}
}

func TestStaticRestrictionCardsHoldTheirPermanentsAtTheUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vault := pushPermanentForTest(g, me.ID, "Mana Vault", "736892cb-a34b-4bb9-b56c-e26e3db207a2", "Artifact")
	ordinary := pushPermanentForTest(g, me.ID, "Ordinary Rock", "", "Artifact")
	if err := g.TapCard(vault, true); err != nil {
		t.Fatal(err)
	}
	if err := g.TapCard(ordinary, true); err != nil {
		t.Fatal(err)
	}
	advanceToUpkeepOf(t, g, 0)
	if c, _ := battlefieldCard(g, vault); !c.Tapped {
		t.Error("Mana Vault untapped during its controller's untap step")
	}
	if c, _ := battlefieldCard(g, ordinary); c.Tapped {
		t.Error("ordinary permanent stayed tapped")
	}
}

func TestFrostBreathResolvesByTappingAndMarkingTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushPermanentForTest(g, me.ID, "A", "", "Creature — Bear")
	b := pushPermanentForTest(g, me.ID, "B", "", "Creature — Bear")
	castCatalogSpell(t, g, "Frost Breath", "Instant", "382097b3-f753-493c-bde4-101c0538feb4", []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c, _ := battlefieldCard(g, id)
		if !c.Tapped || len(c.NextUntapSkips) != 1 {
			t.Errorf("target %s after Frost Breath = tapped %v markers %d", id, c.Tapped, len(c.NextUntapSkips))
		}
	}
}

func TestStunETBCardsPutTheirPrintedStunCountersOnTheTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Alchemax Slayer-Bots", "Artifact Creature — Robot", "cdc7356c-9621-4420-9fda-c8b91df0ec7c", 1},
		{"Dreamdew Entrancer", "Creature — Frog Wizard", "b6b2d63f-5b8c-47ee-8712-58596e0e9e94", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			target := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Target")
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			// Targeted ETB triggers choose the first legal target in the
			// sandbox picker, which is the target we seeded.
			passPriorityAroundTable(t, g)
			answerAnyPendingTargetPrompts(t, g)
			passPriorityAroundTable(t, g)
			c, _ := battlefieldCard(g, target)
			if c.Counters[game.CounterStun] != tc.n || !c.Tapped {
				t.Errorf("target after %s = tapped %v stun %d", tc.name, c.Tapped, c.Counters[game.CounterStun])
			}
		})
	}
}

func TestDreamdewEntrancerDrawsWhenItsETBTargetIsYours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := pushCreatureToBattlefieldForTest(g, me.ID, "Own target")
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Dreamdew Entrancer", "Creature — Frog Wizard", "b6b2d63f-5b8c-47ee-8712-58596e0e9e94", nil)
	passPriorityAroundTable(t, g)
	var prompt *game.PendingChoice
	for _, choice := range g.PendingChoices {
		if choice != nil && choice.Kind == game.PendingChoicePickTarget {
			prompt = choice
			break
		}
	}
	if prompt == nil {
		t.Fatal("Dreamdew Entrancer did not ask for its ETB target")
	}
	if err := g.ResolvePickTarget(prompt.ID, prompt.Chooser, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	// castCatalogSpell injects and then casts one card, so the net change
	// is the two cards drawn by Dreamdew's own-target clause.
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand after casting Dreamdew and drawing two: got %d, want %d", got, handBefore+2)
	}
	if c, _ := battlefieldCard(g, target); !c.Tapped || c.Counters[game.CounterStun] != 3 {
		t.Errorf("own target after Dreamdew: tapped=%v stun=%d", c.Tapped, c.Counters[game.CounterStun])
	}
}

func TestCryogenRelicDrawsOnEntryAndLeaving(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	handBefore := me.Hand.Size()
	relic := castCatalogSpell(t, g, "Cryogen Relic", "Artifact", "106e9c0d-67a2-4db7-97d9-03e1ea1b40b5", nil)
	passPriorityAroundTable(t, g)
	// The helper's injected card nets out; Cryogen's ETB contributes one.
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand after casting Cryogen and its ETB draw: got %d, want %d", got, handBefore+1)
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(relic); err != nil {
			t.Errorf("DestroyPermanentForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand after Cryogen leaves: got %d, want %d", got, handBefore+2)
	}
}

func TestCryogenRelicSacrificeStunsATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	relic := pushPermanentForTest(g, me.ID, "Cryogen Relic", "106e9c0d-67a2-4db7-97d9-03e1ea1b40b5", "Artifact")
	target := pushCreatureToBattlefieldForTest(g, me.ID, "Tapped target")
	if err := g.TapCard(target, true); err != nil {
		t.Fatal(err)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("Cryogen Relic remained on the battlefield after its sacrifice cost")
	}
	if c, _ := battlefieldCard(g, target); c.Counters[game.CounterStun] != 1 {
		t.Errorf("target stun counters = %d, want 1", c.Counters[game.CounterStun])
	}
}
