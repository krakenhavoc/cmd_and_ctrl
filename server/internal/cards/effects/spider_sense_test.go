package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const spiderSenseOracle = "c4485bf9-e7ff-48b5-a983-02900939ee9d"

// TestSpiderSenseTargetsInstantsSorceriesAndTriggeredAbilitiesOnly is
// the target_ability_cards_test.go shape (#1211) applied to the
// three-way clause: an instant, a sorcery and a triggered ability are
// legal; a creature spell and an activated ability are not.
func TestSpiderSenseTargetsInstantsSorceriesAndTriggeredAbilitiesOnly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	ability := taAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, nil)

	instant := batch01OpponentCasts(t, g, opp, "Their Instant", "", "", nil)
	sorcery := batch01OpponentCasts(t, g, opp, "Their Sorcery", "", "", nil)
	creatureSpell := batch01OpponentCasts(t, g, opp, "Their Creature", "", "", nil)
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			switch g.Stack.Cards[i].InstanceID {
			case sorcery:
				g.Stack.Cards[i].TypeLine = "Sorcery"
			case creatureSpell:
				g.Stack.Cards[i].TypeLine = "Creature — Bear"
			}
		}
	})

	if err := g.ActivateAbility(opp.ID, src, game.AbilityParams{
		Label: "Their Rock — an activation",
	}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	activated := acItemLabelled(g, "Their Rock — an activation")

	legal := taLegalTargets(g, me.ID, spiderSenseOracle)
	if !legal[ability] {
		t.Error("Spider-Sense must offer a triggered ability")
	}
	if legal[activated] {
		t.Error("Spider-Sense must not offer an activated ability")
	}
	if !legal[instant] {
		t.Error("Spider-Sense must offer an instant spell")
	}
	if !legal[sorcery] {
		t.Error("Spider-Sense must offer a sorcery spell")
	}
	if legal[creatureSpell] {
		t.Error("Spider-Sense must not offer a creature spell")
	}
}

// TestSpiderSenseWebSlingingReturnsATappedCreatureAndCounters is the
// Daze shape: the bounce is a cost, paid at announce while Spider-
// Sense is still on the stack, and the counter happens on resolution.
func TestSpiderSenseWebSlingingReturnsATappedCreatureAndCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	spiderSense := handCardFull(me, "Spider-Sense", "Instant", "{1}{U}", spiderSenseOracle, []string{"U"})
	creature := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	victim := batch01OpponentCasts(t, g, opp, "Their Instant", "", "", nil)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.CastSpell(me.ID, spiderSense, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "web_slinging",
		AltCostIDs:      []uuid.UUID{creature},
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("Spider-Sense web-slinging: %v", err)
	}
	// The creature is back in hand while Spider-Sense is still on the
	// stack — the bounce is a cost, not part of resolution.
	if !me.Hand.Contains(creature) {
		t.Error("the tapped creature did not return to hand")
	}
	if !g.Stack.Contains(spiderSense) {
		t.Error("Spider-Sense is not on the stack — the cost was paid too late")
	}

	passPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("the targeted spell was not countered")
	}
}

// Web-slinging needs a TAPPED creature; an untapped one cannot pay it.
func TestSpiderSenseWebSlingingRefusesAnUntappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	spiderSense := handCardFull(me, "Spider-Sense", "Instant", "{1}{U}", spiderSenseOracle, []string{"U"})
	creature := pushVanillaCreature(g, me.ID, "Untapped Bear", 2, 2)
	victim := batch01OpponentCasts(t, g, opp, "Their Instant", "", "", nil)

	if err := g.CastSpell(me.ID, spiderSense, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "web_slinging",
		AltCostIDs:      []uuid.UUID{creature},
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err == nil {
		t.Error("web-slinging with an untapped creature should be refused")
	}
	if !g.Battlefield.Contains(creature) {
		t.Error("a rejected payment must not have moved the creature")
	}
}
