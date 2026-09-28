package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_mode_count_test.go — #1590, ADR 0065's 2026-09-27
// amendment: "Choose one. If you control a commander as you cast this
// spell, you may choose both instead." The bound is read at announce
// (CR 601.2b) through ModeSpec.RaiseMaxIf, fixed from then on, and
// only the caster's own control counts.
//
// Every refusal below is paired with an acceptance of the SAME
// announcement on a board where the condition holds, so a refusal can
// only be the mode count talking — never a target or a timing gate.

const willOfTheSultaiOracle = "951a0e34-e610-477f-b441-6680df55a29b"

// cmcSeats is the active seat and an opponent, with the table walked
// to the active seat's main phase.
func cmcSeats(t *testing.T, g *game.Game) (*game.Player, *game.Player) {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return me, opp
}

// jeskasWillBoth is the "both bullets" announcement: bullet one
// targets the opponent, bullet two targets nothing.
func jeskasWillBoth(opp uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: opp, Mode: 0}}
}

func redIn(p *game.Player) int {
	n := 0
	for _, tok := range p.ManaPool {
		if tok.Color == "R" {
			n++
		}
	}
	return n
}

// Without a commander the printed bound holds: one bullet.
func TestJeskasWillBothIsRefusedWithoutACommander(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := cmcSeats(t, g)
	if err := b22TryModal(g, jeskasWillOracle, []int{0, 1}, jeskasWillBoth(opp.ID)); err == nil {
		t.Fatal("no commander on the battlefield: choosing both must be refused")
	}
	if err := b22TryModal(g, jeskasWillOracle, []int{0}, jeskasWillBoth(opp.ID)); err != nil {
		t.Fatalf("one bullet is always legal: %v", err)
	}
}

// With your commander on the battlefield, both is accepted and both
// bullets resolve: the ritual counts the opponent's hand and the
// impulse exiles three.
func TestJeskasWillBothWithYourCommanderResolvesBothBullets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	for i := 0; i < 2; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	want := opp.Hand.Size()
	exiled := g.Exile.Size()
	if err := b22TryModal(g, jeskasWillOracle, []int{0, 1}, jeskasWillBoth(opp.ID)); err != nil {
		t.Fatalf("with your commander out, both is legal: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := redIn(me); got != want {
		t.Errorf("red mana = %d, want %d (one per card in the opponent's hand)", got, want)
	}
	if got := g.Exile.Size() - exiled; got != 3 {
		t.Errorf("exiled %d, want 3", got)
	}
}

// "As you cast this spell" is a check, not a duration: a commander
// that leaves after the announcement does not take a bullet with it.
func TestJeskasWillCommanderLeavingAfterAnnounceKeepsBoth(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	cmdr := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	want := opp.Hand.Size()
	exiled := g.Exile.Size()
	if err := b22TryModal(g, jeskasWillOracle, []int{0, 1}, jeskasWillBoth(opp.ID)); err != nil {
		t.Fatalf("announce with the commander out: %v", err)
	}
	// The commander leaves with the spell on the stack.
	gone, err := g.Battlefield.Remove(cmdr)
	if err != nil {
		t.Fatalf("remove the commander: %v", err)
	}
	me.Graveyard.PushTop(gone)
	passPriorityAroundTable(t, g)
	if got := redIn(me); got != want {
		t.Errorf("red mana = %d, want %d — the ritual still resolves", got, want)
	}
	if got := g.Exile.Size() - exiled; got != 3 {
		t.Errorf("exiled %d, want 3 — the impulse still resolves", got)
	}
}

// Only a commander YOU control counts. An opponent's commander on
// their own side does not; one you have stolen does (the Commander
// Legends ruling: any commander you control).
func TestJeskasWillCountsOnlyACommanderYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	theirs := b21Commander(g, opp.ID, "Their Commander", "Legendary Creature — Elf", 2, 2)
	if err := b22TryModal(g, jeskasWillOracle, []int{0, 1}, jeskasWillBoth(opp.ID)); err == nil {
		t.Fatal("an opponent's commander does not let you choose both")
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == theirs {
			g.Battlefield.Cards[i].Controller = me.ID
		}
	}
	if err := b22TryModal(g, jeskasWillOracle, []int{0, 1}, jeskasWillBoth(opp.ID)); err != nil {
		t.Fatalf("a commander you control — even one you stole — lets you choose both: %v", err)
	}
}

