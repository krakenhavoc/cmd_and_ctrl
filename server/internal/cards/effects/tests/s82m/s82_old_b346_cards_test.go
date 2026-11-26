package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b346_cards_test.go — Khans of Tarkir, batch 346 (S82 older sets,
// #5554): the +1/+1 counter cards. Every test plays the real catalog entry
// through the engine.

const (
	oldB346Fortress  = "10ae7744-5f98-4d6f-993a-ffb2a34b23e7"
	oldB346Bloodsuck = "511e35cd-2e19-4803-a404-b790af0cea60"
	oldB346Boon      = "629b3aee-b8fd-495a-a988-815a0aaf3e82"
	oldB346Growth    = "802ff367-e52e-4f79-aa0f-96c45fe477a0"
	oldB346Vizier    = "9bdef081-1408-4125-82c3-04296dee05d5"
	oldB346Purse     = "bb2cc0f4-6abd-48d6-b442-76eafbfd2309"
	oldB346BringLow  = "cfe4d707-41b3-4897-b13a-dcae8f1f3350"
	oldB346Trap      = "d64adb26-62c4-44aa-bc2e-d05bd1cf657b"
	oldB346Anafenza  = "ebfafabc-5255-4024-be59-7403bdb16ee4"
)

func TestOldB346EveryCardIsRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Ivorytusk Fortress": oldB346Fortress, "Kheru Bloodsucker": oldB346Bloodsuck, "Dragonscale Boon": oldB346Boon,
		"Incremental Growth": oldB346Growth, "Rakshasa Vizier": oldB346Vizier, "Briber's Purse": oldB346Purse,
		"Bring Low": oldB346BringLow, "Trap Essence": oldB346Trap, "Anafenza, the Foremost": oldB346Anafenza,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not in the catalog", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %d, want Full with none", name, spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func oldB346Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB346Counters(g *game.Game, id uuid.UUID) int {
	return effectstest.CounterOn(g, id, game.CounterPlusOne)
}

func TestOldB346IvorytuskFortressUntapsCreaturesWithCountersOnOthersUntapSteps(t *testing.T) {
	g, me, _ := oldB346Table(t)
	effectstest.B12Push(g, me.ID, "Ivorytusk Fortress", "Creature — Elephant", oldB346Fortress, 5, 7)
	counted := effectstest.B12Creature(g, me.ID, "Counted Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	effectstest.AddCountersForTest(g, counted, game.CounterPlusOne, 1)
	effectstest.TapForTest(t, g, counted, plain)
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, counted) {
		t.Error("the creature with a +1/+1 counter stayed tapped in another player's untap step")
	}
	if !effectstest.B16Tapped(t, g, plain) {
		t.Error("a creature without a counter untapped in another player's untap step")
	}
}

func TestOldB346KheruBloodsuckerDrainsWhenABigCreatureYouControlDies(t *testing.T) {
	g, me, opp := oldB346Table(t)
	effectstest.B12Push(g, me.ID, "Kheru Bloodsucker", "Creature — Vampire", oldB346Bloodsuck, 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	small := effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	myLife, oppLife := me.Life, opp.Life

	effectstest.KillCreature(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("a toughness-3 creature or an opponent's creature drained: %d/%d", me.Life, opp.Life)
	}
	effectstest.KillCreature(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+2 || opp.Life != oppLife-2 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, myLife+2, oppLife-2)
	}
}

func TestOldB346KheruBloodsuckerReadsLastKnownToughness(t *testing.T) {
	g, me, opp := oldB346Table(t)
	effectstest.B12Push(g, me.ID, "Kheru Bloodsucker", "Creature — Vampire", oldB346Bloodsuck, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	effectstest.AddCountersForTest(g, bear, game.CounterPlusOne, 2)
	oppLife := opp.Life
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("a 4/4 (2/2 with two counters) dying drained %d, want 2", oppLife-opp.Life)
	}
}

func TestOldB346KheruBloodsuckerSacrificesAnotherCreatureForACounter(t *testing.T) {
	g, me, _ := oldB346Table(t)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Kheru Bloodsucker", "Creature — Vampire", oldB346Bloodsuck, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("B", "C", "C")...) })
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id}}); err == nil {
		t.Fatal("sacrificed itself for \"another creature\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB346Counters(g, id) != 1 {
		t.Errorf("counters = %d, want 1", oldB346Counters(g, id))
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
}

func TestOldB346DragonscaleBoonPutsTwoCountersAndUntaps(t *testing.T) {
	g, me, _ := oldB346Table(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Dragonscale Boon", "Instant", "", oldB346Boon, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if oldB346Counters(g, bear) != 2 {
		t.Errorf("counters = %d, want 2", oldB346Counters(g, bear))
	}
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature is still tapped")
	}
}

