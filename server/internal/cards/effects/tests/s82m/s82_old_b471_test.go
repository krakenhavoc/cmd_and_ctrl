package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b471_test.go — Scars of Mirrodin instants, sorceries, damage and
// removal, S82 OLD batch 471 (#5577): eleven cards, each driven through the
// engine.

const (
	b471ShapeAnew   = "21b7cf4f-1c86-4700-be71-68c974d143b0"
	b471Miasma      = "43d69e13-27df-43a9-b610-6628dc02e3b4"
	b471Fulgent     = "7996904a-1d7c-4d95-8fd8-431272c7c7e0"
	b471Memoricide  = "deeb942b-cf50-43f7-90a9-bdce4e42123a"
	b471Ignus       = "41353001-eba4-45b5-8dd2-09d49b63b6b1"
	b471Eruption    = "8143642e-ce93-47f1-a601-3a4f5731e1d7"
	b471Angler      = "887a9f2f-17b4-4a20-a464-9665cde6f293"
	b471Embersmith  = "e16cec8b-e0ac-44ee-91cf-658e204ca120"
	b471MeltTerrain = "1c0940d6-9f0d-4d84-ab28-ff6595254675"
	b471Spider      = "23d4f436-a417-4a26-b5b8-c698f11a186b"
	b471Allergy     = "bd3ce6a2-9efc-4dab-a357-9f2aa0a4c0c6"
)

func TestS82OldB471EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b471ShapeAnew: "Shape Anew", b471Miasma: "Psychic Miasma", b471Fulgent: "Fulgent Distraction",
		b471Memoricide: "Memoricide", b471Ignus: "Tunnel Ignus", b471Eruption: "Cerebral Eruption",
		b471Angler: "Tangle Angler", b471Embersmith: "Embersmith", b471MeltTerrain: "Melt Terrain",
		b471Spider: "Acid Web Spider", b471Allergy: "Flesh Allergy",
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

// b471Cast casts a spell from the active seat's hand at the given targets and
// lets it resolve, leaving any prompt it asks open.
func b471Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func b471Library(p *game.Player, name, typeLine string, mv string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: mv, Owner: p.ID, Controller: p.ID})
	return id
}

func TestS82OldB471ShapeAnewSacrificesThenPutsTheFirstArtifactFromTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	old := effectstest.B12Permanent(g, opp.ID, "Old Trinket", "Artifact")
	fresh := b471Library(opp, "Fresh Trinket", "Artifact", "{2}")
	b471Library(opp, "Bear", "Creature — Bear", "{1}{G}")
	b471Library(opp, "Forest", "Basic Land — Forest", "")
	lib := opp.Library.Size()
	b471Cast(t, g, "Shape Anew", "Sorcery", b471ShapeAnew, game.TargetRef{Kind: game.TargetCard, ID: old})
	if g.Battlefield.Contains(old) {
		t.Error("the targeted artifact was not sacrificed")
	}
	if !opp.Graveyard.Contains(old) {
		t.Error("the sacrificed artifact is not in its owner's graveyard")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Fatal("the first artifact card in the library did not enter")
	}
	if c := effectstest.FindBattlefieldCardForTest(g, fresh); c.Controller != opp.ID {
		t.Errorf("the new artifact entered under %v, want its putter %v", c.Controller, opp.ID)
	}
	if got := opp.Library.Size(); got != lib-1 {
		t.Errorf("library size %d → %d, want one fewer (the rest was shuffled back in)", lib, got)
	}
	_ = me
}

func TestS82OldB471PsychicMiasmaReturnsOnlyWhenALandIsDiscarded(t *testing.T) {
	for _, land := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		typeLine := "Creature — Bear"
		if land {
			typeLine = "Basic Land — Forest"
		}
		card := effectstest.B13HandCard(opp, "Discard Me", typeLine, "")
		spell := b471Cast(t, g, "Psychic Miasma", "Sorcery", b471Miasma, effectstest.PlayerTarget(opp.ID))
		effectstest.AnswerDiscard(t, g, opp.ID, card)
		effectstest.PassPriorityAroundTable(t, g)
		if !opp.Graveyard.Contains(card) {
			t.Fatalf("land=%v: the card was not discarded", land)
		}
		if got := me.Hand.Contains(spell); got != land {
			t.Errorf("land=%v: Psychic Miasma in its owner's hand = %v, want %v", land, got, land)
		}
		if got := me.Graveyard.Contains(spell); got == land {
			t.Errorf("land=%v: Psychic Miasma in the graveyard = %v", land, got)
		}
	}
}

