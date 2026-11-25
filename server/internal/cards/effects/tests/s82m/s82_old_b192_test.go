package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b192_test.go — Guilds of Ravnica, older-sets batch 192 (#5525).
// Every test plays the real catalog entry through the engine.

const (
	oldB192GlaiveOfTheGuildpact = "178a1b0b-30af-4afd-8976-4a0e2133007a"
	oldB192HaazdaMarshal        = "3a1efa7e-21d5-47a1-bcd8-7d4616be0d58"
	oldB192PistonFistCyclops    = "790e3c9e-7d94-43f7-bc14-d83c7d4c889e"
	oldB192MnemonicBetrayal     = "aa41e0e6-d04e-48d6-8450-02b972dd0bb1"
	oldB192GuildmagesForum      = "ace6403d-9fac-4d0f-a6ea-eb2ff3da259d"
	oldB192DrownedSecrets       = "cab305d9-27fb-4a41-adf1-892e15167cb9"
)

func TestOldB192EveryCardIsRegisteredFull(t *testing.T) {
	effectstest.NewCatalogGame(t)
	for _, id := range []string{oldB192GlaiveOfTheGuildpact, oldB192HaazdaMarshal, oldB192PistonFistCyclops,
		oldB192MnemonicBetrayal, oldB192GuildmagesForum, oldB192DrownedSecrets} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB192GlaivePumpsPerGateYouControlAndGrantsVigilanceAndMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glaive := effectstest.PushCatalogPermanent(g, me.ID, "Glaive of the Guildpact", "Artifact — Equipment", oldB192GlaiveOfTheGuildpact, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BrAttach(t, g, glaive, bear)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("with no Gate the Bear is %d/%d, want 2/2", p, tough)
	}
	for _, kw := range []string{"vigilance", "menace"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped creature lacks %s", kw)
		}
	}
	effectstest.B12Permanent(g, me.ID, "Selesnya Guildgate", "Land — Gate")
	effectstest.B12Permanent(g, me.ID, "Boros Guildgate", "Land — Gate")
	effectstest.B12Permanent(g, opp.ID, "Opposing Gate", "Land — Gate")
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("with two Gates the Bear is %d/%d, want 4/2 (an opponent's Gate does not count)", p, tough)
	}
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	if effectstest.EffectivePower(t, g, other) != 2 || effectstest.HasEffectiveKeyword(t, g, other, "menace") {
		t.Error("an unequipped creature got the bonus")
	}
}

func TestOldB192GlaiveEquipsForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	glaive := effectstest.PushCatalogPermanent(g, me.ID, "Glaive of the Guildpact", "Artifact — Equipment", oldB192GlaiveOfTheGuildpact, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, glaive, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("equipped for less than {3}")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, glaive, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("the Glaive did not attach")
	}
}

func TestOldB192HaazdaMarshalMakesASoldierOnlyWithTwoOtherAttackers(t *testing.T) {
	for _, others := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		marshal := effectstest.B12Push(g, me.ID, "Haazda Marshal", "Creature — Human Soldier", oldB192HaazdaMarshal, 1, 1)
		attackers := []uuid.UUID{marshal}
		for i := 0; i < others; i++ {
			attackers = append(attackers, effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2))
		}
		effectstest.DeclareAttack(t, g, opp.ID, attackers...)
		effectstest.PassPriorityAroundTable(t, g)
		soldiers := effectstest.TokensNamed(g, "Soldier")
		want := 0
		if others == 2 {
			want = 1
		}
		if len(soldiers) != want {
			t.Fatalf("%d other attackers: %d Soldier tokens, want %d", others, len(soldiers), want)
		}
		if want == 1 {
			id := soldiers[0].InstanceID
			if p, tough := effectstest.PtOf(t, g, id); p != 1 || tough != 1 || !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
				t.Errorf("the token is %d/%d, want a 1/1 with lifelink", p, tough)
			}
			if soldiers[0].Controller != me.ID {
				t.Error("the token is not under your control")
			}
		}
	}
}

func TestOldB192PistonFistCyclopsAttacksOnlyAfterYouCastAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	cyclops := effectstest.B12Push(g, me.ID, "Piston-Fist Cyclops", "Creature — Cyclops", oldB192PistonFistCyclops, 4, 3)
	if !effectstest.HasEffectiveKeyword(t, g, cyclops, "defender") {
		t.Error("the Cyclops lacks defender")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(cyclops, opp.ID); err == nil {
		t.Fatal("the Cyclops attacked before an instant or sorcery was cast")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[g2.Turn.ActiveSeat], g2.Seats[(g2.Turn.ActiveSeat+1)%4]
	cyclops2 := effectstest.B12Push(g2, me2.ID, "Piston-Fist Cyclops", "Creature — Cyclops", oldB192PistonFistCyclops, 4, 3)
	effectstest.AdvanceToMain(t, g2)
	// A creature spell does not count.
	effectstest.CastFromSeat(t, g2, me2, "Grizzly Bears", "Creature — Bear", "", effectstest.CardTargets()[:0])
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.CastFromSeat(t, g2, me2, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp2.ID}})
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(cyclops2, opp2.ID); err != nil {
		t.Fatalf("the Cyclops could not attack after an instant was cast: %v", err)
	}
}

