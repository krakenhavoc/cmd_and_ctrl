package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b371_cards_test.go — Journey into Nyx, batch 371 (#5558, S82 OLD).
// Every test plays the real catalog entry through the engine.

const (
	oldB371PullOracle      = "1c0a72dc-4756-4d60-b6b4-4c2c69699efd"
	oldB371FontOracle      = "b8805aa6-fa5f-4c70-ad90-8502611bef5a"
	oldB371RitualOracle    = "c74e7441-7563-4954-a1d3-3c0463294455"
	oldB371MacarOracle     = "1ebafd78-269a-4974-88d2-e75b44b744c2"
	oldB371DiscipleOracle  = "34ace274-8401-4e3c-89cd-75fe2337bf7d"
	oldB371ThiefOracle     = "e01289ea-634b-4be5-8b8e-5e1179761a9b"
	oldB371DictateOracle   = "01cb56d7-f859-46fd-b6c9-40d2c4965b69"
	oldB371ChimeraOracle   = "718d72c4-ad37-49b6-860a-0b921c0de978"
	oldB371WeaverOracle    = "562684ca-6e3c-47df-9e94-8ff772c25a96"
	oldB371IreOracle       = "fb3b365b-eb8c-4695-bfd6-88060ad20741"
	oldB371ChariotOracle   = "5a40339d-7821-4f37-a957-1a6ebdd993e0"
	oldB371DakraOracle     = "63f05280-74de-4fe3-a507-07fb125685c0"
	oldB371BearTypeLine    = "Creature — Bear"
	oldB371EnchantmentLine = "Enchantment"
)

func TestOldB371CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB371PullOracle:     "Pull from the Deep",
		oldB371FontOracle:     "Font of Return",
		oldB371RitualOracle:   "Ritual of the Returned",
		oldB371MacarOracle:    "King Macar, the Gold-Cursed",
		oldB371DiscipleOracle: "Disciple of Deceit",
		oldB371ThiefOracle:    "Daring Thief",
		oldB371DictateOracle:  "Dictate of Karametra",
		oldB371ChimeraOracle:  "Riptide Chimera",
		oldB371WeaverOracle:   "Renowned Weaver",
		oldB371IreOracle:      "Thassa's Ire",
		oldB371ChariotOracle:  "Chariot of Victory",
		oldB371DakraOracle:    "Dakra Mystic",
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

func oldB371Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB371Card(ids ...uuid.UUID) []game.TargetRef {
	var out []game.TargetRef
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func oldB371Float(t *testing.T, g *game.Game, p *game.Player, colored string, generic int) {
	t.Helper()
	if colored != "" {
		effectstest.FloatMana(t, g, p, colored)
	}
	effectstest.B09AddColorless(p, generic)
}

// oldB371Inspire taps the creature and untaps its controller's permanents, the
// way an untap step does, and settles what that triggers.
func oldB371Inspire(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = true })
	if err := g.UntapAll(p.ID); err != nil {
		t.Fatalf("UntapAll: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func oldB371InGraveyard(p *game.Player, name, typeLine string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	return id
}

func TestOldB371ChariotOfVictoryGrantsThreeKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB371Seats(g)
	effectstest.AdvanceToMain(t, g)
	chariot := effectstest.PushCatalogPermanent(g, me.ID, "Chariot of Victory", "Artifact — Equipment", oldB371ChariotOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB371BearTypeLine, 2, 2)
	oldB371Float(t, g, me, "", 1)
	effectstest.EquipTo(t, g, me.ID, chariot, bear)
	effectstest.AssertKeywords(t, g, bear, "first strike", "trample", "haste")
}

func TestOldB371FontOfReturnReturnsUpToThreeCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB371Seats(g)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	a := oldB371InGraveyard(me, "A", oldB371BearTypeLine, 1, 1)
	b := oldB371InGraveyard(me, "B", oldB371BearTypeLine, 1, 1)
	spell := oldB371InGraveyard(me, "Spell", "Sorcery", 0, 0)
	font := effectstest.PushCatalogPermanent(g, me.ID, "Font of Return", oldB371EnchantmentLine, oldB371FontOracle, false)
	oldB371Float(t, g, me, "{B}", 3)
	if err := g.ActivateCatalogAbility(me.ID, font, 0, game.ActivateAbilityParams{Targets: oldB371Card(a, b)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.ZoneOfCard(g, font) != game.ZoneGraveyard {
		t.Errorf("the Font is in %v, want sacrificed as a cost", effectstest.ZoneOfCard(g, font))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the targeted creature cards did not return to hand")
	}
	if effectstest.ZoneOfCard(g, spell) != game.ZoneGraveyard {
		t.Error("a non-creature card was returned")
	}
}

func TestOldB371ThassasIreTapsOrUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB371Seats(g)
	effectstest.AdvanceToMain(t, g)
	ire := effectstest.PushCatalogPermanent(g, me.ID, "Thassa's Ire", oldB371EnchantmentLine, oldB371IreOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB371BearTypeLine, 2, 2)
	oldB371Float(t, g, me, "{U}", 3)
	if err := g.ActivateCatalogAbility(me.ID, ire, 0, game.ActivateAbilityParams{Targets: oldB371Card(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Error("the creature was not tapped")
	}
}

func TestOldB371PullFromTheDeepReturnsAnInstantAndASorceryThenExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB371Seats(g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	instant := oldB371InGraveyard(me, "Shock", "Instant", 0, 0)
	sorcery := oldB371InGraveyard(me, "Divination", "Sorcery", 0, 0)
	creature := oldB371InGraveyard(me, "Bear", oldB371BearTypeLine, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Pull from the Deep", "Sorcery", oldB371PullOracle, oldB371Card(creature)); err == nil {
		t.Fatal("a creature card was a legal target")
	}
	pull := effectstest.CastCatalogSpell(t, g, "Pull from the Deep", "Sorcery", oldB371PullOracle, oldB371Card(instant, sorcery))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(instant) || !me.Hand.Contains(sorcery) {
		t.Error("the instant and sorcery did not return to hand")
	}
	if effectstest.ZoneOfCard(g, creature) != game.ZoneGraveyard {
		t.Error("the creature card moved")
	}
	if effectstest.ZoneOfCard(g, pull) != game.ZoneExile {
		t.Errorf("Pull from the Deep is in %v, want exile", effectstest.ZoneOfCard(g, pull))
	}
}

func TestOldB371RenownedWeaverMakesASpiderEnchantmentCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB371Seats(g)
	effectstest.AdvanceToMain(t, g)
	weaver := effectstest.B12Push(g, me.ID, "Renowned Weaver", "Creature — Human Shaman", oldB371WeaverOracle, 1, 1)
	oldB371Float(t, g, me, "{G}", 1)
	if err := g.ActivateCatalogAbility(me.ID, weaver, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, weaver) != game.ZoneGraveyard {
		t.Error("the Weaver was not sacrificed")
	}
	found := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Spider" {
			found = true
			if !c.IsEnchantment() || !c.IsCreature() || !c.HasColor("G") || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "reach") {
				t.Error("the Spider is not a green reach enchantment creature")
			}
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 1 || tough != 3 {
				t.Errorf("Spider is %d/%d, want 1/3", p, tough)
			}
		}
	}
	if !found {
		t.Fatal("no Spider token")
	}
}

