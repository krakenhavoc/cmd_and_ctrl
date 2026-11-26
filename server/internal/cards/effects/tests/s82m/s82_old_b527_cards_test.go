package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b527_cards_test.go — Eventide, S82 older sets batch 527 (#5590).

const (
	eve3Fable   = "24fe8bc3-7e3a-435b-a232-7f80787aba1f"
	eve3Rise    = "dfd3adec-033b-47ed-a227-ea1f48b4057c"
	eve3Talara  = "08c7a6f5-7e25-4f8d-9e4a-aec7248e45fa"
	eve3Selkie  = "1f1c369a-3209-405b-9880-0e10b86e5698"
	eve3Zealot  = "7445122d-a943-493c-b7b6-3699cc07db0e"
	eve3Leering = "a5cfd1e4-cd14-454a-9734-b5f6952d6a50"
)

func TestS82OldB527EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eve3Fable: "Fable of Wolf and Owl", eve3Rise: "Rise of the Hobgoblins", eve3Talara: "Talara's Battalion",
		eve3Selkie: "Selkie Hedge-Mage", eve3Zealot: "Kithkin Zealot", eve3Leering: "Leering Emblem",
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

// eve3Cast casts a card from my hand for real, paying `pool`, and settles the stack.
func eve3Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle, cost string, colors []string, pool string) uuid.UUID {
	t.Helper()
	id := eve3Hand(g, me, name, typeLine, oracle, cost, colors)
	if pool != "" {
		effectstest.AddPool(t, g, me.ID, pool)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func eve3Hand(g *game.Game, me *game.Player, name, typeLine, oracle, cost string, colors []string) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Colors: colors, Owner: me.ID, Controller: me.ID})
	return id
}

func eve3Count(g *game.Game, owner uuid.UUID, name string) int {
	n, _ := effectstest.CountTokensNamed(g, owner, name)
	return n
}

func TestS82OldB527FableMakesTheTokenOfEachColourOfSpellCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Fable of Wolf and Owl", "Enchantment", eve3Fable, 0, 0)
	effectstest.AdvanceToMain(t, g)
	// A green-blue spell triggers both abilities; each is a "may".
	eve3Cast(t, g, me, "Gold Spell", "Instant", "", "", []string{"G", "U"}, "")
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if w, b := eve3Count(g, me.ID, "Wolf"), eve3Count(g, me.ID, "Bird"); w != 1 || b != 0 {
		t.Errorf("Wolves %d, Birds %d; want 1, 0 (the Bird was declined)", w, b)
	}
	// A red spell triggers neither.
	eve3Cast(t, g, me, "Red Spell", "Instant", "", "", []string{"R"}, "")
	effectstest.PassPriorityAroundTable(t, g)
	if w, b := eve3Count(g, me.ID, "Wolf"), eve3Count(g, me.ID, "Bird"); w != 1 || b != 0 {
		t.Errorf("a red spell changed the tokens: Wolves %d, Birds %d", w, b)
	}
}

func TestS82OldB527RiseOfTheHobgoblinsPaysXAndGrantsFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	red := effectstest.B16Creature(g, me.ID, "Red", "Creature — Goblin", 2, 2, "R")
	green := effectstest.B16Creature(g, me.ID, "Green", "Creature — Elf", 2, 2, "G")
	rise := eve3Cast(t, g, me, "Rise of the Hobgoblins", "Enchantment", eve3Rise, "{R/W}{R/W}", []string{"R", "W"}, "{R}{W}")
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 3)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := eve3Count(g, me.ID, "Goblin Soldier"); n != 3 {
		t.Fatalf("%d Goblin Soldier tokens, want 3", n)
	}
	effectstest.AddPool(t, g, me.ID, "{W}")
	effectstest.B16Activate(t, g, me.ID, rise, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, red, "first strike")
	if got := effectstest.EffectiveAbilities(t, g, green); effectstest.ContainsString(got, "first strike") {
		t.Errorf("a green creature gained first strike: %v", got)
	}
}

func TestS82OldB527TalarasBattalionNeedsAnotherGreenSpellFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := eve3Hand(g, me, "Talara's Battalion", "Creature — Elf Warrior", eve3Talara, "{1}{G}", []string{"G"})
	effectstest.AddPool(t, g, me.ID, "{C}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no other green spell this turn")
	}
	eve3Cast(t, g, me, "Red Spell", "Instant", "", "", []string{"R"}, "")
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast after only a red spell")
	}
	eve3Cast(t, g, me, "Green Spell", "Instant", "", "", []string{"G"}, "")
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast after a green spell: %v", err)
	}
}

