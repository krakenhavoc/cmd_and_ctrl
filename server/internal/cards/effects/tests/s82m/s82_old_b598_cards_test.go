package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b598_cards_test.go — Betrayers of Kamigawa, activated abilities,
// instants and sorceries, batch 598 (#5602, S82 OLD). Every test plays the
// real catalog entry through the engine.

const (
	oldB598Mindsweeper  = "2ef6f4d6-f755-4402-8339-2a9b6c86f96c"
	oldB598Kanzashi     = "3c41c666-5cf4-4f35-9cb8-c41f1f78f31e"
	oldB598Floodbringer = "6c8bba8b-44de-4bbf-bf6a-3045e7313995"
	oldB598Miko         = "82732de2-b98d-4534-967f-2cf4c9e6bfbe"
	oldB598Palliator    = "ed854708-5a65-4293-bebc-d7c639407ba5"
	oldB598Twist        = "5b33f3f0-80d3-4ea2-b73b-74510c9f6071"
	oldB598Memories     = "7b77fb85-a6d8-4f3e-8493-99ec5b5a6d67"
	oldB598Reduce       = "85bac6ad-b63a-4068-a813-ad685966e171"
	oldB598Uproot       = "891d5c5e-3d4d-4183-b9fd-ab694639d1c3"
	oldB598Tragedies    = "b30e80ef-a4d7-48e2-918e-a347c2d9da59"
)

func TestOldB598CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB598Mindsweeper: "Soratami Mindsweeper", oldB598Kanzashi: "Ornate Kanzashi", oldB598Floodbringer: "Floodbringer",
		oldB598Miko: "Split-Tail Miko", oldB598Palliator: "Kitsune Palliator", oldB598Twist: "Twist Allegiance",
		oldB598Memories: "Enshrined Memories", oldB598Reduce: "Reduce to Dreams", oldB598Uproot: "Uproot",
		oldB598Tragedies: "Three Tragedies",
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

func oldB598Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB598Spell(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	return id
}

func TestOldB598SoratamiMindsweeperReturnsALandToMillTwo(t *testing.T) {
	g, me, opp := oldB598Table(t)
	sweeper := effectstest.PushCatalogPermanent(g, me.ID, "Soratami Mindsweeper", "Creature — Moonfolk Wizard", oldB598Mindsweeper, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, sweeper, "flying") {
		t.Error("the Mindsweeper lacks flying")
	}
	land := effectstest.CcLand(g, me.ID, "Island", "Basic Land — Island")
	effectstest.SeedLibrary(opp, "A", "B", "C", "D")
	size := opp.Library.Size()
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := effectstest.RavB2Activate(t, g, me.ID, sweeper, 0, game.ActivateAbilityParams{Targets: target}, "C", "C"); err == nil {
		t.Fatal("the ability activated without returning a land")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, sweeper, 0, game.ActivateAbilityParams{Targets: target, ReturnIDs: []uuid.UUID{land}})
	if !me.Hand.Contains(land) {
		t.Error("the land did not return to hand")
	}
	if opp.Library.Size() != size-2 {
		t.Errorf("library %d, want %d", opp.Library.Size(), size-2)
	}
}

func TestOldB598FloodbringerReturnsALandToTapALand(t *testing.T) {
	g, me, opp := oldB598Table(t)
	bringer := effectstest.PushCatalogPermanent(g, me.ID, "Floodbringer", "Creature — Moonfolk Wizard", oldB598Floodbringer, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, bringer, "flying") {
		t.Error("Floodbringer lacks flying")
	}
	back := effectstest.CcLand(g, me.ID, "Plains", "Basic Land — Plains")
	theirs := effectstest.CcLand(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := effectstest.RavB2Activate(t, g, me.ID, bringer, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(theirs)}, "C", "C"); err == nil {
		t.Fatal("the ability activated without returning a land")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, bringer, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(theirs), ReturnIDs: []uuid.UUID{back}})
	if !me.Hand.Contains(back) {
		t.Error("the land paid for the cost did not return to hand")
	}
	if !effectstest.IsTapped(g, theirs) {
		t.Error("the target land was not tapped")
	}
}

