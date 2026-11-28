package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b818_cards_test.go — Alliances, +1/+1 and other counters and
// cumulative upkeep, batch 818 (#5634, S82 OLD). Every test plays the real
// catalog entry through the engine.

const (
	oldB818Splintering = "184f68a1-0c13-42f4-aee6-3ca744f6a30b"
	oldB818Vermin      = "4d18bea4-c9b2-4fc7-948f-2583589994d7"
	oldB818Blessing    = "ae0f65b4-3814-45fe-a869-8585f6869278"
	oldB818Skycaptain  = "e600136a-8b4f-4725-a29e-021e0cfc7d0a"
	oldB818WarRiders   = "258921f2-df26-4d8a-ad34-8f44d0158ae7"
	oldB818ThoughtLash = "81fe7c5e-503e-488c-b06c-1615ed38e5a3"
	oldB818Dystopia    = "d0e6203b-cef2-470f-8c79-7b139290e2d5"
	oldB818Tornado     = "dd283473-6423-4a1f-b5a3-77234cdce643"
	oldB818Decree      = "fca14cbd-5ffc-441f-bccb-7ec22b03acbb"
)

func TestOldB818CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB818Splintering: "Splintering Wind", oldB818Vermin: "Diseased Vermin", oldB818Blessing: "Nature's Blessing",
		oldB818Skycaptain: "Rogue Skycaptain", oldB818WarRiders: "Varchild's War-Riders", oldB818ThoughtLash: "Thought Lash",
		oldB818Dystopia: "Dystopia", oldB818Tornado: "Tornado", oldB818Decree: "Royal Decree",
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

func oldB818Mana(p *game.Player, colors ...string) {
	p.ManaPool.AddMana(effectstest.ManaTokens(colors...)...)
}

func TestOldB818DystopiaCumulativeUpkeepPaysLifeAndSacrificesGreenOrWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Dystopia", "Enchantment", oldB818Dystopia, false)
	green := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	white := effectstest.B12Creature(g, opp.ID, "Knight", "Creature — Knight", 2, 2)
	blue := effectstest.B12Creature(g, opp.ID, "Merfolk", "Creature — Merfolk", 1, 1)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, green).Colors = []string{"G"}
		effectstest.FindBattlefieldCardForTest(g, white).Colors = []string{"W"}
		effectstest.FindBattlefieldCardForTest(g, blue).Colors = []string{"U"}
	})
	effectstest.BeginUpkeep(t, g, opp.ID)
	effectstest.AnswerSacrifice(t, g, opp.ID, white)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(white) || !g.Battlefield.Contains(green) || !g.Battlefield.Contains(blue) {
		t.Errorf("the opponent sacrifices one green or white permanent of their choice (green %v, white %v, blue %v)",
			g.Battlefield.Contains(green), g.Battlefield.Contains(white), g.Battlefield.Contains(blue))
	}
	life := me.Life
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want 1 paid for one age counter", life, me.Life)
	}
}

func TestOldB818RoyalDecreeDamagesTheControllerOfATappedSwampOrRedPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Royal Decree", "Enchantment", oldB818Decree, false)
	swamp := effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	life := opp.Life
	if err := g.ActivateManaAbility(opp.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Forest: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Fatalf("a tapped Forest cost %d life", life-opp.Life)
	}
	if err := g.ActivateManaAbility(opp.ID, swamp, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Swamp: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d after a Swamp tapped", opp.Life, life-1)
	}
	red := effectstest.PushVanillaCreature(g, me.ID, "Red Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, red).Colors = []string{"R"} })
	mine := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, red)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine-1 {
		t.Errorf("an attacking red creature tapped: your life %d, want %d", me.Life, mine-1)
	}
}

