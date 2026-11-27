package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b602_cards_test.go — Mirrodin, S82 OLD batch 602 (#5606). Each
// test plays the real catalog entry and asserts what its text says.

const (
	oldB602Crown     = "0ad26c65-bc2c-4aa7-8c5d-e4c8f0629cb7"
	oldB602Spellbomb = "3f351696-1402-4b69-b924-f0ba14320287"
	oldB602Pendulum  = "4f88ddfa-5478-426a-a978-dfc64c687db1"
	oldB602Membrane  = "5a2deb96-01f4-48d5-bec8-44e4f141f551"
	oldB602Fatespin  = "676935c8-22e0-427f-9566-16778cea0873"
	oldB602Gambit    = "78497c54-7179-41b7-9454-7b915fd96cd5"
	oldB602Augur     = "d5409a7c-b928-4819-a560-e80226cf61f0"
	oldB602Fate      = "e23bf5fe-e494-473e-8502-606106ee1820"
	oldB602Spoils    = "ef86f861-1fdf-4be0-9497-8adc064af3de"
)

func TestOldB602CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB602Crown: "Mindstorm Crown", oldB602Spellbomb: "Lifespark Spellbomb", oldB602Pendulum: "Liar's Pendulum",
		oldB602Membrane: "Psychic Membrane", oldB602Fatespin: "Fatespinner", oldB602Gambit: "Fiery Gambit",
		oldB602Augur: "Lumengrid Augur", oldB602Fate: "Shared Fate", oldB602Spoils: "Spoils of the Vault",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB602Steps(t *testing.T, g *game.Game) []game.Step {
	t.Helper()
	seat, seq := g.Turn.ActiveSeat, g.Turn.Seq
	var steps []game.Step
	for i := 0; i < 40; i++ {
		if g.Turn.ActiveSeat != seat || g.Turn.Seq != seq {
			return steps
		}
		steps = append(steps, g.Turn.Step)
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("the turn never ended")
	return nil
}

func TestOldB602MindstormCrownDrawsOnAnEmptyHandAndHurtsOtherwise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mindstorm Crown", "Artifact", oldB602Crown, 0, 0)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 1 || me.Life != life {
		t.Errorf("empty hand: hand %d, life %d → %d; want 1 card drawn and no damage", me.Hand.Size(), life, me.Life)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	life, hand := me.Life, me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life-1 {
		t.Errorf("a card in hand: hand %d → %d, life %d → %d; want no draw and 1 damage", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestOldB602LifesparkSpellbombAnimatesALandOrDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bomb := effectstest.B12Push(g, me.ID, "Lifespark Spellbomb", "Artifact", oldB602Spellbomb, 0, 0)
	land := effectstest.PushLand(g, me.ID, "Forest")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		// A creature is a legal pick only if it is a land: the Bear is not.
		effectstest.ApaMana(me, "G")
	} else {
		t.Fatal("a nonland permanent was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(land)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(bomb) {
		t.Error("the Spellbomb was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardCopy(t, g, land)
	if p, tough := effectstest.EffectivePTOf(t, g, land); p != 3 || tough != 3 {
		t.Errorf("the land is %d/%d, want 3/3", p, tough)
	}
	if !c.IsLand() {
		t.Error("the animated land stopped being a land")
	}
	effectstest.DcCleanup(g)
	if _, ok := effectstest.BattlefieldCard(g, land); !ok {
		t.Fatal("the land left the battlefield")
	}
	if effectstest.BattlefieldCardCopy(t, g, land).IsCreature() {
		t.Error("the land is still a creature after the cleanup step")
	}
}

func TestOldB602LifesparkSpellbombDrawsForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bomb := effectstest.B12Push(g, me.ID, "Lifespark Spellbomb", "Artifact", oldB602Spellbomb, 0, 0)
	effectstest.ApaMana(me, "C")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, bomb, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || g.Battlefield.Contains(bomb) {
		t.Errorf("hand %d → %d, bomb on battlefield %v; want a card drawn and the Spellbomb gone", hand, me.Hand.Size(), g.Battlefield.Contains(bomb))
	}
}

func TestOldB602PsychicMembraneMayDrawWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.B12Push(g, me.ID, "Psychic Membrane", "Creature — Wall", oldB602Membrane, 0, 3)
	attacker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Bear",
		Keywords: []string{"vigilance"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(attacker, me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn for blocking", hand, me.Hand.Size())
	}
}

func TestOldB602LumengridAugurLootsAnyPlayerAndUntapsForAnArtifact(t *testing.T) {
	for _, artifact := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		augur := effectstest.B12Push(g, me.ID, "Lumengrid Augur", "Creature — Vedalken Wizard", oldB602Augur, 2, 2)
		g.WithWriteLock(func() {
			me.Hand.Cards = nil
			for i := range g.Battlefield.Cards {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		})
		typeLine := "Instant"
		if artifact {
			typeLine = "Artifact"
		}
		held := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: held, Name: "Held", TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		effectstest.ApaMana(me, "C")
		if err := g.ActivateCatalogAbility(me.ID, augur, 0, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, held)
		if !me.Graveyard.Contains(held) {
			t.Fatalf("artifact=%v: the picked card was not discarded", artifact)
		}
		c, _ := effectstest.BattlefieldCard(g, augur)
		if c.Tapped == artifact {
			t.Errorf("artifact=%v: the Augur's tapped state is %v; want it untapped only when an artifact was discarded", artifact, c.Tapped)
		}
	}
}

func TestOldB602FatespinnerLetsTheOpponentSkipOneStepOrPhase(t *testing.T) {
	for _, tc := range []struct {
		index   int
		skipped func(game.Step) bool
		kept    func(game.Step) bool
	}{
		{0, func(s game.Step) bool { return s == game.StepDraw }, func(s game.Step) bool { return s == game.StepPrecombatMain }},
		{1, func(s game.Step) bool { return s == game.StepPrecombatMain || s == game.StepPostcombatMain }, func(s game.Step) bool { return s == game.StepDraw }},
		{2, func(s game.Step) bool { return game.PhaseOf(s) == game.PhaseCombat }, func(s game.Step) bool { return s == game.StepPrecombatMain }},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Fatespinner", "Creature — Human Wizard", oldB602Fatespin, 1, 2)
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestOptionPickFor(g, me.ID) != nil {
			t.Fatal("the Fatespinner's controller was asked to choose")
		}
		effectstest.AnswerOptionPick(t, g, opp.ID, tc.index)
		steps := oldB602Steps(t, g)
		for _, s := range steps {
			if tc.skipped(s) {
				t.Errorf("choice %d: step %s was not skipped", tc.index, s)
			}
		}
		kept := false
		for _, s := range steps {
			kept = kept || tc.kept(s)
		}
		if !kept {
			t.Errorf("choice %d: the other steps were skipped too: %v", tc.index, steps)
		}
	}
}

func TestOldB602FieryGambitScalesWithTheFlipsWon(t *testing.T) {
	reached := map[int]bool{}
	for attempt := 0; attempt < 400 && len(reached) < 3; attempt++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 9, 9)
		land := effectstest.PushLand(g, me.ID, "Island")
		g.WithWriteLock(func() { g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Tapped = true })
		oppLife, hand := opp.Life, me.Hand.Size()
		stopAt := attempt%3 + 1
		effectstest.CastCatalogSpell(t, g, "Fiery Gambit", "Sorcery", oldB602Gambit, effectstest.CardTarget(bear))
		effectstest.PassPriorityAroundTable(t, g)
		wins, lost := 0, false
		for i := 0; i < 20 && !lost; i++ {
			call := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
			if call == nil {
				break
			}
			before := len(g.Events)
			if err := g.ResolveCoinCall(call.ID, me.ID, "heads"); err != nil {
				t.Fatal(err)
			}
			won := false
			for _, ev := range g.Events[before:] {
				if ev.Kind == game.EventFlipCoin {
					won = ev.Won
				}
			}
			if !won {
				lost = true
				break
			}
			wins++
			if wins == stopAt {
				effectstest.AnswerMayChoice(t, g, me.ID, false)
				break
			}
			effectstest.AnswerMayChoice(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
		dealt := 0
		if c, ok := effectstest.BattlefieldCard(g, bear); ok {
			dealt = c.DamageMarked
		} else {
			dealt = 9
		}
		switch {
		case lost:
			if dealt != 0 || opp.Life != oppLife || me.Hand.Size() != hand {
				t.Fatalf("attempt %d: a lost flip still had an effect (damage %d, life %d → %d, hand %d → %d)", attempt, dealt, oppLife, opp.Life, hand, me.Hand.Size())
			}
		default:
			reached[wins] = true
			if dealt != 3 {
				t.Errorf("%d wins: the creature has %d damage, want 3", wins, dealt)
			}
			wantLife := oppLife
			if wins >= 2 {
				wantLife -= 6
			}
			if opp.Life != wantLife {
				t.Errorf("%d wins: the opponent's life %d → %d, want %d", wins, oppLife, opp.Life, wantLife)
			}
			c := effectstest.BattlefieldCardCopy(t, g, land)
			if wins >= 3 {
				if me.Hand.Size() < hand+9 || c.Tapped {
					t.Errorf("3 wins: hand %d → %d, land tapped %v; want nine drawn and the land untapped", hand, me.Hand.Size(), c.Tapped)
				}
			} else if me.Hand.Size() != hand || !c.Tapped {
				t.Errorf("%d wins: hand %d → %d, land tapped %v; want no draw and no untap", wins, hand, me.Hand.Size(), c.Tapped)
			}
		}
	}
	for wins := 1; wins <= 3; wins++ {
		if !reached[wins] {
			t.Errorf("no attempt stopped after exactly %d wins; the %d-win clause is untested", wins, wins)
		}
	}
}

func TestOldB602SpoilsOfTheVaultFindsTheNameAndExilesTheRestAtALifeEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	after := effectstest.SweepFTop(me, "After", "Sorcery")
	wanted := effectstest.SweepFTop(me, "Wanted", "Sorcery")
	a := effectstest.SweepFTop(me, "Junk A", "Sorcery")
	b := effectstest.SweepFTop(me, "Junk B", "Sorcery")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Spoils of the Vault", "Instant", oldB602Spoils, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Wanted")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, wanted) {
		t.Error("the named card is not in hand")
	}
	if !effectstest.InExile(g, a) || !effectstest.InExile(g, b) {
		t.Error("a revealed card that was not named was not exiled")
	}
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want -2 (two exiled cards)", life, me.Life)
	}
	if !me.Library.Contains(after) {
		t.Error("a card below the named one left the library")
	}
}

