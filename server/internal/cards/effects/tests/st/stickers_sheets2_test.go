package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	effectstest.AdvanceToMain(t, g)
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
	s := effectstest.NewStkScene(t, "Demonic Tourist Laser", "Outlast")
	effectstest.AdvanceToMain(t, s.G)
	s.Activate(0, "R")
	if got := effectstest.CountersOn(s.G, s.Host, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if c := s.Card(); !c.Tapped {
		t.Error("outlast did not tap the creature")
	}
}

func TestStkDemonicTouristDiesForSevenTickets(t *testing.T) {
	s := effectstest.NewStkScene(t, "Demonic Tourist Laser", "you get seven")
	before := stkTickets(s.G, s.Me)
	s.DestroyHost()
	if got := stkTickets(s.G, s.Me) - before; got != 7 {
		t.Errorf("tickets gained = %d, want 7", got)
	}
}

// Elemental Time Flamingo: exile it, then cast a nonland card from the
// graveyard this turn.
func TestStkFlamingoGivesAGraveyardCastThisTurn(t *testing.T) {
	s := effectstest.NewStkScene(t, "Elemental Time Flamingo", "cast target nonland card")
	bears := effectstest.PushGraveyardPermanent(s.Me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	forest := effectstest.PushGraveyardPermanent(s.Me, "Forest", "Basic Land — Forest", "")
	if err := effectstest.P2Activate(t, s.G, s.Me, s.Host, 0, "", effectstest.P2Card(forest)); err == nil {
		t.Fatal("a land card was a legal target")
	}
	s.Activate(0, "", effectstest.P2Card(bears))
	if effects.OnBattlefield(s.G, s.Host) {
		t.Error("the Flamingo was not exiled as the cost")
	}
	var perm *game.CastPermission
	s.G.ReadSnapshot(func() {
		c, _ := s.G.LookupCardForEffect(bears)
		perm = s.G.CastPermissionForLocked(s.Me.ID, c, game.ZoneGraveyard)
	})
	if perm == nil {
		t.Fatal("the Bears have no cast permission from the graveyard")
	}
	if perm.AltCostKey != "" || perm.ExileOnResolution {
		t.Errorf("the permission is %+v, want a plain cast at the printed cost", perm)
	}
	// Only that card: another one in the graveyard is not castable.
	other := effectstest.PushGraveyardPermanent(s.Me, "Other Bears", "Creature — Bear", "{1}{G}")
	s.G.ReadSnapshot(func() {
		c, _ := s.G.LookupCardForEffect(other)
		if s.G.CastPermissionForLocked(s.Me.ID, c, game.ZoneGraveyard) != nil {
			t.Error("a card the ability did not name is castable")
		}
	})
}

func TestStkFlamingoDrainsWhenACreatureYouControlDies(t *testing.T) {
	s := effectstest.NewStkScene(t, "Elemental Time Flamingo", "Whenever a creature you control dies")
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 1)
	theirs := effectstest.P2Push(s.G, s.Opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
	life, oppLife := s.Me.Life, s.Opp.Life
	effectstest.B25Destroy(s.G, theirs)
	s.Settle()
	if s.Me.Life != life || s.Opp.Life != oppLife {
		t.Error("an opponent's creature dying triggered it")
	}
	effectstest.B25Destroy(s.G, other)
	s.Settle()
	if s.Me.Life != life+1 || s.Opp.Life != oppLife-1 {
		t.Errorf("life %d→%d, opponent %d→%d, want +1 and -1", life, s.Me.Life, oppLife, s.Opp.Life)
	}
	s.DestroyHost()
	if s.Me.Life != life+2 {
		t.Errorf("the creature itself dying did not trigger it: life %d", s.Me.Life)
	}
}

// Eternal Acrobat Toast: the flicker, and the untap.
func TestStkEternalAcrobatToastFlickersACreature(t *testing.T) {
	s := effectstest.NewStkScene(t, "Eternal Acrobat Toast", "exile target creature you control")
	// A card with an oracle ID: a card with none is read as a token, which
	// ceases to exist in exile (CR 111.7).
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "00000000-0000-0000-0000-0000000000a1", 1, 1)
	effectstest.AttackWith(t, s.G, s.Opp.ID, s.Host)
	s.PickTargets(effectstest.P2Card(other))
	// It came back as a new object (CR 400.7), the one card on the
	// battlefield with its name.
	if n := stkNamedOnBattlefield(s.G, s.Me.ID, "Other"); n != 1 {
		t.Fatalf("%d permanents named Other, want exactly 1 after the flicker", n)
	}
	if effects.OnBattlefield(s.G, other) {
		t.Error("the creature is the same object: it was not exiled and returned")
	}
}

