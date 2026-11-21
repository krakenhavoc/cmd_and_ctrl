package testso

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nova Pentacle: the controller names targets for nothing, picks which
// opponent chooses, and that opponent picks any creature — one the
// controller controls included. Nothing is paid until they have; the
// redirection then lands on the opponent's pick.
func TestNovaPentacleOpponentChoosesTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	pentacle := effectstest.PushCatalogPermanent(g, me.ID, "Nova Pentacle", "Artifact", effectstest.OtNovaPentacle, false)
	mine := effectstest.Pr7Creature(g, me.ID, "My Bear", 2, "G")
	theirs := effectstest.Pr7Creature(g, a.ID, "Their Bear", 2, "G")
	src := effectstest.Pr7Creature(g, b.ID, "Pinger", 1, "R")

	// The controller may not name the opponent's target themselves.
	if err := g.ActivateCatalogAbility(me.ID, pentacle, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.OtCard(mine)}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a payload naming the opponent's target: %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, pentacle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.FindBattlefieldCardForTest(g, pentacle).Tapped {
		t.Fatal("the Pentacle was tapped before the target was chosen (CR 601.2c precedes 601.2h)")
	}
	// Every opponent could choose a creature, so the controller is asked.
	if got := effectstest.OtOfferedSeats(effectstest.OtOptionPick(t, g, me.ID)); len(got) != 3 {
		t.Fatalf("offered %d opponents, want 3", len(got))
	}
	effectstest.OtChooseOpponent(t, g, me.ID, a.ID)
	p := effectstest.OtPickFor(t, g, a.ID, me.ID)
	for _, want := range []uuid.UUID{mine, theirs, src} {
		if !slices.Contains(p.PickTargetCards, want) {
			t.Fatalf("the opponent's pick does not offer %s: any creature is a legal target", want)
		}
	}
	effectstest.OtAnswer(t, g, a.ID, effectstest.OtCard(mine))
	item := effectstest.OtItemFrom(g, pentacle)
	if item == nil {
		t.Fatal("no ability on the stack after the opponent chose")
	}
	if len(item.TargetChoosers) != 1 || item.TargetChoosers[0].Player != a.ID {
		t.Fatalf("TargetChoosers %+v, want the chosen opponent", item.TargetChoosers)
	}
	if !effectstest.FindBattlefieldCardForTest(g, pentacle).Tapped {
		t.Fatal("the cost was not paid once the announcement completed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life || effectstest.Pr6Marked(g, mine) != 2 {
		t.Fatalf("life %d (want %d), the opponent's pick took %d (want 2)", me.Life, life, effectstest.Pr6Marked(g, mine))
	}
}

// The controller still targets: an opponent cannot name their own
// hexproof creature (the Cuombajj Witches ruling), and with one opponent
// able to choose, nobody is asked which.
func TestOpponentsChoiceIsJudgedForTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	pentacle := effectstest.PushCatalogPermanent(g, me.ID, "Nova Pentacle", "Artifact", effectstest.OtNovaPentacle, false)
	shrouded := effectstest.ApaPush(g, a.ID, a.ID, game.Card{Name: "Hexproof Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Keywords: []string{"hexproof"}})
	plain := effectstest.Pr7Creature(g, me.ID, "Plain Bear", 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, pentacle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// Any opponent could pick the plain bear; ask.
	effectstest.OtChooseOpponent(t, g, me.ID, a.ID)
	p := effectstest.OtPickFor(t, g, a.ID, me.ID)
	if slices.Contains(p.PickTargetCards, shrouded) {
		t.Fatal("the opponent was offered their own hexproof creature: the controller is the one targeting")
	}
	if !slices.Contains(p.PickTargetCards, plain) {
		t.Fatal("the plain creature is not offered")
	}
	if err := g.ResolvePickTargets(p.ID, a.ID, []game.TargetRef{effectstest.OtCard(shrouded)}); err == nil {
		t.Fatal("the hexproof creature was accepted as the opponent's pick")
	}
}

// Evangelize: only an opponent who controls a creature can choose "a
// creature they control", so with one such opponent nobody is asked
// which; the pick offers only their creatures, and the caster gains
// control. A creature that has changed hands by resolution is no longer
// "they control", and the spell does nothing (the ruling).
func TestEvangelizeTheyControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.Pr7Creature(g, a.ID, "Their Bear", 2, "G")
	mine := effectstest.Pr7Creature(g, me.ID, "My Bear", 2, "G")
	effectstest.CastCatalogSpell(t, g, "Evangelize", "Sorcery", effectstest.OtEvangelize, nil)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick {
			t.Fatal("the caster was asked which opponent with only one able to choose")
		}
	}
	p := effectstest.OtPickFor(t, g, a.ID, me.ID)
	if !slices.Equal(p.PickTargetCards, []uuid.UUID{theirs}) {
		t.Fatalf("offered %v, want only the opponent's own creature", p.PickTargetCards)
	}
	if err := g.ResolvePickTargets(p.ID, a.ID, []game.TargetRef{effectstest.OtCard(mine)}); err == nil {
		t.Fatal("the opponent named a creature they don't control")
	}
	effectstest.OtAnswer(t, g, a.ID, effectstest.OtCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, theirs); c == nil || c.Controller != me.ID {
		t.Fatal("the caster did not gain control of the opponent's pick")
	}

	// Again, and this time the pick changes hands in response.
	other := effectstest.Pr7Creature(g, a.ID, "Second Bear", 2, "G")
	effectstest.CastCatalogSpell(t, g, "Evangelize", "Sorcery", effectstest.OtEvangelize, nil)
	effectstest.OtAnswer(t, g, a.ID, effectstest.OtCard(other))
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), other, b.ID, game.IndefiniteDuration(), "test theft")
		g.RecomputeLayersIfStaleLocked()
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, other); c == nil || c.Controller != b.ID {
		t.Fatal("Evangelize took a creature the chosen opponent no longer controls")
	}
}

