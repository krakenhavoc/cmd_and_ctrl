package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b544_cards_test.go — Lorwyn clash cards, S82 OLD batch 544 (#5593).
// Every test plays the real catalog entry through the engine. A clash is
// rigged by what is on top of each library: the greater mana value wins.

const (
	oldB544Springjack = "97b0b719-d948-4671-97ba-1b2be8cc55e4"
	oldB544Cleaning   = "9b3b81cf-f020-4e1e-95c7-db44ca153dc1"
	oldB544Strangle   = "af64ce09-b7d9-4db9-905e-c4356f242ced"
	oldB544Rascal     = "bdec1d22-2f89-424b-b48f-46c9158943af"
	oldB544Hoodlums   = "c48ce36d-46c3-4432-8711-1b10f12efc90"
	oldB544Adder      = "c95e75fb-2bf2-431c-b7ca-d84236bffe5c"
	oldB544Ambush     = "e5a06062-0513-48af-b92a-420ebbeb1695"
)

func TestOldB544EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB544Springjack: "Springjack Knight", oldB544Cleaning: "Spring Cleaning", oldB544Strangle: "Weed Strangle",
		oldB544Rascal: "Paperfin Rascal", oldB544Hoodlums: "Bog Hoodlums", oldB544Adder: "Adder-Staff Boggart",
		oldB544Ambush: "Gilt-Leaf Ambush",
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

// oldB544Top puts a card of the given mana cost on top of p's library.
func oldB544Top(p *game.Player, name, cost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = cost
	p.Library.PushTop(c)
	return c.InstanceID
}

// oldB544Rig makes `me` win the clash, lose it, or tie.
func oldB544Rig(me, opp *game.Player, outcome string) {
	switch outcome {
	case "win":
		oldB544Top(me, "My Five", "{3}{R}{R}")
		oldB544Top(opp, "Their One", "{U}")
	case "lose":
		oldB544Top(me, "My One", "{R}")
		oldB544Top(opp, "Their Five", "{3}{U}{U}")
	default:
		oldB544Top(me, "My Two", "{1}{R}")
		oldB544Top(opp, "Their Two", "{1}{U}")
	}
}

// oldB544Clash answers a clash with `opponent`: the controller picks them
// and both clashers keep their card on top.
func oldB544Clash(t *testing.T, g *game.Game, opponent uuid.UUID) {
	t.Helper()
	answered := false
	for i := 0; i < 8; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionPick {
				open = c
				break
			}
		}
		if open == nil {
			break
		}
		pick := 0
		if !answered {
			for j, o := range open.PickOptions {
				if o.Player == opponent {
					pick = j
				}
			}
			answered = true
		}
		if err := g.ResolveOptionPick(open.ID, open.Chooser, pick); err != nil {
			t.Fatalf("answer the clash question: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB544ClashCreaturesGrowOnlyOnAWin(t *testing.T) {
	for _, c := range []struct{ name, typeLine, oracle string }{
		{"Paperfin Rascal", "Creature — Merfolk Rogue", oldB544Rascal},
		{"Adder-Staff Boggart", "Creature — Goblin Warrior", oldB544Adder},
		{"Bog Hoodlums", "Creature — Goblin Warrior", oldB544Hoodlums},
	} {
		for outcome, want := range map[string]int{"win": 1, "lose": 0, "tie": 0} {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			oldB544Rig(me, opp, outcome)
			id := effectstest.CastCatalogSpell(t, g, c.name, c.typeLine, c.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			oldB544Clash(t, g, opp.ID)
			if got := effectstest.Pr8Counters(g, id, game.CounterPlusOne); got != want {
				t.Errorf("after a %s %s has %d +1/+1 counters, want %d", outcome, c.name, got, want)
			}
		}
	}
}

func TestOldB544BogHoodlumsCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Bog Hoodlums", "Creature — Goblin Warrior", oldB544Hoodlums, false)
	effectstest.AssertRestrictions(t, g, id, game.CantBlock)
}

func TestOldB544WeedStrangleDestroysAndGainsLifeEqualToToughnessOnAWin(t *testing.T) {
	for outcome, gain := range map[string]int{"win": 4, "lose": 0} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
		oldB544Rig(me, opp, outcome)
		life := me.Life
		effectstest.CastCatalogSpell(t, g, "Weed Strangle", "Sorcery", oldB544Strangle, effectstest.CardRefs(wall))
		effectstest.PassPriorityAroundTable(t, g)
		oldB544Clash(t, g, opp.ID)
		if g.Battlefield.Contains(wall) {
			t.Fatalf("after a %s the creature was not destroyed", outcome)
		}
		if me.Life != life+gain {
			t.Errorf("after a %s life %d -> %d, want +%d", outcome, life, me.Life, gain)
		}
	}
}

func TestOldB544SpringCleaningDestroysTheTargetAndEveryOpposingEnchantmentOnAWin(t *testing.T) {
	for outcome, sweepGone := range map[string]bool{"win": true, "lose": false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		target := effectstest.B12Permanent(g, opp.ID, "Aura A", "Enchantment")
		other := effectstest.B12Permanent(g, opp.ID, "Aura B", "Enchantment")
		mine := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment")
		oldB544Rig(me, opp, outcome)
		effectstest.CastCatalogSpell(t, g, "Spring Cleaning", "Instant", oldB544Cleaning, effectstest.CardRefs(target))
		effectstest.PassPriorityAroundTable(t, g)
		oldB544Clash(t, g, opp.ID)
		if g.Battlefield.Contains(target) {
			t.Errorf("after a %s the targeted enchantment survived", outcome)
		}
		if g.Battlefield.Contains(other) == sweepGone {
			t.Errorf("after a %s the other opposing enchantment survived = %v, want %v", outcome, g.Battlefield.Contains(other), !sweepGone)
		}
		if !g.Battlefield.Contains(mine) {
			t.Errorf("after a %s the caster's own enchantment was destroyed", outcome)
		}
	}
}

func TestOldB544SpringjackKnightGrantsDoubleStrikeOnAWin(t *testing.T) {
	for outcome, want := range map[string]bool{"win": true, "lose": false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		knight := effectstest.PushCatalogPermanent(g, me.ID, "Springjack Knight", "Creature — Kithkin Knight", oldB544Springjack, false)
		effectstest.Unsick(g, knight)
		ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
		oldB544Rig(me, opp, outcome)
		effectstest.Pr7bAttack(t, g, opp.ID, knight)
		if pc := effectstest.FirstPendingTarget(g); pc != nil {
			if err := g.ResolvePickTarget(pc.ID, pc.Chooser, game.TargetRef{Kind: game.TargetCard, ID: ally}); err != nil {
				t.Fatalf("choose the target: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		oldB544Clash(t, g, opp.ID)
		if got := effectstest.P1g2HasKeyword(g, ally, "double strike"); got != want {
			t.Errorf("after a %s the target has double strike = %v, want %v", outcome, got, want)
		}
	}
}

func TestOldB544GiltLeafAmbushMakesTwoElvesAndGivesThemDeathtouchOnAWin(t *testing.T) {
	for outcome, want := range map[string]bool{"win": true, "lose": false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		oldB544Rig(me, opp, outcome)
		effectstest.CastCatalogSpell(t, g, "Gilt-Leaf Ambush", "Kindred Instant — Elf", oldB544Ambush, nil)
		effectstest.PassPriorityAroundTable(t, g)
		oldB544Clash(t, g, opp.ID)
		var elves []uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Elf Warrior" {
				elves = append(elves, c.InstanceID)
			}
		}
		if len(elves) != 2 {
			t.Fatalf("after a %s there are %d Elf Warriors, want 2", outcome, len(elves))
		}
		for _, id := range elves {
			if got := effectstest.P1g2HasKeyword(g, id, "deathtouch"); got != want {
				t.Errorf("after a %s an Elf has deathtouch = %v, want %v", outcome, got, want)
			}
		}
	}
}
