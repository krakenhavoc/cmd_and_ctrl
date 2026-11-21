package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sorinWalker puts one of the three on the battlefield as a real entry
// (so its statics are stamped) with `loyalty`.
func sorinWalker(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, OracleID: oracle,
		TypeLine: "Legendary Planeswalker — Sorin", Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

func sorinTable(t *testing.T) (g *game.Game, me, opp *game.Player, seat int) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	seat = g.Turn.ActiveSeat
	me, opp = g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	effectstest.AdvanceToMainOf(t, g, seat)
	return g, me, opp, seat
}

func sorinActivate(t *testing.T, g *game.Game, me *game.Player, sorin uuid.UUID, idx int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, sorin, idx, params); err != nil {
		t.Fatalf("activate ability %d: %v", idx, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Sorin, Vampire Lord ---------------------------------------------------

func TestSorinVampireLordPlusOnePumpsUpToOneCreature(t *testing.T) {
	g, me, _, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Vampire Lord", effectstest.SorinVampireLordOracle, 4)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")

	sorinActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{Targets: effectstest.DgCardRef(bear)})
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power = %d, want 4 (+2/+0)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 2 {
		t.Errorf("toughness = %d, want 2 (the +1 is +2/+0)", got)
	}
	if got := effectstest.LoyaltyOf(g, sorin); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
}

func TestSorinVampireLordPlusOneWithNoTargetIsLegal(t *testing.T) {
	g, me, _, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Vampire Lord", effectstest.SorinVampireLordOracle, 4)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")

	sorinActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{})
	if got := effectstest.LoyaltyOf(g, sorin); got != 5 {
		t.Errorf("loyalty = %d, want 5 — \"up to one\" may be zero", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power = %d, want 2 — nothing was targeted", got)
	}
}

func TestSorinVampireLordMinusTwoDealsFourAndGainsFour(t *testing.T) {
	g, me, opp, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Vampire Lord", effectstest.SorinVampireLordOracle, 4)
	myLife, theirLife := me.Life, opp.Life

	sorinActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != theirLife-4 {
		t.Errorf("opponent life = %d, want %d", opp.Life, theirLife-4)
	}
	if me.Life != myLife+4 {
		t.Errorf("my life = %d, want %d", me.Life, myLife+4)
	}
	if got := effectstest.LoyaltyOf(g, sorin); got != 2 {
		t.Errorf("loyalty = %d, want 2", got)
	}
}

// The −8 grants the steal to the Vampires you control NOW, not to a
// non-Vampire, not to an opponent's Vampire, and not to a Vampire that
// arrives later; the steal is indefinite while the grant ends with the turn.
func TestSorinVampireLordMinusEightGrantsTheStealToYourVampires(t *testing.T) {
	g, me, opp, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Vampire Lord", effectstest.SorinVampireLordOracle, 8)
	vamp := effectstest.B16Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2, "B")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirVamp := effectstest.B16Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 2, 2, "B")
	prey := effectstest.B16Creature(g, opp.ID, "Prey", "Creature — Ogre", 3, 3, "R")

	sorinActivate(t, g, me, sorin, 2, game.ActivateAbilityParams{})
	idx, ref := effectstest.GrantedActivatedIndex(t, g, vamp)
	if idx < 0 {
		t.Fatal("my Vampire did not gain the steal")
	}
	if i, _ := effectstest.GrantedActivatedIndex(t, g, bear); i >= 0 {
		t.Error("a non-Vampire gained the steal")
	}
	if i, _ := effectstest.GrantedActivatedIndex(t, g, theirVamp); i >= 0 {
		t.Error("an opponent's Vampire gained the steal")
	}
	late := effectstest.B16Creature(g, me.ID, "Late Vampire", "Creature — Vampire", 2, 2, "B")
	if i, _ := effectstest.GrantedActivatedIndex(t, g, late); i >= 0 {
		t.Error("a Vampire that arrived after the ability resolved gained the steal (CR 611.2c)")
	}

	effectstest.B16Activate(t, g, me.ID, vamp, idx, game.ActivateAbilityParams{Ref: ref, Targets: effectstest.DgCardRef(prey)})
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, prey); !ok || c.Controller != me.ID {
		t.Errorf("the target creature is not under my control (found=%v)", ok)
	}
	if !effectstest.B16Tapped(t, g, vamp) {
		t.Error("{T}: the Vampire that used the steal is tapped")
	}

	effectstest.AdvancePastCleanupForTest(t, g)
	if i, _ := effectstest.GrantedActivatedIndex(t, g, vamp); i >= 0 {
		t.Error("the granted steal outlived the turn")
	}
	if c, ok := effectstest.BattlefieldCard(g, prey); !ok || c.Controller != me.ID {
		t.Error("the stolen creature went back at end of turn; the control has no stated duration (CR 611.2a)")
	}
}

// --- Sorin, Imperious Bloodlord --------------------------------------------