// Both bullets resolve in PRINTED order (CR 608.2c) even when the
// caster clicked the second one first: Will of the Mardu makes its
// Warriors before it counts your creatures for the damage. The
// commander (1) plus one Warrior (target player is you, controlling
// one creature) is 2 damage, which kills a 2/2; the other order would
// deal 1.
func TestWillOfTheMarduBothResolvesInPrintedOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 1, 1)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// Announced damage-first: occurrence 0 is option 1 (the creature),
	// occurrence 1 is option 0 (the player).
	if err := b22TryModal(g, b21WillOfTheMarduOracle, []int{1, 0}, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Mode: 0},
		{Kind: game.TargetPlayer, ID: me.ID, Mode: 1},
	}); err != nil {
		t.Fatalf("both with a commander: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b14CreaturesControlled(g, me.ID); n != 2 {
		t.Errorf("you control %d creatures, want 2 (commander + one Warrior)", n)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the Warriors come first, so the damage counts them: 2 damage kills the 2/2")
	}
}

// The rest of the Will family on this seam — one acceptance each with
// a commander, one refusal without.
func TestWillCycleChoosesBothOnlyWithACommander(t *testing.T) {
	t.Run("Akroma's Will", func(t *testing.T) {
		g := newCatalogGame(t)
		me, _ := cmcSeats(t, g)
		if err := b22TryModal(g, akromasWillOracle, []int{0, 1}, nil); err == nil {
			t.Fatal("no commander: both refused")
		}
		cmdr := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
		if err := b22TryModal(g, akromasWillOracle, []int{0, 1}, nil); err != nil {
			t.Fatalf("with a commander: %v", err)
		}
		passPriorityAroundTable(t, g)
		for _, kw := range []string{"flying", "double strike", "lifelink", "indestructible"} {
			if !hasEffectiveKeyword(t, g, cmdr, kw) {
				t.Errorf("both bullets resolved, so the commander has %s", kw)
			}
		}
	})
	t.Run("Drown in Dreams", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := cmcSeats(t, g)
		for i := 0; i < 10; i++ {
			opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
		}
		try := func() error {
			id := handCardFull(me, "Drown in Dreams", "Instant", "{X}{2}{U}", b19DrownInDreamsOracle, nil)
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, XValue: 1, Targets: []game.TargetRef{
				{Kind: game.TargetPlayer, ID: me.ID, Mode: 0},
				{Kind: game.TargetPlayer, ID: opp.ID, Mode: 1},
			}})
			if err != nil {
				_, _ = me.Hand.Remove(id)
			}
			return err
		}
		if err := try(); err == nil {
			t.Fatal("no commander: both refused")
		}
		b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
		hand, lib := me.Hand.Size(), opp.Library.Size()
		if err := try(); err != nil {
			t.Fatalf("with a commander: %v", err)
		}
		passPriorityAroundTable(t, g)
		// try() seeds the spell and casts it (net zero); X=1 draws one.
		if me.Hand.Size() != hand+1 {
			t.Errorf("hand %d, want %d — the draw bullet reads its OWN target, you", me.Hand.Size(), hand+1)
		}
		if opp.Library.Size() != lib-2 {
			t.Errorf("opponent milled %d, want 2 — the mill bullet reads its OWN target", lib-opp.Library.Size())
		}
	})
	t.Run("Will of the Abzan", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := cmcSeats(t, g)
		dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
		targets := []game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0},
			{Kind: game.TargetCard, ID: dead, Mode: 1},
		}
		if err := b22TryModal(g, b24WillOfTheAbzanOracle, []int{0, 1}, targets); err == nil {
			t.Fatal("no commander: both refused")
		}
		b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
		life := opp.Life
		if err := b22TryModal(g, b24WillOfTheAbzanOracle, []int{0, 1}, targets); err != nil {
			t.Fatalf("with a commander: %v", err)
		}
		passPriorityAroundTable(t, g)
		if opp.Life != life-3 {
			t.Errorf("the edict bullet: opponent at %d, want %d", opp.Life, life-3)
		}
		if !g.Battlefield.Contains(dead) {
			t.Error("the reanimation bullet returns the creature card")
		}
	})
}

