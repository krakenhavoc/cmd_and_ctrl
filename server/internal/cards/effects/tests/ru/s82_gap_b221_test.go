package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b221_test.go — Commander Legends: Battle for Baldur's Gate
// creatures, GAP batch 221 (#5948). Every test plays the real catalog entry
// through the engine.

const (
	gap221Sprite     = "a9d8ab76-70a4-475e-b87e-4737c090553a"
	gap221Alora      = "a26432ff-ae60-454d-a6de-fec3feafa669"
	gap221Astral     = "a20e2e1a-90d8-4073-96e9-efce86f0178e"
	gap221Looter     = "f74e53e1-e575-4386-a128-5c87124987d4"
	gap221Gale       = "cca15007-2faf-4696-a4c5-2d7b6c1ec5b5"
	gap221Merfolk    = "f0f79389-95f1-4110-9d3d-91779bc71604"
	gap221Illithid   = "299ffd3a-29cf-4667-b191-dd30d09f799c"
	gap221MistDragon = "c6867301-296c-4840-aad4-1ad8345c091c"
	gap221Ulder      = "792b9752-2d46-48ca-abdd-408a787612da"
)

func TestGap221EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gap221Sprite: "Sprite Dragon", gap221Alora: "Alora, Merry Thief", gap221Astral: "Astral Dragon",
		gap221Looter: "Dragonborn Looter", gap221Gale: "Gale, Waterdeep Prodigy", gap221Merfolk: "Gray Harbor Merfolk",
		gap221Illithid: "Illithid Harvester", gap221Illithid + "#1": "Plant Tadpoles",
		gap221MistDragon: "Juvenile Mist Dragon", gap221Ulder: "Duke Ulder Ravengard",
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

func TestGap221SpriteDragonGrowsOnNoncreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	dragon := effectstest.PushCatalogPermanent(g, me.ID, "Sprite Dragon", "Creature — Faerie Dragon", gap221Sprite, false)
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, dragon); n != 0 {
		t.Fatalf("%d counters after a creature spell, want 0", n)
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, dragon); n != 1 {
		t.Errorf("%d counters after a noncreature spell, want 1", n)
	}
	abilities := effectstest.EffectiveAbilities(t, g, dragon)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasAbility(abilities, kw) {
			t.Errorf("Sprite Dragon has no %s", kw)
		}
	}
}

func gap221MakeCommander(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).IsCommander = true })
}

func TestGap221GrayHarborMerfolkGrowsWithACommanderCreatureOrPlaneswalker(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		commander      bool
		want           int
	}{
		{"no commander", "Creature — Bear", false, 0},
		{"commander creature", "Creature — Bear", true, 2},
		{"commander planeswalker", "Legendary Planeswalker — Test", true, 2},
		{"commander artifact", "Legendary Artifact", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			other := effectstest.PushCatalogPermanent(g, me.ID, "Other", tc.typeLine, "", false)
			if tc.commander {
				gap221MakeCommander(g, other)
			}
			merfolk := effectstest.B12Push(g, me.ID, "Gray Harbor Merfolk", "Creature — Merfolk Rogue", gap221Merfolk, 0, 3)
			if p, _ := effectstest.PtOf(t, g, merfolk); p != tc.want {
				t.Errorf("power %d, want %d", p, tc.want)
			}
		})
	}
}

func TestGap221GrayHarborMerfolkCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merfolk := effectstest.B12Push(g, me.ID, "Gray Harbor Merfolk", "Creature — Merfolk Rogue", gap221Merfolk, 0, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.BrAttack(t, g, merfolk)
	if err := g.DeclareBlocker(blocker, merfolk); err == nil {
		t.Error("a creature blocked Gray Harbor Merfolk")
	}
}

func TestGap221DragonbornLooterLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	looter := effectstest.PushCatalogPermanent(g, me.ID, "Dragonborn Looter", "Creature — Dragon Rogue", gap221Looter, false)
	effectstest.ApaMana(me, "C")
	hand, gy := me.Hand.Size(), me.Graveyard.Size()
	if err := g.ActivateCatalogAbility(me.ID, looter, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() < hand {
		t.Errorf("hand %d → %d after loot", hand, me.Hand.Size())
	}
	if c, _ := g.LookupCardForEffect(looter); !c.Tapped {
		t.Error("Looter did not tap")
	}
	_ = gy
	// Tapped, it cannot activate again.
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, looter, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a tapped Looter activated")
	}
}

func gap221OtherOpponents(g *game.Game, me uuid.UUID) []*game.Player {
	var out []*game.Player
	for _, p := range g.Seats {
		if p.ID != me {
			out = append(out, p)
		}
	}
	return out
}

