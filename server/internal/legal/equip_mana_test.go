package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// equip_mana_test.go — #4600, the enumerator half of Ronin, Shadow
// Stalker's "Spend this mana only to cast Equipment spells or activate
// equip abilities" (CR 106.6, CR 702.6). The enumerator prices an
// activation under the spend context the activation pays under, so an
// equip ability is a move for mana only an equip ability may spend, and
// the Equipment's other abilities are not. The engine half is
// game/equip_mana_test.go; the card is in
// cards/effects/ronin_shadow_stalker_test.go.

const (
	oracleRonin       = "8ceb1acf-5ff3-4f1a-af04-9f07ad454526"
	oracleShadowspear = "8b27326f-e7b8-4a4d-b589-df459246d19a"
)

func TestEquipOnlyManaMakesAnEquipAbilityAMoveAndNothingElse(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	ronin := battlefieldCard(g, active, game.Card{
		Name: "Ronin, Shadow Stalker", TypeLine: "Legendary Creature — Human Rogue Hero",
		Power: 3, Toughness: 3, OracleID: oracleRonin,
	})
	spear := battlefieldCard(g, active, game.Card{
		Name: "Shadowspear", TypeLine: "Legendary Artifact — Equipment", OracleID: oracleShadowspear,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateManaAbility(active.ID, ronin, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("Ronin's mana: %v", err)
	}

	acts := activationsOf(legal.EnumerateFor(g, active.ID), spear)
	if len(acts) == 0 {
		t.Fatal("Ronin's mana pays Shadowspear's equip {2}: want it offered")
	}
	for _, m := range acts {
		if !strings.Contains(m.Label, "Equip {2}") {
			t.Errorf("%q is offered for equip-only mana: the Equipment's other ability is not an equip ability", m.Label)
		}
	}
	dispatchAll(t, g, active.ID, acts)
}

// Casts: with Ronin's mana floating, an Equipment spell is a move and a
// creature spell and a plain artifact spell are not. The auto-tapper
// never plans Ronin's own ability (a life payment with no {T} is the
// player's to make, the rule for every "Pay N life" mana ability that
// does not tap), so before it is fired no cast is offered at all.
func TestRoninManaMakesOnlyEquipmentSpellsMoves(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	ronin := battlefieldCard(g, active, game.Card{
		Name: "Ronin, Shadow Stalker", TypeLine: "Legendary Creature — Human Rogue Hero",
		Power: 3, Toughness: 3, OracleID: oracleRonin,
	})
	blade := handCard(active, game.Card{Name: "Test Blade", TypeLine: "Artifact — Equipment", ManaCost: "{1}{R}"})
	bear := handCard(active, game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{R}"})
	rock := handCard(active, game.Card{Name: "Test Rock", TypeLine: "Artifact", ManaCost: "{1}{R}"})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	for _, id := range []uuid.UUID{blade, bear, rock} {
		if len(castMovesOf(moves, id)) != 0 {
			t.Fatal("a cast is offered before Ronin's mana is made: the auto-tapper must not plan a life payment")
		}
	}

	if err := g.ActivateManaAbility(active.ID, ronin, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("Ronin's mana: %v", err)
	}
	moves = legal.EnumerateFor(g, active.ID)
	if len(castMovesOf(moves, blade)) == 0 {
		t.Error("an Equipment spell is not offered for Ronin's mana")
	}
	if len(castMovesOf(moves, bear)) != 0 {
		t.Error("a creature spell is offered for Equipment-only mana")
	}
	if len(castMovesOf(moves, rock)) != 0 {
		t.Error("a non-Equipment artifact spell is offered for Equipment-only mana")
	}
	dispatchAll(t, g, active.ID, castMovesOf(moves, blade))
}
