package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b512_test.go — Shards of Alara, "Life gain and loss" (S82 OLD batch
// 512, #5588): each card driven through a real cast or activation.

const (
	oldB512AgentOracle   = "06ada754-da78-4eda-8be7-3ac4f06115c4"
	oldB512ChaliceOracle = "236f318e-9bd1-4eb0-8b19-af015b3a74ea"
	oldB512WindOracle    = "2f051ee4-9f2e-414d-b763-5fd78f2d4e1b"
	oldB512GobletOracle  = "3e070d7a-6a1e-4cce-b196-73f4052c049c"
	oldB512GraceOracle   = "42946216-59e2-46dc-b770-5e604512f949"
	oldB512HymnOracle    = "cc7d99e8-f9dc-47d3-b6c0-e9308d33c3c0"
	oldB512JundOracle    = "d2567448-f8a5-4bf3-9804-f5aac3b645ce"
)

func TestS82OldB512AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB512AgentOracle, oldB512ChaliceOracle, oldB512WindOracle, oldB512GobletOracle,
		oldB512GraceOracle, oldB512HymnOracle, oldB512JundOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB512Perm(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.SeedPermanentWithOracle(g, g.Seats[0].ID, name, typeLine, oracle)
	effectstest.UntapForTest(g, id)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return id
}

func oldB512Activate(t *testing.T, g *game.Game, src uuid.UUID, ability int, mana string, targets ...game.TargetRef) error {
	t.Helper()
	if mana != "" {
		effectstest.FloatMana(t, g, g.Seats[0], mana)
	}
	return g.ActivateCatalogAbility(g.Seats[0].ID, src, ability, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

func TestS82OldB512ThoughtcutterAgentDrainsOneAndRevealsTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	agent := oldB512Perm(t, g, "Thoughtcutter Agent", "Artifact Creature — Human Rogue", oldB512AgentOracle)
	card := effectstest.PushCatalogHandCard(opp, "Hidden Card", "Sorcery", "")
	life := opp.Life
	if err := oldB512Activate(t, g, agent, 0, "{U}{B}", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, opp.Life)
	}
	c, ok := g.LookupCardForEffect(card)
	if !ok {
		t.Fatal("the hand card is gone")
	}
	for _, seat := range g.Seats {
		if !c.KnownBy[seat.ID] {
			t.Errorf("seat %v does not know the revealed card", seat.ID)
		}
	}
}

func TestS82OldB512MarbleChaliceGainsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chalice := oldB512Perm(t, g, "Marble Chalice", "Artifact", oldB512ChaliceOracle)
	life := me.Life
	if err := oldB512Activate(t, g, chalice, 0, ""); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
	if err := oldB512Activate(t, g, chalice, 0, ""); err == nil {
		t.Error("activated while tapped")
	}
}

func TestS82OldB512OnyxGobletAndJundBattlemageDrain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	goblet := oldB512Perm(t, g, "Onyx Goblet", "Artifact", oldB512GobletOracle)
	life := opp.Life
	if err := oldB512Activate(t, g, goblet, 0, "", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("goblet: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("goblet: life %d -> %d, want -1", life, opp.Life)
	}
	mage := oldB512Perm(t, g, "Jund Battlemage", "Creature — Human Shaman", oldB512JundOracle)
	if err := oldB512Activate(t, g, mage, 0, "{B}", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("mage drain: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("mage: life %d, want %d", opp.Life, life-2)
	}
	effectstest.UntapForTest(g, mage)
	if err := oldB512Activate(t, g, mage, 1, "{G}"); err != nil {
		t.Fatalf("mage token: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Saproling" && c.CurrentPower() == 1 && c.CurrentToughness() == 1 && c.HasColor("G") {
			found = true
		}
	}
	if !found {
		t.Error("no 1/1 green Saproling token")
	}
}

func TestS82OldB512WindwrightMageFliesOnlyWithAnArtifactCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Windwright Mage", "Artifact Creature — Human Wizard", oldB512WindOracle, 2, 2)
	if effectstest.HasKeywordOf(g, mage, "flying") {
		t.Fatal("flies with an empty graveyard")
	}
	effectstest.PushGraveyardPermanent(opp, "Their Rock", "Artifact", "")
	effectstest.S29Bump(g)
	if effectstest.HasKeywordOf(g, mage, "flying") {
		t.Fatal("an artifact in an opponent's graveyard gave it flying")
	}
	effectstest.PushGraveyardPermanent(me, "My Rock", "Artifact", "")
	effectstest.S29Bump(g)
	if !effectstest.HasKeywordOf(g, mage, "flying") {
		t.Error("no flying with an artifact card in its controller's graveyard")
	}
	if spec, _ := effects.Lookup(oldB512WindOracle); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "lifelink" {
		t.Errorf("printed keywords %v, want lifelink", spec.PrintedKeywords)
	}
}

func TestS82OldB512SoulsGraceGainsLifeEqualToThePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Wurm", 5, 5)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Soul's Grace", "Instant", oldB512GraceOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d -> %d, want +5", life, me.Life)
	}
}

func TestS82OldB512InvincibleHymnSetsLifeToTheLibrarySize(t *testing.T) {
	for _, start := range []int{3, 90} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		g.WithWriteLock(func() { me.Life = start })
		size := me.Library.Size()
		effectstest.CastCatalogSpell(t, g, "Invincible Hymn", "Sorcery", oldB512HymnOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != size {
			t.Errorf("from %d: life %d, want library size %d", start, me.Life, size)
		}
	}
}
