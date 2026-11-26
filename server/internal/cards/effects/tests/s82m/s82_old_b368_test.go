package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OLD batch 368 (S82, #5558): Journey into Nyx bestow cards — Sightless
// Brawler, Gnarled Scarhide, Hypnotic Siren, Crystalline Nautilus, Mogis's
// Warhound and Spirespine.

const (
	oldB368Brawler  = "0f7d76a9-3f47-4cfa-882b-1790256d15b2"
	oldB368Scarhide = "1ef73cc0-fe72-4040-a932-bcd4b1a704a3"
	oldB368Siren    = "54aede54-4f7d-40f5-bad0-2448293e2cf0"
	oldB368Nautilus = "77661282-1c98-4f70-b617-c75633ab04c8"
	oldB368Warhound = "bb87917f-33fb-4024-9317-0992c2f8f55e"
	oldB368Spire    = "de0edcc8-1595-4dbb-ac8a-46dcd96299cb"
)

func TestOldB368CardsAreRegisteredFullWithTheirBestowOffer(t *testing.T) {
	for oracle, want := range map[string]struct{ name, cost string }{
		oldB368Brawler:  {"Sightless Brawler", "{4}{W}"},
		oldB368Scarhide: {"Gnarled Scarhide", "{3}{B}"},
		oldB368Siren:    {"Hypnotic Siren", "{5}{U}{U}"},
		oldB368Nautilus: {"Crystalline Nautilus", "{3}{U}{U}"},
		oldB368Warhound: {"Mogis's Warhound", "{2}{R}"},
		oldB368Spire:    {"Spirespine", "{4}{G}"},
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != want.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", want.name, spec)
			continue
		}
		alt := game.AlternativeCostByKey(oracle, game.BestowKey)
		if alt == nil || !alt.Bestow || alt.ManaCost != want.cost || alt.Targets == nil {
			t.Errorf("%s: bestow offer %+v, want Bestow %s", want.name, alt, want.cost)
		}
	}
}

// oldB368Bestowed casts the card for its bestow cost onto a fresh 2/2 host the
// active seat controls (or `hostOwner` does) and settles the spell.
func oldB368Bestowed(t *testing.T, g *game.Game, name, oracle string, p, tough int, hostOwner uuid.UUID) (aura, host uuid.UUID) {
	t.Helper()
	aura = effectstest.BestowCard(t, g, name, "Enchantment Creature — Test", oracle, p, tough)
	host = effectstest.B12Creature(g, hostOwner, "Host", "Creature — Bear", 2, 2)
	effectstest.Unsick(g, host)
	effectstest.CastBestowed(t, g, aura, host)
	effectstest.PassPriorityAroundTable(t, g)
	s := effectstest.ReadBestowState(g, aura)
	if !s.OnBattlefield || s.AttachedTo != host || s.Creature {
		t.Fatalf("%s did not end up attached to the host as an Aura: %+v", name, s)
	}
	return aura, host
}

func TestOldB368GnarledScarhideBestowedPumpsAndStopsBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_, host := oldB368Bestowed(t, g, "Gnarled Scarhide", oldB368Scarhide, 2, 1, me.ID)
	if p, tough := effectstest.PtOf(t, g, host); p != 4 || tough != 3 {
		t.Errorf("the host is %d/%d, want 4/3", p, tough)
	}
	if effectstest.AuraRestrictions(t, g, host)&game.CantBlock == 0 {
		t.Error("the host can still block")
	}
}

func TestOldB368GnarledScarhideAsACreatureCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Gnarled Scarhide", "Enchantment Creature — Minotaur", oldB368Scarhide, 2, 1)
	if effectstest.AuraRestrictions(t, g, id)&game.CantBlock == 0 {
		t.Error("the Scarhide can block")
	}
}

func TestOldB368HypnoticSirenStealsPumpsAndGivesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	_, host := oldB368Bestowed(t, g, "Hypnotic Siren", oldB368Siren, 1, 1, opp.ID)
	if c := effectstest.LayeredCard(t, g, host); c.Controller != me.ID {
		t.Error("I do not control the enchanted creature")
	}
	if p, tough := effectstest.PtOf(t, g, host); p != 3 || tough != 3 {
		t.Errorf("the host is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the host has no flying")
	}
}

