package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushKnight puts a 2/2 Knight on the battlefield under `owner`.
func pushKnight(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Human Knight",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// --- the front face ------------------------------------------------

// TestInvasionOfNewPhyrexiaMakesXKnights is the front face: X = 3
// announced, three 2/2 white-and-blue Knights with vigilance, and a
// battle on the battlefield with six defense counters.
func TestInvasionOfNewPhyrexiaMakesXKnights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]

	id := effectstest.ImportToHand(effectstest.NewPhyrexiaRow(), owner)
	effectstest.ToMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast the Siege for X = 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
	effectstest.PassPriorityAroundTable(t, g)

	knights := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Knight" || c.Controller != owner.ID {
			continue
		}
		knights++
		eff := c.Effective()
		if eff.Power != 2 || eff.Toughness != 2 {
			t.Errorf("Knight is %d/%d, want 2/2", eff.Power, eff.Toughness)
		}
		if !game.HasKeyword(&c, "vigilance") {
			t.Error("the Knight token has no vigilance")
		}
		if len(c.Colors) != 2 {
			t.Errorf("Knight colours = %v, want white and blue", c.Colors)
		}
	}
	if knights != 3 {
		t.Errorf("made %d Knights, want X = 3", knights)
	}

	battle := effectstest.BattlefieldCardFor(g, id)
	if battle == nil {
		t.Fatal("the Siege never reached the battlefield")
	}
	if got := battle.Counters[game.CounterDefense]; got != 6 {
		t.Errorf("defense = %d, want the printed 6", got)
	}
}

// TestInvasionOfNewPhyrexiaAtXZeroMakesNothing — "(X can be 0)" is
// not printed on this one, but X = 0 is always a legal announcement
// (CR 601.2b) and the card must not object to it.
func TestInvasionOfNewPhyrexiaAtXZeroMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]

	id := effectstest.ImportToHand(effectstest.NewPhyrexiaRow(), owner)
	effectstest.ToMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{XValue: 0}); err != nil {
		t.Fatalf("cast the Siege for X = 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountNamed(g, "Knight"); n != 0 {
		t.Errorf("X = 0 made %d Knights", n)
	}
	if effectstest.BattlefieldCardFor(g, id) == nil {
		t.Error("the Siege did not enter")
	}
}

// --- the transformed cast -------------------------------------------

// TestInvasionOfNewPhyrexiaBecomesTeferiAkosa is the whole card: the
// Siege enters, is defeated, is exiled with a grant for face 1, and
// the back face is cast for nothing out of exile — landing as a
// four-loyalty planeswalker whose own abilities are reachable under
// the "#1" catalog key.
func TestInvasionOfNewPhyrexiaBecomesTeferiAkosa(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]

	row := effectstest.NewPhyrexiaRow()
	imported := deck.ToGameCard(row, false)
	if imported.FaceCount() != 2 {
		t.Fatalf("imported %d faces, want 2", imported.FaceCount())
	}
	if imported.Faces[1].StartingLoyalty != 4 {
		t.Fatalf("back-face loyalty imported as %d, want 4 — the planeswalker would land dead",
			imported.Faces[1].StartingLoyalty)
	}

	id := effectstest.ImportToHand(row, owner)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move the Siege to the battlefield: %v", err)
	}
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.DefeatBattle(t, g, id, 6)
	if !effectstest.InExile(g, id) {
		t.Fatal("the defeated Siege was not exiled")
	}
	grant := effectstest.ExileGrantFor(g, id)
	if face, ok := grant.NamedFace(); !ok || face != effects.SiegeBackFace || grant.Player != owner.ID {
		t.Fatalf("grant = %+v, want face 1 for the battle's controller", grant)
	}

	effectstest.ToMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the transformed back face: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	landed := effectstest.BattlefieldCardFor(g, id)
	if landed == nil {
		t.Fatal("Teferi Akosa of Zhalfir never reached the battlefield")
	}
	if landed.Name != "Teferi Akosa of Zhalfir" {
		t.Fatalf("permanent is %q, want the planeswalker", landed.Name)
	}
	if landed.IsBattle() {
		t.Fatal("the defeated battle re-entered as a battle")
	}
	if got := landed.Counters[game.CounterLoyalty]; got != 4 {
		t.Errorf("loyalty = %d, want the printed 4", got)
	}
	// The back face's own catalog entry is what we are really after:
	// its −2 makes an emblem, and nothing but the "#1" key can find
	// the declaration of it.
	if err := g.ActivateCatalogAbility(owner.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2 on the transformed planeswalker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.EmblemsOf(g, owner.ID)); n != 1 {
		t.Errorf("the back face's −2 made %d emblems, want 1", n)
	}
}

