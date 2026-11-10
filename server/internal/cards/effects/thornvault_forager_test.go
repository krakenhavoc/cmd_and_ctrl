package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// thornvault_forager_test.go — #5397: forage as a MANA ability's cost
// (CR 605.1a, CR 701.61a), on the real catalog card, Thornvault Forager
// (Bloomburrow, #4467):
//
//	"{T}: Add {G}.
//	 {T}, Forage: Add two mana in any combination of colors.
//	 {3}{G}, {T}: Search your library for a Squirrel card, reveal it,
//	 put it into your hand, then shuffle."

const thornvaultForagerOracle = "7b18cfad-a1cf-46c2-bcdd-7762c13d50e5"

// tfForageRow is the forage mana ability's index.
const tfForageRow = 1

func TestThornvaultForagerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(thornvaultForagerOracle)
	if !ok {
		t.Fatal("Thornvault Forager is not registered")
	}
	if spec.Name != "Thornvault Forager" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
	if len(spec.ManaAbilities) != 2 || !spec.ManaAbilities[tfForageRow].Cost.Forage || spec.ManaAbilities[0].Cost.Forage {
		t.Error("want a plain {T}: Add {G} and a {T}, Forage row")
	}
}

// tfForager is Thornvault Forager under me, untapped and able to tap.
func tfForager(g *game.Game, me *game.Player) uuid.UUID {
	return pushCatalogPermanent(g, me.ID, "Thornvault Forager", "Creature — Squirrel Ranger", thornvaultForagerOracle, false)
}

func tfForage(g *game.Game, me *game.Player, forager uuid.UUID, branch *int, exile, sac []uuid.UUID) error {
	return g.ActivateManaAbility(me.ID, forager, tfForageRow, game.ManaAbilityParams{
		CostBranch: branch, ExileIDs: exile, SacrificeIDs: sac, Colors: []string{"R", "U"},
	})
}

