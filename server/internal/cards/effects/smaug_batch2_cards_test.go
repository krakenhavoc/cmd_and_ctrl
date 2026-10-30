package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// smaug_batch2_cards_test.go — the Smaug Unchained deck's Treasure cards
// (#4724, batch 2): Decadent Dragon // Expensive Taste, Kellogg,
// Dangerous Mind, Knuckles the Echidna and Mastermind Plum.

const (
	kelloggOracle         = "904b2603-00c3-44d2-a43c-c25439dd8038"
	knucklesOracle        = "71ae0fab-eb3e-4226-9fcc-29186f98a569"
	mastermindPlumOracle  = "4605d8b3-b714-4717-8ff3-155fabfc1999"
	smaugB2ColorlessCost  = "{2}{C}"
	smaugB2CheapSpellCost = "{1}{B}"
)

// --- Decadent Dragon // Expensive Taste --------------------------------

// decadentDragon builds the printed card the way deck.toGameCard does:
// per-face data, then SetFace(0).
func decadentDragon(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   decadentDragonOracleID,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:      "Decadent Dragon",
				TypeLine:  "Creature — Dragon",
				ManaCost:  "{2}{R}{R}",
				Colors:    []string{"R"},
				Power:     4,
				Toughness: 4,
			},
			{
				Name:     "Expensive Taste",
				TypeLine: "Instant — Adventure",
				ManaCost: "{2}{B}",
				Colors:   []string{"B"},
			},
		},
	}
	c.SetFace(0)
	return c
}

// castExpensiveTaste puts the card in the active player's hand at a main
// phase and casts its Adventure half at `target`.
func castExpensiveTaste(t *testing.T, g *game.Game, target uuid.UUID) (*game.Player, uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c := decadentDragon(me.ID)
	me.Hand.PushTop(c)
	advanceToMain(t, g)
	err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{
		Face:    1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}},
	})
	return me, c.InstanceID, err
}

// The creature attacks, has its two keywords, and makes a Treasure.
func TestDecadentDragonAttackMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := decadentDragon(me.ID)
	g.Battlefield.PushTop(dragon)

	if !effectiveAbilitiesContain(t, g, dragon.InstanceID, "flying") ||
		!effectiveAbilitiesContain(t, g, dragon.InstanceID, "trample") {
		t.Error("Decadent Dragon has flying and trample")
	}
	declareAttack(t, g, opp.ID, dragon.InstanceID)
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 1 {
		t.Errorf("Treasures after the attack = %d, want 1", got)
	}
}

// Expensive Taste exiles the top two cards of the target opponent's
// library face down, for the caster alone to look at and play for as long
// as they stay exiled. No "mana of any type" rider: that is Outrageous
// Robbery's, not this card's.
func TestExpensiveTasteExilesTheTopTwoFaceDownForYouToPlay(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	n := len(opp.Library.Cards)
	top := []uuid.UUID{opp.Library.Cards[n-1].InstanceID, opp.Library.Cards[n-2].InstanceID}
	third := opp.Library.Cards[n-3].InstanceID

	me, card, err := castExpensiveTaste(t, g, opp.ID)
	if err != nil {
		t.Fatalf("cast Expensive Taste: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := n - len(opp.Library.Cards); got != 2 {
		t.Fatalf("exiled %d cards, want 2", got)
	}
	if !opp.Library.Contains(third) {
		t.Error("the third card of the library was exiled")
	}
	for _, id := range top {
		perm := e2Permission(g, id)
		if perm == nil || perm.Player != me.ID {
			t.Fatalf("card %s permission %+v, want a grant to the caster", id, perm)
		}
		if perm.CastOnly {
			t.Errorf("card %s: a cast-only grant, but the card says play (lands too)", id)
		}
		if perm.AnyType || perm.AnyColor {
			t.Errorf("card %s: mana of any type or colour, which this card does not print", id)
		}
		if perm.Duration.Kind != game.WhileInZone {
			t.Errorf("card %s: duration %v, want for as long as it remains exiled", id, perm.Duration.Kind)
		}
		e2AssertFaceDownFor(t, g, id, me.ID)
	}
	// CR 715.3d: the resolved Adventure is exiled, and the creature may be
	// cast from there.
	if !g.Exile.Contains(card) {
		t.Fatal("the resolved Adventure spell is not in exile")
	}
	perm := g.CastPermissionOnCardByIDForEffect(card)
	if face, ok := perm.NamedFace(); !ok || face != 0 {
		t.Errorf("adventure grant faces = %v, want just the creature face [0]", perm.Faces)
	}
}

// The grant is the printed one: a stolen {2}{C} card cannot be paid with
// three Swamps, because the card does not let mana be spent as though it
// were any type.
func TestExpensiveTasteGrantsNoAnyTypeMana(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seer := game.NewCard("Their Seer", opp.ID)
	seer.TypeLine = "Creature — Eldrazi"
	seer.ManaCost = smaugB2ColorlessCost
	opp.Library.PushTop(seer)

	me, _, err := castExpensiveTaste(t, g, opp.ID)
	if err != nil {
		t.Fatalf("cast Expensive Taste: %v", err)
	}
	passPriorityAroundTable(t, g)
	e2AssertFaceDownFor(t, g, seer.InstanceID, me.ID)
	e2Swamps(g, me.ID, 3)
	if err := g.CastSpell(me.ID, seer.InstanceID, game.CastSpellParams{FromZone: "exile", AutoTap: true, Strict: true}); err == nil {
		t.Fatal("a {2}{C} card was cast off three Swamps: the grant spends mana as though it were any type")
	}
}

// A stolen card the caster can pay for is castable from exile, face up
// once it is cast.
func TestExpensiveTasteCardIsCastableFromExile(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := game.NewCard("Their Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	bear.ManaCost = smaugB2CheapSpellCost
	opp.Library.PushTop(bear)

	me, _, err := castExpensiveTaste(t, g, opp.ID)
	if err != nil {
		t.Fatalf("cast Expensive Taste: %v", err)
	}
	passPriorityAroundTable(t, g)
	e2Swamps(g, me.ID, 2)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{FromZone: "exile", AutoTap: true}); err != nil {
		t.Fatalf("cast the stolen card from exile: %v", err)
	}
	if c := e2Card(t, g, bear.InstanceID); c.FaceDown {
		t.Error("the cast card is still face down")
	}
}

// "Target opponent": the caster's own library is not a legal target.
func TestExpensiveTasteTargetsOnlyAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if _, _, err := castExpensiveTaste(t, g, me.ID); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Expensive Taste at its caster: got %v, want ErrIllegalTarget", err)
	}
}

