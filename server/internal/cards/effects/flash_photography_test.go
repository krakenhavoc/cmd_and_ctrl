package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flash_photography_test.go — Flash Photography (#5363): "You may cast
// this spell as though it had flash if it targets a permanent you
// control" (Spec.FlashIf, CR 601.3b), a token copy of target permanent
// (CR 707.2) and flashback {4}{U}{U} (CR 702.34a), whose graveyard cast
// the permission opens at instant speed too.

const flashPhotographyOracle = "a92ae7b0-cbbb-44a8-bfd8-2a68f4c1ce1a"

// fpTable is a two-seat-or-more catalog table in `me`'s precombat main,
// with a bear on each side of it.
func fpTable(t *testing.T) (g *game.Game, me *game.Player, mine, theirs uuid.UUID) {
	t.Helper()
	g, me, opp := p7Table(t)
	mine = b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs = b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	return g, me, mine, theirs
}

func fpHand(g *game.Game, me *game.Player) uuid.UUID {
	return flashIfHandCard(g, me, "Flash Photography", "Sorcery", "{2}{U}{U}", flashPhotographyOracle)
}

func fpGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Flash Photography", TypeLine: "Sorcery",
			ManaCost: "{2}{U}{U}", OracleID: flashPhotographyOracle, Owner: me.ID, Controller: me.ID})
	})
	return id
}

func fpCast(target uuid.UUID, alt string) game.CastSpellParams {
	p := game.CastSpellParams{Strict: true, AlternativeCost: alt,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}
	if alt == "flashback" {
		p.FromZone = "graveyard"
	}
	return p
}

// fpTokenCopies counts the tokens named `name` controlled by `who`.
func fpTokenCopies(g *game.Game, who uuid.UUID, name string) int {
	n := 0
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && IsToken(c) && c.Controller == who {
				n++
			}
		}
	})
	return n
}

