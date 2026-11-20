package testsm

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Breeches, Eager Pillager ---------------------------------------

func TestBreechesEagerPillagerGivesEachPirateADifferentBullet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	breeches := effectstest.B12Push(g, me.ID, "Breeches, Eager Pillager", "Legendary Creature — Goblin Pirate",
		"15361770-c6dc-4db7-b745-aa03e4b866ea", 3, 3)
	p1 := effectstest.B12Creature(g, me.ID, "Deck Hand", "Creature — Human Pirate", 1, 1)
	p2 := effectstest.B12Creature(g, me.ID, "Powder Monkey", "Creature — Goblin Pirate", 1, 1)
	p3 := effectstest.B12Creature(g, me.ID, "Bosun", "Creature — Orc Pirate", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall of Wood", "Creature — Wall", 0, 3)
	topCard := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: topCard, Name: "Impulse Target", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})

	// Five attackers, four of them Pirates: four triggers, not five.
	effectstest.DeclareAttack(t, g, opp.ID, breeches, p1, p2, p3, bear)
	if n := len(effectstest.ModePickChoicesFor(g, me.ID)); n != 4 {
		t.Fatalf("four Pirates attacked, want four mode_picks, got %d", n)
	}

	// The target bullet first, so its target prompt is answered at once.
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 1)
	effectstest.PickCard(t, g, me.ID, blocker)
	c := effectstest.ModePickChoicesFor(g, me.ID)[0]
	if slices.Contains(c.ModeOptionIndex, 1) {
		t.Fatalf("the can't-block bullet is taken off the other prompts: %v", c.ModeOptionIndex)
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.AnswerMode(t, g, effectstest.ModePickChoicesFor(g, me.ID)[0], me.ID, 2)
	// "Removed from the stack with no effect": the fourth Pirate's
	// instance has nothing left to choose.
	if n := len(effectstest.ModePickChoicesFor(g, me.ID)); n != 0 {
		t.Fatalf("the fourth trigger is withdrawn, got %d prompts", n)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.B16CountNamed(g, "Treasure"); n != 1 {
		t.Errorf("one Treasure, got %d", n)
	}
	if r := effectstest.EffectiveOf(t, g, blocker).Restrictions; r&game.CantBlock == 0 {
		t.Errorf("the targeted creature can't block this turn: restrictions %v", r)
	}
	if z := effectstest.ZoneOf(g, topCard); z != game.ZoneExile {
		t.Errorf("the top card of the library was exiled, got zone %q", z)
	}
}

// --- Parapet Thrasher -----------------------------------------------

func TestParapetThrasherUsesADifferentBulletInEachDamageStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	victim := g.Seats[(seat+1)%4]
	other1, other2 := g.Seats[(seat+2)%4], g.Seats[(seat+3)%4]
	thrasher := effectstest.B12Push(g, me.ID, "Parapet Thrasher", "Creature — Dragon", "722e23d5-4f2a-43cd-bbd5-baf56b76e68d", 4, 3)
	drake := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Swift Drake", TypeLine: "Creature — Dragon",
		Power: 1, Toughness: 1, Keywords: []string{"first strike"}, Owner: me.ID, Controller: me.ID,
	})
	victimsArtifact := effectstest.B12Permanent(g, victim.ID, "Victim's Relic", "Artifact")
	othersArtifact := effectstest.B12Permanent(g, other1.ID, "Other's Relic", "Artifact")
	lives := map[uuid.UUID]int{victim.ID: victim.Life, other1.ID: other1.Life, other2.ID: other2.Life}

	effectstest.DeclareAttack(t, g, victim.ID, thrasher, drake)
	effectstest.AdvanceUntilModePick(t, g, me.ID)
	// The first-strike step is its own batch: the Drake alone connects.
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	// #3860: "that opponent controls" — only the damaged player's
	// artifacts are offered, and another opponent's is refused.
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt for the artifact bullet")
	}
	if !effectstest.HasID(p.PickTargetCards, victimsArtifact) || effectstest.HasID(p.PickTargetCards, othersArtifact) {
		t.Fatalf("only the damaged opponent's artifact is offered, got %v", p.PickTargetCards)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: othersArtifact}); err == nil {
		t.Fatal("an artifact another opponent controls is not a legal target")
	}
	effectstest.PickCard(t, g, me.ID, victimsArtifact)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, victimsArtifact) != game.ZoneGraveyard {
		t.Error("the damaged opponent's artifact is destroyed")
	}
	if effectstest.ZoneOf(g, othersArtifact) != game.ZoneBattlefield {
		t.Error("an artifact the damaged opponent doesn't control is untouched")
	}

	// The regular damage step: the artifact bullet is used this turn.
	effectstest.AdvanceUntilModePick(t, g, me.ID)
	effectstest.ChooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, victim := effectstest.SeatsForCombat(g)
	thrasher := effectstest.B12Push(g, me.ID, "Parapet Thrasher", "Creature — Dragon", "722e23d5-4f2a-43cd-bbd5-baf56b76e68d", 4, 3)
	relic := effectstest.B12Permanent(g, victim.ID, "Victim's Relic", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	// A non-Dragon connecting alone triggers nothing.
	effectstest.AttackWith(t, g, victim.ID, bear)
	effectstest.NoModePick(t, g, me.ID, "a Bear is not a Dragon")

	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%4)
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+3)%4)
	effectstest.AttackWith(t, g, victim.ID, thrasher)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOf(g, relic); z != game.ZoneGraveyard {
		t.Errorf("the damaged opponent's artifact is destroyed, zone %q", z)
	}
}