func TestStkEternalAcrobatToastUntaps(t *testing.T) {
	s := effectstest.NewStkScene(t, "Eternal Acrobat Toast", "Untap another target permanent")
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 1)
	stkTap(s.G, other)
	if err := effectstest.P2Activate(t, s.G, s.Me, s.Host, 0, "", effectstest.P2Card(s.Host)); err == nil {
		t.Fatal("the permanent itself was a legal target of 'another target permanent'")
	}
	s.Activate(0, "", effectstest.P2Card(other))
	if stkTapped(s.G, other) {
		t.Error("the other permanent is still tapped")
	}
}

// Familiar Beeble Mascot: untap on attack; pump when a creature enters.
func TestStkFamiliarBeebleAttackUntaps(t *testing.T) {
	s := effectstest.NewStkScene(t, "Familiar Beeble Mascot", "untap target permanent")
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 1)
	stkTap(s.G, other)
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, s.Host)
	s.PickTargets(effectstest.P2Card(other))
	if stkTapped(s.G, other) {
		t.Error("the target is still tapped")
	}
}

func TestStkFamiliarBeebleEntersPump(t *testing.T) {
	s := effectstest.NewStkScene(t, "Familiar Beeble Mascot", "creatures you control get +1/+1")
	effectstest.CastAndResolveCreature(t, s.G, "Grizzly Bears", "Creature — Bear", "")
	s.Settle()
	if p, tough := effectstest.P2PT(s.G, s.Host); p != 3 || tough != 3 {
		t.Errorf("host = %d/%d, want 3/3", p, tough)
	}
	// The pump is until end of turn only.
	effectstest.AdvanceToEndStepOf(t, s.G, s.G.Turn.ActiveSeat)
	for i := 0; i < 3; i++ {
		_, _ = s.G.AdvanceStep()
	}
}

// Geek Lotus Warrior.
func TestStkGeekLotusWarriorPumps(t *testing.T) {
	s := effectstest.NewStkScene(t, "Geek Lotus Warrior", "+2/+0")
	s.Activate(0, "RR")
	if p, _ := effectstest.P2PT(s.G, s.Host); p != 4 {
		t.Errorf("power = %d, want 4", p)
	}
}

func TestStkGeekLotusWarriorPingsWhenACreatureEnters(t *testing.T) {
	s := effectstest.NewStkScene(t, "Geek Lotus Warrior", "deals 2 damage to target player")
	life := s.Opp.Life
	effectstest.CastAndResolveCreature(t, s.G, "Grizzly Bears", "Creature — Bear", "")
	s.Settle()
	s.PickTargets(game.TargetRef{Kind: game.TargetPlayer, ID: s.Opp.ID})
	if got := life - s.Opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2", got)
	}
}

// Giant Mana Cake.
func TestStkGiantManaCakeLeavesForFood(t *testing.T) {
	s := effectstest.NewStkScene(t, "Giant Mana Cake", "two Food tokens")
	s.G.WithWriteLock(func() { _ = s.G.BounceToHandForEffect(s.Host) })
	s.Settle()
	if got := stkNamedOnBattlefield(s.G, s.Me.ID, "Food"); got != 2 {
		t.Errorf("Food tokens = %d, want 2", got)
	}
}

func TestStkGiantManaCakeDiesDestroys(t *testing.T) {
	s := effectstest.NewStkScene(t, "Giant Mana Cake", "destroy target nonland permanent")
	theirs := effectstest.P2Push(s.G, s.Opp.ID, "Theirs", "Artifact", "", 0, 0)
	land := effectstest.P2Push(s.G, s.Opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.B25Destroy(s.G, s.Host)
	p := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if p == nil || effectstest.HasID(p.PickTargetCards, land) || !effectstest.HasID(p.PickTargetCards, theirs) {
		t.Fatalf("legal targets %+v: want the artifact and not the land", p)
	}
	s.PickTargets(effectstest.P2Card(theirs))
	if effects.OnBattlefield(s.G, theirs) {
		t.Error("the target was not destroyed")
	}
}

// Goblin Coward Parade: mentor, and the leaves trigger.
func TestStkGoblinCowardParadeMentors(t *testing.T) {
	s := effectstest.NewStkScene(t, "Goblin Coward Parade", "Mentor")
	small := effectstest.P2Push(s.G, s.Me.ID, "Small", "Creature — Bear", "", 1, 1)
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, s.Host, small)
	s.PickTargets(effectstest.P2Card(small))
	if got := effectstest.CountersOn(s.G, small, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the lesser attacker = %d, want 1", got)
	}
	if got := effectstest.CountersOn(s.G, s.Host, game.CounterPlusOne); got != 0 {
		t.Errorf("the mentor got %d counters", got)
	}
}

