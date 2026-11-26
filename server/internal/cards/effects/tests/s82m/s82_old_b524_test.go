package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eventide graveyard cards (S82 older sets, batch 524): every test plays the
// real catalog entry through the engine.

const (
	oldB524HagHedgeMage = "364e8bd6-52e5-4397-8d2a-10c4706e5a45"
	oldB524Treefolk     = "8a011d0d-da29-4287-b50c-12746abfa38c"
	oldB524CreakGhoul   = "94c47804-bb64-4a91-85c5-9241c39685aa"
	oldB524Evershrike   = "9f148f1f-53c5-477d-ba91-a2415db65adf"
	oldB524Merrow       = "dad36dc0-39aa-46f9-819b-cad7a578c037"
	oldB524Rekindled    = "dd99061d-510d-4935-b44e-54721505515f"
	oldB524Pacifism     = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"
)

func TestOldB524EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB524HagHedgeMage: "Hag Hedge-Mage", oldB524Treefolk: "Tilling Treefolk",
		oldB524CreakGhoul: "Creakwood Ghoul", oldB524Evershrike: "Evershrike",
		oldB524Merrow: "Merrow Bonegnawer", oldB524Rekindled: "Rekindled Flame",
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

// oldB524Drive answers whatever the table is waiting on until the stack is
// empty: trigger order, target picks (with `pick`), "you may" prompts (yes
// when `yes`), and card choices (the first `ChooseMin` candidates).
func oldB524Drive(t *testing.T, g *game.Game, yes bool, pick func(c *game.PendingChoice) []game.TargetRef) {
	t.Helper()
	for i := 0; i < 60; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open == nil {
			if effectstest.StackFullyEmpty(g) {
				return
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			continue
		}
		var err error
		switch open.Kind {
		case game.PendingChoiceTriggerOrder:
			err = g.ResolveTriggerOrder(open.ID, open.Chooser, append([]uuid.UUID(nil), open.TriggerOrderIDs...))
		case game.PendingChoicePickTarget:
			err = g.ResolvePickTargets(open.ID, open.Chooser, pick(open))
		case game.PendingChoiceTriggerPrompt:
			err = g.ResolveTriggerPrompt(open.ID, open.Chooser, yes)
		case game.PendingChoiceChooseCards:
			n := open.ChooseMin
			if yes {
				n = open.ChooseMax
			}
			if n > len(open.ChooseCards) {
				n = len(open.ChooseCards)
			}
			err = g.ResolveChooseCards(open.ID, open.Chooser, append([]uuid.UUID(nil), open.ChooseCards[:n]...))
		default:
			t.Fatalf("an unexpected prompt %s", open.Kind)
		}
		if err != nil {
			t.Fatalf("answering %s: %v", open.Kind, err)
		}
	}
	t.Fatal("the table never settled")
}

func oldB524Cast(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, p, tough int) uuid.UUID {
	t.Helper()
	return effectstest.PlsB1Cast(t, g, name, typeLine, oracle, cost, p, tough)
}

func oldB524Land(g *game.Game, owner uuid.UUID, sub string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: sub, TypeLine: "Basic Land — " + sub, Owner: owner, Controller: owner,
	})
}

func TestOldB524MerrowBonegnawerExilesAGraveyardCardAndUntapsOnABlackSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	merrow := effectstest.B12Push(g, me.ID, "Merrow Bonegnawer", "Creature — Merfolk Rogue", oldB524Merrow, 1, 1)
	theirs := effectstest.PushGraveyardPermanent(opp, "Theirs", "Creature — Test", "{1}")
	if err := g.ActivateCatalogAbility(me.ID, merrow, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("{T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, theirs) != game.ZoneExile {
		t.Error("the targeted player's card was not exiled")
	}
	if !effectstest.ArnTapped(t, g, merrow) {
		t.Fatal("the Merrow did not tap for its ability")
	}
	oldB524Cast(t, g, "Red Spell", "Sorcery", "", "{R}", 0, 0)
	oldB524Drive(t, g, true, nil)
	if !effectstest.ArnTapped(t, g, merrow) {
		t.Fatal("it untapped on a red spell")
	}
	oldB524Cast(t, g, "Black Spell", "Sorcery", "", "{B}", 0, 0)
	oldB524Drive(t, g, true, nil)
	if effectstest.ArnTapped(t, g, merrow) {
		t.Error("it did not untap on a black spell")
	}
}

