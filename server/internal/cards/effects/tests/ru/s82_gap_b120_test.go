package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b120_test.go — S82 set GAP, batch 120 (Tales of Middle-earth
// Commander sorceries, instants and lands), through the real catalog.

const (
	gap120Mordor    = "60fa9f5f-2360-4d5e-b291-1318803e3c9f"
	gap120Lidless   = "287aae84-b098-456e-b9cb-8eac51c36360"
	gap120Courage   = "3f7452aa-b075-47d9-b61f-061d61c7d1b4"
	gap120Mouth     = "7c05d239-39fc-4d34-a853-e3d591f4a235"
	gap120TooGreed  = "af65bbf9-e328-4b8b-8caa-c6dff743867b"
	gap120Urborg    = "b6114962-035e-4e7f-9009-4739bf83a05a"
	gap120Seasons   = "c46738a4-e0d4-41dc-a20d-4800f00cea62"
	gap120Hammer    = "c7476beb-7923-4994-8476-bc69187ecb72"
	gap120Homeward  = "cb8ec2e4-8223-4172-8f2c-37c918a573fa"
	gap120WakeDrake = "cfa5b168-d112-4b92-ac42-8bba30f15d77"
)

func gap120Named(g *game.Game, owner uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == name {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

func gap120Land(g *game.Game, owner uuid.UUID, snow bool) uuid.UUID {
	if snow {
		return effectstest.BlbB1Push(g, owner, "Snow-Covered Swamp", "Basic Snow Land — Swamp", "", 0, 0)
	}
	return effectstest.BlbB1Push(g, owner, "Swamp", "Basic Land — Swamp", "", 0, 0)
}

func gap120ControllerOf(g *game.Game, id uuid.UUID) uuid.UUID {
	var who uuid.UUID
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			who = c.Controller
		}
	})
	return who
}

func gap120Power(g *game.Game, id uuid.UUID) int {
	p := -99
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			p = c.PowerForComparison()
		}
	})
	return p
}

func TestGap120Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		gap120Mordor: "Mordor on the March", gap120Lidless: "Lidless Gaze",
		gap120Courage: "Courageous Resolve", gap120Mouth: "Mouth of Ronom",
		gap120TooGreed: "Too Greedily, Too Deep", gap120Urborg: "Urborg",
		gap120Seasons: "Seasons Past", gap120Hammer: "Hammerheim",
		gap120Homeward: "Homeward Path", gap120WakeDrake: "Wake the Dragon",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestGap120CourageousResolveGrantsProtectionDrawsAndHasFatefulHour(t *testing.T) {
	for _, life := range []int{5, 6} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		src := effectstest.BlbB1Push(g, opp.ID, "Attacker", "Creature — Bear", "", 2, 2)
		hand := me.Hand.Size()
		me.Life = life
		effectstest.CastCatalogSpell(t, g, "Courageous Resolve", "Instant", gap120Courage, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.P1g2HasKeyword(g, bear, game.ProtectionFromEachOfYourOpponents) {
			t.Errorf("life %d: the creature has no protection from each of your opponents", life)
		}
		if got := me.Hand.Size(); got != hand+1 {
			t.Errorf("life %d: hand %d -> %d, want +1 (pushed, cast, drew)", life, hand, got)
		}
		// Damage that would take you to 0 or below.
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 9) })
		if life == 5 && me.Life != 5 {
			t.Errorf("fateful hour (life 5): life %d after 9 damage, want it unchanged", me.Life)
		}
		if life == 6 && me.Life != -3 {
			t.Errorf("no fateful hour (life 6): life %d after 9 damage, want -3", me.Life)
		}
		// Gaining life is still allowed.
		if life == 5 {
			effectstest.GainLifeNow(t, g, src, me, 2)
			if me.Life != 7 {
				t.Errorf("fateful hour: gaining 2 left life at %d, want 7", me.Life)
			}
		}
	}
}