func TestStkGoblinCowardParadeLeavesDestroysABigCreature(t *testing.T) {
	s := effectstest.NewStkScene(t, "Goblin Coward Parade", "destroy target creature with power 4 or greater")
	big := effectstest.P2Push(s.G, s.Opp.ID, "Big", "Creature — Giant", "", 4, 4)
	small := effectstest.P2Push(s.G, s.Opp.ID, "Small", "Creature — Bear", "", 3, 3)
	effectstest.B25Destroy(s.G, s.Host)
	p := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if p == nil || effectstest.HasID(p.PickTargetCards, small) || !effectstest.HasID(p.PickTargetCards, big) {
		t.Fatalf("legal targets %+v: want only the power-4 creature", p)
	}
	if err := s.G.ResolvePickTargets(p.ID, s.Me.ID, []game.TargetRef{effectstest.P2Card(big)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, s.G, s.Me.ID, true)
	s.Settle()
	if effects.OnBattlefield(s.G, big) {
		t.Error("the big creature was not destroyed")
	}
}

// Happy Dead Squirrel: {T}: Add {C}{C}.
func TestStkHappyDeadSquirrelMakesTwoColorless(t *testing.T) {
	s := effectstest.NewStkScene(t, "Happy Dead Squirrel", "Add {C}{C}")
	if err := effectstest.TapGrantedMana(t, s.G, s.Me.ID, s.Host, ""); err != nil {
		t.Fatalf("tap: %v", err)
	}
	if got := effectstest.PoolCount(s.Me, "C"); got != 2 {
		t.Errorf("{C} in the pool = %d, want 2", got)
	}
}

// Jetpack Death Seltzer: monstrosity 3.
func TestStkJetpackDeathSeltzerMonstrosity(t *testing.T) {
	s := effectstest.NewStkScene(t, "Jetpack Death Seltzer", "Monstrosity 3")
	s.Activate(0, "RRR")
	if got := effectstest.CountersOn(s.G, s.Host, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
}

// Misunderstood Trapeze Elf: +X/+X for the generic mana in the spell's cost.
func TestStkMisunderstoodTrapezeElfPumpsByGenericMana(t *testing.T) {
	s := effectstest.NewStkScene(t, "Misunderstood Trapeze Elf", "generic mana")
	stkCastFromHand(t, s.G, s.Me, "Test Spell", "Instant", "{3}{R}")
	s.Settle()
	if p, tough := effectstest.P2PT(s.G, s.Host); p != 5 || tough != 5 {
		t.Errorf("host = %d/%d, want 5/5 (+3/+3 for {3})", p, tough)
	}
}

// A spell with no generic mana gives +0/+0.
func TestStkMisunderstoodTrapezeElfNoGenericNoPump(t *testing.T) {
	s := effectstest.NewStkScene(t, "Misunderstood Trapeze Elf", "generic mana")
	stkCastFromHand(t, s.G, s.Me, "Test Spell", "Instant", "{R}{R}")
	s.Settle()
	if p, tough := effectstest.P2PT(s.G, s.Host); p != 2 || tough != 2 {
		t.Errorf("host = %d/%d, want 2/2", p, tough)
	}
}

// Mystic Doom Sandwich: must be blocked if able, and +1/+1 for each blocker.
func TestStkMysticDoomSandwichIsPumpedByItsBlockers(t *testing.T) {
	s := effectstest.NewStkScene(t, "Mystic Doom Sandwich", "must be blocked")
	a := effectstest.B12Creature(s.G, s.Opp.ID, "Wall A", "Creature — Wall", 0, 4)
	b := effectstest.B12Creature(s.G, s.Opp.ID, "Wall B", "Creature — Wall", 0, 4)
	if c := s.Card(); len(c.Effective().BlockRequirements) == 0 {
		t.Error("the sticker gave no block requirement")
	}
	effectstest.UnblockedAttackIntoBlockers(t, s.G, s.Host)
	if err := s.G.DeclareBlocker(a, s.Host); err != nil {
		t.Fatalf("block A: %v", err)
	}
	if err := s.G.DeclareBlocker(b, s.Host); err != nil {
		t.Fatalf("block B: %v", err)
	}
	if err := s.G.FinishBlocks(s.Opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	s.Settle()
	if p, tough := effectstest.P2PT(s.G, s.Host); p != 4 || tough != 4 {
		t.Errorf("host = %d/%d, want 4/4 (+1/+1 for each of two blockers)", p, tough)
	}
}

// Narrow-Minded Baloney Fireworks: attack gains 2 life.
func TestStkBaloneyFireworksAttackGains2(t *testing.T) {
	s := effectstest.NewStkScene(t, "Narrow-Minded Baloney Fireworks", "gain 2 life")
	life := s.Me.Life
	s.Attack()
	if got := s.Me.Life - life; got != 2 {
		t.Errorf("gained %d, want 2", got)
	}
}

// Playable Delusionary Hydra: loot, and the attack trigger.
func TestStkHydraLoots(t *testing.T) {
	s := effectstest.NewStkScene(t, "Playable Delusionary Hydra", "Draw a card, then discard")
	effectstest.SeedLibrary(s.Me, "A", "B")
	lib := s.Me.Library.Size()
	s.Activate(0, "")
	if got := lib - s.Me.Library.Size(); got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if effectstest.LatestChooseCardsFor(s.G, s.Me.ID) == nil {
		t.Error("no discard question follows the draw")
	}
}

func TestStkHydraAttackGainsAndDraws(t *testing.T) {
	s := effectstest.NewStkScene(t, "Playable Delusionary Hydra", "gain 3 life and draw a card")
	effectstest.SeedLibrary(s.Me, "A", "B")
	life, hand := s.Me.Life, s.Me.Hand.Size()
	s.Attack()
	if s.Me.Life != life+3 || s.Me.Hand.Size() != hand+1 {
		t.Errorf("life %d→%d hand %d→%d, want +3 and +1", life, s.Me.Life, hand, s.Me.Hand.Size())
	}
}

// Primal Elder Kitty.
func TestStkPrimalElderKittyPumps(t *testing.T) {
	s := effectstest.NewStkScene(t, "Primal Elder Kitty", "+1/-1")
	s.Activate(0, "R")
	if p, tough := effectstest.P2PT(s.G, s.Host); p != 3 || tough != 1 {
		t.Errorf("host = %d/%d, want 3/1", p, tough)
	}
}

func TestStkPrimalElderKittyDiesForCountersEqualToItsPower(t *testing.T) {
	s := effectstest.NewStkScene(t, "Primal Elder Kitty", "X +1/+1 counters")
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 1)
	g := s.G
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(s.Host, game.CounterPlusOne, 1) })
	effectstest.B25Destroy(s.G, s.Host)
	p := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if p == nil {
		t.Fatal("no target question")
	}
	if err := s.G.ResolvePickTargets(p.ID, s.Me.ID, []game.TargetRef{effectstest.P2Card(other)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, s.G, s.Me.ID, true)
	s.Settle()
	if got := effectstest.CountersOn(s.G, other, game.CounterPlusOne); got != 3 {
		t.Errorf("counters on the target = %d, want 3 (its power as it died: 2 + 1)", got)
	}
}

// Sassy Gremlin Blood.
func TestStkSassyGremlinBloodAttackMakesATreasure(t *testing.T) {
	s := effectstest.NewStkScene(t, "Sassy Gremlin Blood", "Treasure")
	s.Attack()
	if got := stkNamedOnBattlefield(s.G, s.Me.ID, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
}

func TestStkSassyGremlinBloodGivesFlying(t *testing.T) {
	s := effectstest.NewStkScene(t, "Sassy Gremlin Blood", "gains flying")
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 1)
	s.Activate(0, "RRR", effectstest.P2Card(other))
	if !effectstest.P2Keyword(s.G, other, "flying") {
		t.Error("the target did not gain flying")
	}
}

// Slimy Burrito Illusion: bushido 2.
func TestStkSlimyBurritoBushido(t *testing.T) {
	s := effectstest.NewStkScene(t, "Slimy Burrito Illusion", "Bushido 2")
	wall := effectstest.B12Creature(s.G, s.Opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.UnblockedAttackIntoBlockers(t, s.G, s.Host)
	if err := s.G.DeclareBlocker(wall, s.Host); err != nil {
		t.Fatalf("block: %v", err)
	}
	if err := s.G.FinishBlocks(s.Opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	s.Settle()
	if p, tough := effectstest.P2PT(s.G, s.Host); p != 4 || tough != 4 {
		t.Errorf("host = %d/%d, want 4/4", p, tough)
	}
}

// Snazzy Aether Homunculus.
func TestStkSnazzyAetherHomunculusGivesAllCreatureTypes(t *testing.T) {
	s := effectstest.NewStkScene(t, "Snazzy Aether Homunculus", "all creature types")
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 1)
	s.Activate(0, "R", effectstest.P2Card(other))
	c := effectstest.StkCard(t, s.G, other)
	if !game.HasAllCreatureTypes(&c) {
		t.Error("the target did not become every creature type")
	}
}

func TestStkSnazzyAetherHomunculusMagecraftDraws(t *testing.T) {
	s := effectstest.NewStkScene(t, "Snazzy Aether Homunculus", "Magecraft")
	effectstest.SeedLibrary(s.Me, "A", "B", "C")
	lib := s.Me.Library.Size()
	stkCastFromHand(t, s.G, s.Me, "Test Instant", "Instant", "{R}")
	s.Settle()
	if got := lib - s.Me.Library.Size(); got != 1 {
		t.Errorf("cards drawn = %d, want 1", got)
	}
}

// Space Fungus Snickerdoodle: battle cry.
func TestStkSpaceFungusBattleCry(t *testing.T) {
	s := effectstest.NewStkScene(t, "Space Fungus Snickerdoodle", "Battle cry")
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 1)
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, s.Host, other)
	s.Settle()
	if p, _ := effectstest.P2PT(s.G, other); p != 2 {
		t.Errorf("the other attacker's power = %d, want 2", p)
	}
	if p, _ := effectstest.P2PT(s.G, s.Host); p != 2 {
		t.Errorf("the battle cry creature got +1/+0 itself: power %d", p)
	}
}

