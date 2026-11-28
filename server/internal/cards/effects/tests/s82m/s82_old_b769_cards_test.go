package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b769_cards_test.go — Tempest cards, S82 OLD batch 769 (#5628), played
// through the real catalog entries.

const (
	o769Minion        = "2593944c-1bde-4307-b3ca-f7ecbeb027c1"
	o769Selenia       = "3db1ac09-e35a-4f50-97c4-f65dbc28daef"
	o769Warmth        = "cccc9664-0e3d-44c4-959c-abeaba25bbab"
	o769Havoc         = "d95aeeb0-b722-4f70-85fc-6866be4854bd"
	o769OrimsPrayer   = "f387b1c6-25a6-4fff-90d3-fff027b247a5"
	o769Legerdemain   = "107d5f67-1395-444f-b612-c5cc770df956"
	o769CorpseDance   = "59b02ca7-e7bd-4820-ac32-87ea1a8c64c0"
	o769AbandonHope   = "8adbba6e-03ef-4278-aec5-8a4496b377a8"
	o769ScorchedEarth = "fd1ce3e0-9c2d-48b6-9a14-4e71921cdfa4"
	o769SpiritMirror  = "42e041c0-ffbc-4bf8-a5aa-f0d14aad7112"
	o769Sarcomancy    = "bd1e93da-7524-4bdd-9738-accd76bda7f7"
	o769MongrelPack   = "e1208d39-86e7-4eae-8dd6-b5bfb29def84"
)

func TestS82OldB769CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o769Minion: "Minion of the Wastes", o769Selenia: "Selenia, Dark Angel", o769Warmth: "Warmth",
		o769Havoc: "Havoc", o769OrimsPrayer: "Orim's Prayer", o769Legerdemain: "Legerdemain",
		o769CorpseDance: "Corpse Dance", o769AbandonHope: "Abandon Hope", o769ScorchedEarth: "Scorched Earth",
		o769SpiritMirror: "Spirit Mirror", o769Sarcomancy: "Sarcomancy", o769MongrelPack: "Mongrel Pack",
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

func TestS82OldB769WarmthGainsLifeForAnOpponentsRedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Warmth", "Enchantment", o769Warmth, false)
	life := me.Life
	effectstest.B14CastAt(t, g, me, "My Bolt", "Sorcery", "", []string{"R"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("the controller's own red spell gained life: %d, want %d", me.Life, life)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B14CastAt(t, g, opp, "Their Growth", "Sorcery", "", []string{"G"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an opponent's green spell gained life: %d, want %d", me.Life, life)
	}
	effectstest.B14CastAt(t, g, opp, "Their Bolt", "Sorcery", "", []string{"R"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life is %d, want %d after an opponent's red spell", me.Life, life+2)
	}
}

func TestS82OldB769HavocCostsAnOpponentCastingAWhiteSpellTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Havoc", "Enchantment", o769Havoc, false)
	mine := me.Life
	effectstest.B14CastAt(t, g, me, "My Blessing", "Sorcery", "", []string{"W"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("the controller's own white spell cost life: %d, want %d", me.Life, mine)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	theirs := opp.Life
	effectstest.B14CastAt(t, g, opp, "Their Growth", "Sorcery", "", []string{"G"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs {
		t.Errorf("a green spell cost life: %d, want %d", opp.Life, theirs)
	}
	effectstest.B14CastAt(t, g, opp, "Their Blessing", "Sorcery", "", []string{"W"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-2 {
		t.Errorf("the caster is at %d, want %d", opp.Life, theirs-2)
	}
	if me.Life != mine {
		t.Errorf("the controller lost life too: %d, want %d", me.Life, mine)
	}
}

func TestS82OldB769OrimsPrayerGainsLifeForEachAttackerOnYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, me := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Orim's Prayer", "Enchantment", o769OrimsPrayer, false)
	a := effectstest.B12Creature(g, attacker.ID, "Wall A", "Creature — Wall", 0, 3)
	b := effectstest.B12Creature(g, attacker.ID, "Wall B", "Creature — Wall", 0, 3)
	c := effectstest.B12Creature(g, attacker.ID, "Wall C", "Creature — Wall", 0, 3)
	life := me.Life
	effectstest.BrAttack(t, g, a, b, c)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life is %d, want %d (one for each of three attackers)", me.Life, life+3)
	}
}

func TestS82OldB769OrimsPrayerIgnoresAnAttackOnSomeoneElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, me, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Orim's Prayer", "Enchantment", o769OrimsPrayer, false)
	a := effectstest.B12Creature(g, attacker.ID, "Wall A", "Creature — Wall", 0, 3)
	life := me.Life
	effectstest.AttackWith(t, g, other.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an attack on another player gained life: %d, want %d", me.Life, life)
	}
}

func TestS82OldB769SeleniaPaysTwoLifeToReturnToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sel := effectstest.B12Push(g, me.ID, "Selenia, Dark Angel", "Legendary Creature — Phyrexian Angel", o769Selenia, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, sel, "flying") {
		t.Error("Selenia does not fly")
	}
	me.Life = 1
	if err := g.ActivateCatalogAbility(me.ID, sel, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability was activated with only 1 life to pay")
	}
	me.Life = 20
	if err := g.ActivateCatalogAbility(me.ID, sel, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != 18 {
		t.Errorf("life is %d, want 18 (the cost)", me.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sel) || !me.Hand.Contains(sel) {
		t.Error("Selenia did not return to its owner's hand")
	}
}

func TestS82OldB769MinionOfTheWastesIsAsBigAsTheLifePaid(t *testing.T) {
	for _, pay := range []int{0, 7} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		me.Life = 40
		minion := effectstest.CastCatalogSpell(t, g, "Minion of the Wastes", "Creature — Minion", o769Minion, nil)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.PayAmountPrompt(g, me.ID)
		if c == nil {
			t.Fatal("no life prompt as the Minion entered")
		}
		if pa := c.PayAmount; pa.ResourceOrEnergy() != game.PayResourceLife || pa.Max != 40 {
			t.Fatalf("prompt = %+v, want life up to 40", *pa)
		}
		if err := g.ResolvePayAmount(c.ID, me.ID, pay); err != nil {
			t.Fatalf("pay %d: %v", pay, err)
		}
		g.RunStateChecksForTest()
		if me.Life != 40-pay {
			t.Errorf("life %d, want %d", me.Life, 40-pay)
		}
		if pay == 0 {
			if g.Battlefield.Contains(minion) || !me.Graveyard.Contains(minion) {
				t.Error("a Minion that paid nothing is 0/0 and should have died")
			}
			continue
		}
		p, tough := effectstest.PtOf(t, g, minion)
		if p != pay || tough != pay {
			t.Errorf("the Minion is %d/%d, want %d/%d", p, tough, pay, pay)
		}
		if !effectstest.HasEffectiveKeyword(t, g, minion, "trample") {
			t.Error("the Minion has no trample")
		}
	}
}

func TestS82OldB769LegerdemainExchangesTwoCreaturesOrTwoArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	rock := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	ench := effectstest.B12Push(g, opp.ID, "Their Aura", "Enchantment", "", 0, 0)
	// An artifact and a creature share no type; neither does an enchantment.
	for name, targets := range map[string][]game.TargetRef{
		"creature and artifact":    effectstest.CardRefs(mine, rock),
		"creature and enchantment": effectstest.CardRefs(mine, ench),
		"the same creature twice":  effectstest.CardRefs(mine, mine),
	} {
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "Legerdemain", "Sorcery", o769Legerdemain, targets); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
	effectstest.CastCatalogSpell(t, g, "Legerdemain", "Sorcery", o769Legerdemain, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(mine); !ok || c.Controller != opp.ID {
		t.Error("my Bear did not change hands")
	}
	if c, ok := g.LookupCardForEffect(theirs); !ok || c.Controller != me.ID {
		t.Error("their Ogre did not change hands")
	}
}

func TestS82OldB769LegerdemainSharesOnlyArtifactOrCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Both are enchantments; one is an artifact and the other a creature, so
	// they share only the enchantment type, which does not count.
	art := effectstest.B12Push(g, me.ID, "Enchanted Relic", "Enchantment Artifact", "", 0, 0)
	cre := effectstest.B12Creature(g, opp.ID, "Enchanted Beast", "Enchantment Creature — Beast", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Legerdemain", "Sorcery", o769Legerdemain, effectstest.CardRefs(art, cre)); err == nil {
		t.Error("two permanents that share only the enchantment type were accepted")
	}
	art2 := effectstest.B12Push(g, opp.ID, "Their Relic", "Artifact", "", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Legerdemain", "Sorcery", o769Legerdemain, effectstest.CardRefs(art, art2))
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(art); !ok || c.Controller != opp.ID {
		t.Error("the two artifacts did not trade places")
	}
}

func TestS82OldB769CorpseDanceReturnsTheTopCreatureWithHasteAndExilesIt(t *testing.T) {
	for _, buyback := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		me.Graveyard.Cards = nil
		low := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: low, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		top := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: top, Name: "Fresh Ogre", TypeLine: "Creature — Ogre", Owner: me.ID, Controller: me.ID})
		sorcery := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: sorcery, Name: "Spent Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		var optional []int
		if buyback {
			optional = []int{0}
		}
		spell, err := effectstest.CastWithOptionalCosts(t, g, "Corpse Dance", "Instant", o769CorpseDance, nil, optional, nil)
		if err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(top) {
			t.Fatal("the top creature card did not return")
		}
		if g.Battlefield.Contains(low) || !me.Graveyard.Contains(sorcery) {
			t.Error("a card other than the top creature moved")
		}
		if !effectstest.HasEffectiveKeyword(t, g, top, "haste") {
			t.Error("the returned creature has no haste")
		}
		effectstest.TmpB1AssertBuyback(t, me, spell, buyback)
		effectstest.AdvanceToEndStepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(top) || !effectstest.InExile(g, top) {
			t.Error("the returned creature was not exiled at the beginning of the next end step")
		}
	}
}

func TestS82OldB769CorpseDanceWithNoCreatureStillBuysBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	spell, err := effectstest.CastWithOptionalCosts(t, g, "Corpse Dance", "Instant", o769CorpseDance, nil, []int{0}, nil)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TmpB1AssertBuyback(t, me, spell, true)
}

func TestS82OldB769AbandonHopeDiscardsXAndTakesXCardsOfTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Hand.Cards = nil
	spell := effectstest.AvB24Hand(g, me, game.Card{Name: "Abandon Hope", TypeLine: "Sorcery", ManaCost: "{X}{1}{B}", OracleID: o769AbandonHope})
	d1 := effectstest.HandCard(me, "Spare One", "Instant")
	d2 := effectstest.HandCard(me, "Spare Two", "Instant")
	effectstest.AdvanceToMain(t, g)
	opp.Hand.Cards = nil
	keep := effectstest.HandSpell(opp, "Keep", "Sorcery", "")
	toss1 := effectstest.HandSpell(opp, "Toss One", "Sorcery", "")
	toss2 := effectstest.HandSpell(opp, "Toss Two", "Sorcery", "")
	effectstest.FloatForTest(g, me, "BCCC")
	refs := effectstest.PlayerRefs(opp.ID)
	// One discard at X=2 is refused.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{d1}, Targets: refs}); err == nil {
		t.Fatal("the cast was accepted with one discard at X=2")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{d1, d2}, Targets: refs}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !me.Graveyard.Contains(d1) || !me.Graveyard.Contains(d2) {
		t.Error("the cost did not discard two cards")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if len(pick.DiscardOptions) != 3 || pick.FromPlayer != opp.ID {
		t.Fatalf("pick offers %d cards from %v, want the target's hand of 3", len(pick.DiscardOptions), pick.FromPlayer)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{toss1, toss2}); err != nil {
		t.Fatalf("ResolvePendingChoice: %v", err)
	}
	if !opp.Graveyard.Contains(toss1) || !opp.Graveyard.Contains(toss2) || !opp.Hand.Contains(keep) {
		t.Error("the chosen two cards were not the ones discarded")
	}
}

func TestS82OldB769ScorchedEarthDiscardsXLandsAndDestroysXLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Hand.Cards = nil
	spell := effectstest.AvB24Hand(g, me, game.Card{Name: "Scorched Earth", TypeLine: "Sorcery", ManaCost: "{X}{R}", OracleID: o769ScorchedEarth})
	lands := []uuid.UUID{
		effectstest.HandCard(me, "Mountain", "Basic Land — Mountain"),
		effectstest.HandCard(me, "Forest", "Basic Land — Forest"),
	}
	nonland := effectstest.HandCard(me, "Shock", "Instant")
	effectstest.AdvanceToMain(t, g)
	targets := effectstest.CardRefs(
		effectstest.AcPermanent(g, opp.ID, "Island", "Basic Land — Island", 0, 0),
		effectstest.AcPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp", 0, 0))
	effectstest.FloatForTest(g, me, "RCC")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{lands[0], nonland}, Targets: targets}); err == nil {
		t.Fatal("a nonland card paid the land discard")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: lands, Targets: targets}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, tr := range targets {
		if !opp.Graveyard.Contains(tr.ID) {
			t.Errorf("target land %s was not destroyed", tr.ID)
		}
	}
	if !me.Hand.Contains(nonland) {
		t.Error("the nonland card left the hand")
	}
}

