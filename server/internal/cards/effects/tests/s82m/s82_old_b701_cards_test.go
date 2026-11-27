package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b701_cards_test.go — Starter 1999 sorceries, S82 OLD batch 701
// (#5620), played through the real catalog entries.

const (
	o701EyeSpy       = "0fef3f4f-18b6-452c-a634-c870709abcf5"
	o701Piracy       = "131acf69-fa5c-4ce7-9e39-2677af4f8907"
	o701Ransack      = "5b4a400b-435d-4140-984b-65f034c6a9fc"
	o701FalsePeace   = "7962db58-dbd9-4b94-8a21-a1625da4c384"
	o701NaturesCloak = "bb493c95-5f6d-405b-aac6-c56fe6b6f42f"
	o701WindSail     = "c7fb6ff0-4555-4bd1-b00d-ba2ed1cb4afc"
)

func TestS82OldB701CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o701EyeSpy: "Eye Spy", o701Piracy: "Piracy", o701Ransack: "Ransack",
		o701FalsePeace: "False Peace", o701NaturesCloak: "Nature's Cloak", o701WindSail: "Wind Sail",
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

func TestS82OldB701EyeSpyMayMillTheTopCard(t *testing.T) {
	for _, mill := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		top := effectstest.SsCard(opp.Library, opp.ID, "Their Top", "Sorcery", 0, 0)
		effectstest.CastCatalogSpell(t, g, "Eye Spy", "Sorcery", o701EyeSpy, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if prompt == nil {
			t.Fatal("the controller was not asked")
		}
		var picks []uuid.UUID
		if mill {
			picks = []uuid.UUID{top}
		}
		if err := g.ResolveChooseCards(prompt.ID, me.ID, picks); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := opp.Graveyard.Contains(top); got != mill {
			t.Errorf("mill=%v: the card is in the graveyard = %v", mill, got)
		}
		if opp.Library.Contains(top) == mill {
			t.Errorf("mill=%v: the card is still in the library = %v", mill, opp.Library.Contains(top))
		}
	}
}

func TestS82OldB701PiracyLetsYouTapLandsYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Piracy", "Sorcery", o701Piracy, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var mine, theirs bool
	g.WithWriteLock(func() {
		mine = g.MayTapForeignLandsLocked(me.ID)
		theirs = g.MayTapForeignLandsLocked(opp.ID)
	})
	if !mine {
		t.Error("the caster has no permission to tap lands they don't control")
	}
	if theirs {
		t.Error("the permission leaked to another player")
	}
}

func TestS82OldB701RansackPutsAnyNumberOnTheBottomAndTheRestOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	for i := 0; i < 6; i++ {
		ids = append(ids, effectstest.SsCard(opp.Library, opp.ID, "Card", "Sorcery", 0, 0))
	}
	size := len(effectstest.LibraryOrder(opp))
	// SsCard pushes on top: the last one is the top card.
	top5 := []uuid.UUID{ids[5], ids[4], ids[3], ids[2], ids[1]}
	effectstest.CastCatalogSpell(t, g, "Ransack", "Sorcery", o701Ransack, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the controller was not asked to arrange the five cards")
	}
	bottom := []uuid.UUID{top5[0], top5[1]}
	top := []uuid.UUID{top5[4], top5[3], top5[2]}
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, bottom, top); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	order := effectstest.LibraryOrder(opp)
	if len(order) != size {
		t.Fatalf("library size %d, want %d", len(order), size)
	}
	// The library's last element is its top card.
	for i, want := range top {
		if got := order[len(order)-1-i]; got != want {
			t.Errorf("position %d from the top: got %v, want %v", i, got, want)
		}
	}
	// The bottom lane is top-first, so its last card is the library's bottom.
	if order[0] != bottom[1] || order[1] != bottom[0] {
		t.Errorf("bottom of the library: got %v, want %v then %v", order[:2], bottom[1], bottom[0])
	}
}

// oldB701CombatInNextTurnOf walks the start of `seat`'s next turn through its end
// and reports whether any combat step began in it.
func oldB701CombatInNextTurnOf(t *testing.T, g *game.Game, seat int) bool {
	t.Helper()
	for i := 0; i < 400 && (g.Turn.ActiveSeat != seat || g.Turn.Step != game.StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to seat %d: %v", seat, err)
		}
	}
	combat := false
	for i := 0; i < 64 && g.Turn.ActiveSeat == seat; i++ {
		if game.PhaseOf(g.Turn.Step) == game.PhaseCombat {
			combat = true
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep in seat %d's turn: %v", seat, err)
		}
	}
	return combat
}

func TestS82OldB701FalsePeaceSkipsTheTargetsNextCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "False Peace", "Sorcery", o701FalsePeace, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if oldB701CombatInNextTurnOf(t, g, 1) {
		t.Error("the target's next turn still had a combat phase")
	}
	if !oldB701CombatInNextTurnOf(t, g, 2) {
		t.Error("the next player's turn lost its combat")
	}
}

func TestS82OldB701NaturesCloakGrantsForestwalkToGreenCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	red := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Bear", TypeLine: "Creature — Bear",
		Colors: []string{"R"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Nature's Cloak", "Sorcery", o701NaturesCloak, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, green, "forestwalk") {
		t.Error("a green creature you control did not gain forestwalk")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, red, "forestwalk") {
		t.Error("a red creature gained forestwalk")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "forestwalk") {
		t.Error("an opponent's green creature gained forestwalk")
	}
}

func TestS82OldB701WindSailGivesFlyingToOneOrTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Wind Sail", "Sorcery", o701WindSail, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
			t.Error("a target did not gain flying")
		}
	}
	if effectstest.HasKeywordOnBattlefield(t, g, c, "flying") {
		t.Error("a non-target gained flying")
	}
}
