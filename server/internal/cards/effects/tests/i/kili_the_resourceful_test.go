package testsi

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// kiliTable is seat 0 at its precombat main with Kíli and a Bonesplitter,
// and a Bear to equip. No mana in the pool.
func kiliTable(t *testing.T, story bool) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Kíli the Resourceful", OracleID: effectstest.KiliTheResourcefulOracle,
		TypeLine: "Legendary Creature — Dwarf Scout", Power: 1, Toughness: 2})
	axe := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Bonesplitter", OracleID: effectstest.KiliBonesplitterOracle,
		TypeLine: "Artifact — Equipment"})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if story {
		effectstest.GiveStory(g, me)
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	return g, me, axe, bear
}

// With a story and no equip yet this turn, the {0} twin equips for free;
// it is then spent for the turn, and the normal row still costs {1}.
func TestHobKiliPaysZeroForTheFirstEquipEachTurn(t *testing.T) {
	g, me, axe, bear := kiliTable(t, true)
	row := effectstest.EquipRow(t, g, axe)
	tw, ok := effectstest.TwinOf(g, me.ID, axe, row)
	if !ok {
		t.Fatal("no {0} twin with an enduring story and no equip yet")
	}
	if err := g.ActivateCatalogAbility(me.ID, axe, tw.Index, game.ActivateAbilityParams{
		Ref: tw.Ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("the {0} equip with an empty pool: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, axe, bear) {
		t.Fatal("Bonesplitter is not attached to the Bear")
	}
	if _, ok := effectstest.TwinOf(g, me.ID, axe, row); ok {
		t.Error("the twin is still offered after the first equip of the turn")
	}
	err := g.ActivateCatalogAbility(me.ID, axe, tw.Index, game.ActivateAbilityParams{
		Ref: tw.Ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if err == nil {
		t.Error("a second {0} equip in one turn was allowed")
	}
}

// Activating the normal row first spends the alternative: it is for the
// FIRST equip ability you activate.
func TestHobKiliAPaidEquipFirstSpendsTheAlternative(t *testing.T) {
	g, me, axe, bear := kiliTable(t, true)
	row := effectstest.EquipRow(t, g, axe)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, axe, row, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("the paid equip: %v", err)
	}
	if _, ok := effectstest.TwinOf(g, me.ID, axe, row); ok {
		t.Error("the {0} twin is offered after a paid equip this turn")
	}
}

// Without a story there is no alternative, and a twin's index is refused.
func TestHobKiliOffersNothingWithoutAStory(t *testing.T) {
	g, me, axe, bear := kiliTable(t, false)
	row := effectstest.EquipRow(t, g, axe)
	if _, ok := effectstest.TwinOf(g, me.ID, axe, row); ok {
		t.Fatal("a {0} twin without an enduring story")
	}
	err := g.ActivateCatalogAbility(me.ID, axe, game.AltEquipIndexBase+row, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if !errors.Is(err, game.ErrAlternativeCostUnavailable) {
		t.Errorf("a twin without a story: %v, want ErrAlternativeCostUnavailable", err)
	}
}

// A new turn opens the alternative again.
func TestHobKiliAlternativeReturnsNextTurn(t *testing.T) {
	g, me, axe, bear := kiliTable(t, true)
	row := effectstest.EquipRow(t, g, axe)
	tw, _ := effectstest.TwinOf(g, me.ID, axe, row)
	if err := g.ActivateCatalogAbility(me.ID, axe, tw.Index, game.ActivateAbilityParams{
		Ref: tw.Ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("first {0} equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AroundTheTableToMain(t, g)
	if _, ok := effectstest.TwinOf(g, me.ID, axe, row); !ok {
		t.Error("the alternative did not come back the next turn")
	}
}

// The twin keeps the mirrored row's target clause: Blackblade's "Equip
// legendary creature {3}" paid with {0} still equips only a legendary
// creature; its "Equip {7}" twin equips anything.
func TestHobKiliTwinKeepsTheRowsTargets(t *testing.T) {
	g, me, _, bear := kiliTable(t, true)
	blade := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Blackblade Reforged", OracleID: effectstest.KiliBlackbladeOracle,
		TypeLine: "Legendary Artifact — Equipment"})
	rows := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.KiliBlackbladeOracle})
	legendaryRow, anyRow := -1, -1
	for i, r := range rows {
		if r.Equip && r.Label == "Equip legendary creature {3}" {
			legendaryRow = i
		} else if r.Equip {
			anyRow = i
		}
	}
	tw, ok := effectstest.TwinOf(g, me.ID, blade, legendaryRow)
	if !ok || anyRow < 0 {
		t.Fatal("setup: Blackblade's twins are missing")
	}
	if err := g.Clone().ActivateCatalogAbility(me.ID, blade, tw.Index, game.ActivateAbilityParams{
		Ref: tw.Ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("the legendary-creature twin equipped a nonlegendary Bear")
	}
	other, _ := effectstest.TwinOf(g, me.ID, blade, anyRow)
	if err := g.ActivateCatalogAbility(me.ID, blade, other.Index, game.ActivateAbilityParams{
		Ref: other.Ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Errorf("the Equip {7} twin at {0}: %v", err)
	}
}

// The enumerator offers the twin's move, addressed by the twin's index and
// ref, and only while the alternative is open; the move it sends works.
func TestHobKiliTwinIsOfferedToTheEnumerator(t *testing.T) {
	g, me, axe, bear := kiliTable(t, true)
	row := effectstest.EquipRow(t, g, axe)
	tw, _ := effectstest.TwinOf(g, me.ID, axe, row)
	if tw.Shape.Cost.Mana != "{0}" || !tw.Shape.Equip {
		t.Errorf("twin shape cost %q equip %v, want {0} and an equip", tw.Shape.Cost.Mana, tw.Shape.Equip)
	}
	offered := func() bool {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != "activate_ability" || m.Source != axe {
				continue
			}
			var p struct {
				Index int    `json:"ability_index"`
				Ref   string `json:"ref"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("move params: %v", err)
			}
			if p.Index == tw.Index && p.Ref == tw.Ref {
				return true
			}
		}
		return false
	}
	if !offered() {
		t.Fatal("the enumerator does not offer the {0} twin")
	}
	if err := g.ActivateCatalogAbility(me.ID, axe, tw.Index, game.ActivateAbilityParams{
		Ref: tw.Ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("the offered twin: %v", err)
	}
	if offered() {
		t.Error("the enumerator still offers the twin after the first equip")
	}
}

// Another Dwarf or an Equipment entering draws, once a turn.
func TestHobKiliDrawsOnceATurnForADwarfOrEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Kíli the Resourceful", OracleID: effectstest.KiliTheResourcefulOracle,
		TypeLine: "Legendary Creature — Dwarf Scout", Power: 1, Toughness: 2})
	hand := me.Hand.Size()
	effectstest.HobEnter(t, g, me.ID, game.Card{Name: "Some Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("Kíli drew for an Elf")
	}
	effectstest.HobEnter(t, g, me.ID, game.Card{Name: "Spare Sword", TypeLine: "Artifact — Equipment"})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after an Equipment entered, want %d", me.Hand.Size(), hand+1)
	}
	effectstest.HobEnter(t, g, me.ID, game.Card{Name: "Some Dwarf", TypeLine: "Creature — Dwarf", Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Kíli drew twice in one turn: hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// The table sees the twin as a row of its own: its index and ref, a {0}
// cost, a label that names the alternative, and the mirrored row's
// targets.
func TestHobKiliTwinIsARowOnTheWire(t *testing.T) {
	g, me, axe, bear := kiliTable(t, true)
	row := effectstest.EquipRow(t, g, axe)
	tw, _ := effectstest.TwinOf(g, me.ID, axe, row)
	v := effectstest.ActivatedRowOf(t, g, me.ID, axe, tw.Index)
	if v.Ref != tw.Ref || v.ManaCost != "{0}" || !v.Equip {
		t.Errorf("twin row ref %q cost %q equip %v, want %q, {0}, true", v.Ref, v.ManaCost, v.Equip, tw.Ref)
	}
	if !strings.Contains(v.Label, "pay {0} instead (Kíli the Resourceful)") {
		t.Errorf("twin row label %q does not name the alternative", v.Label)
	}
	if v.LegalTargets == nil || !slices.Contains(v.LegalTargets.Cards, bear.String()) {
		t.Errorf("twin row targets %+v do not offer the Bear", v.LegalTargets)
	}
}
