package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_target_cards_test.go — #2055, ADR 0019's 2026-10-16
// amendment: "target … of an opponent's choice". Each card is announced
// for real: the controller's payload names its own targets only, the
// controller names an opponent when more than one could choose, that
// opponent answers a pick_target, and only then is anything paid or put
// on the stack.

const (
	otNovaPentacle      = "db121503-8a34-498a-829c-72c33798369b"
	otEvangelize        = "23a60803-6b0f-48ee-aa3b-12a3c5fc62cf"
	otArena             = "12326d23-1fb5-4084-a23a-d6e2c90ea36c"
	otMagusOfTheArena   = "44865261-16f8-42d2-a388-a57173142eb0"
	otCuombajjWitches   = "638eeb16-9e0e-4cc6-b97e-8ff0df81ca58"
	otVolcanicOffering  = "05e0a10d-532a-451c-a54e-27e64969fe0e"
	otMausoleumTurnkey  = "d4a61685-7149-4d14-bc17-85b060945687"
	otOpponentChoiceTag = "of an opponent's choice"
)

// otOptionPick is the open option pick addressed to `chooser`.
func otOptionPick(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == chooser {
			return c
		}
	}
	t.Fatalf("no option pick open for %s", chooser)
	return nil
}

// otOfferedSeats is the seats an option pick offers.
func otOfferedSeats(c *game.PendingChoice) []uuid.UUID {
	var out []uuid.UUID
	for _, o := range c.PickOptions {
		out = append(out, o.Player)
	}
	return out
}

// otChooseOpponent answers the controller's "which opponent" question.
func otChooseOpponent(t *testing.T, g *game.Game, chooser, opp uuid.UUID) {
	t.Helper()
	c := otOptionPick(t, g, chooser)
	for i, o := range c.PickOptions {
		if o.Player == opp {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("opponent %s is not offered (%v)", opp, otOfferedSeats(c))
}

// otPickFor is the open pick_target addressed to `chooser`, asserted to
// be an opponent's pick for `controller`.
func otPickFor(t *testing.T, g *game.Game, chooser, controller uuid.UUID) *game.PendingChoice {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target open for %s", chooser)
	}
	if p.PickTargetFor != controller {
		t.Fatalf("pick_target is for %s, want %s", p.PickTargetFor, controller)
	}
	return p
}

// otAnswer answers chooser's open pick_target with refs.
func otAnswer(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target open for %s", chooser)
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

func otCard(id uuid.UUID) game.TargetRef   { return game.TargetRef{Kind: game.TargetCard, ID: id} }
func otPlayer(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }

// otItemFrom is the stack item whose source is `source`.
func otItemFrom(g *game.Game, source uuid.UUID) *game.StackItem {
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == source {
			return it
		}
	}
	return nil
}

// Nova Pentacle: the controller names targets for nothing, picks which
// opponent chooses, and that opponent picks any creature — one the
// controller controls included. Nothing is paid until they have; the
// redirection then lands on the opponent's pick.
func TestNovaPentacleOpponentChoosesTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	pentacle := pushCatalogPermanent(g, me.ID, "Nova Pentacle", "Artifact", otNovaPentacle, false)
	mine := pr7Creature(g, me.ID, "My Bear", 2, "G")
	theirs := pr7Creature(g, a.ID, "Their Bear", 2, "G")
	src := pr7Creature(g, b.ID, "Pinger", 1, "R")

	// The controller may not name the opponent's target themselves.
	if err := g.ActivateCatalogAbility(me.ID, pentacle, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{otCard(mine)}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a payload naming the opponent's target: %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, pentacle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if findBattlefieldCardForTest(g, pentacle).Tapped {
		t.Fatal("the Pentacle was tapped before the target was chosen (CR 601.2c precedes 601.2h)")
	}
	// Every opponent could choose a creature, so the controller is asked.
	if got := otOfferedSeats(otOptionPick(t, g, me.ID)); len(got) != 3 {
		t.Fatalf("offered %d opponents, want 3", len(got))
	}
	otChooseOpponent(t, g, me.ID, a.ID)
	p := otPickFor(t, g, a.ID, me.ID)
	for _, want := range []uuid.UUID{mine, theirs, src} {
		if !slices.Contains(p.PickTargetCards, want) {
			t.Fatalf("the opponent's pick does not offer %s: any creature is a legal target", want)
		}
	}
	otAnswer(t, g, a.ID, otCard(mine))
	item := otItemFrom(g, pentacle)
	if item == nil {
		t.Fatal("no ability on the stack after the opponent chose")
	}
	if len(item.TargetChoosers) != 1 || item.TargetChoosers[0].Player != a.ID {
		t.Fatalf("TargetChoosers %+v, want the chosen opponent", item.TargetChoosers)
	}
	if !findBattlefieldCardForTest(g, pentacle).Tapped {
		t.Fatal("the cost was not paid once the announcement completed")
	}
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life || pr6Marked(g, mine) != 2 {
		t.Fatalf("life %d (want %d), the opponent's pick took %d (want 2)", me.Life, life, pr6Marked(g, mine))
	}
}

