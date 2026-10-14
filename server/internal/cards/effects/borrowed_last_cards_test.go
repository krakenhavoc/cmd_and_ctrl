package effects

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// borrowed_last_cards_test.go — ADR 0155 PR 4 (#3202), the last cards on
// the borrow seam: the resolved borrows (Quicksilver Elemental, Grell
// Philosopher, Havengul Lich), the library-top lenders (Conspicuous Snoop,
// Skill Borrower) and Scheming Fence. Each test asserts what a player
// gets: which permanent has which ability, what pays for it and what
// using it does.

const (
	slQuicksilverOracle = "c88e3aff-c539-4e8b-8ee1-a43b3d0962b4"
	slGrellOracle       = "d5dc00cb-1510-4ce9-8dd9-0722f65d34b4"
	slLichOracle        = "1fb02bad-0d1f-437f-a269-f51cc19b280f"
	slSnoopOracle       = "f5d1bd3c-0e65-4999-a810-881b2389b40e"
	slSkillOracle       = "9e862e4a-254d-4d39-b2cf-4beed6db761f"
	slFenceOracle       = "cef87a98-61a0-4a88-950f-867144f685e3"

	slShadeOracle       = "9af1b6c4-295f-41d2-b3a0-0863798330d4" // Evernight Shade: {B}: +1/+1 until end of turn
	slPuzzleknotOracle  = "3334b218-01e3-47a3-8a28-5cc2decbedea" // Woodweaver's Puzzleknot: {2}{G}, sac: gain 3, {E}{E}{E}
	slSharpshooterOracl = "d81285b7-a718-411a-8be3-ecc0cfe0bcb0" // Goblin Sharpshooter: {T}: 1 damage
	slLilianaOracle     = "0ba134d8-ee7d-48ec-8dc6-57942b8e9261" // Liliana of the Veil
)

// slToMain walks the turn to the active player's precombat main phase.
func slToMain(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; g.Turn.Step != game.StepPrecombatMain; i++ {
		if i > 12 {
			t.Fatalf("never reached the main phase (at %v)", g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// slActivateLent activates the borrower's first lent activated row, with
// the ref the view would send.
func slActivateLent(g *game.Game, p *game.Player, borrower uuid.UUID, params game.ActivateAbilityParams) error {
	var idx = -1
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(borrower)
		_, origins := game.ActivatedAbilitiesWithOrigins(c)
		for i, o := range origins {
			if o.Borrowed {
				idx, params.Ref = i, o.Ref
				return
			}
		}
	})
	if idx < 0 {
		return errors.New("no lent activated row")
	}
	return g.ActivateCatalogAbility(p.ID, borrower, idx, params)
}

func slPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c := gaLayered(t, g, id)
	p, _, _ := c.PowerToughness()
	return p
}

// --- Quicksilver Elemental ------------------------------------------------

func TestBorrowedLastQuicksilverKeepsTheTargetsAbilitiesUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qs := gaPut(g, me.ID, "Quicksilver Elemental", "Creature — Elemental", slQuicksilverOracle)
	shade := gaPut(g, opp.ID, "Evernight Shade", "Creature — Shade", slShadeOracle)
	if idx, _ := blBorrowedActivated(t, g, qs); len(idx) != 0 {
		t.Fatalf("the Elemental has %d lent rows before it activated", len(idx))
	}
	b06AddMana(me, "U")
	elActivate(t, g, me, qs, 0, shade)
	if idx, _ := blBorrowedActivated(t, g, qs); len(idx) != 1 {
		t.Fatalf("the Elemental has %d lent rows, want the Shade's 1", len(idx))
	}
	// The Shade leaves: what the Elemental gained stays (CR 608.2h).
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, shade); err != nil {
		t.Fatalf("move the Shade: %v", err)
	}
	if idx, _ := blBorrowedActivated(t, g, qs); len(idx) != 1 {
		t.Fatalf("the Shade left and the Elemental has %d lent rows, want 1", len(idx))
	}
	// Blue mana pays the lent {B}, and "this creature" is the Elemental.
	before := slPower(t, g, qs)
	b06AddMana(me, "U")
	if err := slActivateLent(g, me, qs, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pay the lent {B} with blue mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := slPower(t, g, qs); got != before+1 {
		t.Errorf("the Elemental's power is %d after the lent pump, want %d", got, before+1)
	}
	// Red mana is not blue: the rider names one colour.
	b06AddMana(me, "R")
	if err := slActivateLent(g, me, qs, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("red mana paid the lent {B}")
	}
	// Until end of turn: the next turn the row is gone.
	advancePastCleanupForTest(t, g)
	if idx, _ := blBorrowedActivated(t, g, qs); len(idx) != 0 {
		t.Errorf("the turn ended and the Elemental still has %d lent rows", len(idx))
	}
}

