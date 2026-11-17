package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ob_nixilis_the_adversary_test.go — Ob Nixilis, the Adversary and the
// casualty X seam (#5668).

const obNixilisAdversaryOracle = "a6abb6af-7ea2-43f3-8718-2a077ad0c947"

func TestObNixilisTheAdversaryIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(obNixilisAdversaryOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
	if spec.StartingLoyalty != 3 {
		t.Errorf("StartingLoyalty = %d, want 3", spec.StartingLoyalty)
	}
}

// obNixInHand puts Ob Nixilis in the active seat's hand at its main
// phase, with {1}{B}{R} in the pool.
func obNixInHand(t *testing.T, g *game.Game) (me *game.Player, ob uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ob = uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: ob, Name: "Ob Nixilis, the Adversary",
			TypeLine: "Legendary Planeswalker — Nixilis", ManaCost: "{1}{B}{R}",
			OracleID: obNixilisAdversaryOracle, StartingLoyalty: 3, Owner: me.ID, Controller: me.ID})
		fillPoolColored(me, "B", 1)
		fillPoolColored(me, "R", 1)
		fillPool(me, 1)
	})
	return me, ob
}

// obNixWalkers lists the Ob Nixilis permanents `who` controls: the card
// first, then the tokens.
func obNixWalkers(g *game.Game, who uuid.UUID) (cards, tokens []game.Card) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller != who || c.Name != "Ob Nixilis, the Adversary" {
			continue
		}
		if c.IsToken() {
			tokens = append(tokens, c)
		} else {
			cards = append(cards, c)
		}
	}
	return cards, tokens
}

// The whole clause: any creature may be sacrificed, X is its power as it
// last existed (counters included), and the copy is a nonlegendary token
// with starting loyalty X beside the legendary card with its printed 3.
// A restore point taken with the casualty trigger on the stack keeps X.
func TestObNixilisCasualtyXCopiesWithTheSacrificedPower(t *testing.T) {
	for _, restore := range []bool{false, true} {
		name := "live"
		if restore {
			name = "restored with the trigger on the stack"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, ob := obNixInHand(t, g)
			bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == bear {
						g.Battlefield.Cards[i].Counters = map[string]int{"+1/+1": 3}
					}
				}
			})
			if err := g.CastSpell(me.ID, ob, game.CastSpellParams{
				Strict:        true,
				OptionalCosts: []int{0},
				SacrificeIDs:  []uuid.UUID{bear},
			}); err != nil {
				t.Fatalf("cast with casualty X: %v", err)
			}
			if g.Battlefield.Contains(bear) {
				t.Fatal("the creature was not sacrificed as the spell was cast")
			}
			if triggerOnStack(g, ob) == nil {
				t.Fatal("no casualty trigger on the stack")
			}
			if restore {
				g = restoreRoundTrip(t, g, true)
				me = g.Seats[g.Turn.ActiveSeat]
			}
			passPriorityAroundTable(t, g)
			g.RunStateChecksForTest()

			cards, tokens := obNixWalkers(g, me.ID)
			if len(cards) != 1 || len(tokens) != 1 {
				t.Fatalf("%d cards and %d tokens, want one of each: the copy is not legendary, so both stay", len(cards), len(tokens))
			}
			if !cards[0].IsLegendary() || cards[0].Counters[game.CounterLoyalty] != 3 {
				t.Errorf("the card: legendary %v, loyalty %d; want legendary with 3", cards[0].IsLegendary(), cards[0].Counters[game.CounterLoyalty])
			}
			if tokens[0].IsLegendary() {
				t.Error("the copy is legendary")
			}
			if got := tokens[0].Counters[game.CounterLoyalty]; got != 5 {
				t.Errorf("the copy entered with %d loyalty, want 5 (a 2/2 with three +1/+1 counters)", got)
			}
			if !tokens[0].IsPlaneswalker() || tokens[0].OracleID != obNixilisAdversaryOracle {
				t.Errorf("the copy is not Ob Nixilis: %q %q", tokens[0].TypeLine, tokens[0].OracleID)
			}
			// The exception is a copiable value (CR 707.9b): the token
			// after a restore still has it.
			after := restoreRoundTrip(t, g, true)
			_, restoredTokens := obNixWalkers(after, me.ID)
			if len(restoredTokens) != 1 || restoredTokens[0].StartingLoyalty != 5 || !restoredTokens[0].StartingLoyaltyKnown {
				t.Errorf("the restored copy lost its starting loyalty: %+v", restoredTokens)
			}
		})
	}
}

