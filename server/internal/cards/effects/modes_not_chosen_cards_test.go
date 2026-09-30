package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modes_not_chosen_cards_test.go — ADR 0097 PR 2 (#1749): the cards
// whose modal clause says "choose one that hasn't been chosen THIS
// TURN". Each test plays the card through the engine's own verbs and
// checks three things: a used bullet is not offered again, every
// bullet does what it prints, and a case the card does not apply to
// asks nothing.

// chooseModeNow answers the newest mode_pick owed by `chooser` after
// checking what it offers.
func chooseModeNow(t *testing.T, g *game.Game, chooser uuid.UUID, wantOffered []int, mode int) {
	t.Helper()
	c := modePickChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no mode_pick for %s; want one offering %v", chooser, wantOffered)
	}
	if !slices.Equal(c.ModeOptionIndex, wantOffered) {
		t.Fatalf("mode_pick offers %v, want %v (used %v)", c.ModeOptionIndex, wantOffered, c.ModeUsedIndex)
	}
	answerMode(t, g, c, chooser, mode)
}

// noModePick fails if `chooser` owes a mode_pick.
func noModePick(t *testing.T, g *game.Game, chooser uuid.UUID, why string) {
	t.Helper()
	if c := modePickChoiceFor(g, chooser); c != nil {
		t.Fatalf("%s: unexpected mode_pick offering %v", why, c.ModeOptionIndex)
	}
}

// advanceUntilModePick walks the turn on until `chooser` owes a
// mode_pick.
func advanceUntilModePick(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 40 && modePickChoiceFor(g, chooser) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if modePickChoiceFor(g, chooser) == nil {
		t.Fatal("no mode_pick arrived")
	}
}

func handSize(p *game.Player) int { return p.Hand.Size() }

func poolCount(p *game.Player, color string) int {
	n := 0
	for _, m := range p.ManaPool {
		if m.Color == color {
			n++
		}
	}
	return n
}

func effectiveOf(t *testing.T, g *game.Game, id uuid.UUID) game.Characteristic {
	t.Helper()
	var ch game.Characteristic
	var ok bool
	g.ReadSnapshot(func() {
		var c game.Card
		c, ok = battlefieldCard(g, id)
		if ok {
			ch = c.Effective()
		}
	})
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return ch
}

func zoneOf(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// --- Breeches, Eager Pillager ---------------------------------------

func TestBreechesEagerPillagerGivesEachPirateADifferentBullet(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	breeches := b12Push(g, me.ID, "Breeches, Eager Pillager", "Legendary Creature — Goblin Pirate",
		"15361770-c6dc-4db7-b745-aa03e4b866ea", 3, 3)
	p1 := b12Creature(g, me.ID, "Deck Hand", "Creature — Human Pirate", 1, 1)
	p2 := b12Creature(g, me.ID, "Powder Monkey", "Creature — Goblin Pirate", 1, 1)
	p3 := b12Creature(g, me.ID, "Bosun", "Creature — Orc Pirate", 1, 1)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Wall of Wood", "Creature — Wall", 0, 3)
	topCard := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: topCard, Name: "Impulse Target", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})

	// Five attackers, four of them Pirates: four triggers, not five.
	declareAttack(t, g, opp.ID, breeches, p1, p2, p3, bear)
	if n := len(modePickChoicesFor(g, me.ID)); n != 4 {
		t.Fatalf("four Pirates attacked, want four mode_picks, got %d", n)
	}

	// The target bullet first, so its target prompt is answered at once.
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 1)
	pickCard(t, g, me.ID, blocker)
	c := modePickChoicesFor(g, me.ID)[0]
	if slices.Contains(c.ModeOptionIndex, 1) {
		t.Fatalf("the can't-block bullet is taken off the other prompts: %v", c.ModeOptionIndex)
	}
	answerMode(t, g, c, me.ID, 0)
	answerMode(t, g, modePickChoicesFor(g, me.ID)[0], me.ID, 2)
	// "Removed from the stack with no effect": the fourth Pirate's
	// instance has nothing left to choose.
	if n := len(modePickChoicesFor(g, me.ID)); n != 0 {
		t.Fatalf("the fourth trigger is withdrawn, got %d prompts", n)
	}
	passPriorityAroundTable(t, g)

	if n := b16CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("one Treasure, got %d", n)
	}
	if r := effectiveOf(t, g, blocker).Restrictions; r&game.CantBlock == 0 {
		t.Errorf("the targeted creature can't block this turn: restrictions %v", r)
	}
	if z := zoneOf(g, topCard); z != game.ZoneExile {
		t.Errorf("the top card of the library was exiled, got zone %q", z)
	}
}

