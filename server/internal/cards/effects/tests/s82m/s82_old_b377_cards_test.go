package testss82m

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b377_cards_test.go — Born of the Gods, batch 377 (S82 older sets,
// #5559): five instants and four bestow creatures. Every test plays the real
// catalog entry through the engine.

const (
	oldB377Unravel    = "128e0bd1-ed1d-4f53-8dae-83ee95f6fa2a"
	oldB377Crypsis    = "6077e650-32f9-4954-8286-9c8a3f3227ae"
	oldB377Glimpse    = "87cc3a67-b4f0-41f0-9e34-745ef23aa479"
	oldB377Infatuate  = "ae771e78-1d29-4b8c-8511-9da196a78dcf"
	oldB377Hydra      = "d1580692-ca1c-4682-8cad-7ad380b39454"
	oldB377Everflame  = "2c6b2f4a-4e0e-4dc1-92a8-c07e5cbba5f4"
	oldB377Herald     = "3c04aeda-ae09-4162-ab51-983e29d21225"
	oldB377Flitterstp = "4280191f-e2cb-4cee-a3b4-660b3efd62e9"
	oldB377Spiteful   = "e7e60afe-5ddf-406b-86e1-e82e2a660432"

	oldB377Bear = "Creature — Bear"
	oldB377Ench = "Enchantment Creature — Spirit"
	oldB377LibN = 3
)

var oldB377Cards = map[string]string{
	oldB377Unravel: "Unravel the Aether", oldB377Crypsis: "Crypsis", oldB377Glimpse: "Glimpse the Sun God",
	oldB377Infatuate: "Fated Infatuation", oldB377Hydra: "Aspect of Hydra", oldB377Everflame: "Everflame Eidolon",
	oldB377Herald: "Herald of Torment", oldB377Flitterstp: "Flitterstep Eidolon", oldB377Spiteful: "Spiteful Returned",
}

func oldB377Library(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

func TestOldB377EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB377Cards {
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

func TestOldB377UnravelTheAetherShufflesIntoTheOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	relic := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Relic", TypeLine: "Artifact"})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB377Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Unravel the Aether", "Instant", oldB377Unravel, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Unravel the Aether targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Unravel the Aether", "Instant", oldB377Unravel, effectstest.CardRefs(relic))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || !opp.Library.Contains(relic) {
		t.Errorf("the artifact is on the battlefield %v, in its owner's library %v", g.Battlefield.Contains(relic), opp.Library.Contains(relic))
	}
}

func TestOldB377CrypsisProtectsAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB377Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB377Bear, 2, 2)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, mine).Tapped = true })
	if err := effectstest.CastCatalogSpellErr(t, g, "Crypsis", "Instant", oldB377Crypsis, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("Crypsis targeted an opponent's creature")
	}
	effectstest.CastCatalogSpell(t, g, "Crypsis", "Instant", oldB377Crypsis, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("the creature is still tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, game.ProtectionFromOpponentsCreatures) {
		t.Error("the creature has no protection from creatures your opponents control")
	}
}

func TestOldB377GlimpseTapsXCreaturesAndScrys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB377Library(me, oldB377LibN)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", oldB377Bear, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", oldB377Bear, 2, 2)
	spare := effectstest.B12Creature(g, opp.ID, "Spare", oldB377Bear, 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Glimpse the Sun God", "Instant", "{X}{W}", oldB377Glimpse, game.CastSpellParams{XValue: 2, Targets: effectstest.CardTargets(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, a) || !effectstest.B16Tapped(t, g, b) || effectstest.B16Tapped(t, g, spare) {
		t.Errorf("tapped = %v %v %v, want true true false", effectstest.B16Tapped(t, g, a), effectstest.B16Tapped(t, g, b), effectstest.B16Tapped(t, g, spare))
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("Glimpse the Sun God did not scry")
	}
}

func TestOldB377FatedInfatuationCopiesAndScrysOnlyOnYourTurn(t *testing.T) {
	for _, mine := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		caster := me
		if !mine {
			caster = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		}
		oldB377Library(caster, oldB377LibN)
		src := effectstest.B12Creature(g, caster.ID, "Bear", oldB377Bear, 2, 2)
		effectstest.CastFromHandForTest(t, g, caster, "Fated Infatuation", "Instant", "{U}{U}{U}", oldB377Infatuate, game.CastSpellParams{Targets: effectstest.CardTargets(src)})
		effectstest.PassPriorityAroundTable(t, g)
		n := 0
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Controller == caster.ID && c.Name == "Bear" {
					n++
				}
			}
		})
		if n != 2 {
			t.Errorf("your turn=%v: %d Bears, want 2", mine, n)
		}
		if got := effectstest.ScryChoiceFor(g, caster.ID) != nil; got != mine {
			t.Errorf("your turn=%v: scry prompt = %v", mine, got)
		}
	}
}

func TestOldB377AspectOfHydraCountsGreenDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Green Thing", TypeLine: "Creature — Elf", ManaCost: "{1}{G}{G}", Colors: []string{"G"}, Power: 1, Toughness: 1})
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Blue Thing", TypeLine: "Creature — Wizard", ManaCost: "{U}{U}", Power: 1, Toughness: 1})
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB377Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Aspect of Hydra", "Instant", oldB377Hydra, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the bear is %d/%d, want 4/4 (devotion 2)", p, tough)
	}
}

func TestOldB377BestowedEidolonsPumpAndGrantTheirHost(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		wantP, wantT int
		keyword      string
		unblockable  bool
	}{
		{"Everflame Eidolon", oldB377Everflame, 1, 1, 3, 3, "", false},
		{"Herald of Torment", oldB377Herald, 3, 3, 5, 5, "flying", false},
		{"Flitterstep Eidolon", oldB377Flitterstp, 1, 1, 3, 3, "", true},
		{"Spiteful Returned", oldB377Spiteful, 1, 1, 3, 3, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			card := effectstest.BestowCard(t, g, tc.name, oldB377Ench, tc.oracle, tc.power, tc.tough)
			bear := effectstest.B12Creature(g, me.ID, "Bear", oldB377Bear, 2, 2)
			effectstest.CastBestowed(t, g, card, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if s := effectstest.ReadBestowState(g, card); !s.OnBattlefield || s.AttachedTo != bear || s.Creature {
				t.Fatalf("bestowed %s is not an Aura on the bear: %+v", tc.name, s)
			}
			if p, tough := effectstest.PtOf(t, g, bear); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the bear is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
			if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the bear has no %s", tc.keyword)
			}
			if tc.unblockable {
				effectstest.AssertRestrictions(t, g, bear, game.CantBeBlocked)
			}
		})
	}
}

func TestOldB377FlitterstepEidolonIsUnblockableAsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Flitterstep Eidolon", oldB377Ench, oldB377Flitterstp, 1, 1)
	effectstest.AssertRestrictions(t, g, id, game.CantBeBlocked)
	blocker := effectstest.B12Creature(g, opp.ID, "Bear", oldB377Bear, 2, 2)
	if err := effectstest.PorBlock(t, g, id, opp.ID, blocker); err == nil {
		t.Error("Flitterstep Eidolon was blocked")
	}
}

func TestOldB377EverflameEidolonPumpsItselfOrItsHost(t *testing.T) {
	// As a creature it pumps itself.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Everflame Eidolon", oldB377Ench, oldB377Everflame, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, id); p != 2 {
		t.Errorf("the Eidolon has power %d, want 2", p)
	}

	// Bestowed it pumps the enchanted creature instead.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	card := effectstest.BestowCard(t, g, "Everflame Eidolon", oldB377Ench, oldB377Everflame, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB377Bear, 2, 2)
	effectstest.CastBestowed(t, g, card, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate bestowed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 4/3 (+1/+1 and +1/+0)", p, tough)
	}
}

func TestOldB377HeraldOfTormentCostsALifeEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Herald of Torment", oldB377Ench, oldB377Herald, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID, "flying") {
		t.Error("Herald of Torment has no flying")
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+len(g.Seats)-1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life-1 {
		t.Errorf("life %d → %d, want -1", life, got)
	}
}

func TestOldB377SpitefulReturnedDrainsTheDefenderWhenItOrItsHostAttacks(t *testing.T) {
	// The creature itself attacks.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.B12Push(g, me.ID, "Spiteful Returned", "Enchantment Creature — Zombie", oldB377Spiteful, 1, 1)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.AttackWithTrigger(t, g, id, opp.ID, uuid.Nil)
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("the creature attacked: defender life %d → %d, want -2", life, got)
	}

	// Bestowed, its host attacks.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	card := effectstest.BestowCard(t, g, "Spiteful Returned", "Enchantment Creature — Zombie", oldB377Spiteful, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB377Bear, 2, 2)
	effectstest.CastBestowed(t, g, card, bear)
	effectstest.PassPriorityAroundTable(t, g)
	life = effectstest.LifeOf(g, opp.ID)
	effectstest.AttackWithTrigger(t, g, bear, opp.ID, uuid.Nil)
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("the host attacked: defender life %d → %d, want -2", life, got)
	}
}