func TestS82OldB471FulgentDistractionTapsTwoAndUnattachesEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	sword := effectstest.B12Permanent(g, opp.ID, "Sword", "Artifact — Equipment")
	spare := effectstest.B12Permanent(g, opp.ID, "Spare", "Artifact — Equipment")
	g.WithWriteLock(func() {
		for _, pair := range [][2]uuid.UUID{{sword, a}, {spare, c}} {
			if err := g.AttachForEffect(pair[0], game.TargetRef{Kind: game.TargetCard, ID: pair[1]}); err != nil {
				t.Errorf("AttachForEffect: %v", err)
			}
		}
	})
	b471Cast(t, g, "Fulgent Distraction", "Instant", b471Fulgent,
		game.TargetRef{Kind: game.TargetCard, ID: a}, game.TargetRef{Kind: game.TargetCard, ID: b})
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.FindBattlefieldCardForTest(g, id).Tapped {
			t.Error("a target creature was not tapped")
		}
	}
	if effectstest.FindBattlefieldCardForTest(g, c).Tapped {
		t.Error("an untargeted creature was tapped")
	}
	if got := effectstest.FindBattlefieldCardForTest(g, sword); got.AttachedTo.ID != uuid.Nil {
		t.Error("the Equipment on a target is still attached")
	}
	if !g.Battlefield.Contains(sword) {
		t.Error("the Equipment left the battlefield")
	}
	if got := effectstest.FindBattlefieldCardForTest(g, spare); got.AttachedTo.ID != c {
		t.Error("the Equipment on an untargeted creature was unattached")
	}
}

func TestS82OldB471MemoricideExilesTheNamedCardEverywhere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	inLib := b471Library(opp, "Doomed Bear", "Creature — Bear", "{1}{G}")
	keep := b471Library(opp, "Other Bear", "Creature — Bear", "{1}{G}")
	inHand := effectstest.B13HandCard(opp, "Doomed Bear", "Creature — Bear", "")
	inYard := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: inYard, Name: "Doomed Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	b471Cast(t, g, "Memoricide", "Sorcery", b471Memoricide, effectstest.PlayerTarget(opp.ID))
	effectstest.ActNameFor(t, g, me.ID, "Doomed Bear")
	effectstest.AnswerSearchByID(t, g, me.ID, inLib, inHand, inYard)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{inLib, inHand, inYard} {
		if !g.Exile.Contains(id) {
			t.Errorf("a card named Doomed Bear was not exiled")
		}
	}
	if g.Exile.Contains(keep) {
		t.Error("a card with another name was exiled")
	}
}

func TestS82OldB471MemoricideCannotNameALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	land := b471Library(opp, "Forest", "Basic Land — Forest", "")
	b471Cast(t, g, "Memoricide", "Sorcery", b471Memoricide, effectstest.PlayerTarget(opp.ID))
	effectstest.ActNameFor(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(land) {
		t.Error("a land card was exiled by a nonland name")
	}
}

// b471LandEnters puts a land onto the battlefield under owner's control and
// emits its enter event.
func b471LandEnters(g *game.Game, owner uuid.UUID) {
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner, Controller: owner})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID, Actor: owner})
	})
}

func TestS82OldB471TunnelIgnusPunishesTheSecondLandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.B12Push(g, me.ID, "Tunnel Ignus", "Creature — Elemental", b471Ignus, 2, 1)
	life := opp.Life
	b471LandEnters(g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Fatalf("the first land dealt %d damage, want 0", life-opp.Life)
	}
	b471LandEnters(g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("after the second land life is %d, want %d", opp.Life, life-3)
	}
	// Its own controller's lands don't count.
	mine := me.Life
	b471LandEnters(g, me.ID)
	b471LandEnters(g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Error("the controller's own lands dealt damage to them")
	}
}