// In a main phase the printed "target permanent" stands: a copy of an
// opponent's permanent, under the caster's control.
func TestFlashPhotographySorcerySpeedCopiesAnOpponentsPermanent(t *testing.T) {
	g, me, _, theirs := fpTable(t)
	photo := fpHand(g, me)
	flashIfFund(g, me, "U", "U", "C", "C")
	if err := g.CastSpell(me.ID, photo, fpCast(theirs, "")); err != nil {
		t.Fatalf("main-phase cast onto an opponent's creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := fpTokenCopies(g, me.ID, "Hill Giant"); got != 1 {
		t.Errorf("token copies of the Hill Giant under the caster: %d, want 1", got)
	}
	if !me.Graveyard.Contains(photo) {
		t.Error("Flash Photography is not in its owner's graveyard after resolving")
	}
}

// At instant speed only the card's own permission opens the window, so
// the cast must target a permanent its caster controls (CR 601.3b): an
// opponent's is refused with nothing paid, the caster's own is copied.
func TestFlashPhotographyInstantSpeedOnlyTargetingYourOwn(t *testing.T) {
	g, me, mine, theirs := fpTable(t)
	photo := fpHand(g, me)
	ssToBeginCombat(t, g)
	flashIfFund(g, me, "U", "U", "C", "C")
	pool := len(me.ManaPool)
	if err := g.CastSpell(me.ID, photo, fpCast(theirs, "")); err == nil {
		t.Fatal("instant-speed cast onto an opponent's creature was accepted")
	}
	if g.Stack.Contains(photo) || len(me.ManaPool) != pool {
		t.Fatalf("a refused cast moved the card or spent mana (pool %d)", len(me.ManaPool))
	}
	if err := g.CastSpell(me.ID, photo, fpCast(mine, "")); err != nil {
		t.Fatalf("instant-speed cast onto the caster's own creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := fpTokenCopies(g, me.ID, "Grizzly Bears"); got != 1 {
		t.Errorf("token copies of the bear: %d, want 1", got)
	}
}

// The permission reads no zone: flashback from the graveyard at instant
// speed targeting the caster's own permanent, refused onto an
// opponent's, and the card is exiled as it leaves the stack.
func TestFlashPhotographyFlashbackAtInstantSpeedTargetingYourOwn(t *testing.T) {
	g, me, mine, theirs := fpTable(t)
	photo := fpGraveyard(g, me)
	ssToBeginCombat(t, g)
	flashIfFund(g, me, "U", "U", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, photo, fpCast(theirs, "flashback")); err == nil {
		t.Fatal("instant-speed flashback onto an opponent's creature was accepted")
	}
	if err := g.CastSpell(me.ID, photo, fpCast(mine, "flashback")); err != nil {
		t.Fatalf("instant-speed flashback onto the caster's own creature: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana left after flashback: %d, want 0", len(me.ManaPool))
	}
	passPriorityAroundTable(t, g)
	if got := fpTokenCopies(g, me.ID, "Grizzly Bears"); got != 1 {
		t.Errorf("token copies of the bear: %d, want 1", got)
	}
	if !g.Exile.Contains(photo) {
		t.Error("the flashed-back Flash Photography is not in exile")
	}
}

// Flashback in a main phase takes any permanent, as the printed clause.
func TestFlashPhotographyFlashbackAtSorcerySpeedTakesAnyPermanent(t *testing.T) {
	g, me, _, theirs := fpTable(t)
	photo := fpGraveyard(g, me)
	flashIfFund(g, me, "U", "U", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, photo, fpCast(theirs, "flashback")); err != nil {
		t.Fatalf("main-phase flashback onto an opponent's creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := fpTokenCopies(g, me.ID, "Hill Giant"); got != 1 {
		t.Errorf("token copies of the Hill Giant: %d, want 1", got)
	}
}

// Strict refusals: flashback is not a hand cast, a graveyard cast must
// claim flashback, it pays the flashback cost in full, and the instant-
// speed window stays shut on an opponent's turn without a target of
// the caster's own.
func TestFlashPhotographyStrictRefusals(t *testing.T) {
	t.Run("flashback from hand", func(t *testing.T) {
		g, me, mine, _ := fpTable(t)
		photo := fpHand(g, me)
		flashIfFund(g, me, "U", "U", "C", "C", "C", "C")
		if err := g.CastSpell(me.ID, photo, fpCast(mine, "flashback")); err == nil {
			t.Fatal("flashback claimed from hand was accepted")
		}
	})
	t.Run("graveyard cast for the printed cost", func(t *testing.T) {
		g, me, mine, _ := fpTable(t)
		photo := fpGraveyard(g, me)
		flashIfFund(g, me, "U", "U", "C", "C")
		p := fpCast(mine, "")
		p.FromZone = "graveyard"
		if err := g.CastSpell(me.ID, photo, p); err == nil {
			t.Fatal("a graveyard cast claiming nothing was accepted")
		}
	})
	t.Run("flashback short of mana", func(t *testing.T) {
		g, me, mine, _ := fpTable(t)
		photo := fpGraveyard(g, me)
		flashIfFund(g, me, "U", "U", "C", "C")
		if err := g.CastSpell(me.ID, photo, fpCast(mine, "flashback")); err == nil {
			t.Fatal("flashback paid with {2}{U}{U} was accepted")
		}
		if !me.Graveyard.Contains(photo) {
			t.Error("a refused flashback moved the card")
		}
	})
	t.Run("no target at instant speed", func(t *testing.T) {
		g, me, _, _ := fpTable(t)
		photo := fpHand(g, me)
		ssToBeginCombat(t, g)
		flashIfFund(g, me, "U", "U", "C", "C")
		if err := g.CastSpell(me.ID, photo, game.CastSpellParams{Strict: true}); err == nil {
			t.Fatal("an instant-speed cast with no target was accepted")
		}
	})
}

// The token copies the copiable values (CR 707.2), so the copy's enters
// ability triggers: a token Mulldrifter draws two.
func TestFlashPhotographyCopyTriggersItsEntersAbility(t *testing.T) {
	g, me, _, _ := fpTable(t)
	drifter := pushFlickerCreature(g, me.ID, "Mulldrifter", mulldrifterOracle)
	photo := fpHand(g, me)
	flashIfFund(g, me, "U", "U", "C", "C")
	before := len(me.Hand.Cards)
	if err := g.CastSpell(me.ID, photo, fpCast(drifter, "")); err != nil {
		t.Fatalf("cast onto Mulldrifter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := fpTokenCopies(g, me.ID, "Mulldrifter"); got != 1 {
		t.Fatalf("token copies of Mulldrifter: %d, want 1", got)
	}
	// The cast took Flash Photography out of the hand; the trigger drew two.
	if got, want := len(me.Hand.Cards), before-1+2; got != want {
		t.Errorf("hand size %d, want %d (the token's enters trigger draws two)", got, want)
	}
}

// Registered Full, with the permission's narrower clause and flashback.
func TestFlashPhotographyIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(flashPhotographyOracle)
	if !ok {
		t.Fatal("Flash Photography is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v caveats %v, want Full with none", spec.Completeness, spec.Caveats)
	}
	if spec.FlashIf == nil || spec.FlashIf.Targets == nil {
		t.Error("no FlashIf target clause")
	}
	if fi := game.FlashIfFor(flashPhotographyOracle); fi == nil {
		t.Error("the engine sees no FlashIf for the card")
	}
	var fb bool
	for _, ac := range spec.AlternativeCosts {
		fb = fb || ac.Key == "flashback"
	}
	if !fb {
		t.Error("no flashback offer")
	}
}
