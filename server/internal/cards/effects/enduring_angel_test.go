package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enduring_angel_test.go covers Enduring Angel // Angelic Enforcer and,
// through it, the engine path of #5678 (game/life_instead.go) on the real
// double-faced card: every way a life total falls to 0 or less, the copy
// that cannot transform, and a restore across the CR 616 ordering prompt.
// newCatalogGame leaves seat 0 active.

// enduringAngelRow is the Scryfall `transform` record the importer reads.
func enduringAngelRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Enduring Angel // Angelic Enforcer",
		Layout:   "transform",
		TypeLine: "Creature — Angel // Creature — Angel",
		OracleID: uuid.MustParse(enduringAngelOracleID),
		Keywords: []string{"Flying", "Transform", "Double strike"},
		CardFaces: []cards.CardFace{
			{Name: "Enduring Angel", TypeLine: "Creature — Angel", ManaCost: "{2}{W}{W}{W}", Colors: []string{"W"}, Power: "3", Toughness: "3",
				OracleText: "Flying, double strike\nYou have hexproof.\nIf your life total would be reduced to 0 or less, instead transform this creature and your life total becomes 3. Then if this creature didn't transform this way, you lose the game."},
			{Name: "Angelic Enforcer", TypeLine: "Creature — Angel", Colors: []string{"W"}, Power: "*", Toughness: "*",
				OracleText: "Flying\nYou have hexproof.\nAngelic Enforcer's power and toughness are each equal to your life total.\nWhenever this creature attacks, double your life total."},
		},
	}
}

// putEnduringAngel puts the imported card onto p's battlefield, front
// face up and able to attack.
func putEnduringAngel(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(enduringAngelRow(), p)
	c, ok := cardInZone(p.Hand, id)
	if !ok {
		t.Fatal("Enduring Angel did not reach the hand")
	}
	p.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	return id
}

// putSingleFacedEnduringAngel is a permanent with Enduring Angel's
// abilities on a single-faced object, as a token copy of it is: it has
// no other face to turn to (CR 701.27c).
func putSingleFacedEnduringAngel(g *game.Game, p *game.Player) uuid.UUID {
	return b12Push(g, p.ID, "Enduring Angel", "Creature — Angel", enduringAngelOracleID, 3, 3)
}

func loseLife(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, -n); err != nil && !game.IsStopResolution(err) {
			t.Fatalf("lose %d life: %v", n, err)
		}
	})
}

func lifeChangesOf(g *game.Game, p *game.Player) []int {
	var out []int
	for _, ev := range g.Events {
		if ev.Kind == game.EventChangeLife && ev.Target == p.ID {
			out = append(out, ev.Amount)
		}
	}
	return out
}

func TestEnduringAngelIsRegisteredOnBothFaces(t *testing.T) {
	front, ok := Lookup(enduringAngelOracleID)
	if !ok || front.Name != "Enduring Angel" || front.Completeness != CompletenessFull || len(front.Caveats) != 0 {
		t.Errorf("front face registered %v as %q (%v, caveats %v)", ok, front.Name, front.Completeness, front.Caveats)
	}
	back, ok := Lookup(enduringAngelOracleID + "#1")
	if !ok || back.Name != "Angelic Enforcer" || back.Completeness != CompletenessFull || len(back.Caveats) != 0 {
		t.Errorf("back face registered %v as %q (%v, caveats %v)", ok, back.Name, back.Completeness, back.Caveats)
	}
}

