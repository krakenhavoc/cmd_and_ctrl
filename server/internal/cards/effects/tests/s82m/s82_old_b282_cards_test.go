package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b282_cards_test.go — S82 older sets, Shadows over Innistrad
// batch 282 (#5543): delirium cards.

const (
	oldB282Stallion = "0dc6a90e-c517-445e-9a93-29b3119fb4dc"
	oldB282Gloom    = "245aa786-d607-414f-ab55-b003b2b9faf4"
	oldB282Fiend    = "2aa03852-6c38-4e91-afed-0c4cd5964598"
	oldB282Demon    = "3dda8d3d-89e2-4600-be8d-9b76b4f36e33"
	oldB282Topple   = "4a71f404-e48d-4c86-882e-4e21c8210dea"
	oldB282Slaught  = "518cf9ce-14aa-4f51-81bf-f8c2692a4244"
	oldB282Surgery  = "545f56c6-b10f-4868-8aa8-ea3154980854"
	oldB282Might    = "563a8e37-b90e-4ea1-9366-8502b403e13e"
)

func TestOldB282CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB282Stallion: "Stallion of Ashmouth", oldB282Gloom: "Autumnal Gloom", oldB282Gloom + "#1": "Ancient of the Equinox",
		oldB282Fiend: "Gibbering Fiend", oldB282Demon: "Mindwrack Demon", oldB282Topple: "Topplegeist",
		oldB282Slaught: "To the Slaughter", oldB282Surgery: "Invasive Surgery", oldB282Might: "Might Beyond Reason",
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

func TestOldB282StallionPumpsOnlyWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	horse := effectstest.B12Push(g, me.ID, "Stallion of Ashmouth", "Creature — Nightmare Horse", oldB282Stallion, 3, 3)
	effectstest.B24SeedTypes(me, 3)
	effectstest.ApaMana(me, "B", "C")
	if err := g.ActivateCatalogAbility(me.ID, horse, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the pump was activated with three card types in the graveyard")
	}
	effectstest.GyCard(me, "Old Artifact", "Artifact", "", "")
	effectstest.ApaActivate(t, g, me, horse, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.EffectivePower(t, g, horse), effectstest.EffectiveToughness(t, g, horse); p != 4 || tou != 4 {
		t.Errorf("Stallion is %d/%d, want 4/4", p, tou)
	}
}

func TestOldB282AutumnalGloomMillsAndTransformsOnlyWithDelirium(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		gc := game.Card{
			InstanceID: uuid.New(), OracleID: oldB282Gloom, Layout: game.LayoutTransform, Owner: me.ID, Controller: me.ID,
			Faces: []game.Face{
				{Name: "Autumnal Gloom", TypeLine: "Enchantment", ManaCost: "{2}{G}", Colors: []string{"G"}},
				{Name: "Ancient of the Equinox", TypeLine: "Creature — Treefolk", Colors: []string{"G"}, Power: 4, Toughness: 4},
			},
		}
		gc.SetFace(0)
		gloom := effectstest.PushBattlefieldCardWithTimestamp(g, gc)
		effectstest.SeedLibraryTop(me, "Top", "Land", "")
		before := me.Graveyard.Size()
		effectstest.ApaMana(me, "B")
		effectstest.ApaActivate(t, g, me, gloom, 0, game.ActivateAbilityParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Graveyard.Size(); got != before+1 {
			t.Fatalf("graveyard %d, want %d after the mill", got, before+1)
		}
		if delirium {
			effectstest.B24SeedTypes(me, 4)
		}
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.B12Card(t, g, gloom)
		if got := effects.TransformedPermanent(c); got != delirium {
			t.Errorf("delirium=%v: transformed = %v", delirium, got)
		}
		if delirium {
			if !effectstest.HasEffectiveKeyword(t, g, gloom, "trample") || !effectstest.HasEffectiveKeyword(t, g, gloom, "hexproof") {
				t.Error("Ancient of the Equinox lacks trample or hexproof")
			}
			if p, tou := effectstest.EffectivePower(t, g, gloom), effectstest.EffectiveToughness(t, g, gloom); p != 4 || tou != 4 {
				t.Errorf("Ancient of the Equinox is %d/%d, want 4/4", p, tou)
			}
		}
	}
}

