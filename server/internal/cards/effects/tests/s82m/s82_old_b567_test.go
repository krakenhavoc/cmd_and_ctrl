package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b567_test.go — Coldsnap, batch 567 (#5597): each test plays the
// real catalog entry through the real engine paths.

const (
	b567Scoundrel = "44c437ca-5d36-4bb3-9450-baf0c93ff4ce"
	b567Allosaur  = "176508ba-c107-4694-8dd7-ea34fca3b299"
	b567Chill     = "bcc53728-3f6e-48cf-a042-393a2d8098fc"
	b567Heidar    = "58ac4cce-506e-4d6e-a19c-0e875fbfcd81"
	b567RuneSnag  = "49bf3805-cc9e-4550-8a01-171282aa49c9"
	b567Cryomancr = "23c881f8-37c7-4899-890e-9f6e26f9f560"
	b567Sunscour  = "155f14dd-1516-4b9f-ae39-0a6172c69a66"
	b567Strider   = "f2074c37-6e76-4460-9666-010c7f38ae66"
	b567Furrier   = "3f086b72-6e7c-40d4-b995-d5f94c494462"
	b567Cryoclasm = "77c14618-ced1-4a46-a7fd-06cb2f7f38ff"

	b567Unsummon     = "837182db-1bf3-4a2c-bd01-1af9d9873561"
	b567TemporalAdpt = "fd9b5462-6578-4bf6-8026-23eaf5af3eee"
)

func TestS82OldB567CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b567Scoundrel: "Krovikan Scoundrel", b567Allosaur: "Allosaurus Rider", b567Chill: "Chill to the Bone",
		b567Heidar: "Heidar, Rimewind Master", b567RuneSnag: "Rune Snag", b567Cryomancr: "Rimewind Cryomancer",
		b567Sunscour: "Sunscour", b567Strider: "Karplusan Strider", b567Furrier: "Goblin Furrier", b567Cryoclasm: "Cryoclasm",
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

func b567SnowLand(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, "Snow-Covered "+subtype, "Basic Snow Land — "+subtype, "", 0, 0)
}

func b567Land(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, subtype, "Basic Land — "+subtype, "", 0, 0)
}

func TestS82OldB567AllosaurusRiderPitchesTwoGreenCardsAndCountsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		b567Land(g, me.ID, "Forest")
	}
	g1 := effectstest.HandCardFull(me, "Green One", "Instant", "{G}", "", []string{"G"})
	g2 := effectstest.HandCardFull(me, "Green Two", "Instant", "{G}", "", []string{"G"})
	red := effectstest.HandCardFull(me, "Red", "Instant", "{R}", "", []string{"R"})
	rider := effectstest.HandCardFull(me, "Allosaurus Rider", "Creature — Elf Warrior", "{5}{G}{G}", b567Allosaur, []string{"G"})
	pitch := func(ids ...uuid.UUID) error {
		return g.CastSpell(me.ID, rider, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: ids})
	}
	if err := pitch(g1, red); err == nil {
		t.Fatal("a red card was accepted as a green card")
	}
	if err := pitch(g1); err == nil {
		t.Fatal("one card is not two")
	}
	if err := pitch(g1, g2); err != nil {
		t.Fatalf("pitching two green cards: %v", err)
	}
	if me.Hand.Contains(g1) || me.Hand.Contains(g2) || !g.Exile.Contains(g1) || !g.Exile.Contains(g2) {
		t.Error("both green cards are exiled as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rider) {
		t.Fatal("the Rider did not resolve")
	}
	if p, tough := effectstest.PtOf(t, g, rider); p != 4 || tough != 4 {
		t.Errorf("with three lands %d/%d, want 4/4", p, tough)
	}
	b567Land(g, g.Seats[1].ID, "Island")
	b567Land(g, me.ID, "Forest")
	if p, tough := effectstest.PtOf(t, g, rider); p != 5 || tough != 5 {
		t.Errorf("with four lands of mine (and one of theirs) %d/%d, want 5/5", p, tough)
	}
}

func TestS82OldB567ChillToTheBoneDestroysOnlyNonsnowCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snow := effectstest.B16Creature(g, opp.ID, "Snow Bear", "Snow Creature — Bear", 2, 2, "G")
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "B", "B")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Chill to the Bone", "Instant", b567Chill, effectstest.CardTarget(snow)); err == nil {
		t.Fatal("a snow creature was accepted as the target")
	}
	effectstest.CastCatalogSpell(t, g, "Chill to the Bone", "Instant", b567Chill, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the nonsnow creature survived")
	}
	if !g.Battlefield.Contains(snow) {
		t.Error("the snow creature was destroyed")
	}
}

