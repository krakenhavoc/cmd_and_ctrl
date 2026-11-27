package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB549Eulogist = "21017dc0-9b79-48b8-8eae-561abfad31d5"
	oldB549Knight   = "4e6c71e8-aeaf-404e-84c2-c79f2a1cbe4b"
	oldB549Boggart  = "593b229b-579b-4052-8a00-dac787b149b1"
	oldB549Tauntng  = "5c5d1a62-3b83-4812-82bb-d95b39b990b4"
	oldB549Nectar   = "6d0a2e5f-af10-4754-81d7-b7afa6282d7d"
	oldB549Judge    = "b4eafd12-5f3e-43e6-aa0a-6bb5f4f5782b"
	oldB549Heal     = "cc65a99f-ac7f-4ded-bb00-a5daf51e4bcd"
	oldB549BogAsh   = "e3d4dd61-d646-449c-84f4-0ff927ca6b0a"
)

func TestOldB549CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB549Eulogist: "Elvish Eulogist", oldB549Knight: "Knight of Meadowgrain",
		oldB549Boggart: "Quill-Slinger Boggart", oldB549Tauntng: "Faerie Tauntings",
		oldB549Nectar: "Nectar Faerie", oldB549Judge: "Judge of Currents",
		oldB549Heal: "Heal the Scars", oldB549BogAsh: "Bog-Strider Ash",
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

func TestOldB549KnightOfMeadowgrainHasFirstStrikeAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := effectstest.B12Push(g, me.ID, "Knight of Meadowgrain", "Creature — Kithkin Knight", oldB549Knight, 2, 2)
	for _, kw := range []string{"first strike", "lifelink"} {
		if !effectstest.HasKeywordOf(g, knight, kw) {
			t.Errorf("the Knight lacks %s", kw)
		}
	}
}

func TestOldB549ElvishEulogistGainsOneForEachElfCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	eulogist := effectstest.B12Push(g, me.ID, "Elvish Eulogist", "Creature — Elf Shaman", oldB549Eulogist, 1, 1)
	g.WithWriteLock(func() {
		for _, tl := range []string{"Creature — Elf Warrior", "Creature — Elf Druid", "Creature — Human"} {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: tl, Owner: me.ID, Controller: me.ID})
		}
	})
	life := me.Life
	effectstest.ApaActivate(t, g, me, eulogist, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(eulogist) {
		t.Error("the Eulogist was not sacrificed")
	}
	// Two Elf cards already there, and the Eulogist itself, an Elf card by now.
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestOldB549JudgeOfCurrentsOffersLifeWhenAMerfolkBecomesTapped(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B12Push(g, me.ID, "Judge of Currents", "Creature — Merfolk Wizard", oldB549Judge, 1, 1)
		other := effectstest.B12Push(g, me.ID, "Other Merfolk", "Creature — Merfolk", "", 1, 1)
		bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		oppsMerfolk := effectstest.B12Push(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Merfolk", "Creature — Merfolk", "", 1, 1)
		effectstest.TapForTest(t, g, bear, oppsMerfolk)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OwedMays(g, me.ID) != 0 {
			t.Fatal("a non-Merfolk, or an opponent's Merfolk, becoming tapped offered life")
		}
		life := me.Life
		effectstest.TapForTest(t, g, other)
		effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if yes {
			want++
		}
		if me.Life != want {
			t.Errorf("accept %v: life %d, want %d", yes, me.Life, want)
		}
	}
}

func TestOldB549FaerieTauntingsDrainsOnlyOnYourCastsDuringAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, me := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Faerie Tauntings", "Kindred Enchantment — Faerie", oldB549Tauntng, 0, 0)

	effectstest.AangAdvanceToMain(t, g, 1) // your own turn: silent
	oppLife := opp.Life
	effectstest.KataraCastInstant(t, g, me)
	if effectstest.OwedMays(g, me.ID) != 0 || opp.Life != oppLife {
		t.Fatal("a spell cast on your own turn triggered the Tauntings")
	}

	effectstest.AangAdvanceToMain(t, g, 0) // an opponent's turn: their cast is silent, yours drains
	effectstest.KataraCastInstant(t, g, opp)
	if effectstest.OwedMays(g, me.ID) != 0 {
		t.Fatal("an opponent's spell triggered the Tauntings")
	}
	effectstest.KataraCastInstant(t, g, me)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife {
		t.Errorf("declined: opponent life %d, want %d", opp.Life, oppLife)
	}
	effectstest.KataraCastInstant(t, g, me)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("accepted: opponent life %d, want %d", opp.Life, oppLife-1)
	}
}

