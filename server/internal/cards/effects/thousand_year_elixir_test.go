package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// thousand_year_elixir_test.go — #3928. "You may activate abilities of
// creatures you control as though those creatures had haste."
//
// CR 302.6 has two halves: a creature that arrived this turn can't pay a
// {T} or {Q} cost, and can't attack. Haste lifts both (CR 702.10b,
// 702.10c). An "as though" effect applies only to what it states
// (CR 609.4), so the Elixir lifts the first half for its controller and
// leaves the second alone. Every reader of the first half (the engine's
// two activation paths, the auto-tapper, the legal-move enumerator and
// the wire) must agree.

const thousandYearElixirOracle = "4dc5726e-2f7e-4c2b-9616-c3301d212f78"

// elixirTable seats the Elixir under `elixirOwner` (when set) and a
// summoning-sick Krenko, Mob Boss and Llanowar Elves under the active
// seat.
func elixirTable(t *testing.T, elixirOwner func(g *game.Game) uuid.UUID) (g *game.Game, me *game.Player, krenko, elves, elixir uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = top100ActiveSeat(g)
	if elixirOwner != nil {
		elixir = pushCatalogPermanent(g, elixirOwner(g), "Thousand-Year Elixir", "Artifact", thousandYearElixirOracle, false)
	}
	krenko = pushCatalogPermanent(g, me.ID, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", krenkoOracle, true)
	elves = pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, true)
	return g, me, krenko, elves, elixir
}

func elixirMine(g *game.Game) uuid.UUID { return top100ActiveSeat(g).ID }

func elixirOpponents(g *game.Game) uuid.UUID { return top100Opponent(g).ID }

func TestThousandYearElixirActivatesSickCreaturesTapAbilities(t *testing.T) {
	g, me, krenko, _, _ := elixirTable(t, elixirMine)
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Krenko's {T} ability the turn he arrived, under the Elixir: %v", err)
	}
	if c, _ := battlefieldCard(g, krenko); !c.Tapped {
		t.Error("Krenko was not tapped to pay the cost")
	}
}

func TestThousandYearElixirActivatesSickCreaturesManaAbilities(t *testing.T) {
	g, me, _, elves, _ := elixirTable(t, elixirMine)
	if err := g.ActivateManaAbility(me.ID, elves, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Llanowar Elves' {T} mana ability the turn it arrived, under the Elixir: %v", err)
	}
}

func TestThousandYearElixirAutoTapperUsesSickCreatures(t *testing.T) {
	cost, err := game.ParseCost("{G}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	g, me, _, elves, _ := elixirTable(t, elixirMine)
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 1 || plan[0] != elves {
		t.Fatalf("auto-tap {G} under the Elixir: plan = %v, ok = %v; want the Elves", plan, ok)
	}
	g, me, _, _, _ = elixirTable(t, nil)
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Fatalf("auto-tap {G} with only a sick Elves and no Elixir: plan = %v, want none", plan)
	}
}

func TestThousandYearElixirWithoutItTheCreaturesStaySick(t *testing.T) {
	for name, owner := range map[string]func(*game.Game) uuid.UUID{
		"no Elixir":            nil,
		"an opponent's Elixir": elixirOpponents,
	} {
		t.Run(name, func(t *testing.T) {
			g, me, krenko, elves, _ := elixirTable(t, owner)
			if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrSummoningSick) {
				t.Errorf("Krenko's {T} ability: err = %v, want ErrSummoningSick", err)
			}
			if err := g.ActivateManaAbility(me.ID, elves, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrSummoningSick) {
				t.Errorf("Elves' mana ability: err = %v, want ErrSummoningSick", err)
			}
		})
	}
}

// TestThousandYearElixirGrantsNoHaste is CR 609.4: the creature still
// can't attack, and it has not gained the keyword.
func TestThousandYearElixirGrantsNoHaste(t *testing.T) {
	g, _, krenko, _, _ := elixirTable(t, elixirMine)
	seat := g.Turn.ActiveSeat
	advanceToStepOf(t, g, seat, game.StepDeclareAttackers)
	makeSick(t, g, krenko)
	c, _ := battlefieldCard(g, krenko)
	if game.HasKeyword(&c, "haste") {
		t.Error("the Elixir gave Krenko haste")
	}
	if err := g.DeclareAttacker(krenko, top100Opponent(g).ID); !errors.Is(err, game.ErrSummoningSick) {
		t.Errorf("Krenko attacking the turn he arrived, under the Elixir: err = %v, want ErrSummoningSick", err)
	}
}

// TestThousandYearElixirEnumeratorAndViewAgree: the bot's moves and the
// client's greying read the engine's gate.
func TestThousandYearElixirEnumeratorAndViewAgree(t *testing.T) {
	for _, withElixir := range []bool{true, false} {
		owner := elixirMine
		if !withElixir {
			owner = nil
		}
		g, me, krenko, elves, _ := elixirTable(t, owner)

		offered := map[uuid.UUID]bool{}
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type == legal.TypeActivateAbility || m.Type == legal.TypeActivateManaAbility {
				offered[m.Source] = true
			}
		}
		if offered[krenko] != withElixir || offered[elves] != withElixir {
			t.Errorf("Elixir %v: enumerator offers Krenko %v, Elves %v; want %v", withElixir, offered[krenko], offered[elves], withElixir)
		}

		// The unfiltered frame: the seeded permanents carry no knowers,
		// so a seat's filtered frame would redact them.
		v := protocol.ViewOfGame(g)
		for _, c := range v.Battlefield.Cards {
			if c.InstanceID != krenko.String() && c.InstanceID != elves.String() {
				continue
			}
			if !c.SummoningSick {
				t.Errorf("Elixir %v: %s lost summoning_sick, which still stops it attacking", withElixir, c.Name)
			}
			if c.ActivatesAsThoughHaste != withElixir || c.ActivationSick() == withElixir {
				t.Errorf("Elixir %v: %s activates_as_though_haste = %v", withElixir, c.Name, c.ActivatesAsThoughHaste)
			}
		}
	}
}

func TestThousandYearElixirUntapsTargetCreature(t *testing.T) {
	g, me, krenko, _, elixir := elixirTable(t, elixirMine)
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Krenko: %v", err)
	}
	passPriorityAroundTable(t, g)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, elixir, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: krenko}},
	}); err != nil {
		t.Fatalf("Elixir's untap: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, krenko); c.Tapped {
		t.Error("the Elixir did not untap Krenko")
	}
}
