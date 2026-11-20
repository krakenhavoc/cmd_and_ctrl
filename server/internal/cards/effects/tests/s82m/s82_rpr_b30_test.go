package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsd "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/d"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b30_test.go — the reprint pass, batch 30 (#5908).

const (
	rprB30UrgeToFeed     = "c9f2f0ae-5869-43dc-a225-4dfe3d37c166"
	rprB30VampiricDragon = "e8941499-2b31-417b-b2d2-10a144826703"
	rprB30VerityCircle   = "1b80b910-c0e6-42e3-929c-2b4c85c850fe"
	rprB30Bladebrand     = "b9cf9f1e-0280-43b6-9d58-262052dcb8a8"
	rprB30ClearShot      = "0b8defd2-8530-47d7-a0c6-27aa5162dff6"
	rprB30Dissipate      = "4b2e9aa9-6f91-41de-84b6-e9a89be06a53"
	rprB30Immolation     = "9f40ad89-3767-4837-a078-f2dcfaf368df"
	rprB30Stormrider     = "2c378f72-917c-49d4-ba68-51ebaf18d976"
	rprB30TimberlandGd   = "bd8ba76f-d9e1-4c4d-83e1-6782fa3d01ba"
	rprB30UnrulyMob      = "05d0b904-a680-4f2f-aaa1-e227a7fe7e79"
)

func TestRprB30CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB30UrgeToFeed: "Urge to Feed", rprB30VampiricDragon: "Vampiric Dragon",
		rprB30VerityCircle: "Verity Circle", rprB30Bladebrand: "Bladebrand",
		rprB30ClearShot: "Clear Shot", cardsd.DelverOracle: "Delver of Secrets",
		cardsd.DelverOracle + "#1": "Insectile Aberration", rprB30Dissipate: "Dissipate",
		rprB30Immolation: "Immolation", rprB30Stormrider: "Stormrider Spirit",
		rprB30TimberlandGd: "Timberland Guide", rprB30UnrulyMob: "Unruly Mob",
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

func rprB30Ref(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func rprB30Toughness(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("no card %s", id)
		}
		n = c.CurrentToughness()
	})
	return n
}

func TestRprB30UrgeToFeedShrinksAndCountersTheVampiresYouTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 5)
	v1 := effectstest.B12Creature(g, me.ID, "Vamp One", "Creature — Vampire", 2, 2)
	v2 := effectstest.B12Creature(g, me.ID, "Vamp Two", "Creature — Vampire Knight", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Urge to Feed", "Instant", rprB30UrgeToFeed, []game.TargetRef{rprB30Ref(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := rprB30Toughness(t, g, victim); got != 2 {
		t.Fatalf("target toughness = %d, want 2 after -3/-3", got)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, v1)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, v1); !tapped {
		t.Error("the picked Vampire is not tapped")
	}
	if n := effectstest.CountersOn(g, v1, game.CounterPlusOne); n != 1 {
		t.Errorf("picked Vampire has %d +1/+1 counters, want 1", n)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, v2); tapped {
		t.Error("an unpicked Vampire was tapped")
	}
	if n := effectstest.CountersOn(g, v2, game.CounterPlusOne); n != 0 {
		t.Errorf("unpicked Vampire has %d counters, want 0", n)
	}
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("a non-Vampire got %d counters", n)
	}
}

func TestRprB30UrgeToFeedOffersOnlyUntappedVampiresAndMayTapNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 2)
	v1 := effectstest.B12Creature(g, me.ID, "Vamp One", "Creature — Vampire", 2, 2)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped Vamp", "Creature — Vampire", 2, 2)
	if err := g.TapCard(tapped, true); err != nil {
		t.Fatal(err)
	}
	effectstest.CastCatalogSpell(t, g, "Urge to Feed", "Instant", rprB30UrgeToFeed, []game.TargetRef{rprB30Ref(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(victim) {
		t.Error("a 4/2 should die to -3/-3")
	}
	for _, id := range []uuid.UUID{v1, tapped} {
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 0 {
			t.Errorf("declining the tap still put %d counters on a Vampire", n)
		}
	}
}

func TestRprB30VampiricDragonGrowsWhenACreatureItDamagedDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Vampiric Dragon", "Creature — Vampire Dragon", rprB30VampiricDragon, 5, 5)
	pinged := effectstest.B12Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 1)
	other := effectstest.B12Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{rprB30Ref(pinged)}, Strict: true}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MonSettle(t, g)
	if g.Battlefield.Contains(pinged) {
		t.Fatal("1 damage should kill a 1/1")
	}
	if n := effectstest.CountersOn(g, dragon, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Dragon has %d +1/+1 counters after its victim died, want 1", n)
	}
	// A creature the Dragon never damaged dying gives nothing.
	midB17Bolt(t, g, rprB30Ref(other))
	if g.Battlefield.Contains(other) {
		t.Fatal("Bolt should kill the Rat")
	}
	if n := effectstest.CountersOn(g, dragon, game.CounterPlusOne); n != 1 {
		t.Errorf("an undamaged creature dying changed the counters to %d", n)
	}
	// A ping that doesn't kill gives nothing.
	big := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{rprB30Ref(big)}, Strict: true}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, dragon, game.CounterPlusOne); n != 1 {
		t.Errorf("a survivor changed the counters to %d", n)
	}
	// Bolt finishes the damaged Ox: the Dragon damaged it this turn.
	midB17Bolt(t, g, rprB30Ref(big))
	if n := effectstest.CountersOn(g, dragon, game.CounterPlusOne); n != 2 {
		t.Errorf("the damaged Ox died and the Dragon has %d counters, want 2", n)
	}
}

func TestRprB30VampiricDragonPingNeedsTheMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Vampiric Dragon", "Creature — Vampire Dragon", rprB30VampiricDragon, 5, 5)
	target := effectstest.B12Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{rprB30Ref(target)}, Strict: true}); err == nil {
		t.Fatal("the ping resolved with no mana")
	}
	if !g.Battlefield.Contains(target) {
		t.Error("a refused ping killed the Mouse")
	}
}

func TestRprB30VerityCircleDrawsOnAnOpponentsNonAttackTapOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	circle := effectstest.B12Push(g, me.ID, "Verity Circle", "Enchantment", rprB30VerityCircle, 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	_ = circle
	hand := me.Hand.Size()
	// Their creature tapped by an effect: you may draw.
	g.WithWriteLock(func() {
		if err := g.TapTargetForEffect(theirs); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand = %d after an opponent's creature became tapped, want %d", got, hand+1)
	}
	// Your own creature tapping is silent.
	g.WithWriteLock(func() {
		if err := g.TapTargetForEffect(mine); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("your own creature tapping drew: hand %d", got)
	}
}

func TestRprB30VerityCircleIgnoresAnAttackDeclaration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Verity Circle", "Enchantment", rprB30VerityCircle, 0, 0)
	attacker := effectstest.B12Creature(g, opp.ID, "Their Raider", "Creature — Goblin", 2, 2)
	hand := me.Hand.Size()
	// The opponent's creature is declared as an attacker by seat 1.
	for g.Turn.ActiveSeat != 1 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.DeclareAttack(t, g, me.ID, attacker)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("an attacking tap offered the draw")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand = %d, want %d: a creature declared as an attacker is not a tap to punish", got, hand)
	}
}

func TestRprB30VerityCircleTapsAFlierlessCreatureForFiveMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	circle := effectstest.B12Push(g, me.ID, "Verity Circle", "Enchantment", rprB30VerityCircle, 0, 0)
	ground := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bird",
		TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.FloatMana(t, g, me, "{U}{U}{U}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, circle, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{rprB30Ref(flier)}, Strict: true}); err == nil {
		t.Fatal("a creature with flying was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, circle, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{rprB30Ref(ground)}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, ground); !tapped {
		t.Error("the target is not tapped")
	}
}

func TestRprB30BladebrandGrantsDeathtouchAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bladebrand", "Instant", rprB30Bladebrand, []game.TargetRef{rprB30Ref(mine)})
	effectstest.PassPriorityAroundTable(t, g)
	var has bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.LookupCardForEffect(mine)
		has = game.HasKeyword(&c, "deathtouch")
	})
	if !has {
		t.Error("the target has no deathtouch")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d (the spell's own card left it, one drawn)", got, hand+1)
	}
}

func TestRprB30ClearShotPumpsFirstThenDealsItsNewPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	other := effectstest.B12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Clear Shot", "Instant", rprB30ClearShot,
		[]game.TargetRef{rprB30Ref(mine), rprB30Ref(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the pump comes first: a 3/3 takes 3 from the now 3/3 Bear")
	}
	if got := effectstest.PowerOf(t, g, mine); got != 3 {
		t.Errorf("power = %d, want 3", got)
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("the damage is one-sided")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Clear Shot", "Instant", rprB30ClearShot,
		[]game.TargetRef{rprB30Ref(mine), rprB30Ref(other)}); err != game.ErrIllegalTarget {
		t.Errorf("a creature you control as the second target: err %v, want ErrIllegalTarget", err)
	}
}

func TestRprB30DissipateExilesTheCounteredSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	effectstest.CastCatalogSpell(t, g, "Dissipate", "Instant", rprB30Dissipate, []game.TargetRef{rprB30Ref(bolt)})
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(bolt); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the countered spell is in %v, want exile", z)
	}
	if g.Seats[1].Life != 40 {
		t.Errorf("the opponent took damage: %d", g.Seats[1].Life)
	}
	_ = me
}

func TestRprB30ImmolationGivesPlusTwoMinusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	small := effectstest.B12Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 2)
	effectstest.CastCatalogSpell(t, g, "Immolation", "Enchantment — Aura", rprB30Immolation, []game.TargetRef{rprB30Ref(big)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PowerOf(t, g, big), rprB30Toughness(t, g, big); p != 4 || tough != 1 {
		t.Errorf("enchanted Bear is %d/%d, want 4/1", p, tough)
	}
	effectstest.CastCatalogSpell(t, g, "Immolation", "Enchantment — Aura", rprB30Immolation, []game.TargetRef{rprB30Ref(small)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 1/2 with +2/-2 has 0 toughness and should die")
	}
	_ = me
}

func TestRprB30StormriderSpiritHasFlashAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastAndResolveCreature(t, g, "Stormrider Spirit", "Creature — Spirit", rprB30Stormrider)
	var flying, flash bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatal("not on the battlefield")
		}
		flying, flash = game.HasKeyword(&c, "flying"), game.HasKeyword(&c, "flash")
	})
	if !flying || !flash {
		t.Errorf("flying=%v flash=%v, want both", flying, flash)
	}
}

func TestRprB30TimberlandGuideCountersATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	guide := effectstest.CastAndResolveCreature(t, g, "Timberland Guide", "Creature — Human Scout", rprB30TimberlandGd)
	effectstest.AnswerTriggerTargets(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, theirs, game.CounterPlusOne); n != 1 {
		t.Errorf("target has %d counters, want 1", n)
	}
	if n := effectstest.CountersOn(g, guide, game.CounterPlusOne); n != 0 {
		t.Errorf("the Guide got %d counters though it was not the target", n)
	}
}

func TestRprB30UnrulyMobGrowsOnlyForAnotherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mob := effectstest.B12Push(g, me.ID, "Unruly Mob", "Creature — Human", rprB30UnrulyMob, 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Mouse", "Creature — Mouse", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 1)
	midB17Bolt(t, g, rprB30Ref(theirs))
	if n := effectstest.CountersOn(g, mob, game.CounterPlusOne); n != 0 {
		t.Fatalf("an opponent's creature dying gave %d counters", n)
	}
	midB17Bolt(t, g, rprB30Ref(mine))
	if n := effectstest.CountersOn(g, mob, game.CounterPlusOne); n != 1 {
		t.Fatalf("my creature died and the Mob has %d counters, want 1", n)
	}
}

func rprB30DelverRow() cards.Card {
	row := effectstest.VowB15Row(cardsd.DelverOracle, "Delver of Secrets", "Creature — Human Wizard", "{U}", "1", "1",
		"Insectile Aberration", "Creature — Human Insect", "3", "2", []string{"U"})
	return row
}

func TestRprB30DelverTransformsOnlyWhenYouRevealAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	delver := effectstest.VowB15Put(t, g, rprB30DelverRow(), me)
	effectstest.VowB15Library(me, "A Bear", "Creature — Bear", "{1}{G}")
	effectstest.BeginUpkeep(t, g, me.ID)
	if f := effectstest.IwFace(g, delver); f != 0 {
		t.Fatalf("a creature on top turned the Delver to face %d", f)
	}
	spell := effectstest.VowB15Library(me, "A Bolt", "Instant", "{R}")
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, delver); f != 0 {
		t.Fatalf("declining the reveal turned the Delver to face %d", f)
	}
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerChooseCards(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, delver); f != 1 {
		t.Fatalf("revealing an instant left the Delver on face %d", f)
	}
	if z := effectstest.VowB15Zone(g, spell); z != game.ZoneLibrary {
		t.Errorf("the revealed card moved to %q, want to stay on top", z)
	}
}

func TestRprB30DelverTransformsForASorceryToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	delver := effectstest.VowB15Put(t, g, rprB30DelverRow(), me)
	sorc := effectstest.VowB15Library(me, "A Ritual", "Sorcery", "{1}{B}")
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerChooseCards(t, g, me.ID, sorc)
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, delver); f != 1 {
		t.Fatalf("revealing a sorcery left the Delver on face %d", f)
	}
}