func TestS82OldB567HeidarNeedsFourSnowPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heidar := effectstest.B12Push(g, me.ID, "Heidar, Rimewind Master", "Legendary Creature — Human Wizard", b567Heidar, 3, 3)
	victim := effectstest.B16Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		b567SnowLand(g, me.ID, "Island")
	}
	b567SnowLand(g, opp.ID, "Island") // not mine: does not count
	effectstest.B06AddMana(me, "C", "C")
	params := game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)}
	if err := g.ActivateCatalogAbility(me.ID, heidar, 0, params); err == nil {
		t.Fatal("activated with three snow permanents")
	}
	b567SnowLand(g, me.ID, "Island")
	if err := g.ActivateCatalogAbility(me.ID, heidar, 0, params); err != nil {
		t.Fatalf("activated with four snow permanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, victim); z != game.ZoneHand {
		t.Errorf("the target is in %v, want its owner's hand", z)
	}
}

func TestS82OldB567RimewindCryomancerCountersAnActivatedAbilityOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cryo := effectstest.B12Push(g, me.ID, "Rimewind Cryomancer", "Creature — Human Wizard", b567Cryomancr, 2, 3)
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", effectstest.GoblinBombardmentOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	victim := effectstest.PushCatalogPermanent(g, opp.ID, "Victim", "Creature — Bear", "", false)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		b567SnowLand(g, me.ID, "Island")
	}
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.RtCardTarget(victim)}); err != nil {
		t.Fatalf("activate Goblin Bombardment: %v", err)
	}
	var ping uuid.UUID
	g.WithWriteLock(func() {
		for id, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemActivated && it.SourceCardID == bomb {
				ping = id
			}
		}
	})
	effectstest.B06AddMana(me, "C")
	params := game.ActivateAbilityParams{Strict: true, Targets: effectstest.RtCardTarget(ping)}
	if err := g.ActivateCatalogAbility(me.ID, cryo, 0, params); err == nil {
		t.Fatal("activated with three snow permanents")
	}
	b567SnowLand(g, me.ID, "Forest")
	bolt := effectstest.BoltFrom(t, g, me, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, cryo, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.RtCardTarget(bolt)}); err == nil {
		t.Fatal("a spell was accepted as the target of \"target activated ability\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, cryo, 0, params); err != nil {
		t.Fatalf("activated with four snow permanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) || effectstest.DamageOn(t, g, victim) != 0 {
		t.Error("the countered Goblin Bombardment still dealt its damage")
	}
}

func TestS82OldB567RuneSnagTaxGrowsWithEachRuneSnagInAnyGraveyard(t *testing.T) {
	for _, tc := range []struct {
		inGraveyards int
		want         string
	}{{0, "{2}"}, {1, "{4}"}, {3, "{8}"}} {
		g := effectstest.NewCatalogGame(t)
		caster, them := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		for i := 0; i < tc.inGraveyards; i++ {
			owner := caster
			if i%2 == 1 {
				owner = them
			}
			owner.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Rune Snag", TypeLine: "Instant", OracleID: b567RuneSnag, Owner: owner.ID, Controller: owner.ID})
		}
		victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
		spell := effectstest.HandCardFull(them, "Rune Snag", "Instant", "", b567RuneSnag, []string{"U"})
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority to the responder: %v", err)
		}
		if err := g.CastSpell(them.ID, spell, game.CastSpellParams{Targets: effectstest.CardRefs(victim)}); err != nil {
			t.Fatalf("casting Rune Snag: %v", err)
		}
		effectstest.PassUntilTaxed(t, g, caster.ID)
		prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
		if prompt.PayCost != tc.want {
			t.Errorf("%d in graveyards: the tax is %q, want %s", tc.inGraveyards, prompt.PayCost, tc.want)
		}
		if g.Stack.Contains(victim) {
			t.Error("a declined tax must counter the spell")
		}
	}
}

