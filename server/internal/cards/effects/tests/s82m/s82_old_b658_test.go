package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b658_test.go — Invasion, batch 658 (#5616): removal.

const (
	oldB658ShivanHarvest   = "27c3d7cd-f92e-4203-9fb5-f6b4776f6ffd"
	oldB658BendOrBreak     = "4362efe4-339d-40e1-bd72-3cc9e41bf93f"
	oldB658TsabosAssassin  = "4d00c44b-8953-4b18-af73-2944321b6cd1"
	oldB658TsabosDecree    = "4e0de4d9-7a30-4a2b-94ba-8239baad2a0a"
	oldB658CapashenUnicorn = "52793400-bc83-402a-9609-928a4d3ac812"
	oldB658SpreadingPlague = "6c134ef7-5c2f-4e4c-ae40-ef338fbf41e6"
	oldB658TeferisCare     = "6c5c7b26-9cd6-4f29-bbb2-a0745153ee5d"
	oldB658TrenchWurm      = "9cf65178-6408-46da-a940-f5cb0960b61f"
	oldB658Wallop          = "e939ef5e-ebe1-48cd-9e9f-eb3fd516165c"
	oldB658Liberate        = "f17e6c85-4813-4b0f-a0e9-d295822e9bb5"
	oldB658Tranquility     = "f671e3c3-cd59-4d06-a1af-5d04892cf74d"
)

func TestOldB658EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB658ShivanHarvest: "Shivan Harvest", oldB658BendOrBreak: "Bend or Break", oldB658TsabosAssassin: "Tsabo's Assassin",
		oldB658TsabosDecree: "Tsabo's Decree", oldB658CapashenUnicorn: "Capashen Unicorn", oldB658SpreadingPlague: "Spreading Plague",
		oldB658TeferisCare: "Teferi's Care", oldB658TrenchWurm: "Trench Wurm", oldB658Wallop: "Wallop",
		oldB658Liberate: "Liberate", oldB658Tranquility: "Tranquility",
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

func oldB658Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB658Colored(g *game.Game, owner uuid.UUID, name, typeLine string, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Power: 2, Toughness: 2,
		Colors: colors, Owner: owner, Controller: owner,
	})
}

func oldB658Wake(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).SummonedThisTurn = false })
}

func oldB658Activate(g *game.Game, who, src uuid.UUID, idx int, p game.ActivateAbilityParams) error {
	p.Strict = true
	return g.ActivateCatalogAbility(who, src, idx, p)
}

func TestOldB658ShivanHarvestSacrificesACreatureToDestroyANonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB658Seats(g)
	src := effectstest.PushCatalogPermanent(g, me.ID, "Shivan Harvest", "Enchantment", oldB658ShivanHarvest, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	basic := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	nonbasic := effectstest.B12Permanent(g, opp.ID, "Wooded Hollow", "Land — Forest")
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := oldB658Activate(g, me.ID, src, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(basic), SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a basic land was a legal target")
	}
	if err := oldB658Activate(g, me.ID, src, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(nonbasic)}); err == nil {
		t.Fatal("activated without sacrificing a creature")
	}
	if err := oldB658Activate(g, me.ID, src, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(nonbasic), SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.OnBattlefieldByID(g, bear) {
		t.Error("the creature was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, nonbasic) || !effectstest.OnBattlefieldByID(g, basic) {
		t.Error("the nonbasic land should be destroyed and the basic one left")
	}
}

func TestOldB658TrenchWurmNeedsTapAndManaAndOnlyHitsNonbasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB658Seats(g)
	wurm := effectstest.PushCatalogPermanent(g, me.ID, "Trench Wurm", "Creature — Wurm", oldB658TrenchWurm, true)
	basic := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	nonbasic := effectstest.B12Permanent(g, opp.ID, "Wooded Hollow", "Land — Forest")
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	if err := oldB658Activate(g, me.ID, wurm, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(nonbasic)}); err == nil {
		t.Fatal("a summoning-sick Wurm paid its {T} cost")
	}
	oldB658Wake(g, wurm)
	if err := oldB658Activate(g, me.ID, wurm, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(basic)}); err == nil {
		t.Fatal("a basic land was a legal target")
	}
	if err := oldB658Activate(g, me.ID, wurm, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(nonbasic)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, nonbasic) || !effectstest.OnBattlefieldByID(g, wurm) {
		t.Error("the nonbasic land should be destroyed and the Wurm stay")
	}
}