// Spooky Clown Mox: {1}, {T}: Tap target creature.
func TestStkSpookyClownMoxTaps(t *testing.T) {
	s := effectstest.NewStkScene(t, "Spooky Clown Mox", "Tap target creature")
	theirs := effectstest.P2Push(s.G, s.Opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
	s.Activate(0, "R", effectstest.P2Card(theirs))
	if !stkTapped(s.G, theirs) {
		t.Error("the target is not tapped")
	}
}

// Squishy Sphinx Ninja: ward {2} and provoke.
func TestStkSquishySphinxNinjaWard(t *testing.T) {
	s := effectstest.NewStkScene(t, "Squishy Sphinx Ninja", "Ward")
	effectstest.AdvanceToMain(t, s.G)
	if err := s.G.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, s.G, s.Opp, s.Host)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(s.G, s.Opp.ID); i++ {
		if err := s.G.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(s.G, s.Opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	effectstest.AnswerPayUnless(t, s.G, s.Opp.ID, false)
	s.Settle()
	if !effects.OnBattlefield(s.G, s.Host) {
		t.Error("the warded creature died though the payment was declined")
	}
}

func TestStkSquishySphinxNinjaProvoke(t *testing.T) {
	s := effectstest.NewStkScene(t, "Squishy Sphinx Ninja", "Provoke")
	theirs := effectstest.B12Creature(s.G, s.Opp.ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, s.Host)
	p := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, theirs) {
		t.Fatalf("no provoke target question offering the defender's creature: %+v", p)
	}
	if err := s.G.ResolvePickTargets(p.ID, s.Me.ID, []game.TargetRef{effectstest.P2Card(theirs)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, s.G, s.Me.ID, true)
	s.Settle()
}

// Sticky Kavu Daredevil.
func TestStkStickyKavuDiesBouncesACreature(t *testing.T) {
	s := effectstest.NewStkScene(t, "Sticky Kavu Daredevil", "return target creature")
	theirs := effectstest.P2Push(s.G, s.Opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
	effectstest.B25Destroy(s.G, s.Host)
	s.PickTargets(effectstest.P2Card(theirs))
	effectstest.AnswerLatestTriggerPrompt(t, s.G, s.Me.ID, true)
	s.Settle()
	if effects.OnBattlefield(s.G, theirs) {
		t.Error("the target is still on the battlefield")
	}
	if !s.Opp.Hand.Contains(theirs) {
		t.Error("the target is not in its owner's hand")
	}
}

func TestStkStickyKavuAttackPumps(t *testing.T) {
	s := effectstest.NewStkScene(t, "Sticky Kavu Daredevil", "creatures you control get +1/+1")
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 1)
	s.Attack()
	if p, _ := effectstest.P2PT(s.G, other); p != 2 {
		t.Errorf("the other creature's power = %d, want 2", p)
	}
	if p, _ := effectstest.P2PT(s.G, s.Host); p != 3 {
		t.Errorf("the host's power = %d, want 3", p)
	}
}

// Trained Blessed Mind.
func TestStkTrainedBlessedMindExiles(t *testing.T) {
	s := effectstest.NewStkScene(t, "Trained Blessed Mind", "Exile target card from a graveyard")
	dead := effectstest.PushGraveyardPermanent(s.Opp, "Dead", "Creature — Bear", "{1}")
	s.Activate(0, "", effectstest.P2Card(dead))
	if s.Opp.Graveyard.Contains(dead) {
		t.Error("the card is still in the graveyard")
	}
}

func TestStkTrainedBlessedMindThreshold(t *testing.T) {
	s := effectstest.NewStkScene(t, "Trained Blessed Mind", "Threshold")
	for i := 0; i < 6; i++ {
		effectstest.PushGraveyardPermanent(s.Me, "Dead", "Creature — Bear", "{1}")
	}
	if p, _ := effectstest.P2PT(s.G, s.Host); p != 2 || s.Has("trample") {
		t.Fatalf("six cards: power %d trample %v, want 2 and none", p, s.Has("trample"))
	}
	g := s.G
	seventh := uuid.New()
	g.WithWriteLock(func() {
		s.Me.Graveyard.PushTop(game.Card{InstanceID: seventh, Name: "Seventh", TypeLine: "Instant", Owner: s.Me.ID, Controller: s.Me.ID})
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: seventh, OldZone: game.ZoneLibrary, NewZone: game.ZoneGraveyard})
	})
	if p, _ := effectstest.P2PT(s.G, s.Host); p != 6 || !s.Has("trample") {
		t.Errorf("seven cards: power %d trample %v, want 6 and trample", p, s.Has("trample"))
	}
	// Another player's graveyard does not count.
	g.WithWriteLock(func() {
		_, _ = s.Me.Graveyard.Remove(seventh)
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: seventh, OldZone: game.ZoneGraveyard, NewZone: game.ZoneHand})
	})
	if p, _ := effectstest.P2PT(s.G, s.Host); p != 2 {
		t.Errorf("back to six: power %d, want 2", p)
	}
}

