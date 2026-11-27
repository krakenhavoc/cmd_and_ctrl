package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 757 (Tempest damage, #5628): Sandstone Warrior, Ancient
// Runes, Bellowing Fiend, Cursed Scroll, Opportunist, Shadowstorm, Deadshot,
// Kindle and Scalding Tongs.

const (
	oldB757Warrior   = "120a0bcd-eca2-44eb-ad8d-5e4077f7e21c"
	oldB757Runes     = "1a128c02-febf-4ce6-b4d1-28274b5c4ff0"
	oldB757Fiend     = "37331d0e-edda-4791-8bcc-fb457f984877"
	oldB757Scroll    = "514752ea-d453-41d4-97cf-b1dcdb4b4a08"
	oldB757Opp       = "5b5524f7-3c68-4c7b-be1b-745ffcad8006"
	oldB757Storm     = "6e66e983-6349-428b-b511-8738c03ef8fc"
	oldB757Deadshot  = "711c10cb-7a85-417d-be24-79d4e4a89993"
	oldB757Kindle    = "9397a23b-770a-43fc-8bc2-2c3abb0683c5"
	oldB757Tongs     = "939a0cdf-b5e8-4d7d-adb3-f7b2fd7aab8f"
	oldB757SoltariID = "d2cbd71a-841b-4ff2-a794-247a030e0233" // Soltari Lancer: shadow
)

func TestOldB757EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB757Warrior: "Sandstone Warrior", oldB757Runes: "Ancient Runes", oldB757Fiend: "Bellowing Fiend",
		oldB757Scroll: "Cursed Scroll", oldB757Opp: "Opportunist", oldB757Storm: "Shadowstorm",
		oldB757Deadshot: "Deadshot", oldB757Kindle: "Kindle", oldB757Tongs: "Scalding Tongs",
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

func TestOldB757SandstoneWarriorFirstStrikeAndFirebreathing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Sandstone Warrior", "Creature — Human Soldier Warrior", oldB757Warrior, false)
	if !effectstest.HasEffectiveKeyword(t, g, id, "first strike") {
		t.Error("no first strike")
	}
	before := effectstest.EffectivePower(t, g, id)
	effectstest.FillPoolColored(me, "R", 2)
	for i := 0; i < 2; i++ {
		if err := effectstest.InvB1Activate(g, me.ID, id, 0); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.EffectivePower(t, g, id); got != before+2 {
		t.Errorf("power %d, want %d", got, before+2)
	}
	if err := effectstest.InvB1Activate(g, me.ID, id, 0); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestOldB757AncientRunesHitsEachUpkeepPlayerForTheirArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ancient Runes", "Enchantment", oldB757Runes, false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	myLife, oppLife := me.Life, opp.Life
	effectstest.BeginUpkeep(t, g, opp.ID)
	if opp.Life != oppLife-3 {
		t.Errorf("opponent life %d → %d, want 3 damage", oppLife, opp.Life)
	}
	effectstest.BeginUpkeep(t, g, me.ID)
	if me.Life != myLife {
		t.Errorf("my life %d → %d, want none (no artifacts)", myLife, me.Life)
	}
}

func TestOldB757BellowingFiendHitsTheVictimsControllerAndYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if !effectstest.HasEffectiveKeyword(t, g, effectstest.PushCatalogPermanent(g, me.ID, "Bellowing Fiend", "Creature — Spirit", oldB757Fiend, false), "flying") {
		t.Error("no flying")
	}
	fiend := effectstest.PushCatalogPermanent(g, me.ID, "Bellowing Fiend", "Creature — Spirit", oldB757Fiend, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life
	effectstest.EmitNoncombatDamage(g, fiend, bear, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 || me.Life != myLife-3 {
		t.Errorf("life me %d → %d, opp %d → %d, want 3 each", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestOldB757BellowingFiendIgnoresDamageToPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fiend := effectstest.PushCatalogPermanent(g, me.ID, "Bellowing Fiend", "Creature — Spirit", oldB757Fiend, false)
	myLife, oppLife := me.Life, opp.Life
	effectstest.EmitNoncombatDamage(g, fiend, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Errorf("a player hit set off the trigger: me %d → %d, opp %d → %d", myLife, me.Life, oppLife, opp.Life)
	}
}

func oldB757NameChoice(g *game.Game, who game.Player) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCardName && c.Chooser == who.ID {
			return c
		}
	}
	return nil
}

func oldB757UseScroll(t *testing.T, name string, handNames ...string) (g *game.Game, oppLife, after int) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Hand.Cards = nil
	for _, n := range handNames {
		effectstest.PushCatalogHandCard(me, n, "Sorcery", "")
	}
	scroll := effectstest.PushCatalogPermanent(g, me.ID, "Cursed Scroll", "Artifact", oldB757Scroll, false)
	effectstest.FillPool(me, 3)
	oppLife = opp.Life
	if err := effectstest.InvB1Activate(g, me.ID, scroll, 0, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := oldB757NameChoice(g, *me)
	if prompt == nil {
		t.Fatal("no card name prompt")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, name); err != nil {
		t.Fatalf("name: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, oppLife, opp.Life
}

func TestOldB757CursedScrollDamagesOnAMatchOnly(t *testing.T) {
	if _, before, after := oldB757UseScroll(t, "Lone Card", "Lone Card"); after != before-2 {
		t.Errorf("a matching reveal: life %d → %d, want 2 damage", before, after)
	}
	if _, before, after := oldB757UseScroll(t, "Other Card", "Lone Card"); after != before {
		t.Errorf("a missing reveal: life %d → %d, want none", before, after)
	}
	if g, before, after := oldB757UseScroll(t, "Anything"); after != before || len(g.Seats[0].Hand.Cards) != 0 {
		t.Errorf("an empty hand: life %d → %d, want none", before, after)
	}
}

func TestOldB757CursedScrollRequiresTapAndThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scroll := effectstest.PushCatalogPermanent(g, me.ID, "Cursed Scroll", "Artifact", oldB757Scroll, false)
	effectstest.FillPool(me, 2)
	if err := effectstest.InvB1Activate(g, me.ID, scroll, 0, effectstest.PlayerTarget(opp.ID)); err == nil {
		t.Error("activated with only two mana")
	}
}

func TestOldB757OpportunistOnlyPingsDamagedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushCatalogPermanent(g, me.ID, "Opportunist", "Creature — Human Soldier", oldB757Opp, false)
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt Bear", "Creature — Bear", 3, 3)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh Bear", "Creature — Bear", 3, 3)
	effectstest.Damage(t, g, src, hurt, 1)
	if err := effectstest.InvB1Activate(g, me.ID, src, 0, game.TargetRef{Kind: game.TargetCard, ID: fresh}); err == nil {
		t.Error("targeted a creature that was not dealt damage this turn")
	}
	if err := effectstest.InvB1Activate(g, me.ID, src, 0, game.TargetRef{Kind: game.TargetCard, ID: hurt}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, hurt); got != 2 {
		t.Errorf("damage on the hurt bear %d, want 2", got)
	}
}

func TestOldB757ShadowstormHitsOnlyShadowCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shadow := effectstest.PushCatalogPermanent(g, opp.ID, "Soltari Lancer", "Creature — Soltari Knight", oldB757SoltariID, false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Soltari Lancer", "Creature — Soltari Knight", oldB757SoltariID, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.CastCatalogSpell(t, g, "Shadowstorm", "Sorcery", oldB757Storm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The Lancers are 1/1 here, so 2 damage kills them; the bear lives.
	for _, id := range []uuid.UUID{shadow, mine} {
		if _, ok := effectstest.CardOnBattlefield(g, id); ok {
			t.Error("a creature with shadow took no damage")
		}
	}
	if effectstest.DamageMarkedOn(g, bear) != 0 {
		t.Error("a creature without shadow took damage")
	}
}

func TestOldB757DeadshotTapsOneAndTheOtherTakesItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Deadshot", "Sorcery", oldB757Deadshot, []game.TargetRef{
		{Kind: game.TargetCard, ID: big}, {Kind: game.TargetCard, ID: small}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, big) {
		t.Error("the first creature was not tapped")
	}
	if _, ok := effectstest.CardOnBattlefield(g, small); ok {
		t.Error("the second creature survived 4 damage")
	}
	_ = me
}

