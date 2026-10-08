package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #2526 — Jarad, Golgari Lich Lord's third ability: "Sacrifice a Swamp
// and a Forest: Return this card from your graveyard to your hand."
// The cost is two permanents of DIFFERENT kinds, which a sacrifice
// clause could not say before TargetSpec.EachOf.

// jaradTable puts Jarad in the active seat's graveyard at a main phase
// and returns the seat.
func jaradTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := pushCatalogGraveyardCard(me, "Jarad, Golgari Lich Lord", "Legendary Creature — Zombie Elf", b17JaradOracle, 2, 2)
	return g, me, id
}

func jaradLand(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return b12Permanent(g, owner, name, typeLine)
}

func TestJaradReturnsForASwampAndAForest(t *testing.T) {
	g, me, jarad := jaradTable(t)
	swamp := jaradLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	forest := jaradLand(g, me.ID, "Forest", "Basic Land — Forest")
	keep := jaradLand(g, me.ID, "Island", "Basic Land — Island")

	if err := g.ActivateCatalogAbility(me.ID, jarad, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{swamp, forest}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	ringSettle(t, g)
	if g.Battlefield.Contains(swamp) || g.Battlefield.Contains(forest) {
		t.Error("the Swamp and the Forest were not sacrificed")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("an Island was sacrificed")
	}
	if !me.Hand.Contains(jarad) || me.Graveyard.Contains(jarad) {
		t.Fatal("Jarad did not return to hand")
	}
}

// Two Swamps are not a Swamp and a Forest. The refusal leaves both on
// the battlefield: a payment is all or nothing.
func TestJaradRefusesTwoSwamps(t *testing.T) {
	g, me, jarad := jaradTable(t)
	a := jaradLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	b := jaradLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	jaradLand(g, me.ID, "Forest", "Basic Land — Forest")

	err := g.ActivateCatalogAbility(me.ID, jarad, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two Swamps: err = %v, want ErrIllegalTarget", err)
	}
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) || !me.Graveyard.Contains(jarad) {
		t.Error("a refused payment moved something")
	}
}

// A Swamp Forest can fill either part but not both: it plus a Forest
// pays, it alone (or with another dual, which also fits) pays only with
// two permanents, and a single permanent never does.
func TestJaradDualLandFillsOnePartOnly(t *testing.T) {
	g, me, jarad := jaradTable(t)
	tomb := jaradLand(g, me.ID, "Overgrown Tomb", "Land — Swamp Forest")
	forest := jaradLand(g, me.ID, "Forest", "Basic Land — Forest")
	island := jaradLand(g, me.ID, "Island", "Basic Land — Island")

	// One dual + an Island: the Island fills neither part.
	if err := g.ActivateCatalogAbility(me.ID, jarad, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{tomb, island}}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("dual + Island: err = %v, want ErrIllegalTarget", err)
	}
	// One permanent is the wrong count.
	if err := g.ActivateCatalogAbility(me.ID, jarad, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{tomb}}); err == nil {
		t.Fatal("a lone dual land paid a two-permanent cost")
	}
	// The dual as the Swamp, the Forest as the Forest.
	if err := g.ActivateCatalogAbility(me.ID, jarad, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest, tomb}}); err != nil {
		t.Fatalf("Forest + dual (either order): %v", err)
	}
	ringSettle(t, g)
	if !me.Hand.Contains(jarad) {
		t.Fatal("Jarad did not return")
	}
}

// The enumerator offers one payment that fills both parts, even when
// the first two permanents of its payment order are both Swamps, and it
// leaves the dual land on the battlefield when basics will do.
func TestJaradEnumeratorFindsASetThatFillsBothParts(t *testing.T) {
	g, me, jarad := jaradTable(t)
	s1 := jaradLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	s2 := jaradLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	tomb := jaradLand(g, me.ID, "Overgrown Tomb", "Land — Swamp Forest")
	forest := jaradLand(g, me.ID, "Forest", "Basic Land — Forest")

	var picks []string
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != "activate_ability" || m.Source != jarad {
			continue
		}
		var p struct {
			AbilityIndex int      `json:"ability_index"`
			SacrificeIDs []string `json:"sacrifice_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params: %v", err)
		}
		if p.AbilityIndex == 1 {
			picks = append(picks, p.SacrificeIDs...)
		}
	}
	if len(picks) != 2 {
		t.Fatalf("enumerator offered picks %v, want exactly one two-permanent payment", picks)
	}
	ids := []uuid.UUID{}
	for _, s := range picks {
		ids = append(ids, uuid.MustParse(s))
	}
	spec := game.ActivatedAbilitiesForCard(game.Card{OracleID: b17JaradOracle})[1].Cost.SacrificeOther
	g.WithWriteLock(func() {
		if !g.SacrificeSetSatisfiedForEffect(spec, ids) {
			t.Errorf("the offered payment %v does not fill both parts", picks)
		}
	})
	for _, id := range ids {
		if id == tomb {
			t.Error("the bot spends the dual land when a Swamp and a Forest are on the table")
		}
	}
	_, _, _ = s1, s2, forest

	// And the offered payment really pays.
	if err := g.ActivateCatalogAbility(me.ID, jarad, 1, game.ActivateAbilityParams{SacrificeIDs: ids}); err != nil {
		t.Fatalf("activating the enumerated payment: %v", err)
	}
}

// With only Swamps there is no payment, so no move.
func TestJaradNotOfferedWithoutBothKinds(t *testing.T) {
	g, me, jarad := jaradTable(t)
	jaradLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	jaradLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != "activate_ability" || m.Source != jarad {
			continue
		}
		var p struct {
			AbilityIndex int `json:"ability_index"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if p.AbilityIndex == 1 {
			t.Fatal("the return was offered with no Forest to sacrifice")
		}
	}
}