// Trendy Circus Pirate.
func TestStkTrendyCircusPirateMakesSquirrels(t *testing.T) {
	s := effectstest.NewStkScene(t, "Trendy Circus Pirate", "Squirrel")
	effectstest.AttackWith(t, s.G, s.Opp.ID, s.Host)
	s.Settle()
	if got := stkNamedOnBattlefield(s.G, s.Me.ID, "Squirrel"); got != 2 {
		t.Errorf("Squirrels = %d, want 2 (the 2 damage dealt)", got)
	}
}

// Unassuming Gelatinous Serpent.
func TestStkGelatinousSerpentDiesReturnsACard(t *testing.T) {
	s := effectstest.NewStkScene(t, "Unassuming Gelatinous Serpent", "return target noncreature")
	spell := effectstest.PushGraveyardPermanent(s.Me, "Rock", "Artifact", "{1}")
	bear := effectstest.PushGraveyardPermanent(s.Me, "Dead Bear", "Creature — Bear", "{1}")
	land := effectstest.PushGraveyardPermanent(s.Me, "Forest", "Basic Land — Forest", "")
	effectstest.B25Destroy(s.G, s.Host)
	p := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if p == nil || effectstest.HasID(p.PickTargetCards, bear) || effectstest.HasID(p.PickTargetCards, land) || !effectstest.HasID(p.PickTargetCards, spell) {
		t.Fatalf("legal targets %+v: want only the noncreature, nonland card", p)
	}
	s.PickTargets(effectstest.P2Card(spell))
	if !s.Me.Hand.Contains(spell) {
		t.Error("the card is not in hand")
	}
}

