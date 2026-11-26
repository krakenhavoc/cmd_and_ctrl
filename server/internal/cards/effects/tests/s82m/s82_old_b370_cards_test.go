package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b370_cards_test.go — Journey into Nyx, batch 370 (#5558, S82 OLD).
// Every test plays the real catalog entry through the engine.

const (
	oldB370SkybindOracle    = "13d34f21-f0ce-4e59-ac3e-6635bf7292d3"
	oldB370BlowOracle       = "25a20f94-1703-4bcd-862d-f2db7014ba90"
	oldB370BearerOracle     = "54e86bc8-ff48-49be-902c-4923d59901a7"
	oldB370DeicideOracle    = "e421d2bd-740a-40dd-8a10-205edd4991ee"
	oldB370CerberusOracle   = "2e6be9ad-d726-4598-92a7-1ba7eda317a5"
	oldB370CyclopsOracle    = "35101c6c-0bcc-4b1f-9cb5-a82e4d2f8220"
	oldB370GiantOracle      = "9d09697f-8c5c-4bf9-811f-2f847ee374f0"
	oldB370CockatriceOracle = "d5006683-f1c4-4a18-bf64-2d075372534c"
	oldB370RevelOracle      = "5d46b8dd-2a4a-4d9f-b4c6-2b45f05081c2"
	oldB370EndOracle        = "70f5e932-0dc5-4c34-b7be-16cd80ec10b7"
	oldB370EaglesOracle     = "8fc154ff-f2a2-4089-a504-f483f545606e"
	oldB370HubrisOracle     = "95fbd573-c04c-408a-a8c3-f7e1ad5ddac0"
)

func TestOldB370CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB370SkybindOracle:    "Skybind",
		oldB370BlowOracle:       "Spiteful Blow",
		oldB370BearerOracle:     "Bearer of the Heavens",
		oldB370DeicideOracle:    "Deicide",
		oldB370CerberusOracle:   "Wildfire Cerberus",
		oldB370CyclopsOracle:    "Gluttonous Cyclops",
		oldB370GiantOracle:      "Swarmborn Giant",
		oldB370CockatriceOracle: "Fleetfeather Cockatrice",
		oldB370RevelOracle:      "Revel of the Fallen God",
		oldB370EndOracle:        "Nightmarish End",
		oldB370EaglesOracle:     "Rise of Eagles",
		oldB370HubrisOracle:     "Hubris",
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

func oldB370Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// oldB370Float floats the coloured part of a cost as mana and the generic part
// as colourless.
func oldB370Float(t *testing.T, g *game.Game, p *game.Player, colored string, generic int) {
	t.Helper()
	effectstest.FloatMana(t, g, p, colored)
	effectstest.B09AddColorless(p, generic)
}

func oldB370Card(ids ...uuid.UUID) []game.TargetRef {
	var out []game.TargetRef
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func TestOldB370HubrisReturnsTheCreatureAndItsAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB370Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Aura", TypeLine: "Enchantment — Aura",
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Aura", TypeLine: "Enchantment — Aura",
		Owner: opp.ID, Controller: opp.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}})
	other := effectstest.B12Creature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hubris", "Instant", oldB370HubrisOracle, oldB370Card(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) || !opp.Hand.Contains(theirs) {
		t.Errorf("the creature and its owner's Aura should be in their owner's hand (bear in %v, aura in %v)",
			effectstest.ZoneOfCard(g, bear), effectstest.ZoneOfCard(g, theirs))
	}
	if !me.Hand.Contains(mine) {
		t.Errorf("my Aura is in %v, want my hand", effectstest.ZoneOfCard(g, mine))
	}
	if effectstest.ZoneOfCard(g, other) != game.ZoneBattlefield {
		t.Error("an unrelated creature was returned")
	}
}

func TestOldB370NightmarishEndShrinksByTheCardsInYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB370Seats(g)
	for me.Hand.Size() > 0 {
		_, _ = me.Hand.PopTop()
	}
	for i := 0; i < 3; i++ {
		effectstest.EoHand(me, "Card", "Sorcery")
	}
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Nightmarish End", "Instant", oldB370EndOracle, oldB370Card(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("the Giant is %d/%d, want 2/2 (-3/-3 for three cards in hand)", p, tough)
	}
}

