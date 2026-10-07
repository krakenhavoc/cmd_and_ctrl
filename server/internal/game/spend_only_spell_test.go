package game

import "testing"

// spend_only_spell_test.go — #2556, ADR 0040's 2026-10-07 amendment:
// the engine half of "spend only …" on a SPELL's own cost. The cards
// that prove it are in cards/effects/spend_only_spells_test.go.

// A basic land's mana is basic-land mana; a nonbasic land's, a
// Treasure's and a creature's are not — and a land is a land either way.
func TestManaSourceKindsSeparateBasicLands(t *testing.T) {
	basic := Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
	nonbasic := Card{Name: "Breeding Pool", TypeLine: "Land — Forest Island"}
	snow := Card{Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest"}
	dryad := Card{Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad"}
	if k := manaSourceKindsOf(basic); !k.Has(ManaSourceLand|ManaSourceBasicLand) || k.HasAny(ManaSourceCreature) {
		t.Errorf("basic Forest kinds = %b", k)
	}
	if k := manaSourceKindsOf(snow); !k.Has(ManaSourceLand | ManaSourceBasicLand | ManaSourceSnow) {
		t.Errorf("snow-covered Forest kinds = %b", k)
	}
	if k := manaSourceKindsOf(nonbasic); !k.Has(ManaSourceLand) || k.HasAny(ManaSourceBasicLand) {
		t.Errorf("Breeding Pool kinds = %b, want a land that is not basic", k)
	}
	if k := manaSourceKindsOf(dryad); !k.Has(ManaSourceLand|ManaSourceCreature) || k.HasAny(ManaSourceBasicLand) {
		t.Errorf("Dryad Arbor kinds = %b, want a nonbasic land creature", k)
	}
}

// A payment restricted by source takes a token only if its recorded
// source is one of the kinds; the token's own restrictions still apply.
func TestSourceOnlyContextFiltersTokensBySource(t *testing.T) {
	ctx := ManaSpendContext{Purpose: SpendPurposeCast, SourceOnly: ManaSourceBasicLand}
	basic := ManaToken{Color: "G", SourceKinds: ManaSourceLand | ManaSourceBasicLand}
	ring := ManaToken{Color: "C", SourceKinds: ManaSourceArtifact}
	spell := ManaToken{Color: "G"}
	restricted := ManaToken{Color: "G", SourceKinds: ManaSourceBasicLand, Restrictions: []string{"type:Artifact"}}
	if !ctx.allowsToken(basic) {
		t.Error("basic-land mana was refused")
	}
	if ctx.allowsToken(ring) || ctx.allowsToken(spell) {
		t.Error("mana from a Sol Ring or with no recorded source was accepted")
	}
	if ctx.allowsToken(restricted) {
		t.Error("a token's own restriction stopped applying")
	}
	if !(ManaSpendContext{Purpose: SpendPurposeCast}).allowsToken(spell) {
		t.Error("with no source restriction every unrestricted token is allowed")
	}

	pool := ManaPool{basic, ring, spell}
	cost, err := ParseCost("{G}{G}")
	if err != nil {
		t.Fatal(err)
	}
	if pool.CanPayFor(cost, 0, ctx) {
		t.Error("a pool with one basic-land mana paid {G}{G} under a basic-land restriction")
	}
	if !pool.CanPayFor(cost, 0, ManaSpendContext{Purpose: SpendPurposeCast}) {
		t.Error("the same pool should pay {G}{G} with no restriction")
	}
}

// Delve cannot eat the restricted X: its mana is coloured, and the fold
// has to happen before X is settled into generic.
func TestDelveLeavesARestrictedXAlone(t *testing.T) {
	cost, err := ParseCost("{X}{1}{B}")
	if err != nil {
		t.Fatal(err)
	}
	cost.SpendOnly = &ManaSpendOnly{Colors: []string{"B"}, XOnly: true}
	if got := delveBudget(cost, 3); got != 1 {
		t.Errorf("budget = %d, want 1 — only the {1}, not the black X", got)
	}
	out := delveAdjusted(cost, 5, 3)
	if out.Generic != 0 || out.XSlots != 0 || out.SpendOnly != nil || len(out.Required) != 4 {
		t.Fatalf("delved cost = %+v, want the {1} gone and X=3 folded to three black requirements plus {B}", out)
	}
	for _, r := range out.Required {
		if len(r.Options) != 1 || r.Options[0] != "B" {
			t.Errorf("requirement %+v is not black", r)
		}
	}
}