// --- Parapet Thrasher -----------------------------------------------

func TestParapetThrasherUsesADifferentBulletInEachDamageStep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	victim := g.Seats[(seat+1)%4]
	other1, other2 := g.Seats[(seat+2)%4], g.Seats[(seat+3)%4]
	thrasher := b12Push(g, me.ID, "Parapet Thrasher", "Creature — Dragon", "722e23d5-4f2a-43cd-bbd5-baf56b76e68d", 4, 3)
	drake := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Swift Drake", TypeLine: "Creature — Dragon",
		Power: 1, Toughness: 1, Keywords: []string{"first strike"}, Owner: me.ID, Controller: me.ID,
	})
	victimsArtifact := b12Permanent(g, victim.ID, "Victim's Relic", "Artifact")
	othersArtifact := b12Permanent(g, other1.ID, "Other's Relic", "Artifact")
	lives := map[uuid.UUID]int{victim.ID: victim.Life, other1.ID: other1.Life, other2.ID: other2.Life}

	declareAttack(t, g, victim.ID, thrasher, drake)
	advanceUntilModePick(t, g, me.ID)
	// The first-strike step is its own batch: the Drake alone connects.
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	// The declared simplification: any opponent's artifact is offered,
	// and one the damaged opponent does not control is not destroyed.
	pickCard(t, g, me.ID, othersArtifact)
	passPriorityAroundTable(t, g)
	if zoneOf(g, othersArtifact) != game.ZoneBattlefield {
		t.Error("an artifact the damaged opponent doesn't control is not destroyed")
	}
	if zoneOf(g, victimsArtifact) != game.ZoneBattlefield {
		t.Error("the damaged opponent's artifact was never the target")
	}

	// The regular damage step: the artifact bullet is used this turn.
	advanceUntilModePick(t, g, me.ID)
	chooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	passPriorityAroundTable(t, g)
	if got := lives[victim.ID] - victim.Life; got != 5 {
		t.Errorf("the damaged opponent took only combat damage (1 + 4), got %d", got)
	}
	for _, p := range []*game.Player{other1, other2} {
		if got := lives[p.ID] - p.Life; got != 4 {
			t.Errorf("each OTHER opponent takes 4: %s lost %d", p.Name, got)
		}
	}
}

func TestParapetThrasherDestroysTheDamagedOpponentsArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := seatsForCombat(g)
	thrasher := b12Push(g, me.ID, "Parapet Thrasher", "Creature — Dragon", "722e23d5-4f2a-43cd-bbd5-baf56b76e68d", 4, 3)
	relic := b12Permanent(g, victim.ID, "Victim's Relic", "Artifact")
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	// A non-Dragon connecting alone triggers nothing.
	attackWith(t, g, victim.ID, bear)
	noModePick(t, g, me.ID, "a Bear is not a Dragon")

	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%4)
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+3)%4)
	attackWith(t, g, victim.ID, thrasher)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if z := zoneOf(g, relic); z != game.ZoneGraveyard {
		t.Errorf("the damaged opponent's artifact is destroyed, zone %q", z)
	}
}

// --- Galadriel, Light of Valinor -------------------------------------

func TestGaladrielLightOfValinorAllianceBullets(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	galadriel := b12Push(g, me.ID, "Galadriel, Light of Valinor", "Legendary Creature — Elf Noble",
		"faf44683-339d-4fc0-8029-1e867c69de0e", 3, 3)

	enterCreatureFor(t, g, opp.ID, 1)
	noModePick(t, g, me.ID, "an opponent's creature is not alliance")

	enterCreatureFor(t, g, me.ID, 1)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	passPriorityAroundTable(t, g)
	if n := poolCount(me, "G"); n != 3 {
		t.Errorf("Add {G}{G}{G}: %d green in the pool", n)
	}

	enterCreatureFor(t, g, me.ID, 1)
	chooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, galadriel, game.CounterPlusOne); n != 1 {
		t.Errorf("a +1/+1 counter on each creature you control, Galadriel has %d", n)
	}

	hand := handSize(me)
	enterCreatureFor(t, g, me.ID, 1)
	chooseModeNow(t, g, me.ID, []int{2}, 2)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 {
		t.Errorf("scry 2, then draw: hand %d → %d", hand, handSize(me))
	}

	enterCreatureFor(t, g, me.ID, 1)
	noModePick(t, g, me.ID, "all three bullets are used this turn")
}

