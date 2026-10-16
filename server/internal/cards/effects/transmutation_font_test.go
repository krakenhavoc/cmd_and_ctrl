package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #3940 — Transmutation Font's tutor: "{3}, {T}, Sacrifice three
// artifact tokens with different names: Search your library for an
// artifact card, put it onto the battlefield, then shuffle. Activate
// only as a sorcery." The cost's "with different names" is a rule over
// the SET of picks (TargetSpec.Different on the sacrifice clause).

const fontTutor = 3

// fontTable puts the Font on the active seat's battlefield at a main
// phase, with {3} floating and an artifact and a sorcery in the library.
func fontTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	font := pushCatalogPermanent(g, me.ID, "Transmutation Font", "Artifact", b32TransmutationFontOracle, false)
	pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	pushLibraryCardForTest(me, game.Card{Name: "Sol Ring", TypeLine: "Artifact"})
	fillPool(me, 3)
	return g, me, font
}

// fontTutorPayments is every sacrifice payment the enumerator offers for
// the tutor.
func fontTutorPayments(t *testing.T, g *game.Game, seat, font uuid.UUID) [][]uuid.UUID {
	t.Helper()
	var out [][]uuid.UUID
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != "activate_ability" || m.Source != font {
			continue
		}
		var p struct {
			AbilityIndex int      `json:"ability_index"`
			SacrificeIDs []string `json:"sacrifice_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params: %v", err)
		}
		if p.AbilityIndex != fontTutor {
			continue
		}
		var ids []uuid.UUID
		for _, s := range p.SacrificeIDs {
			ids = append(ids, uuid.MustParse(s))
		}
		out = append(out, ids)
	}
	return out
}

func TestTransmutationFontTutorsForThreeTokensWithDifferentNames(t *testing.T) {
	g, me, font := fontTable(t)
	clues := pushTokens(g, me.ID, ClueToken(), 2)
	blood := pushToken(g, me.ID, BloodToken())
	food := pushToken(g, me.ID, FoodToken())

	b16Activate(t, g, me.ID, font, fontTutor, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{clues[0], blood, food}})
	ringSettle(t, g)
	takeSearchPrompts(t, g)
	for _, id := range []uuid.UUID{clues[0], blood, food} {
		if g.Battlefield.Contains(id) {
			t.Error("a paid token is still on the battlefield")
		}
	}
	if !g.Battlefield.Contains(clues[1]) {
		t.Error("the unpaid Clue was sacrificed")
	}
	if findBattlefieldByName(g, "Sol Ring") == uuid.Nil {
		t.Error("the artifact card is put onto the battlefield")
	}
}

// Two Clues and a Blood are three artifact tokens but only two names.
// The refusal names the rule and moves nothing: a payment is all or
// nothing.
func TestTransmutationFontRefusesTwoTokensWithOneName(t *testing.T) {
	g, me, font := fontTable(t)
	clues := pushTokens(g, me.ID, ClueToken(), 2)
	blood := pushToken(g, me.ID, BloodToken())
	pushToken(g, me.ID, FoodToken())

	err := g.ActivateCatalogAbility(me.ID, font, fontTutor, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{clues[0], clues[1], blood}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two Clues: err = %v, want ErrIllegalTarget", err)
	}
	for _, id := range append(clues, blood) {
		if !g.Battlefield.Contains(id) {
			t.Error("a refused payment sacrificed something")
		}
	}
	if b16Tapped(t, g, font) {
		t.Error("a refused payment tapped the Font")
	}
}

// The clause is "artifact TOKENS": a nontoken artifact with a third name
// does not pay it.
func TestTransmutationFontRefusesANontokenArtifact(t *testing.T) {
	g, me, font := fontTable(t)
	clue := pushToken(g, me.ID, ClueToken())
	blood := pushToken(g, me.ID, BloodToken())
	stone := b12Permanent(g, me.ID, "Mind Stone", "Artifact")

	err := g.ActivateCatalogAbility(me.ID, font, fontTutor, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{clue, blood, stone}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a nontoken artifact: err = %v, want ErrIllegalTarget", err)
	}
}

// CR 201.2b: objects have different names only if each has a name. A
// token with no name is no part of the payment, even beside two named
// ones, and it is no candidate either.
func TestTransmutationFontRefusesATokenWithNoName(t *testing.T) {
	g, me, font := fontTable(t)
	clue := pushToken(g, me.ID, ClueToken())
	blood := pushToken(g, me.ID, BloodToken())
	nameless := BloodToken()
	nameless.Name = ""
	unnamed := pushToken(g, me.ID, nameless)

	err := g.ActivateCatalogAbility(me.ID, font, fontTutor, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{clue, blood, unnamed}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a nameless token: err = %v, want ErrIllegalTarget", err)
	}
	if pays := fontTutorPayments(t, g, me.ID, font); len(pays) != 0 {
		t.Errorf("the tutor was offered with two named tokens and a nameless one: %v", pays)
	}
}

// The enumerator's payment is three names even when the first three of
// the payment order share one, and with only two names on the board the
// tutor is not offered at all.
func TestTransmutationFontEnumeratorPaysThreeNames(t *testing.T) {
	g, me, font := fontTable(t)
	clues := pushTokens(g, me.ID, ClueToken(), 3)
	blood := pushToken(g, me.ID, BloodToken())
	if pays := fontTutorPayments(t, g, me.ID, font); len(pays) != 0 {
		t.Fatalf("offered with two names on the board: %v", pays)
	}

	food := pushToken(g, me.ID, FoodToken())
	pays := fontTutorPayments(t, g, me.ID, font)
	if len(pays) != 1 || len(pays[0]) != 3 {
		t.Fatalf("payments %v, want one payment of three", pays)
	}
	seen := map[uuid.UUID]bool{}
	clueCount := 0
	for _, id := range pays[0] {
		seen[id] = true
		for _, c := range clues {
			if id == c {
				clueCount++
			}
		}
	}
	if clueCount != 1 || !seen[blood] || !seen[food] {
		t.Errorf("payment %v, want one Clue, the Blood and the Food", pays[0])
	}

	// And the offered payment really pays.
	if err := g.ActivateCatalogAbility(me.ID, font, fontTutor, game.ActivateAbilityParams{SacrificeIDs: pays[0]}); err != nil {
		t.Fatalf("activating the enumerated payment: %v", err)
	}
}

// The view ships each option's name as its key under `different`, which
// is what greys a second Clue in the client's picker.
func TestTransmutationFontViewShipsTheNames(t *testing.T) {
	g, me, font := fontTable(t)
	clues := pushTokens(g, me.ID, ClueToken(), 2)
	blood := pushToken(g, me.ID, BloodToken())
	b12Permanent(g, me.ID, "Mind Stone", "Artifact")

	row := activatedRowOf(t, g, me.ID, font, fontTutor)
	opts := row.SacrificeOptions
	if opts == nil {
		t.Fatal("the tutor row carries no sacrifice options")
	}
	if opts.Min != 3 || opts.Max != 3 {
		t.Errorf("bounds %d..%d, want 3..3", opts.Min, opts.Max)
	}
	if len(opts.Cards) != 3 {
		t.Errorf("options %v, want the three tokens (not the Mind Stone)", opts.Cards)
	}
	if opts.Different == nil || opts.Different.Label != "each have a different name" {
		t.Fatalf("different = %+v, want the name rule", opts.Different)
	}
	keys := opts.Different.Keys
	if keys[clues[0].String()] != "clue" || keys[clues[1].String()] != "clue" || keys[blood.String()] != "blood" {
		t.Errorf("keys = %v, want each token's name", keys)
	}
}

// Register keeps the rule to the shape its readers read: an activated
// ability's fixed count of two or more, without a kind rule beside it,
// and off every other sacrifice site.
func TestSacrificeDifferentRegistrationGuard(t *testing.T) {
	bad := func(name string, f func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s: accepted", name)
			}
		}()
		f()
	}
	fresh := func() *game.TargetSpec {
		return SacrificeNDifferent(3, "three artifact tokens with different names", EachDifferentName(), Artifact(), IsTokenPredicate()).SacrificeOther
	}
	bad("a count of one", func() {
		s := fresh()
		s.Min, s.Max = 1, 1
		checkSacrificeClause("Test", "ability 0", s, true, true, false)
	})
	bad("an open count", func() {
		s := fresh()
		s.Min, s.Max = 1, 0
		checkSacrificeClause("Test", "ability 0", s, true, true, false)
	})
	bad("no key", func() {
		s := fresh()
		s.Different = &game.TargetDifference{Label: "each have a different name"}
		checkSacrificeClause("Test", "ability 0", s, true, true, false)
	})
	bad("beside a kind rule", func() {
		s := fresh()
		s.EachOf = []game.SacrificeKind{SacrificeSubtype("a Clue", "Clue"), SacrificeSubtype("a Food", "Food"), SacrificeSubtype("a Blood", "Blood")}
		checkSacrificeClause("Test", "ability 0", s, true, true, false)
	})
	bad("on a mana ability", func() { refuseSacrificeDifferent("Test", "mana ability 0", fresh()) })

	// The Font's own clause passes.
	checkSacrificeClause("Test", "ability 0", fresh(), true, true, false)
	refuseSacrificeDifferent("Test", "mana ability 0", SacrificeN(3, "three artifacts", Artifact()).SacrificeOther)
}