func TestOldB371RiptideChimeraReturnsAnEnchantmentEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Riptide Chimera", "Enchantment Creature — Chimera", oldB371ChimeraOracle, 3, 4)
	shrine := effectstest.PushCatalogPermanent(g, me.ID, "Shrine", oldB371EnchantmentLine, "", false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB371BearTypeLine, 2, 2)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Riptide Chimera queued no return pick")
	}
	if p.ChooseMin != 1 {
		t.Error("the return is mandatory")
	}
	for _, id := range p.ChooseCards {
		if id == bear {
			t.Error("a non-enchantment was offered")
		}
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{shrine}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(shrine) || !me.Hand.Contains(shrine) {
		t.Error("the enchantment did not return to hand")
	}
}

func TestOldB371DictateOfKarametraAddsAMatchingMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Dictate of Karametra", oldB371EnchantmentLine, oldB371DictateOracle, false)
	effectstest.AssertKeywords(t, g, effectstest.PushCatalogPermanent(g, me.ID, "Dictate of Karametra", oldB371EnchantmentLine, oldB371DictateOracle, false), "flash")
	mine := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Permanent(g, them.ID, "Island", "Basic Land — Island")
	effectstest.TapForMana(t, g, me.ID, mine)
	if got := effectstest.SortedPool(me); len(got) < 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("my pool = %v, want green mana from the land and at least one more", got)
	}
	effectstest.TapForMana(t, g, them.ID, theirs)
	if got := effectstest.SortedPool(them); len(got) < 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("their pool = %v, want blue mana: 'a player', not 'you'", got)
	}
}