func TestSorinImperiousFirstPlusOneGrantsAndCountersAVampire(t *testing.T) {
	g, me, _, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Imperious Bloodlord", effectstest.SorinImperiousBloodlordOr, 4)
	vamp := effectstest.B16Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2, "B")

	sorinActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{Targets: effectstest.DgCardRef(vamp)})
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, vamp), kw) {
			t.Errorf("the Vampire did not gain %s", kw)
		}
	}
	if got := effectstest.CounterCount(g, vamp, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 — it is a Vampire", got)
	}
}

func TestSorinImperiousFirstPlusOneNoCounterOnANonVampire(t *testing.T) {
	g, me, _, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Imperious Bloodlord", effectstest.SorinImperiousBloodlordOr, 4)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")

	sorinActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{Targets: effectstest.DgCardRef(bear)})
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "deathtouch") || !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "lifelink") {
		t.Error("the Bear should gain deathtouch and lifelink")
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0 — it is not a Vampire", got)
	}
}

func TestSorinImperiousFirstPlusOneOnlyTargetsYourCreatures(t *testing.T) {
	g, me, opp, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Imperious Bloodlord", effectstest.SorinImperiousBloodlordOr, 4)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")

	if err := g.ActivateCatalogAbility(me.ID, sorin, 0, game.ActivateAbilityParams{Targets: effectstest.DgCardRef(theirs)}); err == nil {
		t.Fatal("an opponent's creature was a legal target for \"target creature you control\"")
	}
}

func TestSorinImperiousSecondPlusOneSacrificesAVampireForDamageAndLife(t *testing.T) {
	g, me, opp, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Imperious Bloodlord", effectstest.SorinImperiousBloodlordOr, 4)
	vamp := effectstest.B16Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2, "B")
	myLife, theirLife := me.Life, opp.Life

	if err := g.ActivateCatalogAbility(me.ID, sorin, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, vamp)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(vamp) {
		t.Error("the Vampire was not sacrificed")
	}
	if opp.Life != theirLife-3 {
		t.Errorf("opponent life = %d, want %d", opp.Life, theirLife-3)
	}
	if me.Life != myLife+3 {
		t.Errorf("my life = %d, want %d", me.Life, myLife+3)
	}
	if got := effectstest.LoyaltyOf(g, sorin); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
}

func TestSorinImperiousSecondPlusOneDeclinedDoesNothing(t *testing.T) {
	g, me, opp, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Imperious Bloodlord", effectstest.SorinImperiousBloodlordOr, 4)
	vamp := effectstest.B16Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2, "B")
	myLife, theirLife := me.Life, opp.Life

	if err := g.ActivateCatalogAbility(me.ID, sorin, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(vamp) {
		t.Error("the Vampire was sacrificed after the player said no")
	}
	if opp.Life != theirLife || me.Life != myLife {
		t.Errorf("life moved (me %d→%d, opp %d→%d) with no sacrifice, so \"when you do\" never happened",
			myLife, me.Life, theirLife, opp.Life)
	}
}

// With no Vampire there is nothing to sacrifice and no question.
func TestSorinImperiousSecondPlusOneAsksNothingWithoutAVampire(t *testing.T) {
	g, me, opp, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Imperious Bloodlord", effectstest.SorinImperiousBloodlordOr, 4)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirLife := opp.Life

	if err := g.ActivateCatalogAbility(me.ID, sorin, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
		t.Fatal("asked to sacrifice a Vampire with none in play")
	}
	if !g.Battlefield.Contains(bear) || opp.Life != theirLife {
		t.Error("something happened with no Vampire to sacrifice")
	}
}

func TestSorinImperiousMinusThreePutsAVampireFromHand(t *testing.T) {
	g, me, _, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Imperious Bloodlord", effectstest.SorinImperiousBloodlordOr, 4)
	vamp := effectstest.HandCardForTest(me, "Hand Vampire", "Creature — Vampire", "")
	bear := effectstest.HandCardForTest(me, "Hand Bear", "Creature — Bear", "")
	rock := effectstest.HandCardForTest(me, "Hand Artifact", "Artifact — Vampire", "")

	if err := g.ActivateCatalogAbility(me.ID, sorin, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the put")
	}
	if pick.ChooseMin != 0 {
		t.Errorf("floor = %d, want 0 — \"you may\"", pick.ChooseMin)
	}
	if !effectstest.HasID(pick.ChooseCards, vamp) || effectstest.HasID(pick.ChooseCards, bear) || effectstest.HasID(pick.ChooseCards, rock) {
		t.Errorf("offered %v, want only the Vampire creature card", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, vamp)
	if !g.Battlefield.Contains(vamp) || me.Hand.Contains(vamp) {
		t.Error("the Vampire did not come out of the hand onto the battlefield")
	}
	if got := effectstest.LoyaltyOf(g, sorin); got != 1 {
		t.Errorf("loyalty = %d, want 1", got)
	}
}

func TestSorinImperiousMinusThreeMayDecline(t *testing.T) {
	g, me, _, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Imperious Bloodlord", effectstest.SorinImperiousBloodlordOr, 4)
	vamp := effectstest.HandCardForTest(me, "Hand Vampire", "Creature — Vampire", "")

	if err := g.ActivateCatalogAbility(me.ID, sorin, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if g.Battlefield.Contains(vamp) || !me.Hand.Contains(vamp) {
		t.Error("declined, yet the Vampire moved")
	}
}

// --- Sorin, Vengeful Bloodlord ---------------------------------------------

func TestSorinVengefulGivesLifelinkOnlyOnYourTurn(t *testing.T) {
	g, me, opp, _ := sorinTable(t)
	sorinWalker(g, me.ID, "Sorin, Vengeful Bloodlord", effectstest.SorinVengefulBloodlordOracl, 4)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	walker := sorinWalker(g, me.ID, "Other Walker", "", 3)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")

	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "lifelink") {
		t.Error("my creature lacks lifelink on my turn")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, walker), "lifelink") {
		t.Error("my other planeswalker lacks lifelink on my turn")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "lifelink") {
		t.Error("an opponent's creature has lifelink")
	}

	// On another player's turn the window is closed.
	effectstest.AdvancePastCleanupForTest(t, g)
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "lifelink") {
		t.Error("my creature has lifelink on an opponent's turn")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, walker), "lifelink") {
		t.Error("my planeswalker has lifelink on an opponent's turn")
	}
}

