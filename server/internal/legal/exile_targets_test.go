package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// exile_targets_test.go -- #3259. "Target creature card exiled with The
// Darkness Crystal" is one set everywhere: the enumerator the bot picks
// from, the legal_targets the client's picker highlights, and the
// engine's announce check all walk the same clause over the exile zone,
// narrowed by TargetSpec.ExiledWithSource.

const oracleDarknessCrystal = "660f370a-b25b-4ea5-b765-7abf96899c1c"

func exiledCard(g *game.Game, owner *game.Player, name, typeLine string, by game.PermissionCardRef) uuid.UUID {
	id := uuid.New()
	known := make(map[uuid.UUID]bool, len(g.Seats))
	for _, s := range g.Seats {
		known[s.ID] = true
	}
	g.Exile.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2,
		Owner: owner.ID, Controller: owner.ID, ExiledWith: by, KnownBy: known,
	})
	return id
}

func TestDarknessCrystalEnumeratorAndViewOfferOnlyItsOwnExiledCreatures(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	crystal := battlefieldCard(g, active, game.Card{
		Name: "The Darkness Crystal", TypeLine: "Legendary Artifact", OracleID: oracleDarknessCrystal,
	})
	for i := 0; i < 6; i++ {
		battlefieldCard(g, active, basic("Swamp", "Swamp"))
	}
	mine := exiledCard(g, opp, "Linked Bear", "Creature -- Bear", game.PermissionCardRef{ID: crystal})
	other := exiledCard(g, opp, "Other Bear", "Creature -- Bear", game.PermissionCardRef{ID: uuid.New()})
	unlinked := exiledCard(g, opp, "Loose Bear", "Creature -- Bear", game.PermissionCardRef{})
	land := exiledCard(g, opp, "Linked Land", "Land", game.PermissionCardRef{ID: crystal})
	onBoard := battlefieldCard(g, opp, creature("Board Bear", "{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	seen := boundedTargetsOffered(t, moves, crystal, map[string]bool{mine.String(): true})
	if !seen[mine.String()] {
		t.Errorf("the Crystal's own exiled creature was never offered: %v", labels(moves))
	}
	for name, id := range map[string]uuid.UUID{"another source's": other, "unlinked": unlinked, "linked land": land, "battlefield": onBoard} {
		if seen[id.String()] {
			t.Errorf("the %s card was offered as a target", name)
		}
	}
	dispatchAll(t, g, active.ID, moves)

	v := protocol.ViewOfGameFor(g, active.ID.String())
	var cards []string
	var mode string
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != crystal.String() {
			continue
		}
		for _, ab := range c.ActivatedAbilities {
			if ab.LegalTargets != nil && len(ab.LegalTargets.Cards) > 0 {
				cards, mode = ab.LegalTargets.Cards, ab.TargetMode
			}
		}
	}
	if len(cards) != 1 || cards[0] != mine.String() {
		t.Errorf("view legal_targets = %v, want only %s", cards, mine)
	}
	if mode != "card_in_exile" {
		t.Errorf("view target_mode = %q, want card_in_exile", mode)
	}

	// A card that leaves exile is no longer offered.
	if _, err := g.Exile.Remove(mine); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	for _, m := range activationsOf(legal.EnumerateFor(g, active.ID), crystal) {
		if len(boundTargetIDs(t, m)) > 0 {
			t.Errorf("%q still offers a target once the exile is empty", m.Label)
		}
	}
}
