package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mill_cost_cards_test.go — #3849: the five printed mill-cost cards the
// #3845 component was waiting on, plus the two colours-spent readers
// regraded in #3850.

const (
	derangedAssistantOracle   = "9e1b0034-bd2c-4412-8d81-db3306c8814f"
	manakinAndMillikinOracle  = "ceda813a-7c15-47a9-a8a3-4724f1dfbf35"
	sinisterConcoctionOracle  = "ffc3beec-6485-428d-a251-fa0bc21263a5"
	rotFarmSkeletonOracle     = "f7b2c135-1b99-4b69-89d9-3fef25b509d4"
	charmedPendantOracle      = "12469cb5-a383-47ab-84cd-414b9ced8a8c"
	moonholdOracle            = "9b515bb8-6e37-4b2d-8d56-bb517c3b267c"
	millCostCardsCreatureType = "Artifact Creature — Construct"
)

func TestMillCostCardsAreGradedHonestly(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		caveats      bool
	}{
		{"Deranged Assistant", derangedAssistantOracle, false},
		{"Manakin and Millikin", manakinAndMillikinOracle, false},
		{"Sinister Concoction", sinisterConcoctionOracle, false},
		{"Rot Farm Skeleton", rotFarmSkeletonOracle, false},
		{"Charmed Pendant", charmedPendantOracle, false},
		{"Moonhold", moonholdOracle, false}, // complete since a waived payment asks (#3855)
		{"River's Grasp", riversGraspOracle, false},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok || spec.Name != tc.name {
			t.Errorf("%s: registered as %q (found %v)", tc.name, spec.Name, ok)
			continue
		}
		want := CompletenessFull
		if tc.caveats {
			want = CompletenessCaveats
		}
		if spec.Completeness != want {
			t.Errorf("%s: completeness %v, want %v", tc.name, spec.Completeness, want)
		}
		if tc.caveats != (len(spec.Caveats) == 1) {
			t.Errorf("%s: caveats %v", tc.name, spec.Caveats)
		}
	}
}

// activateMillMana activates the card's one activated ability (the
// mill-for-mana cards are not mana abilities, #3971, CR 605.1a), checks
// it went on the stack with nothing in the pool yet, then lets it resolve.
// It returns the activation's error and, when it succeeded, the pool size
// seen while the ability was still on the stack.
func activateMillMana(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID) (int, error) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
		return 0, err
	}
	onStack := me.ManaPool.Total()
	if len(g.StackMeta) != 1 {
		t.Fatalf("stack holds %d items after activation, want the one ability", len(g.StackMeta))
	}
	passPriorityAroundTable(t, g)
	return onStack, nil
}

// The two mill-for-colourless creatures: the card is milled and the
// source taps as the cost, the ability uses the stack, and the mana lands
// when it resolves. An empty library refuses the ability.
func TestDerangedAssistantAndManakinAndMillikinMillForColourless(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		want                   int
	}{
		{"Deranged Assistant", derangedAssistantOracle, "Creature — Human Wizard", 1},
		{"Manakin and Millikin", manakinAndMillikinOracle, millCostCardsCreatureType, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			src := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			lib := lcLibrary(g, me, 3)

			onStack, err := activateMillMana(t, g, me, src)
			if err != nil {
				t.Fatalf("activate: %v", err)
			}
			if onStack != 0 {
				t.Errorf("pool held %d while the ability was on the stack; not a mana ability, so the mana comes on resolution", onStack)
			}
			got := batch01PoolColors(me)
			if len(got) != tc.want {
				t.Fatalf("pool %v, want %d {C}", got, tc.want)
			}
			for _, c := range got {
				if c != "C" {
					t.Errorf("pool %v, want only colourless", got)
				}
			}
			if !me.Graveyard.Contains(lib[0]) || me.Library.Size() != 2 {
				t.Error("the top card was not milled")
			}
			if c, _ := battlefieldCard(g, src); !c.Tapped {
				t.Error("the source should be tapped")
			}

			// Nothing to mill: refused, and the source stays untapped.
			b02bSetTapped(g, src, false)
			g.WithWriteLock(func() { me.Library.Cards = nil })
			if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantPayLibraryCost) {
				t.Errorf("empty library: activate = %v, want ErrCantPayLibraryCost", err)
			}
			if c, _ := battlefieldCard(g, src); c.Tapped {
				t.Error("a refused activation tapped the source")
			}
		})
	}
}

func TestManakinAndMillikinIsSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushCatalogPermanent(g, me.ID, "Manakin and Millikin", millCostCardsCreatureType, manakinAndMillikinOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a summoning-sick creature paid its {T}")
	}
}