func TestOldB757DeadshotRefusesTheSameCreatureTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Deadshot", "Sorcery", oldB757Deadshot, []game.TargetRef{
		{Kind: game.TargetCard, ID: big}, {Kind: game.TargetCard, ID: big}}); err == nil {
		t.Error("targeted one creature for both clauses")
	}
}

func TestOldB757KindleCountsKindlesInEveryGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Kindle", "Instant")
	effectstest.PushGraveyardCardWithTypeLine(g, opp.ID, "Kindle", "Instant")
	effectstest.PushGraveyardCardWithTypeLine(g, opp.ID, "Not Kindle", "Instant")
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Kindle", "Instant", oldB757Kindle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("life %d → %d, want 4 damage (2 plus two Kindles)", before, opp.Life)
	}
}

func TestOldB757KindleWithNoOthersDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Kindle", "Instant", oldB757Kindle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d → %d, want 2 damage", before, opp.Life)
	}
}

func TestOldB757ScaldingTongsNeedsThreeOrFewerCards(t *testing.T) {
	for _, tc := range []struct {
		hand int
		want int
	}{{3, 1}, {0, 1}, {4, 0}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		me.Hand.Cards = nil
		for i := 0; i < tc.hand; i++ {
			effectstest.PushCatalogHandCard(me, "Filler", "Sorcery", "")
		}
		effectstest.PushCatalogPermanent(g, me.ID, "Scalding Tongs", "Artifact", oldB757Tongs, false)
		before := opp.Life
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID}) })
		if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
			if tc.want == 0 {
				t.Errorf("hand %d: the ability asked for a target", tc.hand)
			}
			effectstest.AnswerTriggerTargets2(t, g, me.ID, opp.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := before - opp.Life; got != tc.want {
			t.Errorf("hand %d: %d damage, want %d", tc.hand, got, tc.want)
		}
	}
}
