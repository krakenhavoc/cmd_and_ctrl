package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b149_cards_test.go — GAP batch 149, Jumpstart 2022 creatures
// (#5948). Every card is played through the engine.

const (
	gap149LurkingDeadeyeOracle      = "d1dd27eb-ba7b-4d5d-a36c-3f350df2ad51"
	gap149MaraudingBoneslasherOracl = "a0bc9c88-8bec-4d92-a078-08883c82cb28"
	gap149NestedGhoulOracle         = "e60b6b71-eea8-42b3-81e2-c8fb1af5e218"
	gap149NezumiBoneReaderOracle    = "c77f8fee-8d2d-4cec-948f-2ca3e5b64070"
	gap149NirkanaAssassinOracle     = "efb64da9-1ca7-4686-a427-e0d67101030c"
	gap149OssuaryRatsOracle         = "a1380d07-f582-46dd-a63a-9806123ebcd3"
	gap149PhyrexianPlaguelordOracle = "aff9e844-9e03-490b-b44f-10d385738cc6"
	gap149PitKeeperOracle           = "e1c5bfae-07cb-4194-8a86-201503e17085"
	gap149PlagueSpitterOracle       = "5feedfb0-30e6-400d-9e28-d541ea1aa14e"
)

func TestGap149EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		gap149LurkingDeadeyeOracle, gap149MaraudingBoneslasherOracl, gap149NestedGhoulOracle,
		gap149NezumiBoneReaderOracle, gap149NirkanaAssassinOracle, gap149OssuaryRatsOracle,
		gap149PhyrexianPlaguelordOracle, gap149PitKeeperOracle, gap149PlagueSpitterOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestGap149LurkingDeadeyeDestroysADamagedCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := gap149Seats(g)
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt Bear", "Creature — Bear", 4, 4)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh Bear", "Creature — Bear", 4, 4)
	effectstest.B27Damage(g, fresh, hurt, 1)
	// Only a damaged creature is a legal target.
	g.WithWriteLock(func() {})
	effectstest.CastHoldCreature(t, g, "Lurking Deadeye", "Creature — Human Assassin", gap149LurkingDeadeyeOracle, 4, 2)
	effectstest.AnswerPickTargetIfOpen(t, g, hurt)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature survived")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("the undamaged creature was destroyed")
	}
	_ = me
}

func TestGap149LurkingDeadeyeCantTargetAnUndamagedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := gap149Seats(g)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh Bear", "Creature — Bear", 4, 4)
	effectstest.CastHoldCreature(t, g, "Lurking Deadeye", "Creature — Human Assassin", gap149LurkingDeadeyeOracle, 4, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fresh) {
		t.Error("an undamaged creature was destroyed")
	}
}

func TestGap149MaraudingBoneslasherBlocksOnlyWithAnotherZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	slasher := effectstest.B12Push(g, opp.ID, "Marauding Boneslasher", "Creature — Zombie Minotaur", gap149MaraudingBoneslasherOracl, 3, 3)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(slasher, attacker); err == nil {
		t.Fatal("Marauding Boneslasher blocked without another Zombie")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	attacker = effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	slasher = effectstest.B12Push(g, opp.ID, "Marauding Boneslasher", "Creature — Zombie Minotaur", gap149MaraudingBoneslasherOracl, 3, 3)
	effectstest.B12Creature(g, opp.ID, "Walking Corpse", "Creature — Zombie", 2, 2)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(slasher, attacker); err != nil {
		t.Fatalf("Marauding Boneslasher can't block beside another Zombie: %v", err)
	}
}

func TestGap149NestedGhoulMakesAZombieWhenDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := gap149Seats(g)
	ghoul := effectstest.B12Push(g, me.ID, "Nested Ghoul", "Creature — Phyrexian Zombie Warrior", gap149NestedGhoulOracle, 4, 2)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.B27Damage(g, src, ghoul, 1)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Phyrexian Zombie" && c.IsToken() {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Phyrexian Zombie tokens, want 1", n)
	}
}

func TestGap149NezumiBoneReaderDiscardsAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := gap149Seats(g)
	reader := effectstest.B12Push(g, me.ID, "Nezumi Bone-Reader", "Creature — Rat Shaman", gap149NezumiBoneReaderOracle, 1, 1)
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
	pitch := effectstest.HandCardForTest(opp, "Pitch Me", "Sorcery", "")
	effectstest.ApaMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, reader, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{victim},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, pitch)
	if !opp.Graveyard.Contains(pitch) {
		t.Error("the target did not discard")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the sacrificed creature is still here")
	}
	// Not at instant speed.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.ApaMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, reader, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{reader},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestGap149NirkanaAssassinGainsDeathtouchOnLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap149Seats(g)
	assassin := effectstest.B12Push(g, me.ID, "Nirkana Assassin", "Creature — Vampire Assassin Ally", gap149NirkanaAssassinOracle, 2, 3)
	if effectstest.HasEffectiveKeyword(t, g, assassin, "deathtouch") {
		t.Fatal("deathtouch before any life gain")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(assassin, me.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, assassin, "deathtouch") {
		t.Error("no deathtouch after gaining life")
	}
}

func TestGap149OssuaryRatsDamagesByCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := gap149Seats(g)
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardTyped(me, "Dead", "Creature — Bear")
	}
	effectstest.PushGraveyardCardTyped(me, "Spell", "Sorcery")
	target := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 5, 5)
	effectstest.CastHoldCreature(t, g, "Ossuary Rats", "Creature — Rat", gap149OssuaryRatsOracle, 3, 2)
	effectstest.AnswerPickTargetIfOpen(t, g, target)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, target).DamageMarked; got != 3 {
		t.Errorf("target has %d damage, want 3", got)
	}
}

func TestGap149PhyrexianPlaguelordBothSacrificeAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := gap149Seats(g)
	lord := effectstest.B12Push(g, me.ID, "Phyrexian Plaguelord", "Creature — Phyrexian Carrier", gap149PhyrexianPlaguelordOracle, 4, 4)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 4, 4)
	effectstest.ApaActivate(t, g, me, lord, 1, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	})
	if !g.Battlefield.Contains(big) {
		t.Fatal("a -1/-1 killed a 4/4")
	}
	effectstest.ApaActivate(t, g, me, lord, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	})
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(big) {
		t.Error("-4/-4 on a 3/3 left it alive")
	}
	if g.Battlefield.Contains(lord) {
		t.Error("the Plaguelord was not sacrificed")
	}
}

func TestGap149PitKeeperNeedsFourCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap149Seats(g)
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardTyped(me, "Dead", "Creature — Bear")
	}
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Pit Keeper", "Creature — Human Wizard", gap149PitKeeperOracle, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d with three creature cards, want %d", got, hand)
	}

	g = effectstest.NewCatalogGame(t)
	me, _ = gap149Seats(g)
	var last uuid.UUID
	for i := 0; i < 4; i++ {
		last = effectstest.PushGraveyardCardTyped(me, "Dead", "Creature — Bear")
	}
	effectstest.CastHoldCreature(t, g, "Pit Keeper", "Creature — Human Wizard", gap149PitKeeperOracle, 2, 1)
	effectstest.PickCard(t, g, me.ID, last)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(last) {
		t.Error("Pit Keeper did not return the creature card with four in the graveyard")
	}
}

func TestGap149PlagueSpitterPingsOnUpkeepAndWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := gap149Seats(g)
	spitter := effectstest.B12Push(g, me.ID, "Plague Spitter", "Creature — Phyrexian Horror", gap149PlagueSpitterOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	myLife, oppLife := me.Life, opp.Life
	effectstest.B27Kill(g, spitter)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-1 || opp.Life != oppLife-1 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, myLife-1, oppLife-1)
	}
	if got := effectstest.FindCardAnywhere(t, g, bear).DamageMarked; got != 1 {
		t.Errorf("the bear has %d damage, want 1", got)
	}
}

func TestGap149PlagueSpitterUpkeepTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := gap149Seats(g)
	effectstest.B12Push(g, me.ID, "Plague Spitter", "Creature — Phyrexian Horror", gap149PlagueSpitterOracle, 2, 2)
	myLife, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID, Step: game.StepUpkeep}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-1 || opp.Life != oppLife-1 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, myLife-1, oppLife-1)
	}
	// Not on an opponent's upkeep.
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: opp.ID, Step: game.StepUpkeep}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-1 {
		t.Errorf("it pinged on an opponent's upkeep (%d)", me.Life)
	}
}

// gap149Seats is the active seat and the next one.
func gap149Seats(g *game.Game) (me, opp *game.Player) {
	seat := g.Turn.ActiveSeat
	return g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}
