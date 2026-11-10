package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// cryptex_test.go — Cryptex (Murders at Karlov Manor, #4469): collect
// evidence 3 as a MANA ability's cost (CR 605.1a, 701.59a), then the
// unlock-counter-gated sacrifice.

const cryptexOracle = "8186a992-9065-4a7f-bb00-1c20e01e8fb6"

func TestCryptexIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(cryptexOracle)
	if !ok {
		t.Fatal("Cryptex is not registered")
	}
	if spec.Name != "Cryptex" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
	if len(spec.ManaAbilities) != 1 || !spec.ManaAbilities[0].Cost.ExileCards.IsCollectEvidence() ||
		spec.ManaAbilities[0].Cost.ExileCards.ManaValueAtLeast != 3 {
		t.Error("want one {T}, Collect evidence 3 mana row")
	}
}

func cryptexOnBoard(g *game.Game, me *game.Player) uuid.UUID {
	return pushCatalogPermanent(g, me.ID, "Cryptex", "Artifact", cryptexOracle, false)
}

func cryptexTap(g *game.Game, me *game.Player, cx uuid.UUID, color string, exile ...uuid.UUID) error {
	return g.ActivateManaAbility(me.ID, cx, 0, game.ManaAbilityParams{ExileIDs: exile, Colors: []string{color}})
}

// Collect evidence 3 pays: the named cards (total mana value 3) are
// exiled, the mana lands without the stack, an unlock counter goes on,
// and one collection is announced.
func TestCryptexCollectsEvidenceForManaAndAnUnlockCounter(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	cx := cryptexOnBoard(g, me)
	gy := mkmCEEvidence(me, "{1}", "{2}", "{5}")
	if err := cryptexTap(g, me, cx, "U", gy[0], gy[1]); err != nil {
		t.Fatalf("collect evidence 3 with exactly 3: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool = %v, want [U]", got)
	}
	if g.Stack.Size() != 0 {
		t.Error("a mana ability used the stack (CR 605.3b)")
	}
	if !isTapped(g, cx) {
		t.Error("Cryptex did not tap")
	}
	if !inExile(g, gy[0]) || !inExile(g, gy[1]) || !me.Graveyard.Contains(gy[2]) {
		t.Error("only the two named cards are exiled")
	}
	if n := countersOn(g, cx, "unlock"); n != 1 {
		t.Errorf("%d unlock counters, want 1", n)
	}
	wantOneCollection(t, g, me.ID, 3, gy[0], gy[1])
}

// Strict refusals: a pick below the floor, no pick, or a card that is not
// in the activator's graveyard. Each is refused with nothing paid.
func TestCryptexRefusesShortOrForeignEvidence(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	cx := cryptexOnBoard(g, me)
	gy := mkmCEEvidence(me, "{1}", "{1}")
	theirs := mkmCEEvidence(opp, "{3}")
	hand := handSpell(me, "Hand Card", "Sorcery", "{4}")
	cases := map[string]error{
		"a total of 2":         cryptexTap(g, me, cx, "R", gy...),
		"nothing":              cryptexTap(g, me, cx, "R"),
		"an opponent's card":   cryptexTap(g, me, cx, "R", theirs[0]),
		"a card in hand":       cryptexTap(g, me, cx, "R", hand),
		"one card named twice": cryptexTap(g, me, cx, "R", gy[0], gy[0], gy[1]),
	}
	for name, err := range cases {
		if err == nil {
			t.Errorf("%s: activation accepted", name)
		}
	}
	if isTapped(g, cx) || len(poolColors(me)) != 0 || me.Graveyard.Size() != 2 || countersOn(g, cx, "unlock") != 0 ||
		len(evidenceEvents(g)) != 0 {
		t.Fatal("a refused activation paid something")
	}
}

