package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b120_cards_test.go — S82 OLD batch 120 (#5512, Ikoria: Lair of
// Behemoths): Patagia Tiger, Blitz Leech, Lavabrink Venturer, Footfall Crater,
// Solid Footing, Mystic Subdual, Lukka, Coppercoat Outcast, Vivien, Monsters'
// Advocate, Alert Heedbonder and Humble Naturalist. Each test plays the real
// catalog entry through the engine.

const (
	oldB120PatagiaTiger   = "256ab710-4210-4142-973f-22ca029d5474"
	oldB120BlitzLeech     = "7679942f-fcbc-4750-8e8c-fe0ad8103c11"
	oldB120Venturer       = "ea66ffa3-d084-434f-91a6-fdff49806285"
	oldB120FootfallCrater = "ca4fdf89-a74e-4086-8109-b0f9dd91710e"
	oldB120SolidFooting   = "edc91e1a-a305-4c14-9d85-077a410528fa"
	oldB120MysticSubdual  = "fdd423d5-f19b-4bb2-98ca-75cc3c058d32"
	oldB120Lukka          = "1ca97394-4e8c-4698-bea1-554f9b10927b"
	oldB120Vivien         = "7256b03e-4fd5-4a0a-9b20-7b88870edfec"
	oldB120Heedbonder     = "683a06b6-af77-440f-81b9-8958597a7324"
	oldB120Naturalist     = "ff3a69a7-a244-4958-9e8b-d153942c1ba6"
)

func TestOldB120EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB120PatagiaTiger:   "Patagia Tiger",
		oldB120BlitzLeech:     "Blitz Leech",
		oldB120Venturer:       "Lavabrink Venturer",
		oldB120FootfallCrater: "Footfall Crater",
		oldB120SolidFooting:   "Solid Footing",
		oldB120MysticSubdual:  "Mystic Subdual",
		oldB120Lukka:          "Lukka, Coppercoat Outcast",
		oldB120Vivien:         "Vivien, Monsters' Advocate",
		oldB120Heedbonder:     "Alert Heedbonder",
		oldB120Naturalist:     "Humble Naturalist",
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

// oldB120Target answers the trigger's target prompt when one is open, then
// lets the trigger resolve.
func oldB120Target(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	if p := effectstest.LatestPickTarget(g, chooser); p != nil {
		effectstest.PickCard(t, g, chooser, target)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// oldB120Library pushes cards onto a player's library, the last one on top.
func oldB120Library(g *game.Game, p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(cards))
	g.WithWriteLock(func() {
		for _, c := range cards {
			c.InstanceID = uuid.New()
			c.Owner, c.Controller = p.ID, p.ID
			p.Library.PushTop(c)
			ids = append(ids, c.InstanceID)
		}
	})
	return ids
}

func oldB120Filler(n int) []game.Card {
	out := make([]game.Card, n)
	for i := range out {
		out[i] = game.Card{Name: "Filler", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, ManaCost: "{G}"}
	}
	return out
}

// --- Patagia Tiger ---------------------------------------------------------

func TestOldB120PatagiaTigerPumpsATargetHumanYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	human := effectstest.B12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirHuman := effectstest.B12Creature(g, opp.ID, "Their Human", "Creature — Human", 2, 2)
	tiger := effectstest.RfdCast(t, g, "Patagia Tiger", "Creature — Cat", oldB120PatagiaTiger, 3, 4)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, bear) || effectstest.HasID(p.PickTargetCards, theirHuman) || effectstest.HasID(p.PickTargetCards, tiger) {
			t.Errorf("only a Human you control is a legal target: %v", p.PickTargetCards)
		}
	}
	oldB120Target(t, g, me.ID, human)
	if p, to := effectstest.PtOf(t, g, human); p != 4 || to != 4 {
		t.Errorf("your Human is %d/%d, want 4/4", p, to)
	}
	for name, id := range map[string]uuid.UUID{"your Bear": bear, "their Human": theirHuman} {
		if p, to := effectstest.PtOf(t, g, id); p != 2 || to != 2 {
			t.Errorf("%s is %d/%d, want 2/2", name, p, to)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, tiger, "flying") {
		t.Error("the Tiger has no flying")
	}
}

func TestOldB120PatagiaTigerWithNoHumanDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	tiger := effectstest.RfdCast(t, g, "Patagia Tiger", "Creature — Cat", oldB120PatagiaTiger, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.PtOf(t, g, bear); p != 2 || to != 2 {
		t.Errorf("the Bear is %d/%d, want 2/2", p, to)
	}
	if p, to := effectstest.PtOf(t, g, tiger); p != 3 || to != 4 {
		t.Errorf("the Tiger is %d/%d, want 3/4 (it is not a Human)", p, to)
	}
}

// --- Blitz Leech -------------------------------------------------------------

func TestOldB120BlitzLeechShrinksAndStripsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 3, 3)
	effectstest.PutCounters(t, g, big, game.CounterPlusOne, 2)
	effectstest.PutCounters(t, g, big, "shield", 1)
	leech := effectstest.RfdCast(t, g, "Blitz Leech", "Creature — Leech", oldB120BlitzLeech, 5, 2)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Errorf("a creature you control is not a legal target: %v", p.PickTargetCards)
	}
	oldB120Target(t, g, me.ID, big)
	if p, to := effectstest.PtOf(t, g, big); p != 1 || to != 1 {
		t.Errorf("the target is %d/%d, want 1/1 (3/3, -2/-2, no counters)", p, to)
	}
	if n := effectstest.B12Counter(t, g, big, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters left, want 0", n)
	}
	if n := effectstest.B12Counter(t, g, big, "shield"); n != 0 {
		t.Errorf("%d shield counters left, want 0", n)
	}
	if p, to := effectstest.PtOf(t, g, mine); p != 3 || to != 3 {
		t.Errorf("your Bear is %d/%d, want 3/3", p, to)
	}
	if spec, _ := effects.Lookup(oldB120BlitzLeech); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Blitz Leech keywords %v, want flash", spec.PrintedKeywords)
	}
	if !effectstest.HasEffectiveKeyword(t, g, leech, "flash") {
		t.Error("the Leech has no flash")
	}
}

func TestOldB120BlitzLeechStripsTheCountersBeforeTheCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, small, game.CounterPlusOne, 1)
	effectstest.RfdCast(t, g, "Blitz Leech", "Creature — Leech", oldB120BlitzLeech, 5, 2)
	oldB120Target(t, g, me.ID, small)
	// 3/3 with one counter, -2/-2 and the counter gone is 0/0 and dies.
	if effectstest.ZoneOf(g, small) == game.ZoneBattlefield {
		t.Error("the creature survived -2/-2 with its counter removed")
	}
}

// --- Lavabrink Venturer ------------------------------------------------------

func TestOldB120LavabrinkVenturerHasProtectionFromTheChosenParity(t *testing.T) {
	for _, tc := range []struct {
		choice              string
		oddBlocked, evenOK  bool
		evenBlocked, oddOK_ bool
	}{
		{choice: "odd", oddBlocked: true, evenOK: true},
		{choice: "even", evenBlocked: true, oddOK_: true},
	} {
		t.Run(tc.choice, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			odd := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "One Drop",
				TypeLine: "Creature — Bear", ManaCost: "{R}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
			even := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Two Drop",
				TypeLine: "Creature — Bear", ManaCost: "{1}{R}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
			zero := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Token Bear",
				TypeLine: "Token Creature — Bear", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
			venturer := effectstest.RfdCast(t, g, "Lavabrink Venturer", "Creature — Human Soldier", oldB120Venturer, 3, 3)
			if len(g.PendingChoices) == 0 {
				t.Fatal("the Venturer asked nothing as it entered")
			}
			effectstest.BlbSTPickOption(t, g, me.ID, tc.choice)
			effectstest.PassPriorityAroundTable(t, g)

			dealt := func(src uuid.UUID) int {
				before := effectstest.DamageMarkedOn(g, venturer)
				g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, venturer, 1) })
				return effectstest.DamageMarkedOn(g, venturer) - before
			}
			wantOdd, wantEven := 1, 1
			if tc.choice == "odd" {
				wantOdd = 0
			} else {
				wantEven = 0
			}
			if got := dealt(odd); got != wantOdd {
				t.Errorf("a mana value 1 source dealt %d, want %d", got, wantOdd)
			}
			if got := dealt(even); got != wantEven {
				t.Errorf("a mana value 2 source dealt %d, want %d", got, wantEven)
			}
			// A token with no cost is mana value 0, which is even.
			if got := dealt(zero); got != wantEven {
				t.Errorf("a mana value 0 source dealt %d, want %d (zero is even)", got, wantEven)
			}
		})
	}
}

