package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b508_test.go — Shards of Alara, batch 508 (#5588). Every test plays
// the real catalog entry through the engine.

const (
	oldB508Plowbeast  = "07237694-b5d5-4c92-a429-55e343b8dd1c"
	oldB508Scream     = "287d5a5d-e477-4d5b-815a-d83c73f4792d"
	oldB508Submersion = "37985fef-696e-4c6b-9a60-29c5051b5a25"
	oldB508Roar       = "40ad5a52-37ad-4524-be06-9e263a907d31"
	oldB508Rannet     = "562aaf81-e70a-401d-a638-bb7ce4f40626"
	oldB508Sphinx     = "6452873b-1bb2-44b3-b630-88f15962b6e8"
	oldB508Silence    = "6eefa487-1542-4d24-80bd-4893d9d908c8"
	oldB508Wave       = "7e607fc7-ec22-4249-8936-70a026f9996d"
	oldB508Dragon     = "85a38b39-5806-4df9-95cc-8d608e2b5e47"
	oldB508Bear       = "Creature — Bear"
)

func TestOldB508EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB508Plowbeast: "Yoked Plowbeast", oldB508Scream: "Resounding Scream", oldB508Submersion: "Volcanic Submersion",
		oldB508Roar: "Resounding Roar", oldB508Rannet: "Ridge Rannet", oldB508Sphinx: "Sphinx's Herald",
		oldB508Silence: "Resounding Silence", oldB508Wave: "Resounding Wave", oldB508Dragon: "Dragon's Herald",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
		cycles := false
		for _, a := range spec.Activated {
			cycles = cycles || a.Cycling
		}
		if cycles != (name != "Sphinx's Herald" && name != "Dragon's Herald") {
			t.Errorf("%s: cycling = %v", name, cycles)
		}
	}
}

func TestOldB508VanillaBeastsCycleForTwo(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Yoked Plowbeast", "Creature — Beast", oldB508Plowbeast},
		{"Ridge Rannet", "Creature — Beast", oldB508Rannet},
		{"Volcanic Submersion", "Sorcery", oldB508Submersion},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedLibrary(me, "A", "B")
			hand := len(me.Hand.Cards)
			id, _ := effectstest.CycleFromHand(t, g, tc.name, tc.typeLine, tc.oracle, "{C}{C}")
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Graveyard.Contains(id) {
				t.Error("the cycled card is not in the graveyard")
			}
			if got := len(me.Hand.Cards); got != hand+1 {
				t.Errorf("hand %d -> %d, want one card drawn", hand, got)
			}
		})
	}
}

func TestOldB508VolcanicSubmersionDestroysAnArtifactOrLandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB508Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Volcanic Submersion", "Sorcery", oldB508Submersion, effectstest.CardTargets(bear)); err == nil {
		t.Fatal("targeted a creature")
	}
	land := effectstest.SeedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	for _, id := range []uuid.UUID{land, rock} {
		effectstest.CastCatalogSpell(t, g, "Volcanic Submersion", "Sorcery", oldB508Submersion, effectstest.CardTargets(id))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(id) {
			t.Error("the target was not destroyed")
		}
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was destroyed")
	}
}

func TestOldB508ResoundingScreamDiscardsAtRandomAndCyclesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	for i := 0; i < 4; i++ {
		effectstest.PushHandCardOfType(opp, "Card", "Instant")
	}
	hand := len(opp.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Resounding Scream", "Sorcery", oldB508Scream, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(opp.Hand.Cards); got != hand-1 {
		t.Errorf("hand %d, want %d after the spell", got, hand-1)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	for i := 0; i < 4; i++ {
		effectstest.PushHandCardOfType(opp, "Card", "Instant")
	}
	effectstest.SeedLibrary(me, "A", "B")
	hand = len(opp.Hand.Cards)
	effectstest.CycleFromHand(t, g, "Resounding Scream", "Sorcery", oldB508Scream, "{C}{C}{C}{C}{C}{U}{B}{R}")
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(opp.Hand.Cards); got != hand-2 {
		t.Errorf("hand %d, want %d after the cycle trigger", got, hand-2)
	}
}

func TestOldB508ResoundingRoarPumpsThreeAndCyclesForSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB508Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Resounding Roar", "Instant", oldB508Roar, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 5 || tt != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tt)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	bear = effectstest.B12Creature(g, me.ID, "Bear", oldB508Bear, 2, 2)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.CycleFromHand(t, g, "Resounding Roar", "Instant", oldB508Roar, "{C}{C}{C}{C}{C}{R}{G}{W}")
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 8 || tt != 8 {
		t.Errorf("the Bear is %d/%d, want 8/8", p, tt)
	}
}