// The controller still targets: an opponent cannot name their own
// hexproof creature (the Cuombajj Witches ruling), and with one opponent
// able to choose, nobody is asked which.
func TestOpponentsChoiceIsJudgedForTheController(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	pentacle := pushCatalogPermanent(g, me.ID, "Nova Pentacle", "Artifact", otNovaPentacle, false)
	shrouded := apaPush(g, a.ID, a.ID, game.Card{Name: "Hexproof Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Keywords: []string{"hexproof"}})
	plain := pr7Creature(g, me.ID, "Plain Bear", 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, pentacle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// Any opponent could pick the plain bear; ask.
	otChooseOpponent(t, g, me.ID, a.ID)
	p := otPickFor(t, g, a.ID, me.ID)
	if slices.Contains(p.PickTargetCards, shrouded) {
		t.Fatal("the opponent was offered their own hexproof creature: the controller is the one targeting")
	}
	if !slices.Contains(p.PickTargetCards, plain) {
		t.Fatal("the plain creature is not offered")
	}
	if err := g.ResolvePickTargets(p.ID, a.ID, []game.TargetRef{otCard(shrouded)}); err == nil {
		t.Fatal("the hexproof creature was accepted as the opponent's pick")
	}
}

// Evangelize: only an opponent who controls a creature can choose "a
// creature they control", so with one such opponent nobody is asked
// which; the pick offers only their creatures, and the caster gains
// control. A creature that has changed hands by resolution is no longer
// "they control", and the spell does nothing (the ruling).
func TestEvangelizeTheyControl(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := pr7Creature(g, a.ID, "Their Bear", 2, "G")
	mine := pr7Creature(g, me.ID, "My Bear", 2, "G")
	castCatalogSpell(t, g, "Evangelize", "Sorcery", otEvangelize, nil)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick {
			t.Fatal("the caster was asked which opponent with only one able to choose")
		}
	}
	p := otPickFor(t, g, a.ID, me.ID)
	if !slices.Equal(p.PickTargetCards, []uuid.UUID{theirs}) {
		t.Fatalf("offered %v, want only the opponent's own creature", p.PickTargetCards)
	}
	if err := g.ResolvePickTargets(p.ID, a.ID, []game.TargetRef{otCard(mine)}); err == nil {
		t.Fatal("the opponent named a creature they don't control")
	}
	otAnswer(t, g, a.ID, otCard(theirs))
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, theirs); c == nil || c.Controller != me.ID {
		t.Fatal("the caster did not gain control of the opponent's pick")
	}

	// Again, and this time the pick changes hands in response.
	other := pr7Creature(g, a.ID, "Second Bear", 2, "G")
	castCatalogSpell(t, g, "Evangelize", "Sorcery", otEvangelize, nil)
	otAnswer(t, g, a.ID, otCard(other))
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), other, b.ID, game.IndefiniteDuration(), "test theft")
		g.RecomputeLayersIfStaleLocked()
	})
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, other); c == nil || c.Controller != b.ID {
		t.Fatal("Evangelize took a creature the chosen opponent no longer controls")
	}
}