// The rider is about THIS creature's abilities: the Shade's own {B} is
// not payable with blue mana, even by the Elemental's controller.
func TestBorrowedLastQuicksilverRiderCoversOnlyItsOwnAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gaPut(g, me.ID, "Quicksilver Elemental", "Creature — Elemental", slQuicksilverOracle)
	shade := gaPut(g, me.ID, "Evernight Shade", "Creature — Shade", slShadeOracle)
	b06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("blue mana paid the Shade's own {B}")
	}
}

// ADR 0155 owner decision 2: a target the catalog has no entry for lends
// nothing, and the Elemental names it as not automated for the turn, even
// after it has gone.
func TestBorrowedLastQuicksilverNamesAnUnautomatedTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qs := gaPut(g, me.ID, "Quicksilver Elemental", "Creature — Elemental", slQuicksilverOracle)
	odd := gaPut(g, opp.ID, "Mystery Beast", "Creature — Beast", "no-such-oracle")
	b06AddMana(me, "U")
	elActivate(t, g, me, qs, 0, odd)
	if idx, _ := blBorrowedActivated(t, g, qs); len(idx) != 0 {
		t.Fatalf("an uncatalogued target lent %d rows", len(idx))
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, odd); err != nil {
		t.Fatalf("move the target: %v", err)
	}
	var named []*game.Card
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardPtrForEffect(qs)
		named = g.UnautomatedLendersOf(c)
	})
	if len(named) != 1 || named[0].InstanceID != odd || named[0].Name != "Mystery Beast" {
		t.Errorf("unautomated lenders = %+v, want the Mystery Beast", named)
	}
}

// A table holding a resolved borrow is a restore point, and the restored
// game still lends the row; a file naming a row this catalog does not
// have is refused, as an unknown effect key is.
func TestBorrowedLastQuicksilverBorrowRestores(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qs := gaPut(g, me.ID, "Quicksilver Elemental", "Creature — Elemental", slQuicksilverOracle)
	shade := gaPut(g, opp.ID, "Evernight Shade", "Creature — Shade", slShadeOracle)
	b06AddMana(me, "U")
	elActivate(t, g, me, qs, 0, shade)

	restored := restoreThroughJSON(t, g)
	if idx, _ := blBorrowedActivated(t, restored, qs); len(idx) != 1 {
		t.Fatalf("the restored Elemental has %d lent rows, want 1", len(idx))
	}

	raw := mustMarshal(t, g.CaptureSnapshot())
	bad := strings.Replace(string(raw), `"key":"`+slShadeOracle+`"`, `"key":"no-such-card"`, 1)
	if bad == string(raw) {
		t.Fatal("the snapshot names no lent row")
	}
	var snap game.GameSnapshot
	if err := json.Unmarshal([]byte(bad), &snap); err != nil {
		t.Fatal(err)
	}
	if _, err := snap.RestoreStrict(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Errorf("RestoreStrict of an unknown lent row = %v, want ErrUnknownEffectKey", err)
	}
}

