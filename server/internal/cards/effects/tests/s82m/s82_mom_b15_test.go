package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b15_test.go — S82 batch 15: the ten March of the Machine
// Phyrexian transformers with simple back faces.

type momB15Face struct {
	name, typeLine, cost, color string
	power, toughness            int
}

type momB15Row struct {
	oracle      string
	front, back momB15Face
	activate    string // the mana to float for the activation (generic and coloured)
	lifeColor   string // the colour the Phyrexian symbol would take
}

var momB15Rows = []momB15Row{
	{"80ecb069-36e2-490d-9f6c-ef553c00e997", momB15Face{"Aetherblade Agent", "Creature — Human Rogue", "{1}{B}", "B", 1, 1}, momB15Face{"Gitaxian Mindstinger", "Creature — Phyrexian Rogue", "", "B", 3, 3}, "{C}{C}{C}{C}", "U"},
	{"ea405cb4-a21b-4a26-98f3-251e65ea2aef", momB15Face{"Blightreaper Thallid", "Creature — Fungus", "{1}{B}", "B", 2, 2}, momB15Face{"Blightsower Thallid", "Creature — Phyrexian Fungus", "", "B", 3, 3}, "{C}{C}{C}", "G"},
	{"b06067e7-330e-47de-be59-c2874eedba1d", momB15Face{"Bonded Herdbeast", "Creature — Beast", "{4}{G}", "G", 4, 5}, momB15Face{"Plated Kilnbeast", "Creature — Phyrexian Beast", "", "G", 7, 5}, "{C}{C}{C}{C}", "R"},
	{"ac5487ba-6d6e-47ee-ae29-776e186789cc", momB15Face{"Captive Weird", "Creature — Weird", "{U}", "U", 1, 3}, momB15Face{"Compleated Conjurer", "Creature — Phyrexian Weird", "", "U", 3, 3}, "{C}{C}{C}", "R"},
	{"3bafc004-1c90-4939-9dcb-72667c170e39", momB15Face{"Gnottvold Hermit", "Creature — Troll", "{3}{G}", "G", 4, 4}, momB15Face{"Chrome Host Hulk", "Creature — Phyrexian Troll", "", "G", 5, 5}, "{C}{C}{C}{C}{C}", "U"},
	{"c4e51cba-e5f7-4f44-9ae9-9223a2efd004", momB15Face{"Harried Artisan", "Creature — Human Artificer", "{2}{R}", "R", 2, 3}, momB15Face{"Phyrexian Skyflayer", "Creature — Phyrexian Artificer", "", "R", 3, 4}, "{C}{C}{C}", "W"},
	{"bf0ff4b9-3543-4cfa-bfcd-f9504d27f00f", momB15Face{"Herbology Instructor", "Creature — Treefolk Druid", "{1}{G}", "G", 1, 3}, momB15Face{"Malady Invoker", "Creature — Phyrexian Treefolk", "", "G", 3, 3}, "{C}{C}{C}{C}{C}{C}", "B"},
	{"f08d8fc5-c7fb-4913-92fc-06a0d4311d71", momB15Face{"Khenra Spellspear", "Creature — Jackal Warrior", "{1}{R}", "R", 2, 2}, momB15Face{"Gitaxian Spellstalker", "Creature — Phyrexian Jackal", "", "R", 3, 3}, "{C}{C}{C}", "U"},
	{"d849be9d-0447-4f0b-a382-3f4ceb98f9a5", momB15Face{"Order of the Mirror", "Creature — Human Knight", "{1}{U}", "U", 2, 1}, momB15Face{"Order of the Alabaster Host", "Creature — Phyrexian Knight", "", "U", 3, 3}, "{C}{C}{C}", "W"},
	{"f810e185-4498-483b-ab37-0c583dae0eaf", momB15Face{"Pyretic Prankster", "Creature — Devil", "{1}{R}", "R", 2, 1}, momB15Face{"Glistening Goremonger", "Creature — Phyrexian Devil", "", "R", 3, 2}, "{C}{C}{C}", "B"},
}

func momB15Row_(name string) momB15Row {
	for _, r := range momB15Rows {
		if r.front.name == name {
			return r
		}
	}
	panic("no row " + name)
}

