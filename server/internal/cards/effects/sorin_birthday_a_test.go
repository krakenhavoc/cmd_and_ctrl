package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sorin_birthday_a_test.go — #4077 slice a: the four Sorin planeswalkers
// of the Sorin's Birthday Party deck request, driven through the real
// loyalty path with the real catalog wired.

const (
	sorinMirthlessOracle   = "fbffaf9f-40e7-4113-a037-533b733cebce"
	sorinGrimNemesisOracle = "0a01dd05-e289-489f-b3ad-88e15e157bd0"
	sorinLordOracle        = "5c780dda-ee76-482a-a610-53c321c4fdec"
	sorinSolemnOracle      = "1f1f1ae7-dd00-4825-832e-3a049cf8ce73"
)

// sorinATable is a fresh game at the active seat's precombat main phase
// with `name` on the battlefield at `loyalty`.
func sorinATable(t *testing.T, name, oracle string, loyalty int) (g *game.Game, seat int, me *game.Player, sorin uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	seat = g.Turn.ActiveSeat
	me = g.Seats[seat]
	sorin = pushWalkerForTest(g, me.ID, name, oracle, loyalty)
	advanceToMainOf(t, g, seat)
	return g, seat, me, sorin
}

func sorinAActivate(t *testing.T, g *game.Game, me *game.Player, sorin uuid.UUID, idx int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, sorin, idx, params); err != nil {
		t.Fatalf("ability %d: %v", idx, err)
	}
}

func sorinInHand(p *game.Player, id uuid.UUID) bool { return p.Hand.Contains(id) }

// --- Sorin the Mirthless ----------------------------------------------

// The +1 taken: the card goes to hand and its mana value is paid in life.
func TestSorinMirthlessPlusOneTakesTheCardAndPaysItsManaValue(t *testing.T) {
	g, _, me, sorin := sorinATable(t, "Sorin the Mirthless", sorinMirthlessOracle, 4)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Three Drop", TypeLine: "Creature — Test", ManaCost: "{1}{1}{1}", Owner: me.ID, Controller: me.ID})
	life := me.Life

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, top)

	if !sorinInHand(me, top) {
		t.Fatal("the revealed card did not reach the hand")
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("life lost = %d, want the card's mana value 3", got)
	}
	if got := loyaltyOf(g, sorin); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
}

// Declining the "you may": the card stays on top and no life is paid.
func TestSorinMirthlessPlusOneDeclinedPaysNothing(t *testing.T) {
	g, _, me, sorin := sorinATable(t, "Sorin the Mirthless", sorinMirthlessOracle, 4)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Three Drop", TypeLine: "Creature — Test", ManaCost: "{1}{1}{1}", Owner: me.ID, Controller: me.ID})
	life := me.Life

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)

	if sorinInHand(me, top) {
		t.Error("a declined look put the card in hand")
	}
	if me.Life != life {
		t.Errorf("life changed on a declined look: %d → %d", life, me.Life)
	}
}

// A land has mana value 0: taking it costs nothing, and it is not a draw.
func TestSorinMirthlessPlusOneOnALandCostsNothing(t *testing.T) {
	g, _, me, sorin := sorinATable(t, "Sorin the Mirthless", sorinMirthlessOracle, 4)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	life := me.Life

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, top)

	if !sorinInHand(me, top) {
		t.Fatal("the land did not reach the hand")
	}
	if me.Life != life {
		t.Errorf("a land cost %d life", life-me.Life)
	}
}

// The −2 token is the printed 2/3 black Vampire with flying and lifelink.
func TestSorinMirthlessMinusTwoMakesAFlyingLifelinkVampire(t *testing.T) {
	g, _, me, sorin := sorinATable(t, "Sorin the Mirthless", sorinMirthlessOracle, 4)

	sorinAActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	var tok *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Vampire" {
			tok = &g.Battlefield.Cards[i]
		}
	}
	if tok == nil {
		t.Fatal("no Vampire token")
	}
	if tok.Controller != me.ID || !tok.IsToken() {
		t.Error("the Vampire is not a token under Sorin's controller")
	}
	if p, tough := effectivePTOf(t, g, tok.InstanceID); p != 2 || tough != 3 {
		t.Errorf("token is %d/%d, want 2/3", p, tough)
	}
	for _, kw := range []string{"flying", "lifelink"} {
		if !hasKeywordOf(g, tok.InstanceID, kw) {
			t.Errorf("token lacks %s", kw)
		}
	}
	if !tok.HasColor("B") {
		t.Error("token is not black")
	}
}