func TestGap221JuvenileMistDragonTapsAndFreezesOnePerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opps := gap221OtherOpponents(g, me.ID)
	a := effectstest.PushCreatureToBattlefieldForTest(g, opps[0].ID, "A's Bear")
	b := effectstest.PushCreatureToBattlefieldForTest(g, opps[1].ID, "B's Bear")
	mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "My Bear")
	effectstest.CastHoldCreature(t, g, "Juvenile Mist Dragon", "Creature — Dragon", gap221MistDragon, 4, 3)
	effectstest.AnswerPickTarget(t, g, a)
	effectstest.AnswerPickTarget(t, g, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := g.LookupCardForEffect(id); !c.Tapped {
			t.Errorf("%s did not tap", c.Name)
		}
	}
	if c, _ := g.LookupCardForEffect(mine); c.Tapped {
		t.Error("my own creature tapped")
	}
	// The next untap step of its controller skips it.
	effectstest.AdvanceToMainOf(t, g, 1)
	if c, _ := g.LookupCardForEffect(a); !c.Tapped {
		t.Error("A's Bear untapped during its controller's next untap step")
	}
}

func TestGap221AloraMakesAnAttackerUnblockableAndBouncesItAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Alora, Merry Thief", "Legendary Creature — Halfling Rogue", gap221Alora, 3, 2)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bear); err == nil {
		t.Error("the Bear was blocked after Alora's trigger")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear is still on the battlefield at the end step")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the Bear did not return to its owner's hand")
	}
}

func TestGap221DukeUlderGivesAnotherCreatureHasteAndMyriad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ulder := effectstest.B12Push(g, me.ID, "Duke Ulder Ravengard", "Legendary Creature — Human Noble Soldier", gap221Ulder, 5, 5)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	abilities := effectstest.EffectiveAbilities(t, g, bear)
	if !effectstest.HasAbility(abilities, "haste") {
		t.Errorf("the Bear has no haste: %v", abilities)
	}
	// Myriad is a granted trigger: attacking one opponent at a four-player
	// table offers a token copy attacking each of the other two.
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.InstanceID != bear && effects.IsToken(c) {
			copies++
		}
	}
	if copies != 2 {
		t.Errorf("%d myriad copies of the Bear, want 2", copies)
	}
	_ = ulder
}

func TestGap221AstralDragonCopiesANoncreaturePermanentTwiceAsDragons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Test Rock", "Artifact", "", false)
	effectstest.CastHoldCreature(t, g, "Astral Dragon", "Creature — Dragon", gap221Astral, 4, 4)
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if !c.IsToken() || c.Name != "Test Rock" {
			continue
		}
		n++
		p, tough := effectstest.PtOf(t, g, c.InstanceID)
		if !c.IsCreature() || p != 3 || tough != 3 {
			t.Errorf("token is creature=%v %d/%d, want a 3/3 creature", c.IsCreature(), p, tough)
		}
		if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, c.InstanceID), "flying") {
			t.Error("token has no flying")
		}
		if c.TypeLine == "" || !containsAll(c.TypeLine, "Artifact", "Creature", "Dragon") {
			t.Errorf("token type line %q lacks Artifact Creature Dragon", c.TypeLine)
		}
	}
	if n != 2 {
		t.Errorf("%d Test Rock tokens, want 2", n)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		found := false
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestGap221GaleOffersTheOtherTypeFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Gale, Waterdeep Prodigy", "Legendary Creature — Human Wizard", gap221Gale, false)
	sorcery, instant := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sorcery, Name: "Old Sorcery", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: instant, Name: "Old Instant", TypeLine: "Instant", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	effectstest.AnswerPickTarget(t, g, sorcery)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, sorcery)
}

func TestGap221GaleNeedsTheOtherTypeInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Gale, Waterdeep Prodigy", "Legendary Creature — Human Wizard", gap221Gale, false)
	instant := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: instant, Name: "Old Instant", TypeLine: "Instant", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("an instant spell offered an instant card from the graveyard")
	}
}

func TestGap221IllithidHarvesterTurnsTappedNontokenCreaturesFaceDownAs22Horrors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped Ogre", "Creature — Ogre", 5, 5)
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped Ogre", "Creature — Ogre", 5, 5)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true })
	effectstest.CastHoldCreature(t, g, "Illithid Harvester", "Creature — Horror", gap221Illithid, 4, 4)
	effectstest.AnswerPickTarget(t, g, tapped)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(tapped)
	if !c.FaceDown {
		t.Fatalf("%s was not turned face down", c.Name)
	}
	if p, tough := effectstest.PtOf(t, g, tapped); p != 2 || tough != 2 {
		t.Errorf("face-down creature is %d/%d, want 2/2", p, tough)
	}
	if u, _ := g.LookupCardForEffect(untapped); u.FaceDown {
		t.Error("an untapped creature was turned face down")
	}
	_ = me
}

func TestGap221PlantTadpolesTapsXCreaturesAndHoldsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Ogre", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Ogre", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Ogre", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Plant Tadpoles", TypeLine: "Sorcery — Adventure", OracleID: gap221Illithid + "#1",
		ManaCost: "{X}{U}{U}", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "U", "U", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, want := range []struct {
		id     uuid.UUID
		tapped bool
	}{{a, true}, {b, true}, {c, false}} {
		if got, _ := g.LookupCardForEffect(want.id); got.Tapped != want.tapped {
			t.Errorf("%s tapped=%v, want %v", got.Name, got.Tapped, want.tapped)
		}
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if got, _ := g.LookupCardForEffect(a); !got.Tapped {
		t.Error("A untapped during its controller's next untap step")
	}
}