func o769Reflections(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Reflection" {
			n++
		}
	}
	return n
}

func TestS82OldB769SpiritMirrorMakesAReflectionOnlyWhenThereIsNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Spirit Mirror", "Enchantment", o769SpiritMirror, false)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := o769Reflections(g); n != 1 {
		t.Fatalf("%d Reflections after the first upkeep, want 1", n)
	}
	n, tok := effectstest.B14Tokens(g, me.ID, "Reflection")
	if n != 1 || tok.Power != 2 || tok.Toughness != 2 || !tok.HasColorWord("W") {
		t.Errorf("token = %d, %d/%d, want one 2/2 white Reflection", n, tok.Power, tok.Toughness)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := o769Reflections(g); n != 1 {
		t.Errorf("%d Reflections after the next upkeep, want still 1", n)
	}
}

func TestS82OldB769SpiritMirrorDestroysAnyReflectionForNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mirror := effectstest.PushCatalogPermanent(g, me.ID, "Spirit Mirror", "Enchantment", o769SpiritMirror, false)
	refl := effectstest.B12Creature(g, opp.ID, "Reflection", "Creature — Reflection", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, mirror, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("a non-Reflection was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, mirror, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(refl)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(refl) || !opp.Graveyard.Contains(refl) {
		t.Error("the Reflection was not destroyed")
	}
}

