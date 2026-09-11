package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hashaton_test.go — the Hashaton, Scarab's Fist decklist batch: the
// commander's token-copy engine, the reanimation package, and the
// conditional enters-tapped land cycles.
//
// Two assertions in here are the point of the whole batch:
//
//   - Reanimating from an OPPONENT's graveyard has to put the
//     creature under the CASTER's control. It used to route to the
//     graveyard's owner, which handed the creature straight back.
//   - A conditional enters-tapped land has to ENTER tapped, not
//     enter untapped and get tapped. Both leave Tapped == true, so
//     the discriminator is the tap event count, exactly as in
//     temples_test.go.

const (
	hashatonOracle        = "db266661-f783-4907-9e52-6963eec05431"
	reanimateOracle       = "a044474a-cd72-4e9d-bd8d-a08f2de9cdc0"
	zombifyOracle         = "bb95db4d-5017-4121-bf79-d68476602d8c"
	lateToDinnerOracle    = "31bf199b-dfb1-428e-96a5-eb25104e2b43"
	darkslickShoresOracle = "a2b48695-f7d7-42ce-a8a0-2a723428542a"
	glacialFortressOracle = "027dd013-baa7-4111-b3c9-f4d1414e9c45"
	desertedBeachOracle   = "f0ec8681-da50-466b-8cdd-1dc710deccd9"
	testOgreOracle        = "00000000-0000-4000-8000-00000000ffff"
)

// seedGraveyardCreature puts a creature card with a real mana cost
// into a player's graveyard, so a reanimation target exists and its
// mana value is something other than zero.
func seedGraveyardCreature(p *game.Player, name, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Zombie Giant",
		ManaCost:   manaCost,
		Power:      3,
		Toughness:  3,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// seedLandOnBattlefield drops a land straight onto the battlefield
// without going through the play path — for setting up the board
// state a conditional enters-tapped land reads.
func seedLandOnBattlefield(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// emptyHandToLibrary clears a player's hand so a random discard is
// deterministic. The tests below are about what the discard
// triggers, not about which card is picked.
func emptyHandToLibrary(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() {
		for p.Hand.Size() > 0 {
			c, err := p.Hand.Top()
			if err != nil {
				return
			}
			_, _ = game.MoveCard(p.Hand, p.Library, c.InstanceID)
		}
	})
}

// --- reanimation -------------------------------------------------

// TestReanimateTakesAnOpponentsCreature is the controller fix. "Put
// target creature card from A GRAVEYARD onto the battlefield UNDER
// YOUR CONTROL" is the whole reason to play Reanimate in a deck full
// of removal, and it used to hand the creature back to the opponent.
func TestReanimateTakesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	victim := seedGraveyardCreature(opp, "Bulky Zombie", "{4}{B}")
	lifeBefore := me.Life

	castCatalogSpell(t, g, "Reanimate", "Sorcery", reanimateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	card, ok := battlefieldCard(g, victim)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := seedGraveyardCreature(me, "My Zombie", "{4}{B}")
	lifeBefore := me.Life

	castCatalogSpell(t, g, "Zombify", "Sorcery", zombifyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)

	card, ok := battlefieldCard(g, mine)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := seedGraveyardCreature(me, "Dinner Guest", "{2}{W}")

	castCatalogSpell(t, g, "Late to Dinner", "Sorcery", lateToDinnerOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)

	if _, ok := battlefieldCard(g, mine); !ok {
		t.Error("the reanimated creature is not on the battlefield")
	}
	if findBattlefieldByName(g, "Food") == uuid.Nil {
		t.Error("no Food token was created")
	}
}

// TestReanimationFiresTheOnETBHook pins the second of the engine's
// two independently-wired ETB mechanisms. Triggered abilities
// harvest off the emitted EventETB and always worked; Spec.OnETB
// runs through fireETBHookLocked, which this path never called — so
// a reanimated Solemn Simulacrum fetched nothing.
//
// A Temple is the probe because its ETB lives entirely in OnETB (a
// scry) and nowhere else. It also double-checks the controller fix:
// the scry is queued for the Temple's CONTROLLER, so the prompt
// arriving for the reanimating player rather than the graveyard's
// owner is the same assertion twice over.
func TestReanimationFiresTheOnETBHook(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	id := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Temple of Silence",
		TypeLine:   "Land",
		OracleID:   templeOfSilenceOracle,
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

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the Temple is not on the battlefield")
	}
	if card.Controller != me.ID {
		t.Errorf("controller: got %s, want %s", card.Controller, me.ID)
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("the Temple's OnETB did not fire — no scry queued for the new controller")
	}
	if scryChoiceFor(g, opp.ID) != nil {
		t.Error("the scry was queued for the graveyard's owner instead of the new controller")
	}
}

// --- Hashaton ----------------------------------------------------

// hashatonDiscardOgre seeds Hashaton on the battlefield under `me`,
// empties the hand, puts one distinctive creature card in it and
// discards it. Returns the discarded card's ID.
//
// The ogre's printed values are all deliberately unlike the token's:
// 2/5, red, legendary, with subtypes. Every one of them is something
// the "except it's a 4/4 black Zombie" clause has to overwrite or
// keep, so the assertions can tell a real copy from a template.
func hashatonDiscardOgre(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	pushCatalogPermanent(g, me.ID, "Hashaton, Scarab's Fist",
		"Legendary Creature — Zombie Wizard", hashatonOracle, false)
	emptyHandToLibrary(g, me)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Ogre",
		TypeLine:   "Legendary Creature — Ogre Warrior",
		OracleID:   testOgreOracle,
		ManaCost:   "{2}{R}",
		Power:      2,
		Toughness:  5,
		Colors:     []string{"R"},
		Owner:      me.ID,
		Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	return id
}

func TestHashatonCopiesADiscardedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hashatonDiscardOgre(t, g, me)

	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("discarding a creature did not offer the {2}{U} payment")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	answerPayUnless(t, g, me.ID, true)

	tokenID := findBattlefieldByName(g, "Test Ogre")
	if tokenID == uuid.Nil {
		t.Fatal("paying {2}{U} created no token copy")
	}
	tok, _ := battlefieldCard(g, tokenID)
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
	if tok.OracleID != testOgreOracle {
		t.Errorf("token oracle ID: got %q, want the copied card's %q", tok.OracleID, testOgreOracle)
	}
	// "except it's a 4/4 black Zombie" REPLACES the creature types
	// and keeps the supertypes.
	if !containsFoldASCII(tok.TypeLine, "token") {
		t.Errorf("token type line %q is missing the Token supertype", tok.TypeLine)
	}
	if !containsFoldASCII(tok.TypeLine, "legendary") {
		t.Errorf("token type line %q dropped the copied card's supertype", tok.TypeLine)
	}
	if !containsFoldASCII(tok.TypeLine, "zombie") {
		t.Errorf("token type line %q is not a Zombie", tok.TypeLine)
	}
	if containsFoldASCII(tok.TypeLine, "ogre") || containsFoldASCII(tok.TypeLine, "warrior") {
		t.Errorf("token type line %q kept the copied creature types; Zombie replaces them", tok.TypeLine)
	}
}

// TestHashatonDeclineCreatesNothing — "if you do" means the token is
// contingent on the payment, so declining leaves the board alone.
func TestHashatonDeclineCreatesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hashatonDiscardOgre(t, g, me)

	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment prompt to decline")
	}
	answerPayUnless(t, g, me.ID, false)

	if findBattlefieldByName(g, "Test Ogre") != uuid.Nil {
		t.Error("declining the cost still made a token")
	}
}

// TestHashatonIgnoresNoncreatureDiscards — the trigger reads "a
// creature card", so a discarded land is not a trigger at all and
// there is no prompt to decline.
func TestHashatonIgnoresNoncreatureDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Hashaton, Scarab's Fist",
		"Legendary Creature — Zombie Wizard", hashatonOracle, false)
	emptyHandToLibrary(g, me)
	handCard(me, "Island", "Basic Land — Island")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)

	if hasPayUnlessFor(g, me.ID) {
		t.Error("discarding a land offered Hashaton's payment")
	}
	if findBattlefieldByName(g, "Island") != uuid.Nil {
		t.Error("discarding a land made a token copy")
	}
}

// TestTokenCopyKeepsCatalogAbilities — the reason a token copy
// carries the copied card's oracle ID. Every ability hook in the
// catalog keys on it, so a copy of Serra Angel really has flying and
// vigilance rather than being a blank 4/4 with her name.
func TestTokenCopyKeepsCatalogAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: src,
		Name:       "Serra Angel",
		TypeLine:   "Creature — Angel",
		OracleID:   serraAngelOracle,
		ManaCost:   "{3}{W}{W}",
		Power:      4,
		Toughness:  4,
		Owner:      me.ID,
		Controller: me.ID,
	})

	g.WithWriteLock(func() {
		ctx := NewContext(g, nil)
		_ = CreateTokenCopy{Controller: me.ID, Copy: src, N: 1}.Apply(ctx)
	})

	tokenID := findBattlefieldByName(g, "Serra Angel")
	if tokenID == uuid.Nil {
		t.Fatal("no token copy was created")
	}
	abilities := effectiveAbilities(t, g, tokenID)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")

	id := playLandFromHand(t, g, "Darkslick Shores", darkslickShoresOracle)
	card, ok := battlefieldCard(g, id)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	}

	id := playLandFromHand(t, g, "Darkslick Shores", darkslickShoresOracle)
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Darkslick Shores is not on the battlefield")
	}
	if !card.Tapped {
		t.Error("entered untapped with three other lands")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped, not been tapped", n)
	}
}

// TestFastlandIgnoresOpponentLands — "you control", not "are on the
// battlefield".
func TestFastlandIgnoresOpponentLands(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	for i := 0; i < 5; i++ {
		seedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	}

	id := playLandFromHand(t, g, "Darkslick Shores", darkslickShoresOracle)
	card, _ := battlefieldCard(g, id)
	if card.Tapped {
		t.Error("an opponent's lands turned the fastland off")
	}
}

// TestChecklandReadsLandTypes — Glacial Fortress wants a Plains or an
// Island, and a Swamp is neither.
func TestChecklandReadsLandTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")

	id := playLandFromHand(t, g, "Glacial Fortress", glacialFortressOracle)
	card, _ := battlefieldCard(g, id)
	if !card.Tapped {
		t.Error("entered untapped with no Plains or Island")
	}
}

func TestChecklandEntersUntappedWithTheRightType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")

	id := playLandFromHand(t, g, "Glacial Fortress", glacialFortressOracle)
	card, _ := battlefieldCard(g, id)
	if card.Tapped {
		t.Error("entered tapped despite a Plains")
	}
}

// TestSlowlandIsTheFastlandInverse — "two or MORE other lands".
func TestSlowlandIsTheFastlandInverse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")

	id := playLandFromHand(t, g, "Deserted Beach", desertedBeachOracle)
	card, _ := battlefieldCard(g, id)
	if !card.Tapped {
		t.Error("entered untapped with only one other land")
	}
}

func TestSlowlandEntersUntappedLate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")

	id := playLandFromHand(t, g, "Deserted Beach", desertedBeachOracle)
	card, _ := battlefieldCard(g, id)
	if card.Tapped {
		t.Error("entered tapped with two other lands")
	}
}
