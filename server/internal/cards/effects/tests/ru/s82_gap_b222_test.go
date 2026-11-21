package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b222_test.go — Commander Legends: Battle for Baldur's Gate
// creatures, GAP batch 222 (#5948). Every test plays the real catalog entry
// through the engine.

const (
	gap222Kenku        = "e262b226-9463-40a9-95b2-fe2c3eafc043"
	gap222Doppelganger = "3b8eb602-01e8-432e-8dfc-8fdb62c3aa43"
	gap222Pixie        = "89f1aefd-feb6-45ce-a992-707d027714e7"
	gap222Adept        = "32e5a4ac-69c7-4de0-8b9f-dd86682f2ffd"
	gap222Oceanus      = "bf5d44e7-9ff1-4c9a-a62e-628aa5be720d"
	gap222Familiar     = "de9cee6e-d95b-4e8b-99e4-3b43a0a542aa"
	gap222Whale        = "6de2d53c-06e3-4c56-a5cf-c9cb2db42527"
	gap222Renari       = "99204279-2a50-4dd6-84ec-f84ba31554cf"
	gap222Bane         = "5d22ab00-b454-4407-93d7-65b47f56cbaa"
	gap222Sapphire     = "53015c41-af2d-43e2-a690-d8877537b8dd"
	gap222SeaHag       = "788a1d3f-8d73-4ecd-97ac-77906a873cfe"
	gap222Instigator   = "8c2dd8dc-8623-4518-b3fd-dcb9f647fa98"
)

func TestGap222EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gap222Kenku: "Kenku Artificer", gap222Doppelganger: "Mocking Doppelganger",
		gap222Pixie: "Moonshae Pixie", gap222Pixie + "#1": "Pixie Dust", gap222Adept: "Nimbleclaw Adept",
		gap222Oceanus: "Oceanus Dragon", gap222Familiar: "Pseudodragon Familiar", gap222Whale: "Pursued Whale",
		gap222Renari: "Renari, Merchant of Marvels", gap222Bane: "Sailors' Bane",
		gap222Sapphire: "Sapphire Dragon", gap222Sapphire + "#1": "Psionic Pulse",
		gap222SeaHag: "Sea Hag", gap222SeaHag + "#1": "Aquatic Ingress", gap222Instigator: "Sly Instigator",
	}
	for oracle, name := range want {
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

func TestGap222KenkuArtificerAnimatesANoncreatureArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Test Rock", "Artifact", "", false)
	effectstest.CastHoldCreature(t, g, "Kenku Artificer", "Creature — Bird Artificer", gap222Kenku, 1, 1)
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, rock); n != 3 {
		t.Errorf("%d counters, want 3", n)
	}
	c := effectstest.LayeredCard(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtypeWord("Homunculus") {
		t.Errorf("type line %q, want an Homunculus artifact creature", c.TypeLine)
	}
	if p, tough := effectstest.PtOf(t, g, rock); p != 3 || tough != 3 {
		t.Errorf("%d/%d, want 3/3 (0/0 with three counters)", p, tough)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, rock), "flying") {
		t.Error("the artifact has no flying")
	}
}

func TestGap222KenkuArtificerRefusesACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Artifact Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Kenku Artificer", "Creature — Bird Artificer", gap222Kenku, 1, 1)
	// The only artifact is a creature: the trigger has no legal target.
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Error("an artifact creature was accepted as a noncreature artifact target")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, bear); n != 0 {
		t.Errorf("%d counters on an artifact creature, want 0", n)
	}
}

func TestGap222MockingDoppelgangerCopiesAnOpponentsCreatureAndGoadsTheSameName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Creature", "Creature — Ogre", 3, 3)
	id := effectstest.CastCatalogSpell(t, g, "Mocking Doppelganger", "Creature — Shapeshifter", gap222Doppelganger, nil)
	effectstest.ResolveWithCopyChoice(t, g, theirs)
	c := effectstest.LayeredCard(t, g, id)
	if c.Name != "Grizzly Bear" {
		t.Fatalf("the Doppelganger is %q, want a copy of Grizzly Bear", c.Name)
	}
	for name, want := range map[string]struct {
		id     uuid.UUID
		goaded bool
	}{"their Bear": {theirs, true}, "my Bear": {mine, true}, "the copy": {id, false}, "an unrelated creature": {other, false}} {
		if got := effectstest.LayeredCard(t, g, want.id).Goaded(); got != want.goaded {
			t.Errorf("%s: goaded = %v, want %v", name, got, want.goaded)
		}
	}
}