// momB15Card builds the card as the importer does: two faces, front up.
func (r momB15Row) Card(owner uuid.UUID, face int) game.Card {
	mk := func(f momB15Face) game.Face {
		return game.Face{Name: f.name, TypeLine: f.typeLine, ManaCost: f.cost, Colors: []string{f.color}, Power: f.power, Toughness: f.toughness}
	}
	c := game.Card{
		InstanceID: uuid.New(), OracleID: r.oracle, Layout: game.LayoutTransform,
		Owner: owner, Controller: owner,
		Faces: []game.Face{mk(r.front), mk(r.back)},
	}
	c.SetFace(face)
	return c
}

func (r momB15Row) Put(g *game.Game, owner uuid.UUID, face int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, r.Card(owner, face))
}

func momB15FaceName(g *game.Game, id uuid.UUID) string {
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		return ""
	}
	return c.Name
}

// transform activates the transform ability (ability 0) with `mana` floating.
func (r momB15Row) transform(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if mana != "" {
		effectstest.DcMana(t, g, me.ID, mana)
	}
	err := g.ActivateCatalogAbility(me.ID, id, 0, params)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestMomB15AllRegisteredFullWithBothFaces(t *testing.T) {
	for _, r := range momB15Rows {
		for i, name := range []string{r.front.name, r.back.name} {
			oracle := r.oracle
			if i == 1 {
				oracle += "#1"
			}
			spec, ok := effects.Lookup(oracle)
			if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
				t.Errorf("%s: %+v %v", name, spec, ok)
			}
		}
	}
}

// Every card: {N}{X/P} turns it over, once the Phyrexian symbol is paid
// with its colour or with 2 life, at sorcery speed only.
func TestMomB15TransformAbilityIsPaidInManaOrLifeAtSorcerySpeed(t *testing.T) {
	for _, r := range momB15Rows {
		t.Run(r.front.name, func(t *testing.T) {
			t.Run("paid in mana", func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				id := r.Put(g, me.ID, 0)
				life := me.Life
				if err := r.transform(t, g, me, id, r.activate+"{"+r.lifeColor+"}", game.ActivateAbilityParams{Strict: true}); err != nil {
					t.Fatalf("activate: %v", err)
				}
				if got := momB15FaceName(g, id); got != r.back.name {
					t.Fatalf("face = %q, want %q", got, r.back.name)
				}
				if me.Life != life {
					t.Errorf("life %d -> %d, want unchanged", life, me.Life)
				}
				if p, tough := effectstest.EffectivePTOf(t, g, id); p != r.back.power || tough != r.back.toughness {
					t.Errorf("back face is %d/%d, want %d/%d", p, tough, r.back.power, r.back.toughness)
				}
			})
			t.Run("paid in life", func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				id := r.Put(g, me.ID, 0)
				life := me.Life
				if err := r.transform(t, g, me, id, r.activate, game.ActivateAbilityParams{Strict: true, PhyrexianLife: 1}); err != nil {
					t.Fatalf("activate: %v", err)
				}
				if got := momB15FaceName(g, id); got != r.back.name {
					t.Fatalf("face = %q, want %q", got, r.back.name)
				}
				if me.Life != life-2 {
					t.Errorf("life %d -> %d, want 2 paid", life, me.Life)
				}
			})
			t.Run("a pool that can't pay", func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				id := r.Put(g, me.ID, 0)
				if err := r.transform(t, g, me, id, "{C}", game.ActivateAbilityParams{Strict: true}); err == nil {
					t.Fatal("activated without paying")
				}
				if got := momB15FaceName(g, id); got != r.front.name {
					t.Errorf("face = %q, want it unchanged", got)
				}
			})
			t.Run("not at instant speed", func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				id := r.Put(g, me.ID, 0)
				effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
				effectstest.DcMana(t, g, me.ID, r.activate+"{"+r.lifeColor+"}")
				if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
					t.Fatal("activated outside a main phase")
				}
			})
		})
	}
}

