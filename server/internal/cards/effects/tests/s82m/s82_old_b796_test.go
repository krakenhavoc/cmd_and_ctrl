package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b796_test.go — Mirage, batch 796 (#5633). Every test plays the real
// catalog entry through the engine.

const (
	oldB796Alarum   = "18e96674-6fd8-4046-a3fe-7ba80138d85d"
	oldB796Shaman   = "2f14e92d-307f-4854-b7ec-a2f6ce03dced"
	oldB796Brush    = "33a981ac-878d-4257-936f-55b4e6543c12"
	oldB796Wurm     = "457705c9-5b0c-4c3a-89dd-80286fac6a6c"
	oldB796Essence  = "48e2f2c3-9d99-4535-915f-551fbb59ade5"
	oldB796Zuberi   = "5a43b164-61a4-4dcb-9397-45356ad4260e"
	oldB796Raid     = "5b4b08a5-3af8-4c33-ab5e-7bee7f79e77a"
	oldB796Civic    = "6e93cb9a-44c8-49fb-b6a7-bfdf7bd725fe"
	oldB796Dragon   = "6fb6e37e-d774-4a50-8ea3-b955535c2065"
	oldB796Salamand = "745a2815-8487-4b78-93b7-d1d4ef6adc8f"
	oldB796Hordes   = "79bfef02-3586-424a-aff9-a6ec0d4f1135"
)

func TestOldB796EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB796Alarum: "Alarum", oldB796Shaman: "Blighted Shaman", oldB796Brush: "Brushwagg",
		oldB796Wurm: "Jungle Wurm", oldB796Essence: "Radiant Essence", oldB796Zuberi: "Zuberi, Golden Feather",
		oldB796Raid: "Nocturnal Raid", oldB796Civic: "Civic Guildmage", oldB796Dragon: "Teeka's Dragon",
		oldB796Salamand: "Pyric Salamander", oldB796Hordes: "Horrible Hordes",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB796Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB796Creature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int, colors []string, kws ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		OracleID: oracle, Power: p, Toughness: tough, Colors: colors, Keywords: kws, Owner: owner, Controller: owner})
}

func oldB796Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[0]
	c := game.NewCard(name, me.ID)
	c.TypeLine, c.OracleID = typeLine, oracle
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		return err
	}
	effectstest.PassPriorityAroundTable(t, g)
	return nil
}

func TestOldB796AlarumUntapsAndPumpsANonattackingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := oldB796Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, []string{"G"})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := oldB796Cast(t, g, "Alarum", "Instant", oldB796Alarum, oldB796Card(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the Bear is still tapped")
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 3/5", p, tough)
	}
}

func TestOldB796AlarumRefusesAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := oldB796Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, []string{"G"})
	effectstest.Pr7bAttack(t, g, opp.ID, bear)
	if err := oldB796Cast(t, g, "Alarum", "Instant", oldB796Alarum, oldB796Card(bear)); err == nil {
		t.Error("targeted an attacking creature")
	}
}

func TestOldB796BlightedShamanSacrificesASwampOrACreatureToPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	s := effectstest.BlbB1Push(g, me.ID, "Blighted Shaman", "Creature — Human Cleric Shaman", oldB796Shaman, 1, 1)
	swamp := effectstest.PushCatalogPermanent(g, me.ID, "Swamp", "Basic Land — Swamp", "", false)
	bear := oldB796Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, []string{"G"})
	effectstest.B16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{Strict: true,
		Targets: oldB796Card(bear), SacrificeIDs: []uuid.UUID{swamp}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(swamp) {
		t.Error("the Swamp was not sacrificed")
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
	}
}