func TestOldB192MnemonicBetrayalExilesOpposingGraveyardsAndLetsYouCastFromThemWithAnyMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, third := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4], g.Seats[(g.Turn.ActiveSeat+2)%4]
	bolt := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
		OracleID: effectstest.LightningBoltOracle, Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	bear := effectstest.PushGraveyardCardTyped(opp, "Bear", "Creature — Bear")
	theirs := effectstest.PushGraveyardCardTyped(third, "Third Bear", "Creature — Bear")
	mine := effectstest.PushGraveyardCardTyped(me, "My Bear", "Creature — Bear")
	betrayal := effectstest.CastCatalogSpell(t, g, "Mnemonic Betrayal", "Sorcery", oldB192MnemonicBetrayal, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Bolt": bolt, "Bear": bear, "Third Bear": theirs} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled from its owner's graveyard", name)
		}
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("your own graveyard was exiled")
	}
	if !g.Exile.Contains(betrayal) {
		t.Error("Mnemonic Betrayal did not exile itself")
	}
	// Cast the Bolt from exile paying with blue mana.
	life := opp.Life
	effectstest.FloatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, FromZone: "exile",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast the exiled Bolt with blue mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the opponent is at %d life, want %d", opp.Life, life-3)
	}
	// At the beginning of the next end step the cards still exiled go back.
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bear) || !third.Graveyard.Contains(theirs) {
		t.Error("the cards still exiled did not return to their owners' graveyards")
	}
	if g.Exile.Contains(bear) || g.Exile.Contains(theirs) {
		t.Error("a card is still exiled")
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the resolved Bolt is not in its owner's graveyard")
	}
}

func TestOldB192MnemonicBetrayalCannotPlayALandFromThoseCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	land := effectstest.PushGraveyardCardTyped(opp, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Mnemonic Betrayal", "Sorcery", oldB192MnemonicBetrayal, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(land) {
		t.Fatal("the land was not exiled")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("a land card was cast")
	}
}

func TestOldB192GuildmagesForumAddsColourlessAndCountersAMulticoloredCreature(t *testing.T) {
	for _, tc := range []struct {
		name     string
		cost     string
		colors   []string
		forum    string
		counters int
	}{
		{"a multicolored creature", "{G}{W}", []string{"G", "W"}, "W", 1},
		{"a monocolored creature", "{G}{G}", []string{"G"}, "G", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			forum := effectstest.SeedPermanentWithOracle(g, me.ID, "Guildmages' Forum", "Land", oldB192GuildmagesForum)
			effectstest.FloatForTest(g, me, "C")
			effectstest.ActivateManaFor(t, g, me.ID, forum, 1, game.ManaAbilityParams{Colors: []string{tc.forum}})
			effectstest.FloatForTest(g, me, "G")
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gold Bear", TypeLine: "Creature — Bear", ManaCost: tc.cost,
				Colors: tc.colors, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, id, "+1/+1"); got != tc.counters {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.counters)
			}
		})
	}
}

func TestOldB192GuildmagesForumTapsForColourless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	forum := effectstest.SeedPermanentWithOracle(g, me.ID, "Guildmages' Forum", "Land", oldB192GuildmagesForum)
	effectstest.ActivateManaFor(t, g, me.ID, forum, 0, game.ManaAbilityParams{})
	if n := me.ManaPool.Total(); n != 1 {
		t.Errorf("the pool holds %d mana, want 1", n)
	}
}

func oldB192SettlePlayer(t *testing.T, g *game.Game, me *game.Player, pick uuid.UUID) {
	t.Helper()
	for j := 0; j < 12; j++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PbPlayer(pick)); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOldB192DrownedSecretsMillsTwoOnlyForYourBlueSpells(t *testing.T) {
	for _, tc := range []struct {
		name   string
		colors []string
		cost   string
		mill   int
	}{
		{"a blue spell", []string{"U"}, "{U}", 2},
		{"a red spell", []string{"R"}, "{R}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			effectstest.PushCatalogPermanent(g, me.ID, "Drowned Secrets", "Enchantment", oldB192DrownedSecrets, false)
			effectstest.AdvanceToMain(t, g)
			effectstest.SeedLibrary(opp, "A", "B", "C", "D")
			grave := opp.Graveyard.Size()
			spell := effectstest.HandCardFull(me, "Spell", "Instant", tc.cost, "", tc.colors)
			effectstest.FillPoolColored(me, tc.colors[0], 1)
			if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			oldB192SettlePlayer(t, g, me, opp.ID)
			if got := opp.Graveyard.Size() - grave; got != tc.mill {
				t.Errorf("the target milled %d cards, want %d", got, tc.mill)
			}
		})
	}
}

func TestOldB192DrownedSecretsCanTargetYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Drowned Secrets", "Enchantment", oldB192DrownedSecrets, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	grave := me.Graveyard.Size()
	spell := effectstest.HandCardFull(me, "Spell", "Instant", "{U}", "", []string{"U"})
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	oldB192SettlePlayer(t, g, me, me.ID)
	if got := me.Graveyard.Size() - grave; got != 3 { // two milled, plus the resolved Spell
		t.Errorf("the graveyard grew by %d, want 3", got)
	}
}
