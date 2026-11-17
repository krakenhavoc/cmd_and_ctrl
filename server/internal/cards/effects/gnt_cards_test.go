package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Game Night (#5523): Avatar of Growth, Inspired Sphinx and Militant
// Angel. Every test runs at the four-seat table newCatalogGame deals, so
// "the number of opponents you have" is three until someone leaves.

const (
	gntAvatarOracle = "07db0374-3297-49c3-886d-a6bb42f7bb18"
	gntSphinxOracle = "61cceb14-9568-4a5e-9b56-07afa6407df2"
	gntAngelOracle  = "ad645324-204a-49d2-b128-f98eee7c9bb9"
)

func TestGntCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gntAvatarOracle: "Avatar of Growth",
		gntSphinxOracle: "Inspired Sphinx",
		gntAngelOracle:  "Militant Angel",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// gntConcede has a player leave the game through the real concession
// path (CR 800.4a), not a flag flip.
func gntConcede(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	if err := g.Concede(p.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
}

// gntAvatar is Avatar of Growth as printed, for a strict cast.
func gntAvatar() game.Card {
	return game.Card{
		Name: "Avatar of Growth", TypeLine: "Creature — Elemental Avatar", ManaCost: "{4}{G}{G}",
		OracleID: gntAvatarOracle, Colors: []string{"G"}, Power: 4, Toughness: 4,
	}
}

// gntCastWithForests puts c in the active seat's hand, gives the seat
// `forests` more untapped Forests, and casts c strictly with auto-tap, so
// the cost modifiers are applied and the mana is really paid.
func gntCastWithForests(t *testing.T, g *game.Game, c game.Card, forests int) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	for i := 0; i < forests; i++ {
		pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(c)
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, AutoTap: true})
}

// --- Avatar of Growth ----------------------------------------------------------

func TestGntAvatarOfGrowthCostsOneLessPerOpponentAndEveryoneFetchesTwoBasics(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	mine := []uuid.UUID{
		b07LibraryLand(me, "Forest", "Basic Land — Forest"),
		b07LibraryLand(me, "Forest", "Basic Land — Forest"),
		b07LibraryLand(me, "Forest", "Basic Land — Forest"),
	}
	theirPlains := b07LibraryLand(opps[0], "Plains", "Basic Land — Plains")
	b07LibraryLand(opps[1], "Island", "Basic Land — Island")
	b07LibraryLand(opps[1], "Island", "Basic Land — Island")
	lastIsland := b07LibraryLand(opps[2], "Island", "Basic Land — Island")
	b07LibraryLand(opps[2], "Swamp", "Basic Land — Swamp")

	// Three opponents: {4}{G}{G} costs {1}{G}{G}.
	if _, err := gntCastWithForests(t, g, gntAvatar(), 2); err == nil {
		t.Fatal("Avatar of Growth cast for two mana with three opponents")
	}
	id, err := gntCastWithForests(t, g, gntAvatar(), 1)
	if err != nil {
		t.Fatalf("Avatar of Growth for three mana with three opponents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Avatar of Growth is not on the battlefield")
	}
	assertKeywords(t, g, id, "trample")

	if n := b07OpenSearchPrompts(g); n != 4 {
		t.Fatalf("%d search prompts, want one for each of the four players", n)
	}
	landsBefore := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		landsBefore[p.ID] = b07LandsOnBattlefield(g, p.ID)
	}
	answerSearchByID(t, g, me.ID, mine[0], mine[1])
	answerSearchByID(t, g, opps[0].ID, theirPlains)
	answerSearchFailToFind(t, g, opps[1].ID)
	answerSearchByID(t, g, opps[2].ID, lastIsland)
	passPriorityAroundTable(t, g)

	for p, want := range map[*game.Player]int{me: 2, opps[0]: 1, opps[1]: 0, opps[2]: 1} {
		if got := b07LandsOnBattlefield(g, p.ID) - landsBefore[p.ID]; got != want {
			t.Errorf("seat %d put %d lands onto the battlefield, want %d", p.Seat, got, want)
		}
	}
	for _, land := range []uuid.UUID{mine[0], mine[1], theirPlains, lastIsland} {
		if c, ok := battlefieldCard(g, land); !ok || c.Tapped {
			t.Errorf("fetched land %s: on battlefield %v, tapped %v; want it untapped on the battlefield", land, ok, c.Tapped)
		}
	}
	if !me.Library.Contains(mine[2]) {
		t.Error("the third Forest left the library: the search is for up to two")
	}
}