func TestOldB818ThoughtLashExilesTheTopCardToPayAndTheWholeLibraryWhenUnpaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lash := effectstest.PushCatalogPermanent(g, me.ID, "Thought Lash", "Enchantment", oldB818ThoughtLash, false)
	before := me.Library.Size()
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != before-1 || !g.Battlefield.Contains(lash) {
		t.Fatalf("library %d → %d, on battlefield %v: paying exiles one card and keeps it", before, me.Library.Size(), g.Battlefield.Contains(lash))
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 0 {
		t.Errorf("library has %d cards, want 0 after not paying", me.Library.Size())
	}
	if g.Battlefield.Contains(lash) {
		t.Error("Thought Lash was not sacrificed")
	}
}

func TestOldB818ThoughtLashPreventsOneDamageForAnExiledCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lash := effectstest.PushCatalogPermanent(g, me.ID, "Thought Lash", "Enchantment", oldB818ThoughtLash, false)
	effectstest.AdvanceToMain(t, g)
	before := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, lash, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != before-1 {
		t.Errorf("library %d → %d, want one card exiled as the cost", before, me.Library.Size())
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.BoltOracleCombat, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want 2 (3 damage, 1 prevented)", life, me.Life)
	}
}

func TestOldB818TornadoCostsMoreLifeEachTimeAndWorksOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tornado := effectstest.PushCatalogPermanent(g, me.ID, "Tornado", "Enchantment", oldB818Tornado, false)
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	b := effectstest.B12Permanent(g, opp.ID, "Pebble", "Artifact")
	life := me.Life
	oldB818Mana(me, "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, tornado, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(a)}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || effectstest.CountersOn(g, tornado, "velocity") != 1 || me.Life != life {
		t.Fatalf("first use: destroyed %v, velocity %d, life %d → %d (no life with no counters)", !g.Battlefield.Contains(a), effectstest.CountersOn(g, tornado, "velocity"), life, me.Life)
	}
	if err := g.ActivateCatalogAbility(me.ID, tornado, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(b)}); err == nil {
		t.Fatal("activated twice in one turn")
	}
	for turn, i := g.Turn.Seq, 0; i < 100 && (g.Turn.Seq == turn || g.Turn.Step != game.StepPrecombatMain); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	oldB818Mana(me, "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, tornado, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(b)}); err != nil {
		t.Fatalf("next turn's activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 || g.Battlefield.Contains(b) {
		t.Errorf("second use: life %d → %d (want -3), destroyed %v", life, me.Life, !g.Battlefield.Contains(b))
	}
}

func TestOldB818NaturesBlessingGivesACounterOrAKeywordThatLasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	blessing := effectstest.PushCatalogPermanent(g, me.ID, "Nature's Blessing", "Enchantment", oldB818Blessing, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	discard := effectstest.HandCardFull(me, "Spare", "Instant", "{U}", "", []string{"U"})
	if err := g.ActivateCatalogAbility(me.ID, blessing, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("activated without discarding")
	}
	oldB818Mana(me, "G", "W")
	if err := g.ActivateCatalogAbility(me.ID, blessing, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear), DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // banding
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "banding") {
		t.Error("no banding")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "banding") {
		t.Error("banding did not last past the turn it was given")
	}
	next := effectstest.HandCardFull(me, "Spare Two", "Instant", "{U}", "", []string{"U"})
	oldB818Mana(me, "G", "W")
	if err := g.ActivateCatalogAbility(me.ID, blessing, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear), DiscardIDs: []uuid.UUID{next}}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // +1/+1 counter
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("counters %d, want 1", effectstest.CountersOn(g, bear, game.CounterPlusOne))
	}
}

