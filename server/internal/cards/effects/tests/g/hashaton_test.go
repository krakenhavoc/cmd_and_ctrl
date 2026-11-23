package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- reanimation -------------------------------------------------

// TestReanimateTakesAnOpponentsCreature is the controller fix. "Put
// target creature card from A GRAVEYARD onto the battlefield UNDER
// YOUR CONTROL" is the whole reason to play Reanimate in a deck full
// of removal, and it used to hand the creature back to the opponent.
func TestReanimateTakesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	victim := effectstest.SeedGraveyardCreature(opp, "Bulky Zombie", "{4}{B}")
	lifeBefore := me.Life

	effectstest.CastCatalogSpell(t, g, "Reanimate", "Sorcery", effectstest.ReanimateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)

	card, ok := effectstest.BattlefieldCard(g, victim)
	if !ok {
		t.Fatal("the reanimated creature is not on the battlefield")
	}
	if card.Controller != me.ID {
		t.Errorf("controller is not the caster: got %s, want %s", card.Controller, me.ID)
	}
	if card.Owner != opp.ID {
		t.Errorf("ownership should not change: got %s, want %s", card.Owner, opp.ID)
	}
	// {4}{B} is mana value 5.
	if got := lifeBefore - me.Life; got != 5 {
		t.Errorf("life paid: got %d, want 5 (the card's mana value)", got)
	}
	if opp.Graveyard.Contains(victim) {
		t.Error("the creature is still in the opponent's graveyard")
	}
}

// TestZombifyReturnsYourOwnCreature — the same move one word
// narrower. Owner and controller coincide, and no life is paid.
func TestZombifyReturnsYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.SeedGraveyardCreature(me, "My Zombie", "{4}{B}")
	lifeBefore := me.Life

	effectstest.CastCatalogSpell(t, g, "Zombify", "Sorcery", effectstest.ZombifyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)

	card, ok := effectstest.BattlefieldCard(g, mine)
	if !ok {
		t.Fatal("the reanimated creature is not on the battlefield")
	}
	if card.Controller != me.ID {
		t.Errorf("controller: got %s, want %s", card.Controller, me.ID)
	}
	if me.Life != lifeBefore {
		t.Errorf("Zombify costs no life: got %d, want %d", me.Life, lifeBefore)
	}
}

// TestLateToDinnerReanimatesAndMakesFood — both halves are printed
// unconditionally, so the Food is not contingent on the creature.
func TestLateToDinnerReanimatesAndMakesFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.SeedGraveyardCreature(me, "Dinner Guest", "{2}{W}")

	effectstest.CastCatalogSpell(t, g, "Late to Dinner", "Sorcery", effectstest.LateToDinnerOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)

	if _, ok := effectstest.BattlefieldCard(g, mine); !ok {
		t.Error("the reanimated creature is not on the battlefield")
	}
	if effectstest.FindBattlefieldByName(g, "Food") == uuid.Nil {
		t.Error("no Food token was created")
	}
}

// TestReanimationFiresTheAsEntersHook pins the second of the engine's
// two independently-wired entry mechanisms. Triggered abilities
// harvest off the emitted EventETB and always worked; Spec.AsEnters
// runs through fireETBHookLocked, which this path once never called —
// so a reanimated Solemn Simulacrum fetched nothing.
//
// Vanquisher's Banner is the probe because its entry behaviour lives
// entirely in AsEnters (the CR 614.12 creature-type choice) and
// nowhere else — the Temples, the previous probe, moved their scry
// onto the stack in #578. It also double-checks the controller fix:
// the prompt is queued for the Banner's CONTROLLER, so it arriving
// for the reanimating player rather than the graveyard's owner is
// the same assertion twice over.

