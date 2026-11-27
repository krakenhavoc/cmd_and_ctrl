package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	b719MoltenHydra     = "cd5bc95d-4915-4c5a-9b57-2730adf69797"
	b719Anthroplasm     = "f7536d94-4062-4ccc-a086-1c47954b24c7"
	b719ThranLens       = "0ac723bf-ce86-4028-a693-0ef498e4c82d"
	b719SlowMotion      = "151388a4-a24b-4089-a323-e7c1d5e6490a"
	b719Intervene       = "380b386f-90c8-45c5-a611-1ed6803b52d2"
	b719BouncingBeebles = "4fa46972-d953-4328-9c31-96e134234133"
	b719RankAndFile     = "c0b8ab40-3c27-48bd-ae19-fb1a155a0522"
	b719SpawningPool    = "f3bf22cf-0a6f-4fb6-ba82-63ce290308d6"
)

func b719Ready(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).SummonedThisTurn = false })
}

func TestB719MoltenHydraGrowsThenSpendsItsCountersAsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := effectstest.B12Push(g, me.ID, "Molten Hydra", "Creature — Hydra", b719MoltenHydra, 1, 1)
	b719Ready(g, hydra)
	effectstest.P7Activate(t, g, me, hydra, 0, game.ActivateAbilityParams{})
	effectstest.P7Activate(t, g, me, hydra, 0, game.ActivateAbilityParams{})
	if c := effectstest.ApaLive(g, hydra); c == nil || c.Counters[game.CounterPlusOne] != 2 || c.CurrentPower() != 3 {
		t.Fatalf("after two activations: %+v", c)
	}
	before := opp.Life
	effectstest.P7Activate(t, g, me, hydra, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != before-2 {
		t.Errorf("life %d, want %d (one damage per counter removed)", opp.Life, before-2)
	}
	c := effectstest.ApaLive(g, hydra)
	if c == nil || c.Counters[game.CounterPlusOne] != 0 || !c.Tapped {
		t.Errorf("the Hydra should be tapped with no counters: %+v", c)
	}
}

func TestB719MoltenHydraNeedsToTapForTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := effectstest.B12Push(g, me.ID, "Molten Hydra", "Creature — Hydra", b719MoltenHydra, 1, 1)
	b719Ready(g, hydra)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, hydra).Tapped = true })
	err := g.ActivateCatalogAbility(me.ID, hydra, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if err == nil {
		t.Fatal("a tapped Hydra paid {T}")
	}
}

func TestB719AnthroplasmEntersWithTwoCountersAndResetsToX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Anthroplasm", "Creature — Shapeshifter", b719Anthroplasm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == b719Anthroplasm {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("it did not enter the battlefield")
	}
	if c := effectstest.ApaLive(g, id); c.Counters[game.CounterPlusOne] != 2 || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Fatalf("entered as %d/%d with %d counters, want a 2/2 with two", c.CurrentPower(), c.CurrentToughness(), c.Counters[game.CounterPlusOne])
	}
	b719Ready(g, id)
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{XValue: 4})
	if c := effectstest.ApaLive(g, id); c == nil || c.Counters[game.CounterPlusOne] != 4 || c.CurrentPower() != 4 || !c.Tapped {
		t.Fatalf("after X=4: %+v", c)
	}
}

func TestB719AnthroplasmWithXZeroLosesItsCountersAndDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Anthroplasm", "Creature — Shapeshifter", b719Anthroplasm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == b719Anthroplasm {
			id = c.InstanceID
		}
	}
	b719Ready(g, id)
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{XValue: 0})
	g.RunStateChecksForTest()
	if effects.OnBattlefield(g, id) {
		t.Error("a 0/0 with no counters stayed on the battlefield")
	}
}

func TestB719ThranLensMakesEveryPermanentColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	green := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	lens := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Thran Lens", OracleID: b719ThranLens, TypeLine: "Artifact", ManaCost: "{2}"})
	if c := effectstest.ApaLive(g, green); c == nil || !c.IsColorless() {
		t.Error("a green creature is not colorless under Thran Lens")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(lens) })
	if c := effectstest.ApaLive(g, green); c == nil || !c.HasColor("G") {
		t.Error("the creature did not get its colour back when Thran Lens left")
	}
}

func b719AtUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat && i > 0 {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func b719SlowMotionOn(t *testing.T, g *game.Game, owner *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.ApaPush(g, owner.ID, owner.ID, game.Card{Name: "Slow Motion", OracleID: b719SlowMotion, TypeLine: "Enchantment — Aura", Colors: []string{"U"}})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatal(err)
		}
	})
	return aura
}