// --- Kellogg, Dangerous Mind -------------------------------------------

func pushKellogg(g *game.Game, controller uuid.UUID) uuid.UUID {
	return b11Push(g, controller, "Kellogg, Dangerous Mind", "Legendary Creature — Human Mercenary", kelloggOracle, 3, 2)
}

func TestKelloggAttackMakesATreasureAndHasHisKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kellogg := pushKellogg(g, me.ID)
	for _, kw := range []string{"first strike", "haste"} {
		if !effectiveAbilitiesContain(t, g, kellogg, kw) {
			t.Errorf("Kellogg lacks %s", kw)
		}
	}
	declareAttack(t, g, opp.ID, kellogg)
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 1 {
		t.Errorf("Treasures after the attack = %d, want 1", got)
	}
}

// Five Treasures take a creature for as long as you control Kellogg; four
// are refused; and losing Kellogg gives the creature back.
func TestKelloggStealsForAsLongAsYouControlHim(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kellogg := pushKellogg(g, me.ID)
	victim := ctrlPushCreature(g, opp.ID, "Their Bear")
	treasures := pushTokens(g, me.ID, TreasureToken(), 5)
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, kellogg, 0, game.ActivateAbilityParams{
		SacrificeIDs: treasures[:4], Targets: ltCardTarget(victim),
	}); err != game.ErrInvalidParam {
		t.Fatalf("four Treasures: got %v, want ErrInvalidParam", err)
	}
	b16Activate(t, g, me.ID, kellogg, 0, game.ActivateAbilityParams{
		SacrificeIDs: treasures, Targets: ltCardTarget(victim),
	})
	if countTreasures(g, me.ID) != 0 {
		t.Error("all five Treasures are sacrificed")
	}
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s, want Kellogg's controller", got)
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(kellogg) })
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Error("Kellogg left the battlefield and the creature stayed stolen")
	}
}

// "For as long as YOU control Kellogg": an opponent taking Kellogg ends
// the steal even though he stays on the battlefield.
func TestKelloggStealEndsWhenKelloggChangesHands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kellogg := pushKellogg(g, me.ID)
	victim := ctrlPushCreature(g, opp.ID, "Their Bear")
	treasures := pushTokens(g, me.ID, TreasureToken(), 5)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, kellogg, 0, game.ActivateAbilityParams{
		SacrificeIDs: treasures, Targets: ltCardTarget(victim),
	})
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s, want Kellogg's controller", got)
	}

	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), kellogg, opp.ID, game.IndefiniteDuration(), "test — take Kellogg")
	})
	if got := controllerOf(t, g, kellogg); got != opp.ID {
		t.Fatalf("setup: Kellogg's controller %s, want the opponent", got)
	}
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Error("Kellogg changed hands and the creature stayed stolen")
	}
}

