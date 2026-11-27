package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b738_cards_test.go — Exodus, "Damage", batch 738 (#5626). Every test
// plays the real catalog entry through the engine.

const (
	o738Pandemonium = "073551ca-2e84-4d49-8484-9890718cc483"
	o738Oath        = "2b749918-f425-4243-aabb-7350bb74403c"
	o738Spellshock  = "6579bb00-3886-447b-a265-3e72e98bbe2c"
	o738Keeper      = "8cf12b01-7c0a-488f-9527-404a545c70c4"
	o738PitSpawn    = "f536acf3-ff8e-48da-8509-00e7ce4a1678"
	o738Salamander  = "fcca3c9d-a2b6-4f29-b43c-2f061074a332"
)

func TestS82OldB738EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o738Pandemonium, o738Oath, o738Spellshock, o738Keeper, o738PitSpawn, o738Salamander} {
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

func o738SetLife(g *game.Game, p *game.Player, life int) {
	g.WithWriteLock(func() { p.Life = life })
}

func TestS82OldB738SpellshockDamagesAnyPlayerWhoCastsASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, g.Seats[1].ID, "Spellshock", "Enchantment", o738Spellshock, 0, 0)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage to the caster", life, me.Life)
	}
}

func TestS82OldB738KeeperOfTheFlameNeedsAnOpponentWithMoreLifeAsYouActivate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	keeper := effectstest.B12Push(g, me.ID, "Keeper of the Flame", "Creature — Human Wizard", o738Keeper, 1, 2)
	foe := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	o738SetLife(g, me, 20)
	o738SetLife(g, opp, 20)
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Targets: foe, Strict: true}); err == nil {
		t.Fatal("targeted an opponent with no more life than you")
	}
	o738SetLife(g, opp, 25)
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Targets: foe, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// As you activate: the totals changing in response does not stop it.
	o738SetLife(g, opp, 10)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 8 {
		t.Errorf("opponent life %d, want 8 (10 less 2)", opp.Life)
	}
}

func TestS82OldB738ScaldingSalamanderMayDamageEachDefendingCreatureWithoutFlying(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		sal := effectstest.B12Push(g, me.ID, "Scalding Salamander", "Creature — Salamander", o738Salamander, 2, 1)
		ground := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
		flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Their Bird", TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
			Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
		effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
		if err := g.DeclareAttacker(sal, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		want := 0
		if yes {
			want = 1
		}
		if n := effectstest.DamageMarkedOn(g, ground); n != want {
			t.Errorf("yes=%v: their grounded creature has %d damage, want %d", yes, n, want)
		}
		if n := effectstest.DamageMarkedOn(g, flier); n != 0 {
			t.Errorf("yes=%v: their flier was damaged", yes)
		}
		if n := effectstest.DamageMarkedOn(g, mine); n != 0 {
			t.Errorf("yes=%v: my own creature was damaged", yes)
		}
	}
}

func o738PitSpawnUpkeep(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spawn := effectstest.B12Push(g, me.ID, "Pit Spawn", "Creature — Demon", o738PitSpawn, 6, 4)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	return g, me, spawn
}

func TestS82OldB738PitSpawnStaysWhenYouPayAndIsSacrificedWhenYouDont(t *testing.T) {
	g, me, spawn := o738PitSpawnUpkeep(t)
	effectstest.B06AddMana(me, "B", "B")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt for the upkeep trigger")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spawn) {
		t.Error("Pit Spawn should stay after paying {B}{B}")
	}
	g2, me2, spawn2 := o738PitSpawnUpkeep(t)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.HasPayUnlessFor(g2, me2.ID) {
		effectstest.AnswerPayUnless(t, g2, me2.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(spawn2) {
		t.Error("Pit Spawn stayed without payment")
	}
}

func TestS82OldB738PitSpawnExilesACreatureItDamages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spawn := effectstest.B12Push(g, me.ID, "Pit Spawn", "Creature — Demon", o738PitSpawn, 6, 4)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.UnblockedAttackIntoBlockers(t, g, spawn)
	if err := g.DeclareBlocker(wall, spawn); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if z := g.FindCardZoneForEffect(wall); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the blocker is in %+v, want exile (it survived the 6 damage with 9 toughness)", z)
	}
}

