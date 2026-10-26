package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_sheets2_test.go — the ability stickers of the sheets, continued
// (ADR 0208, #4307).

func stkNamedOnBattlefield(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name {
			n++
		}
	}
	return n
}

func stkTap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

// stkCastCreatureFromHand puts a creature card with a mana cost in the
// active seat's hand and casts it, so the cast and enters events happen.
func stkCastFromHand(t *testing.T, g *game.Game, me *game.Player, name, typeLine, manaCost string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// Demonic Tourist Laser: outlast {1}, and seven tickets when it dies.
func TestStkDemonicTouristOutlast(t *testing.T) {
	s := newStkScene(t, "Demonic Tourist Laser", "Outlast")
	advanceToMain(t, s.g)
	s.activate(0, "R")
	if got := countersOn(s.g, s.host, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if c := s.card(); !c.Tapped {
		t.Error("outlast did not tap the creature")
	}
}

func TestStkDemonicTouristDiesForSevenTickets(t *testing.T) {
	s := newStkScene(t, "Demonic Tourist Laser", "you get seven")
	before := stkTickets(s.g, s.me)
	s.destroyHost()
	if got := stkTickets(s.g, s.me) - before; got != 7 {
		t.Errorf("tickets gained = %d, want 7", got)
	}
}

// Elemental Time Flamingo: exile it, then cast a nonland card from the
// graveyard this turn.
func TestStkFlamingoGivesAGraveyardCastThisTurn(t *testing.T) {
	s := newStkScene(t, "Elemental Time Flamingo", "cast target nonland card")
	bears := pushGraveyardPermanent(s.me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	forest := pushGraveyardPermanent(s.me, "Forest", "Basic Land — Forest", "")
	if err := p2Activate(t, s.g, s.me, s.host, 0, "", p2Card(forest)); err == nil {
		t.Fatal("a land card was a legal target")
	}
	s.activate(0, "", p2Card(bears))
	if onBattlefield(s.g, s.host) {
		t.Error("the Flamingo was not exiled as the cost")
	}
	var perm *game.CastPermission
	s.g.ReadSnapshot(func() {
		c, _ := s.g.LookupCardForEffect(bears)
		perm = s.g.CastPermissionForLocked(s.me.ID, c, game.ZoneGraveyard)
	})
	if perm == nil {
		t.Fatal("the Bears have no cast permission from the graveyard")
	}
	if perm.AltCostKey != "" || perm.ExileOnResolution {
		t.Errorf("the permission is %+v, want a plain cast at the printed cost", perm)
	}
	// Only that card: another one in the graveyard is not castable.
	other := pushGraveyardPermanent(s.me, "Other Bears", "Creature — Bear", "{1}{G}")
	s.g.ReadSnapshot(func() {
		c, _ := s.g.LookupCardForEffect(other)
		if s.g.CastPermissionForLocked(s.me.ID, c, game.ZoneGraveyard) != nil {
			t.Error("a card the ability did not name is castable")
		}
	})
}

func TestStkFlamingoDrainsWhenACreatureYouControlDies(t *testing.T) {
	s := newStkScene(t, "Elemental Time Flamingo", "Whenever a creature you control dies")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	theirs := p2Push(s.g, s.opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
	life, oppLife := s.me.Life, s.opp.Life
	b25Destroy(s.g, theirs)
	s.settle()
	if s.me.Life != life || s.opp.Life != oppLife {
		t.Error("an opponent's creature dying triggered it")
	}
	b25Destroy(s.g, other)
	s.settle()
	if s.me.Life != life+1 || s.opp.Life != oppLife-1 {
		t.Errorf("life %d→%d, opponent %d→%d, want +1 and -1", life, s.me.Life, oppLife, s.opp.Life)
	}
	s.destroyHost()
	if s.me.Life != life+2 {
		t.Errorf("the creature itself dying did not trigger it: life %d", s.me.Life)
	}
}

// Eternal Acrobat Toast: the flicker, and the untap.
func TestStkEternalAcrobatToastFlickersACreature(t *testing.T) {
	s := newStkScene(t, "Eternal Acrobat Toast", "exile target creature you control")
	// A card with an oracle ID: a card with none is read as a token, which
	// ceases to exist in exile (CR 111.7).
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "00000000-0000-0000-0000-0000000000a1", 1, 1)
	attackWith(t, s.g, s.opp.ID, s.host)
	s.pickTargets(p2Card(other))
	// It came back as a new object (CR 400.7), the one card on the
	// battlefield with its name.
	if n := stkNamedOnBattlefield(s.g, s.me.ID, "Other"); n != 1 {
		t.Fatalf("%d permanents named Other, want exactly 1 after the flicker", n)
	}
	if onBattlefield(s.g, other) {
		t.Error("the creature is the same object: it was not exiled and returned")
	}
}

func TestStkEternalAcrobatToastUntaps(t *testing.T) {
	s := newStkScene(t, "Eternal Acrobat Toast", "Untap another target permanent")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	stkTap(s.g, other)
	if err := p2Activate(t, s.g, s.me, s.host, 0, "", p2Card(s.host)); err == nil {
		t.Fatal("the permanent itself was a legal target of 'another target permanent'")
	}
	s.activate(0, "", p2Card(other))
	if stkTapped(s.g, other) {
		t.Error("the other permanent is still tapped")
	}
}

// Familiar Beeble Mascot: untap on attack; pump when a creature enters.
func TestStkFamiliarBeebleAttackUntaps(t *testing.T) {
	s := newStkScene(t, "Familiar Beeble Mascot", "untap target permanent")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	stkTap(s.g, other)
	declareAttack(t, s.g, s.opp.ID, s.host)
	s.pickTargets(p2Card(other))
	if stkTapped(s.g, other) {
		t.Error("the target is still tapped")
	}
}

func TestStkFamiliarBeebleEntersPump(t *testing.T) {
	s := newStkScene(t, "Familiar Beeble Mascot", "creatures you control get +1/+1")
	castAndResolveCreature(t, s.g, "Grizzly Bears", "Creature — Bear", "")
	s.settle()
	if p, tough := p2PT(s.g, s.host); p != 3 || tough != 3 {
		t.Errorf("host = %d/%d, want 3/3", p, tough)
	}
	// The pump is until end of turn only.
	advanceToEndStepOf(t, s.g, s.g.Turn.ActiveSeat)
	for i := 0; i < 3; i++ {
		_, _ = s.g.AdvanceStep()
	}
}

// Geek Lotus Warrior.
func TestStkGeekLotusWarriorPumps(t *testing.T) {
	s := newStkScene(t, "Geek Lotus Warrior", "+2/+0")
	s.activate(0, "RR")
	if p, _ := p2PT(s.g, s.host); p != 4 {
		t.Errorf("power = %d, want 4", p)
	}
}

func TestStkGeekLotusWarriorPingsWhenACreatureEnters(t *testing.T) {
	s := newStkScene(t, "Geek Lotus Warrior", "deals 2 damage to target player")
	life := s.opp.Life
	castAndResolveCreature(t, s.g, "Grizzly Bears", "Creature — Bear", "")
	s.settle()
	s.pickTargets(game.TargetRef{Kind: game.TargetPlayer, ID: s.opp.ID})
	if got := life - s.opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2", got)
	}
}

// Giant Mana Cake.
func TestStkGiantManaCakeLeavesForFood(t *testing.T) {
	s := newStkScene(t, "Giant Mana Cake", "two Food tokens")
	s.g.WithWriteLock(func() { _ = s.g.BounceToHandForEffect(s.host) })
	s.settle()
	if got := stkNamedOnBattlefield(s.g, s.me.ID, "Food"); got != 2 {
		t.Errorf("Food tokens = %d, want 2", got)
	}
}

func TestStkGiantManaCakeDiesDestroys(t *testing.T) {
	s := newStkScene(t, "Giant Mana Cake", "destroy target nonland permanent")
	theirs := p2Push(s.g, s.opp.ID, "Theirs", "Artifact", "", 0, 0)
	land := p2Push(s.g, s.opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	b25Destroy(s.g, s.host)
	p := latestPickTarget(s.g, s.me.ID)
	if p == nil || hasID(p.PickTargetCards, land) || !hasID(p.PickTargetCards, theirs) {
		t.Fatalf("legal targets %+v: want the artifact and not the land", p)
	}
	s.pickTargets(p2Card(theirs))
	if onBattlefield(s.g, theirs) {
		t.Error("the target was not destroyed")
	}
}

// Goblin Coward Parade: mentor, and the leaves trigger.
func TestStkGoblinCowardParadeMentors(t *testing.T) {
	s := newStkScene(t, "Goblin Coward Parade", "Mentor")
	small := p2Push(s.g, s.me.ID, "Small", "Creature — Bear", "", 1, 1)
	declareAttack(t, s.g, s.opp.ID, s.host, small)
	s.pickTargets(p2Card(small))
	if got := countersOn(s.g, small, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the lesser attacker = %d, want 1", got)
	}
	if got := countersOn(s.g, s.host, game.CounterPlusOne); got != 0 {
		t.Errorf("the mentor got %d counters", got)
	}
}

func TestStkGoblinCowardParadeLeavesDestroysABigCreature(t *testing.T) {
	s := newStkScene(t, "Goblin Coward Parade", "destroy target creature with power 4 or greater")
	big := p2Push(s.g, s.opp.ID, "Big", "Creature — Giant", "", 4, 4)
	small := p2Push(s.g, s.opp.ID, "Small", "Creature — Bear", "", 3, 3)
	b25Destroy(s.g, s.host)
	p := latestPickTarget(s.g, s.me.ID)
	if p == nil || hasID(p.PickTargetCards, small) || !hasID(p.PickTargetCards, big) {
		t.Fatalf("legal targets %+v: want only the power-4 creature", p)
	}
	if err := s.g.ResolvePickTargets(p.ID, s.me.ID, []game.TargetRef{p2Card(big)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	answerLatestTriggerPrompt(t, s.g, s.me.ID, true)
	s.settle()
	if onBattlefield(s.g, big) {
		t.Error("the big creature was not destroyed")
	}
}

// Happy Dead Squirrel: {T}: Add {C}{C}.
func TestStkHappyDeadSquirrelMakesTwoColorless(t *testing.T) {
	s := newStkScene(t, "Happy Dead Squirrel", "Add {C}{C}")
	if err := tapGrantedMana(t, s.g, s.me.ID, s.host, ""); err != nil {
		t.Fatalf("tap: %v", err)
	}
	if got := poolCount(s.me, "C"); got != 2 {
		t.Errorf("{C} in the pool = %d, want 2", got)
	}
}

// Jetpack Death Seltzer: monstrosity 3.
func TestStkJetpackDeathSeltzerMonstrosity(t *testing.T) {
	s := newStkScene(t, "Jetpack Death Seltzer", "Monstrosity 3")
	s.activate(0, "RRR")
	if got := countersOn(s.g, s.host, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
}

// Misunderstood Trapeze Elf: +X/+X for the generic mana in the spell's cost.
func TestStkMisunderstoodTrapezeElfPumpsByGenericMana(t *testing.T) {
	s := newStkScene(t, "Misunderstood Trapeze Elf", "generic mana")
	stkCastFromHand(t, s.g, s.me, "Test Spell", "Instant", "{3}{R}")
	s.settle()
	if p, tough := p2PT(s.g, s.host); p != 5 || tough != 5 {
		t.Errorf("host = %d/%d, want 5/5 (+3/+3 for {3})", p, tough)
	}
}

// A spell with no generic mana gives +0/+0.
func TestStkMisunderstoodTrapezeElfNoGenericNoPump(t *testing.T) {
	s := newStkScene(t, "Misunderstood Trapeze Elf", "generic mana")
	stkCastFromHand(t, s.g, s.me, "Test Spell", "Instant", "{R}{R}")
	s.settle()
	if p, tough := p2PT(s.g, s.host); p != 2 || tough != 2 {
		t.Errorf("host = %d/%d, want 2/2", p, tough)
	}
}

// Mystic Doom Sandwich: must be blocked if able, and +1/+1 for each blocker.
func TestStkMysticDoomSandwichIsPumpedByItsBlockers(t *testing.T) {
	s := newStkScene(t, "Mystic Doom Sandwich", "must be blocked")
	a := b12Creature(s.g, s.opp.ID, "Wall A", "Creature — Wall", 0, 4)
	b := b12Creature(s.g, s.opp.ID, "Wall B", "Creature — Wall", 0, 4)
	if c := s.card(); len(c.Effective().BlockRequirements) == 0 {
		t.Error("the sticker gave no block requirement")
	}
	unblockedAttackIntoBlockers(t, s.g, s.host)
	if err := s.g.DeclareBlocker(a, s.host); err != nil {
		t.Fatalf("block A: %v", err)
	}
	if err := s.g.DeclareBlocker(b, s.host); err != nil {
		t.Fatalf("block B: %v", err)
	}
	if err := s.g.FinishBlocks(s.opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	s.settle()
	if p, tough := p2PT(s.g, s.host); p != 4 || tough != 4 {
		t.Errorf("host = %d/%d, want 4/4 (+1/+1 for each of two blockers)", p, tough)
	}
}

// Narrow-Minded Baloney Fireworks: attack gains 2 life.
func TestStkBaloneyFireworksAttackGains2(t *testing.T) {
	s := newStkScene(t, "Narrow-Minded Baloney Fireworks", "gain 2 life")
	life := s.me.Life
	s.attack()
	if got := s.me.Life - life; got != 2 {
		t.Errorf("gained %d, want 2", got)
	}
}

// Playable Delusionary Hydra: loot, and the attack trigger.
func TestStkHydraLoots(t *testing.T) {
	s := newStkScene(t, "Playable Delusionary Hydra", "Draw a card, then discard")
	seedLibrary(s.me, "A", "B")
	lib := s.me.Library.Size()
	s.activate(0, "")
	if got := lib - s.me.Library.Size(); got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if latestChooseCardsFor(s.g, s.me.ID) == nil {
		t.Error("no discard question follows the draw")
	}
}

func TestStkHydraAttackGainsAndDraws(t *testing.T) {
	s := newStkScene(t, "Playable Delusionary Hydra", "gain 3 life and draw a card")
	seedLibrary(s.me, "A", "B")
	life, hand := s.me.Life, s.me.Hand.Size()
	s.attack()
	if s.me.Life != life+3 || s.me.Hand.Size() != hand+1 {
		t.Errorf("life %d→%d hand %d→%d, want +3 and +1", life, s.me.Life, hand, s.me.Hand.Size())
	}
}

// Primal Elder Kitty.
func TestStkPrimalElderKittyPumps(t *testing.T) {
	s := newStkScene(t, "Primal Elder Kitty", "+1/-1")
	s.activate(0, "R")
	if p, tough := p2PT(s.g, s.host); p != 3 || tough != 1 {
		t.Errorf("host = %d/%d, want 3/1", p, tough)
	}
}

func TestStkPrimalElderKittyDiesForCountersEqualToItsPower(t *testing.T) {
	s := newStkScene(t, "Primal Elder Kitty", "X +1/+1 counters")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	g := s.g
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(s.host, game.CounterPlusOne, 1) })
	b25Destroy(s.g, s.host)
	p := latestPickTarget(s.g, s.me.ID)
	if p == nil {
		t.Fatal("no target question")
	}
	if err := s.g.ResolvePickTargets(p.ID, s.me.ID, []game.TargetRef{p2Card(other)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	answerLatestTriggerPrompt(t, s.g, s.me.ID, true)
	s.settle()
	if got := countersOn(s.g, other, game.CounterPlusOne); got != 3 {
		t.Errorf("counters on the target = %d, want 3 (its power as it died: 2 + 1)", got)
	}
}

// Sassy Gremlin Blood.
func TestStkSassyGremlinBloodAttackMakesATreasure(t *testing.T) {
	s := newStkScene(t, "Sassy Gremlin Blood", "Treasure")
	s.attack()
	if got := stkNamedOnBattlefield(s.g, s.me.ID, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
}

func TestStkSassyGremlinBloodGivesFlying(t *testing.T) {
	s := newStkScene(t, "Sassy Gremlin Blood", "gains flying")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	s.activate(0, "RRR", p2Card(other))
	if !p2Keyword(s.g, other, "flying") {
		t.Error("the target did not gain flying")
	}
}

// Slimy Burrito Illusion: bushido 2.
func TestStkSlimyBurritoBushido(t *testing.T) {
	s := newStkScene(t, "Slimy Burrito Illusion", "Bushido 2")
	wall := b12Creature(s.g, s.opp.ID, "Wall", "Creature — Wall", 0, 4)
	unblockedAttackIntoBlockers(t, s.g, s.host)
	if err := s.g.DeclareBlocker(wall, s.host); err != nil {
		t.Fatalf("block: %v", err)
	}
	if err := s.g.FinishBlocks(s.opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	s.settle()
	if p, tough := p2PT(s.g, s.host); p != 4 || tough != 4 {
		t.Errorf("host = %d/%d, want 4/4", p, tough)
	}
}

// Snazzy Aether Homunculus.
func TestStkSnazzyAetherHomunculusGivesAllCreatureTypes(t *testing.T) {
	s := newStkScene(t, "Snazzy Aether Homunculus", "all creature types")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	s.activate(0, "R", p2Card(other))
	c := stkCard(t, s.g, other)
	if !game.HasAllCreatureTypes(&c) {
		t.Error("the target did not become every creature type")
	}
}

func TestStkSnazzyAetherHomunculusMagecraftDraws(t *testing.T) {
	s := newStkScene(t, "Snazzy Aether Homunculus", "Magecraft")
	seedLibrary(s.me, "A", "B", "C")
	lib := s.me.Library.Size()
	stkCastFromHand(t, s.g, s.me, "Test Instant", "Instant", "{R}")
	s.settle()
	if got := lib - s.me.Library.Size(); got != 1 {
		t.Errorf("cards drawn = %d, want 1", got)
	}
}

// Space Fungus Snickerdoodle: battle cry.
func TestStkSpaceFungusBattleCry(t *testing.T) {
	s := newStkScene(t, "Space Fungus Snickerdoodle", "Battle cry")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	declareAttack(t, s.g, s.opp.ID, s.host, other)
	s.settle()
	if p, _ := p2PT(s.g, other); p != 2 {
		t.Errorf("the other attacker's power = %d, want 2", p)
	}
	if p, _ := p2PT(s.g, s.host); p != 2 {
		t.Errorf("the battle cry creature got +1/+0 itself: power %d", p)
	}
}

// Spooky Clown Mox: {1}, {T}: Tap target creature.
func TestStkSpookyClownMoxTaps(t *testing.T) {
	s := newStkScene(t, "Spooky Clown Mox", "Tap target creature")
	theirs := p2Push(s.g, s.opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
	s.activate(0, "R", p2Card(theirs))
	if !stkTapped(s.g, theirs) {
		t.Error("the target is not tapped")
	}
}

// Squishy Sphinx Ninja: ward {2} and provoke.
func TestStkSquishySphinxNinjaWard(t *testing.T) {
	s := newStkScene(t, "Squishy Sphinx Ninja", "Ward")
	advanceToMain(t, s.g)
	if err := s.g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, s.g, s.opp, s.host)
	for i := 0; i < 8 && !hasPayUnlessFor(s.g, s.opp.ID); i++ {
		if err := s.g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(s.g, s.opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	answerPayUnless(t, s.g, s.opp.ID, false)
	s.settle()
	if !onBattlefield(s.g, s.host) {
		t.Error("the warded creature died though the payment was declined")
	}
}

func TestStkSquishySphinxNinjaProvoke(t *testing.T) {
	s := newStkScene(t, "Squishy Sphinx Ninja", "Provoke")
	theirs := b12Creature(s.g, s.opp.ID, "Theirs", "Creature — Bear", 1, 1)
	declareAttack(t, s.g, s.opp.ID, s.host)
	p := latestPickTarget(s.g, s.me.ID)
	if p == nil || !hasID(p.PickTargetCards, theirs) {
		t.Fatalf("no provoke target question offering the defender's creature: %+v", p)
	}
	if err := s.g.ResolvePickTargets(p.ID, s.me.ID, []game.TargetRef{p2Card(theirs)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	answerLatestTriggerPrompt(t, s.g, s.me.ID, true)
	s.settle()
}

// Sticky Kavu Daredevil.
func TestStkStickyKavuDiesBouncesACreature(t *testing.T) {
	s := newStkScene(t, "Sticky Kavu Daredevil", "return target creature")
	theirs := p2Push(s.g, s.opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
	b25Destroy(s.g, s.host)
	s.pickTargets(p2Card(theirs))
	answerLatestTriggerPrompt(t, s.g, s.me.ID, true)
	s.settle()
	if onBattlefield(s.g, theirs) {
		t.Error("the target is still on the battlefield")
	}
	if !s.opp.Hand.Contains(theirs) {
		t.Error("the target is not in its owner's hand")
	}
}

func TestStkStickyKavuAttackPumps(t *testing.T) {
	s := newStkScene(t, "Sticky Kavu Daredevil", "creatures you control get +1/+1")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	s.attack()
	if p, _ := p2PT(s.g, other); p != 2 {
		t.Errorf("the other creature's power = %d, want 2", p)
	}
	if p, _ := p2PT(s.g, s.host); p != 3 {
		t.Errorf("the host's power = %d, want 3", p)
	}
}

// Trained Blessed Mind.
func TestStkTrainedBlessedMindExiles(t *testing.T) {
	s := newStkScene(t, "Trained Blessed Mind", "Exile target card from a graveyard")
	dead := pushGraveyardPermanent(s.opp, "Dead", "Creature — Bear", "{1}")
	s.activate(0, "", p2Card(dead))
	if s.opp.Graveyard.Contains(dead) {
		t.Error("the card is still in the graveyard")
	}
}

func TestStkTrainedBlessedMindThreshold(t *testing.T) {
	s := newStkScene(t, "Trained Blessed Mind", "Threshold")
	for i := 0; i < 6; i++ {
		pushGraveyardPermanent(s.me, "Dead", "Creature — Bear", "{1}")
	}
	if p, _ := p2PT(s.g, s.host); p != 2 || s.has("trample") {
		t.Fatalf("six cards: power %d trample %v, want 2 and none", p, s.has("trample"))
	}
	g := s.g
	seventh := uuid.New()
	g.WithWriteLock(func() {
		s.me.Graveyard.PushTop(game.Card{InstanceID: seventh, Name: "Seventh", TypeLine: "Instant", Owner: s.me.ID, Controller: s.me.ID})
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: seventh, OldZone: game.ZoneLibrary, NewZone: game.ZoneGraveyard})
	})
	if p, _ := p2PT(s.g, s.host); p != 6 || !s.has("trample") {
		t.Errorf("seven cards: power %d trample %v, want 6 and trample", p, s.has("trample"))
	}
	// Another player's graveyard does not count.
	g.WithWriteLock(func() {
		_, _ = s.me.Graveyard.Remove(seventh)
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: seventh, OldZone: game.ZoneGraveyard, NewZone: game.ZoneHand})
	})
	if p, _ := p2PT(s.g, s.host); p != 2 {
		t.Errorf("back to six: power %d, want 2", p)
	}
}

// Trendy Circus Pirate.
func TestStkTrendyCircusPirateMakesSquirrels(t *testing.T) {
	s := newStkScene(t, "Trendy Circus Pirate", "Squirrel")
	attackWith(t, s.g, s.opp.ID, s.host)
	s.settle()
	if got := stkNamedOnBattlefield(s.g, s.me.ID, "Squirrel"); got != 2 {
		t.Errorf("Squirrels = %d, want 2 (the 2 damage dealt)", got)
	}
}

// Unassuming Gelatinous Serpent.
func TestStkGelatinousSerpentDiesReturnsACard(t *testing.T) {
	s := newStkScene(t, "Unassuming Gelatinous Serpent", "return target noncreature")
	spell := pushGraveyardPermanent(s.me, "Rock", "Artifact", "{1}")
	bear := pushGraveyardPermanent(s.me, "Dead Bear", "Creature — Bear", "{1}")
	land := pushGraveyardPermanent(s.me, "Forest", "Basic Land — Forest", "")
	b25Destroy(s.g, s.host)
	p := latestPickTarget(s.g, s.me.ID)
	if p == nil || hasID(p.PickTargetCards, bear) || hasID(p.PickTargetCards, land) || !hasID(p.PickTargetCards, spell) {
		t.Fatalf("legal targets %+v: want only the noncreature, nonland card", p)
	}
	s.pickTargets(p2Card(spell))
	if !s.me.Hand.Contains(spell) {
		t.Error("the card is not in hand")
	}
}

func TestStkGelatinousSerpentMillsTwiceTheDamage(t *testing.T) {
	s := newStkScene(t, "Unassuming Gelatinous Serpent", "mills twice")
	seedLibrary(s.opp, "A", "B", "C", "D", "E", "F")
	lib := s.opp.Library.Size()
	attackWith(t, s.g, s.opp.ID, s.host)
	s.settle()
	if got := lib - s.opp.Library.Size(); got != 4 {
		t.Errorf("milled %d, want 4 (twice the 2 damage)", got)
	}
}

// Unglued Pea-Brained Dinosaur.
func TestStkPeaBrainedDinosaurMakesTwoCreatureMana(t *testing.T) {
	s := newStkScene(t, "Unglued Pea-Brained Dinosaur", "Add {2}")
	if err := tapGrantedMana(t, s.g, s.me.ID, s.host, ""); err != nil {
		t.Fatalf("tap: %v", err)
	}
	if got := poolCount(s.me, "C"); got != 2 {
		t.Errorf("{C} in the pool = %d, want 2", got)
	}
}

func TestStkPeaBrainedDinosaurAnimatesAnArtifact(t *testing.T) {
	s := newStkScene(t, "Unglued Pea-Brained Dinosaur", "4/4 artifact creature")
	rock := p2Push(s.g, s.me.ID, "Rock", "Artifact", "", 0, 0)
	theirs := p2Push(s.g, s.opp.ID, "Their Rock", "Artifact", "", 0, 0)
	advanceTo(t, s.g, game.StepBeginCombat)
	s.settle()
	p := latestPickTarget(s.g, s.me.ID)
	if p == nil || hasID(p.PickTargetCards, theirs) || hasID(p.PickTargetCards, s.host) || !hasID(p.PickTargetCards, rock) {
		t.Fatalf("legal targets %+v: want only my noncreature artifact", p)
	}
	s.pickTargets(p2Card(rock))
	ch := effectiveOf(t, s.g, rock)
	if ch.Power != 4 || ch.Toughness != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", ch.Power, ch.Toughness)
	}
	if !p2Keyword(s.g, rock, "flying") {
		t.Error("the artifact has no flying")
	}
}

// Unhinged Beast Hunt.
func TestStkUnhingedBeastHuntGains1(t *testing.T) {
	s := newStkScene(t, "Unhinged Beast Hunt", "You gain 1 life")
	life := s.me.Life
	s.activate(0, "")
	if got := s.me.Life - life; got != 1 {
		t.Errorf("gained %d, want 1", got)
	}
}

func TestStkUnhingedBeastHuntTapsCreaturesWithTheSamePowerOrToughness(t *testing.T) {
	s := newStkScene(t, "Unhinged Beast Hunt", "tap each creature")
	samePower := p2Push(s.g, s.opp.ID, "Same power", "Creature — Bear", "", 2, 5)
	sameToughness := p2Push(s.g, s.opp.ID, "Same toughness", "Creature — Bear", "", 4, 2)
	neither := p2Push(s.g, s.opp.ID, "Neither", "Creature — Bear", "", 3, 3)
	mine := p2Push(s.g, s.me.ID, "Mine", "Creature — Bear", "", 2, 2)
	s.attack()
	if !stkTapped(s.g, samePower) || !stkTapped(s.g, sameToughness) {
		t.Error("a creature with the same power or toughness was not tapped")
	}
	if stkTapped(s.g, neither) {
		t.Error("a creature with neither was tapped")
	}
	if stkTapped(s.g, mine) {
		t.Error("my own creature was tapped")
	}
}

// Unique Charmed Pants.
func TestStkUniqueCharmedPantsTapsForAnyColor(t *testing.T) {
	s := newStkScene(t, "Unique Charmed Pants", "Add one mana of any color")
	if err := tapGrantedMana(t, s.g, s.me.ID, s.host, "U"); err != nil {
		t.Fatalf("tap: %v", err)
	}
	if got := poolCount(s.me, "U"); got != 1 {
		t.Errorf("{U} in the pool = %d, want 1", got)
	}
}

func TestStkUniqueCharmedPantsAttackPumpsByTypeCount(t *testing.T) {
	s := newStkSceneOn(t, "Unique Charmed Pants", "number of supertypes", "Legendary Artifact Creature — Golem Construct", 2, 2)
	s.attack()
	// Legendary; Artifact, Creature; Golem, Construct: five.
	if p, _ := p2PT(s.g, s.host); p != 7 {
		t.Errorf("power = %d, want 7 (2 + 5)", p)
	}
}

func TestStkUniqueCharmedPantsNotForABrushwagg(t *testing.T) {
	s := newStkSceneOn(t, "Unique Charmed Pants", "number of supertypes", "Creature — Brushwagg", 2, 2)
	s.attack()
	if p, _ := p2PT(s.g, s.host); p != 2 {
		t.Errorf("a Brushwagg got +X/+0: power %d", p)
	}
}

// Unsanctioned Ancient Juggler: bolster 1.
func TestStkAncientJugglerBolsters(t *testing.T) {
	s := newStkScene(t, "Unsanctioned Ancient Juggler", "bolster 1")
	small := p2Push(s.g, s.me.ID, "Small", "Creature — Bear", "", 1, 1)
	declareAttack(t, s.g, s.opp.ID, s.host)
	s.settle()
	if got := countersOn(s.g, small, game.CounterPlusOne); got != 1 {
		t.Errorf("the creature with the least toughness got %d counters, want 1", got)
	}
}

// Unstable Robot Dragon.
func TestStkRobotDragonSwitches(t *testing.T) {
	s := newStkSceneOn(t, "Unstable Robot Dragon", "Switch this creature", "Creature — Dragon", 5, 2)
	s.activate(0, "R")
	if p, tough := p2PT(s.g, s.host); p != 2 || tough != 5 {
		t.Errorf("host = %d/%d, want 2/5", p, tough)
	}
}

func TestStkRobotDragonAttackPumps(t *testing.T) {
	s := newStkScene(t, "Unstable Robot Dragon", "+5/+5")
	s.attack()
	if p, tough := p2PT(s.g, s.host); p != 7 || tough != 7 {
		t.Errorf("host = %d/%d, want 7/7", p, tough)
	}
}

// Werewolf Lightning Mage.
func TestStkWerewolfLightningMageLandfall(t *testing.T) {
	s := newStkScene(t, "Werewolf Lightning Mage", "Landfall")
	b31PlayLand(t, s.g, "Forest", "Basic Land — Forest", "")
	s.settle()
	if got := countersOn(s.g, s.host, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestStkWerewolfLightningMageShrinksABlocker(t *testing.T) {
	s := newStkScene(t, "Werewolf Lightning Mage", "blocks this creature")
	wall := b12Creature(s.g, s.opp.ID, "Wall", "Creature — Wall", 0, 5)
	unblockedAttackIntoBlockers(t, s.g, s.host)
	if err := s.g.DeclareBlocker(wall, s.host); err != nil {
		t.Fatalf("block: %v", err)
	}
	if err := s.g.FinishBlocks(s.opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	s.settle()
	if p, tough := p2PT(s.g, wall); p != -4 || tough != 1 {
		t.Errorf("the blocker = %d/%d, want -4/1", p, tough)
	}
}

// Yawgmoth Merfolk Soul.
func TestStkYawgmothMerfolkSoulDiscards(t *testing.T) {
	s := newStkScene(t, "Yawgmoth Merfolk Soul", "discards a card")
	pushHandCard(s.g, s.opp)
	hand := s.opp.Hand.Size()
	b25Destroy(s.g, s.host)
	s.pickTargets(game.TargetRef{Kind: game.TargetPlayer, ID: s.opp.ID})
	passPriorityAroundTable(t, s.g)
	if p := latestChooseCardsFor(s.g, s.opp.ID); p != nil {
		if err := s.g.ResolveChooseCards(p.ID, s.opp.ID, p.ChooseCards[:1]); err != nil {
			t.Fatalf("discard: %v", err)
		}
	}
	if got := hand - s.opp.Hand.Size(); got != 1 {
		t.Errorf("the player discarded %d cards, want 1", got)
	}
}

func TestStkYawgmothMerfolkSoulMakesFiveClownRobots(t *testing.T) {
	s := newStkScene(t, "Yawgmoth Merfolk Soul", "Clown Robot")
	s.destroyHost()
	if got := stkNamedOnBattlefield(s.g, s.me.ID, "Clown Robot"); got != 5 {
		t.Errorf("Clown Robots = %d, want 5", got)
	}
}

// Zombie Cheese Magician.
func TestStkZombieCheeseMagicianDrawsThatMany(t *testing.T) {
	s := newStkScene(t, "Zombie Cheese Magician", "draw that many cards")
	seedLibrary(s.me, "A", "B", "C", "D")
	hand := s.me.Hand.Size()
	attackWith(t, s.g, s.opp.ID, s.host)
	s.settle()
	if got := s.me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (the damage dealt)", got)
	}
}