// The −7: 13 damage to any target and 13 life.
func TestSorinMirthlessMinusSevenDealsThirteenAndGainsThirteen(t *testing.T) {
	g, _, me, sorin := sorinATable(t, "Sorin the Mirthless", sorinMirthlessOracle, 7)
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine, theirs := me.Life, them.Life

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	})
	passPriorityAroundTable(t, g)

	if got := theirs - them.Life; got != 13 {
		t.Errorf("damage = %d, want 13", got)
	}
	if got := me.Life - mine; got != 13 {
		t.Errorf("life gained = %d, want 13", got)
	}
}

// --- Sorin, Grim Nemesis ----------------------------------------------

// The +1: the card goes to hand, every opponent (and not the controller)
// loses life equal to its mana value.
func TestSorinGrimNemesisPlusOneDrainsEachOpponentForTheCard(t *testing.T) {
	g, _, me, sorin := sorinATable(t, "Sorin, Grim Nemesis", sorinGrimNemesisOracle, 6)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Four Drop", TypeLine: "Creature — Test", ManaCost: "{1}{1}{1}{1}", Owner: me.ID, Controller: me.ID})
	lives := b17Life(g)

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	if !sorinInHand(me, top) {
		t.Fatal("the revealed card did not reach the hand")
	}
	for i, p := range g.Seats {
		want := lives[i] - 4
		if p.ID == me.ID {
			want = lives[i]
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
	if got := loyaltyOf(g, sorin); got != 7 {
		t.Errorf("loyalty = %d, want 7", got)
	}
}

// The −X: X damage to the creature, X life to the controller, X loyalty
// paid.
func TestSorinGrimNemesisMinusXDamagesACreatureAndGainsLife(t *testing.T) {
	g, seat, me, sorin := sorinATable(t, "Sorin, Grim Nemesis", sorinGrimNemesisOracle, 6)
	them := g.Seats[(seat+1)%len(g.Seats)]
	ogre := pushSizedCreatureForTest(g, them.ID, "Ogre", "Creature — Ogre", "{2}{R}", []string{"R"}, 3, 3)
	life := me.Life

	sorinAActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{
		XValue:  3,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ogre}},
	})
	if got := loyaltyOf(g, sorin); got != 3 {
		t.Errorf("loyalty after −3 = %d, want 3", got)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(ogre) {
		t.Error("a 3/3 survived 3 damage")
	}
	if got := me.Life - life; got != 3 {
		t.Errorf("life gained = %d, want 3", got)
	}
}

// A planeswalker is a legal target, and the damage removes loyalty.
func TestSorinGrimNemesisMinusXHitsAPlaneswalker(t *testing.T) {
	g, seat, me, sorin := sorinATable(t, "Sorin, Grim Nemesis", sorinGrimNemesisOracle, 6)
	them := g.Seats[(seat+1)%len(g.Seats)]
	walker := pushWalkerForTest(g, them.ID, "Some Walker", "00000000-0000-0000-0000-000000000abc", 5)

	sorinAActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: walker}},
	})
	passPriorityAroundTable(t, g)

	if got := loyaltyOf(g, walker); got != 3 {
		t.Errorf("walker loyalty = %d, want 3 after 2 damage", got)
	}
}

// A player is not a legal target: "target creature or planeswalker".
func TestSorinGrimNemesisMinusXRefusesAPlayer(t *testing.T) {
	g, seat, me, sorin := sorinATable(t, "Sorin, Grim Nemesis", sorinGrimNemesisOracle, 6)
	them := g.Seats[(seat+1)%len(g.Seats)]

	err := g.ActivateCatalogAbility(me.ID, sorin, 1, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a player: got %v, want ErrIllegalTarget", err)
	}
}

// The −9: as many 1/1 lifelink Vampire Knights as the highest life total
// at the table, the controller's own included.
func TestSorinGrimNemesisMinusNineMakesTokensEqualToTheHighestLife(t *testing.T) {
	g, seat, me, sorin := sorinATable(t, "Sorin, Grim Nemesis", sorinGrimNemesisOracle, 9)
	// The highest total belongs to the LAST seat, not the controller: the
	// count is "among all players".
	for i, p := range g.Seats {
		p.Life = 3 + i
	}
	g.Seats[len(g.Seats)-1].Life = 17
	me.Life = 4
	_ = seat

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	if got := countNamed(g, "Vampire Knight"); got != 17 {
		t.Fatalf("Vampire Knights = %d, want the highest life total 17", got)
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name != "Vampire Knight" {
			continue
		}
		if c.Controller != me.ID {
			t.Fatal("a token is not under Sorin's controller")
		}
		if !hasKeywordOf(g, c.InstanceID, "lifelink") || !c.HasColor("B") {
			t.Fatal("a token is not a black lifelink creature")
		}
		break
	}
}

