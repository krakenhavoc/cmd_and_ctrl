package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kozilek_discard_mana_value_test.go — #2190: "Discard a card with mana
// value X: Counter target spell with mana value X." The discard cost
// whose card is the announcement (game.DiscardCost.ManaValueX, ADR
// 0113's 2026-10-07 amendment) paired with a spell target clause
// bounded by the same X (WithManaValueEqualsX).

const kozilekTheGreatDistortionOracle = "4c1c1537-e519-4e2f-9bc2-d34b289d4487"

// kozilekTable is a table where seat 1 holds Kozilek and a hand of
// mana values 0, 1, 3 and 4 (and one card whose joined split cost the
// engine cannot read), and seat 0 — the active seat — is about to cast.
func kozilekTable(t *testing.T) (g *game.Game, caster, me *game.Player, kozilek uuid.UUID, hand map[int]uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	caster, me = g.Seats[0], g.Seats[1]
	kozilek = pushCatalogPermanent(g, me.ID, "Kozilek, the Great Distortion", "Legendary Creature — Eldrazi",
		kozilekTheGreatDistortionOracle, false)
	hand = map[int]uuid.UUID{
		0: ccHandCard(me, "Their Land", "Basic Land — Forest", ""),
		1: ccHandCard(me, "Their One", "Instant", "{G}"),
		3: ccHandCard(me, "Their Three", "Instant", "{2}{G}"),
		4: ccHandCard(me, "Their Four", "Instant", "{3}{G}"),
	}
	return g, caster, me, kozilek, hand
}

func kozilekTarget(spell uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: spell}}
}

func kozilekActivate(g *game.Game, me *game.Player, kozilek uuid.UUID, x int, discard uuid.UUID, spell uuid.UUID) error {
	return g.ActivateCatalogAbility(me.ID, kozilek, 0, game.ActivateAbilityParams{
		XValue:     x,
		DiscardIDs: []uuid.UUID{discard},
		Targets:    kozilekTarget(spell),
	})
}

// The whole card: a mana-value-3 card pays to counter a mana-value-3
// spell, the cost is paid with the ability on the stack, and the spell
// does not resolve.
func TestKozilekCountersASpellOfTheDiscardedValue(t *testing.T) {
	g, _, me, kozilek, hand := kozilekTable(t)
	spell := castXSpell(t, g, "Their Bolt", "Sorcery", "", "{2}{R}", 0, nil)
	if !g.Stack.Contains(spell) {
		t.Fatal("setup: the spell should be on the stack")
	}
	if err := kozilekActivate(g, me, kozilek, 3, hand[3], spell); err != nil {
		t.Fatalf("activate at X=3: %v", err)
	}
	if me.Hand.Contains(hand[3]) || !me.Graveyard.Contains(hand[3]) {
		t.Error("the discarded card is not in the graveyard")
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell was countered before the ability resolved")
	}
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) {
		t.Error("the spell is still on the stack")
	}
	if !g.Seats[0].Graveyard.Contains(spell) {
		t.Error("a countered spell goes to its owner's graveyard")
	}
}

// The announcement, the card and the target have to agree, and a
// refusal leaves everything where it was.
func TestKozilekRefusesACardOrTargetThatDisagreesWithX(t *testing.T) {
	g, _, me, kozilek, hand := kozilekTable(t)
	spell := castXSpell(t, g, "Their Bolt", "Sorcery", "", "{2}{R}", 0, nil)
	handBefore := me.Hand.Size()
	for _, tc := range []struct {
		why     string
		x       int
		discard uuid.UUID
		wantErr error
	}{
		{"X=3 announced over a mana value 4 card", 3, hand[4], game.ErrInvalidParam},
		{"X=4 announced over a mana value 3 card", 4, hand[3], game.ErrInvalidParam},
		{"X=0 announced over a mana value 3 card", 0, hand[3], game.ErrInvalidParam},
		{"a mana value 1 card for a mana value 3 spell", 1, hand[1], game.ErrIllegalTarget},
		{"a land for a mana value 3 spell", 0, hand[0], game.ErrIllegalTarget},
		{"a card that is not in hand", 3, uuid.New(), nil},
	} {
		err := kozilekActivate(g, me, kozilek, tc.x, tc.discard, spell)
		if err == nil {
			t.Errorf("%s: activation accepted", tc.why)
			continue
		}
		if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
			t.Errorf("%s: err = %v, want %v", tc.why, err, tc.wantErr)
		}
		if me.Hand.Size() != handBefore || len(g.StackMeta) != 1 {
			t.Fatalf("%s: a refused activation paid something (hand %d, stack items %d)", tc.why, me.Hand.Size(), len(g.StackMeta))
		}
	}
	// Naming two cards, or none, is not "a card".
	for _, ids := range [][]uuid.UUID{nil, {hand[3], hand[4]}} {
		err := g.ActivateCatalogAbility(me.ID, kozilek, 0, game.ActivateAbilityParams{
			XValue: 3, DiscardIDs: ids, Targets: kozilekTarget(spell),
		})
		if err == nil {
			t.Errorf("%d cards named: activation accepted", len(ids))
		}
	}
}

