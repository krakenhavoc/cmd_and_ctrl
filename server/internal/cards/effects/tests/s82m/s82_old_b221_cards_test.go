package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b221_cards_test.go — Ixalan, batch 221 (#5532). Each test plays the
// real catalog entry and asserts what its text says.

const (
	oldB221Keeper   = "183e2796-2fc1-4f0e-b1be-95285d2ea046"
	oldB221Deacon   = "5212f0a8-7341-4960-835b-2f4600ca1b2d"
	oldB221Stampede = "3d6af1ee-16d2-4b82-8d9f-6d956603a746"
	oldB221Warrior  = "2af1930c-a5f2-4e7e-910f-e84458b051ac"
	oldB221Roar     = "40ee79dd-cf53-415c-8e3c-d1a125a735db"
	oldB221Plunder  = "597afe7c-a4ba-48d0-9c37-d9affe5ba195"
	oldB221Raptors  = "5f554df8-c70f-4212-aec7-597fe8a88553"
	oldB221Speaker  = "68b26aeb-4952-4ca1-9b4a-82bdca16461b"
)

func TestOldB221CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB221Keeper: "Ixalli's Keeper", oldB221Deacon: "Anointed Deacon", oldB221Stampede: "Dinosaur Stampede",
		oldB221Warrior: "Deeproot Warrior", oldB221Roar: "Rallying Roar", oldB221Plunder: "Deadeye Plunderers",
		oldB221Raptors: "Thrash of Raptors", oldB221Speaker: "Kumena's Speaker",
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

func oldB221Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB221Attack(g *game.Game, ids []uuid.UUID, defender uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			for _, id := range ids {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].AttackingTarget = defender
				}
			}
		}
	})
}

func TestOldB221ThrashOfRaptorsGrowsBesideAnotherDinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB221Seats(g)
	thrash := effectstest.B12Push(g, me.ID, "Thrash of Raptors", "Creature — Dinosaur", oldB221Raptors, 3, 3)
	if effectstest.EffectivePower(t, g, thrash) != 3 || effectstest.HasEffectiveKeyword(t, g, thrash, "trample") {
		t.Fatal("alone it is a plain 3/3: it is not its own other Dinosaur")
	}
	effectstest.B12Creature(g, opp.ID, "Their Dino", "Creature — Dinosaur", 1, 1)
	if effectstest.EffectivePower(t, g, thrash) != 3 {
		t.Error("an opponent's Dinosaur gave the bonus")
	}
	pal := effectstest.B12Creature(g, me.ID, "Pal", "Creature — Dinosaur", 1, 1)
	if effectstest.EffectivePower(t, g, thrash) != 5 || !effectstest.HasEffectiveKeyword(t, g, thrash, "trample") {
		t.Fatal("beside a Dinosaur it is not 5/3 with trample")
	}
	if effectstest.EffectiveToughness(t, g, thrash) != 3 {
		t.Error("the bonus is +2/+0")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pal) })
	g.RunStateChecksForTest()
	if effectstest.EffectivePower(t, g, thrash) != 3 || effectstest.HasEffectiveKeyword(t, g, thrash, "trample") {
		t.Error("the bonus outlived the other Dinosaur")
	}
}

func TestOldB221KumenasSpeakerNeedsAnotherMerfolkOrAnyIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB221Seats(g)
	sp := effectstest.B12Push(g, me.ID, "Kumena's Speaker", "Creature — Merfolk Shaman", oldB221Speaker, 1, 1)
	if p, tt := effectstest.PtOf(t, g, sp); p != 1 || tt != 1 {
		t.Fatalf("alone it is %d/%d, want 1/1 (it is not its own other Merfolk)", p, tt)
	}
	effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	effectstest.B12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 1, 1)
	if p, _ := effectstest.PtOf(t, g, sp); p != 1 {
		t.Error("an opponent's Island or Merfolk gave the bonus")
	}
	island := effectstest.B12Permanent(g, me.ID, "My Island", "Basic Land — Island")
	if p, tt := effectstest.PtOf(t, g, sp); p != 2 || tt != 2 {
		t.Errorf("with an Island it is %d/%d, want 2/2", p, tt)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(island) })
	g.RunStateChecksForTest()
	if p, _ := effectstest.PtOf(t, g, sp); p != 1 {
		t.Error("the bonus outlived the Island")
	}
	effectstest.B12Creature(g, me.ID, "Pal", "Creature — Merfolk", 1, 1)
	if p, tt := effectstest.PtOf(t, g, sp); p != 2 || tt != 2 {
		t.Errorf("with another Merfolk it is %d/%d, want 2/2", p, tt)
	}
}

