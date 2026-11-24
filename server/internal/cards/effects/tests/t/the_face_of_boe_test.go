package testst

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	cardsth "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/th"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	faceOfBoeOracle   = "acef7666-53b1-43c5-8107-236bea22a8b3"
	faceOfBoeRiftBolt = "2b8afa9f-4236-4c02-a8d5-3c145caecfd6"
)

func TestTheFaceOfBoeIsRegisteredFull(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, faceOfBoeOracle, "The Face of Boe")
}

func faceOfBoeTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	boe := effectstest.PushCatalogPermanent(g, me.ID, "The Face of Boe", "Legendary Creature — Alien Advisor", faceOfBoeOracle, false)
	bolt := effectstest.HandCardFull(me, "Rift Bolt", "Sorcery", "{2}{R}", faceOfBoeRiftBolt, []string{"R"})
	plain := effectstest.HandSpell(me, "Plain Sorcery", "Sorcery", "{R}")
	return g, me, opp, boe, bolt, plain
}

// The offer names only the card with suspend, and casting it costs its
// suspend cost, {R}, rather than {2}{R}. It is a normal spell: it deals
// its 3 damage.
func TestTheFaceOfBoeCastsASuspendCardForItsSuspendCost(t *testing.T) {
	g, me, opp, boe, bolt, plain := faceOfBoeTable(t)
	if err := g.ActivateCatalogAbility(me.ID, boe, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ResolutionCastOfferFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no offer")
	}
	if !slices.Contains(prompt.DiscardOptions, bolt) || slices.Contains(prompt.DiscardOptions, plain) {
		t.Fatalf("offer %v, want only the suspend card", prompt.DiscardOptions)
	}
	effectstest.FloatForTest(g, me, "R")
	life := opp.Life
	if err := g.CastForResolutionChoice(me.ID, prompt.ID, bolt, game.CastSpellParams{
		AlternativeCost: cardsth.FaceOfBoeAltCostKey,
		Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast for {R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-3)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("the offer is still open")
	}
}

// Activate only as a sorcery; and declining keeps the card in hand, with
// no cast left for later.
func TestTheFaceOfBoeIsSorcerySpeedAndDeclineKeepsTheCard(t *testing.T) {
	g, me, opp, boe, bolt, _ := faceOfBoeTable(t)
	if err := g.ActivateCatalogAbility(me.ID, boe, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if z := effectstest.E2Zone(g, bolt); z != game.ZoneHand {
		t.Fatalf("the card is in %q", z)
	}
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{AlternativeCost: cardsth.FaceOfBoeAltCostKey, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("cast for the suspend cost after declining")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.UntapForTest(g, boe)
	if err := g.ActivateCatalogAbility(me.ID, boe, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated outside a main phase")
	}
}