// --- Grell Philosopher ----------------------------------------------------

func TestBorrowedLastGrellLendsTheArtifactToEachHorrorYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knot := gaPut(g, opp.ID, "Woodweaver's Puzzleknot", "Artifact", slPuzzleknotOracle)
	gaPut(g, me.ID, "Sol Ring", "Artifact", b15SolRingOracle) // mine: not a legal target
	horror := gaPut(g, me.ID, "Nameless Horror", "Creature — Horror", "")
	elf := blElves(g, me)
	grell := enterFromHand(t, g, me.ID, "Grell Philosopher", "Creature — Horror Wizard", slGrellOracle)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, knot)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{grell, horror} {
		if idx, _ := blBorrowedActivated(t, g, id); len(idx) != 1 {
			t.Fatalf("Horror %s has %d lent rows, want the Puzzleknot's 1", id, len(idx))
		}
	}
	if idx, _ := blBorrowedActivated(t, g, elf); len(idx) != 0 {
		t.Errorf("a non-Horror gained %d rows", len(idx))
	}
	// A Horror that arrives later gains nothing (CR 611.2c).
	late := gaPut(g, me.ID, "Late Horror", "Creature — Horror", "")
	if idx, _ := blBorrowedActivated(t, g, late); len(idx) != 0 {
		t.Errorf("a Horror that arrived after the resolution gained %d rows", len(idx))
	}
	// Grell leaves; the rider rides the rows: {2}{G} paid with blue alone,
	// and "sacrifice this artifact" sacrifices the Horror.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, grell); err != nil {
		t.Fatalf("move Grell: %v", err)
	}
	life := me.Life
	b06AddMana(me, "U", "U", "U")
	if err := slActivateLent(g, me, horror, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pay the lent {2}{G} with blue mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d", me.Life, life+3)
	}
	if z := g.FindCardZoneForEffect(horror); z == nil || z.Kind != game.ZoneGraveyard {
		t.Error("the lent sacrifice did not sacrifice the Horror")
	}
	if z := g.FindCardZoneForEffect(knot); z == nil || z.Kind != game.ZoneBattlefield {
		t.Error("the lender was sacrificed")
	}
}

// "And at the beginning of your upkeep": the same ability, a second
// condition.
func TestBorrowedLastGrellTriggersAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knot := gaPut(g, opp.ID, "Woodweaver's Puzzleknot", "Artifact", slPuzzleknotOracle)
	grell := gaPut(g, me.ID, "Grell Philosopher", "Creature — Horror Wizard", slGrellOracle)
	advanceToUpkeepOf(t, g, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, knot)
	passPriorityAroundTable(t, g)
	if idx, _ := blBorrowedActivated(t, g, grell); len(idx) != 1 {
		t.Fatalf("Grell has %d lent rows after its upkeep trigger, want 1", len(idx))
	}
}

// --- Havengul Lich --------------------------------------------------------

func TestBorrowedLastHavengulLichGainsTheAbilitiesOfTheCardItCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slToMain(t, g)
	lich := gaPut(g, me.ID, "Havengul Lich", "Creature — Zombie Wizard", slLichOracle)
	var shade uuid.UUID
	g.WithWriteLock(func() {
		c := game.Card{InstanceID: uuid.New(), Name: "Evernight Shade", TypeLine: "Creature — Shade",
			ManaCost: "{3}{B}", OracleID: slShadeOracle, Owner: opp.ID, Controller: opp.ID}
		shade = c.InstanceID
		opp.Graveyard.PushTop(c)
	})
	b06AddMana(me, "C")
	elActivate(t, g, me, lich, 0, shade)
	if idx, _ := blBorrowedActivated(t, g, lich); len(idx) != 0 {
		t.Fatalf("the Lich has %d lent rows before the card was cast", len(idx))
	}
	b06AddMana(me, "B", "C", "C", "C")
	if err := g.CastSpell(me.ID, shade, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast the Shade from the opponent's graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if idx, _ := blBorrowedActivated(t, g, lich); len(idx) != 1 {
		t.Fatalf("the Lich has %d lent rows after the cast, want the Shade's 1", len(idx))
	}
	before := slPower(t, g, lich)
	b06AddMana(me, "B")
	if err := slActivateLent(g, me, lich, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the lent pump: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := slPower(t, g, lich); got != before+1 {
		t.Errorf("the Lich's power is %d, want %d", got, before+1)
	}
}