// CR 605.1a (#3971): a cost or effect that moves a card to or from a
// library takes an ability out of the mana abilities, so each of the five
// declares an activated ability and no mana ability. Back-out: restore a
// ManaAbilities row on one and this fails.
func TestLibraryMovingManaCardsAreNotManaAbilities(t *testing.T) {
	for _, oracle := range []string{
		derangedAssistantOracle, manakinAndMillikinOracle, charmedPendantOracle,
		b11MillikinOracle, samiChromaticSphere,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if len(spec.ManaAbilities) != 0 {
			t.Errorf("%s declares %d mana abilities; CR 605.1a says it has none", spec.Name, len(spec.ManaAbilities))
		}
		if len(spec.Activated) != 1 {
			t.Errorf("%s declares %d activated abilities, want 1", spec.Name, len(spec.Activated))
		}
	}
}

func sinisterTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, opp := exileCostTable(t)
	lcLibrary(g, me, 3)
	victim := pushCatalogPermanent(g, opp.ID, "Victim", "Creature — Bear", "test-victim", false)
	src := pushCatalogPermanent(g, me.ID, "Sinister Concoction", "Enchantment", sinisterConcoctionOracle, false)
	return g, me, victim, src
}

// All five components are paid at announce, the destroy happens on
// resolution, and the enchantment is already gone.
func TestSinisterConcoctionPaysFiveCostsAndDestroysTheCreature(t *testing.T) {
	g, me, victim, src := sinisterTable(t)
	discard := pushHandCard(g, me)
	floatMana(t, g, me, "{B}")
	life := me.Life
	top, _ := me.Library.Top()

	err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{
		Targets:    []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		DiscardIDs: []uuid.UUID{discard},
	})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	if !me.Graveyard.Contains(top.InstanceID) || !me.Graveyard.Contains(discard) || !me.Graveyard.Contains(src) {
		t.Error("the milled card, the discard and the enchantment should all be in the graveyard")
	}
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the creature died before the ability resolved")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target creature survived")
	}
}

// An empty library, or an empty hand, leaves everything unpaid.
func TestSinisterConcoctionRefusedWithAnUnpayableCostPaysNothing(t *testing.T) {
	for _, tc := range []string{"library", "hand"} {
		t.Run(tc, func(t *testing.T) {
			g, me, victim, src := sinisterTable(t)
			discard := pushHandCard(g, me)
			floatMana(t, g, me, "{B}")
			life := me.Life
			params := game.ActivateAbilityParams{
				Targets:    []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
				DiscardIDs: []uuid.UUID{discard},
			}
			g.WithWriteLock(func() {
				if tc == "library" {
					me.Library.Cards = nil
				} else {
					me.Hand.Cards = nil
				}
			})
			if err := g.ActivateCatalogAbility(me.ID, src, 0, params); err == nil {
				t.Fatal("activated with an unpayable cost")
			}
			if me.Life != life || !g.Battlefield.Contains(src) || me.ManaPool.Total() != 1 {
				t.Error("a refused activation paid part of its cost")
			}
		})
	}
}

func rotFarmTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := pushCatalogGraveyardCard(me, "Rot Farm Skeleton", "Creature — Plant Skeleton", rotFarmSkeletonOracle, 4, 1)
	return g, me, id
}

func TestRotFarmSkeletonReturnsFromTheGraveyardForFourMilledCards(t *testing.T) {
	g, me, id := rotFarmTable(t)
	lcLibrary(g, me, 6)
	floatMana(t, g, me, "{C}{C}{B}{G}")

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate from graveyard: %v", err)
	}
	if me.Library.Size() != 2 {
		t.Errorf("library %d, want 2 after milling four", me.Library.Size())
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the skeleton is not on the battlefield")
	}
	if c.Tapped {
		t.Error("it returns untapped")
	}
}

func TestRotFarmSkeletonNeedsFourCardsToMill(t *testing.T) {
	g, me, id := rotFarmTable(t)
	lcLibrary(g, me, 3)
	floatMana(t, g, me, "{C}{C}{B}{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with three cards in the library")
	}
	if me.Library.Size() != 3 || me.ManaPool.Total() != 4 {
		t.Error("a refused activation paid part of its cost")
	}
}

func TestRotFarmSkeletonIsSorcerySpeedAndGraveyardOnly(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: rotFarmSkeletonOracle})
	if len(abilities) != 1 {
		t.Fatalf("%d activated abilities, want 1", len(abilities))
	}
	ab := abilities[0]
	if !ab.SorcerySpeed || ab.Cost.Mill != 4 || ab.Cost.Mana != "{2}{B}{G}" {
		t.Errorf("cost %+v sorcery %v", ab.Cost, ab.SorcerySpeed)
	}
	if !game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) || game.AbilityFunctionsFromZone(ab, game.ZoneBattlefield) {
		t.Error("the ability should function from the graveyard only")
	}
}

func TestRotFarmSkeletonCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCatalogPermanent(g, me.ID, "Attacker", "Creature — Bear", "test-attacker", false)
	skeleton := pushCatalogPermanent(g, opp.ID, "Rot Farm Skeleton", "Creature — Plant Skeleton", rotFarmSkeletonOracle, false)
	plain := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Plain Blocker", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(skeleton, attacker); err == nil {
		t.Error("Rot Farm Skeleton blocked")
	}
	if err := g.DeclareBlocker(plain, attacker); err != nil {
		t.Errorf("control: an ordinary creature could not block: %v", err)
	}
}

