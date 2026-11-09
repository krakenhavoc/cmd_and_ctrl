package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// toby_beastie_befriender_test.go — #5396: Toby, Beastie Befriender, and
// its Beast's "This token can't attack or block alone." (CR 506.5,
// 508.1c, 509.1b; game.AttackLimitNotAlone, game.BlockRule.NotAlone).

const tobyOracle = "b4c27c89-fa50-4cf0-bf37-039732d16fef"

// tobyBlockTable: seat 0 attacks seat 1 with two vanilla 2/2s; seat 1
// holds one Beast and, with `bear`, a 2/2 Bear. The table stands in
// declare blockers.
func tobyBlockTable(t *testing.T, bear bool) (g *game.Game, def *game.Player, atks [2]uuid.UUID, beast, bearID uuid.UUID) {
	t.Helper()
	return tobyBlockTableWith(t, bear, apaCreature("Bear", "", 2, 2))
}

// tobyBlockTableWith is tobyBlockTable with the companion given.
func tobyBlockTableWith(t *testing.T, bear bool, companion game.Card) (g *game.Game, def *game.Player, atks [2]uuid.UUID, beast, bearID uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, def := g.Seats[0], g.Seats[1]
	atks[0] = pushVanillaCreature(g, me.ID, "Raider One", 2, 2)
	atks[1] = pushVanillaCreature(g, me.ID, "Raider Two", 2, 2)
	beast = apaPush(g, def.ID, def.ID, tobyBeastToken())
	if bear {
		bearID = apaPush(g, def.ID, def.ID, companion)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: atks[0], Target: def.ID},
		{Attacker: atks[1], Target: def.ID},
	}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	return g, def, atks, beast, bearID
}

func tobyBlocking(g *game.Game, id uuid.UUID) bool {
	c := findBattlefieldCardForTest(g, id)
	return c != nil && len(c.BlockedAttackers()) > 0
}

// tobyBlockOptions is the defender's block options, each as the set of
// blockers it names.
func tobyBlockOptions(g *game.Game, seat uuid.UUID) [][]uuid.UUID {
	var out [][]uuid.UUID
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, o := range g.BlockOptionsLocked(seat, 4) {
			var ids []uuid.UUID
			for _, d := range o.Blocks {
				ids = append(ids, d.Blocker)
			}
			out = append(out, ids)
		}
	})
	return out
}

func TestTobyCreatesTheBeastAsPrinted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castAndResolveCreature(t, g, "Toby, Beastie Befriender", "Legendary Creature — Human Wizard", tobyOracle)
	passPriorityAroundTable(t, g)
	var beast *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.IsToken() && c.Name == "Beast" && c.Controller == me.ID {
			beast = c
		}
	}
	if beast == nil {
		t.Fatal("no Beast token entered")
	}
	eff := beast.Effective()
	if eff.Power != 4 || eff.Toughness != 4 || len(eff.Colors) != 1 || eff.Colors[0] != "W" {
		t.Errorf("Beast is %d/%d %v, want a 4/4 white", eff.Power, eff.Toughness, eff.Colors)
	}
	if !game.CantBlockAlone(beast) {
		t.Error("the Beast can block alone")
	}
	if got := game.TokenTextForCard(*beast); got != "This token can't attack or block alone." {
		t.Errorf("token text %q", got)
	}
}

// Alone, the Beast's block is refused at the verb and it is not offered
// alone; with no legal block the defender's declaration completed as the
// step began, and nothing was stored.
func TestTobyBeastCantBlockAlone(t *testing.T) {
	g, def, atks, beast, _ := tobyBlockTable(t, false)

	if g.BlockDeclarationStatusOf(def.ID) != game.BlockDeclarationDeclared {
		t.Error("a defender whose only creature can't block alone was left declaring")
	}
	err := g.DeclareBlocker(beast, atks[0])
	var re *game.BlockRefusedError
	if !errors.As(err, &re) || re.Reason != game.BlockReasonCantBlockAlone || re.Blocker != beast {
		t.Fatalf("declaring the Beast alone: %v, want cant_block_alone", err)
	}
	if got := re.Sentence(def.ID); got != "Beast can't block alone: declare another blocker with it." {
		t.Errorf("sentence %q", got)
	}
	if !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("refusal does not wrap ErrIllegalBlock: %v", err)
	}
	if tobyBlocking(g, beast) {
		t.Error("a refused block was stored")
	}
	if opts := tobyBlockOptions(g, def.ID); len(opts) != 0 {
		t.Errorf("block options offered with the Beast alone: %v", opts)
	}
}

