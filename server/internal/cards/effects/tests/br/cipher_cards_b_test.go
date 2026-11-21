package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cibCastAndEncode casts a (targeted) cipher sorcery and encodes it on
// `creature` once its own text has resolved.
func cibCastAndEncode(t *testing.T, g *game.Game, me *game.Player, name, oracle string, targets []game.TargetRef, creature uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, "Sorcery", oracle, targets)
	pick := effectstest.CipherResolveToPrompt(t, g, me.ID)
	if pick == nil {
		t.Fatalf("%s: cipher asks whether to encode the card", name)
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if len(o.Cards) == 1 && o.Cards[0] == creature {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("%s: the creature is not offered: %+v", name, pick.PickOptions)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, idx)
	g.SettleResolution()
	if on, ok := effectstest.EncodedOn(g, id); !ok || on != creature {
		t.Fatalf("%s is not encoded on the chosen creature", name)
	}
	return id
}

func cibCard(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// Shadow Slice: the target opponent loses 3; encoded, each hit casts a
// copy at an opponent of the copy's caster's choosing.
func TestCibShadowSliceDrainsAndCiphers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	third := g.Seats[(g.Turn.ActiveSeat+2)%4]
	bear := effectstest.SeedBear(g, me.ID)
	oppLife, thirdLife := opp.Life, third.Life

	id := cibCastAndEncode(t, g, me, "Shadow Slice", effectstest.CibShadowSliceOracle, effectstest.CibPlayer(opp.ID), bear)
	if opp.Life != oppLife-3 {
		t.Fatalf("target opponent at %d, want %d", opp.Life, oppLife-3)
	}

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	cp := effectstest.CopyOf(t, g, "Shadow Slice")
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, effectstest.PlayerTarget(third.ID)); err != nil {
		t.Fatalf("the free copy targets another opponent: %v", err)
	}
	if pick := effectstest.CipherResolveToPrompt(t, g, me.ID); pick != nil {
		t.Fatalf("the copy is not a card, yet cipher asked %q", pick.Reason)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if third.Life != thirdLife-3 {
		t.Errorf("the copy's target at %d, want %d", third.Life, thirdLife-3)
	}
	if !effectstest.InExile(g, id) {
		t.Error("the encoded card stays in exile")
	}
}

// Shadow Slice may not target its caster.
func TestCibShadowSliceTargetsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.HandCardFull(me, "Shadow Slice", "Sorcery", "{4}{B}", effectstest.CibShadowSliceOracle, []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CibPlayer(me.ID)}); err == nil {
		t.Error("Shadow Slice targeting its caster must be refused")
	}
}

// Stolen Identity: a token copy of the target under the caster's
// control; the free copy cast through the encoded creature copies a
// different target.
func TestCibStolenIdentityCopiesAndCiphers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.SeedBear(g, me.ID)
	ogre := effectstest.PushVanillaCreature(g, opp.ID, "Their Ogre", 3, 3)
	wall := effectstest.PushVanillaCreature(g, opp.ID, "Their Wall", 0, 4)

	cibCastAndEncode(t, g, me, "Stolen Identity", effectstest.CibStolenIdentityOracle, cibCard(ogre), mine)
	if n := cibControlledNamed(g, me.ID, "Their Ogre"); n != 1 {
		t.Fatalf("%d copies of the ogre under the caster's control, want 1", n)
	}

	effectstest.DealCombatDamageToPlayer(g, mine, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	cp := effectstest.CopyOf(t, g, "Stolen Identity")
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, cibCard(wall)...); err != nil {
		t.Fatalf("the free copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := cibControlledNamed(g, me.ID, "Their Wall"); n != 1 {
		t.Errorf("%d copies of the wall under the caster's control, want 1", n)
	}
}