func TestOldB371DakraMysticPutsRevealedCardsIntoGraveyardsOrEveryoneDraws(t *testing.T) {
	for _, mill := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB371Seats(g)
		effectstest.AdvanceToMain(t, g)
		mystic := effectstest.B12Push(g, me.ID, "Dakra Mystic", "Creature — Merfolk Wizard", oldB371DakraOracle, 1, 1)
		mine := effectstest.TopOfLibrary(me, "My Top", "Sorcery", "{1}")
		theirs := effectstest.TopOfLibrary(opp, "Their Top", "Sorcery", "{1}")
		myHand, oppHand := me.Hand.Size(), opp.Hand.Size()
		oldB371Float(t, g, me, "{U}", 0)
		if err := g.ActivateCatalogAbility(me.ID, mystic, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerConfirmFor(t, g, me, mill)
		effectstest.PassPriorityAroundTable(t, g)
		if mill {
			if effectstest.ZoneOfCard(g, mine) != game.ZoneGraveyard || effectstest.ZoneOfCard(g, theirs) != game.ZoneGraveyard {
				t.Errorf("mill: the revealed cards are in %v and %v, want graveyards", effectstest.ZoneOfCard(g, mine), effectstest.ZoneOfCard(g, theirs))
			}
			if me.Hand.Size() != myHand || opp.Hand.Size() != oppHand {
				t.Error("mill: someone drew")
			}
		} else {
			if me.Hand.Size() != myHand+1 || opp.Hand.Size() != oppHand+1 {
				t.Errorf("no mill: hands %d→%d and %d→%d, want each player to draw one", myHand, me.Hand.Size(), oppHand, opp.Hand.Size())
			}
		}
	}
}

func TestOldB371RitualOfTheReturnedMakesAZombieTheSizeOfTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB371Seats(g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	dead := oldB371InGraveyard(me, "Big Bear", oldB371BearTypeLine, 5, 3)
	spell := oldB371InGraveyard(me, "Spell", "Sorcery", 0, 0)
	if err := effectstest.CastCatalogSpellErr(t, g, "Ritual of the Returned", "Instant", oldB371RitualOracle, oldB371Card(spell)); err == nil {
		t.Fatal("a noncreature card was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Ritual of the Returned", "Instant", oldB371RitualOracle, oldB371Card(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, dead) != game.ZoneExile {
		t.Fatalf("the creature card is in %v, want exile", effectstest.ZoneOfCard(g, dead))
	}
	found := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Zombie" {
			found = true
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 5 || tough != 3 || !c.HasColor("B") {
				t.Errorf("Zombie is %d/%d, want a black 5/3", p, tough)
			}
		}
	}
	if !found {
		t.Fatal("no Zombie token")
	}
}

