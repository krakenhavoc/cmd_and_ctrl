package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b743_test.go — Exodus, S82 OLD batch 743 (#5626).

const (
	oldB743Grollub    = "22c8c952-2713-418e-8259-aea09b4e740b"
	oldB743Zealots    = "3fa67c92-6328-4d0a-95e3-656d447ab20e"
	oldB743KeeperLite = "62d1300f-2b76-42b9-8e6c-f920712cd547"
	oldB743Carnophage = "9ac77cfd-7d78-4c8a-bffc-17649a852114"
	oldB743Convalesce = "9ad86d0e-3363-4d98-b0bb-dc491be1ddd5"
	oldB743ManaBreach = "221d0305-94db-4791-9dc7-89ee96c7fcd6"
	oldB743Song       = "4c543799-1b72-4f32-a5be-475c0134ac0f"
	oldB743Lieges     = "87038354-6762-4069-be34-7d40ae308a8d"
	oldB743Onslaught  = "ae9ca82c-e07e-4a41-a387-0ef7d6df14b6"
)

func TestOldB743CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB743Grollub: "Grollub", oldB743Zealots: "Zealots en-Dal", oldB743KeeperLite: "Keeper of the Light",
		oldB743Carnophage: "Carnophage", oldB743Convalesce: "Convalescence", oldB743ManaBreach: "Mana Breach",
		oldB743Song: "Song of Serenity", oldB743Lieges: "Oath of Lieges", oldB743Onslaught: "Onslaught",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func TestOldB743GrollubGivesEachOpponentTheDamageItTook(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grollub := effectstest.B12Push(g, me.ID, "Grollub", "Creature — Beast", oldB743Grollub, 3, 3)
	foe := effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.B27Damage(g, foe, grollub, 2)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want += 2
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", p.Seat, p.Life, want)
		}
	}
}

func TestOldB743ZealotsGainsLifeOnlyWhileEveryNonlandPermanentIsWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Zealots en-Dal", "Creature — Human Soldier", oldB743Zealots, "{3}{W}", 2, 4, "W")
	effectstest.AcPermanent(g, me.ID, "Plains", "Basic Land — Plains", 0, 0)
	effectstest.AcPermanent(g, me.ID, "Angel", "Creature — Angel", 2, 2, "W", "U")
	life := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 1 {
		t.Errorf("gained %d life with every nonland permanent white, want 1", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.B27Push(g2, me2.ID, "Zealots en-Dal", "Creature — Human Soldier", oldB743Zealots, "{3}{W}", 2, 4, "W")
	effectstest.AcPermanent(g2, me2.ID, "Rock", "Artifact", 0, 0)
	life2 := me2.Life
	effectstest.AdvanceToUpkeepOf(t, g2, 0)
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Life != life2 {
		t.Errorf("gained life with a colorless artifact on the table: %d → %d", life2, me2.Life)
	}
}

func TestOldB743KeeperOfTheLightNeedsAnOpponentWithMoreLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Keeper of the Light", "Creature — Human Wizard", oldB743KeeperLite, false)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	g.WithWriteLock(func() { me.Life, opp.Life = 20, 20 })
	effectstest.DcMana(t, g, me.ID, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("an opponent with equal life was accepted")
	}
	g.WithWriteLock(func() { me.Life, opp.Life = 10, 11 })
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("targeting yourself was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The life totals are read as the ability is activated: the gain happens
	// though the opponent's life drops below mine in response.
	g.WithWriteLock(func() { opp.Life = 1 })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 13 {
		t.Errorf("life %d, want 13", me.Life)
	}
}

func TestOldB743CarnophageTapsUnlessYouPayALife(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.B27Push(g, me.ID, "Carnophage", "Creature — Zombie", oldB743Carnophage, "{B}", 2, 2, "B")
		life := me.Life
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		answer := 0 // tap it
		if pay {
			answer = 1
		}
		effectstest.AnswerOptionPick(t, g, me.ID, answer)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.AlliTapped(t, g, id); got == pay {
			t.Errorf("pay=%v: tapped=%v", pay, got)
		}
		if got := life - me.Life; (got == 1) != pay {
			t.Errorf("pay=%v: life %d → %d", pay, life, me.Life)
		}
	}
}

func TestOldB743ConvalescenceGainsLifeAtTenOrLess(t *testing.T) {
	for _, tc := range []struct{ start, want int }{{10, 11}, {5, 6}, {11, 11}, {20, 20}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushPermanentForTest(g, me.ID, "Convalescence", oldB743Convalesce, "Enchantment")
		g.WithWriteLock(func() { me.Life = tc.start })
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != tc.want {
			t.Errorf("start %d: life %d, want %d", tc.start, me.Life, tc.want)
		}
	}
}

