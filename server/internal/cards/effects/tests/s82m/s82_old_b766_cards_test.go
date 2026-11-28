package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b766_cards_test.go — S82 OLD batch 766 (#5628, Tempest: card draw and
// library). Each test plays the real catalog entry.

const (
	oldB766MnemonicSliver = "101a5c2d-b0b4-45a7-b236-df474ac47356"
	oldB766ManaSeverance  = "34c66f52-dcbf-4d0a-bf7f-cb294bd93e9c"
	oldB766BoobyTrap      = "3e46381a-7850-4e50-9213-e1a180299977"
	oldB766FoolsTome      = "65589538-d3c1-4dfd-a328-71f57f556212"
	oldB766Shocker        = "8981d1d1-e102-4a3d-96ad-3904c83d1611"
	oldB766Interdict      = "93771480-2672-4bc7-81a8-011d3b8878b0"
	oldB766Insight        = "c77afbeb-cc9b-4770-9a48-e7460562895f"
	oldB766SacredGuide    = "de9e8b71-897d-4a9d-942b-af10f31e0425"
	oldB766FugitiveDruid  = "e05f3236-dc82-4353-87ab-d26ae7caa0e7"
	oldB766Sorcerer       = "5e961d15-5972-4e4b-9385-1cd7cd7c6bbe"
)

func TestOldB766EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB766MnemonicSliver: "Mnemonic Sliver",
		oldB766ManaSeverance:  "Mana Severance",
		oldB766BoobyTrap:      "Booby Trap",
		oldB766FoolsTome:      "Fool's Tome",
		oldB766Shocker:        "Shocker",
		oldB766Interdict:      "Interdict",
		oldB766Insight:        "Insight",
		oldB766SacredGuide:    "Sacred Guide",
		oldB766FugitiveDruid:  "Fugitive Druid",
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

func TestOldB766MnemonicSliverGivesEverySliverASacrificeToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Mnemonic Sliver", "Creature — Sliver", oldB766MnemonicSliver, 2, 2)
	mine := effectstest.TspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	theirs := effectstest.TspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
	bear := effectstest.B12Permanent(g, me.ID, "Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, bear); idx >= 0 {
		t.Error("a non-Sliver was given the ability")
	}
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, theirs); idx < 0 {
		t.Error("an opponent's Sliver was not given the ability")
	}
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.TspB1ActivateGranted(t, g, me.ID, mine, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the Sliver was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB766ManaSeveranceExilesAnyNumberOfLandsAndShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	var lands []uuid.UUID
	for _, n := range []string{"Forest", "Island", "Mountain"} {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Basic Land — " + n, Owner: me.ID, Controller: me.ID})
		lands = append(lands, id)
	}
	spell := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: spell, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Mana Severance", "Sorcery", oldB766ManaSeverance, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pc := effectstest.ChkB1OpenSearch(g)
	if pc == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(pc.ID, pc.Chooser, lands[:2]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, want := range []game.ZoneKind{game.ZoneExile, game.ZoneExile, game.ZoneLibrary} {
		if got := effectstest.ZoneOf(g, lands[i]); got != want {
			t.Errorf("land %d is in %s, want %s", i, got, want)
		}
	}
	if got := effectstest.ZoneOf(g, spell); got != game.ZoneLibrary {
		t.Errorf("the nonland card is in %s", got)
	}
}

func TestOldB766FoolsTomeDrawsOnlyWithAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tome := effectstest.PushCatalogPermanent(g, me.ID, "Fool's Tome", "Artifact — Book", oldB766FoolsTome, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the Tome drew with cards in hand")
	}
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the Tome refused an empty hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d, want 1", me.Hand.Size())
	}
}