func TestGap120CourageousResolveWorksWithNoTargetAndStopsTheLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	hand := me.Hand.Size()
	me.Life = 3
	effectstest.CastCatalogSpell(t, g, "Courageous Resolve", "Instant", gap120Courage, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d with no target, want the draw anyway", hand, got)
	}
	// A "lose life" effect does nothing either.
	effectstest.GainLifeNow(t, g, opp.ID, me, -5)
	if me.Life != 3 {
		t.Errorf("life %d after losing 5, want 3 (you can't lose life this turn)", me.Life)
	}
}

func TestGap120LidlessGazeExilesEachTopCardForYouToPlayAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	tops := map[uuid.UUID]*game.Player{}
	for _, p := range g.Seats {
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: "Top of " + p.ID.String(), TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		tops[id] = p
	}
	effectstest.CastCatalogSpell(t, g, "Lidless Gaze", "Sorcery", gap120Lidless, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for id, p := range tops {
		if !g.Exile.Contains(id) {
			t.Errorf("the top card of %s's library was not exiled", p.ID)
		}
		if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, id), me.ID) {
			t.Errorf("you may not play the exiled card of %s", p.ID)
		}
	}
	_, _ = opp, other
	// Flashback is {2}{B}{R}: refused without the mana, allowed with it.
	fb := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: fb, Name: "Lidless Gaze", TypeLine: "Sorcery", OracleID: gap120Lidless, ManaCost: "{2}{B}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, fb, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err == nil {
		t.Error("flashback cast with no mana")
	}
	effectstest.B06AddMana(me, "B", "R", "G", "G")
	if err := g.CastSpell(me.ID, fb, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(fb) {
		t.Error("the flashed back spell was not exiled")
	}
}

func TestGap120MordorOnTheMarchCopiesAGraveyardCreatureWithHasteAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	dead := effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	spec, _ := effects.Lookup(gap120Mordor)
	if len(spec.Triggered) != 1 {
		t.Fatalf("want the storm trigger, got %d triggers", len(spec.Triggered))
	}
	effectstest.CastCatalogSpell(t, g, "Mordor on the March", "Sorcery", gap120Mordor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no prompt to choose a creature card")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) {
		t.Error("the creature card was not exiled")
	}
	toks := gap120Named(g, me.ID, "Dead Bear")
	if len(toks) != 1 {
		t.Fatalf("%d Dead Bear tokens, want 1", len(toks))
	}
	if !effectstest.P1g2HasKeyword(g, toks[0], "haste") {
		t.Error("the token has no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if len(gap120Named(g, me.ID, "Dead Bear")) != 0 {
		t.Error("the token was not exiled at the next end step")
	}
}

func TestGap120MouthOfRonomDamagesForSnowMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mouth := effectstest.PushCatalogPermanent(g, me.ID, "Mouth of Ronom", "Snow Land", gap120Mouth, false)
	victim := effectstest.BlbB1Push(g, opp.ID, "Victim", "Creature — Bear", "", 4, 4)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 5; i++ {
		gap120Land(g, me.ID, false)
	}
	if err := g.ActivateCatalogAbility(me.ID, mouth, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Fatal("{4}{S} paid with nonsnow lands: accepted")
	}
	snow := gap120Land(g, me.ID, true)
	if err := g.ActivateCatalogAbility(me.ID, mouth, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	_ = snow
	if g.Battlefield.Contains(mouth) {
		t.Error("the land was not sacrificed")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the 4/4 survived 4 damage")
	}
}

func TestGap120TooGreedilyReanimatesAndHitsEveryOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	big := effectstest.PushGraveyardPermanent(opp, "Their Dead Giant", "Creature — Giant", "{4}")
	// Give it power through the card record.
	g.WithWriteLock(func() {
		for i := range opp.Graveyard.Cards {
			if opp.Graveyard.Cards[i].InstanceID == big {
				opp.Graveyard.Cards[i].Power, opp.Graveyard.Cards[i].Toughness = 3, 3
			}
		}
	})
	mine := effectstest.BlbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Too Greedily, Too Deep", "Sorcery", gap120TooGreed, []game.TargetRef{{Kind: game.TargetCard, ID: big}})
	effectstest.PassPriorityAroundTable(t, g)
	if len(gap120Named(g, me.ID, "Their Dead Giant")) != 1 {
		t.Fatal("the creature did not enter under your control")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("the 3 damage did not kill the other creatures")
	}
	if len(gap120Named(g, me.ID, "Their Dead Giant")) != 1 {
		t.Error("the reanimated creature hit itself")
	}
}

