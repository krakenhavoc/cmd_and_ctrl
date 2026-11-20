package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestAshnodsAltarEatsACreatureForTwoColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	altar := effectstest.SeedPermanentWithOracle(g, me.ID, "Ashnod's Altar", "Artifact", effectstest.AshnodsAltarOracle)
	food := effectstest.SeedCreature(g, "Doomed Traveler", me.ID)

	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{
		SacrificeIDs: []uuid.UUID{food},
	}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}

	if me.ManaPool.Total() != 2 {
		t.Fatalf("mana pool: got %d tokens, want 2", me.ManaPool.Total())
	}
	for i, tok := range me.ManaPool.Units() {
		if tok.Color != "C" {
			t.Errorf("token[%d]: got color %q, want C", i, tok.Color)
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, food); ok {
		t.Error("sacrificed creature still on the battlefield")
	}
	// No tap component — the Altar stays up and can go again.
	card, ok := effectstest.BattlefieldCard(g, altar)
	if !ok {
		t.Fatal("Altar left the battlefield")
	}
	if card.Tapped {
		t.Error("Altar tapped; it has no tap cost")
	}
}

func TestAshnodsAltarActivatesRepeatedlyInOneTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	altar := effectstest.SeedPermanentWithOracle(g, me.ID, "Ashnod's Altar", "Artifact", effectstest.AshnodsAltarOracle)
	first := effectstest.SeedCreature(g, "Bear One", me.ID)
	second := effectstest.SeedCreature(g, "Bear Two", me.ID)

	for i, victim := range []uuid.UUID{first, second} {
		if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{
			SacrificeIDs: []uuid.UUID{victim},
		}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
	}
	if me.ManaPool.Total() != 4 {
		t.Errorf("mana pool: got %d tokens, want 4", me.ManaPool.Total())
	}
}

func TestAshnodsAltarRejectsSacrificingItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	altar := effectstest.SeedPermanentWithOracle(g, me.ID, "Ashnod's Altar", "Artifact", effectstest.AshnodsAltarOracle)
	effectstest.SeedCreature(g, "Bear", me.ID)

	// "Sacrifice a creature" — the Altar is an Artifact, so it is not
	// a legal choice for its own cost.
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{
		SacrificeIDs: []uuid.UUID{altar},
	}); err == nil {
		t.Fatal("sacrificing the Altar to its own cost was accepted")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("mana pool: got %d tokens after a rejected activation, want 0", me.ManaPool.Total())
	}
	if _, ok := effectstest.BattlefieldCard(g, altar); !ok {
		t.Error("Altar left the battlefield on a rejected activation")
	}
}

func TestPhyrexianAltarPromptsForColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	altar := effectstest.SeedPermanentWithOracle(g, me.ID, "Phyrexian Altar", "Artifact", effectstest.PhyrexianAltarOracle)
	food := effectstest.SeedCreature(g, "Doomed Traveler", me.ID)

	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{
		SacrificeIDs: []uuid.UUID{food},
	}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, food); ok {
		t.Error("sacrificed creature still on the battlefield")
	}
	// A multi-option pipe queues a mana_pick rather than dropping a
	// token straight in — the same path Birds of Paradise takes.
	if len(g.PendingChoices) == 0 {
		t.Fatal("no pending choice; expected a mana_pick for the any-color slot")
	}
	got := g.PendingChoices[len(g.PendingChoices)-1]
	if got.Kind != game.PendingChoiceMana {
		t.Errorf("choice kind: got %q, want %q", got.Kind, game.PendingChoiceMana)
	}
	if got.Chooser != me.ID {
		t.Errorf("chooser: got %v, want %v", got.Chooser, me.ID)
	}
}

// TestAltarViewOffersOnlyYourCreatures is the client's half of the
// contract: the picker's candidate list is stamped, and filtered to
// the controller's own creatures (CR 701.21a) — you may not eat an
// opponent's blocker to pay your own cost.
func TestAltarViewOffersOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	altar := effectstest.SeedPermanentWithOracle(g, me.ID, "Ashnod's Altar", "Artifact", effectstest.AshnodsAltarOracle)
	effectstest.KnownToTable(g, altar)
	mine := effectstest.SeedCreature(g, "My Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", them.ID)

	view := protocol.ViewOfGameFor(g, me.ID.String())
	var found *protocol.ManaAbilityView
	for i := range view.Battlefield.Cards {
		c := view.Battlefield.Cards[i]
		if c.InstanceID != altar.String() {
			continue
		}
		if len(c.ManaAbilities) == 0 {
			t.Fatal("Altar view carries no mana abilities")
		}
		found = &c.ManaAbilities[0]
	}
	if found == nil {
		t.Fatal("Altar not found in the battlefield view")
	}
	if found.SacrificeLabel == "" {
		t.Error("SacrificeLabel empty; the client has no banner copy for the picker")
	}
	if found.SacrificeOptions == nil {
		t.Fatal("SacrificeOptions nil; the client picker would have nothing to show")
	}
	cards := found.SacrificeOptions.Cards
	if len(cards) != 1 {
		t.Fatalf("sacrifice options: got %v, want just %v (mine)", cards, mine)
	}
	if cards[0] != mine.String() {
		t.Errorf("sacrifice option: got %s, want %s (mine); theirs is %s",
			cards[0], mine, theirs)
	}
	if len(found.SacrificeOptions.Players) != 0 {
		t.Error("sacrifice options list players; a sacrifice cost takes a permanent")
	}
}

// TestAltarViewHidesOptionsWhenBoardIsEmpty keeps the client's greying
// honest: with nothing to eat, the option list is empty rather than
// listing something unusable.
func TestAltarViewHidesOptionsWhenBoardIsEmpty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	altar := effectstest.SeedPermanentWithOracle(g, me.ID, "Ashnod's Altar", "Artifact", effectstest.AshnodsAltarOracle)
	effectstest.KnownToTable(g, altar)

	view := protocol.ViewOfGameFor(g, me.ID.String())
	seen := false
	for _, c := range view.Battlefield.Cards {
		if c.InstanceID != altar.String() || len(c.ManaAbilities) == 0 {
			continue
		}
		seen = true
		opts := c.ManaAbilities[0].SacrificeOptions
		if opts != nil && len(opts.Cards) != 0 {
			t.Errorf("sacrifice options %v with an empty board", opts.Cards)
		}
	}
	if !seen {
		t.Fatal("Altar's mana ability missing from the view; the check above was vacuous")
	}
}