// Stolen Identity copies an artifact too.
func TestCibStolenIdentityCopiesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Stolen Identity", "Sorcery", effectstest.CibStolenIdentityOracle, cibCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if n := cibControlledNamed(g, me.ID, "Their Rock"); n != 1 {
		t.Errorf("%d token copies of the artifact, want 1", n)
	}
}

func cibControlledNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

// Undercity Plague: one life, then a discard the player chooses, then a
// sacrifice of a permanent they choose — in that order.
func TestCibUndercityPlagueLosesDiscardsThenSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	ogre := effectstest.PushVanillaCreature(g, opp.ID, "Their Ogre", 3, 3)
	life, hand := opp.Life, opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Undercity Plague", "Sorcery", effectstest.CibUndercityPlagueOracle, effectstest.CibPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		t.Fatal("the sacrifice waits for the discard")
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), hand-1)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, ogre)
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the player sacrifices a permanent")
	}
}

// Voidwalk: the creature is exiled now and returns, under its owner's
// control, at the beginning of the next end step.
func TestCibVoidwalkBlinksUntilTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	ogre := effectstest.PushVanillaCreature(g, opp.ID, "Their Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Voidwalk", "Sorcery", effectstest.CibVoidwalkOracle, cibCard(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) || !effectstest.InExile(g, ogre) {
		t.Fatal("the target is exiled")
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.FindBattlefieldByName(g, "Their Ogre")
	if back == uuid.Nil {
		t.Fatal("the creature returns at the next end step")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == back && c.Controller != opp.ID {
			t.Errorf("it returns under %v, want its owner %v", c.Controller, opp.ID)
		}
	}
}

// Whispering Madness: every hand is discarded, then every player draws
// the greatest number discarded.
func TestCibWhisperingMadnessWheelsToTheGreatest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	for me.Hand.Size() > 1 {
		_, _ = me.Hand.Remove(me.Hand.Cards[0].InstanceID)
	}
	for i := 0; i < 2; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	}
	greatest := 0
	for _, p := range g.Seats {
		greatest = max(greatest, p.Hand.Size())
	}
	mine := me.Hand.Cards[0].InstanceID
	effectstest.CastCatalogSpell(t, g, "Whispering Madness", "Sorcery", effectstest.CibWhisperingMadnessOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) {
		t.Error("the caster's hand is discarded too")
	}
	for i, p := range g.Seats {
		if p.Hand.Size() != greatest {
			t.Errorf("seat %d holds %d, want the greatest discarded, %d", i, p.Hand.Size(), greatest)
		}
	}
}

// Writ of Return: the creature card comes back tapped — and is a
// creature the caster controls, so cipher may encode the card on it.
func TestCibWritOfReturnReanimatesTappedAndEncodesOnIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := effectstest.GyCard(me, "Dead Ogre", "Creature — Ogre", "{2}{R}", "")
	id := effectstest.CastCatalogSpell(t, g, "Writ of Return", "Sorcery", effectstest.CibWritOfReturnOracle, cibCard(dead))
	pick := effectstest.CipherResolveToPrompt(t, g, me.ID)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Dead Ogre" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("the creature card returns to the battlefield")
	}
	if !back.Tapped {
		t.Error("it returns tapped")
	}
	if pick == nil {
		t.Fatal("the returned creature is one cipher may encode the card on")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	g.SettleResolution()
	if on, ok := effectstest.EncodedOn(g, id); !ok || on != back.InstanceID {
		t.Error("Writ of Return is encoded on the creature it returned")
	}
}

// Writ of Return may not target an opponent's graveyard.
func TestCibWritOfReturnTargetsYourGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := effectstest.GyCard(opp, "Their Dead Ogre", "Creature — Ogre", "{2}{R}", "")
	id := effectstest.HandCardFull(me, "Writ of Return", "Sorcery", "{3}{B}{B}", effectstest.CibWritOfReturnOracle, []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cibCard(theirs)}); err == nil {
		t.Error("a creature card in an opponent's graveyard must be refused")
	}
}
