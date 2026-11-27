package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b576_cards_test.go — Ravnica: City of Guilds cards, S82 OLD batch
// 576 (#5600), played through the real catalog entries.

const (
	o576Banshee    = "56b69e46-a1e4-4545-99e2-ee4385b7e429"
	o576Sentry     = "66742e88-88cd-4c89-b04b-806577fe5f67"
	o576Augermage  = "7f08c592-8ddd-4f52-95c3-5649cbb24d3a"
	o576Spawnbrokr = "ae050bc8-f289-4484-a07f-dfe0f0a962dc"
	o576Mob        = "af55096c-af6b-479f-8257-144e596e89d8"
	o576Hierophant = "ccaea194-4b21-4485-ba08-d8d030c9418d"
	o576Dismisser  = "e6ed8820-4276-4ead-ba82-4654a92c415d"
	o576CreatureTL = "Creature — Bear"
)

func TestS82OldB576CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o576Banshee: "Keening Banshee", o576Sentry: "Molten Sentry", o576Augermage: "Sadistic Augermage",
		o576Spawnbrokr: "Spawnbroker", o576Mob: "Vindictive Mob", o576Hierophant: "Stone-Seeder Hierophant",
		o576Dismisser: "Vedalken Dismisser",
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

func TestS82OldB576KeeningBansheeFliesAndShrinksATargetOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", o576CreatureTL, 4, 4)
	banshee := effectstest.CastCatalogSpell(t, g, "Keening Banshee", "Creature — Spirit", o576Banshee, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, victim), effectstest.EffectiveToughness(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("the victim is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, banshee, "flying") {
		t.Error("the Banshee has no flying")
	}
}

func TestS82OldB576MoltenSentryEntersAsFiveTwoHasteOrTwoFiveDefender(t *testing.T) {
	seen := map[string]int{}
	for i := 0; i < 40; i++ {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.CastCatalogSpell(t, g, "Molten Sentry", "Creature — Elemental", o576Sentry, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(id) {
			t.Fatal("the Sentry did not resolve")
		}
		p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id)
		haste := effectstest.HasKeywordOnBattlefield(t, g, id, "haste")
		defender := effectstest.HasKeywordOnBattlefield(t, g, id, "defender")
		switch {
		case p == 5 && tough == 2 && haste && !defender:
			seen["heads"]++
		case p == 2 && tough == 5 && defender && !haste:
			seen["tails"]++
		default:
			t.Fatalf("flip %d: the Sentry is %d/%d, haste=%v defender=%v", i+1, p, tough, haste, defender)
		}
	}
	if seen["heads"] == 0 || seen["tails"] == 0 {
		t.Errorf("outcomes over 40 flips: %v, want both", seen)
	}
}

func TestS82OldB576SadisticAugermageMakesEveryPlayerPutACardBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aug := effectstest.B12Push(g, me.ID, "Sadistic Augermage", "Creature — Human Wizard", o576Augermage, 3, 1)
	for _, p := range g.Seats {
		for p.Hand.Size() > 0 {
			g.WithWriteLock(func() { _, _ = p.Hand.PopTop() })
		}
	}
	mine1 := effectstest.PushCatalogHandCard(me, "Mine One", "Sorcery", "")
	mine2 := effectstest.PushCatalogHandCard(me, "Mine Two", "Sorcery", "")
	theirs := effectstest.PushCatalogHandCard(opp, "Their Only", "Sorcery", "")
	effectstest.UsgB3Destroy(t, g, aug)
	effectstest.PassPriorityAroundTable(t, g)
	if prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); prompt != nil {
		if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{mine2}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	var myTop, theirTop uuid.UUID
	g.ReadSnapshot(func() {
		c, _ := me.Library.Top()
		myTop = c.InstanceID
		c, _ = opp.Library.Top()
		theirTop = c.InstanceID
	})
	if myTop != mine2 || !me.Hand.Contains(mine1) {
		t.Errorf("my chosen card is not on top of my library (top %v) or the other left my hand", myTop)
	}
	if theirTop != theirs || opp.Hand.Contains(theirs) {
		t.Error("the opponent's only card was not put on top of their library")
	}
}