func TestOldB371DaringThiefExchangesPermanentsThatShareACardType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB371Seats(g)
	effectstest.AdvanceToMain(t, g)
	thief := effectstest.B12Push(g, me.ID, "Daring Thief", "Creature — Human Rogue", oldB371ThiefOracle, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB371BearTypeLine, 1, 1)
	theirCreature := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB371BearTypeLine, 4, 4)
	theirRock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, thief).Tapped = true })
	if err := g.UntapAll(me.ID); err != nil {
		t.Fatalf("UntapAll: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CloudDeckPickTargets(t, g, me.ID, mine)
	pc := effectstest.LatestPickTarget(g, me.ID)
	if pc == nil {
		t.Fatal("no second target prompt")
	}
	for _, id := range pc.PickTargetCards {
		if id == theirRock {
			t.Fatal("an artifact that shares no type with the chosen creature was offered")
		}
	}
	effectstest.CloudDeckPickTargets(t, g, me.ID, theirCreature)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, mine); c.Controller != opp.ID {
		t.Errorf("my creature is controlled by %v, want the opponent", c.Controller)
	}
	if c := effectstest.FindBattlefieldCardByID(g, theirCreature); c.Controller != me.ID {
		t.Errorf("their creature is controlled by %v, want me", c.Controller)
	}
}

func TestOldB371DiscipleOfDeceitDiscardsANonlandAndFetchesTheSameManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB371Seats(g)
	effectstest.AdvanceToMain(t, g)
	disciple := effectstest.B12Push(g, me.ID, "Disciple of Deceit", "Creature — Human Rogue", oldB371DiscipleOracle, 1, 3)
	for me.Hand.Size() > 0 {
		_, _ = me.Hand.PopTop()
	}
	land := effectstest.EoHand(me, "Forest", "Basic Land — Forest")
	spell := effectstest.EoHand(me, "Cheap Spell", "Sorcery")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == spell {
				me.Hand.Cards[i].ManaCost = "{2}"
			}
		}
	})
	match := effectstest.TopOfLibrary(me, "Match", "Artifact", "{2}")
	effectstest.TopOfLibrary(me, "Other", "Artifact", "{3}")
	oldB371Inspire(t, g, me, disciple)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no discard pick")
	}
	for _, id := range pick.ChooseCards {
		if id == land {
			t.Fatal("a land card was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	if effectstest.ZoneOfCard(g, spell) != game.ZoneGraveyard {
		t.Fatalf("the spell is in %v, want the graveyard", effectstest.ZoneOfCard(g, spell))
	}
	effectstest.AnswerSearchByID(t, g, me.ID, match)
	if !me.Hand.Contains(match) {
		t.Errorf("the card with the same mana value is in %v, want hand", effectstest.ZoneOfCard(g, match))
	}
}

func TestOldB371DiscipleOfDeceitDecliningDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB371Seats(g)
	effectstest.AdvanceToMain(t, g)
	disciple := effectstest.B12Push(g, me.ID, "Disciple of Deceit", "Creature — Human Rogue", oldB371DiscipleOracle, 1, 3)
	spell := effectstest.EoHand(me, "Cheap Spell", "Sorcery")
	oldB371Inspire(t, g, me, disciple)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no discard pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if !me.Hand.Contains(spell) {
		t.Error("the card was discarded after declining")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("a search opened without a discard")
	}
}

func TestOldB371KingMacarExilesACreatureAndMakesGold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB371Seats(g)
	effectstest.AdvanceToMain(t, g)
	macar := effectstest.B12Push(g, me.ID, "King Macar, the Gold-Cursed", "Legendary Creature — Human Noble", oldB371MacarOracle, 2, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB371BearTypeLine, 2, 2)
	oldB371Inspire(t, g, me, macar)
	effectstest.B04WaitForPickOrSkip(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, victim) != game.ZoneExile {
		t.Fatalf("the creature is in %v, want exile", effectstest.ZoneOfCard(g, victim))
	}
	if n := effectstest.OnBattlefieldNamed(g, "Gold"); n != 1 {
		t.Errorf("%d Gold tokens, want 1", n)
	}
}