func TestOldB368HypnoticSirenIsAFlyingCreatureWhenCastNormally(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Hypnotic Siren", "Enchantment Creature — Siren", oldB368Siren, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("the Siren has no flying")
	}
}

func TestOldB368SightlessBrawlerHostCantAttackAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	_, host := oldB368Bestowed(t, g, "Sightless Brawler", oldB368Brawler, 3, 2, me.ID)
	if p, tough := effectstest.PtOf(t, g, host); p != 5 || tough != 4 {
		t.Errorf("the host is %d/%d, want 5/4", p, tough)
	}
	companion := effectstest.B12Creature(g, me.ID, "Companion", "Creature — Bear", 2, 2)
	effectstest.Unsick(g, companion)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the host attacking alone: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(companion, opp.ID); err != nil {
		t.Fatalf("the companion was refused: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("passing with two attackers: %v", err)
	}
}

func TestOldB368SightlessBrawlerCreatureCantAttackAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	id := effectstest.B12Push(g, me.ID, "Sightless Brawler", "Enchantment Creature — Human Warrior", oldB368Brawler, 3, 2)
	effectstest.Unsick(g, id)
	companion := effectstest.B12Creature(g, me.ID, "Companion", "Creature — Bear", 2, 2)
	effectstest.Unsick(g, companion)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the Brawler alone: %v, want ErrAttackLimit", err)
	}
}

func TestOldB368MogissWarhoundHostAttacksEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	_, host := oldB368Bestowed(t, g, "Mogis's Warhound", oldB368Warhound, 2, 2, me.ID)
	if p, tough := effectstest.PtOf(t, g, host); p != 4 || tough != 4 {
		t.Errorf("the host is %d/%d, want 4/4", p, tough)
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != host {
		t.Fatalf("pass with the host at home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

func TestOldB368MogissWarhoundCreatureAttacksEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := effectstest.B12Push(g, me.ID, "Mogis's Warhound", "Enchantment Creature — Dog", oldB368Warhound, 2, 2)
	effectstest.Unsick(g, id)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != id {
		t.Fatalf("pass with the Warhound at home = %v, want a refusal naming it", err)
	}
}

func TestOldB368SpirespineBlocksEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_, host := oldB368Bestowed(t, g, "Spirespine", oldB368Spire, 4, 1, me.ID)
	if p, tough := effectstest.PtOf(t, g, host); p != 6 || tough != 3 {
		t.Errorf("the host is %d/%d, want 6/3", p, tough)
	}
	if reqs := effectstest.LayeredCard(t, g, host).Effective().BlockRequirements; len(reqs) != 1 || reqs[0].Kind != game.BlockRequirementBlocks {
		t.Errorf("the host's block requirements are %+v, want one blocks-each-combat", reqs)
	}
	id := effectstest.B12Push(g, me.ID, "Spirespine", "Enchantment Creature — Beast", oldB368Spire, 4, 1)
	if reqs := effectstest.LayeredCard(t, g, id).Effective().BlockRequirements; len(reqs) != 1 {
		t.Errorf("the Spirespine's block requirements are %+v, want one", reqs)
	}
}

func TestOldB368CrystallineNautilusSacrificesItselfAndItsHostWhenTargeted(t *testing.T) {
	// As a creature.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	creature := effectstest.B12Push(g, me.ID, "Crystalline Nautilus", "Enchantment Creature — Nautilus", oldB368Nautilus, 4, 4)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTargets(creature))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(creature) {
		t.Error("the Nautilus survived being targeted")
	}

	// Bestowed: the host is pumped and sacrificed when targeted, taking the Aura with it.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	aura, host := oldB368Bestowed(t, g, "Crystalline Nautilus", oldB368Nautilus, 4, 4, me.ID)
	if p, tough := effectstest.PtOf(t, g, host); p != 6 || tough != 6 {
		t.Errorf("the host is %d/%d, want 6/6", p, tough)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardTargets(host))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the host survived being targeted")
	}
	if g.Battlefield.Contains(aura) && effectstest.ReadBestowState(g, aura).AttachedTo == host {
		t.Error("the bestowed Aura is still attached to a host that is gone")
	}
}
