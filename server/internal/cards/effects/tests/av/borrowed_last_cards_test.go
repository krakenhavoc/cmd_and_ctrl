package testsav

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
	c := effectstest.GaLayered(t, g, id)
	p, _, _ := c.PowerToughness()
	return p
}

// --- Quicksilver Elemental ------------------------------------------------

func TestBorrowedLastQuicksilverKeepsTheTargetsAbilitiesUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qs := effectstest.GaPut(g, me.ID, "Quicksilver Elemental", "Creature — Elemental", effectstest.SlQuicksilverOracle)
	shade := effectstest.GaPut(g, opp.ID, "Evernight Shade", "Creature — Shade", effectstest.SlShadeOracle)
	if idx, _ := blBorrowedActivated(t, g, qs); len(idx) != 0 {
		t.Fatalf("the Elemental has %d lent rows before it activated", len(idx))
	}
	effectstest.B06AddMana(me, "U")
	effectstest.ElActivate(t, g, me, qs, 0, shade)
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
	effectstest.B06AddMana(me, "U")
	if err := slActivateLent(g, me, qs, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pay the lent {B} with blue mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := slPower(t, g, qs); got != before+1 {
		t.Errorf("the Elemental's power is %d after the lent pump, want %d", got, before+1)
	}
	// Red mana is not blue: the rider names one colour.
	effectstest.B06AddMana(me, "R")
	if err := slActivateLent(g, me, qs, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("red mana paid the lent {B}")
	}
	// Until end of turn: the next turn the row is gone.
	effectstest.AdvancePastCleanupForTest(t, g)
	if idx, _ := blBorrowedActivated(t, g, qs); len(idx) != 0 {
		t.Errorf("the turn ended and the Elemental still has %d lent rows", len(idx))
	}
}

// The rider is about THIS creature's abilities: the Shade's own {B} is
// not payable with blue mana, even by the Elemental's controller.
func TestBorrowedLastQuicksilverRiderCoversOnlyItsOwnAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.GaPut(g, me.ID, "Quicksilver Elemental", "Creature — Elemental", effectstest.SlQuicksilverOracle)
	shade := effectstest.GaPut(g, me.ID, "Evernight Shade", "Creature — Shade", effectstest.SlShadeOracle)
	effectstest.B06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("blue mana paid the Shade's own {B}")
	}
}

// ADR 0155 owner decision 2: a target the catalog has no entry for lends
// nothing, and the Elemental names it as not automated for the turn, even
// after it has gone.
func TestBorrowedLastQuicksilverNamesAnUnautomatedTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qs := effectstest.GaPut(g, me.ID, "Quicksilver Elemental", "Creature — Elemental", effectstest.SlQuicksilverOracle)
	odd := effectstest.GaPut(g, opp.ID, "Mystery Beast", "Creature — Beast", "no-such-oracle")
	effectstest.B06AddMana(me, "U")
	effectstest.ElActivate(t, g, me, qs, 0, odd)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qs := effectstest.GaPut(g, me.ID, "Quicksilver Elemental", "Creature — Elemental", effectstest.SlQuicksilverOracle)
	shade := effectstest.GaPut(g, opp.ID, "Evernight Shade", "Creature — Shade", effectstest.SlShadeOracle)
	effectstest.B06AddMana(me, "U")
	effectstest.ElActivate(t, g, me, qs, 0, shade)

	restored := effectstest.RestoreThroughJSON(t, g)
	if idx, _ := blBorrowedActivated(t, restored, qs); len(idx) != 1 {
		t.Fatalf("the restored Elemental has %d lent rows, want 1", len(idx))
	}

	raw := effectstest.MustMarshal(t, g.CaptureSnapshot())
	bad := strings.Replace(string(raw), `"key":"`+effectstest.SlShadeOracle+`"`, `"key":"no-such-card"`, 1)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knot := effectstest.GaPut(g, opp.ID, "Woodweaver's Puzzleknot", "Artifact", effectstest.SlPuzzleknotOracle)
	effectstest.GaPut(g, me.ID, "Sol Ring", "Artifact", effectstest.B15SolRingOracle) // mine: not a legal target
	horror := effectstest.GaPut(g, me.ID, "Nameless Horror", "Creature — Horror", "")
	elf := effectstest.BlElves(g, me)
	grell := effectstest.EnterFromHand(t, g, me.ID, "Grell Philosopher", "Creature — Horror Wizard", effectstest.SlGrellOracle)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, knot)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{grell, horror} {
		if idx, _ := blBorrowedActivated(t, g, id); len(idx) != 1 {
			t.Fatalf("Horror %s has %d lent rows, want the Puzzleknot's 1", id, len(idx))
		}
	}
	if idx, _ := blBorrowedActivated(t, g, elf); len(idx) != 0 {
		t.Errorf("a non-Horror gained %d rows", len(idx))
	}
	// A Horror that arrives later gains nothing (CR 611.2c).
	late := effectstest.GaPut(g, me.ID, "Late Horror", "Creature — Horror", "")
	if idx, _ := blBorrowedActivated(t, g, late); len(idx) != 0 {
		t.Errorf("a Horror that arrived after the resolution gained %d rows", len(idx))
	}
	// Grell leaves; the rider rides the rows: {2}{G} paid with blue alone,
	// and "sacrifice this artifact" sacrifices the Horror.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, grell); err != nil {
		t.Fatalf("move Grell: %v", err)
	}
	life := me.Life
	effectstest.B06AddMana(me, "U", "U", "U")
	if err := slActivateLent(g, me, horror, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pay the lent {2}{G} with blue mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knot := effectstest.GaPut(g, opp.ID, "Woodweaver's Puzzleknot", "Artifact", effectstest.SlPuzzleknotOracle)
	grell := effectstest.GaPut(g, me.ID, "Grell Philosopher", "Creature — Horror Wizard", effectstest.SlGrellOracle)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, knot)
	effectstest.PassPriorityAroundTable(t, g)
	if idx, _ := blBorrowedActivated(t, g, grell); len(idx) != 1 {
		t.Fatalf("Grell has %d lent rows after its upkeep trigger, want 1", len(idx))
	}
}

