package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// roonc_cards_test.go — #4128 slice c, the Roon of the Hidden Realm
// deck request: Arctic Treeline, Yavimaya Hollow, Extraordinary Journey,
// Lonis Cryptozoologist, Moon-Blessed Cleric, Palinchron, Surprise
// Deployment, Titan of Industry, and the audit of Wood Elves.

const (
	rooncTreelineOracle   = "32fb0096-546c-4633-b6e3-ba3f2d9fa49c"
	rooncHollowOracle     = "53d6113d-acdb-4754-9641-f7991a96c7b9"
	rooncJourneyOracle    = "69288566-3015-45de-b391-6fe079865f0c"
	rooncLonisOracle      = "0ee06ed0-717a-4bfa-b634-d00e013c1f16"
	rooncClericOracle     = "89ed0598-a3ee-4d14-9a35-25e0dd80d9fb"
	rooncPalinchronOracle = "0023888e-7bec-43e0-8dee-d1a4eb94b372"
	rooncDeployOracle     = "61a1748c-e772-4338-9edd-720cd5d2fc12"
	rooncTitanOracle      = "5d8acafe-c13e-43ca-a89e-0cfd12d95662"
	rooncWoodElvesOracle  = "8973bd99-20f8-4867-90ef-50392147ee1b"
)

// rooncAllNine lists every card of the slice, so one test can pin that
// each is registered, Full, and carries no caveat.
var rooncAllNine = map[string]string{
	rooncTreelineOracle:   "Arctic Treeline",
	rooncHollowOracle:     "Yavimaya Hollow",
	rooncJourneyOracle:    "Extraordinary Journey",
	rooncLonisOracle:      "Lonis, Cryptozoologist",
	rooncClericOracle:     "Moon-Blessed Cleric",
	rooncPalinchronOracle: "Palinchron",
	rooncDeployOracle:     "Surprise Deployment",
	rooncTitanOracle:      "Titan of Industry",
	rooncWoodElvesOracle:  "Wood Elves",
}

// Ian's hard requirement for this slice: every card is Full, no caveats.
func TestRooncEverySliceCardIsFullWithNoCaveats(t *testing.T) {
	for oracle, name := range rooncAllNine {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s is registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s completeness = %v, want Full", name, spec.Completeness)
		}
		if len(spec.Caveats) != 0 {
			t.Errorf("%s carries caveats %v", name, spec.Caveats)
		}
	}
}

// --- helpers ---------------------------------------------------------

// rooncHand puts a card in a player's hand and returns its ID.
func rooncHand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

// rooncCastNow casts a card from the active seat's hand at the CURRENT
// step, without walking to a main phase (Surprise Deployment is cast in
// combat).
func rooncCastNow(g *game.Game, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	me := g.Seats[g.Turn.ActiveSeat]
	id := rooncHand(me, c)
	return id, g.CastSpell(me.ID, id, params)
}

// rooncCastMain casts a card from the active seat's hand in a main phase.
func rooncCastMain(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id, err := rooncCastNow(g, c, params)
	if err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return id
}

// rooncPlayLand plays a land from the active seat's hand.
func rooncPlayLand(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.LandDropsPerTurn++
	id := rooncHand(me, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

func rooncRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// rooncDestroy is a bare destroy effect from outside the catalog.
func rooncDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
}

// rooncKnownBy counts the cards of a library that `viewer` knows.
func rooncKnownBy(p *game.Player, viewer uuid.UUID) int {
	n := 0
	for _, c := range p.Library.Cards {
		if c.KnownBy[viewer] {
			n++
		}
	}
	return n
}

// rooncTapped reads a permanent's tapped flag.
func rooncTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// rooncUntappedLands counts a player's untapped battlefield lands.
func rooncUntappedLands(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.IsLand() && !c.Tapped {
			n++
		}
	}
	return n
}

// rooncTapLand puts a tapped land on the battlefield.
func rooncTapLand(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Basic Land — Island",
		Owner: owner, Controller: owner, Tapped: true,
	})
}

// --- Arctic Treeline --------------------------------------------------

// A snow Forest Plains tapland: enters tapped as a real replacement (no
// tap event), and once untapped taps for {G} or {W}.
func TestRooncArcticTreelineEntersTappedAndMakesGreenOrWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := rooncPlayLand(t, g, "Arctic Treeline", "Snow Land — Forest Plains", rooncTreelineOracle)
	if !rooncTapped(t, g, land) {
		t.Fatal("Arctic Treeline enters tapped")
	}
	if n := untapEventsFor(g, land); n != 0 {
		t.Errorf("it ENTERED tapped; %d untap events say it was tapped after the fact", n)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "G")
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G}: pool %v", got)
	}

	// The other colour is a real choice, not a fixed {G}.
	me.ManaPool.EmptyPool()
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "W")
	if got := poolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("tapped for {W}: pool %v", got)
	}
}

