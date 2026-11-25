package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB12Tempt     = "06e1c0fa-767c-4204-972f-d98f770d85f3"
	oldB12Technique = "0ab9539d-3461-4936-95cd-4c229308c4a4"
	oldB12Reckoning = "0d3e0b38-a9ff-489b-8c00-4ca89deb4869"
	oldB12Toshiro   = "11838086-db2f-4588-ae18-4129c9e2b67d"
	oldB12Lamia     = "2c1d288a-22ea-482c-947e-f3ab04172da0"
	oldB12Relentl   = "7404c078-228b-4296-bf1f-62f57bf832d9"
	oldB12Vengevine = "7ed52301-81ea-4e7f-b985-cfab0593cae4"
	oldB12Arden     = "82d51c8a-d42b-48b3-a726-3a8d28d057e5"
	oldB12Cut       = "98a5bf1a-1088-4339-9a9b-6ee5e4956cf1"
)

func TestOldB12CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB12Tempt: "Tempt with Immortality", oldB12Technique: "Incarnation Technique",
		oldB12Reckoning: "Triumphant Reckoning", oldB12Toshiro: "Toshiro Umezawa",
		oldB12Lamia: "Gravebreaker Lamia", oldB12Relentl: "Relentless Dead", oldB12Vengevine: "Vengevine",
		oldB12Arden: "Arden Angel", oldB12Cut: "Cut", oldB12Cut + "#1": "Ribbons",
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

func oldB12Grave(p *game.Player, name, typeLine, cost string, power int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Power: power, Toughness: power, Owner: p.ID, Controller: p.ID})
	return id
}

func oldB12Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id, err := effectstest.Msc17Cast(t, g, p, c, params)
	if err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return id
}

func TestOldB12TriumphantReckoningReturnsArtifactsEnchantmentsAndPlaneswalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	art := oldB12Grave(me, "Rock", "Artifact", "{1}", 0)
	ench := oldB12Grave(me, "Aura Thing", "Enchantment", "{1}{W}", 0)
	pw := oldB12Grave(me, "Walker Jace", "Legendary Planeswalker — Jace", "{2}{U}{U}", 0)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == pw {
				me.Graveyard.Cards[i].StartingLoyalty = 3
			}
		}
	})
	bear := oldB12Grave(me, "Bear", "Creature — Bear", "{1}{G}", 2)
	bolt := oldB12Grave(me, "Shock", "Instant", "{R}", 0)
	oldB12Cast(t, g, me, game.Card{Name: "Triumphant Reckoning", TypeLine: "Sorcery", OracleID: oldB12Reckoning}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{art, ench, pw} {
		if z := effectstest.B12ZoneOf(g, id); z != game.ZoneBattlefield {
			t.Errorf("%s is in %s, want the battlefield", effectstest.B12Card(t, g, id).Name, z)
		}
	}
	for _, id := range []uuid.UUID{bear, bolt} {
		if z := effectstest.B12ZoneOf(g, id); z != game.ZoneGraveyard {
			t.Errorf("%s is in %s, want the graveyard", effectstest.B12Card(t, g, id).Name, z)
		}
	}
}