// --- the +1 ----------------------------------------------------------

// TestTeferiAkosaPlusOneDrawsTwoAndAsksForTwoOrACreature is the
// set-level rule: the discard prompt's floor is ONE, its ceiling is
// two, and which one-card answers are legal is decided by the SET
// predicate rather than by the count.
func TestTeferiAkosaPlusOneDrawsTwoAndAsksForTwoOrACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	pw := effectstest.PushTeferiAkosa(g, owner.ID, 4)
	effectstest.AdvanceToMainOf(t, g, seat)

	// A hand of exactly one land plus one creature card, so the two
	// answers the clause allows are "both" and "the creature".
	for _, c := range append([]uuid.UUID(nil), effectstest.HandIDs(g, owner.ID)...) {
		if err := g.MoveCardByID(
			game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
			game.ZoneRef{Kind: game.ZoneLibrary, Owner: owner.ID}, c,
		); err != nil {
			t.Fatalf("clear the hand: %v", err)
		}
	}
	before := g.PlayerByIDForEffect(owner.ID).Hand.Size()
	if before != 0 {
		t.Fatalf("hand not cleared: %d cards", before)
	}
	land := effectstest.PushToHand(g, owner.ID, "Plains", "Basic Land — Plains")
	beast := effectstest.PushToHand(g, owner.ID, "Bear", "Creature — Bear")

	if err := g.ActivateCatalogAbility(owner.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := g.PlayerByIDForEffect(owner.ID).Hand.Size(); got != 4 {
		t.Fatalf("hand is %d after drawing two onto two, want 4", got)
	}
	prompt := effectstest.DiscardChoiceFor(g, owner.ID)
	if prompt == nil {
		t.Fatal("the +1 asked for no discard")
	}
	if prompt.ChooseMin != 1 || prompt.ChooseMax != 2 {
		t.Errorf("discard bounds = [%d, %d], want [1, 2] — one creature card is enough",
			prompt.ChooseMin, prompt.ChooseMax)
	}

	// One noncreature card is NOT enough, and the refusal is the
	// prompt's Validate rather than its bounds.
	if err := g.ResolveChooseCards(prompt.ID, owner.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("discarding one land satisfied \"unless you discard a creature card\"")
	}
	if effectstest.DiscardChoiceFor(g, owner.ID) == nil {
		t.Fatal("a rejected submit closed the prompt")
	}

	// One creature card is.
	if err := g.ResolveChooseCards(prompt.ID, owner.ID, []uuid.UUID{beast}); err != nil {
		t.Fatalf("discarding one creature card was refused: %v", err)
	}
	if effectstest.DiscardChoiceFor(g, owner.ID) != nil {
		t.Error("the discard prompt is still open after a legal answer")
	}
	if got := g.PlayerByIDForEffect(owner.ID).Hand.Size(); got != 3 {
		t.Errorf("hand is %d after pitching one creature, want 3", got)
	}
	if !effectstest.InGraveyardOf(g, owner.ID, beast) {
		t.Error("the creature card was not discarded")
	}
}