// X is 0 for a creature with power 0: the copy enters with no loyalty
// counters, not the printed 3, and CR 704.5i puts it into the graveyard.
func TestObNixilisCasualtyZeroCopyHasNoLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	me, ob := obNixInHand(t, g)
	wall := pushVanillaCreature(g, me.ID, "Wall", 0, 3)
	if err := g.CastSpell(me.ID, ob, game.CastSpellParams{
		Strict:        true,
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{wall},
	}); err != nil {
		t.Fatalf("cast with casualty 0: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	cards, tokens := obNixWalkers(g, me.ID)
	if len(tokens) != 0 {
		t.Errorf("a copy with starting loyalty 0 is still on the battlefield with %d loyalty", tokens[0].Counters[game.CounterLoyalty])
	}
	if len(cards) != 1 || cards[0].Counters[game.CounterLoyalty] != 3 {
		t.Errorf("the card itself: %+v, want one with 3 loyalty", cards)
	}
}

// Casualty is optional: an unpaid cast is one legendary walker and no
// trigger.
func TestObNixilisWithoutCasualtyIsNotCopied(t *testing.T) {
	g := newCatalogGame(t)
	me, ob := obNixInHand(t, g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.CastSpell(me.ID, ob, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if triggerOnStack(g, ob) != nil {
		t.Fatal("a casualty trigger with no casualty cost paid")
	}
	passPriorityAroundTable(t, g)
	cards, tokens := obNixWalkers(g, me.ID)
	if len(cards) != 1 || len(tokens) != 0 || cards[0].Counters[game.CounterLoyalty] != 3 {
		t.Errorf("%d cards, %d tokens", len(cards), len(tokens))
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was sacrificed though casualty was not chosen")
	}
}

// The cost is one creature the caster controls: a noncreature, an
// opponent's creature, two creatures and no creature are all refused.
func TestObNixilisCasualtyRefusesAnIllegalSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, ob := obNixInHand(t, g)
	rock := pushPermanentForTest(g, me.ID, "Rock", "", "Artifact")
	theirs := pushVanillaCreature(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Theirs", 3, 3)
	a := pushVanillaCreature(g, me.ID, "A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "B", 1, 1)
	for name, ids := range map[string][]uuid.UUID{
		"an artifact":            {rock},
		"an opponent's creature": {theirs},
		"two creatures":          {a, b},
		"nothing":                nil,
	} {
		if err := g.CastSpell(me.ID, ob, game.CastSpellParams{
			Strict:        true,
			OptionalCosts: []int{0},
			SacrificeIDs:  ids,
		}); err == nil {
			t.Fatalf("casualty X paid with %s was accepted", name)
		}
	}
	if !me.Hand.Contains(ob) {
		t.Error("a refused cast moved the card")
	}
}

// Register holds casualty X to the one shape the trigger can read.
func TestCasualtyXIsHeldToItsShape(t *testing.T) {
	sac := sacrificeSpec("a creature", Creature())
	mustPanic(t, "mandatory additional cost CasualtyX", func() {
		checkCasualtyX(Spec{Name: "Mandatory", AdditionalCost: &game.AdditionalCost{Sacrifice: sac, CasualtyX: true}})
	})
	mustPanic(t, "by hand", func() {
		checkCasualtyX(Spec{Name: "By Hand", OptionalCosts: []game.AdditionalCost{{Optional: true, Key: game.KickerKey, Sacrifice: sac, CasualtyX: true}}})
	})
	mustPanic(t, "mandatory sacrifice", func() {
		checkCasualtyX(Spec{Name: "Two Sacrifices", AdditionalCost: &game.AdditionalCost{Sacrifice: sac},
			OptionalCosts: []game.AdditionalCost{CasualtyX()}})
	})
	mustPanic(t, "optional cost", func() {
		checkCasualtyX(Spec{Name: "Buyback Too", OptionalCosts: []game.AdditionalCost{CasualtyX(),
			{Optional: true, Key: game.BuybackKey, Sacrifice: sac}}})
	})
	mustPanic(t, "never a granted one", func() {
		SpellsYouCastHave("Spells you cast have casualty X.", nil, CasualtyX())
	})
	checkCasualtyX(Spec{Name: "Ob Nixilis, the Adversary", OptionalCosts: []game.AdditionalCost{CasualtyX()}})
}

// obNixOnBattlefield puts the walker on the active seat's battlefield at
// its main phase with `loyalty` counters.
func obNixOnBattlefield(t *testing.T, g *game.Game, loyalty int) (me *game.Player, pw uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me = g.Seats[seat]
	pw = pushWalkerForTest(g, me.ID, "Ob Nixilis, the Adversary", obNixilisAdversaryOracle, loyalty)
	advanceToMainOf(t, g, seat)
	return me, pw
}

// The +1: each opponent in turn order chooses, one question at a time;
// an opponent with no card in hand is not asked and loses the life; and
// with a Devil the controller gains 2 after the last opponent.
func TestObNixilisPlusOneAsksEachOpponentThenGainsWithADevil(t *testing.T) {
	g := newCatalogGame(t)
	me, pw := obNixOnBattlefield(t, g, 3)
	seat := g.Turn.ActiveSeat
	loser, discarder, broke := g.Seats[(seat+1)%4], g.Seats[(seat+2)%4], g.Seats[(seat+3)%4]
	g.WithWriteLock(func() {
		broke.Hand.Cards = nil
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Devil", TypeLine: "Token Creature — Devil",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	lives := map[uuid.UUID]int{me.ID: me.Life, loser.ID: loser.Life, discarder.ID: discarder.Life, broke.ID: broke.Life}
	hand := discarder.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	if got := loyaltyOf(g, pw); got != 4 {
		t.Errorf("loyalty after +1 = %d, want 4", got)
	}
	monSettle(t, g)
	if latestOptionPickFor(g, discarder.ID) != nil {
		t.Fatal("the second opponent was asked before the first answered")
	}
	c := latestOptionPickFor(g, loser.ID)
	if c == nil {
		t.Fatal("no question for the first opponent")
	}
	if c.PickOptions[0].LifeCost != 2 {
		t.Errorf("the life branch costs %d, want 2", c.PickOptions[0].LifeCost)
	}
	answerOptionPick(t, g, loser.ID, 0)
	if me.Life != lives[me.ID] {
		t.Fatal("the life was gained before every opponent had chosen")
	}
	answerOptionPick(t, g, discarder.ID, 1)
	if discardChoiceFor(g, discarder.ID) == nil {
		t.Fatal("no discard prompt after choosing to discard")
	}
	answerDiscard(t, g, discarder.ID, discarder.Hand.Cards[0].InstanceID)
	monSettle(t, g)

	if got := lives[loser.ID] - loser.Life; got != 2 {
		t.Errorf("the opponent who chose life lost %d, want 2", got)
	}
	if discarder.Life != lives[discarder.ID] || hand-discarder.Hand.Size() != 1 {
		t.Errorf("the discarding opponent: life %d (was %d), discarded %d", discarder.Life, lives[discarder.ID], hand-discarder.Hand.Size())
	}
	if got := lives[broke.ID] - broke.Life; got != 2 {
		t.Errorf("the opponent with no hand lost %d, want 2 without being asked", got)
	}
	if got := me.Life - lives[me.ID]; got != 2 {
		t.Errorf("gained %d with a Devil, want 2", got)
	}
}

// Without a Demon or Devil there is no life gain; a Demon counts as a
// Devil does.
func TestObNixilisPlusOneGainsOnlyWithADemonOrDevil(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		want           int
	}{
		{"neither", "Creature — Bear", 0},
		{"a Demon", "Creature — Demon", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, pw := obNixOnBattlefield(t, g, 3)
			g.WithWriteLock(func() {
				for _, p := range g.Seats {
					if p.ID != me.ID {
						p.Hand.Cards = nil
					}
				}
				g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Other", TypeLine: tc.typeLine,
					Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
			})
			before := me.Life
			if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("+1: %v", err)
			}
			monSettle(t, g)
			for _, p := range g.Seats {
				if p.ID != me.ID && p.Life != 38 {
					t.Errorf("an opponent with no hand is at %d, want 38", p.Life)
				}
			}
			if got := me.Life - before; got != tc.want {
				t.Errorf("gained %d, want %d", got, tc.want)
			}
		})
	}
}

func TestObNixilisMinusTwoMakesADevil(t *testing.T) {
	g := newCatalogGame(t)
	me, pw := obNixOnBattlefield(t, g, 3)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	if got := loyaltyOf(g, pw); got != 1 {
		t.Errorf("loyalty after −2 = %d, want 1", got)
	}
	monSettle(t, g)
	var devil *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Devil" && c.Controller == me.ID {
			devil = c
		}
	}
	if devil == nil {
		t.Fatal("no Devil token")
	}
	if !devil.IsToken() || devil.Power != 1 || devil.Toughness != 1 || len(devil.Colors) != 1 || devil.Colors[0] != "R" || !devil.HasSubtype("Devil") {
		t.Errorf("the token is not a 1/1 red Devil: %+v", *devil)
	}
}

// The −7: the target draws seven and loses 7. It needs seven loyalty,
// and a player to target.
func TestObNixilisMinusSevenDrawsSevenAndLosesSeven(t *testing.T) {
	g := newCatalogGame(t)
	me, pw := obNixOnBattlefield(t, g, 7)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	hand, life := opp.Hand.Size(), opp.Life
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the −7 was activated with no target")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	monSettle(t, g)
	if got := opp.Hand.Size() - hand; got != 7 {
		t.Errorf("the target drew %d, want 7", got)
	}
	if got := life - opp.Life; got != 7 {
		t.Errorf("the target lost %d, want 7", got)
	}
}

func TestObNixilisMinusSevenNeedsSevenLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	me, pw := obNixOnBattlefield(t, g, 6)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Fatal("the −7 was activated with six loyalty")
	}
}