// #3860: a bullet whose clause has no legal target is not on offer (CR
// 603.3c). "That opponent controls" makes that mean the DAMAGED
// opponent's artifacts: a relic on another opponent's side of the table
// no longer keeps the destroy bullet alive.
func TestParapetThrasherOffersNoArtifactBulletWithoutThatOpponentsArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := effectstest.SeatsForCombat(g)
	thrasher := effectstest.B12Push(g, me.ID, "Parapet Thrasher", "Creature — Dragon", "722e23d5-4f2a-43cd-bbd5-baf56b76e68d", 4, 3)
	var bystander *game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID && s.ID != victim.ID {
			bystander = s
			break
		}
	}
	relic := effectstest.B12Permanent(g, bystander.ID, "Bystander's Relic", "Artifact")

	effectstest.AttackWith(t, g, victim.ID, thrasher)
	effectstest.ChooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOf(g, relic); z != game.ZoneBattlefield {
		t.Errorf("an artifact the damaged opponent doesn't control is untouched, zone %q", z)
	}
}

// --- Galadriel, Light of Valinor -------------------------------------

func TestGaladrielLightOfValinorAllianceBullets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	galadriel := effectstest.B12Push(g, me.ID, "Galadriel, Light of Valinor", "Legendary Creature — Elf Noble",
		"faf44683-339d-4fc0-8029-1e867c69de0e", 3, 3)

	enterCreatureFor(t, g, opp.ID, 1)
	effectstest.NoModePick(t, g, me.ID, "an opponent's creature is not alliance")

	enterCreatureFor(t, g, me.ID, 1)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PoolCount(me, "G"); n != 3 {
		t.Errorf("Add {G}{G}{G}: %d green in the pool", n)
	}

	enterCreatureFor(t, g, me.ID, 1)
	effectstest.ChooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, galadriel, game.CounterPlusOne); n != 1 {
		t.Errorf("a +1/+1 counter on each creature you control, Galadriel has %d", n)
	}

	hand := effectstest.HandSize(me)
	enterCreatureFor(t, g, me.ID, 1)
	effectstest.ChooseModeNow(t, g, me.ID, []int{2}, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+1 {
		t.Errorf("scry 2, then draw: hand %d → %d", hand, effectstest.HandSize(me))
	}

	enterCreatureFor(t, g, me.ID, 1)
	effectstest.NoModePick(t, g, me.ID, "all three bullets are used this turn")
}

// --- The Vision -------------------------------------------------------