// --- Footfall Crater ---------------------------------------------------------

func TestOldB120FootfallCraterGivesTheLandATrampleAndHasteAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	land := effectstest.GaPut(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Footfall Crater", effectstest.AuraTypeLine, oldB120FootfallCrater, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("Footfall Crater enchanted a creature")
	}
	effectstest.PushAuraOnLand(t, g, me.ID, "Footfall Crater", oldB120FootfallCrater, land)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, land)
	if idx < 0 {
		t.Fatal("the enchanted land has no granted ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err == nil {
		t.Error("the ability was activated without a target")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"trample", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the targeted creature has no %s", kw)
		}
	}
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the land did not tap for its ability")
	}
}

func TestOldB120FootfallCraterCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.CycleFromHand(t, g, "Footfall Crater", effectstest.AuraTypeLine, oldB120FootfallCrater, "{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	before := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", before, got)
	}
}

// --- Solid Footing -----------------------------------------------------------

func TestOldB120SolidFootingPumpsAndDamagesByToughnessOnlyWithVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	vigilant := effectstest.PushTribalCreature(g, me.ID, "Vigilant Wall", "Creature — Wall", 0, 4, "vigilance")
	plain := effectstest.PushTribalCreature(g, me.ID, "Plain Wall", "Creature — Wall", 0, 4)
	effectstest.CastCatalogSpell(t, g, "Solid Footing", effectstest.AuraTypeLine, oldB120SolidFooting, effectstest.B16TargetCard(vigilant))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Solid Footing", effectstest.AuraTypeLine, oldB120SolidFooting, effectstest.B16TargetCard(plain))
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"vigilant": vigilant, "plain": plain} {
		if p, to := effectstest.PtOf(t, g, id); p != 1 || to != 5 {
			t.Errorf("the %s wall is %d/%d, want 1/5", name, p, to)
		}
	}
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, vigilant, plain)
	effectstest.PassPriorityAroundTable(t, g)
	// The vigilant wall assigns its toughness (5), the other its power (1).
	if want := before - 6; opp.Life != want {
		t.Errorf("opponent at %d, want %d", opp.Life, want)
	}
}

func TestOldB120SolidFootingHasFlash(t *testing.T) {
	spec, ok := effects.Lookup(oldB120SolidFooting)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Solid Footing keywords %v, want flash", spec.PrintedKeywords)
	}
}

// --- Mystic Subdual ----------------------------------------------------------

func TestOldB120MysticSubdualShrinksAndSilencesTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	host := effectstest.PushTribalCreature(g, opp.ID, "Dragon", "Creature — Dragon", 4, 4, "flying", "haste")
	effectstest.CastCatalogSpell(t, g, "Mystic Subdual", effectstest.AuraTypeLine, oldB120MysticSubdual, effectstest.B16TargetCard(host))
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.PtOf(t, g, host); p != 2 || to != 4 {
		t.Errorf("the Dragon is %d/%d, want 2/4", p, to)
	}
	for _, kw := range []string{"flying", "haste"} {
		if effectstest.HasEffectiveKeyword(t, g, host, kw) {
			t.Errorf("the Dragon still has %s", kw)
		}
	}
	_ = me
	if spec, _ := effects.Lookup(oldB120MysticSubdual); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Mystic Subdual keywords %v, want flash", spec.PrintedKeywords)
	}
}

// --- Alert Heedbonder --------------------------------------------------------

func TestOldB120AlertHeedbonderGainsLifeForEachVigilantCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	heed := effectstest.B12Push(g, me.ID, "Alert Heedbonder", "Creature — Human Scout", oldB120Heedbonder, 2, 4)
	effectstest.PushTribalCreature(g, me.ID, "Knight", "Creature — Knight", 2, 2, "vigilance")
	effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PushTribalCreature(g, opp.ID, "Their Knight", "Creature — Knight", 2, 2, "vigilance")
	if !effectstest.HasEffectiveKeyword(t, g, heed, "vigilance") {
		t.Fatal("the Heedbonder has no vigilance")
	}
	before := me.Life
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if want := before + 2; me.Life != want {
		t.Errorf("life %d, want %d (the Heedbonder and the Knight)", me.Life, want)
	}
}