func TestStkGelatinousSerpentMillsTwiceTheDamage(t *testing.T) {
	s := effectstest.NewStkScene(t, "Unassuming Gelatinous Serpent", "mills twice")
	effectstest.SeedLibrary(s.Opp, "A", "B", "C", "D", "E", "F")
	lib := s.Opp.Library.Size()
	effectstest.AttackWith(t, s.G, s.Opp.ID, s.Host)
	s.Settle()
	if got := lib - s.Opp.Library.Size(); got != 4 {
		t.Errorf("milled %d, want 4 (twice the 2 damage)", got)
	}
}

// Unglued Pea-Brained Dinosaur.
func TestStkPeaBrainedDinosaurMakesTwoCreatureMana(t *testing.T) {
	s := effectstest.NewStkScene(t, "Unglued Pea-Brained Dinosaur", "Add {2}")
	if err := effectstest.TapGrantedMana(t, s.G, s.Me.ID, s.Host, ""); err != nil {
		t.Fatalf("tap: %v", err)
	}
	if got := effectstest.PoolCount(s.Me, "C"); got != 2 {
		t.Errorf("{C} in the pool = %d, want 2", got)
	}
}

func TestStkPeaBrainedDinosaurAnimatesAnArtifact(t *testing.T) {
	s := effectstest.NewStkScene(t, "Unglued Pea-Brained Dinosaur", "4/4 artifact creature")
	rock := effectstest.P2Push(s.G, s.Me.ID, "Rock", "Artifact", "", 0, 0)
	theirs := effectstest.P2Push(s.G, s.Opp.ID, "Their Rock", "Artifact", "", 0, 0)
	effectstest.AdvanceTo(t, s.G, game.StepBeginCombat)
	s.Settle()
	p := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if p == nil || effectstest.HasID(p.PickTargetCards, theirs) || effectstest.HasID(p.PickTargetCards, s.Host) || !effectstest.HasID(p.PickTargetCards, rock) {
		t.Fatalf("legal targets %+v: want only my noncreature artifact", p)
	}
	s.PickTargets(effectstest.P2Card(rock))
	ch := effectstest.EffectiveOf(t, s.G, rock)
	if ch.Power != 4 || ch.Toughness != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", ch.Power, ch.Toughness)
	}
	if !effectstest.P2Keyword(s.G, rock, "flying") {
		t.Error("the artifact has no flying")
	}
}