func TestOldB598SplitTailMikoPreventsTheNextTwo(t *testing.T) {
	g, me, _ := oldB598Table(t)
	miko := effectstest.PushCatalogPermanent(g, me.ID, "Split-Tail Miko", "Creature — Fox Cleric", oldB598Miko, false)
	effectstest.RavB2MustActivate(t, g, me.ID, miko, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}, "W")
	if !effectstest.IsTapped(g, miko) {
		t.Error("the Miko did not tap")
	}
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(miko, me.ID, 3) })
	if me.Life != life-1 {
		t.Errorf("life %d, want %d (2 of 3 damage prevented)", me.Life, life-1)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(miko, me.ID, 3) })
	if me.Life != life-4 {
		t.Errorf("life %d, want %d (the shield was used up)", me.Life, life-4)
	}
}

func TestOldB598KitsunePalliatorShieldsEveryCreatureAndPlayerOnce(t *testing.T) {
	g, me, opp := oldB598Table(t)
	pal := effectstest.PushCatalogPermanent(g, me.ID, "Kitsune Palliator", "Creature — Fox Cleric", oldB598Palliator, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 4, 4)
	effectstest.RavB2MustActivate(t, g, me.ID, pal, 0, game.ActivateAbilityParams{})
	later := effectstest.B12Creature(g, opp.ID, "Latecomer", "Creature — Bear", 4, 4)
	for _, id := range []uuid.UUID{mine, theirs, later} {
		effectstest.DealDamageTo(t, g, pal, id, 2)
	}
	if n := effectstest.DamageMarkedOn(g, mine); n != 1 {
		t.Errorf("my creature took %d, want 1 (one prevented)", n)
	}
	if n := effectstest.DamageMarkedOn(g, theirs); n != 1 {
		t.Errorf("their creature took %d, want 1", n)
	}
	if n := effectstest.DamageMarkedOn(g, later); n != 2 {
		t.Errorf("a creature that entered later took %d, want 2 (no shield)", n)
	}
	for _, p := range []*game.Player{me, opp} {
		life := p.Life
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pal, p.ID, 3) })
		if p.Life != life-2 {
			t.Errorf("a player lost %d, want 2 (one prevented)", life-p.Life)
		}
	}
}

func TestOldB598OrnateKanzashiLetsYouPlayTheOpponentsTopCard(t *testing.T) {
	g, me, opp := oldB598Table(t)
	kanzashi := effectstest.PushCatalogPermanent(g, me.ID, "Ornate Kanzashi", "Artifact", oldB598Kanzashi, false)
	top := effectstest.StackLibrary(opp, "Their Top Card", "Creature — Bear", "{1}{G}")
	effectstest.RavB2MustActivate(t, g, me.ID, kanzashi, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}, "C", "C")
	if !g.Exile.Contains(top) {
		t.Fatal("the opponent's top card was not exiled")
	}
	perm := effectstest.ExiledPermission(g, top)
	if !effectstest.PermissionLive(g, perm, me.ID) {
		t.Error("you have no live permission to play the exiled card")
	}
	if effectstest.PermissionLive(g, perm, opp.ID) {
		t.Error("the opponent was given the permission")
	}
	if !effectstest.IsTapped(g, kanzashi) {
		t.Error("the Kanzashi did not tap")
	}
}