// The view ships the floor and greys the row while the graveyard is
// short; the enumerator offers the move only once the graveyard reaches
// three, naming enough cards, and the move applied as a bot sends it
// pays.
func TestCryptexViewAndEnumeratorReadTheGate(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	cx := cryptexOnBoard(g, me)
	gy := mkmCEEvidence(me, "{2}")
	row := cryptexManaRow(t, g, me.ID, cx)
	if !row.ExileCostAnyNumber || row.ExileCostManaValueMin != 3 || row.ExileCostZone != "graveyard" || row.ExileCostN != 0 {
		t.Errorf("row %+v, want an any-number graveyard exile with a floor of 3", row)
	}
	if row.CantActivate == "" {
		t.Error("a graveyard totalling 2 is not greyed")
	}
	if m := cryptexMove(t, g, me.ID, cx); m != nil {
		t.Errorf("offered %q over a graveyard totalling 2", m.Label)
	}
	gy = append(gy, gyCard(me, "More Evidence", "Sorcery", "{1}", ""))
	row = cryptexManaRow(t, g, me.ID, cx)
	if row.CantActivate != "" || len(row.ExileCostOptions) != 2 {
		t.Errorf("row %+v, want payable with both cards offered", row)
	}
	m := cryptexMove(t, g, me.ID, cx)
	if m == nil {
		t.Fatal("no Cryptex move over a graveyard totalling 3")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: me.ID, Caller: me.ID, Params: m.Params}); err != nil {
		t.Fatalf("move %q: %v", m.Label, err)
	}
	if !inExile(g, gy[0]) || !inExile(g, gy[1]) || countersOn(g, cx, "unlock") != 1 {
		t.Error("the enumerated move did not collect evidence")
	}
}

// The auto-tapper never collects evidence: with Cryptex the only source
// and a full graveyard, a {1} spell is neither offered nor castable with
// auto-tap.
func TestCryptexAutoTapperNeverCollectsEvidence(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	cx := cryptexOnBoard(g, me)
	gy := mkmCEEvidence(me, "{3}", "{3}")
	spell := handSpell(me, "One Generic", "Instant", "{1}")
	if castMove(g, me.ID, spell) {
		t.Error("the {1} spell is offered: only Cryptex's collect evidence could pay it")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err == nil {
		t.Error("the {1} spell was cast with auto-tap")
	}
	if isTapped(g, cx) || !me.Graveyard.Contains(gy[0]) || !me.Graveyard.Contains(gy[1]) {
		t.Error("the auto-tapper spent Cryptex's graveyard")
	}
}

// The sacrifice: refused below five unlock counters, then surveil 3 and
// draw three.
func TestCryptexSacrificesAtFiveUnlockCounters(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	cx := cryptexOnBoard(g, me)
	for i := 0; i < 6; i++ {
		topLibraryCard(me, "Deep", "Sorcery")
	}
	third := topLibraryCard(me, "Third", "Sorcery")
	second := topLibraryCard(me, "Second", "Sorcery")
	top := topLibraryCard(me, "Top", "Sorcery")
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(cx, "unlock", 4); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, cx, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("sacrificed with four unlock counters")
	}
	if !onBattlefield(g, cx) {
		t.Fatal("a refused activation sacrificed Cryptex")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(cx, "unlock", 1); err != nil {
			t.Fatal(err)
		}
	})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, cx, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sacrifice with five unlock counters: %v", err)
	}
	if onBattlefield(g, cx) {
		t.Error("Cryptex was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 3 {
		t.Fatalf("surveil prompt %+v, want one over three cards", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, []uuid.UUID{second, third}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card is not in the graveyard")
	}
	if me.Hand.Size() != hand+3 || !me.Hand.Contains(second) || !me.Hand.Contains(third) {
		t.Errorf("hand %d, want %d holding the two kept cards", me.Hand.Size(), hand+3)
	}
}

func cryptexManaRow(t *testing.T, g *game.Game, viewer, source uuid.UUID) protocol.ManaAbilityView {
	t.Helper()
	activatedRowOfMarkKnown(g)
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, cv := range v.Battlefield.Cards {
		if cv.InstanceID == source.String() && len(cv.ManaAbilities) > 0 {
			return cv.ManaAbilities[0]
		}
	}
	t.Fatalf("no mana row on %s", source)
	return protocol.ManaAbilityView{}
}

// activatedRowOfMarkKnown marks every permanent known to every seat, as
// activatedRowOf does before it reads the view.
func activatedRowOfMarkKnown(g *game.Game) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.KnownBy == nil {
				c.KnownBy = map[uuid.UUID]bool{}
			}
			for _, p := range g.Seats {
				c.KnownBy[p.ID] = true
			}
		}
	})
}

func cryptexMove(t *testing.T, g *game.Game, seat, cx uuid.UUID) *legal.Move {
	t.Helper()
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeActivateManaAbility || m.Source != cx {
			continue
		}
		var p struct {
			ExileIDs []string `json:"exile_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if len(p.ExileIDs) == 0 {
			t.Errorf("move %q names no evidence", m.Label)
		}
		return &m
	}
	return nil
}