// --- Yavimaya Hollow --------------------------------------------------

func TestRooncYavimayaHollowTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hollow := pushCatalogPermanent(g, me.ID, "Yavimaya Hollow", "Legendary Land", rooncHollowOracle, false)
	b28TapForMana(t, g, me.ID, hollow, "")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("tapped for {C}: pool %v", got)
	}
}

// {G}, {T}: Regenerate target creature — the shield replaces the next
// destruction, the land is tapped by the cost, and it needs the {G}.
func TestRooncYavimayaHollowRegeneratesTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hollow := pushCatalogPermanent(g, me.ID, "Yavimaya Hollow", "Legendary Land", rooncHollowOracle, false)
	bear := seedCreature(g, "Bear", me.ID)

	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, hollow, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{rooncRef(bear)},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !rooncTapped(t, g, hollow) {
		t.Error("{T} is part of the cost")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %d: the {G} is part of the cost and was spent", me.ManaPool.Total())
	}
	if n := regenShieldsOn(t, g, bear); n != 0 {
		t.Errorf("shields = %d before the ability resolves: it uses the stack", n)
	}
	passPriorityAroundTable(t, g)
	if n := regenShieldsOn(t, g, bear); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	rooncDestroy(t, g, bear)
	if _, ok := battlefieldCard(g, bear); !ok {
		t.Fatal("CR 701.19a: the shield replaces the destruction")
	}
}

// The mana ability and the regeneration share {T}: using one spends the
// other.
func TestRooncYavimayaHollowTapIsSharedBetweenItsAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	hollow := pushCatalogPermanent(g, me.ID, "Yavimaya Hollow", "Legendary Land", rooncHollowOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	b28TapForMana(t, g, me.ID, hollow, "")
	me.ManaPool.EmptyPool()
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, hollow, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{rooncRef(bear)},
	}); err == nil {
		t.Error("a tapped Hollow cannot pay {T} for the regeneration")
	}
}

// --- Wood Elves (audit) ----------------------------------------------

// Any land with the Forest subtype qualifies (a Snow-Covered Forest, a
// Temple Garden), a Plains does not, and the land enters UNTAPPED.
func TestRooncWoodElvesFetchesAnyForestUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Temple Garden", "Land — Forest Plains"),
		searchTestLand("Snow-Covered Forest", "Basic Snow Land — Forest"),
		searchTestLand("Grizzly Bears", "Creature — Bear"),
	)
	rooncCastMain(t, g, game.Card{Name: "Wood Elves", TypeLine: "Creature — Elf Scout",
		OracleID: rooncWoodElvesOracle, Power: 1, Toughness: 1}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Wood Elves did not offer a search")
	}
	if searchOptionNamed(g, c, "Plains") != uuid.Nil || searchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("only Forest cards are legal finds")
	}
	if searchOptionNamed(g, c, "Temple Garden") == uuid.Nil || searchOptionNamed(g, c, "Snow-Covered Forest") == uuid.Nil {
		t.Error("a nonbasic and a snow Forest both qualify")
	}
	answerSearchNamed(t, g, me.ID, "Snow-Covered Forest")
	passPriorityAroundTable(t, g)

	var forest uuid.UUID
	for _, bc := range g.Battlefield.Cards {
		if bc.Name == "Snow-Covered Forest" {
			forest = bc.InstanceID
		}
	}
	if forest == uuid.Nil {
		t.Fatal("the Forest is not on the battlefield")
	}
	if rooncTapped(t, g, forest) {
		t.Error("Nothing prints 'tapped': the Forest enters untapped")
	}
	// Hidden information: only the fetched card was shown. The rest of
	// the library is unknown to the other seats.
	if n := rooncKnownBy(me, opp.ID); n != 0 {
		t.Errorf("the opponent knows %d library cards after a search that revealed one", n)
	}
	if names := revealedNames(revealsFor(g, opp.ID)); len(names) != 1 || names[0] != "Snow-Covered Forest" {
		t.Errorf("the table was shown %v, want only the fetched Forest", names)
	}
}

// A search with a stated quality may find nothing, and still shuffles.
func TestRooncWoodElvesMayFindNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"))
	rooncCastMain(t, g, game.Card{Name: "Wood Elves", TypeLine: "Creature — Elf Scout",
		OracleID: rooncWoodElvesOracle, Power: 1, Toughness: 1}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 1 {
		t.Errorf("nothing was taken: library = %d", me.Library.Size())
	}
}