func TestOldB602SpoilsOfTheVaultWithANameNotInTheLibraryExilesItAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	n := len(me.Library.Cards)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Spoils of the Vault", "Instant", oldB602Spoils, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "A Card Nobody Printed")
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Library.Cards) != 0 {
		t.Errorf("%d cards left in the library, want it all revealed and exiled", len(me.Library.Cards))
	}
	if me.Life != life-n {
		t.Errorf("life %d → %d, want -%d (a life per exiled card)", life, me.Life, n)
	}
}

func TestOldB602LiarsPendulumDrawsOnlyOnAWrongGuessAndAReveal(t *testing.T) {
	for _, tc := range []struct {
		name        string
		inHand      bool
		guessInHand bool
		reveal      bool
		wantDrawn   bool
	}{
		{"wrong guess, reveal", true, false, true, true},
		{"wrong guess, no reveal", true, false, false, false},
		{"right guess, reveal", true, true, true, false},
		{"wrong guess about an absent card, reveal", false, true, true, true},
		{"right guess about an absent card", false, false, true, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		pendulum := effectstest.B12Push(g, me.ID, "Liar's Pendulum", "Artifact", oldB602Pendulum, 0, 0)
		g.WithWriteLock(func() { me.Hand.Cards = nil })
		other := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: other, Name: "Other Card", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		if tc.inHand {
			me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Secret Card", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		}
		size := me.Hand.Size()
		effectstest.ApaMana(me, "C", "C")
		if err := g.ActivateCatalogAbility(me.ID, pendulum, 0, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.ActNameFor(t, g, me.ID, "Secret Card")
		effectstest.AnswerMayChoice(t, g, opp.ID, tc.guessInHand)
		effectstest.AnswerMayChoice(t, g, me.ID, tc.reveal)
		revealed := false
		for _, c := range me.Hand.Cards {
			if c.InstanceID == other && c.KnownBy[opp.ID] {
				revealed = true
			}
		}
		drawn := me.Hand.Size() == size+1
		wantDrawn := tc.wantDrawn && tc.reveal
		if drawn != wantDrawn || revealed != tc.reveal {
			t.Errorf("%s: drew %v (want %v), revealed %v (want %v)", tc.name, drawn, wantDrawn, revealed, tc.reveal)
		}
	}
}

func TestOldB602SharedFateExilesAnOpponentsTopCardInsteadOfDrawing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Shared Fate", "Enchantment", oldB602Fate, 0, 0)
	for _, p := range g.Seats[2:] {
		g.WithWriteLock(func() { p.Library.Cards = nil })
	}
	top := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID
	myTop := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	hand, lib := me.Hand.Size(), len(me.Library.Cards)
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	if me.Hand.Size() != hand || len(me.Library.Cards) != lib || !me.Library.Contains(myTop) {
		t.Errorf("the draw was not replaced: hand %d → %d, library %d → %d", hand, me.Hand.Size(), lib, len(me.Library.Cards))
	}
	if z := effectstest.E2Zone(g, top); z != game.ZoneExile {
		t.Fatalf("the opponent's top card is in %q, want exile", z)
	}
	perm := effectstest.E2Permission(g, top)
	if perm == nil || perm.Player != me.ID || perm.CastOnly || perm.AnyColor || perm.AnyType {
		t.Errorf("permission %+v, want a play permission for the drawer, paid for normally", perm)
	}
	effectstest.E2AssertFaceDownFor(t, g, top, me.ID)
}

func TestOldB602SharedFateAsksWhichOpponentsLibraryWithSeveral(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Shared Fate", "Enchantment", oldB602Fate, 0, 0)
	g.WithWriteLock(func() { g.Seats[3].Library.Cards = nil })
	top1 := o1.Library.Cards[len(o1.Library.Cards)-1].InstanceID
	top2 := o2.Library.Cards[len(o2.Library.Cards)-1].InstanceID
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no prompt for which opponent's library")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	z1, z2 := effectstest.E2Zone(g, top1), effectstest.E2Zone(g, top2)
	if (z1 == game.ZoneExile) == (z2 == game.ZoneExile) {
		t.Errorf("exactly one library should lose its top card: first %q, second %q", z1, z2)
	}
	if me.Hand.Size() != hand {
		t.Error("the replaced draw still put a card into hand")
	}
}

func TestOldB602SharedFateReplacesEveryPlayersDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Shared Fate", "Enchantment", oldB602Fate, 0, 0)
	for _, p := range g.Seats[2:] {
		g.WithWriteLock(func() { p.Library.Cards = nil })
	}
	mine := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	hand := opp.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(opp.ID, 1); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	if opp.Hand.Size() != hand {
		t.Error("an opponent of the enchantment's controller still drew")
	}
	if z := effectstest.E2Zone(g, mine); z != game.ZoneExile {
		t.Errorf("the controller's top card is in %q, want exile (the drawing opponent exiles from their opponent's library)", z)
	}
	if perm := effectstest.E2Permission(g, mine); perm == nil || perm.Player != opp.ID {
		t.Errorf("permission %+v, want it granted to the player who would have drawn", perm)
	}
}