// tfForageMoves is the forage row's enumerated moves for seat, keyed by
// branch.
func tfForageMoves(t *testing.T, g *game.Game, seat, forager uuid.UUID) map[int]legal.Move {
	t.Helper()
	out := map[int]legal.Move{}
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeActivateManaAbility || m.Source != forager {
			continue
		}
		var p struct {
			AbilityIndex int  `json:"ability_index"`
			CostBranch   *int `json:"cost_branch"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		switch {
		case p.AbilityIndex == tfForageRow && p.CostBranch != nil:
			out[*p.CostBranch] = m
		case p.AbilityIndex == tfForageRow:
			t.Errorf("forage move %q names no branch", m.Label)
		case p.CostBranch != nil:
			t.Errorf("plain move %q names branch %d", m.Label, *p.CostBranch)
		}
	}
	return out
}

func tfPoolHas(t *testing.T, me *game.Player, want ...string) {
	t.Helper()
	got := poolColors(me)
	if len(got) != len(want) {
		t.Fatalf("pool = %v, want %v", got, want)
	}
	have := map[string]int{}
	for _, c := range got {
		have[c]++
	}
	for _, c := range want {
		if have[c] == 0 {
			t.Fatalf("pool = %v, want %v", got, want)
		}
		have[c]--
	}
}

// Branch 0: three cards from the graveyard pay the forage. Exactly those
// three leave, the Forager taps, two mana of the named colours land, and
// one forage is announced.
func TestThornvaultForagerForagesByExilingThree(t *testing.T) {
	g, me, _ := eoTable(t)
	forager := tfForager(g, me)
	gy := fgGraveyard(me, 4)
	zero := 0
	if err := tfForage(g, me, forager, &zero, gy[:3], nil); err != nil {
		t.Fatalf("forage: %v", err)
	}
	tfPoolHas(t, me, "R", "U")
	if !isTapped(g, forager) {
		t.Error("the Forager did not tap")
	}
	for _, id := range gy[:3] {
		if !exiled(g, id) {
			t.Errorf("graveyard card %s was not exiled", id)
		}
	}
	if !inGraveyard(me, gy[3]) {
		t.Error("the fourth graveyard card left; forage exiles exactly three")
	}
	if n := fgForages(g, me.ID); n != 1 {
		t.Errorf("%d forage events, want 1", n)
	}
	if g.Stack.Size() != 0 {
		t.Error("a mana ability used the stack (CR 605.3b)")
	}
}

// Branch 1: a Food pays. The two mana may be the same colour ("any
// combination"), and "whenever you forage" (Corpseberry Cultivator)
// triggers off a forage paid for mana.
func TestThornvaultForagerForagesBySacrificingAFood(t *testing.T) {
	g, me, _ := eoTable(t)
	forager := tfForager(g, me)
	cultivator := pushCatalogPermanent(g, me.ID, "Corpseberry Cultivator", "Creature — Squirrel Warlock", blbFgCorpseberryOracle, false)
	food := fgFoodToken(t, g, me.ID)
	gy := fgGraveyard(me, 2)
	one := 1
	if err := g.ActivateManaAbility(me.ID, forager, tfForageRow, game.ManaAbilityParams{
		CostBranch: &one, SacrificeIDs: []uuid.UUID{food}, Colors: []string{"B", "B"},
	}); err != nil {
		t.Fatalf("forage: %v", err)
	}
	tfPoolHas(t, me, "B", "B")
	if onBattlefield(g, food) {
		t.Error("the Food was not sacrificed")
	}
	if !inGraveyard(me, gy[0]) || !inGraveyard(me, gy[1]) {
		t.Error("the Food branch exiled graveyard cards")
	}
	if n := fgForages(g, me.ID); n != 1 {
		t.Errorf("%d forage events, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if n := fgPlusOnes(t, g, cultivator); n != 1 {
		t.Errorf("Corpseberry Cultivator has %d +1/+1 counters, want 1 from the forage", n)
	}
}

// Two graveyard cards and no Food of mine: forage can't be paid either
// way (CR 118.3, "If you don't have enough cards in your graveyard or a
// Food on the battlefield, you can't choose to forage"). No forage move
// is offered, every attempt is refused with nothing paid, and the plain
// {G} row is still there.
func TestThornvaultForagerRefusedWithTwoCardsAndNoFood(t *testing.T) {
	g, me, opp := eoTable(t)
	forager := tfForager(g, me)
	gy := fgGraveyard(me, 2)
	oppFood := fgFoodToken(t, g, opp.ID)
	if got := tfForageMoves(t, g, me.ID, forager); len(got) != 0 {
		t.Fatalf("offered forage branches %v, want none", got)
	}
	zero, one := 0, 1
	for name, err := range map[string]error{
		"two cards":          tfForage(g, me, forager, &zero, gy, nil),
		"no branch":          tfForage(g, me, forager, nil, gy, nil),
		"an opponent's Food": tfForage(g, me, forager, &one, nil, []uuid.UUID{oppFood}),
	} {
		if err == nil {
			t.Errorf("%s: activation accepted", name)
		}
	}
	if isTapped(g, forager) || len(poolColors(me)) != 0 || !inGraveyard(me, gy[0]) || !inGraveyard(me, gy[1]) ||
		!onBattlefield(g, oppFood) || fgForages(g, me.ID) != 0 {
		t.Error("a refused activation paid something")
	}
	if err := g.ActivateManaAbility(me.ID, forager, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("plain {T}: Add {G}: %v", err)
	}
	tfPoolHas(t, me, "G")
}

// The strict refusals: a branch is required on the forage row and
// refused on the plain one, it must be one the cost has, and the cards
// or the Food named must be the branch's own. Each is refused before
// anything is paid.
func TestThornvaultForagerStrictRefusals(t *testing.T) {
	g, me, _ := eoTable(t)
	forager := tfForager(g, me)
	gy := fgGraveyard(me, 3)
	food := fgFoodToken(t, g, me.ID)
	bauble := pushCatalogPermanent(g, me.ID, "Bauble", "Artifact", "", false)
	hand := handSpell(me, "Shock-ish", "Instant", "{R}")
	zero, one, two, minus := 0, 1, 2, -1
	cases := []struct {
		name    string
		err     error
		wantErr error
	}{
		{"no branch", tfForage(g, me, forager, nil, gy, nil), game.ErrAbilityCostBranch},
		{"branch 2", tfForage(g, me, forager, &two, gy, nil), game.ErrAbilityCostBranch},
		{"branch -1", tfForage(g, me, forager, &minus, gy, nil), game.ErrAbilityCostBranch},
		{"a branch on the plain row", g.ActivateManaAbility(me.ID, forager, 0, game.ManaAbilityParams{CostBranch: &zero}), game.ErrAbilityCostBranch},
		{"graveyard branch naming the Food", tfForage(g, me, forager, &zero, nil, []uuid.UUID{food}), nil},
		{"graveyard branch naming two cards", tfForage(g, me, forager, &zero, gy[:2], nil), nil},
		{"graveyard branch naming a hand card", tfForage(g, me, forager, &zero, []uuid.UUID{gy[0], gy[1], hand}, nil), nil},
		{"Food branch naming the graveyard", tfForage(g, me, forager, &one, gy, nil), nil},
		{"Food branch naming a non-Food artifact", tfForage(g, me, forager, &one, nil, []uuid.UUID{bauble}), nil},
		{"Food branch naming the Food and the cards", tfForage(g, me, forager, &one, gy, []uuid.UUID{food}), nil},
	}
	for _, c := range cases {
		if c.err == nil {
			t.Errorf("%s: activation accepted", c.name)
			continue
		}
		if c.wantErr != nil && !errors.Is(c.err, c.wantErr) {
			t.Errorf("%s: %v, want %v", c.name, c.err, c.wantErr)
		}
	}
	if isTapped(g, forager) || len(poolColors(me)) != 0 || !onBattlefield(g, food) || fgForages(g, me.ID) != 0 {
		t.Fatal("a refused activation paid something")
	}
	for _, id := range gy {
		if !inGraveyard(me, id) {
			t.Fatal("a refused activation exiled a graveyard card")
		}
	}
}

// The auto-tapper never forages. With the Forager the only source, a
// Food and a full graveyard beside it, a {G} spell is paid with the
// plain row, and a {G}{G} or {R} spell — which only a forage could pay —
// is neither offered nor castable with auto-tap.
func TestThornvaultForagerAutoTapperNeverForages(t *testing.T) {
	g, me, _ := eoTable(t)
	forager := tfForager(g, me)
	food := fgFoodToken(t, g, me.ID)
	gy := fgGraveyard(me, 3)

	two := handSpell(me, "Two Green", "Instant", "{G}{G}")
	red := handSpell(me, "One Red", "Instant", "{R}")
	for name, id := range map[string]uuid.UUID{"{G}{G}": two, "{R}": red} {
		if castMove(g, me.ID, id) {
			t.Errorf("%s is offered: only a forage could pay it, and the auto-tapper never forages", name)
		}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true}); err == nil {
			t.Errorf("%s was cast with auto-tap", name)
		}
	}
	if isTapped(g, forager) || !onBattlefield(g, food) || fgForages(g, me.ID) != 0 {
		t.Fatal("a refused auto-tap foraged or tapped the Forager")
	}

	one := handSpell(me, "One Green", "Instant", "{G}")
	if !castMove(g, me.ID, one) {
		t.Fatal("{G} is not offered with the Forager's plain row untapped")
	}
	if err := g.CastSpell(me.ID, one, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {G}: %v", err)
	}
	if !isTapped(g, forager) {
		t.Error("the Forager did not tap for the {G}")
	}
	if !onBattlefield(g, food) || fgForages(g, me.ID) != 0 {
		t.Error("the auto-tapper foraged")
	}
	for _, id := range gy {
		if !inGraveyard(me, id) {
			t.Error("the auto-tapper exiled a graveyard card")
		}
	}
}

// The plain row is an ordinary "{T}: Add {G}" and forages nothing.
func TestThornvaultForagerPlainRowAddsGreen(t *testing.T) {
	g, me, _ := eoTable(t)
	forager := tfForager(g, me)
	food := fgFoodToken(t, g, me.ID)
	if err := g.ActivateManaAbility(me.ID, forager, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {G}: %v", err)
	}
	tfPoolHas(t, me, "G")
	if !isTapped(g, forager) || !onBattlefield(g, food) || fgForages(g, me.ID) != 0 {
		t.Error("the plain row foraged or did not tap")
	}
	zero := 0
	if err := tfForage(g, me, forager, &zero, fgGraveyard(me, 3), nil); err == nil {
		t.Error("the tapped Forager foraged")
	}
}

// Summoning sickness stops both {T} mana rows (CR 302.6).
func TestThornvaultForagerForageRowRefusedWhileSick(t *testing.T) {
	g, me, _ := eoTable(t)
	forager := pushCatalogPermanent(g, me.ID, "Thornvault Forager", "Creature — Squirrel Ranger", thornvaultForagerOracle, true)
	gy := fgGraveyard(me, 3)
	zero := 0
	if err := tfForage(g, me, forager, &zero, gy, nil); !errors.Is(err, game.ErrSummoningSick) {
		t.Fatalf("sick forage: %v, want summoning sick", err)
	}
	if got := tfForageMoves(t, g, me.ID, forager); len(got) != 0 {
		t.Errorf("a sick Forager is offered forage moves %v", got)
	}
}

// "{3}{G}, {T}: Search your library for a Squirrel card": a changeling
// counts, a non-Squirrel is not offered, and the card goes to hand.
func TestThornvaultForagerTutorsASquirrel(t *testing.T) {
	g, me, _ := eoTable(t)
	forager := tfForager(g, me)
	bear := libraryCardFor(me, "Grizzly Bears", "Creature — Bear")
	nut := libraryCardFor(me, "Nut Hoarder", "Creature — Squirrel")
	fillPoolColored(me, "G", 4)
	if err := g.ActivateCatalogAbility(me.ID, forager, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the tutor: %v", err)
	}
	if !isTapped(g, forager) {
		t.Error("the tutor did not tap the Forager")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("the search offered a Bear")
	}
	answerSearchByID(t, g, me.ID, nut)
	if !inHand(me, nut) || inHand(me, bear) {
		t.Error("the Squirrel card did not go to hand alone")
	}
}

// The bot path: the enumerator offers each payable branch as its own
// move, with the cards or the Food named, and the move applied through
// actions.Dispatch — as a bot seat sends it — pays exactly that branch.
func TestThornvaultForagerEnumeratedMovesPayTheirBranch(t *testing.T) {
	for _, branch := range []int{0, 1} {
		g, me, _ := eoTable(t)
		forager := tfForager(g, me)
		food := fgFoodToken(t, g, me.ID)
		gy := fgGraveyard(me, 3)
		moves := tfForageMoves(t, g, me.ID, forager)
		if len(moves) != 2 {
			t.Fatalf("offered branches %v, want both", moves)
		}
		m := moves[branch]
		if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: me.ID, Caller: me.ID, Params: m.Params}); err != nil {
			t.Fatalf("branch %d move %q: %v", branch, m.Label, err)
		}
		if fgForages(g, me.ID) != 1 || !isTapped(g, forager) {
			t.Errorf("branch %d: did not forage", branch)
		}
		if got := onBattlefield(g, food); got != (branch == 0) {
			t.Errorf("branch %d: Food on battlefield = %v", branch, got)
		}
		for _, id := range gy {
			if got := inGraveyard(me, id); got != (branch == 1) {
				t.Errorf("branch %d: graveyard card in graveyard = %v", branch, got)
			}
		}
		// Two picks queued for the any-combination output; the move sent
		// no colours, as a bot's does not.
		for range 2 {
			pick := pendingOfKind(g, game.PendingChoiceMana)
			if pick == nil {
				t.Fatalf("branch %d: no mana pick", branch)
			}
			if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
				t.Fatal(err)
			}
		}
		tfPoolHas(t, me, "G", "G")
	}
}