func TestB719SlowMotionSacrificesTheCreatureUnlessItsControllerPaysAndReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	aura := b719SlowMotionOn(t, g, me, bear)
	b719AtUpkeepOf(t, g, 1)
	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, opp.ID)
	if choice == nil {
		t.Fatal("the enchanted creature's controller was not asked to pay")
	}
	if err := g.ResolvePayUnless(choice.ID, opp.ID, false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, bear) {
		t.Error("the creature was not sacrificed")
	}
	if !me.Hand.Contains(aura) {
		t.Error("Slow Motion did not return to its owner's hand")
	}
}

func TestB719SlowMotionPaidCreatureStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	aura := b719SlowMotionOn(t, g, me, bear)
	b719AtUpkeepOf(t, g, 1)
	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, opp.ID)
	if choice == nil {
		t.Fatal("no prompt")
	}
	effectstest.ApaMana(opp, "C", "C")
	if err := g.ResolvePayUnless(choice.ID, opp.ID, true); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, bear) || !effects.OnBattlefield(g, aura) {
		t.Error("the creature or the Aura left after the tax was paid")
	}
}

func TestB719IntervenCountersASpellThatTargetsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTarget(bear))
	id := effectstest.PutInHand(me, game.Card{Name: "Intervene", TypeLine: "Instant", OracleID: b719Intervene})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bolt)}); err != nil {
		t.Fatalf("cast Intervene: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, bear); got != 0 {
		t.Errorf("the countered Bolt still dealt %d damage", got)
	}
}

func TestB719IntervenRefusesASpellThatTargetsOnlyAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	id := effectstest.PutInHand(me, game.Card{Name: "Intervene", TypeLine: "Instant", OracleID: b719Intervene})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bolt)}); err == nil {
		t.Fatal("Intervene accepted a spell that targets only a player")
	}
}

func TestB719BouncingBeeblesCantBeBlockedWhileTheDefenderControlsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beebles := effectstest.B12Push(g, me.ID, "Bouncing Beebles", "Creature — Beeble", b719BouncingBeebles, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.BrAttack(t, g, beebles)
	if err := g.DeclareBlocker(blocker, beebles); err == nil {
		t.Fatal("a creature blocked the Beebles while its controller had an artifact")
	}
}

func TestB719BouncingBeeblesCanBeBlockedWithoutAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beebles := effectstest.B12Push(g, me.ID, "Bouncing Beebles", "Creature — Beeble", b719BouncingBeebles, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact") // only the DEFENDING player's artifact counts
	effectstest.BrAttack(t, g, beebles)
	if err := g.DeclareBlocker(blocker, beebles); err != nil {
		t.Fatalf("the Beebles could not be blocked with no artifact on the defender's side: %v", err)
	}
}

func TestB719RankAndFileShrinksEveryGreenCreatureUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	green := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Elf", TypeLine: "Creature — Elf", Power: 2, Toughness: 2, Colors: []string{"G"}})
	mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Ent", TypeLine: "Creature — Treefolk", Power: 3, Toughness: 3, Colors: []string{"G"}})
	dies := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Sprout", TypeLine: "Creature — Plant", Power: 1, Toughness: 1, Colors: []string{"G"}})
	black := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Zombie", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}})
	effectstest.CastCatalogSpell(t, g, "Rank and File", "Creature — Zombie", b719RankAndFile, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ApaLive(g, green); c == nil || c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the opponent's green creature is %+v, want 1/1", c)
	}
	if c := effectstest.ApaLive(g, mine); c == nil || c.CurrentPower() != 2 {
		t.Error("my own green creature was not shrunk")
	}
	if effects.OnBattlefield(g, dies) {
		t.Error("a 1/1 green creature survived -1/-1")
	}
	if c := effectstest.ApaLive(g, black); c == nil || c.CurrentPower() != 2 {
		t.Error("a black creature was shrunk")
	}
}

func TestB719SpawningPoolEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Spawning Pool", b719SpawningPool)
	if c := effectstest.ApaLive(g, id); c == nil || !c.Tapped {
		t.Fatal("it did not enter tapped")
	}
}

func TestB719SpawningPoolBecomesARegeneratingSkeletonUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pool := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Spawning Pool", OracleID: b719SpawningPool, TypeLine: "Land"})
	effectstest.P7Activate(t, g, me, pool, 0, game.ActivateAbilityParams{})
	c := effectstest.ApaLive(g, pool)
	if c == nil || !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Skeleton") || !c.HasColor("B") || c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Fatalf("it is %+v, want a 1/1 black Skeleton land creature", c)
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, pool)
	if idx < 0 {
		t.Fatal("it was not given the regeneration ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, pool, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := g.RegenerationShieldsOn(pool); n != 1 {
		t.Errorf("regeneration shields %d, want 1", n)
	}
	// It ends with the turn: a creature no longer, a land still.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if c := effectstest.ApaLive(g, pool); c == nil || c.IsCreature() || !c.IsLand() {
		t.Errorf("after the turn it is %+v, want a plain land", c)
	}
}