// --- Moon-Blessed Cleric ---------------------------------------------

func rooncClericGame(t *testing.T) (*game.Game, *game.Player, *game.Player, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := seedSearchLibrary(me,
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Pacifism", "Enchantment — Aura"),
		searchTestLand("Grizzly Bears", "Creature — Bear"),
		searchTestLand("Rhystic Study", "Enchantment"),
		searchTestLand("Sol Ring", "Artifact"),
	)
	rooncCastMain(t, g, game.Card{Name: "Moon-Blessed Cleric", TypeLine: "Creature — Human Elf Cleric",
		OracleID: rooncClericOracle, Power: 3, Toughness: 2}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	return g, me, opp, ids
}

// Find an enchantment, reveal it, shuffle, put it on top.
func TestRooncMoonBlessedClericPutsTheEnchantmentOnTop(t *testing.T) {
	g, me, opp, _ := rooncClericGame(t)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a search")
	}
	if searchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil || searchOptionNamed(g, c, "Sol Ring") != uuid.Nil || searchOptionNamed(g, c, "Island") != uuid.Nil {
		t.Error("only enchantment cards are legal finds")
	}
	if searchOptionNamed(g, c, "Pacifism") == uuid.Nil || searchOptionNamed(g, c, "Rhystic Study") == uuid.Nil {
		t.Fatal("both enchantments (an Aura among them) are offered")
	}
	answerSearchNamed(t, g, me.ID, "Rhystic Study")
	passPriorityAroundTable(t, g)

	top := me.Library.Cards[len(me.Library.Cards)-1]
	if top.Name != "Rhystic Study" {
		t.Fatalf("the top card is %q, want Rhystic Study", top.Name)
	}
	if me.Library.Size() != 5 {
		t.Errorf("library = %d, the card never left it", me.Library.Size())
	}
	// Revealed: the table is told, and knows that one card and no other.
	if names := revealedNames(revealsFor(g, opp.ID)); len(names) != 1 || names[0] != "Rhystic Study" {
		t.Errorf("the table was shown %v, want only Rhystic Study", names)
	}
	if !top.KnownBy[opp.ID] {
		t.Error("the opponent saw the card go on top")
	}
	if n := rooncKnownBy(me, opp.ID); n != 1 {
		t.Errorf("the opponent knows %d library cards, want exactly the revealed one", n)
	}
}

// "You may search": taking nothing is legal, puts nothing on top and
// reveals nothing. (The engine answers a declined "may search" as a
// search that finds nothing, which still shuffles; see the card's comment.)
func TestRooncMoonBlessedClericMayTakeNothing(t *testing.T) {
	g, me, opp, ids := rooncClericGame(t)
	c := searchChoiceFor(g, me.ID)
	if c == nil || c.SearchMin != 0 {
		t.Fatalf("the search must allow taking nothing: %+v", c)
	}
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != len(ids) {
		t.Fatalf("library = %d, want %d", me.Library.Size(), len(ids))
	}
	if len(revealsFor(g, opp.ID)) != 0 {
		t.Error("nothing was revealed")
	}
	if n := rooncKnownBy(me, opp.ID); n != 0 {
		t.Errorf("the opponent knows %d library cards after a search that found nothing", n)
	}
}

// --- Palinchron -------------------------------------------------------

func TestRooncPalinchronUntapsUpToSevenLandsAnyPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := make([]uuid.UUID, 0, 8)
	for i := 0; i < 8; i++ {
		mine = append(mine, rooncTapLand(g, me.ID, "Island"))
	}
	theirs := rooncTapLand(g, opp.ID, "Their Island")
	rooncCastMain(t, g, game.Card{Name: "Palinchron", TypeLine: "Creature — Illusion",
		OracleID: rooncPalinchronOracle, Power: 4, Toughness: 5}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	p := latestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no untap prompt")
	}
	if p.ChooseMin != 0 || p.ChooseMax != 7 {
		t.Errorf("bounds %d..%d, want 0..7: \"up to seven\"", p.ChooseMin, p.ChooseMax)
	}
	if !hasID(p.ChooseCards, theirs) {
		t.Error("the clause says neither 'target' nor 'you control': an opponent's land is a legal pick")
	}
	if len(p.ChooseCards) != 9 {
		t.Errorf("%d tapped lands offered, want all 9", len(p.ChooseCards))
	}
	// Eight is too many.
	if err := g.ResolveChooseCards(p.ID, me.ID, mine); err == nil {
		t.Fatal("eight lands is more than 'up to seven'")
	}
	picks := append(append([]uuid.UUID(nil), mine[:6]...), theirs)
	answerChooseCards(t, g, me.ID, picks...)
	passPriorityAroundTable(t, g)
	if got := rooncUntappedLands(g, me.ID); got != 6 {
		t.Errorf("%d of my lands untapped, want 6", got)
	}
	if rooncTapped(t, g, theirs) {
		t.Error("the opponent's land untapped too")
	}
	if !rooncTapped(t, g, mine[7]) {
		t.Error("an unchosen land stays tapped")
	}
}