func TestS82OldB576SpawnbrokerExchangesControlOfTwoCreatures(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
		mine := effectstest.B12Creature(g, me.ID, "My Bear", o576CreatureTL, 3, 3)
		small := effectstest.B12Creature(g, opp.ID, "Their Small Bear", o576CreatureTL, 3, 1)
		big := effectstest.B12Creature(g, opp.ID, "Their Big Bear", o576CreatureTL, 4, 4)
		effectstest.CastCatalogSpell(t, g, "Spawnbroker", "Creature — Human Wizard", o576Spawnbrokr, nil)
		effectstest.PassPriorityAroundTable(t, g)
		first := effectstest.LatestPickTarget(g, me.ID)
		if first == nil {
			t.Fatal("no targets were asked for")
		}
		if err := g.ResolvePickTargets(first.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}}); err != nil {
			t.Fatalf("ResolvePickTargets (my creature): %v", err)
		}
		second := effectstest.LatestPickTarget(g, me.ID)
		if second == nil {
			t.Fatal("the second target was not asked for")
		}
		if err := g.ResolvePickTargets(second.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: big, Slot: 1}}); err == nil {
			t.Fatal("a creature with greater power was accepted as the second target")
		}
		if err := g.ResolvePickTargets(second.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: small, Slot: 1}}); err != nil {
			t.Fatalf("ResolvePickTargets (equal power): %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, accept)
		effectstest.PassPriorityAroundTable(t, g)
		mc, _ := effectstest.BattlefieldCard(g, mine)
		sc, _ := effectstest.BattlefieldCard(g, small)
		if accept && (mc.Controller != opp.ID || sc.Controller != me.ID) {
			t.Errorf("accepted: my bear is under %v and theirs under %v, want them swapped", mc.Controller, sc.Controller)
		}
		if !accept && (mc.Controller != me.ID || sc.Controller != opp.ID) {
			t.Error("declined: control changed")
		}
	}
}

func TestS82OldB576VindictiveMobSacrificesACreatureAndSaprolingsCantBlockIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", o576CreatureTL, 1, 1)
	mob := effectstest.CastCatalogSpell(t, g, "Vindictive Mob", "Creature — Human Berserker", o576Mob, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(mob) {
		t.Error("the Mob's trigger did not sacrifice the creature chosen")
	}
	attacker := effectstest.B12Push(g, me.ID, "Vindictive Mob", "Creature — Human Berserker", o576Mob, 5, 5)
	sap := effectstest.TypedCreature(g, opp.ID, "Saproling", "Creature — Saproling", "", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", o576CreatureTL, 2, 2)
	effectstest.AttackIntoBlocks(t, g, opp.ID, attacker)
	if r := effectstest.BlockRefusal(t, g, sap, attacker); r == "" {
		t.Error("a Saproling blocked the Mob")
	}
	if r := effectstest.BlockRefusal(t, g, bear, attacker); r != "" {
		t.Errorf("a Bear may not block the Mob: %q", r)
	}
}

func TestS82OldB576StoneSeederHierophantUntapsItselfOnLandfallAndUntapsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hiero := effectstest.B12Push(g, me.ID, "Stone-Seeder Hierophant", "Creature — Human Druid", o576Hierophant, 1, 1)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.RavB2MustActivate(t, g, me.ID, hiero, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(land)})
	if effectstest.TappedForTest(t, g, land) {
		t.Error("the land is still tapped")
	}
	if !effectstest.TappedForTest(t, g, hiero) {
		t.Fatal("the ability does not tap the Hierophant")
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedForTest(t, g, hiero) {
		t.Error("landfall did not untap the Hierophant")
	}
}

func TestS82OldB576VedalkenDismisserPutsACreatureOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", o576CreatureTL, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Vedalken Dismisser", "Creature — Vedalken Wizard", o576Dismisser, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	var top uuid.UUID
	g.ReadSnapshot(func() {
		c, _ := opp.Library.Top()
		top = c.InstanceID
	})
	if g.Battlefield.Contains(victim) || top != victim {
		t.Errorf("the creature is not on top of its owner's library (top %v)", top)
	}
}