func TestReanimationFiresTheAsEntersHook(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	id := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Vanquisher's Banner",
		TypeLine:   "Artifact",
		OracleID:   effectstest.VanquishersBannerOracle,
		Owner:      opp.ID,
		Controller: opp.ID,
	})

	var err error
	g.WithWriteLock(func() {
		err = g.ReturnFromGraveyardUnderControlForEffect(id, game.ZoneBattlefield, me.ID)
	})
	if err != nil {
		t.Fatalf("ReturnFromGraveyardUnderControlForEffect: %v", err)
	}

	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("the Banner is not on the battlefield")
	}
	if card.Controller != me.ID {
		t.Errorf("controller: got %s, want %s", card.Controller, me.ID)
	}
	if effectstest.ReanimatedTypeChoiceFor(g, me.ID) == nil {
		t.Error("the Banner's AsEnters hook did not fire — no creature-type prompt for the new controller")
	}
	if effectstest.ReanimatedTypeChoiceFor(g, opp.ID) != nil {
		t.Error("the prompt was queued for the graveyard's owner instead of the new controller")
	}
}

func TestHashatonCopiesADiscardedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.HashatonDiscardOgre(t, g, me)

	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("discarding a creature did not offer the {2}{U} payment")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)

	tokenID := effectstest.FindBattlefieldByName(g, "Test Ogre")
	if tokenID == uuid.Nil {
		t.Fatal("paying {2}{U} created no token copy")
	}
	tok, _ := effectstest.BattlefieldCard(g, tokenID)
	if tok.Power != 4 || tok.Toughness != 4 {
		t.Errorf("token is %d/%d, want 4/4", tok.Power, tok.Toughness)
	}
	if !tok.Tapped {
		t.Error("the token should enter tapped")
	}
	if tok.Controller != me.ID {
		t.Errorf("token controller: got %s, want %s", tok.Controller, me.ID)
	}
	if len(tok.Colors) != 1 || tok.Colors[0] != "B" {
		t.Errorf("token colours: got %v, want [B]", tok.Colors)
	}
	if tok.OracleID != effectstest.TestOgreOracle {
		t.Errorf("token oracle ID: got %q, want the copied card's %q", tok.OracleID, effectstest.TestOgreOracle)
	}
	// "except it's a 4/4 black Zombie" REPLACES the creature types
	// and keeps the supertypes.
	if !effects.ContainsFoldASCII(tok.TypeLine, "token") {
		t.Errorf("token type line %q is missing the Token supertype", tok.TypeLine)
	}
	if !effects.ContainsFoldASCII(tok.TypeLine, "legendary") {
		t.Errorf("token type line %q dropped the copied card's supertype", tok.TypeLine)
	}
	if !effects.ContainsFoldASCII(tok.TypeLine, "zombie") {
		t.Errorf("token type line %q is not a Zombie", tok.TypeLine)
	}
	if effects.ContainsFoldASCII(tok.TypeLine, "ogre") || effects.ContainsFoldASCII(tok.TypeLine, "warrior") {
		t.Errorf("token type line %q kept the copied creature types; Zombie replaces them", tok.TypeLine)
	}
}

// TestHashatonDeclineCreatesNothing — "if you do" means the token is
// contingent on the payment, so declining leaves the board alone.
func TestHashatonDeclineCreatesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.HashatonDiscardOgre(t, g, me)

	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment prompt to decline")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)

	if effectstest.FindBattlefieldByName(g, "Test Ogre") != uuid.Nil {
		t.Error("declining the cost still made a token")
	}
}

// TestHashatonIgnoresNoncreatureDiscards — the trigger reads "a
// creature card", so a discarded land is not a trigger at all and
// there is no prompt to decline.
func TestHashatonIgnoresNoncreatureDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Hashaton, Scarab's Fist",
		"Legendary Creature — Zombie Wizard", effectstest.HashatonOracle, false)
	effectstest.EmptyHandToLibrary(g, me)
	effectstest.HandCard(me, "Island", "Basic Land — Island")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("discarding a land offered Hashaton's payment")
	}
	if effectstest.FindBattlefieldByName(g, "Island") != uuid.Nil {
		t.Error("discarding a land made a token copy")
	}
}