// "Activate only as a sorcery": not in combat, not with something on the
// stack.
func TestKelloggStealIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kellogg := pushKellogg(g, me.ID)
	victim := ctrlPushCreature(g, opp.ID, "Their Bear")
	treasures := pushTokens(g, me.ID, TreasureToken(), 5)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, kellogg, 0, game.ActivateAbilityParams{
		SacrificeIDs: treasures, Targets: ltCardTarget(victim),
	}); err == nil {
		t.Fatal("Kellogg's steal was activated outside a main phase")
	}
	if countTreasures(g, me.ID) != 5 {
		t.Error("a refused activation sacrificed Treasures")
	}
}

// --- Knuckles the Echidna ----------------------------------------------

func pushKnuckles(g *game.Game, controller uuid.UUID) uuid.UUID {
	return b11Push(g, controller, "Knuckles the Echidna", "Legendary Creature — Echidna Warrior", knucklesOracle, 2, 4)
}

// Double strike is two combat damage steps (CR 510.4), each its own batch:
// a Treasure in each. Two creatures connecting with one player in the same
// step are one batch and one Treasure (CR 603.2c).
func TestKnucklesMakesATreasurePerDamageStepNotPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knuckles := pushKnuckles(g, me.ID)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	for _, kw := range []string{"double strike", "trample", "haste"} {
		if !effectiveAbilitiesContain(t, g, knuckles, kw) {
			t.Errorf("Knuckles lacks %s", kw)
		}
	}

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{knuckles, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepFirstStrikeDamage)
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 1 {
		t.Fatalf("Treasures after the first-strike step = %d, want 1", got)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	// The regular step: Knuckles again and the Bear, one player, one batch.
	if got := countTreasures(g, me.ID); got != 2 {
		t.Errorf("Treasures after both damage steps = %d, want 2", got)
	}
	if opp.Life != 40-2-2-2 {
		t.Errorf("opponent life %d, want 34", opp.Life)
	}
}

// Two players hit in one damage step are two triggers ("a player", per
// player, CR 603.2c).
func TestKnucklesMakesATreasurePerPlayerHit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	pushKnuckles(g, me.ID)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, other.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := countTreasures(g, me.ID); got != 2 {
		t.Errorf("Treasures = %d, want one per player hit (2)", got)
	}
}

func knucklesWithArtifacts(t *testing.T, n int) (*game.Game, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[1]
	knuckles := pushKnuckles(g, me.ID)
	rocks := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		rocks = append(rocks, s58Rock(g, me.ID, "Rock", "Artifact", "{1}"))
	}
	advanceToUpkeepOf(t, g, 1)
	return g, knuckles, rocks
}

func TestKnucklesWinsAtUpkeepWithThirtyArtifacts(t *testing.T) {
	g, _, _ := knucklesWithArtifacts(t, 30)
	passPriorityAroundTable(t, g)
	if g.State != game.StateEnded || g.Outcome == nil || g.Outcome.Winner != g.Seats[1].ID {
		t.Fatalf("state %s outcome %+v, want seat 1 to win", g.State, g.Outcome)
	}
}

// The intervening if as the ability triggers: twenty-nine artifacts put
// nothing on the stack.
func TestKnucklesDoesNotTriggerWithTwentyNineArtifacts(t *testing.T) {
	g, knuckles, _ := knucklesWithArtifacts(t, 29)
	if n := triggersOnStackFrom(g, knuckles); n != 0 {
		t.Fatalf("Treasure Hunter triggered with 29 artifacts (%d on the stack)", n)
	}
	passPriorityAroundTable(t, g)
	if g.State != game.StateActive {
		t.Fatalf("29 artifacts ended the game: %+v", g.Outcome)
	}
}