func TestOldB221DeadeyePlunderersCountsArtifactsAndMakesTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB221Seats(g)
	pl := effectstest.B12Push(g, me.ID, "Deadeye Plunderers", "Creature — Human Pirate", oldB221Plunder, 3, 3)
	if p, tt := effectstest.PtOf(t, g, pl); p != 3 || tt != 3 {
		t.Fatalf("with no artifacts it is %d/%d, want 3/3", p, tt)
	}
	effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 1, 1)
	if p, tt := effectstest.PtOf(t, g, pl); p != 5 || tt != 5 {
		t.Fatalf("with two artifacts it is %d/%d, want 5/5", p, tt)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B10AddMana(me, "C", "C", "U", "B")
	if err := g.ActivateCatalogAbility(me.ID, pl, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTreasuresControlledBy(g, me.ID); n != 1 {
		t.Fatalf("%d Treasures, want 1", n)
	}
	if p, _ := effectstest.PtOf(t, g, pl); p != 6 {
		t.Errorf("the new Treasure is an artifact: power %d, want 6", p)
	}
	if err := g.ActivateCatalogAbility(me.ID, pl, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestOldB221DeeprootWarriorGrowsWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB221Seats(g)
	w := effectstest.B12Push(g, me.ID, "Deeproot Warrior", "Creature — Merfolk Warrior", oldB221Warrior, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	oldB221Attack(g, []uuid.UUID{w}, opp.ID)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: blocker, CardID: w, Target: w})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, w); p != 3 || tt != 3 {
		t.Errorf("after being blocked it is %d/%d, want 3/3", p, tt)
	}
}

func TestOldB221DinosaurStampedePumpsAttackersAndGivesDinosaursTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB221Seats(g)
	dinoAtk := effectstest.B12Creature(g, me.ID, "Dino Attacker", "Creature — Dinosaur", 2, 2)
	bearAtk := effectstest.B12Creature(g, me.ID, "Bear Attacker", "Creature — Bear", 2, 2)
	dinoIdle := effectstest.B12Creature(g, me.ID, "Dino Idle", "Creature — Dinosaur", 2, 2)
	theirDino := effectstest.B12Creature(g, opp.ID, "Their Dino", "Creature — Dinosaur", 2, 2)
	oldB221Attack(g, []uuid.UUID{dinoAtk, bearAtk}, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Dinosaur Stampede", "Instant", oldB221Stampede, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{dinoAtk: 4, bearAtk: 4, dinoIdle: 2, theirDino: 2} {
		if p, tt := effectstest.PtOf(t, g, id); p != want || tt != 2 {
			t.Errorf("%v is %d/%d, want %d/2", id, p, tt, want)
		}
	}
	for id, want := range map[uuid.UUID]bool{dinoAtk: true, bearAtk: false, dinoIdle: true, theirDino: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "trample"); got != want {
			t.Errorf("%v trample = %v, want %v", id, got, want)
		}
	}
}

func TestOldB221RallyingRoarPumpsAndUntapsYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB221Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].Tapped = true
		}
	})
	effectstest.CastCatalogSpell(t, g, "Rallying Roar", "Instant", oldB221Roar, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("mine is %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("theirs is %d/%d, want 2/2", p, tt)
	}
	if _, c := effectstest.CardWhere(g, me, mine); c.Tapped {
		t.Error("my creature is still tapped")
	}
	if _, c := effectstest.CardWhere(g, opp, theirs); !c.Tapped {
		t.Error("their creature was untapped")
	}
}

func TestOldB221IxallisKeeperSacrificesToPumpAndGiveTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB221Seats(g)
	keeper := effectstest.B12Push(g, me.ID, "Ixalli's Keeper", "Creature — Human Shaman", oldB221Keeper, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	tgt := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Targets: tgt, Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.B10AddMana(me, "G", "G", "G", "G", "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Targets: tgt, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(keeper) {
		t.Error("the Keeper was not sacrificed")
	}
	if p, tt := effectstest.PtOf(t, g, bear); p != 7 || tt != 7 || !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Errorf("the target is %d/%d, want 7/7 with trample", p, tt)
	}
}

func TestOldB221AnointedDeaconPumpsATargetVampireAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB221Seats(g)
	effectstest.B12Push(g, me.ID, "Anointed Deacon", "Creature — Vampire Cleric", oldB221Deacon, 3, 3)
	bat := effectstest.B12Creature(g, me.ID, "Bat", "Creature — Vampire", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bat)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bat); p != 3 || tt != 1 {
		t.Errorf("the Vampire is %d/%d, want 3/1", p, tt)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("a non-Vampire got the pump: power %d", p)
	}
}

func TestOldB221AnointedDeaconMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB221Seats(g)
	effectstest.B12Push(g, me.ID, "Anointed Deacon", "Creature — Vampire Cleric", oldB221Deacon, 3, 3)
	bat := effectstest.B12Creature(g, me.ID, "Bat", "Creature — Vampire", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bat)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bat); p != 1 {
		t.Errorf("declined, yet the Vampire has power %d", p)
	}
}