// The view's block_refusals: while nothing is staged, every attacker
// is refused to the Beast as cant_block_alone; once the Bear blocks,
// none is.
func TestTobyBeastBlockRefusalsView(t *testing.T) {
	g, _, atks, beast, bear := tobyBlockTable(t, true)
	refusals := func() []game.LoneBlockRefusal {
		var out []game.LoneBlockRefusal
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			out = g.LoneBlockRefusalsForEffect()[beast]
		})
		return out
	}
	got := refusals()
	if len(got) != 2 {
		t.Fatalf("view refusals %+v, want both attackers", got)
	}
	for _, r := range got {
		if r.Reason != game.BlockReasonCantBlockAlone || r.Sentence != "Beast can't block alone: declare another blocker with it." {
			t.Errorf("view refusal %+v", r)
		}
	}
	if err := g.DeclareBlocker(bear, atks[0]); err != nil {
		t.Fatalf("Bear block: %v", err)
	}
	if got := refusals(); len(got) != 0 {
		t.Errorf("view refusals beside the Bear: %+v", got)
	}
}

// Beside another blocker of its controller's it blocks, on another
// attacker or the same one; declared first, the companion makes it
// legal.
func TestTobyBeastBlocksBesideAnotherBlocker(t *testing.T) {
	g, def, atks, beast, bear := tobyBlockTable(t, true)

	// The enumerator offers the Beast with a companion, never alone.
	sawGroup := false
	for _, ids := range tobyBlockOptions(g, def.ID) {
		if len(ids) == 1 && ids[0] == beast {
			t.Errorf("the Beast is offered alone: %v", ids)
		}
		if len(ids) == 2 && (ids[0] == beast || ids[1] == beast) {
			sawGroup = true
		}
	}
	if !sawGroup {
		t.Error("the Beast is never offered with a companion")
	}

	if err := g.DeclareBlocker(bear, atks[0]); err != nil {
		t.Fatalf("Bear block: %v", err)
	}
	if err := g.DeclareBlocker(beast, atks[1]); err != nil {
		t.Fatalf("the Beast beside the Bear was refused: %v", err)
	}
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if !tobyBlocking(g, beast) || !tobyBlocking(g, bear) {
		t.Error("both should be blocking")
	}
}

func TestTobyBeastBlocksAsAGroup(t *testing.T) {
	g, def, atks, beast, bear := tobyBlockTable(t, true)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: beast, Attacker: atks[0]},
		{Blocker: bear, Attacker: atks[0]},
	}); err != nil {
		t.Fatalf("the Beast and the Bear together: %v", err)
	}
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if !tobyBlocking(g, beast) || !tobyBlocking(g, bear) {
		t.Error("both should be blocking")
	}
}

// Its companion taken back before the declaration completes, the Beast
// is taken out of the declaration rather than announced alone.
func TestTobyBeastLeftAloneIsTakenOutAtCompletion(t *testing.T) {
	g, def, atks, beast, bear := tobyBlockTable(t, true)
	if err := g.DeclareBlocker(bear, atks[0]); err != nil {
		t.Fatalf("Bear block: %v", err)
	}
	if err := g.DeclareBlocker(beast, atks[1]); err != nil {
		t.Fatalf("Beast block: %v", err)
	}
	if err := g.RemoveFromCombat(bear); err != nil {
		t.Fatalf("taking the Bear back: %v", err)
	}
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if tobyBlocking(g, beast) {
		t.Error("the Beast blocks alone")
	}
	if g.BlockedAttackerForEffect(atks[1]) {
		t.Error("the attacker the lone Beast was pointed at reads blocked")
	}
}