// Flying, double strike, and "You have hexproof" on both faces: an
// opponent's Bolt at the Angel's controller is refused before and after
// the transform, and their own is not.
func TestEnduringAngelKeywordsAndHexproofOnBothFaces(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := putEnduringAngel(t, g, opp)
	for _, kw := range []string{"flying", "double strike"} {
		if !hasEffectiveKeyword(t, g, angel, kw) {
			t.Errorf("Enduring Angel lacks %s", kw)
		}
	}
	if p, tough := effectivePTOf(t, g, angel); p != 3 || tough != 3 {
		t.Errorf("Enduring Angel is %d/%d, want 3/3", p, tough)
	}
	if err := boltAtPlayerErr(t, g, me, opp.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at the Angel's controller: got %v, want ErrIllegalTarget", err)
	}

	iwTransform(t, g, angel)
	if !hasEffectiveKeyword(t, g, angel, "flying") || hasEffectiveKeyword(t, g, angel, "double strike") {
		t.Error("Angelic Enforcer has flying and not double strike")
	}
	if err := boltAtPlayerErr(t, g, me, opp.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at the Enforcer's controller: got %v, want ErrIllegalTarget", err)
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(angel); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if err := boltAtPlayerErr(t, g, me, opp.ID); err != nil {
		t.Errorf("with the Angel gone the Bolt is legal: %v", err)
	}
}

// A loss that would leave the total at 0 or less: the Angel transforms,
// the total becomes 3, and the player is not dead at the next check. A
// smaller loss is left alone.
func TestEnduringAngelReplacesALethalLifeLoss(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	angel := putEnduringAngel(t, g, me)

	loseLife(t, g, me, me.Life-1)
	if me.Life != 1 || iwFace(g, angel) != 0 {
		t.Fatalf("a loss down to 1: life %d, face %d; want 1 and the front face", me.Life, iwFace(g, angel))
	}

	me.Life = 10
	loseLife(t, g, me, 10)
	if me.Life != 3 {
		t.Errorf("losing 10 at 10: life %d, want 3", me.Life)
	}
	if iwFace(g, angel) != 1 {
		t.Fatalf("Enduring Angel did not transform (face %d)", iwFace(g, angel))
	}
	if got := lifeChangesOf(g, me); got[len(got)-1] != -7 {
		t.Errorf("the life change landed as %d, want -7 (CR 119.5)", got[len(got)-1])
	}
	g.RunStateChecksForTest()
	if me.Eliminated {
		t.Fatal("the player lost the game at 3 life")
	}
	if p, tough := effectivePTOf(t, g, angel); p != 3 || tough != 3 {
		t.Errorf("Angelic Enforcer is %d/%d at 3 life, want 3/3", p, tough)
	}

	// Angelic Enforcer has no such ability: the next lethal loss lands.
	loseLife(t, g, me, 5)
	if me.Life != -2 {
		t.Errorf("a lethal loss with only the Enforcer: life %d, want -2", me.Life)
	}
}

// "Your life total becomes 0" (a set-life effect, CR 119.5) and a payment
// of all your life (CR 119.4) are reductions to 0 too; the payment is
// still paid.
func TestEnduringAngelReplacesSetLifeAndAPayment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	mine := putEnduringAngel(t, g, me)
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: opp.ID})
		if err := b31LifeBecomes(ctx, me.ID, 0); err != nil {
			t.Fatalf("life total becomes 0: %v", err)
		}
	})
	if me.Life != 3 || iwFace(g, mine) != 1 {
		t.Errorf("set to 0: life %d, face %d; want 3 and transformed", me.Life, iwFace(g, mine))
	}

	theirs := putEnduringAngel(t, g, opp)
	opp.Life = 6
	g.WithWriteLock(func() {
		if err := g.PayLifeForEffect(uuid.Nil, opp.ID, 6); err != nil {
			t.Fatalf("paying 6 life at 6: %v", err)
		}
	})
	if opp.Life != 3 || iwFace(g, theirs) != 1 {
		t.Errorf("paying all 6: life %d, face %d; want 3 and transformed", opp.Life, iwFace(g, theirs))
	}
}

// Damage is still dealt whole; only the life it costs is replaced (CR
// 120.3a). At 2 life the total becomes 3 by a gain of 1, and at 3 no
// life moves at all.
func TestEnduringAngelReplacesLethalDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushVanillaCreature(g, me.ID, "Pinger", 1, 1)

	angel := putEnduringAngel(t, g, opp)
	opp.Life = 10
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 12) })
	if opp.Life != 3 || iwFace(g, angel) != 1 {
		t.Fatalf("12 damage at 10: life %d, face %d; want 3 and transformed", opp.Life, iwFace(g, angel))
	}
	ev, ok := lastDamageTo(g, opp.ID)
	if !ok || ev.Amount != 12 || ev.DamageLifeLoss() != 7 {
		t.Errorf("damage event: %d dealt, %d life lost; want 12 and 7", ev.Amount, ev.DamageLifeLoss())
	}

	second := putEnduringAngel(t, g, opp)
	opp.Life = 2
	before := len(lifeChangesOf(g, opp))
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 5) })
	if opp.Life != 3 || iwFace(g, second) != 1 {
		t.Fatalf("5 damage at 2: life %d, face %d; want 3 and transformed", opp.Life, iwFace(g, second))
	}
	if got := lifeChangesOf(g, opp)[before:]; len(got) != 1 || got[0] != 1 {
		t.Errorf("5 damage at 2: life changes %v, want one gain of 1", got)
	}
	if ev, _ := lastDamageTo(g, opp.ID); ev.Amount != 5 || ev.DamageLifeLoss() != 0 {
		t.Errorf("damage event: %d dealt, %d life lost; want 5 and 0", ev.Amount, ev.DamageLifeLoss())
	}

	third := putEnduringAngel(t, g, opp)
	before = len(lifeChangesOf(g, opp))
	loseLife(t, g, opp, 4)
	if opp.Life != 3 || iwFace(g, third) != 1 {
		t.Errorf("losing 4 at 3: life %d, face %d; want 3 and transformed", opp.Life, iwFace(g, third))
	}
	if got := lifeChangesOf(g, opp)[before:]; len(got) != 0 {
		t.Errorf("losing 4 at 3 moved no life but emitted %v", got)
	}
}