func TestGap222MockingDoppelgangerOffersOnlyOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Mocking Doppelganger", "Creature — Shapeshifter", gap222Doppelganger, nil)
	for i := 0; i < 8 && effectstest.CopyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := effectstest.CopyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if !effectstest.HasID(pc.CopyOptions, theirs) || effectstest.HasID(pc.CopyOptions, mine) {
		t.Errorf("CopyOptions = %v, want only the opponent's creature", pc.CopyOptions)
	}
}

func TestGap222MoonshaePixieDrawsForEachOpponentDealtCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, attacker)
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Moonshae Pixie", "Creature — Faerie", gap222Pixie, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn (one opponent was dealt combat damage)", hand, got)
	}
}

func TestGap222PixieDustGivesUpToThreeCreaturesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Pixie Dust", TypeLine: "Instant — Adventure", OracleID: gap222Pixie + "#1",
		ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "U", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{a, b} {
		if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, c), "flying") {
			t.Errorf("%s has no flying", c)
		}
	}
}

func TestGap222NimbleclawAdeptUntapsTwoOthersOncePerTurnAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	adept := effectstest.PushCatalogPermanent(g, me.ID, "Nimbleclaw Adept", "Creature — Dragon Wizard", gap222Adept, false)
	a := effectstest.PushCatalogPermanent(g, me.ID, "Rock A", "Artifact", "", false)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Rock B", "Artifact", "", false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, a).Tapped = true
		effectstest.FindBattlefieldCardForTest(g, b).Tapped = true
	})
	params := game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}
	if err := g.ActivateCatalogAbility(me.ID, adept, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := g.LookupCardForEffect(id); c.Tapped {
			t.Errorf("%s is still tapped", c.Name)
		}
	}
	// Once each turn: untap the Adept and try again.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, adept).Tapped = false })
	params.Strict = true
	if err := g.ActivateCatalogAbility(me.ID, adept, 0, params); err == nil {
		t.Error("the Adept activated twice in one turn")
	}
}

func TestGap222NimbleclawAdeptNeverTargetsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	adept := effectstest.PushCatalogPermanent(g, me.ID, "Nimbleclaw Adept", "Creature — Dragon Wizard", gap222Adept, false)
	a := effectstest.PushCatalogPermanent(g, me.ID, "Rock A", "Artifact", "", false)
	err := g.ActivateCatalogAbility(me.ID, adept, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: adept}}})
	if err == nil {
		t.Error("the Adept targeted itself")
	}
}

func TestGap222OceanusDragonTapsAndGoadsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Oceanus Dragon", "Creature — Dragon", gap222Oceanus, 3, 5)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("my own creature was accepted as a target")
	}
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, victim)
	if !c.Tapped || !c.Goaded() {
		t.Errorf("victim tapped=%v goaded=%v, want both", c.Tapped, c.Goaded())
	}
}

func TestGap222PseudodragonFamiliarGivesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	fam := effectstest.PushCatalogPermanent(g, me.ID, "Pseudodragon Familiar", "Creature — Dragon", gap222Familiar, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "U", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, fam, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "flying") {
		t.Error("the Bear has no flying")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, fam), "flying") {
		t.Error("the Familiar lost its printed flying")
	}
}

func TestGap222PursuedWhaleGivesEachOpponentAPirate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Pursued Whale", "Creature — Whale", gap222Whale, 8, 8)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		n, _ := effectstest.CountTokensNamed(g, p.ID, "Pirate")
		want := 1
		if p.ID == me.ID {
			want = 0
		}
		if n != want {
			t.Errorf("%s controls %d Pirates, want %d", p.Name, n, want)
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Pirate" {
			continue
		}
		if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 1 || tough != 1 {
			t.Errorf("Pirate is %d/%d, want 1/1", p, tough)
		}
	}
}