func TestOldB818DiseasedVerminInfectsAndThenBurnsThePlayerItHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vermin := effectstest.B12Push(g, me.ID, "Diseased Vermin", "Creature — Rat", oldB818Vermin, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, vermin)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, vermin, "infection") != 1 {
		t.Fatalf("infection counters %d, want 1", effectstest.CountersOn(g, vermin, "infection"))
	}
	life, other := opp.Life, g.Seats[2].Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			if len(c.PickTargetPlayers) != 1 || c.PickTargetPlayers[0] != opp.ID {
				t.Fatalf("the upkeep offers %v, want only the opponent it hit", c.PickTargetPlayers)
			}
			if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 || g.Seats[2].Life != other {
		t.Errorf("opponent %d → %d, another opponent %d → %d, want exactly 1 to the one it hit", life, opp.Life, other, g.Seats[2].Life)
	}
}

func TestOldB818RogueSkycaptainWagesOrAnOpponentTakesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := effectstest.B12Push(g, me.ID, "Rogue Skycaptain", "Creature — Human Rogue Mercenary", oldB818Skycaptain, 3, 4)
	effectstest.BeginUpkeep(t, g, me.ID)
	if effectstest.CountersOn(g, cap, "wage") != 1 {
		t.Fatalf("wage counters %d, want 1", effectstest.CountersOn(g, cap, "wage"))
	}
	oldB818Mana(me, "R", "R")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, cap) != me.ID {
		t.Fatal("paying {2} lost the creature")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, cap, "wage") != 2 {
		t.Fatalf("wage counters %d, want 2 on the second upkeep", effectstest.CountersOn(g, cap, "wage"))
	}
	if p := effectstest.RemoraUpkeepPrompt(t, g, me.ID); p.PayCost != "{2}{2}" {
		t.Errorf("the second wage is %q, want {2}{2}", p.PayCost)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, cap) != opp.ID {
		t.Error("the chosen opponent does not control the creature")
	}
	if effectstest.CountersOn(g, cap, "wage") != 0 {
		t.Errorf("wage counters %d, want them all removed", effectstest.CountersOn(g, cap, "wage"))
	}
}

func TestOldB818VarchildsWarRidersMakeAnOpponentCreateASurvivor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	riders := effectstest.B12Push(g, me.ID, "Varchild's War-Riders", "Creature — Human Warrior", oldB818WarRiders, 3, 4)
	effectstest.BeginUpkeep(t, g, me.ID)
	p := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if err := g.ResolvePayUnlessWithCards(p.ID, me.ID, true, []uuid.UUID{opp.ID}); err != nil {
		t.Fatalf("name an opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Survivor", opp.ID); n != 1 {
		t.Errorf("the opponent controls %d Survivors, want 1", n)
	}
	if !g.Battlefield.Contains(riders) {
		t.Error("a paid upkeep sacrificed the War-Riders")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, riders, "trample") {
		t.Error("no trample")
	}
}

func TestOldB818SplinteringWindPingsAndMakesASplinterThatBitesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wind := effectstest.PushCatalogPermanent(g, me.ID, "Splintering Wind", "Enchantment", oldB818Splintering, false)
	effectstest.AdvanceToMain(t, g)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 3, 3)
	life := me.Life
	oldB818Mana(me, "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, wind, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(foe)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageOn(t, g, foe) != 1 {
		t.Errorf("the target has %d damage, want 1", effectstest.DamageOn(t, g, foe))
	}
	var splinter uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Splinter" && c.Controller == me.ID {
			splinter = c.InstanceID
		}
	}
	if splinter == uuid.Nil {
		t.Fatal("no Splinter token")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, splinter); p != 1 || tough != 1 || !effectstest.HasKeywordOnBattlefield(t, g, splinter, "flying") {
		t.Errorf("the Splinter is %d/%d, flying %v", p, tough, effectstest.HasKeywordOnBattlefield(t, g, splinter, "flying"))
	}
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(splinter) {
		t.Fatal("the unpaid Splinter was not sacrificed")
	}
	if me.Life != life-1 || effectstest.DamageOn(t, g, friend) != 1 {
		t.Errorf("life %d, friend damage %d: want 1 to you and 1 to each creature you control", me.Life, effectstest.DamageOn(t, g, friend))
	}
}