// --- Havengul Lich --------------------------------------------------------

func TestBorrowedLastHavengulLichGainsTheAbilitiesOfTheCardItCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slToMain(t, g)
	lich := effectstest.GaPut(g, me.ID, "Havengul Lich", "Creature — Zombie Wizard", effectstest.SlLichOracle)
	var shade uuid.UUID
	g.WithWriteLock(func() {
		c := game.Card{InstanceID: uuid.New(), Name: "Evernight Shade", TypeLine: "Creature — Shade",
			ManaCost: "{3}{B}", OracleID: effectstest.SlShadeOracle, Owner: opp.ID, Controller: opp.ID}
		shade = c.InstanceID
		opp.Graveyard.PushTop(c)
	})
	effectstest.B06AddMana(me, "C")
	effectstest.ElActivate(t, g, me, lich, 0, shade)
	if idx, _ := blBorrowedActivated(t, g, lich); len(idx) != 0 {
		t.Fatalf("the Lich has %d lent rows before the card was cast", len(idx))
	}
	effectstest.B06AddMana(me, "B", "C", "C", "C")
	if err := g.CastSpell(me.ID, shade, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast the Shade from the opponent's graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if idx, _ := blBorrowedActivated(t, g, lich); len(idx) != 1 {
		t.Fatalf("the Lich has %d lent rows after the cast, want the Shade's 1", len(idx))
	}
	before := slPower(t, g, lich)
	effectstest.B06AddMana(me, "B")
	if err := slActivateLent(g, me, lich, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the lent pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	snoop := effectstest.GaPut(g, me.ID, "Conspicuous Snoop", "Creature — Goblin Rogue", effectstest.SlSnoopOracle)
	if idx, _ := blBorrowedActivated(t, g, snoop); len(idx) != 0 {
		t.Fatalf("a filler on top lent %d rows", len(idx))
	}
	slPutOnTop(g, me, "Goblin Sharpshooter", "Creature — Goblin", effectstest.SlSharpshooterOracl)
	idx, origins := blBorrowedActivated(t, g, snoop)
	if len(idx) != 1 {
		t.Fatalf("the Snoop has %d lent rows with a Goblin on top, want 1", len(idx))
	}
	label := game.BorrowedLabel("{T}: Goblin Sharpshooter deals 1 damage to any target", origins[0], "Conspicuous Snoop")
	if !strings.HasPrefix(label, "{T}: Conspicuous Snoop deals") {
		t.Errorf("label = %q, want the Sharpshooter's name read as the Snoop's (CR 201.5b)", label)
	}
	// A non-Goblin goes on top: nothing is lent.
	slPutOnTop(g, me, "Llanowar Elves", "Creature — Elf Druid", effectstest.BlLlanowarOracle)
	if idx, _ := blBorrowedActivated(t, g, snoop); len(idx) != 0 {
		t.Errorf("an Elf on top lent %d rows", len(idx))
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, snoop); len(idx) != 0 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	effectstest.PushTopSource(t, g, me.ID, "Conspicuous Snoop", "Creature — Goblin Rogue", effectstest.SlSnoopOracle)
	bear := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if effectstest.GrantedPermissionOn(g, me.ID, bear, game.ZoneLibrary).Granted() {
		t.Error("a non-Goblin was opened")
	}
	goblin := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Goblin", TypeLine: "Creature — Goblin", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, goblin, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast a Goblin off the top: %v", err)
	}
}