func eve3Land(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: subtype, TypeLine: "Basic Land — " + subtype, Owner: owner, Controller: owner})
}

func TestS82OldB527SelkieHedgeMageChecksForestsAndIslands(t *testing.T) {
	// Two Forests, one Island: only the life gain.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	eve3Land(g, me.ID, "Forest")
	eve3Land(g, me.ID, "Forest")
	eve3Land(g, me.ID, "Island")
	tapped := effectstest.B16Creature(g, g.Seats[1].ID, "Tapped", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, tapped)
	life := me.Life
	eve3Cast(t, g, me, "Selkie Hedge-Mage", "Creature — Merfolk Wizard", eve3Selkie, "{2}{G/U}", []string{"G", "U"}, "{C}{C}{G}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	if c, ok := effectstest.BattlefieldCard(g, tapped); !ok || c.Tapped != true {
		t.Error("with one Island the bounce trigger should not have happened")
	}
	// Two Islands: the bounce, not the life.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g2)
	eve3Land(g2, me2.ID, "Island")
	eve3Land(g2, me2.ID, "Island")
	tapped2 := effectstest.B16Creature(g2, g2.Seats[1].ID, "Tapped", "Creature — Bear", 2, 2)
	untapped := effectstest.B16Creature(g2, g2.Seats[1].ID, "Untapped", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g2, tapped2)
	life = me2.Life
	eve3Cast(t, g2, me2, "Selkie Hedge-Mage", "Creature — Merfolk Wizard", eve3Selkie, "{2}{G/U}", []string{"G", "U"}, "{C}{C}{U}")
	effectstest.PassPriorityAroundTable(t, g2)
	if p := effectstest.LatestPickTarget(g2, me2.ID); p != nil {
		effectstest.B17PickCards(t, g2, me2.ID, tapped2)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g2, me2.ID, true)
	effectstest.PassPriorityAroundTable(t, g2)
	if _, ok := effectstest.BattlefieldCard(g2, tapped2); ok {
		t.Error("the tapped creature was not returned to hand")
	}
	if _, ok := effectstest.BattlefieldCard(g2, untapped); !ok {
		t.Error("an untapped creature was bounced")
	}
	if me2.Life != life {
		t.Errorf("life changed with no Forests: %d → %d", life, me2.Life)
	}
}

func TestS82OldB527KithkinZealotGainsPerBlackOrRedPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Creature(g, opp.ID, "Black", "Creature — Zombie", 1, 1, "B")
	effectstest.B16Creature(g, opp.ID, "Red", "Creature — Goblin", 1, 1, "R")
	effectstest.B16Creature(g, opp.ID, "Both", "Creature — Ogre", 1, 1, "B", "R")
	effectstest.B16Creature(g, opp.ID, "White", "Creature — Soldier", 1, 1, "W")
	effectstest.B16Creature(g, me.ID, "MyBlack", "Creature — Zombie", 1, 1, "B")
	life := me.Life
	eve3Cast(t, g, me, "Kithkin Zealot", "Creature — Kithkin Cleric", eve3Zealot, "{1}{W}", []string{"W"}, "{C}{W}")
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3 (black, red and the both-coloured one; not white or mine)", life, me.Life)
	}
}

func TestS82OldB527LeeringEmblemPumpsTheEquippedCreatureOnEachCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	emblem := effectstest.B12Push(g, me.ID, "Leering Emblem", "Artifact — Equipment", eve3Leering, 0, 0)
	effectstest.AdvanceToMain(t, g)
	// Unequipped: a cast does nothing.
	eve3Cast(t, g, me, "Spell", "Instant", "", "", []string{"R"}, "")
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Fatalf("unequipped cast changed power to %d", p)
	}
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, emblem, bear)
	eve3Cast(t, g, me, "Spell", "Instant", "", "", []string{"R"}, "")
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped Bear is %d/%d after a cast, want 4/4", p, tough)
	}
	eve3Cast(t, g, me, "Spell", "Instant", "", "", []string{"R"}, "")
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 6 {
		t.Errorf("a second cast: power %d, want 6", p)
	}
}
