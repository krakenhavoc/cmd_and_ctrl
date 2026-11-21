package testsso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	g, _, me, sorin := effectstest.SorinATable(t, "Sorin the Mirthless", effectstest.SorinMirthlessOracle, 4)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Three Drop", TypeLine: "Creature — Test", ManaCost: "{1}{1}{1}", Owner: me.ID, Controller: me.ID})
	life := me.Life

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, top)

	if !sorinInHand(me, top) {
		t.Fatal("the revealed card did not reach the hand")
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("life lost = %d, want the card's mana value 3", got)
	}
	if got := effectstest.LoyaltyOf(g, sorin); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
}

// Declining the "you may": the card stays on top and no life is paid.
func TestSorinMirthlessPlusOneDeclinedPaysNothing(t *testing.T) {
	g, _, me, sorin := effectstest.SorinATable(t, "Sorin the Mirthless", effectstest.SorinMirthlessOracle, 4)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Three Drop", TypeLine: "Creature — Test", ManaCost: "{1}{1}{1}", Owner: me.ID, Controller: me.ID})
	life := me.Life

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)

	if sorinInHand(me, top) {
		t.Error("a declined look put the card in hand")
	}
	if me.Life != life {
		t.Errorf("life changed on a declined look: %d → %d", life, me.Life)
	}
}

// A land has mana value 0: taking it costs nothing, and it is not a draw.
func TestSorinMirthlessPlusOneOnALandCostsNothing(t *testing.T) {
	g, _, me, sorin := effectstest.SorinATable(t, "Sorin the Mirthless", effectstest.SorinMirthlessOracle, 4)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	life := me.Life

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, top)

	if !sorinInHand(me, top) {
		t.Fatal("the land did not reach the hand")
	}
	if me.Life != life {
		t.Errorf("a land cost %d life", life-me.Life)
	}
}

// The −2 token is the printed 2/3 black Vampire with flying and lifelink.
func TestSorinMirthlessMinusTwoMakesAFlyingLifelinkVampire(t *testing.T) {
	g, _, me, sorin := effectstest.SorinATable(t, "Sorin the Mirthless", effectstest.SorinMirthlessOracle, 4)

	sorinAActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

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
	if p, tough := effectstest.EffectivePTOf(t, g, tok.InstanceID); p != 2 || tough != 3 {
		t.Errorf("token is %d/%d, want 2/3", p, tough)
	}
	for _, kw := range []string{"flying", "lifelink"} {
		if !effectstest.HasKeywordOf(g, tok.InstanceID, kw) {
			t.Errorf("token lacks %s", kw)
		}
	}
	if !tok.HasColor("B") {
		t.Error("token is not black")
	}
}

// The −7: 13 damage to any target and 13 life.
func TestSorinMirthlessMinusSevenDealsThirteenAndGainsThirteen(t *testing.T) {
	g, _, me, sorin := effectstest.SorinATable(t, "Sorin the Mirthless", effectstest.SorinMirthlessOracle, 7)
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine, theirs := me.Life, them.Life

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)

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
	g, _, me, sorin := effectstest.SorinATable(t, "Sorin, Grim Nemesis", effectstest.SorinGrimNemesisOracle, 6)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Four Drop", TypeLine: "Creature — Test", ManaCost: "{1}{1}{1}{1}", Owner: me.ID, Controller: me.ID})
	lives := effectstest.B17Life(g)

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

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
	if got := effectstest.LoyaltyOf(g, sorin); got != 7 {
		t.Errorf("loyalty = %d, want 7", got)
	}
}

// The −X: X damage to the creature, X life to the controller, X loyalty
// paid.
func TestSorinGrimNemesisMinusXDamagesACreatureAndGainsLife(t *testing.T) {
	g, seat, me, sorin := effectstest.SorinATable(t, "Sorin, Grim Nemesis", effectstest.SorinGrimNemesisOracle, 6)
	them := g.Seats[(seat+1)%len(g.Seats)]
	ogre := effectstest.PushSizedCreatureForTest(g, them.ID, "Ogre", "Creature — Ogre", "{2}{R}", []string{"R"}, 3, 3)
	life := me.Life

	sorinAActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{
		XValue:  3,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ogre}},
	})
	if got := effectstest.LoyaltyOf(g, sorin); got != 3 {
		t.Errorf("loyalty after −3 = %d, want 3", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(ogre) {
		t.Error("a 3/3 survived 3 damage")
	}
	if got := me.Life - life; got != 3 {
		t.Errorf("life gained = %d, want 3", got)
	}
}

// A planeswalker is a legal target, and the damage removes loyalty.
func TestSorinGrimNemesisMinusXHitsAPlaneswalker(t *testing.T) {
	g, seat, me, sorin := effectstest.SorinATable(t, "Sorin, Grim Nemesis", effectstest.SorinGrimNemesisOracle, 6)
	them := g.Seats[(seat+1)%len(g.Seats)]
	walker := effectstest.PushWalkerForTest(g, them.ID, "Some Walker", "00000000-0000-0000-0000-000000000abc", 5)

	sorinAActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: walker}},
	})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LoyaltyOf(g, walker); got != 3 {
		t.Errorf("walker loyalty = %d, want 3 after 2 damage", got)
	}
}