func TestBorrowedLastSkillBorrowerLendsAnArtifactOrCreatureOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skill := effectstest.GaPut(g, me.ID, "Skill Borrower", "Artifact Creature — Human Wizard", effectstest.SlSkillOracle)
	slPutOnTop(g, me, "Sol Ring", "Artifact", effectstest.B15SolRingOracle)
	if idx, _ := effectstest.BlBorrowedMana(t, g, skill); len(idx) != 1 {
		t.Fatalf("Sol Ring on top lent %d mana rows, want 1", len(idx))
	}
	effectstest.BlTapLentMana(t, g, me, skill)
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 {
		t.Errorf("the lent Sol Ring made %v, want two colourless", got)
	}
	slPutOnTop(g, me, "Forest", "Basic Land — Forest", "")
	if idx, _ := effectstest.BlBorrowedMana(t, g, skill); len(idx) != 0 {
		t.Errorf("a land on top lent %d mana rows", len(idx))
	}
}

func TestBorrowedLastSchemingFenceTakesTheChosenPermanentsAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := effectstest.GaPut(g, opp.ID, "Sol Ring", "Artifact", effectstest.B15SolRingOracle)
	effectstest.GaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	fence := effectstest.EnterFromHand(t, g, me.ID, "Scheming Fence", "Creature — Human Citizen", effectstest.SlFenceOracle)
	effectstest.SlChoose(t, g, me.ID, ring)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.GaLayered(t, g, fence).ChosenPermanent.ID; got != ring {
		t.Fatalf("chosen = %s, want the Sol Ring", got)
	}
	// The Fence has the Ring's mana ability, and the Ring can't use it.
	if idx, _ := effectstest.BlBorrowedMana(t, g, fence); len(idx) != 1 {
		t.Fatalf("the Fence has %d lent mana rows, want Sol Ring's 1", len(idx))
	}
	if err := g.ActivateManaAbility(opp.ID, ring, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the chosen Sol Ring tapped for mana")
	}
	// The Fence is summoning sick: the lent {T} is refused (CR 302.6).
	_, origins := effectstest.BlBorrowedMana(t, g, fence)
	idx, _ := effectstest.BlBorrowedMana(t, g, fence)
	if err := g.ActivateManaAbility(me.ID, fence, idx[0], game.ManaAbilityParams{Ref: origins[0].Ref}); err == nil {
		t.Error("a summoning-sick Fence tapped for the lent mana")
	}
	// The Ring leaves: nothing more is lent, and a new Ring is not chosen.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, ring); err != nil {
		t.Fatalf("move the Ring: %v", err)
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, fence); len(idx) != 0 {
		t.Errorf("the Ring left and %d rows are still lent", len(idx))
	}
}

func TestBorrowedLastSchemingFenceLeavesOutLoyaltyAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liliana := effectstest.GaPut(g, opp.ID, "Liliana of the Veil", "Legendary Planeswalker — Liliana", effectstest.SlLilianaOracle)
	if err := g.AddCounter(liliana, game.CounterLoyalty, 3); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	fence := effectstest.EnterFromHand(t, g, me.ID, "Scheming Fence", "Creature — Human Citizen", effectstest.SlFenceOracle)
	effectstest.SlChoose(t, g, me.ID, liliana)
	effectstest.PassPriorityAroundTable(t, g)
	if idx, _ := blBorrowedActivated(t, g, fence); len(idx) != 0 {
		t.Errorf("the Fence borrowed %d of Liliana's loyalty abilities", len(idx))
	}
}

func TestBorrowedLastSchemingFenceMayChooseNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := effectstest.GaPut(g, opp.ID, "Sol Ring", "Artifact", effectstest.B15SolRingOracle)
	fence := effectstest.EnterFromHand(t, g, me.ID, "Scheming Fence", "Creature — Human Citizen", effectstest.SlFenceOracle)
	effectstest.SlChoose(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if idx, _ := effectstest.BlBorrowedMana(t, g, fence); len(idx) != 0 {
		t.Errorf("nothing chosen and %d rows are lent", len(idx))
	}
	if err := g.ActivateManaAbility(opp.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("nothing chosen and the Ring can't tap for mana: %v", err)
	}
}