func TestOldB743ManaBreachReturnsALandOfTheCaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushPermanentForTest(g, me.ID, "Mana Breach", oldB743ManaBreach, "Enchantment")
	myForest := effectstest.AcPermanent(g, me.ID, "Forest", "Basic Land — Forest", 0, 0)
	myIsland := effectstest.AcPermanent(g, me.ID, "Island", "Basic Land — Island", 0, 0)
	theirs := effectstest.AcPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp", 0, 0)
	spell := effectstest.HandCardOf(g, me, "Shock", "Instant", "{R}", "")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 1 || prompt.ChooseMax != 1 {
		t.Fatalf("the caster was not asked for a land: %+v", prompt)
	}
	if err := g.ResolveOwnPermanents(prompt.ID, me.ID, []uuid.UUID{theirs}); err == nil {
		t.Error("an opponent's land was accepted")
	}
	if err := g.ResolveOwnPermanents(prompt.ID, me.ID, []uuid.UUID{myIsland}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(myIsland) || !me.Hand.Contains(myIsland) {
		t.Error("the chosen land did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(myForest) || !g.Battlefield.Contains(theirs) {
		t.Error("another land left the battlefield")
	}
}

func TestOldB743ManaBreachAlsoHitsAnOpponentsCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushPermanentForTest(g, me.ID, "Mana Breach", oldB743ManaBreach, "Enchantment")
	mine := effectstest.AcPermanent(g, me.ID, "Forest", "Basic Land — Forest", 0, 0)
	theirs := effectstest.AcPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp", 0, 0)
	spell := effectstest.HandCardOf(g, opp, "Shock", "Instant", "{R}", "")
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The caster is the opponent, and their land is the only candidate.
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
	if prompt == nil {
		t.Fatal("the opponent was not asked for a land")
	}
	if err := g.ResolveOwnPermanents(prompt.ID, opp.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the opponent's land did not return to their hand")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the Breach's controller lost a land to an opponent's spell")
	}
}

func TestOldB743SongOfSerenityStopsEnchantedCreaturesAttackingAndBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Song of Serenity", oldB743Song, "Enchantment")
	enchanted := effectstest.B12Creature(g, opp.ID, "Enchanted", "Creature — Bear", 2, 2)
	equipped := effectstest.B12Creature(g, opp.ID, "Equipped", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	aura := effectstest.AcPermanent(g, me.ID, "Pacifism-like", "Enchantment — Aura", 0, 0)
	sword := effectstest.AcPermanent(g, me.ID, "Sword", "Artifact — Equipment", 0, 0)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: enchanted}
		effectstest.FindBattlefieldCardForTest(g, sword).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: equipped}
	})
	both := game.CantAttack | game.CantBlock
	if got := effectstest.RestrictionsOf(t, g, enchanted); got&both != both {
		t.Errorf("an enchanted creature has restrictions %v, want can't attack or block", got)
	}
	for name, id := range map[string]uuid.UUID{"equipped": equipped, "plain": plain} {
		if got := effectstest.RestrictionsOf(t, g, id); got&both != 0 {
			t.Errorf("a %s creature has restrictions %v", name, got)
		}
	}
	// It ends with the Aura.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	if got := effectstest.RestrictionsOf(t, g, enchanted); got&both != 0 {
		t.Errorf("a creature whose Aura left still has restrictions %v", got)
	}
}

func TestOldB743OathOfLiegesUpkeepPlayerMaySearchForABasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep, poor, rich := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Lieges", oldB743Lieges, "Enchantment")
	effectstest.AcPermanent(g, upkeep.ID, "Lone Forest", "Basic Land — Forest", 0, 0)
	for i := 0; i < 3; i++ {
		effectstest.AcPermanent(g, owner.ID, "Owner Land", "Basic Land — Plains", 0, 0)
	}
	for i := 0; i < 2; i++ {
		effectstest.AcPermanent(g, rich.ID, "Rich Land", "Basic Land — Island", 0, 0)
	}
	_ = poor
	ids := effectstest.SeedSearchLibrary(upkeep,
		game.Card{Name: "Nonbasic", TypeLine: "Land"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
	)
	mountain := ids[1]
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	ch := effectstest.PickTargetPrompt(t, g)
	if ch.Chooser != upkeep.ID {
		t.Fatalf("the prompt is for %s, want the upkeep player", ch.Chooser)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range ch.PickTargetPlayers {
		offered[id] = true
	}
	if !offered[owner.ID] || !offered[rich.ID] || offered[upkeep.ID] || offered[poor.ID] || len(offered) != 2 {
		t.Fatalf("offered %v, want only the opponents with more lands than the upkeep player", ch.PickTargetPlayers)
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: poor.ID}); err == nil {
		t.Fatal("a player with fewer lands was accepted")
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: owner.ID}); err != nil {
		t.Fatalf("picking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, upkeep, true)
	c := effectstest.SearchChoiceFor(g, upkeep.ID)
	if c == nil || len(c.SearchCards) != 1 {
		t.Fatalf("search prompt %+v, want only the basic land card offered", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, upkeep.ID, []uuid.UUID{mountain}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !g.Battlefield.Contains(mountain) {
		t.Error("the basic land did not enter the battlefield")
	}
	if effectstest.FindBattlefieldCardForTest(g, mountain).Controller != upkeep.ID {
		t.Error("the land entered under another player's control")
	}
}

func TestOldB743OathOfLiegesHasNoTargetWithoutALandLead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Lieges", oldB743Lieges, "Enchantment")
	effectstest.AcPermanent(g, upkeep.ID, "Forest", "Basic Land — Forest", 0, 0)
	effectstest.AcPermanent(g, owner.ID, "Plains", "Basic Land — Plains", 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.LatestPickTarget(g, upkeep.ID); p != nil {
		t.Fatalf("asked for a target with nobody holding more lands: %+v", p)
	}
}

func TestOldB743OnslaughtTapsATargetWhenYouCastACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushPermanentForTest(g, me.ID, "Onslaught", oldB743Onslaught, "Enchantment")
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	// A noncreature spell does nothing.
	bolt := effectstest.HandCardOf(g, me, "Shock", "Instant", "{R}", "")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("a noncreature spell triggered Onslaught")
	}
	effectstest.CastCatalogSpell(t, g, "Bear Cub", "Creature — Bear", "", nil)
	effectstest.PickCard(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AlliTapped(t, g, foe) {
		t.Error("the target was not tapped")
	}
	if effectstest.AlliTapped(t, g, other) {
		t.Error("an untargeted creature was tapped")
	}
}