// Arena: the controller's creature first, then the opponent of their
// choice picks a creature they control; both are tapped and fight.
func TestArenaTapsAndFights(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
	}{
		{"Arena", "Land", effectstest.OtArena},
		{"Magus of the Arena", "Creature — Human Wizard", effectstest.OtMagusOfTheArena},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
			arena := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Test", Power: 5, Toughness: 5})
			theirs := effectstest.ApaPush(g, a.ID, a.ID, game.Card{Name: "Small", TypeLine: "Creature — Test", Power: 2, Toughness: 2})
			bs := effectstest.ApaPush(g, b.ID, b.ID, game.Card{Name: "Other", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
			if err := g.ActivateCatalogAbility(me.ID, arena, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.OtCard(mine)}}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			// Two opponents control a creature; the third does not and is
			// not offered.
			offered := effectstest.OtOfferedSeats(effectstest.OtOptionPick(t, g, me.ID))
			if len(offered) != 2 || !slices.Contains(offered, a.ID) || !slices.Contains(offered, b.ID) {
				t.Fatalf("offered %v, want the two opponents who control a creature", offered)
			}
			effectstest.OtChooseOpponent(t, g, me.ID, a.ID)
			p := effectstest.OtPickFor(t, g, a.ID, me.ID)
			if !slices.Equal(p.PickTargetCards, []uuid.UUID{theirs}) {
				t.Fatalf("offered %v, want the chosen opponent's creature only (not %s)", p.PickTargetCards, bs)
			}
			effectstest.OtAnswer(t, g, a.ID, effectstest.OtCard(theirs))
			effectstest.PassPriorityAroundTable(t, g)
			if c := effectstest.FindBattlefieldCardForTest(g, mine); c == nil || !c.Tapped || c.DamageMarked != 2 {
				t.Fatalf("the controller's creature: %+v, want tapped with 2 damage", c)
			}
			if effectstest.FindBattlefieldCardForTest(g, theirs) != nil {
				t.Fatal("the opponent's 2/2 survived a fight with a 5/5")
			}
		})
	}
}

