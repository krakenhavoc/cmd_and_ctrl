package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b640_cards_test.go — Planeshift, batch 640 (#5615). Each test plays
// the real catalog entry and asserts what its text says.

const (
	b640FuryOracle     = "2e482522-2eef-4f6c-887a-fcc51be61dc2"
	b640SpecterOracle  = "5811effb-5010-44b9-bb97-9d7eb358327f"
	b640StrafeOracle   = "6ac7990b-b86e-4725-974c-d6affa2af2f9"
	b640SlingshotOrcl  = "b377595f-c6c2-4c54-b462-73ef4eba519b"
	b640TahngarthOracl = "c5b4131e-13aa-4a22-b8f5-e5f9d366cb4f"
	b640SingeOracle    = "e7fbdebf-2275-4bf6-b364-6a1849765f94"
	b640BearType       = "Creature — Bear"
)

func TestOldB640EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b640FuryOracle:     "Planeswalker's Fury",
		b640SpecterOracle:  "Doomsday Specter",
		b640StrafeOracle:   "Strafe",
		b640SlingshotOrcl:  "Slingshot Goblin",
		b640TahngarthOracl: "Tahngarth, Talruum Hero",
		b640SingeOracle:    "Singe",
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

func b640Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func b640Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string, targets ...game.TargetRef) error {
	t.Helper()
	effectstest.DcMana(t, g, me.ID, mana)
	err := effectstest.InvB1Activate(g, me.ID, id, 0, targets...)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func b640Ready(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

func TestOldB640PlaneswalkersFuryDealsTheRevealedCardsManaValue(t *testing.T) {
	g, me, opp := b640Table(t)
	fury := effectstest.B12Push(g, me.ID, "Planeswalker's Fury", "Enchantment", b640FuryOracle, 0, 0)
	opp.Hand.Cards = nil
	effectstest.PutInHand(opp, game.Card{Name: "Five Drop", TypeLine: "Sorcery", ManaCost: "{4}{R}"})
	life := opp.Life
	if err := b640Activate(t, g, me, fury, "{R}{R}{R}{R}", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := life - opp.Life; got != 5 {
		t.Errorf("the opponent lost %d life, want 5", got)
	}
	if opp.Hand.Size() != 1 {
		t.Error("the revealed card left the opponent's hand")
	}
}

func TestOldB640PlaneswalkersFuryDoesNothingToAnEmptyHandOrToYou(t *testing.T) {
	g, me, opp := b640Table(t)
	fury := effectstest.B12Push(g, me.ID, "Planeswalker's Fury", "Enchantment", b640FuryOracle, 0, 0)
	opp.Hand.Cards = nil
	life := opp.Life
	if err := b640Activate(t, g, me, fury, "{R}{R}{R}{R}", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if opp.Life != life {
		t.Errorf("an empty hand cost %d life", life-opp.Life)
	}
	if err := b640Activate(t, g, me, fury, "{R}{R}{R}{R}", game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("the controller was accepted as the target opponent")
	}
}

func TestOldB640DoomsdaySpecterReturnsABlueOrBlackCreatureWhenItEnters(t *testing.T) {
	g, me, _ := b640Table(t)
	blue := effectstest.B27Push(g, me.ID, "Blue Bear", b640BearType, "", "{1}{U}", 2, 2, "U")
	red := effectstest.B27Push(g, me.ID, "Red Bear", b640BearType, "", "{1}{R}", 2, 2, "R")
	specter := effectstest.CastCatalogSpell(t, g, "Doomsday Specter", "Creature — Specter", b640SpecterOracle, nil)
	effectstest.PassUntilOnBattlefield(t, g, specter)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no return pick was queued")
	}
	if effectstest.HasID(p.ChooseCards, red) {
		t.Error("a red creature was offered")
	}
	if !effectstest.HasID(p.ChooseCards, blue) {
		t.Error("a blue creature was not offered")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{blue}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(blue) || !me.Hand.Contains(blue) {
		t.Error("the picked creature did not return to its owner's hand")
	}
}

func TestOldB640DoomsdaySpecterMakesTheDamagedPlayerDiscardACardYouChoose(t *testing.T) {
	g, me, opp := b640Table(t)
	specter := effectstest.B27Push(g, me.ID, "Doomsday Specter", "Creature — Specter", b640SpecterOracle, "{2}{U}{B}", 2, 3, "U", "B")
	opp.Hand.Cards = nil
	keep := effectstest.PutInHand(opp, game.Card{Name: "Keep Me", TypeLine: "Sorcery", ManaCost: "{1}"})
	take := effectstest.PutInHand(opp, game.Card{Name: "Take Me", TypeLine: "Sorcery", ManaCost: "{2}"})
	effectstest.DeclareAttack(t, g, opp.ID, specter)
	effectstest.NeoToDamage(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if pick.Chooser != me.ID {
		t.Fatalf("the pick belongs to %v, not the Specter's controller", pick.Chooser)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{take}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	if !opp.Graveyard.Contains(take) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card was not the one discarded")
	}
}

func TestOldB640StrafeDamagesANonredCreatureOnly(t *testing.T) {
	g, _, opp := b640Table(t)
	red := effectstest.B27Push(g, opp.ID, "Red Bear", b640BearType, "", "{1}{R}", 2, 9, "R")
	green := effectstest.B27Push(g, opp.ID, "Green Bear", b640BearType, "", "{1}{G}", 2, 9, "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Strafe", "Sorcery", b640StrafeOracle, effectstest.CardTarget(red)); err == nil {
		t.Fatal("a red creature was accepted as the target")
	}
	effectstest.CastCatalogSpell(t, g, "Strafe", "Sorcery", b640StrafeOracle, effectstest.CardTarget(green))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, green); got != 3 {
		t.Errorf("the creature was dealt %d damage, want 3", got)
	}
}

func TestOldB640SlingshotGoblinDamagesABlueCreatureOnly(t *testing.T) {
	g, me, opp := b640Table(t)
	goblin := effectstest.B27Push(g, me.ID, "Slingshot Goblin", "Creature — Goblin", b640SlingshotOrcl, "{2}{R}", 2, 2, "R")
	b640Ready(g, goblin)
	blue := effectstest.B27Push(g, opp.ID, "Blue Bear", b640BearType, "", "{1}{U}", 2, 9, "U")
	green := effectstest.B27Push(g, opp.ID, "Green Bear", b640BearType, "", "{1}{G}", 2, 9, "G")
	if err := b640Activate(t, g, me, goblin, "{R}", effectstest.CardTarget(green)...); err == nil {
		t.Fatal("a non-blue creature was accepted as the target")
	}
	if err := b640Activate(t, g, me, goblin, "{R}", effectstest.CardTarget(blue)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.DamageOn(t, g, blue); got != 2 {
		t.Errorf("the creature was dealt %d damage, want 2", got)
	}
	if !effectstest.FindCardAnywhere(t, g, goblin).Tapped {
		t.Error("the {T} cost was not paid")
	}
}

func TestOldB640TahngarthAndItsTargetDealDamageToEachOther(t *testing.T) {
	g, me, opp := b640Table(t)
	tahngarth := effectstest.B12Push(g, me.ID, "Tahngarth, Talruum Hero", "Legendary Creature — Minotaur Warrior", b640TahngarthOracl, 4, 4)
	b640Ready(g, tahngarth)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b640BearType, 2, 9)
	if err := b640Activate(t, g, me, tahngarth, "{R}{R}", effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.DamageOn(t, g, bear); got != 4 {
		t.Errorf("the target was dealt %d damage, want 4 (Tahngarth's power)", got)
	}
	if got := effectstest.DamageOn(t, g, tahngarth); got != 2 {
		t.Errorf("Tahngarth was dealt %d damage, want 2 (the target's power)", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, tahngarth, "vigilance") {
		t.Error("Tahngarth lacks vigilance")
	}
}

func TestOldB640SingeDamagesTheCreatureAndMakesItBlack(t *testing.T) {
	g, _, opp := b640Table(t)
	bear := effectstest.B27Push(g, opp.ID, "Green Bear", b640BearType, "", "{1}{G}", 2, 9, "G")
	effectstest.CastCatalogSpell(t, g, "Singe", "Instant", b640SingeOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, bear); got != 1 {
		t.Errorf("the creature was dealt %d damage, want 1", got)
	}
	if got := effectstest.TmtB9Colors(t, g, bear); got != "B" {
		t.Errorf("colors = %q, want only black", got)
	}
}