// --- Humble Naturalist -------------------------------------------------------

func TestOldB120HumbleNaturalistManaPaysOnlyForCreatureSpells(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Humble Naturalist", "Creature — Human Druid", oldB120Naturalist, true)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("a summoning-sick Naturalist tapped for mana (it has no haste)")
	}
	naturalist := effectstest.PushCatalogPermanent(g, me.ID, "Humble Naturalist", "Creature — Human Druid", oldB120Naturalist, false)
	rats := effectstest.PushCatalogPermanent(g, me.ID, "Crypt Rats", "Creature — Rat", effectstest.CryptRatsOracle, false)
	if err := g.ActivateManaAbility(me.ID, naturalist, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tap the Naturalist for {R}: %v", err)
	}
	cost, _ := game.ParseCost("{R}")
	ratsCard, _ := effectstest.BattlefieldCard(g, rats)
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(ratsCard)) {
		t.Error("Humble Naturalist mana pays an ability: stronger than printed")
	}
	sorcery := effectstest.HandSpell(me, "Red Sorcery", "Sorcery", "{R}")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("Humble Naturalist mana paid for a noncreature spell")
	}
	bear := effectstest.HandSpell(me, "Red Creature", "Creature — Bear", "{R}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Naturalist's {R} for a creature spell: %v", err)
	}
}

// --- Lukka, Coppercoat Outcast -----------------------------------------------

const oldB120LukkaLine = "Legendary Planeswalker — Lukka"

func oldB120PushLukka(t *testing.T, g *game.Game, me *game.Player, loyalty int) uuid.UUID {
	t.Helper()
	return effectstest.PushCatalogWalkerWithLine(g, me.ID, "Lukka, Coppercoat Outcast", oldB120LukkaLine, oldB120Lukka, loyalty)
}

func oldB120Activate(t *testing.T, g *game.Game, me *game.Player, walker uuid.UUID, idx int, targets ...uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(targets...)}); err != nil {
		t.Fatalf("activate ability %d: %v", idx, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB120LukkaPlusOneLetsYouCastExiledCreaturesWhileYouControlALukka(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := oldB120Library(g, me, append(oldB120Filler(3),
		game.Card{Name: "Wild Land", TypeLine: "Land"},
		game.Card{Name: "Wild Sorcery", TypeLine: "Sorcery", ManaCost: "{G}"},
		game.Card{Name: "Wild Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 2, Toughness: 2},
	)...)
	filler, land, sorcery, bear := ids[2], ids[3], ids[4], ids[5]
	lukka := oldB120PushLukka(t, g, me, 5)
	oldB120Activate(t, g, me, lukka, 0)
	for name, id := range map[string]uuid.UUID{"the land": land, "the sorcery": sorcery, "the bear": bear} {
		if effectstest.ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("%s is in %q, want exile", name, effectstest.ZoneOf(g, id))
		}
	}
	if effectstest.ZoneOf(g, filler) != game.ZoneLibrary {
		t.Error("a fourth card was exiled")
	}
	effectstest.B06AddMana(me, "G", "G")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a noncreature card exiled this way was cast: only creature cards gain the ability")
	}
	// Without a Lukka planeswalker the creature card stays out of reach.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, lukka); err != nil {
		t.Fatalf("remove Lukka: %v", err)
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the creature card was cast from exile with no Lukka planeswalker")
	}
	// Any Lukka planeswalker opens it again, not only the one that exiled it.
	effectstest.PushCatalogWalkerWithLine(g, me.ID, "Lukka, Other", oldB120LukkaLine, "", 4)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled creature while controlling a Lukka planeswalker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, bear) != game.ZoneBattlefield {
		t.Errorf("the Bear is in %q, want the battlefield", effectstest.ZoneOf(g, bear))
	}
}