func TestOldB766InsightDrawsOnlyForAnOpponentsGreenSpell(t *testing.T) {
	for _, tc := range []struct {
		name, cost string
		opponent   bool
		draws      bool
	}{
		{"Giant Growth", "{G}", true, true},
		{"Opt", "{U}", true, false},
		{"Giant Growth", "{G}", false, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.PushCatalogPermanent(g, me.ID, "Insight", "Enchantment", oldB766Insight, false)
		effectstest.AdvanceToMain(t, g)
		caster := me
		if tc.opponent {
			caster = opp
		}
		effectstest.FloatForTest(g, caster, tc.cost[1:2])
		before := me.Hand.Size()
		id := uuid.New()
		g.WithWriteLock(func() {
			caster.Hand.PushTop(game.Card{InstanceID: id, Name: tc.name, TypeLine: "Instant", ManaCost: tc.cost, Owner: caster.ID, Controller: caster.ID})
		})
		if caster == me {
			before = me.Hand.Size() - 1
		}
		if err := g.CastSpell(caster.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := before
		if tc.draws {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("%s (opponent %v): hand %d, want %d", tc.name, tc.opponent, me.Hand.Size(), want)
		}
	}
}

func TestOldB766FugitiveDruidDrawsWhenAnAuraSpellTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	druid := effectstest.B12Push(g, me.ID, "Fugitive Druid", "Creature — Human Druid", oldB766FugitiveDruid, 3, 2)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	cast := func(name, typeLine, cost, mana string, target uuid.UUID) {
		t.Helper()
		effectstest.FloatForTest(g, me, mana)
		effectstest.CastFromHandForTest(t, g, me, name, typeLine, cost, "", game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(target)})
		effectstest.PassPriorityAroundTable(t, g)
	}
	hand := me.Hand.Size()
	cast("Pacifism", "Enchantment — Aura", "{1}{W}", "CW", other)
	if me.Hand.Size() != hand {
		t.Errorf("an Aura on another creature drew: hand %d, want %d", me.Hand.Size(), hand)
	}
	cast("Pacifism", "Enchantment — Aura", "{1}{W}", "CW", druid)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an Aura on the Druid: hand %d, want %d", me.Hand.Size(), hand+1)
	}
	cast("Shock", "Instant", "{R}", "R", druid)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a non-Aura spell on the Druid drew: hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB766ShockerWheelsThePlayerItDamages(t *testing.T) {
	for _, hand := range []int{3, 0} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		shocker := effectstest.B12Push(g, me.ID, "Shocker", "Creature — Insect", oldB766Shocker, 1, 1)
		effectstest.AdvanceToMain(t, g)
		var old []uuid.UUID
		g.WithWriteLock(func() {
			opp.Hand.Cards = nil
			for i := 0; i < hand; i++ {
				id := uuid.New()
				old = append(old, id)
				opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Old Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
			}
		})
		effectstest.DeclareAttack(t, g, opp.ID, shocker)
		effectstest.AdvanceTo(t, g, game.StepCombatDamage)
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Hand.Size() != hand {
			t.Errorf("hand of %d: the damaged player holds %d cards, want %d", hand, opp.Hand.Size(), hand)
		}
		for _, id := range old {
			if opp.Hand.Contains(id) {
				t.Error("an old card is still in the damaged player's hand")
			}
			if z := effectstest.ZoneOf(g, id); z != game.ZoneGraveyard {
				t.Errorf("a discarded card is in %s", z)
			}
		}
		if effectstest.LifeOf(g, opp.ID) != game.StartingLife-1 {
			t.Errorf("life %d, want %d", effectstest.LifeOf(g, opp.ID), game.StartingLife-1)
		}
	}
}

func TestOldB766InterdictCountersTheAbilityLocksItsSourceAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pinger := effectstest.PushCatalogPermanent(g, me.ID, "Prodigal Sorcerer", "Creature — Human Wizard", oldB766Sorcerer, false)
	twin := effectstest.PushCatalogPermanent(g, me.ID, "Prodigal Sorcerer", "Creature — Human Wizard", oldB766Sorcerer, false)
	effectstest.AdvanceToMain(t, g)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, pinger, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("activate the Sorcerer: %v", err)
	}
	ability := oldB607ActivatedItemOf(g, pinger)
	if ability == uuid.Nil {
		t.Fatal("the ability is not on the stack")
	}
	life := effectstest.LifeOf(g, opp.ID)
	hand := me.Hand.Size()
	effectstest.FloatForTest(g, me, "CU")
	effectstest.CastFromHandForTest(t, g, me, "Interdict", "Instant", "{1}{U}", oldB766Interdict, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(ability)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, opp.ID) != life {
		t.Error("the countered ability still dealt damage")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the draw)", me.Hand.Size(), hand+1)
	}
	effectstest.UntapForTest(g, pinger)
	if err := g.ActivateCatalogAbility(me.ID, pinger, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Error("the countered ability's source activated again this turn")
	}
	if err := g.ActivateCatalogAbility(me.ID, twin, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Errorf("another permanent was locked too: %v", err)
	}
}

func TestOldB766InterdictCannotTargetAnAbilityFromAnotherKindOfSource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spec, _ := effects.Lookup(oldB766Interdict)
	if spec.Targets == nil || !spec.Targets.Abilities {
		t.Fatal("Interdict does not target abilities")
	}
	pred := effects.FromAnArtifactCreatureEnchantmentOrLandSource()
	for _, tc := range []struct {
		typeLine string
		kind     game.StackItemKind
		want     bool
	}{
		{"Artifact", game.StackItemActivated, true},
		{"Creature — Bear", game.StackItemActivated, true},
		{"Enchantment", game.StackItemActivated, true},
		{"Land", game.StackItemActivated, true},
		{"Planeswalker — Jace", game.StackItemActivated, false},
		{"Creature — Bear", game.StackItemTriggered, false},
	} {
		src := effectstest.B12Permanent(g, me.ID, "Source", tc.typeLine)
		item := &game.StackItem{Kind: tc.kind, SourceCardID: src}
		if got := pred(g, me.ID, item); got != tc.want {
			t.Errorf("%s %v: legal = %v, want %v", tc.typeLine, tc.kind, got, tc.want)
		}
	}
}