// --- Sorin, Lord of Innistrad -----------------------------------------

func TestSorinLordPlusOneMakesALifelinkVampire(t *testing.T) {
	g, _, me, sorin := sorinATable(t, "Sorin, Lord of Innistrad", sorinLordOracle, 3)

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	var tok *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Vampire" {
			tok = &g.Battlefield.Cards[i]
		}
	}
	if tok == nil {
		t.Fatal("no Vampire token")
	}
	if p, tough := effectivePTOf(t, g, tok.InstanceID); p != 1 || tough != 1 {
		t.Errorf("token is %d/%d, want 1/1", p, tough)
	}
	if !hasKeywordOf(g, tok.InstanceID, "lifelink") {
		t.Error("token lacks lifelink")
	}
}

// The −2 emblem: +1/+0 to every creature its owner controls, including
// one that arrives later, and it outlives Sorin.
func TestSorinLordMinusTwoEmblemPumpsYourCreaturesForever(t *testing.T) {
	g, seat, me, sorin := sorinATable(t, "Sorin, Lord of Innistrad", sorinLordOracle, 3)
	them := g.Seats[(seat+1)%len(g.Seats)]
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	theirs := pushCreatureToBattlefieldForTest(g, them.ID, "Theirs")

	sorinAActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	if got := me.Emblems.Size(); got != 1 {
		t.Fatalf("emblems = %d, want 1", got)
	}
	if p, tough := effectivePTOf(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("my creature is %d/%d, want 3/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want an unchanged 2/2", p, tough)
	}

	later := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Later", TypeLine: "Creature — Test", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	if p, _ := effectivePTOf(t, g, later); p != 2 {
		t.Errorf("a creature that arrived later has power %d, want 2", p)
	}

	g.WithWriteLock(func() { g.Battlefield.Remove(sorin) })
	if p, _ := effectivePTOf(t, g, mine); p != 3 {
		t.Errorf("after Sorin left, power %d, want the emblem's 3", p)
	}
}

// The −6: destroy up to three, return what reached a graveyard under the
// activator's control. A token is gone for good; an indestructible
// permanent is never destroyed and so never returned.
func TestSorinLordMinusSixDestroysThenReturnsThemUnderYourControl(t *testing.T) {
	g, seat, me, sorin := sorinATable(t, "Sorin, Lord of Innistrad", sorinLordOracle, 6)
	them := g.Seats[(seat+1)%len(g.Seats)]
	bear := pushCreatureToBattlefieldForTest(g, them.ID, "Bear")
	walker := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: walker, Name: "Some Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 4, Owner: them.ID, Controller: them.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	tokenID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: tokenID, Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1,
		Owner: them.ID, Controller: them.ID,
	})
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: bear},
		{Kind: game.TargetCard, ID: walker},
		{Kind: game.TargetCard, ID: tokenID},
	}

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{Targets: targets})
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{bear, walker} {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatalf("card %s did not come back to the battlefield", id)
		}
		if c.Controller != me.ID {
			t.Errorf("%s returned under %s, want the activator", c.Name, c.Controller)
		}
		if c.Owner != them.ID {
			t.Errorf("%s changed owner", c.Name)
		}
	}
	if g.Battlefield.Contains(tokenID) {
		t.Error("the destroyed token is still on the battlefield")
	}
	for _, p := range g.Seats {
		if p.Graveyard.Contains(tokenID) {
			t.Error("a token is in a graveyard")
		}
	}
	if got := loyaltyOf(g, walker); got != 4 {
		t.Errorf("the returned planeswalker has loyalty %d, want its printed 4 (a new object, CR 306.5b)", got)
	}
}

func TestSorinLordMinusSixDoesNotReturnWhatWasNotDestroyed(t *testing.T) {
	g, seat, me, sorin := sorinATable(t, "Sorin, Lord of Innistrad", sorinLordOracle, 6)
	them := g.Seats[(seat+1)%len(g.Seats)]
	god := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: god, Name: "Immortal", TypeLine: "Creature — God", Power: 4, Toughness: 4,
		Keywords: []string{"indestructible"}, Owner: them.ID, Controller: them.ID,
	})

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: god}},
	})
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, god)
	if !ok {
		t.Fatal("the indestructible creature left the battlefield")
	}
	if c.Controller != them.ID {
		t.Error("an indestructible creature changed controller: it was never destroyed, so nothing is returned")
	}
}

// "Other planeswalkers": Sorin himself is not a legal pick, and zero
// targets is a legal activation ("up to three").
func TestSorinLordMinusSixCannotTargetSorinHimself(t *testing.T) {
	g, _, me, sorin := sorinATable(t, "Sorin, Lord of Innistrad", sorinLordOracle, 6)

	err := g.ActivateCatalogAbility(me.ID, sorin, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: sorin}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting Sorin himself: got %v, want ErrIllegalTarget", err)
	}
}