func TestOldB370SpitefulBlowDestroysTheCreatureAndTheLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB370Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	bystander := effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	effectstest.CastCatalogSpell(t, g, "Spiteful Blow", "Sorcery", oldB370BlowOracle, oldB370Card(bear, land))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, bear) != game.ZoneGraveyard || effectstest.ZoneOfCard(g, land) != game.ZoneGraveyard {
		t.Errorf("bear in %v, land in %v, want both in the graveyard", effectstest.ZoneOfCard(g, bear), effectstest.ZoneOfCard(g, land))
	}
	if effectstest.ZoneOfCard(g, bystander) != game.ZoneBattlefield {
		t.Error("another land was destroyed")
	}
}

func TestOldB370RevelMakesFourHastySatyrs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Revel of the Fallen God", "Sorcery", oldB370RevelOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Satyr"); n != 4 {
		t.Fatalf("%d Satyr tokens, want 4", n)
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Satyr" {
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "haste") || !c.HasColor("R") || !c.HasColor("G") {
				t.Errorf("Satyr %v lacks haste or its colors", c.InstanceID)
			}
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 2 || tough != 2 {
				t.Errorf("Satyr is %d/%d, want 2/2", p, tough)
			}
		}
	}
}

func TestOldB370RiseOfEaglesMakesTwoEnchantmentBirdsAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB370Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.CastCatalogSpell(t, g, "Rise of Eagles", "Sorcery", oldB370EaglesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Bird"); n != 2 {
		t.Fatalf("%d Bird tokens, want 2", n)
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Bird" {
			if !c.IsEnchantment() || !c.IsCreature() || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") || !c.HasColor("U") {
				t.Errorf("the Bird is not a blue flying enchantment creature")
			}
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 2 || tough != 2 {
				t.Errorf("Bird is %d/%d, want 2/2", p, tough)
			}
		}
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt")
	}
}

func TestOldB370SkybindExilesANonenchantmentAndReturnsItAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB370Seats(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	shrine := effectstest.PushCatalogPermanent(g, opp.ID, "Their Shrine", "Enchantment", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Skybind", "Enchantment", oldB370SkybindOracle, false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Other Shrine", "Enchantment", "", false)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: other, Actor: me.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, bad := range []uuid.UUID{shrine} {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Fatal("an enchantment was accepted as the target")
			}
		}
	}
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, bear) != game.ZoneExile {
		t.Fatalf("the Bear is in %v, want exile", effectstest.ZoneOfCard(g, bear))
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Their Bear") {
		t.Error("the Bear did not return at the beginning of the next end step")
	}
}

func TestOldB370BearerOfTheHeavensDestroysEverythingAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB370Seats(g)
	effectstest.AdvanceToMain(t, g)
	bearer := effectstest.B12Push(g, me.ID, "Bearer of the Heavens", "Creature — Giant", oldB370BearerOracle, 10, 10)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.B27Kill(g, bearer)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, bear) != game.ZoneBattlefield {
		t.Fatal("permanents were destroyed before the end step")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, bear) != game.ZoneGraveyard || effectstest.ZoneOfCard(g, land) != game.ZoneGraveyard {
		t.Errorf("bear in %v, land in %v, want both destroyed", effectstest.ZoneOfCard(g, bear), effectstest.ZoneOfCard(g, land))
	}
}

func TestOldB370DeicideExilesAnEnchantmentAndOnlyAGodOpensTheSearch(t *testing.T) {
	for _, god := range []bool{true, false} {
		typeLine, name := "Enchantment", "Plain Shrine"
		if god {
			typeLine, name = "Legendary Enchantment Creature — God", "Some God"
		}
		x := effectstest.ExtractionSetUp(t, name, typeLine)
		x.Permanent("target", name, typeLine)
		effectstest.CastCatalogSpell(t, x.G, "Deicide", "Instant", oldB370DeicideOracle, x.Target("target"))
		effectstest.PassPriorityAroundTable(t, x.G)
		x.Exiled(t, "target")
		if !god {
			if c := effectstest.SearchChoiceFor(x.G, x.Me.ID); c != nil {
				t.Errorf("a non-God opened the search")
			}
			x.NotExiled(t, "gy", "gy 2", "hand", "lib")
			continue
		}
		if c := effectstest.SearchChoiceFor(x.G, x.Me.ID); c == nil || x.G.SearchFloorLocked(c) != 0 {
			t.Fatal("a God opened no search, or one that must find something")
		}
		effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["gy"], x.Cards["gy 2"], x.Cards["hand"], x.Cards["lib"])
		x.Exiled(t, "gy", "gy 2", "hand", "lib")
		x.NotExiled(t, "gy other", "hand other", "lib other")
	}
}