// --- Conspicuous Snoop and Skill Borrower ---------------------------------

func slPutOnTop(g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// The top of the library changes with no event at all here (a bare push
// under the write lock): the watch DependsOnLibraryTop arms is what makes
// the Snoop notice.
func TestBorrowedLastConspicuousSnoopLendsAGoblinOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	snoop := gaPut(g, me.ID, "Conspicuous Snoop", "Creature — Goblin Rogue", slSnoopOracle)
	if idx, _ := blBorrowedActivated(t, g, snoop); len(idx) != 0 {
		t.Fatalf("a filler on top lent %d rows", len(idx))
	}
	slPutOnTop(g, me, "Goblin Sharpshooter", "Creature — Goblin", slSharpshooterOracl)
	idx, origins := blBorrowedActivated(t, g, snoop)
	if len(idx) != 1 {
		t.Fatalf("the Snoop has %d lent rows with a Goblin on top, want 1", len(idx))
	}
	label := game.BorrowedLabel("{T}: Goblin Sharpshooter deals 1 damage to any target", origins[0], "Conspicuous Snoop")
	if !strings.HasPrefix(label, "{T}: Conspicuous Snoop deals") {
		t.Errorf("label = %q, want the Sharpshooter's name read as the Snoop's (CR 201.5b)", label)
	}
	// A non-Goblin goes on top: nothing is lent.
	slPutOnTop(g, me, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	if idx, _ := blBorrowedActivated(t, g, snoop); len(idx) != 0 {
		t.Errorf("an Elf on top lent %d rows", len(idx))
	}
	if idx, _ := blBorrowedMana(t, g, snoop); len(idx) != 0 {
		t.Errorf("an Elf on top lent %d mana rows", len(idx))
	}
	// Drawn off the top: the Goblin is on top again.
	g.WithWriteLock(func() { _, _ = me.Library.PopTop() })
	if idx, _ := blBorrowedActivated(t, g, snoop); len(idx) != 1 {
		t.Errorf("the Goblin is back on top and %d rows are lent", len(idx))
	}
	// Every seat sees the top card, and the Snoop may cast a Goblin from it.
	var vis game.LibraryTopVisibility
	g.ReadSnapshot(func() { vis = g.LibraryTopVisibilityLocked(me.ID) })
	if vis != game.LibraryTopRevealed {
		t.Errorf("visibility = %v, want revealed", vis)
	}
}

// "You may cast Goblin spells from the top of your library": a Goblin is
// opened, a non-Goblin is not.
func TestBorrowedLastConspicuousSnoopCastsGoblinsFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushTopSource(t, g, me.ID, "Conspicuous Snoop", "Creature — Goblin Rogue", slSnoopOracle)
	bear := seedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if grantedPermissionOn(g, me.ID, bear, game.ZoneLibrary).Granted() {
		t.Error("a non-Goblin was opened")
	}
	goblin := seedTopOfLibrary(me, game.Card{Name: "Top Goblin", TypeLine: "Creature — Goblin", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, goblin, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast a Goblin off the top: %v", err)
	}
}

func TestBorrowedLastSkillBorrowerLendsAnArtifactOrCreatureOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	skill := gaPut(g, me.ID, "Skill Borrower", "Artifact Creature — Human Wizard", slSkillOracle)
	slPutOnTop(g, me, "Sol Ring", "Artifact", b15SolRingOracle)
	if idx, _ := blBorrowedMana(t, g, skill); len(idx) != 1 {
		t.Fatalf("Sol Ring on top lent %d mana rows, want 1", len(idx))
	}
	blTapLentMana(t, g, me, skill)
	if got := batch01PoolColors(me); len(got) != 2 {
		t.Errorf("the lent Sol Ring made %v, want two colourless", got)
	}
	slPutOnTop(g, me, "Forest", "Basic Land — Forest", "")
	if idx, _ := blBorrowedMana(t, g, skill); len(idx) != 0 {
		t.Errorf("a land on top lent %d mana rows", len(idx))
	}
}