func TestMomB15BackFaceKeywords(t *testing.T) {
	for _, tc := range []struct {
		card string
		want []string
	}{
		{"Bonded Herdbeast", []string{"menace"}},
		{"Harried Artisan", []string{"flying", "haste"}},
		{"Aetherblade Agent", []string{"deathtouch"}},
		{"Khenra Spellspear", []string{"trample", "prowess"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		r := momB15Row_(tc.card)
		front := r.Put(g, me.ID, 0)
		back := r.Put(g, me.ID, 1)
		for _, kw := range tc.want {
			if !effectstest.HasKeywordOnBattlefield(t, g, back, kw) {
				t.Errorf("%s lacks %s", r.back.name, kw)
			}
		}
		if tc.card == "Bonded Herdbeast" && effectstest.HasKeywordOnBattlefield(t, g, front, "menace") {
			t.Error("the front face has menace")
		}
	}
	// "Defender" stays on the Weird only until it turns over.
	g := effectstest.NewCatalogGame(t)
	r := momB15Row_("Captive Weird")
	if !effectstest.HasKeywordOnBattlefield(t, g, r.Put(g, g.Seats[0].ID, 0), "defender") {
		t.Error("Captive Weird lacks defender")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, r.Put(g, g.Seats[0].ID, 1), "defender") {
		t.Error("Compleated Conjurer has defender")
	}
}

func TestMomB15SpellstalkerHasWardAndTwoProwess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	r := momB15Row_("Khenra Spellspear")
	id := r.Put(g, me.ID, 1)
	opp := g.Seats[1]
	_ = opp
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardByID(g, id)
	if n := game.ProwessCount(c); n != 2 {
		t.Errorf("prowess instances = %d, want 2", game.ProwessCount(c))
	}
	if spec, _ := effects.Lookup(r.oracle + "#1"); len(spec.Triggered) == 0 {
		t.Error("no ward trigger declared on the back face")
	}
}

func TestMomB15HerbologyInstructorGainsThreeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	r := momB15Row_("Herbology Instructor")
	life := me.Life
	id := r.Put(g, me.ID, 0)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: me.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

func TestMomB15MaladyInvokerShrinksAnOpponentsCreatureByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	r := momB15Row_("Herbology Instructor")
	id := r.Put(g, me.ID, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 6)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		for _, c := range pick.PickTargetCards {
			if c == mine {
				t.Error("offered my own creature")
			}
		}
		effectstest.AnswerPickTarget(t, g, theirs)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 4 || tough != 3 {
		t.Errorf("their Ogre is %d/%d, want 4/3 (-0/-3)", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 2 || tough != 4 {
		t.Errorf("my Bear is %d/%d, want unchanged", p, tough)
	}
}

func TestMomB15CompleatedConjurerExilesTheTopCardOnTransformOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	r := momB15Row_("Captive Weird")
	back := r.Put(g, me.ID, 1) // enters transformed: it never transformed
	_ = back
	lib, ex := me.Library.Size(), g.Exile.Size()
	if me.Library.Size() != lib || g.Exile.Size() != ex {
		t.Fatal("setup")
	}
	id := r.Put(g, me.ID, 0)
	if err := r.transform(t, g, me, id, "{C}{C}{C}{R}", game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Library.Size() != lib-1 || g.Exile.Size() != ex+1 {
		t.Errorf("library %d -> %d, exile %d -> %d; want the top card exiled", lib, me.Library.Size(), ex, g.Exile.Size())
	}
}

func TestMomB15ChromeHostHulkSetsAnotherCreaturesBaseTo5x5(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	r := momB15Row_("Gnottvold Hermit")
	hulk := r.Put(g, me.ID, 1)
	small := effectstest.B12Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, hulk)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, c := range pick.PickTargetCards {
		if c == hulk {
			t.Error("the Hulk can target itself")
		}
	}
	effectstest.AnswerPickTarget(t, g, small)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, small); p != 5 || tough != 5 {
		t.Errorf("the Mouse is %d/%d, want 5/5", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, hulk); p != 5 || tough != 5 {
		t.Errorf("the Hulk is %d/%d, want 5/5", p, tough)
	}
}

func TestMomB15AlabasterHostShrinksEachBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	r := momB15Row_("Order of the Mirror")
	host := r.Put(g, me.ID, 1)
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 3, 3)
	effectstest.WoeS2Block(g, opp.ID, a, host)
	effectstest.WoeS2Block(g, opp.ID, b, host)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, a); p != 1 || tough != 1 {
		t.Errorf("blocker A is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, b); p != 2 || tough != 2 {
		t.Errorf("blocker B is %d/%d, want 2/2", p, tough)
	}
}