// A player is not a legal target: "target creature or planeswalker".
func TestSorinGrimNemesisMinusXRefusesAPlayer(t *testing.T) {
	g, seat, me, sorin := effectstest.SorinATable(t, "Sorin, Grim Nemesis", effectstest.SorinGrimNemesisOracle, 6)
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
	g, seat, me, sorin := effectstest.SorinATable(t, "Sorin, Grim Nemesis", effectstest.SorinGrimNemesisOracle, 9)
	// The highest total belongs to the LAST seat, not the controller: the
	// count is "among all players".
	for i, p := range g.Seats {
		p.Life = 3 + i
	}
	g.Seats[len(g.Seats)-1].Life = 17
	me.Life = 4
	_ = seat

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountNamed(g, "Vampire Knight"); got != 17 {
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
		if !effectstest.HasKeywordOf(g, c.InstanceID, "lifelink") || !c.HasColor("B") {
			t.Fatal("a token is not a black lifelink creature")
		}
		break
	}
}

// --- Sorin, Lord of Innistrad -----------------------------------------

func TestSorinLordPlusOneMakesALifelinkVampire(t *testing.T) {
	g, _, me, sorin := effectstest.SorinATable(t, "Sorin, Lord of Innistrad", effectstest.SorinLordOracle, 3)

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	var tok *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Vampire" {
			tok = &g.Battlefield.Cards[i]
		}
	}
	if tok == nil {
		t.Fatal("no Vampire token")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, tok.InstanceID); p != 1 || tough != 1 {
		t.Errorf("token is %d/%d, want 1/1", p, tough)
	}
	if !effectstest.HasKeywordOf(g, tok.InstanceID, "lifelink") {
		t.Error("token lacks lifelink")
	}
}

// The −2 emblem: +1/+0 to every creature its owner controls, including
// one that arrives later, and it outlives Sorin.
func TestSorinLordMinusTwoEmblemPumpsYourCreaturesForever(t *testing.T) {
	g, seat, me, sorin := effectstest.SorinATable(t, "Sorin, Lord of Innistrad", effectstest.SorinLordOracle, 3)
	them := g.Seats[(seat+1)%len(g.Seats)]
	mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	theirs := effectstest.PushCreatureToBattlefieldForTest(g, them.ID, "Theirs")

	sorinAActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Emblems.Size(); got != 1 {
		t.Fatalf("emblems = %d, want 1", got)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("my creature is %d/%d, want 3/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want an unchanged 2/2", p, tough)
	}

	later := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Later", TypeLine: "Creature — Test", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	if p, _ := effectstest.EffectivePTOf(t, g, later); p != 2 {
		t.Errorf("a creature that arrived later has power %d, want 2", p)
	}

	g.WithWriteLock(func() { g.Battlefield.Remove(sorin) })
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 3 {
		t.Errorf("after Sorin left, power %d, want the emblem's 3", p)
	}
}

// The −6: destroy up to three, return what reached a graveyard under the
// activator's control. A token is gone for good; an indestructible
// permanent is never destroyed and so never returned.
func TestSorinLordMinusSixDestroysThenReturnsThemUnderYourControl(t *testing.T) {
	g, seat, me, sorin := effectstest.SorinATable(t, "Sorin, Lord of Innistrad", effectstest.SorinLordOracle, 6)
	them := g.Seats[(seat+1)%len(g.Seats)]
	bear := effectstest.PushCreatureToBattlefieldForTest(g, them.ID, "Bear")
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
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{bear, walker} {
		c, ok := effectstest.BattlefieldCard(g, id)
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
	if got := effectstest.LoyaltyOf(g, walker); got != 4 {
		t.Errorf("the returned planeswalker has loyalty %d, want its printed 4 (a new object, CR 306.5b)", got)
	}
}

func TestSorinLordMinusSixDoesNotReturnWhatWasNotDestroyed(t *testing.T) {
	g, seat, me, sorin := effectstest.SorinATable(t, "Sorin, Lord of Innistrad", effectstest.SorinLordOracle, 6)
	them := g.Seats[(seat+1)%len(g.Seats)]
	god := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: god, Name: "Immortal", TypeLine: "Creature — God", Power: 4, Toughness: 4,
		Keywords: []string{"indestructible"}, Owner: them.ID, Controller: them.ID,
	})

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: god}},
	})
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.BattlefieldCard(g, god)
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
	g, _, me, sorin := effectstest.SorinATable(t, "Sorin, Lord of Innistrad", effectstest.SorinLordOracle, 6)

	err := g.ActivateCatalogAbility(me.ID, sorin, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: sorin}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting Sorin himself: got %v, want ErrIllegalTarget", err)
	}
}