// {2}{U}{U}: return it to its owner's hand — and the replay untaps again.
func TestRooncPalinchronBouncesItselfForTwoUU(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pal := pushCatalogPermanent(g, me.ID, "Palinchron", "Creature — Illusion", rooncPalinchronOracle, false)

	fillPoolColored(me, "U", 2)
	fillPoolColored(me, "C", 2)
	if err := g.ActivateCatalogAbility(me.ID, pal, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %d: {2}{U}{U} is all four mana", me.ManaPool.Total())
	}
	if !g.Battlefield.Contains(pal) {
		t.Error("the ability uses the stack: Palinchron is still there until it resolves")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pal) || !me.Hand.Contains(pal) {
		t.Fatal("Palinchron did not return to its owner's hand")
	}
}

// --- Titan of Industry ------------------------------------------------

func rooncTitanModes(t *testing.T, g *game.Game, me *game.Player, modes ...int) {
	t.Helper()
	rooncCastMain(t, g, game.Card{Name: "Titan of Industry", TypeLine: "Creature — Elemental",
		OracleID: rooncTitanOracle, Power: 7, Toughness: 7}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("the enters trigger did not ask for modes: %+v", g.PendingChoices)
	}
	if c.ModeMin != 2 || c.ModeMax != 2 {
		t.Errorf("choose exactly two, got %d–%d", c.ModeMin, c.ModeMax)
	}
	if err := g.ResolveModePick(c.ID, me.ID, modes); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

// Destroy target artifact or enchantment, and target player gains 5:
// two targeted bullets on one trigger, each with its own slot.
func TestRooncTitanOfIndustryDestroysAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := pushArtifactToBattlefieldForTest(g, opp.ID, "Sol Ring", "Artifact", "")
	bear := seedCreature(g, "Their Bear", opp.ID)
	life := opp.Life
	rooncTitanModes(t, g, me, 0, 1)

	// The artifact bullet refuses a creature.
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, rooncRef(bear)); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, rooncRef(ring)); err != nil {
		t.Fatalf("pick the artifact: %v", err)
	}
	p = latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the life bullet asks for its own target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick the player: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ring) {
		t.Error("the artifact is destroyed")
	}
	if opp.Life != life+5 {
		t.Errorf("life %d → %d, want +5 (any target player)", life, opp.Life)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was not a target")
	}
}

// Rhino token and a shield counter — the counter goes on a creature
// chosen as the bullet resolves, and the token is eligible because the
// Rhino bullet is printed first.
func TestRooncTitanOfIndustryMakesARhinoAndShieldsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rooncTitanModes(t, g, me, 2, 3)
	passPriorityAroundTable(t, g)

	var rhino uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Rhino Warrior" {
			rhino = c.InstanceID
			if c.Power != 4 || c.Toughness != 4 || c.Controller != me.ID {
				t.Errorf("the Rhino is %d/%d under %s, want a 4/4 under me", c.Power, c.Toughness, c.Controller)
			}
		}
	}
	if rhino == uuid.Nil {
		t.Fatal("no Rhino Warrior token")
	}
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatalf("the counter bullet asks which creature: %+v", g.PendingChoices)
	}
	if !hasID(p.ChooseCards, rhino) || len(p.ChooseCards) != 2 {
		t.Errorf("candidates %v: want the Titan and the new Rhino", p.ChooseCards)
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{rhino}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, rhino, game.CounterShield); n != 1 {
		t.Fatalf("shield counters on the Rhino = %d, want 1", n)
	}
	// The shield does its job: the destroy removes the counter instead.
	rooncDestroy(t, g, rhino)
	if !g.Battlefield.Contains(rhino) {
		t.Fatal("the shield counter should have saved the Rhino")
	}
	if n := countersOn(g, rhino, game.CounterShield); n != 0 {
		t.Errorf("the shield is spent: %d left", n)
	}
	rooncDestroy(t, g, rhino)
	if g.Battlefield.Contains(rhino) {
		t.Error("with the shield gone the second destroy lands")
	}
}