// Two creatures that each can't block alone may block together.
func TestTwoTobyBeastsBlockTogether(t *testing.T) {
	g, def, atks, beast, second := tobyBlockTableWith(t, true, tobyBeastToken())
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: beast, Attacker: atks[0]},
		{Blocker: second, Attacker: atks[1]},
	}); err != nil {
		t.Fatalf("two Beasts together: %v", err)
	}
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if !tobyBlocking(g, beast) || !tobyBlocking(g, second) {
		t.Error("both Beasts should be blocking")
	}
}

// Another player's blocker is no company (CR 802.4b).
func TestTobyBeastCountsOnlyItsControllersBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, def, other := g.Seats[0], g.Seats[1], g.Seats[2]
	a1 := pushVanillaCreature(g, me.ID, "Raider One", 2, 2)
	a2 := pushVanillaCreature(g, me.ID, "Raider Two", 2, 2)
	beast := apaPush(g, def.ID, def.ID, tobyBeastToken())
	bear := apaPush(g, def.ID, def.ID, apaCreature("Bear", "", 2, 2))
	wall := pushVanillaCreature(g, other.ID, "Wall", 0, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: a1, Target: def.ID},
		{Attacker: a2, Target: other.ID},
	}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, a2); err != nil {
		t.Fatalf("the other player's block: %v", err)
	}
	var re *game.BlockRefusedError
	if err := g.DeclareBlocker(beast, a1); !errors.As(err, &re) || re.Reason != game.BlockReasonCantBlockAlone {
		t.Fatalf("the Beast beside another player's blocker: %v, want cant_block_alone", err)
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: beast, Attacker: a1}, {Blocker: bear, Attacker: a1}}); err != nil {
		t.Fatalf("the Beast beside its controller's Bear: %v", err)
	}
}

// The attack half: alone it is refused, beside another attacker it
// attacks.
func TestTobyBeastCantAttackAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := apaPush(g, me.ID, me.ID, tobyBeastToken())
	advanceTo(t, g, game.StepDeclareAttackers)
	err := g.DeclareAttacker(beast, opp.ID)
	var le *game.AttackLimitError
	if !errors.As(err, &le) || le.Scope != game.AttackLimitNotAlone {
		t.Fatalf("the Beast attacking alone: %v, want the can't-attack-alone refusal", err)
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	beast2 := apaPush(g2, me2.ID, me2.ID, tobyBeastToken())
	bear := apaPush(g2, me2.ID, me2.ID, apaCreature("Bear", "", 2, 2))
	advanceTo(t, g2, game.StepDeclareAttackers)
	if _, err := g2.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: beast2, Target: opp2.ID},
		{Attacker: bear, Target: g2.Seats[2].ID},
	}); err != nil {
		t.Fatalf("the Beast beside the Bear: %v", err)
	}
	if err := g2.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if c := findBattlefieldCardForTest(g2, beast2); c == nil || c.AttackingTarget == uuid.Nil {
		t.Error("the Beast is not attacking")
	}
}

// Four creature tokens: they fly, Toby does not; three: none does.
func TestTobyFlyingAtFourCreatureTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toby := apaPush(g, me.ID, me.ID, game.Card{Name: "Toby, Beastie Befriender", OracleID: tobyOracle,
		TypeLine: "Legendary Creature — Human Wizard", Power: 1, Toughness: 1})
	var tokens []uuid.UUID
	for i := 0; i < 3; i++ {
		tokens = append(tokens, apaPush(g, me.ID, me.ID, tobyBeastToken()))
	}
	// An opponent's creature token and a noncreature token of mine do
	// not count.
	theirs := apaPush(g, opp.ID, opp.ID, tobyBeastToken())
	apaPush(g, me.ID, me.ID, TreasureToken())
	for _, id := range tokens {
		if game.HasKeyword(apaLive(g, id), "flying") {
			t.Fatal("a token flies at three creature tokens")
		}
	}
	tokens = append(tokens, apaPush(g, me.ID, me.ID, TokenCard("1/1 white Soldier")))
	for _, id := range tokens {
		if !game.HasKeyword(apaLive(g, id), "flying") {
			t.Errorf("token %v does not fly at four creature tokens", id)
		}
	}
	if game.HasKeyword(apaLive(g, toby), "flying") {
		t.Error("Toby, not a token, flies")
	}
	if game.HasKeyword(apaLive(g, theirs), "flying") {
		t.Error("an opponent's token flies")
	}
}