// Arena: the controller's creature first, then the opponent of their
// choice picks a creature they control; both are tapped and fight.
func TestArenaTapsAndFights(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
	}{
		{"Arena", "Land", otArena},
		{"Magus of the Arena", "Creature — Human Wizard", otMagusOfTheArena},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
			arena := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			mine := apaPush(g, me.ID, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Test", Power: 5, Toughness: 5})
			theirs := apaPush(g, a.ID, a.ID, game.Card{Name: "Small", TypeLine: "Creature — Test", Power: 2, Toughness: 2})
			bs := apaPush(g, b.ID, b.ID, game.Card{Name: "Other", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
			if err := g.ActivateCatalogAbility(me.ID, arena, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{otCard(mine)}}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			// Two opponents control a creature; the third does not and is
			// not offered.
			offered := otOfferedSeats(otOptionPick(t, g, me.ID))
			if len(offered) != 2 || !slices.Contains(offered, a.ID) || !slices.Contains(offered, b.ID) {
				t.Fatalf("offered %v, want the two opponents who control a creature", offered)
			}
			otChooseOpponent(t, g, me.ID, a.ID)
			p := otPickFor(t, g, a.ID, me.ID)
			if !slices.Equal(p.PickTargetCards, []uuid.UUID{theirs}) {
				t.Fatalf("offered %v, want the chosen opponent's creature only (not %s)", p.PickTargetCards, bs)
			}
			otAnswer(t, g, a.ID, otCard(theirs))
			passPriorityAroundTable(t, g)
			if c := findBattlefieldCardForTest(g, mine); c == nil || !c.Tapped || c.DamageMarked != 2 {
				t.Fatalf("the controller's creature: %+v, want tapped with 2 damage", c)
			}
			if findBattlefieldCardForTest(g, theirs) != nil {
				t.Fatal("the opponent's 2/2 survived a fight with a 5/5")
			}
		})
	}
}