// Sorin is a planeswalker you control, so his own +2 gains you life on your
// turn (CR 702.15b): the proof that lifelink reaches a source that is not a
// creature.
func TestSorinVengefulPlusTwoHitsAPlayerAndLifelinks(t *testing.T) {
	g, me, opp, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Vengeful Bloodlord", effectstest.SorinVengefulBloodlordOracl, 4)
	myLife, theirLife := me.Life, opp.Life

	sorinActivate(t, g, me, sorin, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != theirLife-1 {
		t.Errorf("opponent life = %d, want %d", opp.Life, theirLife-1)
	}
	if me.Life != myLife+1 {
		t.Errorf("my life = %d, want %d — Sorin has lifelink on my turn", me.Life, myLife+1)
	}
	if got := effectstest.LoyaltyOf(g, sorin); got != 6 {
		t.Errorf("loyalty = %d, want 6", got)
	}
}

func TestSorinVengefulMinusXReturnsACreatureOfThatManaValueAsAVampire(t *testing.T) {
	g, me, _, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Vengeful Bloodlord", effectstest.SorinVengefulBloodlordOracl, 4)
	three := effectstest.PushGraveyardCreature(g, me.ID, "Three Drop", "{2}{G}")
	two := effectstest.PushGraveyardCreature(g, me.ID, "Two Drop", "{1}{G}")

	// A creature of the wrong mana value is not a legal target.
	if err := g.ActivateCatalogAbility(me.ID, sorin, 1, game.ActivateAbilityParams{
		XValue: 2, Targets: effectstest.DgCardRef(three)}); err == nil {
		t.Fatal("a mana value 3 creature was a legal target with X = 2")
	}

	sorinActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{XValue: 3, Targets: effectstest.DgCardRef(three)})
	if got := effectstest.LoyaltyOf(g, sorin); got != 1 {
		t.Errorf("loyalty = %d, want 1 — X = 3 comes off 4", got)
	}
	c, ok := effectstest.BattlefieldCard(g, three)
	if !ok {
		t.Fatal("the creature did not return to the battlefield")
	}
	if c.Controller != me.ID {
		t.Error("it returned under someone else's control")
	}
	if !effectstest.Contains(effectstest.EffectiveSubtypes(t, g, three), "Vampire") {
		t.Errorf("subtypes = %v, want Vampire added", effectstest.EffectiveSubtypes(t, g, three))
	}
	if !effectstest.Contains(effectstest.EffectiveSubtypes(t, g, three), "Bear") {
		t.Errorf("subtypes = %v, want its other types kept", effectstest.EffectiveSubtypes(t, g, three))
	}
	if g.Battlefield.Contains(two) {
		t.Error("the other graveyard creature returned too")
	}
}

// The Vampire clause lasts exactly as long as the returned object does.
func TestSorinVengefulVampireClauseEndsWhenTheCreatureLeaves(t *testing.T) {
	g, me, _, _ := sorinTable(t)
	sorin := sorinWalker(g, me.ID, "Sorin, Vengeful Bloodlord", effectstest.SorinVengefulBloodlordOracl, 4)
	one := effectstest.PushGraveyardCreature(g, me.ID, "One Drop", "{G}")

	sorinActivate(t, g, me, sorin, 1, game.ActivateAbilityParams{XValue: 1, Targets: effectstest.DgCardRef(one)})
	if !effectstest.Contains(effectstest.EffectiveSubtypes(t, g, one), "Vampire") {
		t.Fatal("setup: it should be a Vampire while it stays")
	}
	effectstest.B18Kill(t, g, one)
	if g.Battlefield.Contains(one) {
		t.Fatal("setup: the creature should have died")
	}
	if zone, c := effectstest.CardWhere(g, me, one); zone != "graveyard" || c.HasSubtype("Vampire") {
		t.Errorf("in the graveyard as %q with Vampire=%v; the clause should be gone (CR 400.7)", zone, c.HasSubtype("Vampire"))
	}
}
