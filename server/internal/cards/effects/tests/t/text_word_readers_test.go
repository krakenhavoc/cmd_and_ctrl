package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// text_word_readers_test.go — #3872: card files that used to hard-code a
// colour word or basic land type now read it through the object whose
// text it is, so Sleight of Mind and Magical Hack reach them (CR 612.2,
// ADR 0160's 2026-10-15 amendment). One test per kind of door: an effect
// reading its spell's text (ctx.TextColor), a trigger and a static
// reading their source (source.TextColor / TextSubtype), a state trigger's
// query (PermanentQuery.ThroughTextOf), a target clause reading the lent
// text (LentSubtype), and a cast trigger lending its source's text.

const (
	twrSeasingerOracle = "13745b92-1e30-4075-9733-5d00da11408c"
	twrSteamKinOracle  = "eb414a8f-566f-4635-80a1-77b2e18ac8dc"
	twrThassaOracle    = "69e5df2f-be1f-4608-a90f-3e2f51e2fea4"
)

// twrSleightTheSpell casts Sleight of Mind on a spell already on the
// stack and answers its two prompts (indices into WUBRG, then into the
// four colours left).
func twrSleightTheSpell(t *testing.T, g *game.Game, spell uuid.UUID, from, to int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Sleight of Mind", "Instant", effectstest.TwSleightOracle, effectstest.TwTargetRef(spell))
	for i := 0; i < 8 && effectstest.LatestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	effectstest.AnswerOptionPick(t, g, me.ID, from)
	effectstest.AnswerOptionPick(t, g, me.ID, to)
}

// Pyroblast's "if it's blue" is a colour word of the spell's text: made
// "if it's green", its second mode destroys a green permanent and spares
// a blue one.
func TestPyroblastChangedToGreenDestroysAGreenPermanent(t *testing.T) {
	for _, tc := range []struct {
		name      string
		colour    string
		destroyed bool
	}{
		{"a green target is destroyed", "G", true},
		{"a blue target is spared", "U", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
			victim := effectstest.TwPut(g, opp.ID, "Bear", "Creature — Bear", []string{tc.colour})
			pyro := effectstest.CastModal(t, g, "Pyroblast", "Instant", effectstest.B03PyroblastOracle, []int{1}, effectstest.TwTargetRef(victim))
			twrSleightTheSpell(t, g, pyro, effectstest.TwBlue, 3) // W, B, R, G: green is 3
			effectstest.PassPriorityAroundTable(t, g)
			if got := !twOnBattlefield(g, victim); got != tc.destroyed {
				t.Errorf("Pyroblast now reads \"if it's green\": destroyed = %v, want %v", got, tc.destroyed)
			}
		})
	}
}

// Seasinger under Magical Hack (Island → Forest): its target clause
// wants a controller with a Forest, and its state trigger asks about
// Forests, so losing the last Island no longer sacrifices it and losing
// the last Forest does.
func TestSeasingerChangedToForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	island := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	seasinger := effectstest.PushCatalogPermanent(g, me.ID, "Seasinger", "Creature — Merfolk", twrSeasingerOracle, false)
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	if err := g.ActivateCatalogAbility(me.ID, seasinger, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(victim)}); err == nil {
		t.Fatal("setup: printed Seasinger can't target a creature whose controller has only a Forest")
	}

	// Plains, Island, Swamp, Mountain, Forest; then the four others.
	effectstest.TwChange(t, g, "Magical Hack", effectstest.TwHackOracle, seasinger, 1, 3) // Island → Forest

	effectstest.RtActivate(t, g, me.ID, seasinger, 0, effectstest.LtCardTarget(victim))
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("Seasinger now steals a creature whose controller controls a Forest: controller %s", got)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(island) })
	effectstest.PassPriorityAroundTable(t, g)
	if !twOnBattlefield(g, seasinger) {
		t.Fatal("\"When you control no Forests\": losing the last Island no longer sacrifices it")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(forest) })
	effectstest.PassPriorityAroundTable(t, g)
	if twOnBattlefield(g, seasinger) {
		t.Error("with no Forest left, Seasinger is sacrificed")
	}
}