func TestOldB346IncrementalGrowthPutsOneTwoAndThreeOnThreeDifferentCreatures(t *testing.T) {
	g, me, _ := oldB346Table(t)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: c}}
	if _, err := effectstest.CastFromHandErr(g, me, "Incremental Growth", "Sorcery", "", oldB346Growth, game.CastSpellParams{Strict: true, Targets: targets}); err == nil {
		t.Fatal("the same creature was accepted for two target clauses")
	}
	targets[1].ID = b
	effectstest.CastFromHandForTest(t, g, me, "Incremental Growth", "Sorcery", "", oldB346Growth, game.CastSpellParams{Strict: true, Targets: targets})
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 1, b: 2, c: 3} {
		if got := oldB346Counters(g, id); got != want {
			t.Errorf("counters = %d, want %d", got, want)
		}
	}
}

func TestOldB346RakshasaVizierGrowsPerCardExiledFromYourGraveyard(t *testing.T) {
	g, me, opp := oldB346Table(t)
	vizier := effectstest.B12Push(g, me.ID, "Rakshasa Vizier", "Creature — Demon", oldB346Vizier, 4, 4)
	mine1 := effectstest.GraveyardCardOf(g, me, "Mine One", "Creature — Bear", "", "")
	mine2 := effectstest.GraveyardCardOf(g, me, "Mine Two", "Sorcery", "", "")
	theirs := effectstest.GraveyardCardOf(g, opp, "Theirs", "Creature — Bear", "", "")
	move := func(owner *game.Player, id uuid.UUID, to game.ZoneKind) {
		t.Helper()
		dest := game.ZoneRef{Kind: to, Owner: owner.ID}
		if to == game.ZoneExile {
			dest = game.ZoneRef{Kind: to}
		}
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner.ID}, dest, id); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	move(opp, theirs, game.ZoneExile)
	if oldB346Counters(g, vizier) != 0 {
		t.Fatal("an opponent's graveyard card counted")
	}
	move(me, mine1, game.ZoneHand)
	if oldB346Counters(g, vizier) != 0 {
		t.Fatal("a card returned to hand counted as exiled")
	}
	move(me, mine2, game.ZoneExile)
	if oldB346Counters(g, vizier) != 1 {
		t.Errorf("counters = %d, want 1", oldB346Counters(g, vizier))
	}
}

func TestOldB346BribersPurseEntersWithXGemCountersAndSpendsThem(t *testing.T) {
	g, me, opp := oldB346Table(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastFromHandForTest(t, g, me, "Briber's Purse", "Artifact", "{X}", oldB346Purse, game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, id, effects.OldB346GemCounter); got != 2 {
		t.Fatalf("gem counters = %d, want 2", got)
	}
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("C")...) })
	if err := effectstest.ActivateOn(t, g, me.ID, id, 0, theirs); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, id, effects.OldB346GemCounter); got != 1 {
		t.Errorf("gem counters = %d, want 1 after one use", got)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardForTest(g, theirs)
	if c.Effective().Restrictions&game.CantAttack == 0 || c.Effective().Restrictions&game.CantBlock == 0 {
		t.Error("the target can still attack or block")
	}
}