// A spell on the stack counts {X} as the value chosen for it (CR
// 202.3e), so a Fireball cast for X=3 is mana value 4: the four pays,
// the one that a text-only reading of {X}{R} would suggest does not.
func TestKozilekReadsTheXOfASpellOnTheStack(t *testing.T) {
	g, _, me, kozilek, hand := kozilekTable(t)
	spell := castXSpell(t, g, "Their Fireball", "Sorcery", "", "{X}{R}", 3, nil)
	if err := kozilekActivate(g, me, kozilek, 1, hand[1], spell); err == nil {
		t.Fatal("a mana value 1 card countered a mana value 4 spell")
	}
	// The value the view ships for the client to narrow the picker by is
	// the same one: the stack's, not the printed cost's.
	g.ReadSnapshot(func() {
		spec := TargetSpell("target spell with mana value X").WithManaValueEqualsX()
		if got := g.BoundValuesForEffect(spec, []uuid.UUID{spell})[spell]; got != 4 {
			t.Errorf("shipped mana value of the X=3 Fireball = %d, want 4", got)
		}
	})
	if err := kozilekActivate(g, me, kozilek, 4, hand[4], spell); err != nil {
		t.Fatalf("a mana value 4 card for the X=3 Fireball: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) || !g.Seats[0].Graveyard.Contains(spell) {
		t.Error("the Fireball was not countered")
	}
}

// X may be zero: a land pays to counter a spell with no mana cost.
func TestKozilekZeroPaysForAManaValueZeroSpell(t *testing.T) {
	g, _, me, kozilek, hand := kozilekTable(t)
	spell := castCatalogSpell(t, g, "Their Free Spell", "Sorcery", "", nil)
	if err := kozilekActivate(g, me, kozilek, 0, hand[0], spell); err != nil {
		t.Fatalf("a land for a mana value 0 spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) || !g.Seats[0].Graveyard.Contains(spell) {
		t.Error("the mana value 0 spell was not countered")
	}
}

// A card whose cost the engine cannot read has no mana value to
// announce, so it cannot pay; nothing else is affected.
func TestKozilekCannotDiscardACardWithAnUnreadableCost(t *testing.T) {
	g, _, me, kozilek, _ := kozilekTable(t)
	split := ccHandCard(me, "Fire // Ice", "Instant // Instant", "{1}{R} // {1}{U}")
	spell := castXSpell(t, g, "Their Bolt", "Sorcery", "", "{1}{R}", 0, nil)
	for x := 0; x <= 4; x++ {
		if err := kozilekActivate(g, me, kozilek, x, split, spell); err == nil {
			t.Errorf("X=%d: a card with an unreadable cost paid", x)
		}
	}
	if !me.Hand.Contains(split) {
		t.Error("the refused activation discarded the split card")
	}
}

// With nothing on the stack there is no spell to target, so the
// ability cannot be activated at all.
func TestKozilekNeedsASpellToCounter(t *testing.T) {
	g, _, me, kozilek, hand := kozilekTable(t)
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, kozilek, 0, game.ActivateAbilityParams{
		XValue: 3, DiscardIDs: []uuid.UUID{hand[3]},
	}); err == nil {
		t.Fatal("activated with no target")
	}
	if me.Hand.Size() != handBefore {
		t.Error("a refused activation discarded a card")
	}
}

// The card is complete: the cast trigger, menace and the counter
// ability, with no caveat left.
func TestKozilekIsComplete(t *testing.T) {
	spec, ok := Lookup(kozilekTheGreatDistortionOracle)
	if !ok {
		t.Fatal("Kozilek is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v with caveats %v, want full and none", spec.Completeness, spec.Caveats)
	}
}

// The form is an activated ability's. A mana ability announces no X
// (CR 605.3b), and the cost cannot also pay for an {X} or a count.
func TestRegisterRefusesTheManaValueDiscardWhereItCannotWork(t *testing.T) {
	cases := map[string]func() Spec{
		"on a mana ability": func() Spec {
			return Spec{OracleID: "00000000-0000-0000-0000-00000000kz01", Name: "Kz Mana",
				ManaAbilities: []ManaAbility{{
					Cost:     ManaAbilityCost{DiscardCards: &game.DiscardCost{N: 1, Label: "a card", ManaValueX: true}},
					Produced: "{C}",
				}}}
		},
		"beside {X} in the mana cost": func() Spec {
			return Spec{OracleID: "00000000-0000-0000-0000-00000000kz02", Name: "Kz Mana X",
				Activated: []ActivatedAbility{{
					Label:   "{X}, Discard a card with mana value X: Counter target spell with mana value X.",
					Cost:    Plus(ManaCost("{X}"), DiscardCardWithManaValueX("a card with mana value X")),
					Targets: TargetSpell("target spell with mana value X").WithManaValueEqualsX(),
					Effect:  func(*game.Game, *game.StackItem) error { return nil },
				}}}
		},
		"with a predicate": func() Spec {
			cost := DiscardCardWithManaValueX("a card with mana value X")
			cost.DiscardCards.Match = func(game.Card) bool { return true }
			return Spec{OracleID: "00000000-0000-0000-0000-00000000kz03", Name: "Kz Match",
				Activated: []ActivatedAbility{{
					Label:   "Discard a card with mana value X: Counter target spell with mana value X.",
					Cost:    cost,
					Targets: TargetSpell("target spell with mana value X").WithManaValueEqualsX(),
					Effect:  func(*game.Game, *game.StackItem) error { return nil },
				}}}
		},
		"beside discarding this card": func() Spec {
			cost := DiscardCardWithManaValueX("a card with mana value X")
			cost.DiscardSelf = true
			return Spec{OracleID: "00000000-0000-0000-0000-00000000kz04", Name: "Kz Self",
				Activated: []ActivatedAbility{{
					Label:   "Discard a card with mana value X: Counter target spell with mana value X.",
					Cost:    cost,
					Targets: TargetSpell("target spell with mana value X").WithManaValueEqualsX(),
					Effect:  func(*game.Game, *game.StackItem) error { return nil },
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
