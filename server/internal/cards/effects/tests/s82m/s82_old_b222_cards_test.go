package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b222_cards_test.go — Ixalan, batch 222 (#5532). Each test plays the
// real catalog entry and asserts what its text says.

const (
	oldB222Captain  = "68f1da0d-1143-457a-b709-1da3c060a45d"
	oldB222Sentinel = "bfd2ef5d-618d-4062-9cb2-9dbef5432865"
	oldB222Spurring = "8bba7c7d-6110-4d7f-8b28-3c5b9d3448df"
	oldB222Keeper   = "86f2890d-3e5a-48db-bb22-a6cb8cbd39ce"
	oldB222Knight   = "ae5fe3bd-b6cf-47be-bfb0-05204387bb83"
	oldB222Skymarch = "f7ff0046-8365-4ea4-9872-583d140f7b3f"
	oldB222Zeal     = "cc7dcb89-534b-4e86-abd2-8362855e2f66"
)

func TestOldB222CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB222Captain: "Dire Fleet Captain", oldB222Sentinel: "Jace's Sentinel", oldB222Spurring: "Huatli's Spurring",
		oldB222Keeper: "Encampment Keeper", oldB222Knight: "Tilonalli's Knight", oldB222Skymarch: "Duskborne Skymarcher",
		oldB222Zeal: "Vampire's Zeal",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB222Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB222Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB222Attack(g *game.Game, defender uuid.UUID, ids ...uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			for _, id := range ids {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].AttackingTarget = defender
				}
			}
		}
	})
}

func TestOldB222DireFleetCaptainCountsOtherAttackingPirates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB222Seats(g)
	cap := effectstest.B12Push(g, me.ID, "Dire Fleet Captain", "Creature — Orc Pirate", oldB222Captain, 2, 2)
	p1 := effectstest.B12Creature(g, me.ID, "Pirate One", "Creature — Human Pirate", 1, 1)
	effectstest.B12Creature(g, me.ID, "Idle Pirate", "Creature — Human Pirate", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, cap, p1, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, cap); p != 3 || tt != 3 {
		t.Errorf("the Captain is %d/%d, want 3/3 (one other attacking Pirate)", p, tt)
	}
}

func TestOldB222VampiresZealGivesFirstStrikeOnlyToAVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB222Seats(g)
	bat := effectstest.B12Creature(g, opp.ID, "Bat", "Creature — Vampire", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
	_ = me
	effectstest.CastCatalogSpell(t, g, "Vampire's Zeal", "Instant", oldB222Zeal, oldB222Target(bat))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bat); p != 3 || tt != 3 || !effectstest.HasEffectiveKeyword(t, g, bat, "first strike") {
		t.Errorf("the Vampire is %d/%d, want 3/3 with first strike", p, tt)
	}
	effectstest.CastCatalogSpell(t, g, "Vampire's Zeal", "Instant", oldB222Zeal, oldB222Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 || effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Errorf("the Bear is %d/%d, want 3/3 without first strike", p, tt)
	}
}

func TestOldB222HuatlisSpurringIsFourOnlyWithAHuatliPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB222Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Huatli's Spurring", "Instant", oldB222Spurring, oldB222Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 1 {
		t.Fatalf("without a Huatli the Bear is %d/%d, want 3/1", p, tt)
	}
	pw := effectstest.B12Permanent(g, me.ID, "Huatli, Warrior Poet", "Legendary Planeswalker — Huatli")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pw {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLoyalty: 3}
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Huatli's Spurring", "Instant", oldB222Spurring, oldB222Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 7 {
		t.Errorf("with a Huatli the second spell gives +4/+0 instead: power %d, want 7", p)
	}
}

func TestOldB222TilonallisKnightNeedsADinosaur(t *testing.T) {
	for _, withDino := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB222Seats(g)
		knight := effectstest.B12Push(g, me.ID, "Tilonalli's Knight", "Creature — Human Knight", oldB222Knight, 2, 2)
		if withDino {
			effectstest.B12Creature(g, me.ID, "Dino", "Creature — Dinosaur", 1, 1)
		}
		effectstest.DeclareAttack(t, g, opp.ID, knight)
		effectstest.PassPriorityAroundTable(t, g)
		want := 2
		if withDino {
			want = 3
		}
		if p, tt := effectstest.PtOf(t, g, knight); p != want || tt != want {
			t.Errorf("dinosaur=%v: the Knight is %d/%d, want %d/%d", withDino, p, tt, want, want)
		}
	}
}

func TestOldB222JacesSentinelNeedsAJaceToGrowAndBeUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB222Seats(g)
	sen := effectstest.B12Push(g, me.ID, "Jace's Sentinel", "Creature — Merfolk Warrior", oldB222Sentinel, 1, 3)
	if p, _ := effectstest.PtOf(t, g, sen); p != 1 {
		t.Fatalf("without a Jace its power is %d, want 1", p)
	}
	effectstest.B12Permanent(g, opp.ID, "Jace, Their Walker", "Legendary Planeswalker — Jace")
	if p, _ := effectstest.PtOf(t, g, sen); p != 1 {
		t.Error("an opponent's Jace gave the bonus")
	}
	jace := effectstest.B12Permanent(g, me.ID, "Jace, My Walker", "Legendary Planeswalker — Jace")
	if p, tt := effectstest.PtOf(t, g, sen); p != 2 || tt != 3 {
		t.Errorf("with a Jace it is %d/%d, want 2/3", p, tt)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(jace) })
	g.RunStateChecksForTest()
	if p, _ := effectstest.PtOf(t, g, sen); p != 1 {
		t.Error("the bonus outlived the Jace")
	}
}

func TestOldB222EncampmentKeeperPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB222Seats(g)
	keeper := effectstest.B12Push(g, me.ID, "Encampment Keeper", "Creature — Dog", oldB222Keeper, 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.B10AddMana(me, "W", "W", "W", "W", "W", "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(keeper) {
		t.Error("the Keeper was not sacrificed")
	}
	if p, tt := effectstest.PtOf(t, g, mine); p != 4 || tt != 4 {
		t.Errorf("mine is %d/%d, want 4/4", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("theirs is %d/%d, want 2/2", p, tt)
	}
}

func TestOldB222DuskborneSkymarcherPumpsOnlyAnAttackingVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB222Seats(g)
	sky := effectstest.B12Push(g, me.ID, "Duskborne Skymarcher", "Creature — Vampire Cleric", oldB222Skymarch, 1, 1)
	idle := effectstest.B12Creature(g, me.ID, "Idle Vampire", "Creature — Vampire", 2, 2)
	atkBear := effectstest.B12Creature(g, me.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	atkVamp := effectstest.B12Creature(g, me.ID, "Attacking Vampire", "Creature — Vampire", 2, 2)
	oldB222Attack(g, opp.ID, atkBear, atkVamp)
	effectstest.AdvanceToMain(t, g)
	effectstest.B10AddMana(me, "W", "W")
	for _, bad := range []uuid.UUID{idle, atkBear} {
		if err := g.ActivateCatalogAbility(me.ID, sky, 0, game.ActivateAbilityParams{Targets: oldB222Target(bad), Strict: true}); err == nil {
			t.Fatal("targeted a creature that is not an attacking Vampire")
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, sky, 0, game.ActivateAbilityParams{Targets: oldB222Target(atkVamp), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, atkVamp); p != 3 || tt != 3 {
		t.Errorf("the attacking Vampire is %d/%d, want 3/3", p, tt)
	}
}