func TestGntAvatarOfGrowthCountsOnlyOpponentsStillInTheGame(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	advanceToMain(t, g)
	gntConcede(t, g, opps[1])

	// Two opponents: {4}{G}{G} costs {2}{G}{G}.
	if _, err := gntCastWithForests(t, g, gntAvatar(), 3); err == nil {
		t.Fatal("Avatar of Growth cast for three mana with two opponents")
	}
	if _, err := gntCastWithForests(t, g, gntAvatar(), 1); err != nil {
		t.Fatalf("Avatar of Growth for four mana with two opponents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b07OpenSearchPrompts(g); n != 3 {
		t.Fatalf("%d search prompts, want three: the player who left does not search", n)
	}
	for _, p := range []*game.Player{me, opps[0], opps[2]} {
		answerSearchFailToFind(t, g, p.ID)
	}
	if n := b07OpenSearchPrompts(g); n != 0 {
		t.Errorf("%d search prompts left open", n)
	}
}

// --- Inspired Sphinx -----------------------------------------------------------

func TestGntInspiredSphinxDrawsOnePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB37Seats(g)
	advanceToMain(t, g)
	before := me.Hand.Size()
	id := castHoldCreature(t, g, "Inspired Sphinx", "Creature — Sphinx", gntSphinxOracle, 5, 5)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Inspired Sphinx is not on the battlefield")
	}
	assertKeywords(t, g, id, "flying")
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d cards with three opponents, want 3", got)
	}
}

func TestGntInspiredSphinxDoesNotCountAnOpponentWhoLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	advanceToMain(t, g)
	gntConcede(t, g, opps[0])
	before := me.Hand.Size()
	castHoldCreature(t, g, "Inspired Sphinx", "Creature — Sphinx", gntSphinxOracle, 5, 5)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d cards with two opponents left, want 2", got)
	}
}

func TestGntInspiredSphinxMakesAColorlessFlyingThopter(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB37Seats(g)
	advanceToMain(t, g)
	sphinx := pushCatalogPermanent(g, me.ID, "Inspired Sphinx", "Creature — Sphinx", gntSphinxOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, sphinx, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	thopters := trekB37Tokens(g, me.ID, "Thopter")
	if len(thopters) != 1 {
		t.Fatalf("%d Thopters, want 1", len(thopters))
	}
	th := thopters[0]
	if th.Power != 1 || th.Toughness != 1 || len(th.Colors) != 0 || !th.IsArtifact() || !th.IsCreature() || !th.HasSubtype("Thopter") {
		t.Errorf("token = %s %d/%d %v, want a 1/1 colorless Thopter artifact creature", th.TypeLine, th.Power, th.Toughness, th.Colors)
	}
	if !hasEffectiveKeyword(t, g, th.InstanceID, "flying") {
		t.Error("the Thopter has no flying")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("%d mana left in the pool, want the {3}{U} spent", len(me.ManaPool))
	}
}

// --- Militant Angel ------------------------------------------------------------

// gntAngelAfterCombat declares the attacks, finishes combat, and casts
// Militant Angel in the second main phase. Returns the Angel.
func gntAngelAfterCombat(t *testing.T, g *game.Game, attacks map[uuid.UUID]uuid.UUID, leave ...*game.Player) uuid.UUID {
	t.Helper()
	if len(attacks) > 0 {
		trekB37Declare(t, g, attacks)
		passPriorityAroundTable(t, g)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	for _, p := range leave {
		gntConcede(t, g, p)
	}
	id := castHoldCreature(t, g, "Militant Angel", "Creature — Angel", gntAngelOracle, 3, 4)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Militant Angel is not on the battlefield")
	}
	return id
}

func TestGntMilitantAngelMakesAKnightPerOpponentAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	c := pushVanillaCreature(g, me.ID, "Bear C", 2, 2)
	// Two creatures at the first opponent and one at the second: two
	// opponents attacked, each counted once.
	angel := gntAngelAfterCombat(t, g, map[uuid.UUID]uuid.UUID{a: opps[0].ID, b: opps[0].ID, c: opps[1].ID})
	assertKeywords(t, g, angel, "flying", "lifelink")
	knights := trekB37Tokens(g, me.ID, "Knight")
	if len(knights) != 2 {
		t.Fatalf("%d Knights after attacking two opponents, want 2", len(knights))
	}
	for _, k := range knights {
		if k.Power != 2 || k.Toughness != 2 || !slices.Equal(k.Colors, []string{"W"}) || !k.HasSubtype("Knight") {
			t.Errorf("token = %s %d/%d %v, want a 2/2 white Knight", k.TypeLine, k.Power, k.Toughness, k.Colors)
		}
		if !hasEffectiveKeyword(t, g, k.InstanceID, "vigilance") {
			t.Error("a Knight has no vigilance")
		}
	}
}

func TestGntMilitantAngelMakesNothingWithoutAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := trekB37Seats(g)
	gntAngelAfterCombat(t, g, nil)
	if n := len(trekB37Tokens(g, me.ID, "Knight")); n != 0 {
		t.Errorf("%d Knights without attacking anyone, want 0", n)
	}
}

func TestGntMilitantAngelSkipsAnAttackedOpponentWhoLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := trekB37Seats(g)
	attacks := map[uuid.UUID]uuid.UUID{}
	for i, opp := range opps {
		attacks[pushVanillaCreature(g, me.ID, "Bear", 2, 2+i)] = opp.ID
	}
	gntAngelAfterCombat(t, g, attacks, opps[2])
	if n := len(trekB37Tokens(g, me.ID, "Knight")); n != 2 {
		t.Errorf("%d Knights after attacking three opponents, one of whom left, want 2", n)
	}
}