func TestOldB12IncarnationTechniqueMillsFiveThenReturnsAChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	}
	// Push puts at the bottom; the top five are the ones to mill, so seat them on top.
	var top []uuid.UUID
	for i := 0; i < 5; i++ {
		id := uuid.New()
		top = append(top, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Milled Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	}
	old := oldB12Grave(me, "Old Ogre", "Creature — Ogre", "{3}", 3)
	oldB12Cast(t, g, me, game.Card{Name: "Incarnation Technique", TypeLine: "Sorcery", OracleID: oldB12Technique}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// A demonstrate offer may be open: decline it.
	if p := effectstest.LatestTriggerPrompt(g, me.ID); p != nil {
		_ = g.ResolveTriggerPrompt(p.ID, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, id := range top {
		if z := effectstest.B12ZoneOf(g, id); z != game.ZoneGraveyard && z != game.ZoneBattlefield {
			t.Errorf("a top card is in %s, want it milled", z)
		}
	}
	effectstest.SlChoose(t, g, me.ID, top[0])
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, top[0]); z != game.ZoneBattlefield {
		t.Errorf("the chosen freshly milled creature is in %s, want the battlefield", z)
	}
	if z := effectstest.B12ZoneOf(g, old); z != game.ZoneGraveyard {
		t.Errorf("the unchosen creature is in %s, want the graveyard", z)
	}
}

func TestOldB12TemptWithImmortalityReturnsOneForYouAndOneMoreForEachOpponentWhoTakesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	mine1 := oldB12Grave(me, "My Bear", "Creature — Bear", "{1}{G}", 2)
	mine2 := oldB12Grave(me, "My Ogre", "Creature — Ogre", "{3}", 3)
	theirs := oldB12Grave(opp, "Their Elk", "Creature — Elk", "{2}", 2)
	oldB12Cast(t, g, me, game.Card{Name: "Tempt with Immortality", TypeLine: "Sorcery", OracleID: oldB12Tempt}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SlChoose(t, g, me.ID, mine1)
	effectstest.PassPriorityAroundTable(t, g)
	// The first opponent takes the offer; the others decline.
	for _, p := range g.Seats[1:] {
		if effectstest.LatestConfirmFor(g, p.ID) == nil {
			continue
		}
		effectstest.AnswerMayChoice(t, g, p.ID, p.ID == opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if p.ID == opp.ID {
			effectstest.SlChoose(t, g, opp.ID, theirs)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.SlChoose(t, g, me.ID, mine2)
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
	for _, id := range []uuid.UUID{mine1, mine2} {
		if z := effectstest.B12ZoneOf(g, id); z != game.ZoneBattlefield {
			t.Errorf("%s is in %s, want the battlefield", effectstest.B12Card(t, g, id).Name, z)
		}
	}
	if z := effectstest.B12ZoneOf(g, theirs); z != game.ZoneBattlefield {
		t.Errorf("the opponent's creature is in %s, want the battlefield", z)
	}
}

func TestOldB12ToshiroCastsAnInstantFromTheGraveyardAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Toshiro Umezawa", "Legendary Creature — Human Samurai", oldB12Toshiro, 2, 2)
	instant := oldB12Grave(me, "Test Instant", "Instant", "{B}", 0)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(victim) })
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: instant}}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestTriggerPrompt(g, me.ID); p != nil {
		if err := g.ResolveTriggerPrompt(p.ID, me.ID, true); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.FloatForTest(g, me, "B")
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, instant, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, instant); z != game.ZoneExile {
		t.Errorf("the cast instant is in %s, want exile instead of the graveyard", z)
	}
}

func TestOldB12GravebreakerLamiaSearchesToTheGraveyardAndDiscountsGraveyardSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	target := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Nice Card", TypeLine: "Sorcery"})
	oldB12Cast(t, g, me, game.Card{
		Name: "Gravebreaker Lamia", TypeLine: "Enchantment Creature — Snake Lamia", OracleID: oldB12Lamia, Power: 4, Toughness: 4,
	}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if z := effectstest.B12ZoneOf(g, target); z != game.ZoneGraveyard && z != game.ZoneLibrary {
		t.Errorf("the searched card is in %s", z)
	}
	price := func(zone game.ZoneKind) int {
		base, err := game.ParseCost("{3}{R}")
		if err != nil {
			t.Fatal(err)
		}
		out, err := g.ApplyCostModifiers(base, game.CostQuery{
			Card:       game.Card{InstanceID: uuid.New(), Name: "Flashed", TypeLine: "Sorcery", ManaCost: "{3}{R}", Owner: me.ID, Controller: me.ID},
			Controller: me.ID, FromZone: zone,
		})
		if err != nil {
			t.Fatal(err)
		}
		return out.ManaValue()
	}
	if got := price(game.ZoneGraveyard); got != 3 {
		t.Errorf("a spell cast from the graveyard costs %d, want 3", got)
	}
	if got := price(game.ZoneHand); got != 4 {
		t.Errorf("a spell cast from hand costs %d, want 4", got)
	}
}