// --- Scheming Fence -------------------------------------------------------

func slChoose(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			if err := g.ResolveChooseCards(c.ID, chooser, picks); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			return
		}
	}
	t.Fatal("no choose-cards prompt")
}

func TestBorrowedLastSchemingFenceTakesTheChosenPermanentsAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := gaPut(g, opp.ID, "Sol Ring", "Artifact", b15SolRingOracle)
	gaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	fence := enterFromHand(t, g, me.ID, "Scheming Fence", "Creature — Human Citizen", slFenceOracle)
	slChoose(t, g, me.ID, ring)
	passPriorityAroundTable(t, g)
	if got := gaLayered(t, g, fence).ChosenPermanent.ID; got != ring {
		t.Fatalf("chosen = %s, want the Sol Ring", got)
	}
	// The Fence has the Ring's mana ability, and the Ring can't use it.
	if idx, _ := blBorrowedMana(t, g, fence); len(idx) != 1 {
		t.Fatalf("the Fence has %d lent mana rows, want Sol Ring's 1", len(idx))
	}
	if err := g.ActivateManaAbility(opp.ID, ring, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the chosen Sol Ring tapped for mana")
	}
	// The Fence is summoning sick: the lent {T} is refused (CR 302.6).
	_, origins := blBorrowedMana(t, g, fence)
	idx, _ := blBorrowedMana(t, g, fence)
	if err := g.ActivateManaAbility(me.ID, fence, idx[0], game.ManaAbilityParams{Ref: origins[0].Ref}); err == nil {
		t.Error("a summoning-sick Fence tapped for the lent mana")
	}
	// The Ring leaves: nothing more is lent, and a new Ring is not chosen.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, ring); err != nil {
		t.Fatalf("move the Ring: %v", err)
	}
	if idx, _ := blBorrowedMana(t, g, fence); len(idx) != 0 {
		t.Errorf("the Ring left and %d rows are still lent", len(idx))
	}
}

func TestBorrowedLastSchemingFenceLeavesOutLoyaltyAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liliana := gaPut(g, opp.ID, "Liliana of the Veil", "Legendary Planeswalker — Liliana", slLilianaOracle)
	if err := g.AddCounter(liliana, game.CounterLoyalty, 3); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	fence := enterFromHand(t, g, me.ID, "Scheming Fence", "Creature — Human Citizen", slFenceOracle)
	slChoose(t, g, me.ID, liliana)
	passPriorityAroundTable(t, g)
	if idx, _ := blBorrowedActivated(t, g, fence); len(idx) != 0 {
		t.Errorf("the Fence borrowed %d of Liliana's loyalty abilities", len(idx))
	}
}

func TestBorrowedLastSchemingFenceMayChooseNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := gaPut(g, opp.ID, "Sol Ring", "Artifact", b15SolRingOracle)
	fence := enterFromHand(t, g, me.ID, "Scheming Fence", "Creature — Human Citizen", slFenceOracle)
	slChoose(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if idx, _ := blBorrowedMana(t, g, fence); len(idx) != 0 {
		t.Errorf("nothing chosen and %d rows are lent", len(idx))
	}
	if err := g.ActivateManaAbility(opp.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("nothing chosen and the Ring can't tap for mana: %v", err)
	}
}