// And as it resolves (CR 603.4): an artifact lost in response stops the
// win.
func TestKnucklesWinIsCheckedAgainOnResolution(t *testing.T) {
	g, knuckles, rocks := knucklesWithArtifacts(t, 30)
	if n := triggersOnStackFrom(g, knuckles); n != 1 {
		t.Fatalf("Treasure Hunter on the stack = %d, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rocks[0]) })
	passPriorityAroundTable(t, g)
	if g.State != game.StateActive {
		t.Fatalf("the win resolved with 29 artifacts: %+v", g.Outcome)
	}
}

// "Your upkeep": an opponent's upkeep does nothing.
func TestKnucklesDoesNotWinOnAnOpponentsUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[2]
	knuckles := pushKnuckles(g, me.ID)
	for i := 0; i < 30; i++ {
		s58Rock(g, me.ID, "Rock", "Artifact", "{1}")
	}
	advanceToUpkeepOf(t, g, 1)
	if n := triggersOnStackFrom(g, knuckles); n != 0 {
		t.Fatalf("Treasure Hunter triggered on an opponent's upkeep (%d)", n)
	}
	passPriorityAroundTable(t, g)
	if g.State != game.StateActive {
		t.Fatalf("Knuckles won on an opponent's upkeep: %+v", g.Outcome)
	}
}

// --- Mastermind Plum ---------------------------------------------------

func pushPlum(g *game.Game, controller uuid.UUID) uuid.UUID {
	return b11Push(g, controller, "Mastermind Plum", "Legendary Creature — Human Wizard", mastermindPlumOracle, 2, 2)
}

// plumAttack attacks with Plum and answers the trigger's target prompt
// with `pick` (none for uuid.Nil), then resolves it.
func plumAttack(t *testing.T, g *game.Game, me, opp *game.Player, plum, pick uuid.UUID) {
	t.Helper()
	declareAttack(t, g, opp.ID, plum)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("Plum's attack trigger asked for no target")
	}
	var refs []game.TargetRef
	if pick != uuid.Nil {
		refs = []game.TargetRef{{Kind: game.TargetCard, ID: pick}}
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// An artifact card from any graveyard is exiled and makes a Treasure.
func TestMastermindPlumExilesAnArtifactCardForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plum := pushPlum(g, me.ID)
	relic := pushGraveyardPermanent(opp, "Their Relic", "Artifact", "{2}")

	plumAttack(t, g, me, opp, plum, relic)
	if !g.Exile.Contains(relic) {
		t.Fatal("the artifact card was not exiled")
	}
	if got := countTreasures(g, me.ID); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
}

// A card that is not an artifact is exiled and makes nothing.
func TestMastermindPlumNonArtifactCardMakesNoTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plum := pushPlum(g, me.ID)
	bear := pushGraveyardCardForTest(opp, "Their Bear")

	plumAttack(t, g, me, opp, plum, bear)
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature card was not exiled")
	}
	if got := countTreasures(g, me.ID); got != 0 {
		t.Errorf("Treasures = %d, want 0", got)
	}
}

// "Up to one": choosing no target is legal and does nothing.
func TestMastermindPlumMayChooseNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plum := pushPlum(g, me.ID)
	relic := pushGraveyardPermanent(opp, "Their Relic", "Artifact", "{2}")

	plumAttack(t, g, me, opp, plum, uuid.Nil)
	if g.Exile.Contains(relic) {
		t.Error("a card was exiled with no target chosen")
	}
	if got := countTreasures(g, me.ID); got != 0 {
		t.Errorf("Treasures = %d, want 0", got)
	}
}

// The cast trigger: a spell paid partly with Treasure mana draws a card
// and loses a life; one paid with no Treasure mana puts nothing on the
// stack (the intervening if, CR 603.4).
func TestMastermindPlumDrawsOffTreasureMana(t *testing.T) {
	for _, tc := range []struct {
		name     string
		treasure bool
		draws    int
	}{
		{"Treasure mana", true, 1},
		{"no Treasure mana", false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			plum := pushPlum(g, me.ID)
			advanceToMain(t, g)
			if tc.treasure {
				crackATreasureFor(t, g, me, "B")
				floatForTest(g, me, "B")
			} else {
				floatForTest(g, me, "BB")
			}
			hand, life := me.Hand.Size(), me.Life

			castFromHandForTest(t, g, me, "Bear Cub", "Creature — Bear", smaugB2CheapSpellCost, "", game.CastSpellParams{Strict: true})
			if got, want := triggersOnStackFrom(g, plum), tc.draws; got != want {
				t.Errorf("Plum triggers on the stack = %d, want %d", got, want)
			}
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("drew %d, want %d", got, tc.draws)
			}
			if got := life - me.Life; got != tc.draws {
				t.Errorf("lost %d life, want %d", got, tc.draws)
			}
		})
	}
}

// "Whenever YOU cast": an opponent's Treasure-paid spell does nothing.
func TestMastermindPlumIgnoresAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plum := pushPlum(g, opp.ID)
	advanceToMain(t, g)
	crackATreasureFor(t, g, me, "B")
	floatForTest(g, me, "B")
	hand := opp.Hand.Size()

	castFromHandForTest(t, g, me, "Bear Cub", "Creature — Bear", smaugB2CheapSpellCost, "", game.CastSpellParams{Strict: true})
	if n := triggersOnStackFrom(g, plum); n != 0 {
		t.Fatalf("Plum triggered on an opponent's spell (%d)", n)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 0 {
		t.Errorf("Plum's controller drew %d, want 0", got)
	}
}