// Lethal combat damage: the defending player ends combat at 3 life with
// Angelic Enforcer, still in the game.
func TestEnduringAngelReplacesLethalCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := putEnduringAngel(t, g, opp)
	big := pushVanillaCreature(g, me.ID, "Giant", 9, 9)
	small := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	opp.Life = 10

	attackWith(t, g, opp.ID, big, small)
	g.RunStateChecksForTest()

	if iwFace(g, angel) != 1 {
		t.Fatalf("Enduring Angel did not transform under lethal combat damage (face %d)", iwFace(g, angel))
	}
	if opp.Eliminated {
		t.Fatal("the defending player lost the game")
	}
	// The two attackers' damage is one batch of separate events, so the
	// one that would take the total to 0 or less is the one replaced:
	// either order ends between 1 and 3.
	if opp.Life < 1 || opp.Life > 3 {
		t.Errorf("life %d after 11 combat damage at 10, want 1 to 3", opp.Life)
	}
	if p, tough := effectivePTOf(t, g, angel); p != opp.Life || tough != opp.Life {
		t.Errorf("Angelic Enforcer is %d/%d at %d life", p, tough, opp.Life)
	}
}

// Two Enduring Angels: one answers the event, and the other is left for
// the next one.
func TestOnlyOneEnduringAngelTransforms(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a, b := putEnduringAngel(t, g, me), putEnduringAngel(t, g, me)

	loseLife(t, g, me, me.Life+5)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("two copies of one replacement queued %d prompts, want 0", len(g.PendingChoices))
	}
	if me.Life != 3 || iwFace(g, a)+iwFace(g, b) != 1 {
		t.Fatalf("first loss: life %d, faces %d and %d; want 3 and exactly one transformed", me.Life, iwFace(g, a), iwFace(g, b))
	}
	loseLife(t, g, me, 3)
	if me.Life != 3 || iwFace(g, a) != 1 || iwFace(g, b) != 1 {
		t.Errorf("second loss: life %d, faces %d and %d; want 3 and both transformed", me.Life, iwFace(g, a), iwFace(g, b))
	}
}

// "Then if this creature didn't transform this way, you lose the game":
// a single-faced Enduring Angel can't transform, so the total becomes 3
// and its controller loses (CR 104.3e).
func TestEnduringAngelThatCannotTransformLosesTheGame(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	putSingleFacedEnduringAngel(g, opp)
	opp.Life = 8

	loseLife(t, g, opp, 8)
	if opp.Life != 3 {
		t.Errorf("life %d, want 3: the total becomes 3 first", opp.Life)
	}
	if !opp.Eliminated || opp.LossCause != game.LossEffect {
		t.Fatalf("eliminated %v by %q, want an effect loss", opp.Eliminated, opp.LossCause)
	}
	g.RunStateChecksForTest()
	if g.State != game.StateActive || me.Eliminated {
		t.Errorf("the other three play on: state %q", g.State)
	}
}

// The same from lethal combat damage, in the middle of a damage step.
func TestSingleFacedEnduringAngelLosesToLethalCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	putSingleFacedEnduringAngel(g, opp)
	big := pushVanillaCreature(g, me.ID, "Giant", 9, 9)
	small := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	opp.Life = 5

	attackWith(t, g, opp.ID, big, small)
	g.RunStateChecksForTest()

	if !opp.Eliminated || opp.LossCause != game.LossEffect {
		t.Fatalf("eliminated %v by %q, want an effect loss", opp.Eliminated, opp.LossCause)
	}
	if g.State != game.StateActive || me.Eliminated {
		t.Errorf("the other three play on: state %q", g.State)
	}
}