// Cuombajj Witches: 1 to the controller's target and 1 to the target of
// an opponent's choice — the opponent may pick the controller.
func TestCuombajjWitchesTwoPings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	witches := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Cuombajj Witches", effectstest.OtCuombajjWitches, 1, 3))
	if err := g.ActivateCatalogAbility(me.ID, witches, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.OtPlayer(a.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.OtChooseOpponent(t, g, me.ID, b.ID)
	effectstest.OtPickFor(t, g, b.ID, me.ID)
	meLife, aLife := me.Life, a.Life
	effectstest.OtAnswer(t, g, b.ID, effectstest.OtPlayer(me.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != meLife-1 || a.Life != aLife-1 {
		t.Fatalf("life: controller %d (want %d), first target %d (want %d)", me.Life, meLife-1, a.Life, aLife-1)
	}
}

// Volcanic Offering: two clauses an opponent chooses, asked one at a
// time, each with its own opponent; the spell is cast only after both.
func TestVolcanicOfferingTwoOpponentsChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	nonbasic := func(owner uuid.UUID, name string) uuid.UUID {
		return effectstest.ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Land"})
	}
	landA, landB := nonbasic(a.ID, "A's Land"), nonbasic(b.ID, "B's Land")
	bigA := effectstest.ApaPush(g, a.ID, a.ID, game.Card{Name: "A's Giant", TypeLine: "Creature — Giant", Power: 6, Toughness: 6})
	bigB := effectstest.ApaPush(g, b.ID, b.ID, game.Card{Name: "B's Giant", TypeLine: "Creature — Giant", Power: 6, Toughness: 6})
	mine := effectstest.Pr7Creature(g, me.ID, "My Bear", 2, "G")
	spell := effectstest.CastCatalogSpell(t, g, "Volcanic Offering", "Instant", effectstest.OtVolcanicOffering, []game.TargetRef{
		{Kind: game.TargetCard, ID: landA, Slot: 0},
		{Kind: game.TargetCard, ID: bigA, Slot: 2},
	})
	effectstest.OtChooseOpponent(t, g, me.ID, b.ID)
	p := effectstest.OtPickFor(t, g, b.ID, me.ID)
	if slices.Contains(p.PickTargetCards, mine) {
		t.Fatal("a land clause offered a creature")
	}
	effectstest.OtAnswer(t, g, b.ID, effectstest.OtCard(landB))
	if effectstest.OtItemFrom(g, spell) != nil {
		t.Fatal("the spell was cast before its last clause was chosen")
	}
	effectstest.OtChooseOpponent(t, g, me.ID, a.ID)
	p = effectstest.OtPickFor(t, g, a.ID, me.ID)
	if slices.Contains(p.PickTargetCards, mine) {
		t.Fatal("\"you don't control\" is the caster's, and the caster's creature was offered")
	}
	effectstest.OtAnswer(t, g, a.ID, effectstest.OtCard(bigB))
	item := effectstest.OtItemFrom(g, spell)
	if item == nil || len(item.Targets) != 4 || len(item.TargetChoosers) != 2 {
		t.Fatalf("item %+v, want four targets and two choosers", item)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{landA, landB, bigA, bigB} {
		if effectstest.FindBattlefieldCardForTest(g, id) != nil {
			t.Fatalf("%s survived Volcanic Offering", id)
		}
	}
	if effectstest.FindBattlefieldCardForTest(g, mine) == nil {
		t.Fatal("the caster's creature was hit")
	}
}

// Mausoleum Turnkey: the trigger's controller names an opponent, who
// picks the creature card that comes back from the controller's
// graveyard.
func TestMausoleumTurnkeyOpponentPicksTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	weak, strong := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: weak, Name: "Weak", TypeLine: "Creature — Test", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: strong, Name: "Strong", TypeLine: "Creature — Test", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Mausoleum Turnkey", "Creature — Ogre Rogue", effectstest.OtMausoleumTurnkey, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.OtChooseOpponent(t, g, me.ID, a.ID)
	p := effectstest.OtPickFor(t, g, a.ID, me.ID)
	if !slices.Contains(p.PickTargetCards, weak) || !slices.Contains(p.PickTargetCards, strong) {
		t.Fatalf("offered %v, want both creature cards", p.PickTargetCards)
	}
	effectstest.OtAnswer(t, g, a.ID, effectstest.OtCard(weak))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, weak) || effectstest.InHandOf(me, strong) {
		t.Fatal("the controller's hand does not hold exactly the opponent's pick")
	}
}

// A controller who leaves takes the parked announcement with them: the
// opponent's open pick is withdrawn and nothing reaches the stack (CR
// 800.4a). A second announcement of the same ability while one is
// parked is refused.
func TestOpponentsPickIsWithdrawnWhenTheControllerLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	arena := effectstest.PushCatalogPermanent(g, me.ID, "Arena", "Land", effectstest.OtArena, false)
	mine := effectstest.Pr7Creature(g, me.ID, "Mine", 3, "G")
	effectstest.Pr7Creature(g, a.ID, "Theirs", 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, arena, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.OtCard(mine)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.OtPickFor(t, g, a.ID, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, arena, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.OtCard(mine)}}); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a second announcement while one is parked: %v, want ErrChoicePending", err)
	}
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if p := effectstest.LatestPickTarget(g, a.ID); p != nil {
		t.Fatal("the opponent still owes a pick for a player who has left")
	}
	if effectstest.OtItemFrom(g, arena) != nil {
		t.Fatal("the departed player's ability reached the stack")
	}
}

// The opponent's clause is DATA on the catalog: the clause the client and
// the bot skip carries the flag and its printed words.
func TestOpponentChoiceClausesAreDeclared(t *testing.T) {
	for _, oracle := range []string{effectstest.OtNovaPentacle, effectstest.OtEvangelize, effectstest.OtArena, effectstest.OtMagusOfTheArena, effectstest.OtCuombajjWitches, effectstest.OtVolcanicOffering, effectstest.OtMausoleumTurnkey} {
		spec, ok := effects.Lookup(oracle)
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
				if !containsText(c.Label, effectstest.OtOpponentChoiceTag) {
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