func TestMomB15AlabasterHostBlockingIsNotBecomingBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	r := momB15Row_("Order of the Mirror")
	host := r.Put(g, opp.ID, 1)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.WoeS2Block(g, opp.ID, host, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, attacker); p != 2 || tough != 2 {
		t.Errorf("the attacker is %d/%d, want unchanged", p, tough)
	}
}

func TestMomB15GoremongerDiesAndEachOpponentSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	r := momB15Row_("Pyretic Prankster")
	id := r.Put(g, me.ID, 1)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	c1 := effectstest.B12Creature(g, o1.ID, "Their Bear", "Creature — Bear", 2, 2)
	a2 := effectstest.B12Permanent(g, o2.ID, "Their Idol", "Artifact")
	_ = o3
	effectstest.UsgB3Destroy(t, g, id)
	if !g.Battlefield.Contains(mine) {
		t.Error("I sacrificed my own creature")
	}
	if g.Battlefield.Contains(c1) {
		t.Error("the first opponent kept their only creature")
	}
	if g.Battlefield.Contains(a2) {
		t.Error("the second opponent kept their only artifact")
	}
}

func TestMomB15FrontPranksterDoesNothingWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	id := momB15Row_("Pyretic Prankster").Put(g, me.ID, 0)
	c1 := effectstest.B12Creature(g, o1.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.UsgB3Destroy(t, g, id)
	if !g.Battlefield.Contains(c1) {
		t.Error("the front face's death made an opponent sacrifice")
	}
}

func TestMomB15MindstingerDrawsOnCombatDamageToAPlayerOrBattleOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := momB15Row_("Aetherblade Agent").Put(g, me.ID, 1)
	emit := func(combat bool, target uuid.UUID) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: id, CardID: id, Target: target, Amount: 3, Combat: combat, Actor: me.ID})
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	hand := me.Hand.Size()
	emit(false, opp.ID)
	if me.Hand.Size() != hand {
		t.Error("noncombat damage drew a card")
	}
	bystander := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	emit(true, bystander)
	if me.Hand.Size() != hand {
		t.Error("combat damage to a creature drew a card")
	}
	emit(true, opp.ID)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	battle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "A Battle", TypeLine: "Battle — Siege",
		Counters: map[string]int{"defense": 5}, Owner: opp.ID, Controller: opp.ID})
	emit(true, battle)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand = %d, want %d after damage to a battle", me.Hand.Size(), hand+2)
	}
}

func momB15Saprolings(g *game.Game, owner uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == "Phyrexian Saproling" && c.IsToken() {
				n++
			}
		}
	})
	return n
}

func TestMomB15BlightsowerThallidMakesASaprolingWhenItTransformsAndWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	r := momB15Row_("Blightreaper Thallid")
	id := r.Put(g, me.ID, 0)
	if err := r.transform(t, g, me, id, "{C}{C}{C}{G}", game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := momB15Saprolings(g, me.ID); n != 1 {
		t.Fatalf("saprolings after transforming = %d, want 1", n)
	}
	effectstest.UsgB3Destroy(t, g, id)
	if n := momB15Saprolings(g, me.ID); n != 2 {
		t.Errorf("saprolings after dying = %d, want 2", n)
	}
}

func TestMomB15BlightreaperDyingMakesNothingAndEnteringTransformedMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	r := momB15Row_("Blightreaper Thallid")
	front := r.Put(g, me.ID, 0)
	effectstest.UsgB3Destroy(t, g, front)
	if n := momB15Saprolings(g, me.ID); n != 0 {
		t.Errorf("the front face's death made %d Saprolings", n)
	}
	r.Put(g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := momB15Saprolings(g, me.ID); n != 0 {
		t.Errorf("entering transformed made %d Saprolings", n)
	}
}

func TestMomB15PhyrexianSaprolingTokenIsAsPrinted(t *testing.T) {
	tok := effects.TokenCard("1/1 green Phyrexian Saproling")
	if tok.Name != "Phyrexian Saproling" || tok.TypeLine != "Token Creature — Phyrexian Saproling" || tok.Power != 1 || tok.Toughness != 1 || len(tok.Colors) != 1 || tok.Colors[0] != "G" {
		t.Errorf("token = %+v", tok)
	}
}