// A "whenever you cast a <colour> spell" trigger reads the colour
// through its source (YouCast lends it): Runaway Steam-Kin made to care
// about green grows on a green spell and not on a red one.
func TestRunawaySteamKinChangedToGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kin := effectstest.PushCatalogPermanent(g, me.ID, "Runaway Steam-Kin", "Creature — Elemental", twrSteamKinOracle, false)
	effectstest.TwChange(t, g, "Sleight of Mind", effectstest.TwSleightOracle, kin, effectstest.TwRed, 3) // W, U, B, G: green is 3
	counters := func() int {
		return effectstest.GaLayered(t, g, kin).Counters[game.CounterPlusOne]
	}
	twrCastColoured(t, g, "Red Bolt", "R")
	effectstest.PassPriorityAroundTable(t, g)
	if n := counters(); n != 0 {
		t.Fatalf("a red spell no longer triggers it: %d counters", n)
	}
	twrCastColoured(t, g, "Green Growth", "G")
	effectstest.PassPriorityAroundTable(t, g)
	if n := counters(); n != 1 {
		t.Errorf("a green spell triggers it: %d counters, want 1", n)
	}
}

// twrCastColoured casts a vanilla instant of one colour.
func twrCastColoured(t *testing.T, g *game.Game, name, colour string) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Colors: []string{colour},
		Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
}

// Artificial Evolution on a creature SPELL changes its type line as it
// stands on the stack, not only once it resolves.
func TestArtificialEvolutionChangesASpellsTypeLineOnTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	king := effectstest.CastCatalogSpell(t, g, "Goblin King", "Creature — Goblin", effectstest.TwGoblinKingOracle, nil)
	effectstest.CastCatalogSpell(t, g, "Artificial Evolution", "Instant", effectstest.TwEvolutionOracle, effectstest.TwTargetRef(king))
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	var onStack, elf, goblin bool
	g.ReadSnapshot(func() {
		for _, c := range g.Stack.Cards {
			if c.InstanceID == king {
				onStack, elf, goblin = true, c.HasSubtype("Elf"), c.HasSubtype("Goblin")
			}
		}
	})
	if !onStack {
		t.Fatal("setup: Goblin King is still a spell")
	}
	if !elf || goblin {
		t.Errorf("the spell on the stack is an Elf and not a Goblin: Elf %v, Goblin %v", elf, goblin)
	}
}

// A creature type in a printed protection quality is a word of the text:
// "protection from Goblins" under Artificial Evolution (Goblin → Elf) is
// "protection from Elves".
func TestArtificialEvolutionChangesAProtectionFromACreatureType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := effectstest.TwPut(g, me.ID, "Goblin Hunter", "Creature — Human Knight", []string{"W"}, "protection from Goblins")
	effectstest.CastCatalogSpell(t, g, "Artificial Evolution", "Instant", effectstest.TwEvolutionOracle, effectstest.TwTargetRef(knight))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	if effectstest.HasEffectiveKeyword(t, g, knight, "protection from Goblins") || !effectstest.HasEffectiveKeyword(t, g, knight, "protection from Elves") {
		t.Error("protection from Goblins now reads protection from Elves")
	}
}

// A Theros God's devotion is to a colour WORD of its text: Thassa made
// to count devotion to black is a creature on a board of black pips.
func TestThassaChangedToBlackCountsDevotionToBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thassa := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Thassa, God of the Sea", TypeLine: "Legendary Enchantment Creature — God",
		OracleID: twrThassaOracle, ManaCost: "{2}{U}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushCostedPermanentForTest(g, me.ID, "Black Pips", "Enchantment", "{B}{B}{B}{B}{B}")
	if effectstest.GaLayered(t, g, thassa).IsCreature() {
		t.Fatal("setup: devotion to blue is 1, so Thassa isn't a creature")
	}
	effectstest.TwChange(t, g, "Sleight of Mind", effectstest.TwSleightOracle, thassa, effectstest.TwBlue, 1) // W, B, R, G: black is 1
	if !effectstest.GaLayered(t, g, thassa).IsCreature() {
		t.Error("devotion to black is 5: Thassa is a creature")
	}
}