// Will of the Sultai, new with #1590. Both bullets with a commander,
// targeting yourself for the mill: the lands come back (tapped) BEFORE
// the counters count them — two on the battlefield, one already in the
// graveyard and one milled make four.
func TestWillOfTheSultaiBothReturnsLandsThenCountsThem(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	inYard := b17GraveyardCard(me, "Dead Forest", "Basic Land — Forest", "")
	milled := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milled, Name: "Milled Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	targets := []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Mode: 0},
		{Kind: game.TargetCard, ID: bear, Mode: 1},
	}
	if err := b22TryModal(g, willOfTheSultaiOracle, []int{0, 1}, targets); err != nil {
		t.Fatalf("both with a commander: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{inYard, milled} {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatalf("land %s did not return", id)
		}
		if !c.Tapped {
			t.Errorf("%s returns tapped", c.Name)
		}
	}
	if got := b18Counters(t, g, bear, "+1/+1"); got != 4 {
		t.Errorf("+1/+1 counters = %d, want 4 — the returned lands are counted", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the creature gains trample")
	}
}

func TestWillOfTheSultaiOneBulletWithoutACommander(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	both := []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Mode: 0},
		{Kind: game.TargetCard, ID: bear, Mode: 1},
	}
	if err := b22TryModal(g, willOfTheSultaiOracle, []int{0, 1}, both); err == nil {
		t.Fatal("no commander: both refused")
	}
	if err := b22TryModal(g, willOfTheSultaiOracle, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
		t.Fatalf("the counters bullet alone: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, bear, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 (one land)", got)
	}
}

// Flame of Anor's condition is a Wizard, not a commander — the same
// shape with a different predicate — and it is the caster's Wizard.
func TestFlameOfAnorChoosesTwoOnlyWithYourWizard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	rock := b12Permanent(g, opp.ID, "Mana Rock", "Artifact")
	// A 6/6 survives the 5 damage and would not survive a destroy, so
	// it tells the two bullets' target groups apart.
	bear := b16Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6, "G")
	try := func() error {
		id := handCardFull(me, "Flame of Anor", "Instant", "", b16FlameOfAnorOracle, []string{"U", "R"})
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1, 2}, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: rock, Mode: 0},
			{Kind: game.TargetCard, ID: bear, Mode: 1},
		}})
		if err != nil {
			_, _ = me.Hand.Remove(id)
		}
		return err
	}
	if err := try(); err == nil {
		t.Fatal("no Wizard: choosing two is refused")
	}
	b12Creature(g, opp.ID, "Their Wizard", "Creature — Human Wizard", 1, 1)
	if err := try(); err == nil {
		t.Fatal("an opponent's Wizard does not count")
	}
	b12Creature(g, me.ID, "My Wizard", "Creature — Human Wizard", 1, 1)
	if err := try(); err != nil {
		t.Fatalf("with your Wizard, two is legal: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact bullet destroys the rock")
	}
	giant, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("the destroy bullet hit the damage bullet's target — each bullet reads its OWN group")
	}
	if giant.DamageMarked != 5 {
		t.Errorf("the giant has %d damage marked, want 5", giant.DamageMarked)
	}
}

// Register refuses a half-declared or non-raising conditional count.
func TestRegisterRefusesABadConditionalModeCount(t *testing.T) {
	cond := YouControlACommander
	for name, ms := range map[string]*game.ModeSpec{
		"no predicate":      {Options: []game.ModeOption{{Label: "a"}, {Label: "b"}}, Min: 1, Max: 1, RaisedMax: 2},
		"no raised max":     {Options: []game.ModeOption{{Label: "a"}, {Label: "b"}}, Min: 1, Max: 1, RaiseMaxIf: cond},
		"does not raise":    ChooseN("Choose two", 2, 2, Mode("a"), Mode("b")).OrUpToIf(2, cond),
		"more than printed": ChooseOne(Mode("a"), Mode("b")).OrUpToIf(3, cond),
		"unbounded max":     {Options: []game.ModeOption{{Label: "a"}, {Label: "b"}}, Min: 1, Max: 0, RaisedMax: 2, RaiseMaxIf: cond},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("Register must panic")
				}
			}()
			checkRaisedModeMax("Test Card", "Modes", ms)
		})
	}
}
