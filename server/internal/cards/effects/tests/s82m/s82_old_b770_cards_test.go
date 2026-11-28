package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b770_cards_test.go — Tempest cards, S82 OLD batch 770 (#5628), played
// through the real catalog entries.

const (
	o770DauthiGhoul   = "5bf39f9e-4f06-4637-b0fc-5d21b926f6f4"
	o770Mistfolk      = "c836f8e4-6154-4c7a-a8dd-165ff74f9115"
	o770Mindripper    = "ecae5eba-8407-4c71-aeae-ab10fb1e5226"
	o770MetallicSlive = "2dfafd63-83fb-4783-8e12-30018680a905"
	o770Tinviper      = "b0bdc033-29f0-4dcb-87ff-ccd56126f7e5"
	o770Diver         = "4145c153-1ee3-4583-8a60-8471959656eb"
	o770SkyshroudElf  = "84a666a0-e453-4b34-815c-48dbfc9b4f4b"
	o770TradewindRide = "d0314da0-da24-4f6e-b1f6-0d50617bcaaa"
)

func TestS82OldB770CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o770DauthiGhoul: "Dauthi Ghoul", o770Mistfolk: "Thalakos Mistfolk", o770Mindripper: "Dauthi Mindripper",
		o770MetallicSlive: "Metallic Sliver", o770Tinviper: "Coiled Tinviper", o770Diver: "Rootwater Diver",
		o770SkyshroudElf: "Skyshroud Elf", o770TradewindRide: "Tradewind Rider",
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

func TestS82OldB770PrintedKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct{ oracle, name, typeLine, keyword string }{
		{o770DauthiGhoul, "Dauthi Ghoul", "Creature — Dauthi Zombie", "shadow"},
		{o770Mistfolk, "Thalakos Mistfolk", "Creature — Thalakos Illusion", "shadow"},
		{o770Mindripper, "Dauthi Mindripper", "Creature — Dauthi Minion", "shadow"},
		{o770Tinviper, "Coiled Tinviper", "Artifact Creature — Snake", "first strike"},
		{o770TradewindRide, "Tradewind Rider", "Creature — Spirit", "flying"},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		if !effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
			t.Errorf("%s does not have %s", tc.name, tc.keyword)
		}
	}
}

func TestS82OldB770DauthiGhoulGrowsWhenAnyShadowCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ghoul := effectstest.B12Push(g, me.ID, "Dauthi Ghoul", "Creature — Dauthi Zombie", o770DauthiGhoul, 1, 1)
	shadow := effectstest.B12Push(g, opp.ID, "Their Shade", "Creature — Dauthi Minion", o770Mindripper, 2, 1)
	plain := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	counters := func() int {
		c, _ := g.LookupCardForEffect(ghoul)
		return c.Counters[game.CounterPlusOne]
	}
	effectstest.B14Kill(g, plain)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 0 {
		t.Errorf("a creature without shadow dying gave %d counters", counters())
	}
	effectstest.B14Kill(g, shadow)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 1 {
		t.Errorf("a creature with shadow dying gave %d counters, want 1", counters())
	}
	p, tough := effectstest.PtOf(t, g, ghoul)
	if p != 2 || tough != 2 {
		t.Errorf("the Ghoul is %d/%d, want 2/2", p, tough)
	}
}

func TestS82OldB770ThalakosMistfolkGoesOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mist := effectstest.B12Push(g, me.ID, "Thalakos Mistfolk", "Creature — Thalakos Illusion", o770Mistfolk, 2, 1)
	effectstest.AddPool(t, g, me.ID, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, mist, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mist) || !me.Library.Contains(mist) {
		t.Fatal("the Mistfolk is not in its owner's library")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != mist {
		t.Error("the Mistfolk is not on top of the library")
	}
}