func TestS82OldB769SarcomancyMakesAZombieAndPingsYouWithoutOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Sarcomancy", "Enchantment", o769Sarcomancy, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	n, tok := effectstest.B14Tokens(g, me.ID, "Zombie")
	if n != 1 || tok.Power != 2 || tok.Toughness != 2 || !tok.HasColorWord("B") {
		t.Fatalf("token = %d, %d/%d, want one 2/2 black Zombie", n, tok.Power, tok.Toughness)
	}
	life := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d while a Zombie is on the battlefield", me.Life, life)
	}
	effectstest.B14Kill(g, tok.InstanceID)
	g.RunStateChecksForTest()
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d, want %d with no Zombie", me.Life, life-1)
	}
}

func TestS82OldB769SarcomancyAnyZombieStopsTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sarcomancy", "Enchantment", o769Sarcomancy, false)
	effectstest.B12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	life := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d: an opponent's Zombie card counts", me.Life, life)
	}
}

func TestS82OldB769MongrelPackLeavesFourDogsOnlyWhenItDiesInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pack := effectstest.B12Push(g, me.ID, "Mongrel Pack", "Creature — Dog", o769MongrelPack, 4, 1)
	effectstest.B14Kill(g, pack)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.B14Tokens(g, me.ID, "Dog"); n != 0 {
		t.Fatalf("%d Dogs after a death in the main phase, want none", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	pack2 := effectstest.B12Push(g2, me2.ID, "Mongrel Pack", "Creature — Dog", o769MongrelPack, 4, 1)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	effectstest.B14Kill(g2, pack2)
	g2.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g2)
	n, tok := effectstest.B14Tokens(g2, me2.ID, "Dog")
	if n != 4 || tok.Power != 1 || tok.Toughness != 1 || !tok.HasColorWord("G") {
		t.Errorf("tokens = %d, %d/%d, want four 1/1 green Dogs", n, tok.Power, tok.Toughness)
	}
}