func TestGap120UrborgLosesTheChosenKeyword(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	urborg := effectstest.PushCatalogPermanent(g, me.ID, "Urborg", "Legendary Land", gap120Urborg, false)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Keywords: []string{"first strike", "swampwalk"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, urborg, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: walker}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == me.ID {
			if err := g.ResolveOptionPick(c.ID, me.ID, 1); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.P1g2HasKeyword(g, walker, "swampwalk") {
		t.Error("the creature kept swampwalk")
	}
	if !effectstest.P1g2HasKeyword(g, walker, "first strike") {
		t.Error("the creature lost first strike although swampwalk was chosen")
	}
}

func TestGap120SeasonsPastReturnsCardsWithDifferentManaValuesThenGoesToTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	one := effectstest.PushGraveyardPermanent(me, "One A", "Creature — Bear", "{1}")
	oneB := effectstest.PushGraveyardPermanent(me, "One B", "Creature — Bear", "{1}")
	three := effectstest.PushGraveyardPermanent(me, "Three", "Creature — Bear", "{2}{G}")
	hand := me.Hand.Size()
	spell := effectstest.CastCatalogSpell(t, g, "Seasons Past", "Sorcery", gap120Seasons, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no prompt")
	}
	// Two cards with the same mana value are refused as a set.
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{one, oneB}); err == nil {
		t.Fatal("two cards with the same mana value were accepted")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{one, three}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(one) || !me.Hand.Contains(three) || me.Hand.Contains(oneB) {
		t.Error("the wrong cards came back")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2 (the spell's own card went out and two came back)", hand, got)
	}
	if me.Graveyard.Contains(spell) {
		t.Error("Seasons Past went to the graveyard")
	}
	lib := me.Library.Cards
	if len(lib) == 0 || lib[0].InstanceID != spell {
		t.Error("Seasons Past is not on the bottom of its owner's library")
	}
}

func TestGap120HammerheimStripsEveryLandwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	h := effectstest.PushCatalogPermanent(g, me.ID, "Hammerheim", "Legendary Land", gap120Hammer, false)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Keywords: []string{"islandwalk", "forestwalk", "flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, h, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: walker}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"islandwalk", "forestwalk"} {
		if effectstest.P1g2HasKeyword(g, walker, kw) {
			t.Errorf("the creature kept %s", kw)
		}
	}
	if !effectstest.P1g2HasKeyword(g, walker, "flying") {
		t.Error("the creature lost flying")
	}
}

func TestGap120HomewardPathReturnsEveryCreatureToItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	path := effectstest.PushCatalogPermanent(g, me.ID, "Homeward Path", "Land", gap120Homeward, false)
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, path, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap120ControllerOf(g, stolen); got != opp.ID {
		t.Errorf("the creature is controlled by %v, want its owner", got)
	}
}

func TestGap120WakeTheDragonMakesAStealingDragonAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.CastCatalogSpell(t, g, "Wake the Dragon", "Sorcery", gap120WakeDrake, nil)
	effectstest.PassPriorityAroundTable(t, g)
	dragons := gap120Named(g, me.ID, "Dragon")
	if len(dragons) != 1 {
		t.Fatalf("%d Dragons, want 1", len(dragons))
	}
	for _, kw := range []string{"flying", "menace"} {
		if !effectstest.P1g2HasKeyword(g, dragons[0], kw) {
			t.Errorf("the Dragon lacks %s", kw)
		}
	}
	if gap120Power(g, dragons[0]) != 6 {
		t.Errorf("the Dragon is not 6/6")
	}
	// Its combat damage steals an artifact from the damaged player.
	rock := effectstest.BlbB1Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	other := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	_ = other
	effectstest.DealCombatDamageToPlayer(g, dragons[0], opp.ID, 6)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: rock}, true)
	if got := gap120ControllerOf(g, rock); got != me.ID {
		t.Errorf("the artifact is controlled by %v, want the Dragon's controller", got)
	}
}
