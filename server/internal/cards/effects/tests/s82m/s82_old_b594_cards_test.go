package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o594Sakiko   = "18a98d44-54ee-4d6d-ab32-27e294ac8995"
	o594Realm    = "1da3a0c0-43f0-4feb-b20a-414b023e710c"
	o594Oni      = "4ce00fdf-c4da-492a-81c9-8c93dd0bf620"
	o594Fields   = "5cd35d18-115b-4d7c-8d05-3d4c8303f46e"
	o594Heart    = "634d0256-849d-4ef7-b070-b6b8f3242e00"
	o594Threads  = "9d86735c-12e5-457a-beb4-057ab220d18b"
	o594Wings    = "9da74f0f-143d-4db2-81eb-c7873744a2d0"
	o594Cedars   = "ada3c285-3449-4037-8b44-42e128456a65"
	o594Spires   = "c040882e-48ba-469e-9fc6-02e8b68dd58c"
	o594Falls    = "c546747a-1935-4c59-a45f-cad7da41fe38"
	o594Fens     = "f6c9c0e6-b851-4ede-8d7a-190c7cfd9051"
	o594AuraLine = "Enchantment — Aura"
)

func TestO594CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o594Sakiko: "Mark of Sakiko", o594Realm: "Genju of the Realm", o594Oni: "Mark of the Oni", o594Fields: "Genju of the Fields",
		o594Heart: "Heart of Light", o594Threads: "Threads of Disloyalty", o594Wings: "Phantom Wings", o594Cedars: "Genju of the Cedars",
		o594Spires: "Genju of the Spires", o594Falls: "Genju of the Falls", o594Fens: "Genju of the Fens",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func o594Controller(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	return effectstest.GaLayered(t, g, id).Controller
}

func TestO594ThreadsStealsOnlyManaValueTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := effectstest.ApcB1Push(g, opp.ID, "Cheap", "Creature — Bear", "", "{1}{G}", 2, 2)
	dear := effectstest.ApcB1Push(g, opp.ID, "Dear", "Creature — Bear", "", "{2}{G}", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Threads of Disloyalty", o594AuraLine, o594Threads, effectstest.CardRefs(dear)); err == nil {
		t.Fatal("Threads took a creature with mana value 3")
	}
	effectstest.AuraCast(t, g, "Threads of Disloyalty", o594Threads, cheap)
	if got := o594Controller(t, g, cheap); got != me.ID {
		t.Errorf("the enchanted creature is controlled by %v, want the caster", got)
	}
	if got := o594Controller(t, g, dear); got != opp.ID {
		t.Error("the other creature changed hands")
	}
}

func TestO594MarkOfTheOniSacrificesItselfWithoutADemon(t *testing.T) {
	for _, demon := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		aura := effectstest.AuraCast(t, g, "Mark of the Oni", o594Oni, victim)
		if got := o594Controller(t, g, victim); got != me.ID {
			t.Fatalf("demon=%v: control %v, want the caster", demon, got)
		}
		if demon {
			effectstest.B12Creature(g, me.ID, "Imp", "Creature — Demon", 1, 1)
		}
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if gone := !g.Battlefield.Contains(aura); gone == demon {
			t.Errorf("demon=%v: the Aura left the battlefield = %v", demon, gone)
		}
		if got := o594Controller(t, g, victim); (got == opp.ID) == demon {
			t.Errorf("demon=%v: the creature's controller is %v", demon, got)
		}
	}
}

func TestO594HeartOfLightPreventsDamageBothWays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	blk := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.AuraCast(t, g, "Heart of Light", o594Heart, atk)
	effectstest.BrAttack(t, g, atk)
	if err := g.DeclareBlocker(blk, atk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, blk); got != 0 {
		t.Errorf("the blocker was dealt %d damage by the enchanted attacker", got)
	}
	if got := effectstest.Pr6Marked(g, atk); got != 0 {
		t.Errorf("the enchanted attacker was dealt %d damage", got)
	}
	if !g.Battlefield.Contains(blk) || !g.Battlefield.Contains(atk) {
		t.Error("a creature died in a fight Heart of Light should have made harmless")
	}
}

func TestO594PhantomWingsGivesFlyingAndBouncesTheHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	aura := effectstest.AuraCast(t, g, "Phantom Wings", o594Wings, host)
	effectstest.AssertKeywords(t, g, host, "flying")
	effectstest.B16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed")
	}
	if g.Battlefield.Contains(host) || !effectstest.HandHas(me, host) {
		t.Error("the enchanted creature was not returned to its owner's hand")
	}
}

func TestO594MarkOfSakikoAddsKeptGreenManaOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.AuraCast(t, g, "Mark of Sakiko", o594Sakiko, atk)
	effectstest.BrAttack(t, g, atk)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if got := effectstest.SortedPool(me); !slices.Equal(got, []string{"G", "G", "G"}) {
		t.Fatalf("pool after combat damage: %v, want GGG", got)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if got := effectstest.SortedPool(me); !slices.Equal(got, []string{"G", "G", "G"}) {
		t.Errorf("the mana did not survive to the postcombat main: %v", got)
	}
}

type o594Genju struct {
	name, oracle, landType, landLine string
	power, toughness                 int
	color                            string
}