func TestOldB120LukkaMinusTwoExilesYourCreatureAndPutsAGreaterOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Two Drop",
		TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ids := oldB120Library(g, me,
		game.Card{Name: "Bottom Filler", TypeLine: "Land"},
		game.Card{Name: "Three Drop", TypeLine: "Creature — Giant", ManaCost: "{2}{G}", Power: 3, Toughness: 3},
		game.Card{Name: "Another Two", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "A Land", TypeLine: "Land"},
		game.Card{Name: "One Drop", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 1, Toughness: 1},
	)
	filler, three, two, land, one := ids[0], ids[1], ids[2], ids[3], ids[4]
	lukka := oldB120PushLukka(t, g, me, 5)
	if err := g.ActivateCatalogAbility(me.ID, lukka, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Fatal("-2 targeted a creature you do not control")
	}
	oldB120Activate(t, g, me, lukka, 1, mine)
	if z := effectstest.ZoneOf(g, mine); z != game.ZoneExile {
		t.Errorf("the target is in %q, want exile", z)
	}
	if z := effectstest.ZoneOf(g, three); z != game.ZoneBattlefield {
		t.Errorf("the creature with greater mana value is in %q, want the battlefield", z)
	}
	for name, id := range map[string]uuid.UUID{"One Drop": one, "A Land": land, "Another Two": two} {
		if z := effectstest.ZoneOf(g, id); z != game.ZoneLibrary {
			t.Errorf("%s is in %q, want the library (bottom)", name, z)
		}
	}
	if z := effectstest.ZoneOf(g, filler); z != game.ZoneLibrary {
		t.Errorf("the untouched card is in %q, want the library", z)
	}
	// The three revealed cards that did not enter went to the bottom (index 0
	// is the bottom of a library), in a random order.
	var bottom []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range me.Library.Cards[:3] {
			bottom = append(bottom, c.InstanceID)
		}
	})
	for name, id := range map[string]uuid.UUID{"One Drop": one, "A Land": land, "Another Two": two} {
		if !effectstest.HasID(bottom, id) {
			t.Errorf("%s is not among the bottom three cards", name)
		}
	}
	if l := effectstest.B12Counter(t, g, lukka, game.CounterLoyalty); l != 3 {
		t.Errorf("loyalty %d, want 3", l)
	}
}

func TestOldB120LukkaMinusTwoWithNoGreaterCreatureBottomsTheWholeLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big",
		TypeLine: "Creature — Giant", ManaCost: "{5}{G}", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	ids := oldB120Library(g, me,
		game.Card{Name: "Small A", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 1, Toughness: 1},
		game.Card{Name: "Small B", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
	)
	lukka := oldB120PushLukka(t, g, me, 5)
	oldB120Activate(t, g, me, lukka, 1, mine)
	for _, id := range ids {
		if z := effectstest.ZoneOf(g, id); z != game.ZoneLibrary {
			t.Errorf("a revealed card is in %q, want the library", z)
		}
	}
}

func TestOldB120LukkaMinusSevenDealsEachCreaturesPowerToEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 4, 4)
	effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 5)
	// A creature an opponent controls deals nothing.
	effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Giant", "Creature — Giant", 9, 9)
	lukka := oldB120PushLukka(t, g, me, 7)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	oldB120Activate(t, g, me, lukka, 2)
	for _, p := range g.Seats {
		want := before[p.ID]
		if p.ID != me.ID {
			want -= 6
		}
		if p.Life != want {
			t.Errorf("%s is at %d, want %d", p.ID, p.Life, want)
		}
	}
	if effectstest.ZoneOf(g, lukka) != game.ZoneGraveyard {
		t.Errorf("Lukka at 0 loyalty is in %q, want the graveyard", effectstest.ZoneOf(g, lukka))
	}
}

// --- Vivien, Monsters' Advocate ----------------------------------------------

func oldB120PushVivien(t *testing.T, g *game.Game, me *game.Player, loyalty int) uuid.UUID {
	t.Helper()
	return effectstest.PushCatalogWalkerWithLine(g, me.ID, "Vivien, Monsters' Advocate", "Legendary Planeswalker — Vivien", oldB120Vivien, loyalty)
}