// Unhinged Beast Hunt.
func TestStkUnhingedBeastHuntGains1(t *testing.T) {
	s := effectstest.NewStkScene(t, "Unhinged Beast Hunt", "You gain 1 life")
	life := s.Me.Life
	s.Activate(0, "")
	if got := s.Me.Life - life; got != 1 {
		t.Errorf("gained %d, want 1", got)
	}
}

func TestStkUnhingedBeastHuntTapsCreaturesWithTheSamePowerOrToughness(t *testing.T) {
	s := effectstest.NewStkScene(t, "Unhinged Beast Hunt", "tap each creature")
	samePower := effectstest.P2Push(s.G, s.Opp.ID, "Same power", "Creature — Bear", "", 2, 5)
	sameToughness := effectstest.P2Push(s.G, s.Opp.ID, "Same toughness", "Creature — Bear", "", 4, 2)
	neither := effectstest.P2Push(s.G, s.Opp.ID, "Neither", "Creature — Bear", "", 3, 3)
	mine := effectstest.P2Push(s.G, s.Me.ID, "Mine", "Creature — Bear", "", 2, 2)
	s.Attack()
	if !stkTapped(s.G, samePower) || !stkTapped(s.G, sameToughness) {
		t.Error("a creature with the same power or toughness was not tapped")
	}
	if stkTapped(s.G, neither) {
		t.Error("a creature with neither was tapped")
	}
	if stkTapped(s.G, mine) {
		t.Error("my own creature was tapped")
	}
}

// Unique Charmed Pants.
func TestStkUniqueCharmedPantsTapsForAnyColor(t *testing.T) {
	s := effectstest.NewStkScene(t, "Unique Charmed Pants", "Add one mana of any color")
	if err := effectstest.TapGrantedMana(t, s.G, s.Me.ID, s.Host, "U"); err != nil {
		t.Fatalf("tap: %v", err)
	}
	if got := effectstest.PoolCount(s.Me, "U"); got != 1 {
		t.Errorf("{U} in the pool = %d, want 1", got)
	}
}

func TestStkUniqueCharmedPantsAttackPumpsByTypeCount(t *testing.T) {
	s := effectstest.NewStkSceneOn(t, "Unique Charmed Pants", "number of supertypes", "Legendary Artifact Creature — Golem Construct", 2, 2)
	s.Attack()
	// Legendary; Artifact, Creature; Golem, Construct: five.
	if p, _ := effectstest.P2PT(s.G, s.Host); p != 7 {
		t.Errorf("power = %d, want 7 (2 + 5)", p)
	}
}

func TestStkUniqueCharmedPantsNotForABrushwagg(t *testing.T) {
	s := effectstest.NewStkSceneOn(t, "Unique Charmed Pants", "number of supertypes", "Creature — Brushwagg", 2, 2)
	s.Attack()
	if p, _ := effectstest.P2PT(s.G, s.Host); p != 2 {
		t.Errorf("a Brushwagg got +X/+0: power %d", p)
	}
}

