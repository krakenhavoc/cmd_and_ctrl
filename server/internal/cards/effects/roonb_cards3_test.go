package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roonb_cards3_test.go — #4128 slice b: See Double.

func roonbSeeDoubleCast(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "See Double", TypeLine: "Instant", OracleID: roonbSeeDoubleOr,
		Owner: active.ID, Controller: active.ID,
	})
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

func roonbFillGraveyard(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

// Bullet two: a token that's a copy of target creature, under the
// spell's controller, with the creature's abilities.
func TestRoonbSeeDoubleTokenCopiesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := seedCopyableCreature(g, opp.ID, "Serra Angel", "Creature — Angel", 4, 4)
	before := len(g.Battlefield.Cards)
	if _, err := roonbSeeDoubleCast(t, g, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: angel}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.Battlefield.Cards) != before+1 {
		t.Fatalf("battlefield grew by %d, want 1 token", len(g.Battlefield.Cards)-before)
	}
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID != angel && c.Name == "Serra Angel" {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil || !token.IsToken() || token.Controller != me.ID {
		t.Fatalf("copy = %+v, want a Serra Angel token under the caster", token)
	}
	if token.CurrentPower() != 4 || token.CurrentToughness() != 4 {
		t.Errorf("token is %d/%d, want 4/4", token.CurrentPower(), token.CurrentToughness())
	}
	if roonbControllerOf(g, angel) != opp.ID {
		t.Error("the original should be untouched")
	}
}

// Bullet one: copy target spell, with the retarget prompt; the copy is
// controlled by See Double's controller and a copy of a permanent spell
// would become a token (ADR 0043), so it is exercised on an instant here.
func TestRoonbSeeDoubleCopiesASpellAndRetargetsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	lifeA, lifeB := lifeOf(g, victimA), lifeOf(g, victimB)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	if _, err := roonbSeeDoubleCast(t, g, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g.ResolvePickTarget(prompt.ID, me, game.TargetRef{Kind: game.TargetPlayer, ID: victimB}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, victimA); got != lifeA-3 {
		t.Errorf("the original Bolt: life %d, want %d", got, lifeA-3)
	}
	if got := lifeOf(g, victimB); got != lifeB-3 {
		t.Errorf("the copy: life %d, want %d", got, lifeB-3)
	}
}

// "Choose one" unless an opponent has eight or more cards in their
// graveyard; both then (read when the modes are chosen).
func TestRoonbSeeDoubleBothNeedsAnOpponentWithEightGraveyardCards(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	angel := seedCopyableCreature(g, opp.ID, "Serra Angel", "Creature — Angel", 4, 4)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	both := []game.TargetRef{
		{Kind: game.TargetCard, ID: bolt, Mode: 0},
		{Kind: game.TargetCard, ID: angel, Mode: 1},
	}
	roonbFillGraveyard(opp, 7)
	if _, err := roonbSeeDoubleCast(t, g, []int{0, 1}, both); err == nil {
		t.Fatal("seven cards in an opponent's graveyard: choosing both must be refused")
	}
	roonbFillGraveyard(opp, 1)
	if _, err := roonbSeeDoubleCast(t, g, []int{0, 1}, both); err != nil {
		t.Fatalf("eight cards in an opponent's graveyard: both is legal: %v", err)
	}
}

// Your own graveyard does not count: "an opponent".
func TestRoonbSeeDoubleCountsOnlyOpponentsGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := seedCopyableCreature(g, opp.ID, "Serra Angel", "Creature — Angel", 4, 4)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	roonbFillGraveyard(me, 12)
	if _, err := roonbSeeDoubleCast(t, g, []int{0, 1}, []game.TargetRef{
		{Kind: game.TargetCard, ID: bolt, Mode: 0},
		{Kind: game.TargetCard, ID: angel, Mode: 1},
	}); err == nil {
		t.Fatal("the caster's own graveyard must not enable both")
	}
}

// "This spell can't be copied": a Twincast aimed at it resolves and
// makes nothing, and asks nobody to retarget.
func TestRoonbSeeDoubleCantBeCopied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := seedCopyableCreature(g, opp.ID, "Serra Angel", "Creature — Angel", 4, 4)
	// A second creature, so a copy WOULD have somewhere to retarget to
	// and the CR 707.10c prompt would open if the copy were made.
	seedCopyableCreature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	seeDouble, err := roonbSeeDoubleCast(t, g, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: angel}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	castCatalogSpell(t, g, "Twincast", "Instant", twincastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: seeDouble}})
	passPriorityAroundTable(t, g)
	if n := stackCopies(g); n != 0 {
		t.Errorf("%d copies on the stack, want none", n)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a retarget prompt opened for a copy that cannot exist")
	}
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Serra Angel" {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("%d Serra Angel tokens, want exactly the one See Double made", tokens)
	}
}

// The keep-the-targets copy door (Chef's Kiss) refuses it as well.
func TestRoonbSeeDoubleRefusesTheKeepTargetsCopyDoor(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	angel := seedCopyableCreature(g, opp.ID, "Serra Angel", "Creature — Angel", 4, 4)
	seeDouble, err := roonbSeeDoubleCast(t, g, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: angel}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	var copyErr error
	g.WithWriteLock(func() {
		_, copyErr = g.CopySpellKeepingTargetsForEffect(seeDouble, g.Seats[0].ID)
	})
	if !errors.Is(copyErr, game.ErrCardNotFound) {
		t.Fatalf("CopySpellKeepingTargetsForEffect = %v, want ErrCardNotFound", copyErr)
	}
	if n := stackCopies(g); n != 0 {
		t.Errorf("%d copies on the stack", n)
	}
}

// Control: a card WITHOUT the rider is copied by the same door, so the
// two tests above can only be passing on the rider.
func TestRoonbOtherSpellsAreStillCopiedByTheKeepTargetsDoor(t *testing.T) {
	g := newCatalogGame(t)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	var copyErr error
	g.WithWriteLock(func() {
		_, copyErr = g.CopySpellKeepingTargetsForEffect(bolt, g.Seats[0].ID)
	})
	if copyErr != nil {
		t.Fatalf("a Lightning Bolt is copyable: %v", copyErr)
	}
	if n := stackCopies(g); n != 1 {
		t.Errorf("%d copies on the stack, want 1", n)
	}
}
