package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// answers_signal_test.go — ADR 0142's signal PR, through the enumerator.
//
// Owner answer 2: a first strike, double strike or deathtouch grant is a
// combat answer, so Endling's deathtouch sets combat_interacts and not
// interacts. Owner answer 4: "Sacrifice this creature: …" interacts
// while an opponent's item on the stack targets the creature, and
// combat-interacts while it attacks or blocks. A crew row is a combat
// answer (animate).

const (
	oracleSakuraTribeElderAS = "e3afc704-220f-498f-9eaa-0821b17dc24c"
	oracleEndlingAS          = "4da3e510-e495-4077-997f-8efc04f01892"
	oracleSmugglersCopterAS  = "49136bdc-bc50-49a2-999a-1ef9c16ea130"
)

func opponentOf(g *game.Game, seat *game.Player) *game.Player {
	for _, s := range g.Seats {
		if s.ID != seat.ID {
			return s
		}
	}
	return nil
}

// targetedBy puts a spell controlled by `caster` on the stack, aimed at
// `victim`.
func targetedBy(g *game.Game, caster *game.Player, victim uuid.UUID) {
	spell := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{
			InstanceID: spell, Name: "Doom Blade", TypeLine: "Instant",
			Owner: caster.ID, Controller: caster.ID,
		})
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		g.StackMeta[spell] = &game.StackItem{
			ID: spell, Kind: game.StackItemSpell,
			Controller: caster.ID, Owner: caster.ID, SourceCardID: spell,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		}
	})
}

func elderFlags(t *testing.T, g *game.Game, seat *game.Player, elder uuid.UUID) (interacts, combat bool) {
	t.Helper()
	got := movesFrom(legal.EnumerateFor(g, seat.ID), elder, legal.KindActivate)
	if len(got) == 0 {
		t.Fatal("Sakura-Tribe Elder offered no activation")
	}
	for _, m := range got {
		interacts = interacts || m.Interacts
		combat = combat || m.CombatInteracts
	}
	return interacts, combat
}

func TestSelfSacrificeInteractsWhenItsCreatureIsThreatened(t *testing.T) {
	newElder := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
		g := newTable(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		clearHand(seat)
		advanceTo(t, g, game.StepPrecombatMain)
		elder := battlefieldCard(g, seat, game.Card{
			Name: "Sakura-Tribe Elder", TypeLine: "Creature — Snake Shaman", ManaCost: "{1}{G}",
			OracleID: oracleSakuraTribeElderAS, Power: 1, Toughness: 1,
		})
		return g, seat, elder
	}

	t.Run("idle", func(t *testing.T) {
		g, seat, elder := newElder(t)
		if i, c := elderFlags(t, g, seat, elder); i || c {
			t.Errorf("an unthreatened Elder: interacts=%v combat_interacts=%v, want neither", i, c)
		}
	})
	t.Run("an opponent's spell targets it", func(t *testing.T) {
		g, seat, elder := newElder(t)
		targetedBy(g, opponentOf(g, seat), elder)
		if i, c := elderFlags(t, g, seat, elder); !i || c {
			t.Errorf("a targeted Elder: interacts=%v combat_interacts=%v, want interacts only", i, c)
		}
	})
	t.Run("its own controller's spell targets it", func(t *testing.T) {
		g, seat, elder := newElder(t)
		targetedBy(g, seat, elder)
		if i, _ := elderFlags(t, g, seat, elder); i {
			t.Error("an Elder its own controller targets interacts")
		}
	})
	t.Run("another creature is targeted", func(t *testing.T) {
		g, seat, elder := newElder(t)
		other := battlefieldCard(g, seat, game.Card{
			Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2,
		})
		targetedBy(g, opponentOf(g, seat), other)
		if i, _ := elderFlags(t, g, seat, elder); i {
			t.Error("the Elder interacts when only another creature is targeted")
		}
	})
	t.Run("attacking", func(t *testing.T) {
		g, seat, elder := newElder(t)
		opp := opponentOf(g, seat)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == elder {
					g.Battlefield.Cards[i].AttackingTarget = opp.ID
				}
			}
		})
		if i, c := elderFlags(t, g, seat, elder); i || !c {
			t.Errorf("an attacking Elder: interacts=%v combat_interacts=%v, want combat_interacts only", i, c)
		}
	})
}

func TestCombatTierAnswersSetCombatInteracts(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	for range 3 {
		battlefieldCard(g, seat, basic("Swamp", "Swamp"))
	}
	battlefieldCard(g, seat, game.Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2,
	})
	endling := battlefieldCard(g, seat, game.Card{
		Name: "Endling", TypeLine: "Creature — Zombie Shapeshifter", ManaCost: "{2}{B}{B}",
		OracleID: oracleEndlingAS, Power: 3, Toughness: 3,
	})
	copter := battlefieldCard(g, seat, game.Card{
		Name: "Smuggler's Copter", TypeLine: "Artifact — Vehicle", ManaCost: "{2}",
		OracleID: oracleSmugglersCopterAS, Power: 3, Toughness: 3,
	})

	moves := legal.EnumerateFor(g, seat.ID)
	dispatchAll(t, g, seat.ID, moves)

	sawDeathtouch, sawUndying := false, false
	for _, m := range movesFrom(moves, endling, legal.KindActivate) {
		switch {
		case strings.Contains(m.Label, "deathtouch"):
			sawDeathtouch = true
			if m.Interacts || !m.CombatInteracts {
				t.Errorf("Endling's deathtouch: interacts=%v combat_interacts=%v, want combat only (%s)", m.Interacts, m.CombatInteracts, m.Label)
			}
		case strings.Contains(m.Label, "undying"):
			sawUndying = true
			if !m.Interacts || m.CombatInteracts {
				t.Errorf("Endling's undying: interacts=%v combat_interacts=%v, want interacts only (%s)", m.Interacts, m.CombatInteracts, m.Label)
			}
		}
	}
	if !sawDeathtouch || !sawUndying {
		t.Fatalf("Endling's deathtouch or undying row not offered: %v", labels(moves))
	}
	crew := movesFrom(moves, copter, legal.KindActivate)
	if len(crew) == 0 {
		t.Fatalf("Smuggler's Copter offered no crew: %v", labels(moves))
	}
	for _, m := range crew {
		if m.Interacts || !m.CombatInteracts {
			t.Errorf("crew: interacts=%v combat_interacts=%v, want combat only (%s)", m.Interacts, m.CombatInteracts, m.Label)
		}
	}
	for _, m := range moves {
		if m.CombatInteracts && (m.Interacts || m.HasTargets) {
			t.Errorf("%q: combat_interacts set beside interacts or has_targets", m.Label)
		}
	}
}