func TestOldB12RelentlessDeadReturnsToHandForBlackAndAZombieForItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	dead := effectstest.B12Push(g, me.ID, "Relentless Dead", "Creature — Zombie", oldB12Relentl, 2, 2)
	zombie := oldB12Grave(me, "Gravecrawler", "Creature — Zombie", "{1}{B}", 2)
	nonZombie := oldB12Grave(me, "Bear", "Creature — Bear", "{1}{G}", 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(dead) })
	effectstest.FloatForTest(g, me, "BBB")
	picked, paid := false, 0
	for i := 0; i < 12; i++ {
		switch {
		case effectstest.LatestPickTarget(g, me.ID) != nil && !picked:
			p := effectstest.LatestPickTarget(g, me.ID)
			if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: nonZombie}}); err == nil {
				t.Fatal("a non-Zombie was a legal target")
			}
			if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: dead}}); err == nil {
				t.Fatal("Relentless Dead itself was a legal target (it says another)")
			}
			if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: zombie}}); err != nil {
				t.Fatal(err)
			}
			picked = true
		case effectstest.HasPayUnlessFor(g, me.ID):
			effectstest.AnswerPayUnless(t, g, me.ID, true)
			paid++
		case effectstest.TriggerOrderPromptFor(g, me.ID) != nil:
			o := effectstest.TriggerOrderPromptFor(g, me.ID)
			if err := g.ResolveTriggerOrder(o.ID, me.ID, append([]uuid.UUID(nil), o.TriggerOrderIDs...)); err != nil {
				t.Fatal(err)
			}
		default:
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
	if !picked || paid != 2 {
		t.Fatalf("picked %v, paid %d offers; want the Zombie targeted and both payments offered (%+v)", picked, paid, g.PendingChoices)
	}
	if z := effectstest.B12ZoneOf(g, zombie); z != game.ZoneBattlefield {
		t.Errorf("the Zombie is in %s, want the battlefield", z)
	}
	if z := effectstest.B12ZoneOf(g, dead); z != game.ZoneHand {
		t.Errorf("Relentless Dead is in %s, want its owner's hand", z)
	}
	if z := effectstest.B12ZoneOf(g, nonZombie); z != game.ZoneGraveyard {
		t.Errorf("the Bear is in %s, want the graveyard", z)
	}
}

func TestOldB12VengevineReturnsOnTheSecondCreatureSpellOnly(t *testing.T) {
	for _, order := range [][]string{{"Creature — Bear", "Creature — Bear"}, {"Sorcery", "Creature — Bear"}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.ToMain(t, g)
		vine := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: vine, Name: "Vengevine", TypeLine: "Creature — Elemental", OracleID: oldB12Vengevine, Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID})
		for i, ty := range order {
			oldB12Cast(t, g, me, game.Card{Name: "Spell", TypeLine: ty, Power: 1, Toughness: 1}, game.CastSpellParams{})
			effectstest.PassPriorityAroundTable(t, g)
			if p := effectstest.LatestTriggerPrompt(g, me.ID); p != nil {
				if err := g.ResolveTriggerPrompt(p.ID, me.ID, true); err != nil {
					t.Fatal(err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				if i != 1 {
					t.Errorf("order %v: Vengevine's trigger offered after spell %d", order, i+1)
				}
			}
		}
		wantBack := order[0] == "Creature — Bear"
		back := effectstest.B12ZoneOf(g, vine) == game.ZoneBattlefield
		if back != wantBack {
			t.Errorf("order %v: Vengevine on the battlefield = %v, want %v", order, back, wantBack)
		}
	}
}

func TestOldB12ArdenAngelRollsADieInTheGraveyardAtUpkeep(t *testing.T) {
	back, stayed := 0, 0
	for table := 0; table < 60 && (back == 0 || stayed == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		angel := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: angel, Name: "Arden Angel", TypeLine: "Creature — Angel", OracleID: oldB12Arden, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
		effectstest.B12ToMyNextUpkeep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		rolls := len(effectstest.RandomEvents(g, game.EventRollDie))
		if rolls == 0 {
			t.Fatal("no die was rolled at the upkeep")
		}
		if z := effectstest.B12ZoneOf(g, angel); z == game.ZoneBattlefield {
			back++
		} else {
			stayed++
		}
	}
	if back == 0 || stayed == 0 {
		t.Fatalf("never saw both outcomes (back %d, stayed %d)", back, stayed)
	}
}

func TestOldB12CutDealsFourAndRibbonsDrainsXFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Cut // Ribbons", TypeLine: "Sorcery // Sorcery", OracleID: oldB12Cut,
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast Cut: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("4 damage did not kill the 4/4")
	}
	if z := effectstest.B12ZoneOf(g, id); z != game.ZoneGraveyard {
		t.Fatalf("the split card is in %s, want the graveyard for aftermath", z)
	}
}