// Cuombajj Witches: 1 to the controller's target and 1 to the target of
// an opponent's choice — the opponent may pick the controller.
func TestCuombajjWitchesTwoPings(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	witches := apaPush(g, me.ID, me.ID, apaCreature("Cuombajj Witches", otCuombajjWitches, 1, 3))
	if err := g.ActivateCatalogAbility(me.ID, witches, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{otPlayer(a.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	otChooseOpponent(t, g, me.ID, b.ID)
	otPickFor(t, g, b.ID, me.ID)
	meLife, aLife := me.Life, a.Life
	otAnswer(t, g, b.ID, otPlayer(me.ID))
	passPriorityAroundTable(t, g)
	if me.Life != meLife-1 || a.Life != aLife-1 {
		t.Fatalf("life: controller %d (want %d), first target %d (want %d)", me.Life, meLife-1, a.Life, aLife-1)
	}
}

// Volcanic Offering: two clauses an opponent chooses, asked one at a
// time, each with its own opponent; the spell is cast only after both.
func TestVolcanicOfferingTwoOpponentsChoose(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	nonbasic := func(owner uuid.UUID, name string) uuid.UUID {
		return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Land"})
	}
	landA, landB := nonbasic(a.ID, "A's Land"), nonbasic(b.ID, "B's Land")
	bigA := apaPush(g, a.ID, a.ID, game.Card{Name: "A's Giant", TypeLine: "Creature — Giant", Power: 6, Toughness: 6})
	bigB := apaPush(g, b.ID, b.ID, game.Card{Name: "B's Giant", TypeLine: "Creature — Giant", Power: 6, Toughness: 6})
	mine := pr7Creature(g, me.ID, "My Bear", 2, "G")
	spell := castCatalogSpell(t, g, "Volcanic Offering", "Instant", otVolcanicOffering, []game.TargetRef{
		{Kind: game.TargetCard, ID: landA, Slot: 0},
		{Kind: game.TargetCard, ID: bigA, Slot: 2},
	})
	otChooseOpponent(t, g, me.ID, b.ID)
	p := otPickFor(t, g, b.ID, me.ID)
	if slices.Contains(p.PickTargetCards, mine) {
		t.Fatal("a land clause offered a creature")
	}
	otAnswer(t, g, b.ID, otCard(landB))
	if otItemFrom(g, spell) != nil {
		t.Fatal("the spell was cast before its last clause was chosen")
	}
	otChooseOpponent(t, g, me.ID, a.ID)
	p = otPickFor(t, g, a.ID, me.ID)
	if slices.Contains(p.PickTargetCards, mine) {
		t.Fatal("\"you don't control\" is the caster's, and the caster's creature was offered")
	}
	otAnswer(t, g, a.ID, otCard(bigB))
	item := otItemFrom(g, spell)
	if item == nil || len(item.Targets) != 4 || len(item.TargetChoosers) != 2 {
		t.Fatalf("item %+v, want four targets and two choosers", item)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{landA, landB, bigA, bigB} {
		if findBattlefieldCardForTest(g, id) != nil {
			t.Fatalf("%s survived Volcanic Offering", id)
		}
	}
	if findBattlefieldCardForTest(g, mine) == nil {
		t.Fatal("the caster's creature was hit")
	}
}

// Mausoleum Turnkey: the trigger's controller names an opponent, who
// picks the creature card that comes back from the controller's
// graveyard.
func TestMausoleumTurnkeyOpponentPicksTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	weak, strong := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: weak, Name: "Weak", TypeLine: "Creature — Test", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: strong, Name: "Strong", TypeLine: "Creature — Test", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Mausoleum Turnkey", "Creature — Ogre Rogue", otMausoleumTurnkey, nil)
	passPriorityAroundTable(t, g)
	otChooseOpponent(t, g, me.ID, a.ID)
	p := otPickFor(t, g, a.ID, me.ID)
	if !slices.Contains(p.PickTargetCards, weak) || !slices.Contains(p.PickTargetCards, strong) {
		t.Fatalf("offered %v, want both creature cards", p.PickTargetCards)
	}
	otAnswer(t, g, a.ID, otCard(weak))
	passPriorityAroundTable(t, g)
	if !inHandOf(me, weak) || inHandOf(me, strong) {
		t.Fatal("the controller's hand does not hold exactly the opponent's pick")
	}
}

// A controller who leaves takes the parked announcement with them: the
// opponent's open pick is withdrawn and nothing reaches the stack (CR
// 800.4a). A second announcement of the same ability while one is
// parked is refused.
func TestOpponentsPickIsWithdrawnWhenTheControllerLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	arena := pushCatalogPermanent(g, me.ID, "Arena", "Land", otArena, false)
	mine := pr7Creature(g, me.ID, "Mine", 3, "G")
	pr7Creature(g, a.ID, "Theirs", 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, arena, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{otCard(mine)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	otPickFor(t, g, a.ID, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, arena, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{otCard(mine)}}); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a second announcement while one is parked: %v, want ErrChoicePending", err)
	}
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if p := latestPickTarget(g, a.ID); p != nil {
		t.Fatal("the opponent still owes a pick for a player who has left")
	}
	if otItemFrom(g, arena) != nil {
		t.Fatal("the departed player's ability reached the stack")
	}
}

// The opponent's clause is DATA on the catalog: the clause the client and
// the bot skip carries the flag and its printed words.
func TestOpponentChoiceClausesAreDeclared(t *testing.T) {
	for _, oracle := range []string{otNovaPentacle, otEvangelize, otArena, otMagusOfTheArena, otCuombajjWitches, otVolcanicOffering, otMausoleumTurnkey} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		var clauses *game.TargetSpec
		switch {
		case spec.Targets != nil:
			clauses = spec.Targets
		case len(spec.Activated) > 0:
			clauses = spec.Activated[0].Targets
		case len(spec.Triggered) > 0:
			clauses = spec.Triggered[0].Targets
		}
		found := false
		for i := 0; i < clauses.ClauseCount(); i++ {
			c := clauses.Clause(i)
			if c.ChosenByOpponent {
				found = true
				if !containsText(c.Label, otOpponentChoiceTag) {
					t.Errorf("%s clause %d is an opponent's choice labelled %q", spec.Name, i, c.Label)
				}
			}
		}
		if !found {
			t.Errorf("%s declares no opponent's-choice clause", spec.Name)
		}
	}
}

func containsText(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