// A different pair, to pin that the bullets are independent: the Rhino
// plus the life gain.
func TestRooncTitanOfIndustryRhinoAndLifeAreIndependentBullets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	rooncTitanModes(t, g, me, 1, 2)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err != nil {
		t.Fatalf("pick the player: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+5 || b16CountNamed(g, "Rhino Warrior") != 1 {
		t.Errorf("life %d → %d, rhinos %d: want +5 and one Rhino", life, me.Life, b16CountNamed(g, "Rhino Warrior"))
	}
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("the counter bullet was not chosen: nothing to pick")
	}
}

// --- Surprise Deployment ----------------------------------------------

func rooncDeploymentCard() game.Card {
	return game.Card{Name: "Surprise Deployment", TypeLine: "Instant", OracleID: rooncDeployOracle, ManaCost: "{3}{W}"}
}

func TestRooncSurpriseDeploymentIsCastableOnlyDuringCombat(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	if _, err := rooncCastNow(g, rooncDeploymentCard(), game.CastSpellParams{}); err == nil {
		t.Fatal("main phase: 'Cast this spell only during combat' must refuse the cast")
	}
	advanceTo(t, g, game.StepBeginCombat)
	if _, err := rooncCastNow(g, rooncDeploymentCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("beginning of combat is combat: %v", err)
	}
}

// Nonwhite creature cards only; the permanent enters for real; it goes
// back to hand at the next end step.
func TestRooncSurpriseDeploymentPutsANonwhiteCreatureThenReturnsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	white := rooncHand(me, game.Card{Name: "White Knight", TypeLine: "Creature — Human Knight", Colors: []string{"W"}, Power: 2, Toughness: 2})
	artifact := rooncHand(me, game.Card{Name: "Sol Ring", TypeLine: "Artifact"})
	green := rooncHand(me, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})
	colorless := rooncHand(me, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 3, Toughness: 3})
	gold := rooncHand(me, game.Card{Name: "Azorius Guard", TypeLine: "Creature — Human", Colors: []string{"W", "U"}, Power: 2, Toughness: 2})

	advanceTo(t, g, game.StepBeginCombat)
	if _, err := rooncCastNow(g, rooncDeploymentCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatalf("no put prompt: %+v", g.PendingChoices)
	}
	if p.ChooseMin != 0 {
		t.Error("'you may' — declining must be a legal answer")
	}
	if hasID(p.ChooseCards, white) || hasID(p.ChooseCards, gold) {
		t.Error("a white (or part-white) creature is not nonwhite")
	}
	if hasID(p.ChooseCards, artifact) {
		t.Error("a noncreature card is never offered")
	}
	if !hasID(p.ChooseCards, green) || !hasID(p.ChooseCards, colorless) {
		t.Errorf("green and colorless creatures are nonwhite: %v", p.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, green)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Fatal("the creature is on the battlefield")
	}
	if me.Hand.Contains(green) {
		t.Fatal("and no longer in hand")
	}

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) || !me.Hand.Contains(green) {
		t.Error("at the beginning of the next end step the creature returns to hand")
	}
	if !g.Battlefield.Contains(green) && me.Graveyard.Contains(green) {
		t.Error("it went to the graveyard, not the hand")
	}
}

// "Return it only if it's on the battlefield": a creature that died
// first stays dead.
func TestRooncSurpriseDeploymentDoesNotReturnADeadCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	green := rooncHand(me, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})
	advanceTo(t, g, game.StepBeginCombat)
	if _, err := rooncCastNow(g, rooncDeploymentCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, green)
	passPriorityAroundTable(t, g)
	rooncDestroy(t, g, green)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(green) || !me.Graveyard.Contains(green) {
		t.Error("a creature that already died is in the graveyard and stays there")
	}
}

// A flickered creature is a new object: the delayed trigger is about the
// old one, so the new one stays (CR 400.7).
func TestRooncSurpriseDeploymentDoesNotReturnANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	green := rooncHand(me, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})
	advanceTo(t, g, game.StepBeginCombat)
	if _, err := rooncCastNow(g, rooncDeploymentCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, green)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := (Flicker{Target: green}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("flicker: %v", err)
		}
	})
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := b16CountNamed(g, "Green Bear"); got != 1 {
		t.Errorf("the flickered Bear is a new object and stays: %d on the battlefield", got)
	}
	if me.Hand.Size() != 0 {
		for _, c := range me.Hand.Cards {
			if c.Name == "Green Bear" {
				t.Error("the new object must not be bounced")
			}
		}
	}
}