// TestTeferiAkosaPlusOneTakesTwoOfAnything — the other legal answer.
// Two cards pay the clause whatever they are, which is the printed
// "discard two cards" before the "unless".
func TestTeferiAkosaPlusOneTakesTwoOfAnything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	pw := effectstest.PushTeferiAkosa(g, owner.ID, 4)
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	prompt := effectstest.DiscardChoiceFor(g, owner.ID)
	if prompt == nil {
		t.Fatal("the +1 asked for no discard")
	}
	hand := effectstest.HandIDs(g, owner.ID)
	if len(hand) < 2 {
		t.Fatalf("hand of %d is too small for this test", len(hand))
	}
	if err := g.ResolveChooseCards(prompt.ID, owner.ID, hand[:2]); err != nil {
		t.Fatalf("discarding two cards was refused: %v", err)
	}
	for _, id := range hand[:2] {
		if !effectstest.InGraveyardOf(g, owner.ID, id) {
			t.Errorf("card %s was not discarded", id)
		}
	}
}

// --- the −2 -----------------------------------------------------------

// TestTeferiAkosaEmblemPumpsAndWardsYourKnights is the emblem, both
// halves, through the one predicate: your Knights are 3/2 and an
// opponent's spell aimed at one of them raises a ward {1} payment.
func TestTeferiAkosaEmblemPumpsAndWardsYourKnights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := effectstest.PushTeferiAkosa(g, owner.ID, 4)
	mine := pushKnight(g, owner.ID, "My Knight")
	theirs := pushKnight(g, opp.ID, "Their Knight")
	notAKnight := effectstest.PushCreatureToBattlefieldForTest(g, owner.ID, "Just A Bear")
	effectstest.AdvanceToMainOf(t, g, seat)

	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 2 || tough != 2 {
		t.Fatalf("baseline Knight is %d/%d, want 2/2", p, tough)
	}

	// Index 1 is the −2; the +1 and the −3 are 0 and 2.
	if err := g.ActivateCatalogAbility(owner.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	emblems := effectstest.EmblemsOf(g, owner.ID)
	if len(emblems) != 1 {
		t.Fatalf("owner has %d emblems, want 1", len(emblems))
	}
	if emblems[0].Text != "Knights you control get +1/+0 and have ward {1}." {
		t.Errorf("emblem text = %q", emblems[0].Text)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("my Knight is %d/%d, want 3/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's Knight is %d/%d — \"you control\" is not being read", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, notAKnight); p != 2 || tough != 2 {
		t.Errorf("a non-Knight I control is %d/%d — \"Knights\" is not being read", p, tough)
	}

	// The ward: the opponent targets my Knight and is asked to pay.
	// A warded permanent is a LEGAL target, so the announce succeeds
	// and the payment is offered to the SPELL's controller.
	effectstest.CastAtWardedCreature(t, g, opp, mine)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("the emblem's ward did not trigger on an opponent's spell")
	}
	if effectstest.HasPayUnlessFor(g, owner.ID) {
		t.Error("the emblem's controller was asked to pay for somebody else's spell")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("a declined ward must counter the spell — the Knight should live")
	}
}

// TestTeferiAkosasWardIsNotYourOwnProblem — "an opponent controls" is
// measured against the WARDED permanent's controller, so the emblem's
// owner targeting their own Knight pays nothing (CR 702.21a).
func TestTeferiAkosasWardIsNotYourOwnProblem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	pw := effectstest.PushTeferiAkosa(g, owner.ID, 4)
	mine := pushKnight(g, owner.ID, "My Knight")
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.CastAtWardedCreature(t, g, owner, mine)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.HasPayUnlessFor(g, owner.ID) {
		t.Error("the emblem warded its own controller's spell")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("nothing stopped the Doom Blade, so the Knight should be gone")
	}
}