func TestOldB658CapashenUnicornSacrificesItselfToDestroyAnArtifactOrEnchantment(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		ok             bool
	}{{"Rock", "Artifact", true}, {"Pacifism", "Enchantment", true}, {"Bear", "Creature — Bear", false}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := oldB658Seats(g)
		unicorn := effectstest.PushCatalogPermanent(g, me.ID, "Capashen Unicorn", "Creature — Unicorn", oldB658CapashenUnicorn, false)
		victim := effectstest.B12Permanent(g, opp.ID, tc.name, tc.typeLine)
		effectstest.FloatMana(t, g, me, "{W}{W}")
		err := oldB658Activate(g, me.ID, unicorn, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim)})
		if (err == nil) != tc.ok {
			t.Errorf("%s: activation error %v, want legal=%v", tc.name, err, tc.ok)
			continue
		}
		if !tc.ok {
			continue
		}
		if effectstest.OnBattlefieldByID(g, unicorn) {
			t.Errorf("%s: the Unicorn was not sacrificed as a cost", tc.name)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, victim) {
			t.Errorf("%s: the target survived", tc.name)
		}
	}
}

func TestOldB658WallopDestroysOnlyABlueOrBlackFlyer(t *testing.T) {
	for _, tc := range []struct {
		name  string
		color string
		kw    []string
		ok    bool
	}{
		{"blue flyer", "U", []string{"flying"}, true},
		{"black flyer", "B", []string{"flying"}, true},
		{"blue ground", "U", nil, false},
		{"green flyer", "G", []string{"flying"}, false},
	} {
		g := effectstest.NewCatalogGame(t)
		_, opp := oldB658Seats(g)
		victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: tc.name, TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
			Colors: []string{tc.color}, Keywords: tc.kw, Owner: opp.ID, Controller: opp.ID,
		})
		if !tc.ok {
			err := effectstest.CastCatalogSpellErrStrict(t, g, "Wallop", "Sorcery", oldB658Wallop, effectstest.CardTarget(victim))
			if err == nil {
				t.Errorf("%s was a legal target", tc.name)
			}
			continue
		}
		effectstest.CastCatalogSpell(t, g, "Wallop", "Sorcery", oldB658Wallop, effectstest.CardTarget(victim))
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, victim) {
			t.Errorf("%s survived", tc.name)
		}
	}
}

func TestOldB658TranquilityDestroysEveryEnchantmentAndNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB658Seats(g)
	mine := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Tranquility", "Sorcery", oldB658Tranquility, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, mine) || effectstest.OnBattlefieldByID(g, theirs) {
		t.Error("an enchantment survived")
	}
	if !effectstest.OnBattlefieldByID(g, bear) || !effectstest.OnBattlefieldByID(g, rock) {
		t.Error("a non-enchantment was destroyed")
	}
}

func TestOldB658LiberateExilesYourCreatureAndReturnsItAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB658Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Liberate", "Instant", oldB658Liberate, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("a creature you don't control was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Liberate", "Instant", oldB658Liberate, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, bear) || effectstest.B12ZoneOf(g, bear) != game.ZoneExile {
		t.Fatal("the Bear was not exiled")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Bear") {
		t.Error("the Bear did not return at the end step")
	}
}

func TestOldB658TeferisCareSacrificesAnEnchantmentToDestroyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB658Seats(g)
	care := effectstest.PushCatalogPermanent(g, me.ID, "Teferi's Care", "Enchantment", oldB658TeferisCare, false)
	fodder := effectstest.B12Permanent(g, me.ID, "Fodder", "Enchantment")
	target := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{W}{W}")
	if err := oldB658Activate(g, me.ID, care, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear), SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("a creature was a legal target")
	}
	if err := oldB658Activate(g, me.ID, care, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(target)}); err == nil {
		t.Fatal("activated without sacrificing an enchantment")
	}
	if err := oldB658Activate(g, me.ID, care, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(target), SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, target) || effectstest.OnBattlefieldByID(g, fodder) {
		t.Error("the target and the sacrificed enchantment should both be gone")
	}
	if !effectstest.OnBattlefieldByID(g, care) {
		t.Error("Teferi's Care should stay")
	}
}

func TestOldB658TeferisCareCountersAnEnchantmentSpellOnly(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		ok       bool
	}{{"Enchantment", true}, {"Artifact", false}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := oldB658Seats(g)
		care := effectstest.PushCatalogPermanent(g, opp.ID, "Teferi's Care", "Enchantment", oldB658TeferisCare, false)
		spell := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: tc.typeLine, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.FloatMana(t, g, opp, "{U}{U}{U}{U}{U}")
		err := oldB658Activate(g, opp.ID, care, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(spell)})
		if (err == nil) != tc.ok {
			t.Errorf("%s: activation error %v, want legal=%v", tc.typeLine, err, tc.ok)
			continue
		}
		if !tc.ok {
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, spell) || !me.Graveyard.Contains(spell) {
			t.Error("the enchantment spell was not countered")
		}
	}
}