func TestS82OldB471CerebralEruptionDamagesByManaValueAndReturnsOnALand(t *testing.T) {
	for _, land := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		hit := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		big := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		if land {
			b471Library(opp, "Forest", "Basic Land — Forest", "")
		} else {
			b471Library(opp, "Top Card", "Sorcery", "{2}{R}")
		}
		life := opp.Life
		spell := b471Cast(t, g, "Cerebral Eruption", "Sorcery", b471Eruption, effectstest.PlayerTarget(opp.ID))
		want := 3
		if land {
			want = 0
		}
		if opp.Life != life-want {
			t.Errorf("land=%v: their life %d → %d, want -%d", land, life, opp.Life, want)
		}
		if g.Battlefield.Contains(hit) == (want >= 2) {
			t.Errorf("land=%v: their 2/2 on the battlefield = %v", land, g.Battlefield.Contains(hit))
		}
		if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(mine) {
			t.Errorf("land=%v: an unrelated creature died", land)
		}
		if got := me.Hand.Contains(spell); got != land {
			t.Errorf("land=%v: the spell is in its owner's hand = %v", land, got)
		}
	}
}

func TestS82OldB471TangleAnglerForcesABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	angler := effectstest.B21Push(g, me.ID, "Tangle Angler", "Creature — Phyrexian Horror", b471Angler, 1, 5, "G")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, angler, "infect") {
		t.Error("the Angler has no infect")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, angler, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "G")
	effectstest.DeclareAttack(t, g, opp.ID, angler)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers(nil); err == nil {
		t.Error("the defender declared no blockers although the target must block if able")
	}
}

func TestS82OldB471EmbersmithPingsOnlyWhenPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		effectstest.B12Push(g, me.ID, "Embersmith", "Creature — Human Artificer", b471Embersmith, 2, 1)
		b471Cast(t, g, "Trinket", "Artifact", "")
		pick := effectstest.LatestPickTarget(g, me.ID)
		if pick == nil {
			t.Fatalf("pay=%v: no target prompt for the cast trigger", pay)
		}
		if err := g.ResolvePickTarget(pick.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		effectstest.B06AddMana(me, "C")
		life := opp.Life
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 1
		}
		if opp.Life != life-want {
			t.Errorf("pay=%v: life %d → %d, want -%d", pay, life, opp.Life, want)
		}
	}
}

func TestS82OldB471MeltTerrainDestroysALandAndHurtsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	land := effectstest.AangPushLand(g, opp.ID, "Mountain", false)
	life := opp.Life
	b471Cast(t, g, "Melt Terrain", "Sorcery", b471MeltTerrain, game.TargetRef{Kind: game.TargetCard, ID: land})
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, opp.Life)
	}
}

func TestS82OldB471AcidWebSpiderMayDestroyAnEquipment(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
		sword := effectstest.B12Permanent(g, opp.ID, "Sword", "Artifact — Equipment")
		me := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Acid Web Spider", TypeLine: "Creature — Spider", OracleID: b471Spider, Power: 3, Toughness: 5, Owner: me.ID, Controller: me.ID})
		effectstest.ToMainForCost(t, g)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.PickCard(t, g, me.ID, sword)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(sword) == yes {
			t.Errorf("yes=%v: the Equipment is on the battlefield = %v", yes, g.Battlefield.Contains(sword))
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "reach") {
			t.Error("the Spider has no reach")
		}
	}
}

func TestS82OldB471FleshAllergyDestroysAndDrainsPerCreatureThatDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	id := effectstest.B13HandCard(me, "Flesh Allergy", "Sorcery", b471Allergy)
	effectstest.ToMainForCost(t, g)
	// Without a creature to sacrifice it can't be cast strictly.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Fatal("cast with no sacrifice named")
	}
	life := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim), SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || g.Battlefield.Contains(victim) {
		t.Error("the sacrificed or targeted creature is still on the battlefield")
	}
	if opp.Life != life-2 {
		t.Errorf("their life %d → %d, want -2 (the sacrificed and the destroyed creature)", life, opp.Life)
	}
}