func TestOldB282GibberingFiendPingsOnEntryAndAtOpponentUpkeeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	l1, l2, mine := opp.Life, opp2.Life, me.Life
	effectstest.CastAndResolveCreature(t, g, "Gibbering Fiend", "Creature — Devil", oldB282Fiend)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != l1-1 || opp2.Life != l2-1 || me.Life != mine {
		t.Errorf("entry: life %d/%d/%d from %d/%d/%d, want each opponent -1", me.Life, opp.Life, opp2.Life, mine, l1, l2)
	}
	// Without delirium the upkeep trigger does nothing.
	l1 = opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != l1 {
		t.Errorf("no delirium: the opponent lost %d at their upkeep", l1-opp.Life)
	}
	effectstest.B24SeedTypes(me, 4)
	l1, l2 = opp.Life, opp2.Life
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if opp2.Life != l2-1 || opp.Life != l1 {
		t.Errorf("delirium: life %d/%d from %d/%d, want only the upkeep's player -1", opp.Life, opp2.Life, l1, l2)
	}
	mine = me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Error("the Fiend hurt its controller on their own upkeep")
	}
}

func TestOldB282MindwrackDemonMillsFourAndCostsLifeUnlessDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 6; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Land", "")
	}
	effectstest.AdvanceToMain(t, g)
	before := me.Graveyard.Size()
	demon := effectstest.CastAndResolveCreature(t, g, "Mindwrack Demon", "Creature — Demon", oldB282Demon)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != before+4 {
		t.Errorf("graveyard %d, want %d (four milled)", got, before+4)
	}
	if !effectstest.HasEffectiveKeyword(t, g, demon, "flying") {
		t.Error("no flying")
	}
	// Four milled lands is one card type: no delirium, so the upkeep costs 4.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("no delirium: life %d → %d, want -4", life, me.Life)
	}
	effectstest.B24SeedTypes(me, 4)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	life = me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("delirium: life %d → %d, want no loss", life, me.Life)
	}
}

func TestOldB282TopplegeistTapsOnEntryAndAtOpponentUpkeepsWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp2.ID, "Other", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastAndResolveCreature(t, g, "Topplegeist", "Creature — Spirit", oldB282Topple)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, theirs) || effectstest.IsTapped(g, other) || effectstest.IsTapped(g, mine) {
		t.Fatalf("entry: tapped = %v/%v/%v, want only the chosen opposing creature", effectstest.IsTapped(g, theirs), effectstest.IsTapped(g, other), effectstest.IsTapped(g, mine))
	}
	// Opponent 2's upkeep, no delirium: nothing.
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, other) {
		t.Error("no delirium: a creature was tapped at the upkeep")
	}
	effectstest.B24SeedTypes(me, 4)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	// "Theirs" untapped in the untap step, and is tapped again by the trigger.
	if !effectstest.IsTapped(g, theirs) {
		t.Error("delirium: the upkeep player's creature was not tapped")
	}
	if effectstest.IsTapped(g, other) {
		t.Error("delirium: another player's creature was tapped")
	}
}