func TestOldB658TsabosAssassinDestroysACreatureOfAMostCommonColour(t *testing.T) {
	setup := func() (*game.Game, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := oldB658Seats(g)
		assassin := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Tsabo's Assassin", TypeLine: "Creature — Phyrexian Zombie Assassin", OracleID: oldB658TsabosAssassin,
			Power: 1, Toughness: 1, Colors: []string{"B"}, Owner: me.ID, Controller: me.ID,
		})
		green := oldB658Colored(g, opp.ID, "Green Bear", "Creature — Bear", "G")
		oldB658Colored(g, opp.ID, "Other Green Bear", "Creature — Bear", "G")
		colourless := oldB658Colored(g, opp.ID, "Golem", "Artifact Creature — Golem")
		return g, assassin, green, colourless, me.ID
	}
	g, assassin, green, colourless, me := setup()
	if err := oldB658Activate(g, me, assassin, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(colourless)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, colourless) {
		t.Error("a colourless creature shares no colour and must survive")
	}
	g, assassin, green, _, me = setup()
	if err := oldB658Activate(g, me, assassin, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(assassin)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, assassin) {
		t.Error("black is not the most common colour, so the Assassin must survive targeting itself")
	}
	g, assassin, green, _, me = setup()
	if err := oldB658Activate(g, me, assassin, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(green)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := oldB658Activate(g, me, assassin, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(green)}); err == nil {
		t.Error("a tapped Assassin activated again")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, green) {
		t.Error("a green creature, green being most common, survived")
	}
}

func TestOldB658SpreadingPlagueDestroysOtherCreaturesSharingAColourWithTheOneThatEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB658Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Spreading Plague", "Enchantment", oldB658SpreadingPlague, false)
	myGreen := oldB658Colored(g, me.ID, "My Green", "Creature — Bear", "G")
	theirGreen := oldB658Colored(g, opp.ID, "Their Green", "Creature — Bear", "G")
	theirRedGreen := oldB658Colored(g, opp.ID, "Their Gruul", "Creature — Bear", "R", "G")
	theirRed := oldB658Colored(g, opp.ID, "Their Red", "Creature — Bear", "R")
	theirBlue := oldB658Colored(g, opp.ID, "Their Blue", "Creature — Bear", "U")
	golem := oldB658Colored(g, opp.ID, "Golem", "Artifact Creature — Golem")
	entering := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: entering, Name: "Green Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, entering, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]bool{
		myGreen: false, theirGreen: false, theirRedGreen: false, theirRed: true, theirBlue: true, golem: true, entering: true,
	} {
		if effectstest.OnBattlefieldByID(g, id) != want {
			t.Errorf("creature %s on battlefield = %v, want %v", id, !want, want)
		}
	}
}

func TestOldB658SpreadingPlagueDoesNothingForAColourlessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB658Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Spreading Plague", "Enchantment", oldB658SpreadingPlague, false)
	green := oldB658Colored(g, opp.ID, "Their Green", "Creature — Bear", "G")
	golem := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: golem, Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, golem, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, green) || !effectstest.OnBattlefieldByID(g, golem) {
		t.Error("a colourless creature entering destroyed something")
	}
}

func TestOldB658TsabosDecreeDiscardsAndDestroysTheChosenTypeOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB658Seats(g)
	handGoblin := effectstest.PushHandCardOfType(opp, "Hand Goblin", "Creature — Goblin")
	handElf := effectstest.PushHandCardOfType(opp, "Hand Elf", "Creature — Elf")
	handSpell := effectstest.PushHandCardOfType(opp, "Goblin Tale", "Sorcery — Goblin")
	theirGoblin := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 2, 2)
	theirElf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 2, 2)
	myGoblin := effectstest.B12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Tsabo's Decree", "Instant", oldB658TsabosDecree, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	if opp.Hand.Contains(handGoblin) || !opp.Graveyard.Contains(handGoblin) {
		t.Error("the Goblin creature card was not discarded")
	}
	if !opp.Hand.Contains(handElf) || !opp.Hand.Contains(handSpell) {
		t.Error("a card that is not a Goblin creature was discarded")
	}
	if effectstest.OnBattlefieldByID(g, theirGoblin) {
		t.Error("their Goblin survived")
	}
	if !effectstest.OnBattlefieldByID(g, theirElf) || !effectstest.OnBattlefieldByID(g, myGoblin) {
		t.Error("an Elf or the caster's own Goblin was destroyed")
	}
}

// bendPile answers the open "separate your lands" prompt for `who`, putting
// `first` in pile 1.
func oldB658Split(t *testing.T, g *game.Game, who *game.Player, first ...uuid.UUID) {
	t.Helper()
	var c *game.PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceRevealPick && pc.Chooser == who.ID {
			c = pc
		}
	}
	if c == nil {
		t.Fatalf("%s was not asked to separate their lands", who.Name)
	}
	if err := g.ResolveRevealPick(c.ID, who.ID, first); err != nil {
		t.Fatalf("%s's split: %v", who.Name, err)
	}
}

