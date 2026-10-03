package protocol

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emblem_targeting_view_test.go — ADR 0109 Delivery PR 6: the view reads
// the same gates the engine and the enumerator do. An emblem's "can't
// cast" (§5, #1899) greys an opponent's hand card with its clause on
// `cant_cast`; a restriction about graveyards (§6, #1885) leaves every
// graveyard card out of a spell's `legal_targets` and names itself on
// GameView.graveyard_target_bans. Real catalog cards.

const (
	oracleNarsetView     = "e1de0c94-0ecc-425a-9b92-27c2745d07e7"
	oracleGroundSealView = "13f6f960-ef79-4f2c-8874-90fe9e77099e"
	oracleRegrowthView   = "e6e4a8bd-5c40-4654-8de1-0da9afed90fd"
)

// An opponent's Narset emblem stamps its clause on the viewer's
// noncreature hand card and not on a creature card.
func TestHandCardCarriesAnEmblemsCastBan(t *testing.T) {
	g, me, opp := stripTable(t)
	bolt := publicCard(g, me.ID, "Shock", "Instant", "{R}", "")
	bear := publicCard(g, me.ID, "Bear", "Creature — Bear", "{G}", "")
	me.Hand.PushTop(bolt)
	me.Hand.PushTop(bear)
	narset := publicCard(g, opp.ID, "Narset Transcendent", "Legendary Planeswalker — Narset", "{2}{W}{U}", oracleNarsetView)
	g.Battlefield.PushTop(narset)
	var err error
	g.WithWriteLock(func() { err = g.CreateEmblemForEffect(opp.ID, narset.InstanceID) })
	if err != nil {
		t.Fatalf("CreateEmblemForEffect: %v", err)
	}
	if got := ownHandCard(t, g, me, bolt.InstanceID).CantCast; got != "Your opponents can't cast noncreature spells." {
		t.Errorf("the instant's cant_cast = %q, want the emblem's clause", got)
	}
	if got := ownHandCard(t, g, me, bear.InstanceID).CantCast; got != "" {
		t.Errorf("the creature's cant_cast = %q, want none", got)
	}
	if got := seatView(t, g, me, opp).Emblems; len(got) != 1 {
		t.Errorf("the emblem's owner shows %d emblems, want 1", len(got))
	}
}

// Under Ground Seal a graveyard card is in no spell's legal_targets, and
// the frame names the restriction for the graveyard viewer.
func TestGraveyardCardsLeaveLegalTargetsUnderGroundSeal(t *testing.T) {
	g, me, opp := stripTable(t)
	regrowth := publicCard(g, me.ID, "Regrowth", "Sorcery", "{1}{G}", oracleRegrowthView)
	me.Hand.PushTop(regrowth)
	dead := publicCard(g, me.ID, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	me.Graveyard.PushTop(dead)
	targets := func() []string {
		c := ownHandCard(t, g, me, regrowth.InstanceID)
		if c.LegalTargets == nil {
			return nil
		}
		return c.LegalTargets.Cards
	}
	if got := targets(); len(got) != 1 || got[0] != dead.InstanceID.String() {
		t.Fatalf("setup: Regrowth's legal targets = %v, want the graveyard card", got)
	}
	if bans := ViewOfGameFor(g, opp.ID.String()).GraveyardTargetBans; len(bans) != 0 {
		t.Fatalf("a frame with no restriction carries %v", bans)
	}

	g.Battlefield.PushTop(publicCard(g, opp.ID, "Ground Seal", "Enchantment", "{1}{G}", oracleGroundSealView))
	if got := targets(); len(got) != 0 {
		t.Errorf("Regrowth's legal targets under Ground Seal = %v, want none", got)
	}
	bans := ViewOfGameFor(g, opp.ID.String()).GraveyardTargetBans
	if len(bans) != 1 || !strings.HasPrefix(bans[0], "Cards in graveyards can't be the targets") || !strings.HasSuffix(bans[0], " — Ground Seal") {
		t.Errorf("graveyard_target_bans = %v, want the Seal's clause and name", bans)
	}
	if err := g.CastSpell(me.ID, regrowth.InstanceID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead.InstanceID}},
	}); err == nil {
		t.Error("the engine accepted the target the view left out")
	}
}