// --- Charmed Pendant ------------------------------------------------

func pendantTable(t *testing.T, milledCost string) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushCatalogPermanent(g, me.ID, "Charmed Pendant", "Artifact", charmedPendantOracle, false)
	ids := lcLibrary(g, me, 2)
	id := ids[0]
	g.WithWriteLock(func() {
		for i := range me.Library.Cards {
			if me.Library.Cards[i].InstanceID == id {
				me.Library.Cards[i].ManaCost = milledCost
			}
		}
	})
	return g, me, src, id
}

func TestCharmedPendantAddsOneManaPerColouredSymbolOfTheMilledCard(t *testing.T) {
	g, me, src, milled := pendantTable(t, "{3}{U}{U}{B}")
	top, _ := me.Library.Top()
	if top.InstanceID != milled {
		t.Fatal("setup: the milled card is not on top")
	}
	if _, err := activateMillMana(t, g, me, src); err != nil {
		t.Fatalf("activate: %v", err)
	}
	got := map[string]int{}
	for _, c := range batch01PoolColors(me) {
		got[c]++
	}
	if len(got) != 2 || got["U"] != 2 || got["B"] != 1 {
		t.Errorf("pool %v, want {U}{U}{B}", got)
	}
	if !me.Graveyard.Contains(milled) {
		t.Error("the card was not milled")
	}
}

func TestCharmedPendantAddsNothingForAColourlessOrCostlessCard(t *testing.T) {
	for _, cost := range []string{"{4}", ""} {
		g, me, src, _ := pendantTable(t, cost)
		if _, err := activateMillMana(t, g, me, src); err != nil {
			t.Fatalf("cost %q: activate: %v", cost, err)
		}
		if n := me.ManaPool.Total(); n != 0 {
			t.Errorf("cost %q: pool holds %d, want none", cost, n)
		}
	}
}

// CR 701.17c: Rest in Peace exiles the milled card, and the ability that
// refers to it still finds it there, so the mana is added.
func TestCharmedPendantCountsACardRestInPeaceExiled(t *testing.T) {
	g, me, src, milled := pendantTable(t, "{1}{G}{G}")
	castRestInPeace(t, g)
	if _, err := activateMillMana(t, g, me, src); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(milled) {
		t.Fatal("setup: the milled card should be in exile")
	}
	got := batch01PoolColors(me)
	if len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool %v, want {G}{G}", got)
	}
}

func TestCharmedPendantNeedsALibrary(t *testing.T) {
	g, me, src, _ := pendantTable(t, "{G}")
	g.WithWriteLock(func() { me.Library.Cards = nil })
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantPayLibraryCost) {
		t.Errorf("empty library: activate = %v, want ErrCantPayLibraryCost", err)
	}
}

func TestCharmedPendantIsNeverAutoTapped(t *testing.T) {
	g, me, _, _ := pendantTable(t, "{G}")
	lib := me.Library.Size()
	cost, err := game.ParseCost("{G}")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Error("the auto-tapper planned Charmed Pendant")
	}
	if me.Library.Size() != lib {
		t.Error("a probe milled the library")
	}
}

func TestColoredSymbolsAsProduced(t *testing.T) {
	for cost, want := range map[string]string{
		"":             "",
		"{3}{U}{U}{B}": "{U}{U}{B}",
		"{X}{R}":       "{R}",
		"{C}{C}":       "",
		"{W/U}{W/U}":   "{W|U}{W|U}",
		"{2/W}{G/P}":   "{W}{G}",
		"{R/G/P}":      "{R|G}",
	} {
		if got := coloredSymbolsAsProduced(cost); got != want {
			t.Errorf("%q → %q, want %q", cost, got, want)
		}
	}
}

// PaidCost.Milled on the CR 602 path: a stacked ability's record names
// the cards its mill cost put in the graveyard (#3849).
func TestActivatedAbilityRecordsTheMilledCards(t *testing.T) {
	var got []uuid.UUID
	registerForTest(t, Spec{
		OracleID: "test-mc-record", Name: "Mill Recorder",
		Activated: []ActivatedAbility{{
			Label: "Mill Recorder: Note the milled cards.",
			Cost:  MillCardsCost(2),
			Effect: func(_ *game.Game, item *game.StackItem) error {
				got = append([]uuid.UUID(nil), item.Paid.Milled...)
				return nil
			},
		}},
	})
	g, me, _ := exileCostTable(t)
	lib := lcLibrary(g, me, 3)
	src := pushCatalogPermanent(g, me.ID, "Mill Recorder", "Enchantment", "test-mc-record", false)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(got) != 2 || got[0] != lib[0] || got[1] != lib[1] {
		t.Errorf("Milled = %v, want the top two cards %v", got, lib[:2])
	}
}