func TestOldB658BendOrBreakSeparatesNontokenLandsAndTheChosenPileIsDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB658Seats(g)
	if len(g.Seats) != 2 {
		g.Seats = g.Seats[:2]
	}
	_ = opp
	mineA := effectstest.B12Permanent(g, me.ID, "Mountain A", "Basic Land — Mountain")
	mineB := effectstest.B12Permanent(g, me.ID, "Mountain B", "Basic Land — Mountain")
	token := effectstest.B12Permanent(g, me.ID, "Land Token", "Token Land")
	theirA := effectstest.B12Permanent(g, opp.ID, "Forest A", "Basic Land — Forest")
	theirB := effectstest.B12Permanent(g, opp.ID, "Forest B", "Basic Land — Forest")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Bend or Break", "Sorcery", oldB658BendOrBreak, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The token land is not offered: only the two Mountains are separated.
	if c := oldB658OpenSplit(g, me.ID); c == nil || len(c.ChooseCards) != 2 {
		t.Fatalf("my split offered %+v, want the two nontoken lands", c)
	}
	oldB658Split(t, g, me, mineA)
	oldB658Split(t, g, opp, theirA, theirB)
	// Each player's pile is chosen by the one opponent: the opponent chooses
	// for me, I choose for them.
	pick := func(who *game.Player, index int) {
		t.Helper()
		c := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
		if c == nil || c.Chooser != who.ID {
			t.Fatalf("want %s asked to choose a pile, got %+v", who.Name, c)
		}
		if err := g.ResolveOptionPick(c.ID, who.ID, index); err != nil {
			t.Fatalf("%s's pile: %v", who.Name, err)
		}
	}
	pick(opp, 0) // my pile 1 (Mountain A) is chosen: destroyed; Mountain B is tapped
	pick(me, 1)  // their pile 2 (empty) is chosen: nothing destroyed; both Forests are tapped
	if effectstest.OnBattlefieldByID(g, mineA) {
		t.Error("the chosen pile's land survived")
	}
	for _, id := range []uuid.UUID{mineB, token, theirA, theirB, rock} {
		if !effectstest.OnBattlefieldByID(g, id) {
			t.Errorf("%v was destroyed but is not in a chosen pile", id)
		}
	}
	if !effectstest.B13Tapped(t, g, mineB) || !effectstest.B13Tapped(t, g, theirA) || !effectstest.B13Tapped(t, g, theirB) {
		t.Error("the lands in the other piles were not tapped")
	}
	if effectstest.B13Tapped(t, g, token) || effectstest.B13Tapped(t, g, rock) {
		t.Error("a token land or a non-land was tapped")
	}
}

func oldB658OpenSplit(g *game.Game, who uuid.UUID) *game.PendingChoice {
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceRevealPick && pc.Chooser == who {
			return pc
		}
	}
	return nil
}

// With three opponents, the player whose piles they are names the opponent
// who chooses, and a player with no nontoken lands is skipped.
func TestOldB658BendOrBreakLetsTheOwnerNameWhichOpponentChoosesTheirPile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB658Seats(g)
	if len(g.Seats) < 3 {
		t.Skip("needs three opponents")
	}
	a := effectstest.B12Permanent(g, me.ID, "Mountain A", "Basic Land — Mountain")
	b := effectstest.B12Permanent(g, me.ID, "Mountain B", "Basic Land — Mountain")
	effectstest.CastCatalogSpell(t, g, "Bend or Break", "Sorcery", oldB658BendOrBreak, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oldB658Split(t, g, me, a)
	c := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil || c.Chooser != me.ID || len(c.PickOptions) != len(g.Seats)-1 {
		t.Fatalf("want me asked to name one of %d opponents, got %+v", len(g.Seats)-1, c)
	}
	chosen := c.PickOptions[len(c.PickOptions)-1].Player
	if err := g.ResolveOptionPick(c.ID, me.ID, len(c.PickOptions)-1); err != nil {
		t.Fatalf("naming an opponent: %v", err)
	}
	c = effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil || c.Chooser != chosen {
		t.Fatalf("want the named opponent asked to choose a pile, got %+v", c)
	}
	if err := g.ResolveOptionPick(c.ID, chosen, 1); err != nil {
		t.Fatalf("choosing a pile: %v", err)
	}
	if effectstest.OnBattlefieldByID(g, b) || !effectstest.OnBattlefieldByID(g, a) {
		t.Error("pile 2 (Mountain B) should be destroyed and pile 1 left")
	}
	if !effectstest.B13Tapped(t, g, a) {
		t.Error("the other pile's land was not tapped")
	}
}