func TestS82OldB738OathOfMagesLetsTheUpkeepPlayerPingAnOpponentWithMoreLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep, rich, poor := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Mages", o738Oath, "Enchantment")
	o738SetLife(g, owner, 15)
	o738SetLife(g, upkeep, 20)
	o738SetLife(g, rich, 30)
	o738SetLife(g, poor, 10)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	ch := effectstest.PickTargetPrompt(t, g)
	if ch.Chooser != upkeep.ID {
		t.Fatalf("the prompt is for %s, want the upkeep player", ch.Chooser)
	}
	if len(ch.PickTargetPlayers) != 1 || ch.PickTargetPlayers[0] != rich.ID {
		t.Fatalf("offered %v, want only the opponent with more life", ch.PickTargetPlayers)
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: poor.ID}); err == nil {
		t.Fatal("a player with less life was accepted")
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: rich.ID}); err != nil {
		t.Fatalf("picking the richer opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, upkeep, true)
	if rich.Life != 29 {
		t.Errorf("the chosen player has %d life, want 29", rich.Life)
	}
}

func TestS82OldB738OathOfMagesDoesNothingWithoutARicherOpponentAndMayBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Mages", o738Oath, "Enchantment")
	o738SetLife(g, owner, 10)
	o738SetLife(g, upkeep, 20)
	o738SetLife(g, g.Seats[2], 5)
	o738SetLife(g, g.Seats[3], 5)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.LatestPickTarget(g, upkeep.ID); p != nil {
		t.Fatalf("the upkeep player was asked for a target with nobody eligible: %+v", p)
	}
	g2 := effectstest.NewCatalogGame(t)
	owner2, upkeep2 := g2.Seats[0], g2.Seats[1]
	effectstest.PushPermanentForTest(g2, owner2.ID, "Oath of Mages", o738Oath, "Enchantment")
	o738SetLife(g2, owner2, 30)
	o738SetLife(g2, upkeep2, 20)
	effectstest.AdvanceToUpkeepOf(t, g2, 1)
	ch := effectstest.PickTargetPrompt(t, g2)
	if err := g2.ResolvePickTarget(ch.ID, upkeep2.ID, game.TargetRef{Kind: game.TargetPlayer, ID: owner2.ID}); err != nil {
		t.Fatalf("picking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerConfirmFor(t, g2, upkeep2, false)
	if owner2.Life != 30 {
		t.Errorf("declined, yet the owner has %d life", owner2.Life)
	}
}

func TestS82OldB738PandemoniumLetsTheEnteringCreaturesControllerChooseAndDeal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, me, victim := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushPermanentForTest(g, owner.ID, "Pandemonium", o738Pandemonium, "Enchantment")
	id := uuid.New()
	effectstest.PlayerByIDForTest(g, me.ID).Hand.PushTop(game.Card{
		InstanceID: id, Name: "Hill Giant", TypeLine: "Creature — Giant",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter: %v", err)
	}
	ch := effectstest.PickTargetPrompt(t, g)
	if ch.Chooser != me.ID {
		t.Fatalf("the prompt is for %s, want the entering creature's controller %s", ch.Chooser, me.ID)
	}
	life := victim.Life
	if err := g.ResolvePickTarget(ch.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID}); err != nil {
		t.Fatalf("picking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	if victim.Life != life-3 {
		t.Errorf("victim life %d → %d, want 3 damage (the creature's power)", life, victim.Life)
	}
}

func TestS82OldB738PandemoniumMayBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, victim := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Pandemonium", o738Pandemonium, "Enchantment")
	id := uuid.New()
	effectstest.PlayerByIDForTest(g, owner.ID).Hand.PushTop(game.Card{
		InstanceID: id, Name: "Hill Giant", TypeLine: "Creature — Giant",
		Power: 3, Toughness: 3, Owner: owner.ID, Controller: owner.ID,
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter: %v", err)
	}
	ch := effectstest.PickTargetPrompt(t, g)
	life := victim.Life
	if err := g.ResolvePickTarget(ch.ID, owner.ID, game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID}); err != nil {
		t.Fatalf("picking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, owner, false)
	if victim.Life != life {
		t.Errorf("declined, yet the victim lost %d life", life-victim.Life)
	}
}
