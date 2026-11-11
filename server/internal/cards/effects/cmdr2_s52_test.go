package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2S52YarokOracle = "95b53836-18aa-451a-992d-2a111deeeab2"
	cmdr2S52FelixOracle = "95175d35-c815-47bd-8e69-427972b567b6"
)

func TestCmdr2S52YarokDoublesAnyPermanentEntering(t *testing.T) {
	for _, withYarok := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withYarok {
			pushCatalogPermanent(g, me.ID, "Yarok, the Desecrated", "Legendary Creature — Elemental Horror", cmdr2S52YarokOracle, false)
		}
		pushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7", false)
		before := opp.Library.Size()
		// A plain Forest is neither an artifact nor a creature: Panharmonicon
		// would not double it, Yarok does.
		castCatalogSpell(t, g, "Forest", "Basic Land — Forest", "", nil)
		passPriorityAroundTable(t, g)
		got := before - opp.Library.Size()
		want := 3
		if withYarok {
			want = 6
		}
		if got != want {
			t.Errorf("withYarok=%v: milled %d, want %d", withYarok, got, want)
		}
	}
}

func TestCmdr2S52YarokOnlyDoublesItsControllersAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Yarok, the Desecrated", "Legendary Creature — Elemental Horror", cmdr2S52YarokOracle, false)
	pushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7", false)
	before := opp.Library.Size()
	castCatalogSpell(t, g, "Forest", "Basic Land — Forest", "", nil)
	passPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 3 {
		t.Errorf("opponent's Yarok doubled my trigger: milled %d, want 3", got)
	}
}

func TestCmdr2S52FelixDoublesCombatDamageToAPlayerTriggers(t *testing.T) {
	for _, withFelix := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		opp := g.Seats[1]
		bidentID := pushPermanentForTest(g, me.ID, "Bident of Thassa", bidentOracle, "Legendary Enchantment Artifact")
		a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
		if withFelix {
			pushCatalogPermanent(g, me.ID, "Felix Five-Boots", "Legendary Creature — Ooze Rogue", cmdr2S52FelixOracle, false)
		}
		_ = bidentID
		attackWith(t, g, opp.ID, a)
		want := 1
		if withFelix {
			want = 2
		}
		if got := owedMays(g, me.ID); got != want {
			t.Errorf("withFelix=%v: Bident prompts %d, want %d", withFelix, got, want)
		}
	}
}

func TestCmdr2S52FelixIgnoresOpponentsDoublerAndOtherCauses(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Bident of Thassa", bidentOracle, "Legendary Enchantment Artifact")
	pushCatalogPermanent(g, opp.ID, "Felix Five-Boots", "Legendary Creature — Ooze Rogue", cmdr2S52FelixOracle, false)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	attackWith(t, g, opp.ID, a)
	if got := owedMays(g, me.ID); got != 1 {
		t.Errorf("opponent's Felix doubled my trigger: %d prompts, want 1", got)
	}
	spec, ok := Lookup(cmdr2S52FelixOracle)
	if !ok || len(spec.TriggerDoublers) != 1 {
		t.Fatalf("Felix registration: ok=%v", ok)
	}
	// Entering is not a combat-damage cause.
	q := game.TriggerDoublingQuery{Event: game.Event{Kind: game.EventETB}, Ability: &game.TriggeredAbility{}, SourceIsPermanent: true, HasSubject: true}
	if spec.TriggerDoublers[0].Applies(g, q) {
		t.Error("Felix doubled an entering cause")
	}
}