// --- The Vision -------------------------------------------------------

func TestTheVisionNoncreatureSpellBullets(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vision := b12Push(g, me.ID, "The Vision", "Legendary Artifact Creature — Robot Hero",
		"32aee76a-e738-4898-985a-801d1cce15ea", 2, 5)

	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	noModePick(t, g, me.ID, "a creature spell does not trigger it")
	passPriorityAroundTable(t, g)

	hand := handSize(me)
	castCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 2)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 {
		t.Errorf("Technopathy draws a card: hand %d → %d", hand, handSize(me))
	}

	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	chooseModeNow(t, g, me.ID, []int{0, 1}, 0)
	passPriorityAroundTable(t, g)
	if !slices.Contains(effectiveOf(t, g, vision).Abilities, "double strike") {
		t.Error("Solar Beam: double strike until end of turn")
	}

	castCatalogSpell(t, g, "Test Instant 2", "Instant", "", nil)
	chooseModeNow(t, g, me.ID, []int{1}, 1)
	passPriorityAroundTable(t, g)
	if !slices.Contains(effectiveOf(t, g, vision).Abilities, "indestructible") {
		t.Error("Density Control: indestructible until end of turn")
	}

	castCatalogSpell(t, g, "Test Instant 3", "Instant", "", nil)
	noModePick(t, g, me.ID, "all three bullets used this turn")
}

// --- Lita, Little Orphan Amphibian -----------------------------------

func TestLitaAllianceBullets(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lita := b12Push(g, me.ID, "Lita, Little Orphan Amphibian", "Legendary Creature — Mutant Ninja Turtle",
		"212fdb7c-1c7e-4eed-b1ed-bcc14a425da8", 2, 1)

	enterCreatureFor(t, g, me.ID, 1)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, lita, game.CounterPlusOne); n != 1 {
		t.Errorf("a +1/+1 counter on Lita: %d", n)
	}

	enterCreatureFor(t, g, me.ID, 1)
	chooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Food"); n != 1 {
		t.Errorf("a Food token: %d", n)
	}

	enterCreatureFor(t, g, me.ID, 1)
	chooseModeNow(t, g, me.ID, []int{2}, 2)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("Scry 1 asks")
	}
	answerScryKeepAll(t, g, me.ID)

	enterCreatureFor(t, g, me.ID, 1)
	noModePick(t, g, me.ID, "all three bullets used this turn")
}

// --- Kargan Intimidator -----------------------------------------------

func TestKarganIntimidatorActivatedBulletsOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	kargan := pushCatalogPermanent(g, me.ID, "Kargan Intimidator", "Creature — Human Warrior",
		"045bf1fd-f375-47a0-9983-7e2ebfed7854", false)
	theirs := b12Creature(g, opp.ID, "Craven Knight", "Creature — Human Knight", 2, 2)

	b16Activate(t, g, me.ID, kargan, 0, game.ActivateAbilityParams{Modes: []int{0}})
	if p := effectiveOf(t, g, kargan).Power; p != 2 {
		t.Errorf("+1/+1 until end of turn on a 1/1 test body: power %d", p)
	}
	for _, modes := range activationModesOffered(g, me.ID, kargan) {
		if slices.Contains(modes, 0) {
			t.Errorf("the enumerator offered the used bullet: %v", modes)
		}
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, kargan, 0, game.ActivateAbilityParams{Modes: []int{0}}); err == nil {
		t.Fatal("a used bullet is refused")
	}
	if me.Life != life {
		t.Error("a refused activation changes nothing")
	}

	b16Activate(t, g, me.ID, kargan, 0, game.ActivateAbilityParams{Modes: []int{1}, Targets: b16TargetCard(theirs)})
	if !slices.Contains(effectiveOf(t, g, theirs).Subtypes, "Coward") {
		t.Errorf("becomes a Coward: subtypes %v", effectiveOf(t, g, theirs).Subtypes)
	}
	if !slices.Contains(effectiveOf(t, g, theirs).Subtypes, "Knight") {
		t.Error("in addition to its other creature types")
	}

	b16Activate(t, g, me.ID, kargan, 0, game.ActivateAbilityParams{Modes: []int{2}, Targets: b16TargetCard(kargan)})
	if !slices.Contains(effectiveOf(t, g, kargan).Abilities, "trample") {
		t.Error("the Warrior gains trample")
	}
	if got := activationModesOffered(g, me.ID, kargan); len(got) != 0 {
		t.Errorf("with all three used the ability is not offered: %v", got)
	}

	// Cowards can't block Warriors.
	var refusal bool
	g.ReadSnapshot(func() {
		k, _ := battlefieldCard(g, kargan)
		c, _ := battlefieldCard(g, theirs)
		refusal = !g.CanBlockLocked(&k, &c)
	})
	if !refusal {
		t.Error("a Coward can't block a Warrior")
	}
}