// TestTeferiAkosasEmblemOutlivesTeferi — CR 114: an emblem is an
// OBJECT, not a continuous effect the planeswalker sources, so killing
// Teferi changes neither half of it. The ward half is the one worth
// asserting: it is a trigger harvested from the command zone, and a
// granted ward that quietly stopped working when its granter died
// would look exactly like a working one until somebody tried to
// remove a Knight.
func TestTeferiAkosasEmblemOutlivesTeferi(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := effectstest.PushTeferiAkosa(g, owner.ID, 4)
	mine := pushKnight(g, owner.ID, "My Knight")
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(pw); err != nil {
			t.Fatalf("destroy the planeswalker: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pw) {
		t.Fatal("the planeswalker survived being destroyed")
	}

	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("my Knight is %d/%d after Teferi died, want 3/2 — the emblem is its own object", p, tough)
	}
	effectstest.CastAtWardedCreature(t, g, opp, mine)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("the ward died with the planeswalker — it belongs to the EMBLEM")
	}
}

// --- the −3 -----------------------------------------------------------

// TestTeferiAkosaMinusThreeTapsThenShuffles is the reflexive half:
// two creatures tapped, so the trigger's clause is "mana value 2 or
// less" and it is offered only the permanents that fit.
func TestTeferiAkosaMinusThreeTapsThenShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := effectstest.PushTeferiAkosa(g, owner.ID, 4)
	a := effectstest.PushCreatureToBattlefieldForTest(g, owner.ID, "Tapper A")
	b := effectstest.PushCreatureToBattlefieldForTest(g, owner.ID, "Tapper B")

	cheap := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: cheap, Name: "Two Drop", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	dear := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: dear, Name: "Five Drop", TypeLine: "Creature — Giant",
		ManaCost: "{4}{G}", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AdvanceToMainOf(t, g, seat)

	// Index 2 is the −3.
	if err := g.ActivateCatalogAbility(owner.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	tap := effectstest.LatestChooseCards(g, owner.ID)
	if tap == nil {
		t.Fatal("the −3 offered no creatures to tap")
	}
	if tap.ChooseMin != 0 {
		t.Errorf("tap floor = %d, want 0 — \"any number\" includes none", tap.ChooseMin)
	}
	if err := g.ResolveChooseCards(tap.ID, owner.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("tap two creatures: %v", err)
	}
	for _, id := range []uuid.UUID{a, b} {
		if c := effectstest.BattlefieldCardFor(g, id); c == nil || !c.Tapped {
			t.Errorf("creature %s was not tapped", id)
		}
	}

	pick := effectstest.LatestPickTarget(g, owner.ID)
	if pick == nil {
		t.Fatal("no reflexive trigger asked for a target")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.PickTargetCards {
		offered[id] = true
	}
	if !offered[cheap] {
		t.Error("a mana value 2 permanent was not offered to a two-creature tap")
	}
	if offered[dear] {
		t.Error("a mana value 5 permanent was offered when X was 2")
	}
	if offered[a] || offered[b] {
		t.Error("the trigger was offered a permanent its own controller controls")
	}

	if err := g.ResolvePickTarget(pick.ID, owner.ID, game.TargetRef{Kind: game.TargetCard, ID: cheap}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(cheap) {
		t.Error("the targeted permanent is still on the battlefield")
	}
	if !effectstest.InLibrary(g, opp.ID, cheap) {
		t.Error("the targeted permanent did not end up in its owner's library")
	}
}

// TestTeferiAkosaMinusThreeTappingNothingDoesNothing — "when you tap
// one or more creatures this way" is conditional on the tapping, so
// an empty answer creates no trigger at all (CR 603.12).
func TestTeferiAkosaMinusThreeTappingNothingDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := effectstest.PushTeferiAkosa(g, owner.ID, 4)
	effectstest.PushCreatureToBattlefieldForTest(g, owner.ID, "Untapped")
	victim := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Safe")
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	tap := effectstest.LatestChooseCards(g, owner.ID)
	if tap == nil {
		t.Fatal("the −3 offered no creatures to tap")
	}
	if err := g.ResolveChooseCards(tap.ID, owner.ID, nil); err != nil {
		t.Fatalf("tapping nothing was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.LatestPickTarget(g, owner.ID) != nil {
		t.Error("a reflexive trigger was created though nothing was tapped")
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("something got shuffled away after tapping nothing")
	}
}