// Declining puts nothing in and schedules nothing.
func TestRooncSurpriseDeploymentMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	green := rooncHand(me, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})
	advanceTo(t, g, game.StepBeginCombat)
	if _, err := rooncCastNow(g, rooncDeploymentCard(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) || !me.Hand.Contains(green) {
		t.Error("declined: the Bear is still in hand")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(green) {
		t.Error("nothing was scheduled")
	}
}

// --- Extraordinary Journey --------------------------------------------

func rooncJourneyCard() game.Card {
	return game.Card{Name: "Extraordinary Journey", TypeLine: "Enchantment", OracleID: rooncJourneyOracle, ManaCost: "{X}{X}{U}{U}"}
}

// Cast with X=2: exile up to two target creatures, each playable by its
// OWNER for as long as it stays exiled — here the caster's own creature,
// so it can be cast right back, which draws a card.
func TestRooncExtraordinaryJourneyExilesXCreaturesAndDrawsWhenOneIsCastFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedCreature(g, "My Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)
	third := seedCreature(g, "Third Bear", opp.ID)
	journey := rooncCastMain(t, g, rooncJourneyCard(), game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)

	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no target prompt for the enters trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{rooncRef(mine), rooncRef(theirs), rooncRef(third)}); err == nil {
		t.Fatal("'up to X' with X=2 refuses three targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{rooncRef(mine), rooncRef(theirs)}); err != nil {
		t.Fatalf("pick two: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("both targets are exiled")
	}
	if !g.Battlefield.Contains(third) || !g.Battlefield.Contains(journey) {
		t.Error("the unchosen creature and the Journey stay")
	}

	// The OWNER may play it: my creature, cast from exile on my turn.
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast my exiled creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Fatal("the creature is back")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: a creature cast from exile draws a card", hand, me.Hand.Size())
	}
}

// Only once each turn: the second creature out of exile draws nothing.
func TestRooncExtraordinaryJourneyDrawsOnlyOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := seedCreature(g, "Bear A", me.ID)
	b := seedCreature(g, "Bear B", me.ID)
	rooncCastMain(t, g, rooncJourneyCard(), game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{rooncRef(a), rooncRef(b)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)

	hand := me.Hand.Size()
	for _, id := range []uuid.UUID{a, b} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
			t.Fatalf("cast from exile: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want exactly one card across two exile casts in a turn", hand, me.Hand.Size())
	}
}

// "Entered from exile" is not "cast": a flicker returns the creature from
// exile, and it can be ANY player's creature — the card is drawn by the
// Journey's controller.
func TestRooncExtraordinaryJourneyDrawsForAnyPlayersCreatureReturningFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := seedCreature(g, "Their Bear", opp.ID)
	rooncCastMain(t, g, rooncJourneyCard(), game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := (Flicker{Target: theirs}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("flicker: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: a creature entering from exile draws for the Journey's controller", hand, me.Hand.Size())
	}
}

// Not exile, not a card: a creature cast from hand and a token draw nothing.
func TestRooncExtraordinaryJourneyIgnoresHandCastsAndTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rooncCastMain(t, g, rooncJourneyCard(), game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()

	rooncCastMain(t, g, game.Card{Name: "Hand Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		_ = CreateToken{Controller: me.ID, Template: TokenCard("4/4 green Rhino Warrior"), N: 1}.Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d: neither a hand cast nor a token enters from exile", hand, me.Hand.Size())
	}
}

// X = 0 (or a Journey that was not cast) exiles nothing and asks nothing.
func TestRooncExtraordinaryJourneyWithXZeroTargetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedCreature(g, "Bear", me.ID)
	rooncCastMain(t, g, rooncJourneyCard(), game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("X=0: no target prompt")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("X=0 exiles nothing")
	}
}

// A token that is exiled is not a card: it is gone and cannot be played.
func TestRooncExtraordinaryJourneyExilingATokenLeavesNothingToPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var rhino uuid.UUID
	g.WithWriteLock(func() {
		_ = CreateToken{Controller: me.ID, Template: TokenCard("4/4 green Rhino Warrior"), N: 1}.Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Rhino Warrior" {
			rhino = c.InstanceID
		}
	}
	rooncCastMain(t, g, rooncJourneyCard(), game.CastSpellParams{XValue: 1})
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{rooncRef(rhino)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rhino) {
		t.Fatal("the token left")
	}
	if err := g.CastSpell(me.ID, rhino, game.CastSpellParams{}); err == nil {
		t.Error("a token that was exiled cannot be played")
	}
}

// --- Lonis, Cryptozoologist --------------------------------------------