func TestTheVisionNoncreatureSpellBullets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vision := effectstest.B12Push(g, me.ID, "The Vision", "Legendary Artifact Creature — Robot Hero",
		"32aee76a-e738-4898-985a-801d1cce15ea", 2, 5)

	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.NoModePick(t, g, me.ID, "a creature spell does not trigger it")
	effectstest.PassPriorityAroundTable(t, g)

	hand := effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+1 {
		t.Errorf("Technopathy draws a card: hand %d → %d", hand, effectstest.HandSize(me))
	}

	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !slices.Contains(effectstest.EffectiveOf(t, g, vision).Abilities, "double strike") {
		t.Error("Solar Beam: double strike until end of turn")
	}

	effectstest.CastCatalogSpell(t, g, "Test Instant 2", "Instant", "", nil)
	effectstest.ChooseModeNow(t, g, me.ID, []int{1}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !slices.Contains(effectstest.EffectiveOf(t, g, vision).Abilities, "indestructible") {
		t.Error("Density Control: indestructible until end of turn")
	}

	effectstest.CastCatalogSpell(t, g, "Test Instant 3", "Instant", "", nil)
	effectstest.NoModePick(t, g, me.ID, "all three bullets used this turn")
}

// --- Lita, Little Orphan Amphibian -----------------------------------

func TestLitaAllianceBullets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lita := effectstest.B12Push(g, me.ID, "Lita, Little Orphan Amphibian", "Legendary Creature — Mutant Ninja Turtle",
		"212fdb7c-1c7e-4eed-b1ed-bcc14a425da8", 2, 1)

	enterCreatureFor(t, g, me.ID, 1)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, lita, game.CounterPlusOne); n != 1 {
		t.Errorf("a +1/+1 counter on Lita: %d", n)
	}

	enterCreatureFor(t, g, me.ID, 1)
	effectstest.ChooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Food"); n != 1 {
		t.Errorf("a Food token: %d", n)
	}

	enterCreatureFor(t, g, me.ID, 1)
	effectstest.ChooseModeNow(t, g, me.ID, []int{2}, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("Scry 1 asks")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)

	enterCreatureFor(t, g, me.ID, 1)
	effectstest.NoModePick(t, g, me.ID, "all three bullets used this turn")
}

// --- Kargan Intimidator -----------------------------------------------

func TestKarganIntimidatorActivatedBulletsOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	kargan := effectstest.PushCatalogPermanent(g, me.ID, "Kargan Intimidator", "Creature — Human Warrior",
		"045bf1fd-f375-47a0-9983-7e2ebfed7854", false)
	theirs := effectstest.B12Creature(g, opp.ID, "Craven Knight", "Creature — Human Knight", 2, 2)

	effectstest.B16Activate(t, g, me.ID, kargan, 0, game.ActivateAbilityParams{Modes: []int{0}})
	if p := effectstest.EffectiveOf(t, g, kargan).Power; p != 2 {
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

	effectstest.B16Activate(t, g, me.ID, kargan, 0, game.ActivateAbilityParams{Modes: []int{1}, Targets: effectstest.B16TargetCard(theirs)})
	if !slices.Contains(effectstest.EffectiveOf(t, g, theirs).Subtypes, "Coward") {
		t.Errorf("becomes a Coward: subtypes %v", effectstest.EffectiveOf(t, g, theirs).Subtypes)
	}
	if !slices.Contains(effectstest.EffectiveOf(t, g, theirs).Subtypes, "Knight") {
		t.Error("in addition to its other creature types")
	}

	effectstest.B16Activate(t, g, me.ID, kargan, 0, game.ActivateAbilityParams{Modes: []int{2}, Targets: effectstest.B16TargetCard(kargan)})
	if !slices.Contains(effectstest.EffectiveOf(t, g, kargan).Abilities, "trample") {
		t.Error("the Warrior gains trample")
	}
	if got := activationModesOffered(g, me.ID, kargan); len(got) != 0 {
		t.Errorf("with all three used the ability is not offered: %v", got)
	}

	// Cowards can't block Warriors.
	var refusal bool
	g.ReadSnapshot(func() {
		k, _ := effectstest.BattlefieldCard(g, kargan)
		c, _ := effectstest.BattlefieldCard(g, theirs)
		refusal = !g.CanBlockLocked(&k, &c)
	})
	if !refusal {
		t.Error("a Coward can't block a Warrior")
	}
}