func TestOldB796BlightedShamanSacrificedCreatureGivesPlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	s := effectstest.BlbB1Push(g, me.ID, "Blighted Shaman", "Creature — Human Cleric Shaman", oldB796Shaman, 1, 1)
	fodder := oldB796Creature(g, me.ID, "Fodder", "Creature — Rat", "", 1, 1, []string{"B"})
	bear := oldB796Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, []string{"G"})
	effectstest.B16Activate(t, g, me.ID, s, 1, game.ActivateAbilityParams{Strict: true,
		Targets: oldB796Card(bear), SacrificeIDs: []uuid.UUID{fodder}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB796BlightedShamanFirstAbilityWantsASwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	s := effectstest.BlbB1Push(g, me.ID, "Blighted Shaman", "Creature — Human Cleric Shaman", oldB796Shaman, 1, 1)
	forest := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	bear := oldB796Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, []string{"G"})
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true,
		Targets: oldB796Card(bear), SacrificeIDs: []uuid.UUID{forest}}); err == nil {
		t.Error("sacrificed a Forest to a Swamp cost")
	}
}

// oldB796Combat has `attacker` (the active seat's) attacked and blocked by `blockers`.
func oldB796Combat(t *testing.T, g *game.Game, attacker uuid.UUID, blockers ...uuid.UUID) {
	t.Helper()
	opp := g.Seats[1]
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	blocks := map[uuid.UUID]uuid.UUID{}
	for _, b := range blockers {
		blocks[b] = attacker
	}
	if len(blocks) > 0 {
		effectstest.Pr7bBlock(t, g, blocks)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB796BrushwaggGetsMinusTwoPlusTwoWhenBlockedOrBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	b := effectstest.BlbB1Push(g, me.ID, "Brushwagg", "Creature — Brushwagg", oldB796Brush, 3, 2)
	wall := oldB796Creature(g, opp.ID, "Wall", "Creature — Wall", "", 0, 9, []string{"W"})
	oldB796Combat(t, g, b, wall)
	if p, tough := effectstest.PtOf(t, g, b); p != 1 || tough != 4 {
		t.Errorf("the blocked Brushwagg is %d/%d, want 1/4", p, tough)
	}
}

func TestOldB796BrushwaggBlockingGetsTheSameBonus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	atk := oldB796Creature(g, me.ID, "Ogre", "Creature — Ogre", "", 2, 9, []string{"R"})
	b := effectstest.BlbB1Push(g, opp.ID, "Brushwagg", "Creature — Brushwagg", oldB796Brush, 3, 2)
	oldB796Combat(t, g, atk, b)
	if p, tough := effectstest.PtOf(t, g, b); p != 1 || tough != 4 {
		t.Errorf("the blocking Brushwagg is %d/%d, want 1/4", p, tough)
	}
}

func TestOldB796JungleWurmShrinksForEachBlockerBeyondTheFirst(t *testing.T) {
	for _, tc := range []struct {
		name     string
		blockers int
		want     int
	}{{"one blocker", 1, 5}, {"three blockers", 3, 3}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			w := effectstest.BlbB1Push(g, me.ID, "Jungle Wurm", "Creature — Wurm", oldB796Wurm, 5, 5)
			var bl []uuid.UUID
			for i := 0; i < tc.blockers; i++ {
				bl = append(bl, oldB796Creature(g, opp.ID, "Wall", "Creature — Wall", "", 0, 9, []string{"W"}))
			}
			oldB796Combat(t, g, w, bl...)
			if p, tough := effectstest.PtOf(t, g, w); p != tc.want || tough != tc.want {
				t.Errorf("the Wurm is %d/%d, want %d/%d", p, tough, tc.want, tc.want)
			}
		})
	}
}

func TestOldB796RadiantEssenceWantsAnOpposingBlackPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	e := effectstest.BlbB1Push(g, me.ID, "Radiant Essence", "Creature — Spirit", oldB796Essence, 2, 3)
	oldB796Creature(g, me.ID, "My Zombie", "Creature — Zombie", "", 1, 1, []string{"B"})
	if p, tough := effectstest.PtOf(t, g, e); p != 2 || tough != 3 {
		t.Fatalf("with only your own black permanent the Essence is %d/%d, want 2/3", p, tough)
	}
	black := oldB796Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", "", 1, 1, []string{"B"})
	if p, tough := effectstest.PtOf(t, g, e); p != 3 || tough != 5 {
		t.Errorf("with an opposing black permanent the Essence is %d/%d, want 3/5", p, tough)
	}
	effectstest.P1Destroy(t, g, black)
	if p, tough := effectstest.PtOf(t, g, e); p != 2 || tough != 3 {
		t.Errorf("after it left the Essence is %d/%d, want 2/3", p, tough)
	}
}

func TestOldB796ZuberiPumpsOtherGriffinsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	z := oldB796Creature(g, me.ID, "Zuberi, Golden Feather", "Legendary Creature — Griffin", oldB796Zuberi, 3, 3, []string{"W"}, "flying")
	mine := oldB796Creature(g, me.ID, "Griffin", "Creature — Griffin", "", 2, 2, []string{"W"})
	theirs := oldB796Creature(g, opp.ID, "Their Griffin", "Creature — Griffin", "", 2, 2, []string{"W"})
	bear := oldB796Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, []string{"G"})
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"Zuberi", z, 3}, {"your Griffin", mine, 3}, {"their Griffin", theirs, 3}, {"a Bear", bear, 2}} {
		if p, tough := effectstest.PtOf(t, g, tc.id); p != tc.want || tough != tc.want {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.want, tc.want)
		}
	}
}

func TestOldB796NocturnalRaidPumpsBlackCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := oldB796Creature(g, me.ID, "My Zombie", "Creature — Zombie", "", 1, 1, []string{"B"})
	theirs := oldB796Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", "", 1, 1, []string{"B"})
	green := oldB796Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, []string{"G"})
	if err := oldB796Cast(t, g, "Nocturnal Raid", "Instant", oldB796Raid, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"your black creature", mine, 3}, {"their black creature", theirs, 3}, {"a green creature", green, 2}} {
		if p, _ := effectstest.PtOf(t, g, tc.id); p != tc.want {
			t.Errorf("%s has power %d, want %d", tc.name, p, tc.want)
		}
	}
}

func TestOldB796CivicGuildmageToughensAndTucks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	m := effectstest.BlbB1Push(g, me.ID, "Civic Guildmage", "Creature — Human Wizard", oldB796Civic, 1, 1)
	bear := oldB796Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, []string{"G"})
	effectstest.FloatMana(t, g, me, "{G}")
	effectstest.B16Activate(t, g, me.ID, m, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB796Card(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 2/3", p, tough)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == m {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	effectstest.FloatMana(t, g, me, "{U}")
	effectstest.B16Activate(t, g, me.ID, m, 1, game.ActivateAbilityParams{Strict: true, Targets: oldB796Card(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if top, err := me.Library.Top(); g.Battlefield.Contains(bear) || err != nil || top.InstanceID != bear {
		t.Error("the Bear is not on top of its owner's library")
	}
}

func TestOldB796PyricSalamanderPumpsAndIsSacrificedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	s := effectstest.BlbB1Push(g, me.ID, "Pyric Salamander", "Creature — Salamander", oldB796Salamand, 1, 1)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	effectstest.B16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, s); p != 3 {
		t.Errorf("the Salamander has power %d, want 3 after two activations", p)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(s) {
		t.Error("the Salamander survived the end step")
	}
}

func TestOldB796RampageCreaturesGrowForEachBlockerBeyondTheFirst(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, blockers, want  int
	}{
		{"Teeka's Dragon", "Artifact Creature — Dragon", oldB796Dragon, 5, 2, 9},
		{"Horrible Hordes", "Artifact Creature — Spirit", oldB796Hordes, 2, 3, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			atk := effectstest.BlbB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.power)
			var bl []uuid.UUID
			for i := 0; i < tc.blockers; i++ {
				bl = append(bl, oldB796Creature(g, opp.ID, "Wall", "Creature — Wall", "", 0, 20, []string{"W"}, "reach"))
			}
			oldB796Combat(t, g, atk, bl...)
			if p, _ := effectstest.PtOf(t, g, atk); p != tc.want {
				t.Errorf("%s has power %d, want %d", tc.name, p, tc.want)
			}
		})
	}
}