func rooncClue(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clue", TypeLine: "Token Artifact — Clue",
		Owner: owner, Controller: owner,
	})
}

func rooncLonis(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Lonis, Cryptozoologist", "Legendary Creature — Snake Elf Scout", rooncLonisOracle, false)
}

func TestRooncLonisInvestigatesForAnotherNontokenCreatureOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rooncLonis(g, me.ID)

	rooncCastMain(t, g, game.Card{Name: "Hand Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := b16CountNamed(g, "Clue"); got != 1 {
		t.Fatalf("another nontoken creature entering: %d Clues, want 1", got)
	}

	// A token does not, and neither does an opponent's creature.
	g.WithWriteLock(func() {
		_ = CreateToken{Controller: me.ID, Template: TokenCard("4/4 green Rhino Warrior"), N: 1}.Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	passPriorityAroundTable(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Owner: opp.ID, Controller: opp.ID, Power: 2, Toughness: 2})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID})
	})
	passPriorityAroundTable(t, g)
	if got := b16CountNamed(g, "Clue"); got != 1 {
		t.Errorf("a token and an opponent's creature must not investigate: %d Clues", got)
	}
}

func rooncLonisLibrary(opp *game.Player) (top, land, big, small, e uuid.UUID) {
	opp.Library.Cards = nil
	// Pushed bottom-first: the library reads (top first) A, B, C, D, E.
	e = plTop(opp, "E Sorcery", "Sorcery", "{1}{R}")
	small = plTop(opp, "D Trinket", "Artifact", "{1}")
	big = plTop(opp, "C Dragon", "Creature — Dragon", "{2}{R}{R}")
	land = plTop(opp, "B Mountain", "Basic Land — Mountain", "")
	top = plTop(opp, "A Bear", "Creature — Bear", "{1}{G}")
	return
}

func rooncActivateLonis(t *testing.T, g *game.Game, me, opp *game.Player, lonis uuid.UUID, x int) {
	t.Helper()
	clues := make([]uuid.UUID, 0, x)
	for i := 0; i < x; i++ {
		clues = append(clues, rooncClue(g, me.ID))
	}
	if err := g.ActivateCatalogAbility(me.ID, lonis, 0, game.ActivateAbilityParams{
		XValue:       x,
		SacrificeIDs: clues,
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate at X=%d: %v", x, err)
	}
	for _, c := range clues {
		if g.Battlefield.Contains(c) {
			t.Errorf("a Clue was not sacrificed: the sacrifice is a cost")
		}
	}
	if !rooncTapped(t, g, lonis) {
		t.Error("{T} is part of the cost")
	}
}

// Reveal the top three of an opponent's library; put the Bear (mana
// value 2, nonland) onto the battlefield under MY control; the rest go
// to the bottom of THEIR library.
func TestRooncLonisPutsAQualifyingCardUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	lonis := rooncLonis(g, me.ID)
	bear, land, dragon, trinket, sorcery := rooncLonisLibrary(opp)
	rooncActivateLonis(t, g, me, opp, lonis, 3)
	passPriorityAroundTable(t, g)

	// Everyone is shown the three revealed cards, and only those.
	names := revealedNames(revealsFor(g, third.ID))
	if len(names) != 3 || !containsName(names, "A Bear") || !containsName(names, "B Mountain") || !containsName(names, "C Dragon") {
		t.Errorf("the table was shown %v, want the top three", names)
	}
	if containsName(names, "D Trinket") || containsName(names, "E Sorcery") {
		t.Error("a card below the revealed ones was shown")
	}

	p := latestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatalf("no put prompt for the controller: %+v", g.PendingChoices)
	}
	if latestChooseCardsFor(g, opp.ID) != nil {
		t.Error("the opponent is not asked anything: the activator chooses")
	}
	if p.ChooseMin != 0 || p.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1: 'you may put a ... card'", p.ChooseMin, p.ChooseMax)
	}
	if len(p.ChooseCards) != 1 || p.ChooseCards[0] != bear {
		t.Errorf("candidates %v: only the Bear is a nonland permanent card with mana value 3 or less", p.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	bc, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("the Bear is not on the battlefield")
	}
	if bc.Controller != me.ID || bc.Owner != opp.ID {
		t.Errorf("controller %s owner %s: want mine to control, theirs to own", bc.Controller, bc.Owner)
	}
	// The rest are at the bottom of THEIR library; the unrevealed
	// cards are still on top, in order, and still hidden from everyone.
	if opp.Library.Size() != 4 {
		t.Fatalf("their library has %d cards", opp.Library.Size())
	}
	bottom := plBottomIDs(opp, 2)
	if !bottom[land] || !bottom[dragon] {
		t.Errorf("the revealed rest belong on the bottom: %v", bottom)
	}
	n := len(opp.Library.Cards)
	// The top is the last element: D, then E under it.
	if opp.Library.Cards[n-1].InstanceID != trinket || opp.Library.Cards[n-2].InstanceID != sorcery {
		t.Errorf("the unrevealed cards must still be on top in order, got %s then %s",
			opp.Library.Cards[n-1].Name, opp.Library.Cards[n-2].Name)
	}
	for _, viewer := range []uuid.UUID{me.ID, third.ID} {
		if k := rooncKnownBy(opp, viewer); k != 0 {
			t.Errorf("seat %s knows %d of the opponent's library cards: the bottomed cards and the unrevealed ones must be hidden again", viewer, k)
		}
	}
}