// TestTokenCopyKeepsCatalogAbilities — the reason a token copy
// carries the copied card's oracle ID. Every ability hook in the
// catalog keys on it, so a copy of Serra Angel really has flying and
// vigilance rather than being a blank 4/4 with her name.
func TestTokenCopyKeepsCatalogAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: src,
		Name:       "Serra Angel",
		TypeLine:   "Creature — Angel",
		OracleID:   effectstest.SerraAngelOracle,
		ManaCost:   "{3}{W}{W}",
		Power:      4,
		Toughness:  4,
		Owner:      me.ID,
		Controller: me.ID,
	})

	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, nil)
		_ = effects.CreateTokenCopy{Controller: me.ID, Copy: src, N: 1}.Apply(ctx)
	})

	tokenID := effectstest.FindBattlefieldByName(g, "Serra Angel")
	if tokenID == uuid.Nil {
		t.Fatal("no token copy was created")
	}
	abilities := effectstest.EffectiveAbilities(t, g, tokenID)
	want := map[string]bool{"flying": false, "vigilance": false}
	for _, a := range abilities {
		if _, ok := want[a]; ok {
			want[a] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("token copy is missing %q; got %v", k, abilities)
		}
	}
	if !me.Graveyard.Contains(src) {
		t.Error("copying a card should not move it")
	}
}

// --- conditional enters-tapped lands -----------------------------

// TestFastlandEntersUntappedEarly — "unless you control two or fewer
// OTHER lands". Two other lands is inside the window.
func TestFastlandEntersUntappedEarly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")

	id := effectstest.PlayLandFromHand(t, g, "Darkslick Shores", effectstest.DarkslickShoresOracle)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("Darkslick Shores is not on the battlefield")
	}
	if card.Tapped {
		t.Error("entered tapped with only two other lands")
	}
}

// TestFastlandEntersTappedLate — three other lands is outside it.
// The tap-event count is the discriminator between a replacement and
// an OnETB tap: a replacement emits none.
func TestFastlandEntersTappedLate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	}

	id := effectstest.PlayLandFromHand(t, g, "Darkslick Shores", effectstest.DarkslickShoresOracle)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("Darkslick Shores is not on the battlefield")
	}
	if !card.Tapped {
		t.Error("entered untapped with three other lands")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped, not been tapped", n)
	}
}

// TestFastlandIgnoresOpponentLands — "you control", not "are on the
// battlefield".
func TestFastlandIgnoresOpponentLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	for i := 0; i < 5; i++ {
		effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	}

	id := effectstest.PlayLandFromHand(t, g, "Darkslick Shores", effectstest.DarkslickShoresOracle)
	card, _ := effectstest.BattlefieldCard(g, id)
	if card.Tapped {
		t.Error("an opponent's lands turned the fastland off")
	}
}

// TestChecklandReadsLandTypes — Glacial Fortress wants a Plains or an
// Island, and a Swamp is neither.
func TestChecklandReadsLandTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")

	id := effectstest.PlayLandFromHand(t, g, "Glacial Fortress", effectstest.GlacialFortressOracle)
	card, _ := effectstest.BattlefieldCard(g, id)
	if !card.Tapped {
		t.Error("entered untapped with no Plains or Island")
	}
}

func TestChecklandEntersUntappedWithTheRightType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")

	id := effectstest.PlayLandFromHand(t, g, "Glacial Fortress", effectstest.GlacialFortressOracle)
	card, _ := effectstest.BattlefieldCard(g, id)
	if card.Tapped {
		t.Error("entered tapped despite a Plains")
	}
}

// TestSlowlandIsTheFastlandInverse — "two or MORE other lands".
func TestSlowlandIsTheFastlandInverse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")

	id := effectstest.PlayLandFromHand(t, g, "Deserted Beach", effectstest.DesertedBeachOracle)
	card, _ := effectstest.BattlefieldCard(g, id)
	if !card.Tapped {
		t.Error("entered untapped with only one other land")
	}
}

func TestSlowlandEntersUntappedLate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")

	id := effectstest.PlayLandFromHand(t, g, "Deserted Beach", effectstest.DesertedBeachOracle)
	card, _ := effectstest.BattlefieldCard(g, id)
	if card.Tapped {
		t.Error("entered tapped with two other lands")
	}
}