func TestOldB598TwistAllegianceSwapsEveryCreature(t *testing.T) {
	g, me, opp := oldB598Table(t)
	other := g.Seats[2]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	third := effectstest.B12Creature(g, other.ID, "Third Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, theirs).Tapped = true })
	oldB598Spell(t, g, me, "Twist Allegiance", "Sorcery", oldB598Twist, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	controller := func(id uuid.UUID) uuid.UUID {
		var c uuid.UUID
		g.WithWriteLock(func() { c = effectstest.FindBattlefieldCardForTest(g, id).Controller })
		return c
	}
	if controller(mine) != opp.ID || controller(theirs) != me.ID {
		t.Fatal("the two players' creatures did not swap")
	}
	if controller(third) != other.ID {
		t.Error("a third player's creature changed hands")
	}
	if effectstest.IsTapped(g, theirs) {
		t.Error("the stolen creature was not untapped")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, theirs, "haste") || !effectstest.HasKeywordOnBattlefield(t, g, mine, "haste") {
		t.Error("the swapped creatures did not gain haste")
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if controller(mine) != me.ID || controller(theirs) != opp.ID {
		t.Error("control did not return at end of turn")
	}
}

func TestOldB598EnshrinedMemoriesTakesCreaturesAndBottomsTheRest(t *testing.T) {
	g, me, _ := oldB598Table(t)
	deep := effectstest.SeedLibraryTop(me, "Deep Card", "Creature — Bear", "{2}")
	third := effectstest.SeedLibraryTop(me, "Third", "Creature — Bear", "{2}")
	land := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	first := effectstest.SeedLibraryTop(me, "First", "Creature — Bear", "{2}")
	hand, size := me.Hand.Size(), me.Library.Size()
	oldB598Spell(t, g, me, "Enshrined Memories", "Sorcery", oldB598Memories, game.CastSpellParams{XValue: 3})
	for _, id := range []uuid.UUID{first, third} {
		if !me.Hand.Contains(id) {
			t.Errorf("a revealed creature card did not go to hand")
		}
	}
	if me.Hand.Contains(land) || me.Hand.Contains(deep) {
		t.Error("a land or an unrevealed card went to hand")
	}
	if me.Hand.Size() != hand+2 || me.Library.Size() != size-2 {
		t.Errorf("hand %d (want %d), library %d (want %d)", me.Hand.Size(), hand+2, me.Library.Size(), size-2)
	}
	if me.Library.Cards[0].InstanceID != land {
		t.Error("the revealed land is not on the bottom of the library")
	}
}

func TestOldB598ReduceToDreamsReturnsEveryArtifactAndEnchantment(t *testing.T) {
	g, me, opp := oldB598Table(t)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	shrine := effectstest.B12Permanent(g, opp.ID, "Shrine", "Enchantment")
	both := effectstest.B12Permanent(g, opp.ID, "Both", "Enchantment Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	oldB598Spell(t, g, me, "Reduce to Dreams", "Sorcery", oldB598Reduce, game.CastSpellParams{})
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(shrine) || g.Battlefield.Contains(both) {
		t.Error("an artifact or enchantment stayed on the battlefield")
	}
	if !me.Hand.Contains(rock) || !opp.Hand.Contains(shrine) || !opp.Hand.Contains(both) {
		t.Error("an artifact or enchantment did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(land) {
		t.Error("a creature or land was returned")
	}
}

func TestOldB598UprootPutsALandOnTopOfItsOwnersLibrary(t *testing.T) {
	g, me, opp := oldB598Table(t)
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	oldB598Cast := func(target uuid.UUID) error {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Uproot", TypeLine: "Sorcery — Arcane", OracleID: oldB598Uproot, Owner: me.ID, Controller: me.ID})
		})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)})
	}
	if err := oldB598Cast(bear); err == nil {
		t.Fatal("a creature was a legal target")
	}
	if err := oldB598Cast(land); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID != land {
		t.Error("the land is not on top of its owner's library")
	}
}

func TestOldB598ThreeTragediesMakesThePlayerDiscardThree(t *testing.T) {
	g, me, opp := oldB598Table(t)
	opp.Hand.Cards = nil
	hand := effectstest.FillHand(opp, "A", "B", "C", "D", "E")
	oldB598Spell(t, g, me, "Three Tragedies", "Sorcery — Arcane", oldB598Tragedies, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.AnswerDiscard(t, g, opp.ID, hand[0], hand[1], hand[2])
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 2 || !opp.Graveyard.Contains(hand[0]) || !opp.Graveyard.Contains(hand[2]) || opp.Hand.Contains(hand[3]) == false {
		t.Errorf("hand %d, want the two undiscarded cards", opp.Hand.Size())
	}
}