func oldB766SacredGuideTable(t *testing.T, g *game.Game, me *game.Player, colors ...string) (guide uuid.UUID, library []uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { me.Library.Cards = nil })
	for i := len(colors) - 1; i >= 0; i-- { // the last pushed is the top
		id := uuid.New()
		c := game.Card{InstanceID: id, Name: "Card " + colors[i], TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
		if colors[i] != "" {
			c.Colors = []string{colors[i]}
			c.ManaCost = "{" + colors[i] + "}"
		}
		me.Library.PushTop(c)
		library = append([]uuid.UUID{id}, library...)
	}
	guide = effectstest.B12Push(g, me.ID, "Sacred Guide", "Creature — Human Cleric", oldB766SacredGuide, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CW")
	return guide, library
}

func TestOldB766SacredGuideTakesTheFirstWhiteCardAndExilesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	guide, lib := oldB766SacredGuideTable(t, g, me, "B", "G", "W", "W", "U")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, guide, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Sacred Guide: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(guide) {
		t.Error("Sacred Guide was not sacrificed")
	}
	if !me.Hand.Contains(lib[2]) || me.Hand.Size() != hand+1 {
		t.Error("the first white card is not the one card in hand")
	}
	for _, i := range []int{0, 1} {
		if z := effectstest.ZoneOf(g, lib[i]); z != game.ZoneExile {
			t.Errorf("a non-white revealed card is in %s, want exile", z)
		}
	}
	for _, i := range []int{3, 4} {
		if z := effectstest.ZoneOf(g, lib[i]); z != game.ZoneLibrary {
			t.Errorf("a card below the hit is in %s, want the library", z)
		}
	}
}

func TestOldB766SacredGuideWithNoWhiteCardExilesTheWholeLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	guide, lib := oldB766SacredGuideTable(t, g, me, "B", "G", "")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, guide, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Sacred Guide: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: nothing is put into it", me.Hand.Size(), hand)
	}
	for _, id := range lib {
		if z := effectstest.ZoneOf(g, id); z != game.ZoneExile {
			t.Errorf("a revealed card is in %s, want exile", z)
		}
	}
}

func oldB766BoobyTrapEnter(t *testing.T, g *game.Game, me, chosen *game.Player) uuid.UUID {
	t.Helper()
	trap := effectstest.CastCatalogSpell(t, g, "Booby Trap", "Artifact", oldB766BoobyTrap, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, chosen)
	effectstest.ActNameFor(t, g, me.ID, "Lightning Bolt")
	return trap
}

func TestOldB766BoobyTrapStoresAnOpponentAndANonBasicName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	trap := oldB766BoobyTrapEnter(t, g, me, them)
	c := effectstest.FindBattlefieldCardForTest(g, trap)
	if c.ChosenPlayer != them.ID || c.ChosenName != "Lightning Bolt" {
		t.Errorf("chose player %v and name %q, want %v and Lightning Bolt", c.ChosenPlayer, c.ChosenName, them.ID)
	}
}

func TestOldB766BoobyTrapSacrificesAndDealsTenWhenTheChosenPlayerDrawsTheName(t *testing.T) {
	for _, tc := range []struct {
		name   string
		drawer int // seat offset from the controller; 2 is the chosen player
		card   string
		fires  bool
	}{
		{"the chosen player draws the chosen name", 2, "Lightning Bolt", true},
		{"the chosen player draws another card", 2, "Shock", false},
		{"another player draws the chosen name", 3, "Lightning Bolt", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			them := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
			drawer := g.Seats[(g.Turn.ActiveSeat+tc.drawer)%len(g.Seats)]
			trap := oldB766BoobyTrapEnter(t, g, me, them)
			life := effectstest.LifeOf(g, drawer.ID)
			g.WithWriteLock(func() {
				drawer.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: tc.card, TypeLine: "Instant", Owner: drawer.ID, Controller: drawer.ID})
			})
			if err := g.DrawCard(drawer.ID); err != nil {
				t.Fatalf("DrawCard: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := !g.Battlefield.Contains(trap); got != tc.fires {
				t.Errorf("Booby Trap left the battlefield = %v, want %v", got, tc.fires)
			}
			want := life
			if tc.fires {
				want -= 10
			}
			if got := effectstest.LifeOf(g, drawer.ID); got != want {
				t.Errorf("the drawer's life is %d, want %d", got, want)
			}
		})
	}
}