func o594Setup(t *testing.T, gj o594Genju) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushCatalogPermanent(g, me.ID, gj.landType, gj.landLine, "", false)
	effectstest.AdvanceToMain(t, g)
	aura := effectstest.AuraCast(t, g, gj.name, gj.oracle, land)
	if !effectstest.AttachedTo(t, g, aura, land) {
		t.Fatalf("%s did not attach to the %s", gj.name, gj.landType)
	}
	return g, me, aura, land
}

func o594Animate(t *testing.T, g *game.Game, me *game.Player, aura uuid.UUID) {
	t.Helper()
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestO594GenjuBecomeTheirCreatures(t *testing.T) {
	for _, gj := range []o594Genju{
		{"Genju of the Falls", o594Falls, "Island", "Basic Land — Island", 3, 2, "U"},
		{"Genju of the Spires", o594Spires, "Mountain", "Basic Land — Mountain", 6, 1, "R"},
		{"Genju of the Cedars", o594Cedars, "Forest", "Basic Land — Forest", 4, 4, "G"},
		{"Genju of the Fields", o594Fields, "Plains", "Basic Land — Plains", 2, 5, "W"},
		{"Genju of the Fens", o594Fens, "Swamp", "Basic Land — Swamp", 2, 2, "B"},
	} {
		t.Run(gj.name, func(t *testing.T) {
			g, me, aura, land := o594Setup(t, gj)
			if effectstest.GaLayered(t, g, land).IsCreature() {
				t.Fatal("the land was a creature before the ability")
			}
			o594Animate(t, g, me, aura)
			c := effectstest.GaLayered(t, g, land)
			if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Spirit") || !c.HasSubtype(gj.landType) {
				t.Errorf("the land is %q, want a Spirit creature that is still a %s land", c.TypeLine, gj.landType)
			}
			if !slices.Equal(c.EffectiveColors(), []string{gj.color}) {
				t.Errorf("colours %v, want [%s]", c.EffectiveColors(), gj.color)
			}
			if p, tt := effectstest.EffectivePower(t, g, land), effectstest.EffectiveToughness(t, g, land); p != gj.power || tt != gj.toughness {
				t.Errorf("%d/%d, want %d/%d", p, tt, gj.power, gj.toughness)
			}
			if gj.oracle == o594Falls {
				effectstest.AssertKeywords(t, g, land, "flying")
			}
			effectstest.AdvanceToEndStep(t, g)
			effectstest.AdvanceTo(t, g, game.StepCleanup)
			if effectstest.GaLayered(t, g, land).IsCreature() {
				t.Error("the land was still a creature after the turn ended")
			}
		})
	}
}

func TestO594GenjuOfTheRealmMakesALegendaryTramplingLand(t *testing.T) {
	g, me, aura, land := o594Setup(t, o594Genju{"Genju of the Realm", o594Realm, "Swamp", "Basic Land — Swamp", 8, 12, ""})
	o594Animate(t, g, me, aura)
	c := effectstest.GaLayered(t, g, land)
	if !c.IsCreature() || !c.IsLand() || !c.IsLegendary() || !c.HasSubtype("Spirit") {
		t.Errorf("the land is %q, want a legendary Spirit creature land", c.TypeLine)
	}
	if p, tt := effectstest.EffectivePower(t, g, land), effectstest.EffectiveToughness(t, g, land); p != 8 || tt != 12 {
		t.Errorf("%d/%d, want 8/12", p, tt)
	}
	effectstest.AssertKeywords(t, g, land, "trample")
}

func TestO594GenjuOfTheFensPumpsForB(t *testing.T) {
	g, me, aura, land := o594Setup(t, o594Genju{"Genju of the Fens", o594Fens, "Swamp", "Basic Land — Swamp", 2, 2, "B"})
	o594Animate(t, g, me, aura)
	idx, ref := effectstest.BeGrantedRow(t, g, land, "{B}: This creature gets +1/+1 until end of turn.")
	if idx < 0 {
		t.Fatal("the animated Swamp has no granted pump row")
	}
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, land); p != 3 {
		t.Errorf("power %d after one pump, want 3", p)
	}
}

func TestO594GenjuOfTheFieldsGainsLifeForDamageDealt(t *testing.T) {
	g, me, aura, land := o594Setup(t, o594Genju{"Genju of the Fields", o594Fields, "Plains", "Basic Land — Plains", 2, 5, "W"})
	opp := g.Seats[1]
	o594Animate(t, g, me, aura)
	life := me.Life
	effectstest.BrAttack(t, g, land)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if opp.Life != 38 {
		t.Fatalf("the opponent is at %d, want 38 after 2 damage", opp.Life)
	}
	if me.Life-life != 2 {
		t.Errorf("gained %d life, want 2", me.Life-life)
	}
}

func TestO594GenjuReturnsFromTheGraveyardWhenItsLandIsPutThere(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g, me, aura, land := o594Setup(t, o594Genju{"Genju of the Spires", o594Spires, "Mountain", "Basic Land — Mountain", 6, 1, "R"})
		effectstest.Destroy(t, g, land)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HandHas(me, aura); got != yes {
			t.Errorf("yes=%v: Genju in hand = %v (graveyard %v, battlefield %v, stack empty %v)", yes, got, effectstest.GraveyardHas(me, aura), g.Battlefield.Contains(aura), effectstest.StackFullyEmpty(g))
		}
		if !yes && !effectstest.GraveyardHas(me, aura) {
			t.Error("declining left the Genju somewhere other than the graveyard")
		}
	}
}