func TestOldB370DeicideRefusesANonEnchantmentTarget(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Plain Shrine", "Enchantment")
	x.Permanent("target", "Plain Shrine", "Enchantment")
	creature := x.Permanent("bear", "Bear", "Creature — Bear")
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Deicide", "Instant", oldB370DeicideOracle, oldB370Card(creature)); err == nil {
		t.Fatal("Deicide targeted a creature that is not an enchantment")
	}
}

func TestOldB370WildfireCerberusBurnsOpponentsWhenMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB370Seats(g)
	cerberus := effectstest.PushMonster(g, me.ID, "Wildfire Cerberus", oldB370CerberusOracle, 4, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 5)
	life := opp.Life
	oldB370Float(t, g, me, "{R}{R}", 5)
	effectstest.ActivateMonstrosity(t, g, me.ID, cerberus, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, cerberus); n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", n)
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want -2", life, opp.Life)
	}
	if got := effectstest.DamageOn(t, g, theirs); got != 2 {
		t.Errorf("their creature has %d damage, want 2", got)
	}
	if got := effectstest.DamageOn(t, g, mine); got != 0 {
		t.Errorf("my creature has %d damage, want 0", got)
	}
}

func TestOldB370MonstrosityCardsGetTheirCounters(t *testing.T) {
	for _, c := range []struct {
		name, oracle, colored string
		generic, counters     int
	}{
		{"Gluttonous Cyclops", oldB370CyclopsOracle, "{R}{R}", 5, 3},
		{"Fleetfeather Cockatrice", oldB370CockatriceOracle, "{G}{U}", 5, 3},
		{"Swarmborn Giant", oldB370GiantOracle, "{G}{G}", 4, 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := oldB370Seats(g)
			id := effectstest.PushMonster(g, me.ID, c.name, c.oracle, 3, 3)
			oldB370Float(t, g, me, c.colored, c.generic)
			effectstest.ActivateMonstrosity(t, g, me.ID, id, 0)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.PlusOneCounters(g, id); n != c.counters {
				t.Errorf("+1/+1 counters = %d, want %d", n, c.counters)
			}
			if !effectstest.MonstrousNow(g, id) {
				t.Error("not monstrous")
			}
		})
	}
}

func TestOldB370CockatriceKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB370Seats(g)
	id := effectstest.B12Push(g, me.ID, "Fleetfeather Cockatrice", "Creature — Cockatrice", oldB370CockatriceOracle, 3, 3)
	effectstest.AssertKeywords(t, g, id, "flash", "flying", "deathtouch")
}

func TestOldB370SwarmbornGiantHasReachOnlyOnceMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB370Seats(g)
	id := effectstest.PushMonster(g, me.ID, "Swarmborn Giant", oldB370GiantOracle, 6, 6)
	if effectstest.HasEffectiveKeyword(t, g, id, "reach") {
		t.Fatal("reach before it is monstrous")
	}
	oldB370Float(t, g, me, "{G}{G}", 4)
	effectstest.ActivateMonstrosity(t, g, me.ID, id, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "reach") {
		t.Error("no reach once monstrous")
	}
}

func TestOldB370SwarmbornGiantIsSacrificedWhenYouAreDealtCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB370Seats(g)
	giant := effectstest.B12Push(g, me.ID, "Swarmborn Giant", "Creature — Giant", oldB370GiantOracle, 6, 6)
	attacker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: attacker, Target: me.ID, Amount: 2, Combat: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, giant) != game.ZoneGraveyard {
		t.Errorf("the Giant is in %v, want the graveyard", effectstest.ZoneOfCard(g, giant))
	}
}

func TestOldB370SwarmbornGiantIgnoresNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB370Seats(g)
	giant := effectstest.B12Push(g, me.ID, "Swarmborn Giant", "Creature — Giant", oldB370GiantOracle, 6, 6)
	attacker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: attacker, Target: me.ID, Amount: 2})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, giant) != game.ZoneBattlefield {
		t.Error("noncombat damage sacrificed the Giant")
	}
}