func TestOldB524RekindledFlameDealsFourAndComesBackWhenAnOpponentHasNoHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Rekindled Flame", "Sorcery", oldB524Rekindled, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-4)
	}
	var flame uuid.UUID
	for _, c := range me.Graveyard.Cards {
		if c.OracleID == oldB524Rekindled {
			flame = c.InstanceID
		}
	}
	if flame == uuid.Nil {
		t.Fatal("the Flame is not in the graveyard")
	}
	// Every opponent holds cards: nothing triggers.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if n := effectstest.B18TriggerPromptCount(g, me.ID); n != 0 {
		t.Fatalf("%d questions while every opponent holds cards", n)
	}
	// An opponent with no hand: the upkeep trigger offers the return.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	flame = effectstest.PushGraveyardPermanent(me, "Rekindled Flame", "Sorcery", "{2}{R}{R}")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == flame {
			me.Graveyard.Cards[i].OracleID = oldB524Rekindled
		}
	}
	// Walk to the last seat's end step, empty the hand, then on into seat 0's upkeep.
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == 3 && g.Turn.Step == game.StepEnd); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	opp.Hand.Cards = nil
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	oldB524Drive(t, g, true, nil)
	if !me.Hand.Contains(flame) {
		t.Error("Rekindled Flame did not return to its owner's hand")
	}
}

func TestOldB524HagHedgeMageNeedsTwoSwampsToMakeAPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB524Land(g, me.ID, "Swamp")
	oldB524Land(g, me.ID, "Swamp")
	hand := opp.Hand.Size()
	oldB524Cast(t, g, "Hag Hedge-Mage", "Creature — Hag Shaman", oldB524HagHedgeMage, "{2}{B/G}", 2, 2)
	oldB524Drive(t, g, true, func(c *game.PendingChoice) []game.TargetRef {
		return []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
	})
	if opp.Hand.Size() != hand-1 {
		t.Errorf("the target player's hand is %d, want %d", opp.Hand.Size(), hand-1)
	}

	// One Swamp is not enough.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	oldB524Land(g, me.ID, "Swamp")
	hand = opp.Hand.Size()
	oldB524Cast(t, g, "Hag Hedge-Mage", "Creature — Hag Shaman", oldB524HagHedgeMage, "{2}{B/G}", 2, 2)
	oldB524Drive(t, g, true, func(c *game.PendingChoice) []game.TargetRef {
		t.Fatal("a target was asked for with one Swamp")
		return nil
	})
	if opp.Hand.Size() != hand {
		t.Error("a player discarded with only one Swamp")
	}
}

func TestOldB524HagHedgeMagePutsAGraveyardCardOnTopWithTwoForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB524Land(g, me.ID, "Forest")
	oldB524Land(g, me.ID, "Forest")
	card := effectstest.PushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	oldB524Cast(t, g, "Hag Hedge-Mage", "Creature — Hag Shaman", oldB524HagHedgeMage, "{2}{B/G}", 2, 2)
	oldB524Drive(t, g, true, func(c *game.PendingChoice) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: card}}
	})
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != card {
		t.Error("the graveyard card is not on top of the library")
	}
	if effectstest.ZoneOfCard(g, card) != game.ZoneLibrary {
		t.Error("the card did not leave the graveyard")
	}
}