// --- Genku, Future Shaper --------------------------------------------

func TestGenkuFutureShaperMakesADifferentTokenForEachExit(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	b12Push(g, me.ID, "Genku, Future Shaper", "Legendary Creature — Moonfolk Wizard",
		"ca16b735-26c7-4d0b-959b-b031f6cb376e", 2, 5)
	destroy := func(id uuid.UUID) {
		t.Helper()
		g.WithWriteLock(func() {
			if err := (DestroyTarget{Target: id}).Apply(ctxFor(g, &game.StackItem{Controller: me.ID})); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
	}

	destroy(b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	noModePick(t, g, me.ID, "an opponent's permanent leaving")
	enterCreatureFor(t, g, me.ID, 1)
	destroy(findBattlefieldByName(g, "Soldier"))
	noModePick(t, g, me.ID, "a token leaving")

	destroy(b12Permanent(g, me.ID, "My Relic", "Artifact"))
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	passPriorityAroundTable(t, g)
	destroy(b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2))
	chooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	passPriorityAroundTable(t, g)
	destroy(b12Permanent(g, me.ID, "My Totem", "Enchantment"))
	chooseModeNow(t, g, me.ID, []int{2}, 2)
	passPriorityAroundTable(t, g)
	for _, name := range []string{"Fox", "Moonfolk", "Rat"} {
		if n := b16CountNamed(g, name); n != 1 {
			t.Errorf("one %s token, got %d", name, n)
		}
	}
	destroy(b12Permanent(g, me.ID, "My Other Relic", "Artifact"))
	noModePick(t, g, me.ID, "all three bullets used this turn")
}

// --- The Fantastic Four ----------------------------------------------

func TestTheFantasticFourEntersAndFoursShareOneMemory(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	cast := func(c game.Card) {
		t.Helper()
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Hand.PushTop(c)
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
			t.Fatalf("CastSpell %s: %v", c.Name, err)
		}
	}

	cast(game.Card{Name: "The Fantastic Four", TypeLine: "Legendary Creature — Human Hero",
		ManaCost: "{R}{G}{W}{U}", Power: 4, Toughness: 4, OracleID: "a2281c99-3f45-441b-8e78-7f1f29bf1dcd"})
	noModePick(t, g, me.ID, "casting it does not trigger it")
	passPriorityAroundTable(t, g)
	hand := handSize(me)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 3)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 {
		t.Errorf("the entry draws a card: %d → %d", hand, handSize(me))
	}

	cast(game.Card{Name: "Three Drop", TypeLine: "Instant", ManaCost: "{1}{U}{U}"})
	noModePick(t, g, me.ID, "mana value 3, no power")
	passPriorityAroundTable(t, g)

	cast(game.Card{Name: "Four Drop", TypeLine: "Sorcery", ManaCost: "{2}{U}{U}"})
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Wall"); n != 1 {
		t.Errorf("a 0/4 Wall: %d", n)
	}

	lives := make([]int, 4)
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	cast(game.Card{Name: "Big Hitter", TypeLine: "Creature — Ogre", ManaCost: "{R}", Power: 4, Toughness: 1})
	chooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i] - 3
		if p.ID == me.ID {
			want = lives[i]
		}
		if p.Life != want {
			t.Errorf("3 damage to each opponent: %s %d → %d", p.Name, lives[i], p.Life)
		}
	}

	wall := findBattlefieldByName(g, "Wall")
	cast(game.Card{Name: "Sturdy", TypeLine: "Creature — Wall", ManaCost: "{W}", Power: 0, Toughness: 4})
	chooseModeNow(t, g, me.ID, []int{2}, 2)
	pickCard(t, g, me.ID, wall)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, wall, game.CounterPlusOne); n != 2 {
		t.Errorf("two +1/+1 counters on the target: %d", n)
	}

	cast(game.Card{Name: "Another Four", TypeLine: "Instant", ManaCost: "{3}{R}"})
	noModePick(t, g, me.ID, "all four bullets used this turn")
}