// Unsanctioned Ancient Juggler: bolster 1.
func TestStkAncientJugglerBolsters(t *testing.T) {
	s := effectstest.NewStkScene(t, "Unsanctioned Ancient Juggler", "bolster 1")
	small := effectstest.P2Push(s.G, s.Me.ID, "Small", "Creature — Bear", "", 1, 1)
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, s.Host)
	s.Settle()
	if got := effectstest.CountersOn(s.G, small, game.CounterPlusOne); got != 1 {
		t.Errorf("the creature with the least toughness got %d counters, want 1", got)
	}
}

// Unstable Robot Dragon.
func TestStkRobotDragonSwitches(t *testing.T) {
	s := effectstest.NewStkSceneOn(t, "Unstable Robot Dragon", "Switch this creature", "Creature — Dragon", 5, 2)
	s.Activate(0, "R")
	if p, tough := effectstest.P2PT(s.G, s.Host); p != 2 || tough != 5 {
		t.Errorf("host = %d/%d, want 2/5", p, tough)
	}
}

func TestStkRobotDragonAttackPumps(t *testing.T) {
	s := effectstest.NewStkScene(t, "Unstable Robot Dragon", "+5/+5")
	s.Attack()
	if p, tough := effectstest.P2PT(s.G, s.Host); p != 7 || tough != 7 {
		t.Errorf("host = %d/%d, want 7/7", p, tough)
	}
}

// Werewolf Lightning Mage.
func TestStkWerewolfLightningMageLandfall(t *testing.T) {
	s := effectstest.NewStkScene(t, "Werewolf Lightning Mage", "Landfall")
	effectstest.B31PlayLand(t, s.G, "Forest", "Basic Land — Forest", "")
	s.Settle()
	if got := effectstest.CountersOn(s.G, s.Host, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestStkWerewolfLightningMageShrinksABlocker(t *testing.T) {
	s := effectstest.NewStkScene(t, "Werewolf Lightning Mage", "blocks this creature")
	wall := effectstest.B12Creature(s.G, s.Opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.UnblockedAttackIntoBlockers(t, s.G, s.Host)
	if err := s.G.DeclareBlocker(wall, s.Host); err != nil {
		t.Fatalf("block: %v", err)
	}
	if err := s.G.FinishBlocks(s.Opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	s.Settle()
	if p, tough := effectstest.P2PT(s.G, wall); p != -4 || tough != 1 {
		t.Errorf("the blocker = %d/%d, want -4/1", p, tough)
	}
}

// Yawgmoth Merfolk Soul.
func TestStkYawgmothMerfolkSoulDiscards(t *testing.T) {
	s := effectstest.NewStkScene(t, "Yawgmoth Merfolk Soul", "discards a card")
	effectstest.PushHandCard(s.G, s.Opp)
	hand := s.Opp.Hand.Size()
	effectstest.B25Destroy(s.G, s.Host)
	s.PickTargets(game.TargetRef{Kind: game.TargetPlayer, ID: s.Opp.ID})
	effectstest.PassPriorityAroundTable(t, s.G)
	if p := effectstest.LatestChooseCardsFor(s.G, s.Opp.ID); p != nil {
		if err := s.G.ResolveChooseCards(p.ID, s.Opp.ID, p.ChooseCards[:1]); err != nil {
			t.Fatalf("discard: %v", err)
		}
	}
	if got := hand - s.Opp.Hand.Size(); got != 1 {
		t.Errorf("the player discarded %d cards, want 1", got)
	}
}

func TestStkYawgmothMerfolkSoulMakesFiveClownRobots(t *testing.T) {
	s := effectstest.NewStkScene(t, "Yawgmoth Merfolk Soul", "Clown Robot")
	s.DestroyHost()
	if got := stkNamedOnBattlefield(s.G, s.Me.ID, "Clown Robot"); got != 5 {
		t.Errorf("Clown Robots = %d, want 5", got)
	}
}

// Zombie Cheese Magician.
func TestStkZombieCheeseMagicianDrawsThatMany(t *testing.T) {
	s := effectstest.NewStkScene(t, "Zombie Cheese Magician", "draw that many cards")
	effectstest.SeedLibrary(s.Me, "A", "B", "C", "D")
	hand := s.Me.Hand.Size()
	effectstest.AttackWith(t, s.G, s.Opp.ID, s.Host)
	s.Settle()
	if got := s.Me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (the damage dealt)", got)
	}
}
