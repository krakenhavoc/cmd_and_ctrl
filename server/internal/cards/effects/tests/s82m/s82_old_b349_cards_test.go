package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b349_cards_test.go — Khans of Tarkir raid, batch 349 (#5554). Each
// test plays the real catalog entry and asserts what its text says.

const (
	oldB349Skullhunter = "2a87e3a9-10e3-4b00-aeb9-d3336b2fa3b3"
	oldB349Howl        = "76d9c310-9d06-43e8-87a9-fb03318a6cb8"
	oldB349Hordemate   = "9ace6510-7876-48e4-9cae-1c24f2229e03"
	oldB349Saddlebrute = "e1e5ad8c-d425-44ad-a8f6-e9290667f6f5"
	oldB349Hordechief  = "f9ea5b9e-a23d-48cf-8601-56e79ca844f9"
	oldB349Warshrieker = "fe9e9e2e-f304-4da3-a257-6886b0b1896b"
)

func TestOldB349CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB349Skullhunter: "Mardu Skullhunter", oldB349Howl: "Howl of the Horde", oldB349Hordemate: "Timely Hordemate",
		oldB349Saddlebrute: "Bellowing Saddlebrute", oldB349Hordechief: "Mardu Hordechief", oldB349Warshrieker: "Mardu Warshrieker",
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

func TestOldB349SkullhunterEntersTappedAndDiscardsOnlyAfterRaid(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB223Raid(t, g, attacked)
		card := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: card, Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		hand := opp.Hand.Size()
		id := effectstest.CastAndResolveCreature(t, g, "Mardu Skullhunter", "Creature — Human Warrior", oldB349Skullhunter)
		if attacked {
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickPlayer(t, g, me.ID, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
				effectstest.AnswerDiscard(t, g, opp.ID, card)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.PermanentTapped(g, id) {
			t.Errorf("attacked=%v: the Skullhunter entered untapped", attacked)
		}
		want := hand
		if attacked {
			want = hand - 1
		}
		if opp.Hand.Size() != want {
			t.Errorf("attacked=%v: opponent hand %d, want %d", attacked, opp.Hand.Size(), want)
		}
	}
}

func TestOldB349HordemateReturnsASmallCreatureOnlyAfterRaid(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB223Raid(t, g, attacked)
		small := effectstest.PushGraveyardPermanent(me, "Small Creature", "Creature — Bear", "{1}{G}")
		big := effectstest.PushGraveyardPermanent(me, "Big Creature", "Creature — Bear", "{2}{G}")
		effectstest.CastAndResolveCreature(t, g, "Timely Hordemate", "Creature — Human Warrior", oldB349Hordemate)
		if attacked {
			effectstest.B04WaitForPick(t, g, me.ID)
			if effectstest.HasID(effectstest.LatestPickTarget(g, me.ID).PickTargetCards, big) {
				t.Error("a mana value 3 creature card was offered")
			}
			effectstest.PickCard(t, g, me.ID, small)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.FindBattlefieldCardByID(g, small) != nil; got != attacked {
			t.Errorf("attacked=%v: the small creature is on the battlefield = %v", attacked, got)
		}
		if effectstest.FindBattlefieldCardByID(g, big) != nil {
			t.Errorf("attacked=%v: the mana value 3 creature came back", attacked)
		}
	}
}

func TestOldB349SaddlebruteCostsFourLifeUnlessYouAttacked(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB223Raid(t, g, attacked)
		life := me.Life
		effectstest.CastAndResolveCreature(t, g, "Bellowing Saddlebrute", "Creature — Orc Warrior", oldB349Saddlebrute)
		effectstest.PassPriorityAroundTable(t, g)
		want := life - 4
		if attacked {
			want = life
		}
		if me.Life != want {
			t.Errorf("attacked=%v: life %d, want %d", attacked, me.Life, want)
		}
	}
}

func TestOldB349HordechiefMakesAWarriorOnlyAfterRaid(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB223Raid(t, g, attacked)
		effectstest.CastAndResolveCreature(t, g, "Mardu Hordechief", "Creature — Human Warrior", oldB349Hordechief)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if attacked {
			want = 1
		}
		if got := len(effectstest.AvB34Tokens(g, me.ID, "Warrior")); got != want {
			t.Errorf("attacked=%v: %d Warrior tokens, want %d", attacked, got, want)
		}
	}
}

func TestOldB349WarshriekerAddsThreeColorsOnlyAfterRaid(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB223Raid(t, g, attacked)
		effectstest.CastAndResolveCreature(t, g, "Mardu Warshrieker", "Creature — Orc Shaman", oldB349Warshrieker)
		effectstest.PassPriorityAroundTable(t, g)
		colours := effectstest.PoolColours(me)
		if attacked {
			if colours["R"] != 1 || colours["W"] != 1 || colours["B"] != 1 {
				t.Errorf("pool %v, want one each of R, W and B", colours)
			}
		} else if effectstest.PoolSize(g, me.ID) != 0 {
			t.Errorf("pool %v after no attack, want empty", colours)
		}
	}
}

func TestOldB349HowlCopiesTheNextInstantOrSorceryOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB223Raid(t, g, false)
	effectstest.CastCatalogSpell(t, g, "Howl of the Horde", "Sorcery", oldB349Howl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("Howl left %d delayed triggers with no attack, want 1", len(g.DelayedTriggers))
	}
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 1)
	if got := effectstest.LifeOf(g, opp.ID); got != before-6 {
		t.Errorf("opponent life = %d, want %d (the Bolt and one copy)", got, before-6)
	}
}

func TestOldB349HowlWithRaidCopiesAnAdditionalTime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB223Raid(t, g, true)
	effectstest.CastCatalogSpell(t, g, "Howl of the Horde", "Sorcery", oldB349Howl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("Howl left %d delayed triggers after an attack, want 2", len(g.DelayedTriggers))
	}
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 2)
	if got := effectstest.LifeOf(g, opp.ID); got != before-9 {
		t.Errorf("opponent life = %d, want %d (the Bolt and two copies)", got, before-9)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("both triggers are for the next spell only; %d left", len(g.DelayedTriggers))
	}
}
