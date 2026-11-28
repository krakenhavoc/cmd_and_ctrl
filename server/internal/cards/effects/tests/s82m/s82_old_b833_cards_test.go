package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB833BlinkingSpirit    = "a26f6958-9bc8-46f8-8ca7-7f55098e1291"
	oldB833ElderDruid        = "a440a4f7-843f-4186-8ca9-df99f6aa321e"
	oldB833WhaleboneGlider   = "a65fadeb-c9bc-409f-8e90-ff8e55485a4c"
	oldB833ArcumsWeathervane = "b4e57fba-5fa3-4fd5-8f0a-7b61593ade72"
	oldB833FyndhornBow       = "d37fb8bf-d293-4ac3-b744-74e4caade975"
	oldB833ElvishHealer      = "ea46835c-9dac-4e1e-8338-ad99136f511a"
	oldB833ZuranEnchanter    = "ee0f883f-d7c9-4acf-a78f-f733b6f268d3"
	oldB833BarbarianGuides   = "ef8c324a-582b-401e-add1-1985c5596908"
)

var oldB833Cards = map[string]string{
	oldB833BlinkingSpirit: "Blinking Spirit", oldB833ElderDruid: "Elder Druid",
	oldB833WhaleboneGlider: "Whalebone Glider", oldB833ArcumsWeathervane: "Arcum's Weathervane",
	oldB833FyndhornBow: "Fyndhorn Bow", oldB833ElvishHealer: "Elvish Healer",
	oldB833ZuranEnchanter: "Zuran Enchanter", oldB833BarbarianGuides: "Barbarian Guides",
}

func TestOldB833EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB833Cards {
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

func oldB833Permanent(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func oldB833InHand(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func TestOldB833BlinkingSpiritReturnsToHandForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sp := oldB833Permanent(g, me.ID, "Blinking Spirit", "Creature — Spirit", oldB833BlinkingSpirit, 2, 2)
	effectstest.B16Activate(t, g, me.ID, sp, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB833InHand(me, "Blinking Spirit") {
		t.Error("the Spirit did not return to hand")
	}
}

func TestOldB833FyndhornBowGivesFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bow := oldB833Permanent(g, me.ID, "Fyndhorn Bow", "Artifact", oldB833FyndhornBow, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{G}{G}{G}")
	effectstest.B16Activate(t, g, me.ID, bow, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the target did not gain first strike")
	}
}

func TestOldB833WhalebonGliderNeedsPowerThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	glider := oldB833Permanent(g, me.ID, "Whalebone Glider", "Artifact", oldB833WhaleboneGlider, 0, 0)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 4, 4)
	effectstest.FloatMana(t, g, me, "{G}{G}")
	err := g.ActivateCatalogAbility(me.ID, glider, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	})
	if err == nil {
		t.Fatal("a power 4 creature was a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, glider, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, small, "flying") {
		t.Error("the target did not gain flying")
	}
}

func TestOldB833ArcumsWeathervaneTogglesSnow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vane := oldB833Permanent(g, me.ID, "Arcum's Weathervane", "Artifact", oldB833ArcumsWeathervane, 0, 0)
	snow := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest"})
	plain := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	if !effectstest.LayeredCard(t, g, snow).HasSupertype("snow") {
		t.Fatal("the test land is not snow")
	}
	effectstest.FloatMana(t, g, me, "{G}{G}")
	err := g.ActivateCatalogAbility(me.ID, vane, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	})
	if err == nil {
		t.Fatal("a nonsnow land was a legal target of the first ability")
	}
	effectstest.B16Activate(t, g, me.ID, vane, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: snow}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, snow).HasSupertype("snow") {
		t.Error("the land is still snow")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == vane {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	effectstest.FloatMana(t, g, me, "{G}{G}")
	effectstest.B16Activate(t, g, me.ID, vane, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LayeredCard(t, g, plain).HasSupertype("snow") {
		t.Error("the basic land did not become snow")
	}
}

func TestOldB833ElvishHealerPreventsTwoToAGreenCreature(t *testing.T) {
	for _, tc := range []struct {
		name   string
		colors []string
		want   int
	}{{"green", []string{"G"}, 2}, {"other", []string{"R"}, 1}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		healer := oldB833Permanent(g, me.ID, "Elvish Healer", "Creature — Elf Cleric", oldB833ElvishHealer, 1, 2)
		bear := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 5, Colors: tc.colors})
		effectstest.B16Activate(t, g, me.ID, healer, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		})
		effectstest.PassPriorityAroundTable(t, g)
		src := effectstest.B12Creature(g, g.Seats[1].ID, "Source", "Creature — Giant", 5, 5)
		g.WithWriteLock(func() {
			if err := g.DealDamageToCreatureForEffect(src, bear, 3); err != nil {
				t.Fatalf("damage: %v", err)
			}
		})
		g.RunStateChecksForTest()
		c := effectstest.LayeredCard(t, g, bear)
		if c.DamageMarked != 3-tc.want {
			t.Errorf("%s: damage %d, want %d", tc.name, c.DamageMarked, 3-tc.want)
		}
	}
}

func TestOldB833ZuranEnchanterMakesAPlayerDiscardOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1-g.Turn.ActiveSeat]
	ze := oldB833Permanent(g, me.ID, "Zuran Enchanter", "Creature — Human Wizard", oldB833ZuranEnchanter, 1, 1)
	d := effectstest.HandSpell(opp, "Doomed", "Sorcery", "{1}")
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	effectstest.B16Activate(t, g, me.ID, ze, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		effectstest.AnswerDiscard(t, g, opp.ID, d)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if oldB833InHand(opp, "Doomed") {
		t.Error("the player kept the card")
	}
	// Not on an opponent's turn.
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	other := g2.Seats[1-g2.Turn.ActiveSeat]
	ze2 := oldB833Permanent(g2, other.ID, "Zuran Enchanter", "Creature — Human Wizard", oldB833ZuranEnchanter, 1, 1)
	effectstest.FloatMana(t, g2, other, "{B}{B}{B}")
	err := g2.ActivateCatalogAbility(other.ID, ze2, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g2.Seats[g2.Turn.ActiveSeat].ID}},
	})
	if err == nil {
		t.Error("activated on an opponent's turn")
	}
}

func TestOldB833ElderDruidTapsOrUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dr := oldB833Permanent(g, me.ID, "Elder Druid", "Creature — Elf Cleric Druid", oldB833ElderDruid, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{G}{G}{G}{G}")
	effectstest.B16Activate(t, g, me.ID, dr, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no tap-or-untap prompt")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if strings.Contains(strings.ToLower(o.Label), "tap") && !strings.Contains(strings.ToLower(o.Label), "untap") {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("no tap option in %v", pick.PickOptions)
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LayeredCard(t, g, bear).Tapped {
		t.Error("the Bear was not tapped")
	}
}

func TestOldB833BarbarianGuidesGrantsSnowLandwalkThenReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bg := oldB833Permanent(g, me.ID, "Barbarian Guides", "Creature — Human Barbarian", oldB833BarbarianGuides, 1, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	effectstest.B16Activate(t, g, me.ID, bg, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land type prompt")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == "Swamp" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("Swamp is not offered")
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "snow swampwalk") {
		t.Error("the target did not gain snow swampwalk")
	}
	if oldB833InHand(me, "Bear") {
		t.Error("the Bear returned before the end step")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB833InHand(me, "Bear") {
		t.Error("the Bear did not return at the end step")
	}
}