func oldB549Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost string, colors []string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors, Owner: p.ID, Controller: p.ID})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func TestOldB549QuillSlingerBoggartDrainsOnAnyKithkinSpell(t *testing.T) {
	for caster := range 2 {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		effectstest.B12Push(g, me.ID, "Quill-Slinger Boggart", "Creature — Goblin Warrior", oldB549Boggart, 3, 2)
		p := g.Seats[caster] // seat 0 is the Boggart's controller, seat 1 an opponent
		life := opp.Life
		oldB549Cast(t, g, p, "Kithkin Dude", "Kindred Instant — Kithkin", "", []string{"W"})
		effectstest.AnswerTriggerTargets2(t, g, me.ID, opp.ID)
		effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != life-1 {
			t.Errorf("caster seat %d: target life %d, want %d", caster, opp.Life, life-1)
		}
	}
}

func TestOldB549QuillSlingerBoggartIgnoresNonKithkinSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Quill-Slinger Boggart", "Creature — Goblin Warrior", oldB549Boggart, 3, 2)
	p := g.Seats[g.Turn.ActiveSeat]
	oldB549Cast(t, g, p, "Goblin Dude", "Kindred Instant — Goblin", "", []string{"R"})
	if effectstest.OwedMays(g, me.ID) != 0 || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a Goblin spell triggered the Boggart")
	}
}

func TestOldB549NectarFaerieGivesAFaerieOrElfLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	faerie := effectstest.B12Push(g, me.ID, "Nectar Faerie", "Creature — Faerie Wizard", oldB549Nectar, 1, 1)
	elf := effectstest.B12Push(g, me.ID, "Elf", "Creature — Elf", "", 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, faerie, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("a Bear was a legal target")
	}
	effectstest.ApaMana(me, "B")
	effectstest.ApaActivate(t, g, me, faerie, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: elf}}})
	if !effectstest.HasKeywordOf(g, elf, "lifelink") {
		t.Error("the Elf did not gain lifelink")
	}
	if effectstest.HasKeywordOf(g, bear, "lifelink") {
		t.Error("the Bear gained lifelink")
	}
}

func TestOldB549HealTheScarsRegeneratesAndGainsItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	wall := effectstest.B12Push(g, me.ID, "Wall", "Creature — Wall", "", 0, 5)
	life := me.Life
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heal the Scars", TypeLine: "Instant", ManaCost: "{3}{G}", Colors: []string{"G"}, OracleID: oldB549Heal, Owner: me.ID, Controller: me.ID})
	})
	effectstest.ApaMana(me, "G", "G", "G", "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wall}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d, want %d", me.Life, life+5)
	}
	effectstest.B25Destroy(g, wall)
	if !g.Battlefield.Contains(wall) {
		t.Error("the shield did not regenerate the Wall")
	}
}

func TestOldB549BogStriderAshPaysGForTwoLifeOnAnyGoblinSpell(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		effectstest.B12Push(g, me.ID, "Bog-Strider Ash", "Creature — Treefolk Shaman", oldB549BogAsh, 2, 4)
		p := g.Seats[g.Turn.ActiveSeat]
		oldB549Cast(t, g, p, "Kithkin Dude", "Kindred Instant — Kithkin", "", []string{"W"})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("a Kithkin spell triggered the Ash")
		}
		oldB549Cast(t, g, p, "Goblin Dude", "Kindred Instant — Goblin", "", []string{"R"})
		effectstest.ApaMana(me, "G")
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("a Goblin spell did not offer {G}")
		}
		life := me.Life
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if pay {
			want += 2
		}
		if me.Life != want || (pay && me.ManaPool.CountColor("G") != 0) {
			t.Errorf("pay %v: life %d want %d, G left %d", pay, me.Life, want, me.ManaPool.CountColor("G"))
		}
	}
}