func TestOldB508ResoundingSilenceExilesAttackersOnlyAndCyclesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	idle := effectstest.B12Creature(g, me.ID, "Idle", oldB508Bear, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB508Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	spell := effectstest.HandCardFull(me, "Resounding Silence", "Instant", "{3}{W}", oldB508Silence, []string{"W"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(idle)}); err == nil {
		t.Fatal("targeted a creature that is not attacking")
	}
	effectstest.CastFromSeat(t, g, me, "Resounding Silence", "Instant", oldB508Silence, effectstest.CardTargets(attacker))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, attacker) != game.ZoneExile {
		t.Errorf("the attacker is in zone %v, want exile", effectstest.ZoneOfCard(g, attacker))
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	a1 := effectstest.B12Creature(g, me.ID, "A1", oldB508Bear, 2, 2)
	a2 := effectstest.B12Creature(g, me.ID, "A2", oldB508Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a1, a2)
	effectstest.SeedLibrary(me, "X", "Y")
	id := effectstest.PushCatalogHandCard(me, "Resounding Silence", "Instant", oldB508Silence)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}{G}{W}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the cycle trigger asked for no targets")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardTargets(a1, a2)); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, a := range []uuid.UUID{a1, a2} {
		if effectstest.ZoneOfCard(g, a) != game.ZoneExile {
			t.Errorf("an attacker is in zone %v, want exile", effectstest.ZoneOfCard(g, a))
		}
	}
}

func TestOldB508ResoundingWaveBouncesOneAndCyclesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB508Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Resounding Wave", "Instant", oldB508Wave, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !effectstest.HandHas(opp, bear) {
		t.Error("the Bear was not returned to its owner's hand")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	b1 := effectstest.B12Creature(g, opp.ID, "B1", oldB508Bear, 2, 2)
	b2 := effectstest.B12Creature(g, opp.ID, "B2", oldB508Bear, 2, 2)
	b3 := effectstest.B12Creature(g, opp.ID, "B3", oldB508Bear, 2, 2)
	effectstest.SeedLibrary(me, "X", "Y")
	effectstest.CycleFromHand(t, g, "Resounding Wave", "Instant", oldB508Wave, "{C}{C}{C}{C}{C}{W}{U}{B}")
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the cycle trigger asked for no targets")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardTargets(b1, b2)); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{b1, b2} {
		if g.Battlefield.Contains(id) || !effectstest.HandHas(opp, id) {
			t.Error("a targeted permanent was not returned to its owner's hand")
		}
	}
	if !g.Battlefield.Contains(b3) {
		t.Error("the third permanent was returned")
	}
}

func oldB508Herald(t *testing.T, name, typeLine, oracle string, colors []string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		OracleID: oracle, Colors: colors, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	return g, me, id
}

func oldB508Fodder(g *game.Game, me *game.Player, name string, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: oldB508Bear,
		Colors: colors, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
}

func TestOldB508SphinxsHeraldSacrificesThreeColoursAndFetchesTheSovereign(t *testing.T) {
	g, me, herald := oldB508Herald(t, "Sphinx's Herald", "Artifact Creature — Vedalken Wizard", oldB508Sphinx, []string{"U"})
	white := oldB508Fodder(g, me, "White", "W")
	black := oldB508Fodder(g, me, "Black", "B")
	extra := oldB508Fodder(g, me, "Extra", "W")
	sovereign := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Sphinx Sovereign", TypeLine: "Artifact Creature — Sphinx", Power: 6, Toughness: 6},
		game.Card{Name: "Bear", TypeLine: oldB508Bear})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCU")
	// One of each colour: the Herald is the blue creature.
	if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{white, extra, black}}); err == nil {
		t.Fatal("two white creatures and a black one paid for a white, a blue and a black")
	}
	effectstest.FloatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{white, herald, black}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Sphinx Sovereign")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sovereign[0]) {
		t.Error("the Sphinx Sovereign is not on the battlefield")
	}
	for _, id := range []uuid.UUID{white, herald, black} {
		if g.Battlefield.Contains(id) {
			t.Error("a sacrificed creature is still on the battlefield")
		}
	}
	if !g.Battlefield.Contains(extra) {
		t.Error("the extra creature was sacrificed")
	}
}

func TestOldB508DragonsHeraldNeedsABlackARedAndAGreenCreature(t *testing.T) {
	g, me, herald := oldB508Herald(t, "Dragon's Herald", "Creature — Goblin Shaman", oldB508Dragon, []string{"R"})
	black := oldB508Fodder(g, me, "Black", "B")
	green := oldB508Fodder(g, me, "Green", "G")
	black2 := oldB508Fodder(g, me, "Second Black", "B")
	overlord := effectstest.SeedSearchLibrary(me, game.Card{Name: "Hellkite Overlord", TypeLine: "Creature — Dragon", Power: 8, Toughness: 8})
	effectstest.AdvanceToMain(t, g)
	// Two black creatures and the red Herald leave the green unpaid.
	effectstest.FloatForTest(g, me, "CCR")
	if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{black2, herald, black}}); err == nil {
		t.Fatal("two black creatures and a red one paid for black, red and green")
	}
	effectstest.FloatForTest(g, me, "CCR")
	if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{black, herald, green}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Hellkite Overlord")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(overlord[0]) {
		t.Error("the Hellkite Overlord is not on the battlefield")
	}
}