func TestOldB524CreakwoodGhoulExilesAGraveyardCardAndGainsALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := effectstest.B12Push(g, me.ID, "Creakwood Ghoul", "Creature — Plant Zombie", oldB524CreakGhoul, 3, 3)
	theirs := effectstest.PushGraveyardPermanent(opp, "Theirs", "Creature — Test", "{1}")
	life := me.Life
	effectstest.ApaMana(me, "B", "G")
	effectstest.ApaActivate(t, g, me, ghoul, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(theirs)})
	if effectstest.ZoneOfCard(g, theirs) != game.ZoneExile {
		t.Error("the card was not exiled")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestOldB524TillingTreefolkReturnsUpToTwoLandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	l1 := effectstest.PushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	l2 := effectstest.PushGraveyardPermanent(me, "Swamp", "Basic Land — Swamp", "")
	l3 := effectstest.PushGraveyardPermanent(me, "Island", "Basic Land — Island", "")
	bear := effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	oldB524Cast(t, g, "Tilling Treefolk", "Creature — Treefolk Druid", oldB524Treefolk, "{2}{G}", 1, 3)
	oldB524Drive(t, g, true, func(c *game.PendingChoice) []game.TargetRef {
		if effectstest.HasID(c.PickTargetCards, bear) {
			t.Error("a nonland card was offered")
		}
		if c.PickTargetMin != 0 || c.PickTargetMax != 2 {
			t.Errorf("target count %d..%d, want up to two", c.PickTargetMin, c.PickTargetMax)
		}
		return []game.TargetRef{{Kind: game.TargetCard, ID: l1}, {Kind: game.TargetCard, ID: l2}}
	})
	if !me.Hand.Contains(l1) || !me.Hand.Contains(l2) {
		t.Error("the two chosen lands did not return to hand")
	}
	if me.Hand.Contains(l3) || me.Hand.Contains(bear) {
		t.Error("an unchosen card returned")
	}
}

func TestOldB524EvershrikeGrowsForEachAttachedAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shrike := effectstest.B27Push(g, me.ID, "Evershrike", "Creature — Elemental Spirit", oldB524Evershrike, "{3}{W/B}{W/B}", 2, 2, "W", "B")
	effectstest.AssertKeywords(t, g, shrike, "flying")
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Aura", TypeLine: "Enchantment — Aura", Owner: me.ID, Controller: me.ID,
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: shrike},
		})
	}
	if p, tough := effectstest.EffectivePower(t, g, shrike), effectstest.EffectiveToughness(t, g, shrike); p != 6 || tough != 6 {
		t.Errorf("Evershrike is %d/%d, want 6/6 with two Auras", p, tough)
	}
}

func oldB524EvershrikeInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	id := effectstest.PushGraveyardPermanent(me, "Evershrike", "Creature — Elemental Spirit", "{3}{W/B}{W/B}")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == id {
			me.Graveyard.Cards[i].OracleID = oldB524Evershrike
			me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 2, 2
		}
	}
	return id
}

func TestOldB524EvershrikeReturnsWithAnAuraFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shrike := oldB524EvershrikeInGraveyard(g, me)
	aura := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: aura, Name: "Pacifism", TypeLine: "Enchantment — Aura", OracleID: oldB524Pacifism,
		ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "W", "W", "W", "W")
	effectstest.ApaActivate(t, g, me, shrike, 0, game.ActivateAbilityParams{XValue: 2})
	oldB524Drive(t, g, true, nil)
	if !g.Battlefield.Contains(shrike) {
		t.Fatal("Evershrike did not return to the battlefield")
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura was not put onto the battlefield")
	}
	if got := effects.B18AurasAttachedTo(g, shrike); got != 1 {
		t.Errorf("%d Auras attached to Evershrike, want 1", got)
	}
}

func TestOldB524EvershrikeIsExiledWithoutAnAuraOrWhenItIsTooBig(t *testing.T) {
	// No Aura in hand: exiled.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shrike := oldB524EvershrikeInGraveyard(g, me)
	effectstest.ApaMana(me, "W", "W")
	effectstest.ApaActivate(t, g, me, shrike, 0, game.ActivateAbilityParams{XValue: 0})
	oldB524Drive(t, g, true, nil)
	if effectstest.ZoneOfCard(g, shrike) != game.ZoneExile {
		t.Errorf("Evershrike is in %v, want exile", effectstest.ZoneOfCard(g, shrike))
	}

	// An Aura with mana value above X is not offered: exiled.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	shrike = oldB524EvershrikeInGraveyard(g, me)
	aura := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: aura, Name: "Pacifism", TypeLine: "Enchantment — Aura", OracleID: oldB524Pacifism,
		ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "W", "W", "W")
	effectstest.ApaActivate(t, g, me, shrike, 0, game.ActivateAbilityParams{XValue: 1})
	oldB524Drive(t, g, true, nil)
	if effectstest.ZoneOfCard(g, shrike) != game.ZoneExile || !me.Hand.Contains(aura) {
		t.Error("an Aura of mana value 2 was usable with X = 1")
	}
}