func TestGap222PursuedWhaleTaxesOpponentsSpellsThatTargetIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	whale := effectstest.B12Push(g, me.ID, "Pursued Whale", "Creature — Whale", gap222Whale, 8, 8)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
		OracleID: effectstest.LightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	g.Turn.ActiveSeat = 1
	effectstest.ApaMana(opp, "R")
	err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: whale}}})
	if err == nil {
		t.Fatal("a Lightning Bolt targeting the Whale was cast for {R}")
	}
	effectstest.ApaMana(opp, "R", "C", "C", "C")
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: whale}}}); err != nil {
		t.Errorf("the taxed Bolt ({3}{R}) was refused: %v", err)
	}
}

func TestGap222RenariGivesDragonAndArtifactSpellsFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Renari, Merchant of Marvels", "Legendary Creature — Dragon Artificer", gap222Renari, false)
	// In the opponent's turn, on their main phase, only a flash spell may be cast.
	g.Turn.ActiveSeat = 1
	for name, tc := range map[string]struct {
		typeLine string
		ok       bool
	}{
		"Test Dragon":   {"Creature — Dragon", true},
		"Test Artifact": {"Artifact", true},
		"Test Bear":     {"Creature — Bear", false},
	} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: tc.typeLine, Owner: me.ID, Controller: me.ID})
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if (err == nil) != tc.ok {
			t.Errorf("%s (%s): cast err = %v, want success = %v", name, tc.typeLine, err, tc.ok)
		}
	}
}

func gap222ReducerPile(g *game.Game, me, opp *game.Player) {
	add := func(p *game.Player, z *game.Zone, name, typeLine string, layout string) {
		z.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Layout: layout, Owner: p.ID, Controller: p.ID})
	}
	add(me, me.Graveyard, "Old Instant", "Instant", "")
	add(me, me.Graveyard, "Old Sorcery", "Sorcery", "")
	add(me, me.Graveyard, "Old Creature", "Creature — Bear", "")
	add(me, g.Exile, "Exiled Adventurer", "Creature — Bear", game.LayoutAdventure)
	add(me, g.Exile, "Exiled Bear", "Creature — Bear", "")
	add(opp, g.Exile, "Their Instant", "Instant", "")
}

func TestGap222SailorsBaneCostsLessForInstantsSorceriesAndAdventuresYouOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	gap222ReducerPile(g, me, opp)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sailors' Bane", TypeLine: "Creature — Dragon Turtle", OracleID: gap222Bane,
		ManaCost: "{7}{U}{U}", Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	// Three qualifying cards: {7}{U}{U} costs {4}{U}{U}, six mana.
	effectstest.ApaMana(me, "U", "U", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for five mana; the three reducers should leave a cost of six")
	}
	effectstest.ApaMana(me, "U", "U", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("cast for six mana refused: %v", err)
	}
}

func TestGap222SapphireDragonScriesOnAttackAndBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Sapphire Dragon", "Creature — Dragon", gap222Sapphire, 5, 6)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.DeclareAttack(t, g, opp.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	scry := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceScry && c.Chooser == me.ID {
			scry = true
		}
	}
	if !scry {
		t.Error("no scry prompt after attacking")
	}
}

func TestGap222SeaHagShrinksOnlyOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Ogre", 5, 5)
	effectstest.CastHoldCreature(t, g, "Sea Hag", "Creature — Hag", gap222SeaHag, 3, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, theirs); p != 1 {
		t.Errorf("their creature has power %d, want 1", p)
	}
	if p, _ := effectstest.PtOf(t, g, mine); p != 5 {
		t.Errorf("my creature has power %d, want 5", p)
	}
}

func TestGap222AquaticIngressPumpsAndMakesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Aquatic Ingress", TypeLine: "Instant — Adventure", OracleID: gap222SeaHag + "#1",
		ManaCost: "{2}{U}", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "U", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, a); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.BrAttack(t, g, a)
	if err := g.DeclareBlocker(blocker, a); err == nil {
		t.Error("the pumped creature was blocked")
	}
}

func TestGap222SlyInstigatorMakesAnOpponentsCreatureUnblockableAndGoaded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	inst := effectstest.PushCatalogPermanent(g, me.ID, "Sly Instigator", "Creature — Human Wizard", gap222Instigator, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, inst, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, victim)
	if !c.Goaded() {
		t.Error("the creature is not goaded")
	}
	if effectstest.RestrictionsOf(t, g, victim)&game.CantBeBlocked == 0 {
		t.Error("the creature can be blocked")
	}
}