func TestOldB282ToTheSlaughterEdict(t *testing.T) {
	setup := func(t *testing.T, delirium bool) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		bear2 := effectstest.B12Creature(g, opp.ID, "Bear Two", "Creature — Bear", 2, 2)
		walker := effectstest.P1g2Planeswalker(g, opp.ID, 3)
		if delirium {
			effectstest.B24SeedTypes(me, 4)
		}
		effectstest.CastCatalogSpell(t, g, "To the Slaughter", "Instant", oldB282Slaught, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		return g, opp, bear, bear2, walker
	}
	t.Run("without delirium one permanent", func(t *testing.T) {
		g, opp, bear, bear2, walker := setup(t, false)
		effectstest.AnswerSacrifice(t, g, opp.ID, walker)
		g.SettleResolution()
		if g.Battlefield.Contains(walker) || !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(bear2) {
			t.Error("the chosen planeswalker should be the only permanent sacrificed")
		}
	})
	t.Run("with delirium a creature and a planeswalker", func(t *testing.T) {
		g, opp, bear, bear2, walker := setup(t, true)
		for i := 0; i < 2; i++ {
			c := effectstest.SacrificeChoiceFor(g, opp.ID)
			if c == nil {
				break
			}
			// The prompts stack up newest-first: the planeswalker's, then the creature's.
			pick := walker
			if i == 1 {
				pick = bear
			}
			effectstest.AnswerSacrifice(t, g, opp.ID, pick)
		}
		g.SettleResolution()
		if g.Battlefield.Contains(bear) || g.Battlefield.Contains(walker) || !g.Battlefield.Contains(bear2) {
			t.Errorf("bear %v walker %v bear2 %v on the battlefield, want only Bear Two left",
				g.Battlefield.Contains(bear), g.Battlefield.Contains(walker), g.Battlefield.Contains(bear2))
		}
	})
}

func TestOldB282InvasiveSurgerySearchesOnlyWithDelirium(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		x := effectstest.ExtractionSetUp(t, "Divination", "Sorcery")
		if delirium {
			effectstest.B24SeedTypes(x.Me, 4)
		}
		spell := uuid.New()
		x.Opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Divination", TypeLine: "Sorcery", Owner: x.Opp.ID, Controller: x.Opp.ID})
		x.Cards["cast"] = spell
		g := x.G
		effectstest.AdvanceToMainOf(t, g, 1)
		if err := g.CastSpell(x.Opp.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast the sorcery: %v", err)
		}
		surgery := uuid.New()
		x.Me.Hand.PushTop(game.Card{InstanceID: surgery, Name: "Invasive Surgery", TypeLine: "Instant", OracleID: oldB282Surgery, Owner: x.Me.ID, Controller: x.Me.ID})
		if err := g.CastSpell(x.Me.ID, surgery, game.CastSpellParams{Targets: x.Target("cast")}); err != nil {
			t.Fatalf("cast Invasive Surgery: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !x.Opp.Graveyard.Contains(spell) {
			t.Fatalf("delirium=%v: the sorcery was not countered", delirium)
		}
		if !delirium {
			x.NotExiled(t, "cast", "gy", "gy 2", "hand", "lib")
			if x.Searches() != 0 {
				t.Error("no delirium: a search happened")
			}
			continue
		}
		effectstest.WantKeys(t, x.Candidates(t), "cast", "gy", "gy 2", "hand", "lib")
		effectstest.AnswerSearchByID(t, g, x.Me.ID, x.Cards["cast"], x.Cards["gy 2"], x.Cards["hand"])
		x.Exiled(t, "cast", "gy 2", "hand")
		x.NotExiled(t, "gy", "lib", "gy other", "hand other", "lib other")
		if x.Searches() != 1 {
			t.Errorf("%d searches, want 1", x.Searches())
		}
	}
}

func TestOldB282InvasiveSurgeryRefusesAnInstant(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Shock", "Instant")
	x.Cards["cast"] = effectstest.TdCastFromHand(t, x.G, game.Card{Name: "Shock", TypeLine: "Instant"}, game.CastSpellParams{})
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Invasive Surgery", "Instant", oldB282Surgery, x.Target("cast")); err == nil {
		t.Error("an instant spell was a legal target")
	}
}

func TestOldB282MightBeyondReasonCounters(t *testing.T) {
	for _, delirium := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		if delirium {
			effectstest.B24SeedTypes(me, 4)
		}
		effectstest.CastCatalogSpell(t, g, "Might Beyond Reason", "Instant", oldB282Might, effectstest.CardTarget(bear))
		effectstest.PassPriorityAroundTable(t, g)
		want := 2
		if delirium {
			want = 3
		}
		if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != want {
			t.Errorf("delirium=%v: %d counters, want %d", delirium, got, want)
		}
	}
}