// An Enduring Angel that can't transform (Immerwolf holds a non-Human
// Werewolf) does not, and its controller loses; one who can't lose the
// game (Platinum Angel) is left at 3 life.
func TestEnduringAngelUnderCantTransformAndCantLose(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	held := putEnduringAngel(t, g, opp)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == held {
				g.Battlefield.Cards[i].TypeLine = "Creature — Werewolf Angel"
			}
		}
	})
	b12Push(g, opp.ID, "Immerwolf", "Creature — Wolf", iwImmerwolfOracle, 2, 2)
	var can bool
	g.WithWriteLock(func() { can = g.CanTransformForEffect(held) })
	if can {
		t.Fatal("the test's Werewolf Angel can still transform under Immerwolf")
	}
	opp.Life = 4
	loseLife(t, g, opp, 9)
	for _, ev := range g.Events {
		if ev.Kind == game.EventTransform && ev.CardID == held {
			t.Fatal("a Werewolf Enduring Angel transformed under Immerwolf")
		}
	}
	if opp.Life != 3 || !opp.Eliminated {
		t.Errorf("life %d, eliminated %v; want 3 and out of the game", opp.Life, opp.Eliminated)
	}

	putSingleFacedEnduringAngel(g, me)
	pushPermanentForTest(g, me.ID, "Platinum Angel", platinumAngelOrc, "Artifact Creature — Angel")
	me.Life = 4
	loseLife(t, g, me, 9)
	g.RunStateChecksForTest()
	if me.Life != 3 || me.Eliminated {
		t.Errorf("under Platinum Angel: life %d, eliminated %v; want 3 and in the game", me.Life, me.Eliminated)
	}
}

// Angelic Enforcer: power and toughness follow the life total, and its
// attack trigger doubles it (CR 701.10d).
func TestAngelicEnforcerFollowsAndDoublesYourLifeTotal(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := putEnduringAngel(t, g, me)
	iwTransform(t, g, angel)

	me.Life = 17
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	if p, tough := effectivePTOf(t, g, angel); p != 18 || tough != 18 {
		t.Fatalf("Angelic Enforcer is %d/%d at 18 life", p, tough)
	}

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(angel, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if me.Life != 36 {
		t.Errorf("after the attack trigger: life %d, want 36", me.Life)
	}
	if p, tough := effectivePTOf(t, g, angel); p != 36 || tough != 36 {
		t.Errorf("Angelic Enforcer is %d/%d at 36 life", p, tough)
	}
}

// Two different replacements on one loss ask the affected player to
// order them (CR 616.1), and the prompt is a restore point: the game
// that comes back applies the Angel when the answer arrives. The order
// matters. Bloodletter first doubles the loss to 20 and the Angel then
// makes the total 3; the Angel first makes the loss 7, which Bloodletter
// doubles to 14, and the Angel gets no second look at it (CR 614.5), so
// that player is at -4 and loses to the state-based check.
func TestEnduringAngelOrderingPromptSurvivesARestore(t *testing.T) {
	seen := map[int]bool{}
	for _, reversed := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushBloodletter(g, me.ID) // seat 0 is active: an opponent's loss is doubled
		angel := putEnduringAngel(t, g, opp)
		opp.Life = 10

		loseLife(t, g, opp, 10)
		if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
			t.Fatalf("want the CR 616 ordering prompt, got %d choices", len(g.PendingChoices))
		}
		if opp.Life != 10 || iwFace(g, angel) != 0 {
			t.Fatalf("before the answer: life %d, face %d", opp.Life, iwFace(g, angel))
		}

		r := restoreRoundTrip(t, g, true)
		rOpp := r.Seats[1]
		if len(r.PendingChoices) != 1 || rOpp.ID != opp.ID {
			t.Fatalf("the restored game has %d choices, want the prompt", len(r.PendingChoices))
		}
		c := r.PendingChoices[0]
		order := append([]game.ReplacementEffectID(nil), c.ReplacementEffectIDs...)
		if len(order) != 2 {
			t.Fatalf("the prompt orders %d effects, want 2", len(order))
		}
		if reversed {
			order[0], order[1] = order[1], order[0]
		}
		if err := r.ResolveReplacementOrder(c.ID, c.Chooser, order); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		// The player at -4 has left the game and the Angel with them, so
		// the transform is read off the log.
		transformed := false
		for _, ev := range r.Events {
			transformed = transformed || (ev.Kind == game.EventTransform && ev.CardID == angel)
		}
		if !transformed {
			t.Errorf("reversed=%v: Enduring Angel did not transform in the restored game", reversed)
		}
		if rOpp.Eliminated != (rOpp.Life <= 0) {
			t.Errorf("reversed=%v: life %d, eliminated %v", reversed, rOpp.Life, rOpp.Eliminated)
		}
		seen[rOpp.Life] = true
	}
	if !seen[3] || !seen[-4] || len(seen) != 2 {
		t.Errorf("the two orders ended at %v, want 3 (Bloodletter first) and -4 (the Angel first)", seen)
	}
}