// Declining ("you may") bottoms all three and enters nothing.
func TestRooncLonisMayDeclineAndEverythingGoesToTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	lonis := rooncLonis(g, me.ID)
	bear, land, dragon, _, _ := rooncLonisLibrary(opp)
	rooncActivateLonis(t, g, me, opp, lonis, 3)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("declined: nothing enters")
	}
	bottom := plBottomIDs(opp, 3)
	if !bottom[bear] || !bottom[land] || !bottom[dragon] {
		t.Errorf("all three revealed cards go to the bottom: %v", bottom)
	}
}

// Mana value X or less is a real bound: X=1 offers neither the Bear
// (MV 2) nor anything else, so nothing is asked and the revealed card
// is bottomed.
func TestRooncLonisManaValueBoundAndNoCandidateNoPrompt(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	lonis := rooncLonis(g, me.ID)
	bear, _, _, _, _ := rooncLonisLibrary(opp)
	rooncActivateLonis(t, g, me, opp, lonis, 1)
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("with no qualifying card there is no prompt")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("a mana value 2 card is not 'X or less' at X=1")
	}
	if !plBottomIDs(opp, 1)[bear] {
		t.Error("the one revealed card goes to the bottom")
	}
}

// A library shorter than X reveals what it has.
func TestRooncLonisShortLibraryRevealsWhatItHas(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	lonis := rooncLonis(g, me.ID)
	opp.Library.Cards = nil
	bear := plTop(opp, "A Bear", "Creature — Bear", "{1}{G}")
	rooncActivateLonis(t, g, me, opp, lonis, 3)
	passPriorityAroundTable(t, g)
	p := latestChooseCardsFor(g, me.ID)
	if p == nil || len(p.ChooseCards) != 1 || p.ChooseCards[0] != bear {
		t.Fatalf("the single card is offered: %+v", p)
	}
	answerChooseCards(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || opp.Library.Size() != 0 {
		t.Error("the Bear enters and the library is empty")
	}
}

// The permanent is its OWNER's card: it dies to the opponent's graveyard.
func TestRooncLonisStolenPermanentDiesToItsOwnersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	lonis := rooncLonis(g, me.ID)
	bear, _, _, _, _ := rooncLonisLibrary(opp)
	rooncActivateLonis(t, g, me, opp, lonis, 2)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	rooncDestroy(t, g, bear)
	if !opp.Graveyard.Contains(bear) || me.Graveyard.Contains(bear) {
		t.Error("the Bear goes to its owner's graveyard, not the controller's")
	}
}

// rooncVisibleOptions counts the card options a viewer is shown across
// every pending choice, through the real per-viewer filter.
func rooncVisibleOptions(g *game.Game, viewer uuid.UUID) int {
	v := protocol.FilterViewFor(protocol.ViewOfGame(g), viewer.String())
	n := 0
	for _, c := range v.PendingChoices {
		n += len(c.Options)
	}
	return n
}

// The put prompt belongs to the activator: the other seats' filtered
// views carry no candidates for it.
func TestRooncLonisPromptIsInvisibleToOtherSeats(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	lonis := rooncLonis(g, me.ID)
	rooncLonisLibrary(opp)
	rooncActivateLonis(t, g, me, opp, lonis, 3)
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) == nil {
		t.Fatal("setup: the activator should be holding the prompt")
	}
	if n := rooncVisibleOptions(g, me.ID); n != 1 {
		t.Errorf("the activator sees %d candidates, want 1", n)
	}
	for _, viewer := range []uuid.UUID{opp.ID, g.Seats[2].ID} {
		if n := rooncVisibleOptions(g, viewer); n != 0 {
			t.Errorf("seat %s is shown %d candidates of a prompt that is not theirs", viewer, n)
		}
	}
}