// --- Genku, Future Shaper --------------------------------------------

func TestGenkuFutureShaperMakesADifferentTokenForEachExit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.B12Push(g, me.ID, "Genku, Future Shaper", "Legendary Creature — Moonfolk Wizard",
		"ca16b735-26c7-4d0b-959b-b031f6cb376e", 2, 5)
	destroy := func(id uuid.UUID) {
		t.Helper()
		g.WithWriteLock(func() {
			if err := (effects.DestroyTarget{Target: id}).Apply(effectstest.CtxFor(g, &game.StackItem{Controller: me.ID})); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
	}

	destroy(effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	effectstest.NoModePick(t, g, me.ID, "an opponent's permanent leaving")
	enterCreatureFor(t, g, me.ID, 1)
	destroy(effectstest.FindBattlefieldByName(g, "Soldier"))
	effectstest.NoModePick(t, g, me.ID, "a token leaving")

	destroy(effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact"))
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	destroy(effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2))
	effectstest.ChooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	destroy(effectstest.B12Permanent(g, me.ID, "My Totem", "Enchantment"))
	effectstest.ChooseModeNow(t, g, me.ID, []int{2}, 2)
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Fox", "Moonfolk", "Rat"} {
		if n := effectstest.B16CountNamed(g, name); n != 1 {
			t.Errorf("one %s token, got %d", name, n)
		}
	}
	destroy(effectstest.B12Permanent(g, me.ID, "My Other Relic", "Artifact"))
	effectstest.NoModePick(t, g, me.ID, "all three bullets used this turn")
}

// --- The Fantastic Four ----------------------------------------------

func TestTheFantasticFourEntersAndFoursShareOneMemory(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
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
	effectstest.NoModePick(t, g, me.ID, "casting it does not trigger it")
	effectstest.PassPriorityAroundTable(t, g)
	hand := effectstest.HandSize(me)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+1 {
		t.Errorf("the entry draws a card: %d → %d", hand, effectstest.HandSize(me))
	}

	cast(game.Card{Name: "Three Drop", TypeLine: "Instant", ManaCost: "{1}{U}{U}"})
	effectstest.NoModePick(t, g, me.ID, "mana value 3, no power")
	effectstest.PassPriorityAroundTable(t, g)

	cast(game.Card{Name: "Four Drop", TypeLine: "Sorcery", ManaCost: "{2}{U}{U}"})
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Wall"); n != 1 {
		t.Errorf("a 0/4 Wall: %d", n)
	}

	lives := make([]int, 4)
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	cast(game.Card{Name: "Big Hitter", TypeLine: "Creature — Ogre", ManaCost: "{R}", Power: 4, Toughness: 1})
	effectstest.ChooseModeNow(t, g, me.ID, []int{1, 2}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i] - 3
		if p.ID == me.ID {
			want = lives[i]
		}
		if p.Life != want {
			t.Errorf("3 damage to each opponent: %s %d → %d", p.Name, lives[i], p.Life)
		}
	}

	wall := effectstest.FindBattlefieldByName(g, "Wall")
	cast(game.Card{Name: "Sturdy", TypeLine: "Creature — Wall", ManaCost: "{W}", Power: 0, Toughness: 4})
	effectstest.ChooseModeNow(t, g, me.ID, []int{2}, 2)
	effectstest.PickCard(t, g, me.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, wall, game.CounterPlusOne); n != 2 {
		t.Errorf("two +1/+1 counters on the target: %d", n)
	}

	cast(game.Card{Name: "Another Four", TypeLine: "Instant", ManaCost: "{3}{R}"})
	effectstest.NoModePick(t, g, me.ID, "all four bullets used this turn")
}