func TestOldB346BringLowDealsThreeOrFiveWithACounter(t *testing.T) {
	g, me, opp := oldB346Table(t)
	plain := effectstest.B12Creature(g, opp.ID, "Plain", "Creature — Wall", 0, 9)
	counted := effectstest.B12Creature(g, opp.ID, "Counted", "Creature — Wall", 0, 7)
	effectstest.AddCountersForTest(g, counted, game.CounterPlusOne, 1)
	effectstest.AdvanceToMain(t, g)
	for _, c := range []struct {
		id   uuid.UUID
		want int
	}{{plain, 3}, {counted, 5}} {
		effectstest.CastFromHandForTest(t, g, me, "Bring Low", "Instant", "", oldB346BringLow, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(c.id)})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.Pr6Marked(g, c.id); got != c.want {
			t.Errorf("damage = %d, want %d", got, c.want)
		}
	}
}

func TestOldB346TrapEssenceCountersACreatureSpellAndGrowsACreature(t *testing.T) {
	g, me, _ := oldB346Table(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	oppSpell := effectstest.CastFromHandForTest(t, g, me, "Bear Spell", "Creature — Bear", "", "", game.CastSpellParams{})
	effectstest.CastFromHandForTest(t, g, me, "Trap Essence", "Instant", "", oldB346Trap, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: oppSpell}, {Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(oppSpell) {
		t.Error("the creature spell resolved")
	}
	if oldB346Counters(g, bear) != 2 {
		t.Errorf("counters = %d, want 2", oldB346Counters(g, bear))
	}
}

func TestOldB346AnafenzaPutsACounterOnAnotherTappedCreatureWhenAttacking(t *testing.T) {
	g, me, opp := oldB346Table(t)
	effectstest.AdvanceToMain(t, g)
	anafenza := effectstest.B12Push(g, me.ID, "Anafenza, the Foremost", "Legendary Creature — Human Soldier", oldB346Anafenza, 4, 4)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	untapped := effectstest.B12Creature(g, me.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, tapped)
	effectstest.DeclareAttack(t, g, opp.ID, anafenza)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target was asked for")
	}
	if effectstest.HasID(p.PickTargetCards, untapped) || effectstest.HasID(p.PickTargetCards, anafenza) {
		t.Errorf("offered %v: an untapped creature or Anafenza itself", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, tapped)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB346Counters(g, tapped) != 1 {
		t.Errorf("counters = %d, want 1", oldB346Counters(g, tapped))
	}
}

func TestOldB346AnafenzaExilesOpponentsCreatureCardsInsteadOfTheGraveyard(t *testing.T) {
	g, me, opp := oldB346Table(t)
	effectstest.B12Push(g, me.ID, "Anafenza, the Foremost", "Legendary Creature — Human Soldier", oldB346Anafenza, 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.KillCreature(t, g, me.ID, mine)
	if opp.Graveyard.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Error("an opponent's dying creature was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own creature was exiled")
	}
	// A creature card that was never on the battlefield: discarded from hand.
	card := effectstest.HandCardFull(opp, "Their Hand Bear", "Creature — Bear", "", "", nil)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, card); err != nil {
		t.Fatal(err)
	}
	if opp.Graveyard.Contains(card) || !g.Exile.Contains(card) {
		t.Error("an opponent's creature card from their hand was not exiled")
	}
	// A noncreature card still goes to the graveyard.
	spell := effectstest.HandCardFull(opp, "Their Sorcery", "Sorcery", "", "", nil)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, spell); err != nil {
		t.Fatal(err)
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("a noncreature card was exiled")
	}
}

func TestOldB346TrapEssenceRefusesANoncreatureSpell(t *testing.T) {
	g, me, _ := oldB346Table(t)
	effectstest.AdvanceToMain(t, g)
	sorcery := effectstest.CastFromHandForTest(t, g, me, "Plain Sorcery", "Sorcery", "", "", game.CastSpellParams{})
	_, err := effectstest.CastFromHandErr(g, me, "Trap Essence", "Instant", "", oldB346Trap, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: sorcery}}})
	if err == nil {
		t.Fatal("Trap Essence targeted a noncreature spell")
	}
}