func TestS82OldB567SunscourPitchesTwoWhiteCardsAndDestroysEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2, "G")
	land := b567Land(g, opp.ID, "Island")
	effectstest.AdvanceToMain(t, g)
	w1 := effectstest.HandCardFull(me, "White One", "Instant", "{W}", "", []string{"W"})
	w2 := effectstest.HandCardFull(me, "White Two", "Instant", "{W}", "", []string{"W"})
	blue := effectstest.HandCardFull(me, "Blue", "Instant", "{U}", "", []string{"U"})
	sun := effectstest.HandCardFull(me, "Sunscour", "Sorcery", "{5}{W}{W}", b567Sunscour, []string{"W"})
	cast := func(ids ...uuid.UUID) error {
		return g.CastSpell(me.ID, sun, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: ids})
	}
	if err := cast(w1, blue); err == nil {
		t.Fatal("a blue card was accepted as a white card")
	}
	if err := cast(w1); err == nil {
		t.Fatal("one card is not two")
	}
	if err := cast(w1, w2); err != nil {
		t.Fatalf("pitching two white cards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a creature survived Sunscour")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was destroyed")
	}
}

func TestS82OldB567KarplusanStriderCantBeTargetedByBlueOrBlackSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	strider := effectstest.B12Push(g, opp.ID, "Karplusan Strider", "Creature — Yeti", b567Strider, 3, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "B", "B", "B", "B", "U", "R")
	blue := effectstest.HandCardFull(me, "Unsummon", "Instant", "{U}", b567Unsummon, []string{"U"})
	if err := g.CastSpell(me.ID, blue, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(strider)}); err == nil {
		t.Error("a blue spell was allowed to target the Strider")
	}
	black := effectstest.HandCardFull(me, "Chill to the Bone", "Instant", "{3}{B}", b567Chill, []string{"B"})
	if err := g.CastSpell(me.ID, black, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(strider)}); err == nil {
		t.Error("a black spell was allowed to target the Strider")
	}
	// A red spell and an ability from a blue source may.
	bolt := effectstest.HandCardFull(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle, []string{"R"})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(strider)}); err != nil {
		t.Errorf("a red spell targeting the Strider: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	adept := effectstest.B12Push(g, me.ID, "Temporal Adept", "Creature — Human Wizard", b567TemporalAdpt, 1, 3)
	effectstest.B06AddMana(me, "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, adept, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(strider)}); err != nil {
		t.Errorf("an ability from a blue source targeting the Strider: %v", err)
	}
}

func TestS82OldB567GoblinFurrierDealsNoDamageToSnowCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	furrier := effectstest.B12Push(g, me.ID, "Goblin Furrier", "Creature — Goblin Warrior", b567Furrier, 2, 2)
	snow := effectstest.B16Creature(g, opp.ID, "Snow Bear", "Snow Creature — Bear", 2, 5, "G")
	plain := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 5, "G")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(furrier, snow, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(furrier, plain, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if got := effectstest.DamageOn(t, g, snow); got != 0 {
		t.Errorf("the snow creature has %d damage, want 0", got)
	}
	if got := effectstest.DamageOn(t, g, plain); got != 2 {
		t.Errorf("the nonsnow creature has %d damage, want 2", got)
	}
	// Damage from another source is not prevented.
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(plain, snow, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if got := effectstest.DamageOn(t, g, snow); got != 2 {
		t.Errorf("another source dealt %d to the snow creature, want 2", got)
	}
	_ = me
}

func TestS82OldB567CryoclasmDestroysAPlainsOrIslandAndHurtsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := b567Land(g, opp.ID, "Forest")
	island := b567Land(g, opp.ID, "Island")
	plains := b567Land(g, me.ID, "Plains")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R", "R")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Cryoclasm", "Sorcery", b567Cryoclasm, effectstest.CardTarget(forest)); err == nil {
		t.Fatal("a Forest was accepted as the target")
	}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Cryoclasm", "Sorcery", b567Cryoclasm, effectstest.CardTarget(island))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(island) {
		t.Error("the Island survived")
	}
	if got := life - opp.Life; got != 3 {
		t.Errorf("the land's controller lost %d, want 3", got)
	}
	// Your own Plains is a legal target; its controller is you.
	life = me.Life
	effectstest.CastCatalogSpell(t, g, "Cryoclasm", "Sorcery", b567Cryoclasm, effectstest.CardTarget(plains))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(plains) || life-me.Life != 3 {
		t.Errorf("own Plains on battlefield %v, life lost %d, want it gone and 3", g.Battlefield.Contains(plains), life-me.Life)
	}
}