// The view ships the union as the options and each part's candidates as
// each_of, which is what holds the client's confirm button.
func TestJaradViewShipsEachPart(t *testing.T) {
	g, me, jarad := jaradTable(t)
	swamp := jaradLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	forest := jaradLand(g, me.ID, "Forest", "Basic Land — Forest")
	tomb := jaradLand(g, me.ID, "Overgrown Tomb", "Land — Swamp Forest")
	jaradLand(g, me.ID, "Island", "Basic Land — Island")

	view := protocol.ViewOfGameFor(g, me.ID.String())
	var opts *protocol.LegalTargetsView
	for _, p := range view.Seats {
		if p.ID != me.ID.String() {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.InstanceID != jarad.String() {
				continue
			}
			for _, a := range c.ZoneAbilities {
				if a.SacrificeOptions != nil {
					opts = a.SacrificeOptions
				}
			}
		}
	}
	if opts == nil {
		t.Fatal("Jarad's graveyard row carries no sacrifice options")
	}
	if opts.Min != 2 || opts.Max != 2 {
		t.Errorf("bounds %d..%d, want 2..2", opts.Min, opts.Max)
	}
	if len(opts.Cards) != 3 {
		t.Errorf("options %v, want the Swamp, the Forest and the dual (not the Island)", opts.Cards)
	}
	if len(opts.EachOf) != 2 || opts.EachOf[0].Label != "a Swamp" || opts.EachOf[1].Label != "a Forest" {
		t.Fatalf("each_of = %+v, want a Swamp then a Forest", opts.EachOf)
	}
	has := func(g protocol.SacrificeGroupView, id uuid.UUID) bool {
		for _, c := range g.Cards {
			if c == id.String() {
				return true
			}
		}
		return false
	}
	if !has(opts.EachOf[0], swamp) || !has(opts.EachOf[0], tomb) || has(opts.EachOf[0], forest) {
		t.Errorf("the Swamp part lists %v", opts.EachOf[0].Cards)
	}
	if !has(opts.EachOf[1], forest) || !has(opts.EachOf[1], tomb) || has(opts.EachOf[1], swamp) {
		t.Errorf("the Forest part lists %v", opts.EachOf[1].Cards)
	}
}

// Register refuses a set rule whose count is not its entry count, and a
// one-entry rule: a cost the validator and the picker would read
// differently must fail at boot, not at the table.
func TestSacrificeSetRuleRegistrationGuard(t *testing.T) {
	bad := func(name string, mutate func(*game.TargetSpec)) {
		t.Helper()
		cost := SacrificeEach("a Swamp and a Forest",
			SacrificeSubtype("a Swamp", "Swamp"),
			SacrificeSubtype("a Forest", "Forest"))
		mutate(cost.SacrificeOther)
		defer func() {
			if recover() == nil {
				t.Errorf("%s: checkSacrificeClause accepted it", name)
			}
		}()
		checkSacrificeClause("Test", "ability 0", cost.SacrificeOther, true, true, false)
	}
	bad("count below the entries", func(s *game.TargetSpec) { s.Min, s.Max = 1, 1 })
	bad("open count", func(s *game.TargetSpec) { s.Min, s.Max = 2, 0 })
	bad("one entry", func(s *game.TargetSpec) { s.EachOf = s.EachOf[:1] })
	bad("an entry that names nothing", func(s *game.TargetSpec) { s.EachOf[1] = game.SacrificeKind{Label: "a Forest"} })
}