func TestOldB120VivienPlusOneMakesABeastWithTheChosenCounter(t *testing.T) {
	for _, tc := range []struct{ label, kind string }{
		{"Reach", game.CounterReach},
		{"Vigilance", game.CounterVigilance},
		{"Trample", game.CounterTrample},
	} {
		t.Run(tc.label, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.ToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			vivien := oldB120PushVivien(t, g, me, 3)
			oldB120Activate(t, g, me, vivien, 0)
			effectstest.BlbSTPickOption(t, g, me.ID, tc.label)
			effectstest.PassPriorityAroundTable(t, g)
			var beast uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Name == "Beast" && c.IsToken() && c.Controller == me.ID {
					beast = c.InstanceID
				}
			}
			if beast == uuid.Nil {
				t.Fatal("no Beast token")
			}
			if p, to := effectstest.PtOf(t, g, beast); p != 3 || to != 3 {
				t.Errorf("the Beast is %d/%d, want 3/3", p, to)
			}
			for _, kind := range []string{game.CounterReach, game.CounterVigilance, game.CounterTrample} {
				want := 0
				if kind == tc.kind {
					want = 1
				}
				if n := effectstest.B12Counter(t, g, beast, kind); n != want {
					t.Errorf("%d %s counters, want %d", n, kind, want)
				}
			}
			if !effectstest.HasEffectiveKeyword(t, g, beast, tc.kind) {
				t.Errorf("the Beast has no %s", tc.kind)
			}
			if l := effectstest.B12Counter(t, g, vivien, game.CounterLoyalty); l != 4 {
				t.Errorf("loyalty %d, want 4", l)
			}
		})
	}
}

func TestOldB120VivienCastsCreatureSpellsFromTheTopOfTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB120PushVivien(t, g, me, 3)
	spec, _ := effects.Lookup(oldB120Vivien)
	if spec.LibraryTopVisible != game.LibraryTopOwner {
		t.Errorf("LibraryTopVisible = %v, want its controller", spec.LibraryTopVisible)
	}
	effectstest.B06AddMana(me, "G", "G")
	ids := oldB120Library(g, me,
		game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{G}"},
	)
	bear, sorcery := ids[0], ids[1]
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a noncreature spell was cast from the top of the library")
	}
	g.WithWriteLock(func() { me.Library.Cards = me.Library.Cards[:len(me.Library.Cards)-1] })
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("cast the creature on top: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOf(g, bear); z != game.ZoneBattlefield {
		t.Errorf("the Bear is in %q, want the battlefield", z)
	}
}

func TestOldB120VivienMinusTwoSearchesForALesserCreatureWhenYouNextCastOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vivien := oldB120PushVivien(t, g, me, 3)
	ids := oldB120Library(g, me,
		game.Card{Name: "Five Drop", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Power: 5, Toughness: 5},
		game.Card{Name: "Three Drop", TypeLine: "Creature — Giant", ManaCost: "{2}{G}", Power: 3, Toughness: 3},
		game.Card{Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Some Land", TypeLine: "Land"},
	)
	five, three, two := ids[0], ids[1], ids[2]
	oldB120Activate(t, g, me, vivien, 1)
	// A noncreature spell does not use it up.
	effectstest.B06AddMana(me, "G", "G", "G")
	sorcery := effectstest.HandSpell(me, "A Sorcery", "Sorcery", "{G}")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast a sorcery: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("a noncreature spell triggered the search")
	}
	effectstest.B06AddMana(me, "G", "G")
	creature := effectstest.HandSpell(me, "Three Mana Bear", "Creature — Bear", "{2}{G}")
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast a creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("no search prompt for the creature card with lesser mana value")
	}
	// Only the card with mana value below 3 is offered.
	c := effectstest.SearchChoiceFor(g, me.ID)
	if effectstest.SearchOptionNamed(g, c, "Two Drop") == uuid.Nil {
		t.Error("the Two Drop is not offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Three Drop") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Five Drop") != uuid.Nil {
		t.Error("a creature with mana value 3 or more is offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Two Drop")
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOf(g, two); z != game.ZoneBattlefield {
		t.Errorf("the Two Drop is in %q, want the battlefield", z)
	}
	if effectstest.ZoneOf(g, three) != game.ZoneLibrary || effectstest.ZoneOf(g, five) != game.ZoneLibrary {
		t.Error("the other creature cards left the library")
	}
	// It fires once: a second creature spell finds nothing more.
	effectstest.B06AddMana(me, "G", "G", "G")
	again := effectstest.HandSpell(me, "Another Bear", "Creature — Bear", "{2}{G}")
	if err := g.CastSpell(me.ID, again, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast a second creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("the delayed trigger fired twice")
	}
}