func TestSorinLordMinusSixWithNoTargetsIsLegal(t *testing.T) {
	g, _, me, sorin := effectstest.SorinATable(t, "Sorin, Lord of Innistrad", effectstest.SorinLordOracle, 6)

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Sorin, Solemn Visitor --------------------------------------------

// The +1: +1/+0 and lifelink on the creatures you control as it
// resolves, until your next turn begins; a creature that arrives later
// gets neither.
func TestSorinSolemnPlusOneLastsUntilYourNextTurnAndLocksTheSet(t *testing.T) {
	g, seat, me, sorin := effectstest.SorinATable(t, "Sorin, Solemn Visitor", effectstest.SorinSolemnOracle, 4)
	them := g.Seats[(seat+1)%len(g.Seats)]
	mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Mine")
	theirs := effectstest.PushCreatureToBattlefieldForTest(g, them.ID, "Theirs")

	sorinAActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 3 {
		t.Errorf("my creature has power %d, want 3", p)
	}
	if !effectstest.HasKeywordOf(g, mine, "lifelink") {
		t.Error("my creature lacks lifelink")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 || effectstest.HasKeywordOf(g, theirs, "lifelink") {
		t.Error("an opponent's creature was affected")
	}
	later := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Later", TypeLine: "Creature — Test", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	if p, _ := effectstest.EffectivePTOf(t, g, later); p != 1 || effectstest.HasKeywordOf(g, later, "lifelink") {
		t.Error("a creature that arrived after resolution was affected (CR 611.2c)")
	}

	// Still on during every other player's turn…
	effectstest.AdvanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 3 || !effectstest.HasKeywordOf(g, mine, "lifelink") {
		t.Error("the effect ended before my next turn")
	}
	// …and gone once my next turn begins.
	effectstest.AdvanceToMainOf(t, g, seat)
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 2 || effectstest.HasKeywordOf(g, mine, "lifelink") {
		t.Error("the effect outlasted the start of my next turn")
	}
}

func TestSorinSolemnMinusTwoMakesAFlyingVampire(t *testing.T) {
	g, _, me, sorin := effectstest.SorinATable(t, "Sorin, Solemn Visitor", effectstest.SorinSolemnOracle, 4)

	sorinAActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	var tok *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Vampire" {
			tok = &g.Battlefield.Cards[i]
		}
	}
	if tok == nil {
		t.Fatal("no Vampire token")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, tok.InstanceID); p != 2 || tough != 2 {
		t.Errorf("token is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.HasKeywordOf(g, tok.InstanceID, "flying") || !tok.HasColor("B") {
		t.Error("token is not a black flyer")
	}
}

// The −6 emblem: at the beginning of each opponent's upkeep, that
// player sacrifices a creature of their choice. Not on its owner's own
// upkeep.
func TestSorinSolemnEmblemMakesEachOpponentSacrificeInTheirUpkeep(t *testing.T) {
	g, seat, me, sorin := effectstest.SorinATable(t, "Sorin, Solemn Visitor", effectstest.SorinSolemnOracle, 6)
	next := (seat + 1) % len(g.Seats)
	opp := g.Seats[next]
	mineA := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "MineA")
	oppA := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "OppA")
	oppB := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "OppB")

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Emblems.Size(); got != 1 {
		t.Fatalf("emblems = %d, want 1", got)
	}

	effectstest.AdvanceToUpkeepOf(t, g, next)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, oppA)
	effectstest.PassPriorityAroundTable(t, g)

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
	g, seat, me, sorin := effectstest.SorinATable(t, "Sorin, Solemn Visitor", effectstest.SorinSolemnOracle, 6)
	mineA := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "MineA")

	sorinAActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	// Walk once around the table to the owner's next upkeep.
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the emblem asked its own owner to sacrifice")
	}
	if !g.Battlefield.Contains(mineA) {
		t.Error("the emblem made its own owner sacrifice")
	}
}