func TestS82OldB770DauthiMindripperSacrificesToMakeTheDefenderDiscardThree(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		opp.Hand.Cards = nil
		var hand []uuid.UUID
		for i := 0; i < 5; i++ {
			hand = append(hand, effectstest.HandSpell(opp, "Card", "Sorcery", ""))
		}
		rip := effectstest.B12Push(g, me.ID, "Dauthi Mindripper", "Creature — Dauthi Minion", o770Mindripper, 2, 1)
		effectstest.DeclareAttack(t, g, opp.ID, rip)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.ConfirmChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("the controller was not asked whether to sacrifice")
		}
		if err := g.ResolveConfirm(c.ID, me.ID, accept); err != nil {
			t.Fatalf("answer: %v", err)
		}
		if !accept {
			if !g.Battlefield.Contains(rip) || opp.Hand.Size() != 5 {
				t.Error("declining sacrificed the creature or made a player discard")
			}
			continue
		}
		if g.Battlefield.Contains(rip) {
			t.Error("the Mindripper was not sacrificed")
		}
		effectstest.AnswerDiscard(t, g, opp.ID, hand[0], hand[1], hand[2])
		if opp.Hand.Size() != 2 {
			t.Errorf("the defender has %d cards, want 2", opp.Hand.Size())
		}
	}
}

func TestS82OldB770RootwaterDiverReturnsAnArtifactCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	diver := effectstest.B12Push(g, me.ID, "Rootwater Diver", "Creature — Merfolk", o770Diver, 1, 1)
	art := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: art, Name: "Old Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, diver, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("a nonartifact card was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, diver, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(art)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(diver) {
		t.Error("the Diver was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(art) || me.Graveyard.Contains(art) {
		t.Error("the artifact card did not return to hand")
	}
}

func TestS82OldB770SkyshroudElfMakesGreenFromATapAndRedOrWhiteFromOneMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	elf := effectstest.B12Push(g, me.ID, "Skyshroud Elf", "Creature — Elf Druid", o770SkyshroudElf, 1, 1)
	spec, _ := effects.Lookup(o770SkyshroudElf)
	if len(spec.ManaAbilities) != 2 || spec.ManaAbilities[0].Produced != "{G}" || spec.ManaAbilities[1].Produced != "{R|W}" ||
		!spec.ManaAbilities[0].Cost.Tap || spec.ManaAbilities[1].Cost.Mana != "{1}" || spec.ManaAbilities[1].Cost.Tap {
		t.Fatalf("mana abilities = %+v", spec.ManaAbilities)
	}
	effectstest.ActivateManaFor(t, g, me.ID, elf, 0, game.ManaAbilityParams{})
	if pool := effectstest.SortedPool(me); len(pool) != 1 || pool[0] != "G" {
		t.Fatalf("the tap ability left %v in the pool, want [G]", pool)
	}
	if !effectstest.TappedForTest(t, g, elf) {
		t.Error("the tap ability did not tap the Elf")
	}
	// The {1} ability works while the Elf is tapped, and makes red or white only.
	me.ManaPool = game.ManaPool{}
	effectstest.AddPool(t, g, me.ID, "{C}")
	if err := g.ActivateManaAbility(me.ID, elf, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("the {1} ability made green")
	}
	effectstest.ActivateManaFor(t, g, me.ID, elf, 1, game.ManaAbilityParams{Colors: []string{"R"}})
	if pool := effectstest.SortedPool(me); len(pool) != 1 || pool[0] != "R" {
		t.Errorf("the {1} ability left %v in the pool, want [R]", pool)
	}
}

func TestS82OldB770TradewindRiderTapsTwoOthersToBounceAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rider := effectstest.B12Push(g, me.ID, "Tradewind Rider", "Creature — Spirit", o770TradewindRide, 1, 4)
	a := effectstest.B12Creature(g, me.ID, "Helper A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Helper B", "Creature — Bear", 1, 1)
	target := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	// One helper is not enough.
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(target), TapIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("the ability was activated with one creature tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, rider, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(target), TapIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range []uuid.UUID{rider, a, b} {
		if !effectstest.TappedForTest(t, g, id) {
			t.Errorf("%s is untapped", id)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) || !opp.Hand.Contains(target) {
		t.Error("the target did not return to its owner's hand")
	}
}