func TestSorinLordMinusSixWithNoTargetsIsLegal(t *testing.T) {
	g, _, me, sorin := sorinATable(t, "Sorin, Lord of Innistrad", sorinLordOracle, 6)

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
}

// --- Sorin, Solemn Visitor --------------------------------------------

// The +1: +1/+0 and lifelink on the creatures you control as it
// resolves, until your next turn begins; a creature that arrives later
// gets neither.
func TestSorinSolemnPlusOneLastsUntilYourNextTurnAndLocksTheSet(t *testing.T) {
	g, seat, me, sorin := sorinATable(t, "Sorin, Solemn Visitor", sorinSolemnOracle, 4)
	them := g.Seats[(seat+1)%len(g.Seats)]
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	theirs := pushCreatureToBattlefieldForTest(g, them.ID, "Theirs")

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	if p, _ := effectivePTOf(t, g, mine); p != 3 {
		t.Errorf("my creature has power %d, want 3", p)
	}
	if !hasKeywordOf(g, mine, "lifelink") {
		t.Error("my creature lacks lifelink")
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 || hasKeywordOf(g, theirs, "lifelink") {
		t.Error("an opponent's creature was affected")
	}
	later := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Later", TypeLine: "Creature — Test", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	if p, _ := effectivePTOf(t, g, later); p != 1 || hasKeywordOf(g, later, "lifelink") {
		t.Error("a creature that arrived after resolution was affected (CR 611.2c)")
	}

	// Still on during every other player's turn…
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if p, _ := effectivePTOf(t, g, mine); p != 3 || !hasKeywordOf(g, mine, "lifelink") {
		t.Error("the effect ended before my next turn")
	}
	// …and gone once my next turn begins.
	advanceToMainOf(t, g, seat)
	if p, _ := effectivePTOf(t, g, mine); p != 2 || hasKeywordOf(g, mine, "lifelink") {
		t.Error("the effect outlasted the start of my next turn")
	}
}

func TestSorinSolemnMinusTwoMakesAFlyingVampire(t *testing.T) {
	g, _, me, sorin := sorinATable(t, "Sorin, Solemn Visitor", sorinSolemnOracle, 4)

	sorinAActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	var tok *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Vampire" {
			tok = &g.Battlefield.Cards[i]
		}
	}
	if tok == nil {
		t.Fatal("no Vampire token")
	}
	if p, tough := effectivePTOf(t, g, tok.InstanceID); p != 2 || tough != 2 {
		t.Errorf("token is %d/%d, want 2/2", p, tough)
	}
	if !hasKeywordOf(g, tok.InstanceID, "flying") || !tok.HasColor("B") {
		t.Error("token is not a black flyer")
	}
}

// The −6 emblem: at the beginning of each opponent's upkeep, that
// player sacrifices a creature of their choice. Not on its owner's own
// upkeep.
func TestSorinSolemnEmblemMakesEachOpponentSacrificeInTheirUpkeep(t *testing.T) {
	g, seat, me, sorin := sorinATable(t, "Sorin, Solemn Visitor", sorinSolemnOracle, 6)
	next := (seat + 1) % len(g.Seats)
	opp := g.Seats[next]
	mineA := pushCreatureToBattlefieldForTest(g, me.ID, "MineA")
	oppA := pushCreatureToBattlefieldForTest(g, opp.ID, "OppA")
	oppB := pushCreatureToBattlefieldForTest(g, opp.ID, "OppB")

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Emblems.Size(); got != 1 {
		t.Fatalf("emblems = %d, want 1", got)
	}

	advanceToUpkeepOf(t, g, next)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, oppA)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(oppA) {
		t.Error("the opponent's chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(oppB) {
		t.Error("the opponent sacrificed more than one creature")
	}
	if !g.Battlefield.Contains(mineA) {
		t.Error("the emblem's owner sacrificed on an opponent's upkeep")
	}
}

func TestSorinSolemnEmblemSpareTheOwnersOwnUpkeep(t *testing.T) {
	g, seat, me, sorin := sorinATable(t, "Sorin, Solemn Visitor", sorinSolemnOracle, 6)
	mineA := pushCreatureToBattlefieldForTest(g, me.ID, "MineA")

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	// Walk once around the table to the owner's next upkeep.
	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)

	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the emblem asked its own owner to sacrifice")
	}
	if !g.Battlefield.Contains(mineA) {
		t.Error("the emblem made its own owner sacrifice")
	}
}
