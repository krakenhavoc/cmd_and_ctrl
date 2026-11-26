package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b354_test.go — S82 older sets, batch 354: Magic 2015 damage cards
// (#5556).

const (
	oldB354GlacialCrasher     = "42441362-d638-423a-8928-ab3fd9b2d4f2"
	oldB354MasterPredicaments = "503447d9-2489-46d7-b5aa-a14bbb22dcc8"
	oldB354HeatRay            = "76ec76b9-da0f-4b9d-ad0a-d734052a5f2b"
	oldB354AltacBloodseeker   = "7f9d5219-0306-4c39-8640-dcb4314042de"
	oldB354XathridSlyblade    = "931c0cd3-06fb-42da-86f2-476d08a238f9"
	oldB354CircleOfFlame      = "c31801e6-2444-4158-8e19-2afe751d7cff"
	oldB354RazorfootGriffin   = "c7ecb785-2bfc-4953-b356-db7c27406dbb"
	oldB354SanctifiedCharge   = "e89b3133-4cd6-44fb-ae86-a01b86bb970c"
)

func TestOldB354EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB354GlacialCrasher: "Glacial Crasher", oldB354MasterPredicaments: "Master of Predicaments",
		oldB354HeatRay: "Heat Ray", oldB354AltacBloodseeker: "Altac Bloodseeker",
		oldB354XathridSlyblade: "Xathrid Slyblade", oldB354CircleOfFlame: "Circle of Flame",
		oldB354RazorfootGriffin: "Razorfoot Griffin", oldB354SanctifiedCharge: "Sanctified Charge",
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

func TestOldB354RazorfootGriffinHasFlyingAndFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Razorfoot Griffin", "Creature — Griffin", oldB354RazorfootGriffin, 2, 2)
	effectstest.AdvanceToMain(t, g)
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestOldB354GlacialCrasherNeedsAMountainOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	crasher := effectstest.B12Push(g, me.ID, "Glacial Crasher", "Creature — Elemental", oldB354GlacialCrasher, 5, 5)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(crasher, opp.ID); err == nil {
		t.Fatal("the Crasher attacked with no Mountain on the battlefield")
	}
	// A Mountain under the opponent's control counts.
	effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	if err := g.DeclareAttacker(crasher, opp.ID); err != nil {
		t.Fatalf("the Crasher could not attack with a Mountain in play: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, crasher, "trample") {
		t.Error("no trample")
	}
}

func TestOldB354HeatRayDealsXDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heat Ray", TypeLine: "Instant", ManaCost: "{X}{R}", OracleID: oldB354HeatRay, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "R", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, Targets: effectstest.CardTarget(big)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, big); got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
	// Not castable at a player.
	id2 := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id2, Name: "Heat Ray", TypeLine: "Instant", ManaCost: "{X}{R}", OracleID: oldB354HeatRay, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "R")
	if err := g.CastSpell(me.ID, id2, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("Heat Ray targeted a player")
	}
}

func TestOldB354SanctifiedChargePumpsAllAndGivesWhiteOnesFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	white := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "White Knight", TypeLine: "Creature — Knight", Colors: []string{"W"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	green := effectstest.B12Creature(g, me.ID, "Green Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Sanctified Charge", "Instant", oldB354SanctifiedCharge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{white, green} {
		if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 3 {
			t.Errorf("%s is %d/%d, want 4/3", id, p, tough)
		}
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opposing creature is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, white, "first strike") {
		t.Error("the white creature has no first strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, green, "first strike") {
		t.Error("the green creature got first strike")
	}
}

func TestOldB354XathridSlybladeLosesHexproofForGainingFirstStrikeAndDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	blade := effectstest.B12Push(g, me.ID, "Xathrid Slyblade", "Creature — Human Assassin", oldB354XathridSlyblade, 2, 1)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, blade, "hexproof") {
		t.Fatal("no hexproof")
	}
	if err := g.ActivateCatalogAbility(me.ID, blade, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no mana")
	}
	effectstest.FloatForTest(g, me, "CCCB")
	effectstest.B16Activate(t, g, me.ID, blade, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, blade, "hexproof") {
		t.Error("still has hexproof")
	}
	for _, kw := range []string{"first strike", "deathtouch"} {
		if !effectstest.HasEffectiveKeyword(t, g, blade, kw) {
			t.Errorf("no %s", kw)
		}
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if !effectstest.HasEffectiveKeyword(t, g, blade, "hexproof") {
		t.Error("hexproof did not come back")
	}
}

func TestOldB354AltacBloodseekerWatchesOpposingCreaturesDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seeker := effectstest.B12Push(g, me.ID, "Altac Bloodseeker", "Creature — Human Berserker", oldB354AltacBloodseeker, 2, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, seeker); p != 2 {
		t.Errorf("power %d after my own creature died, want 2", p)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, seeker); p != 4 {
		t.Errorf("power %d after an opposing creature died, want 4", p)
	}
	for _, kw := range []string{"first strike", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, seeker, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestOldB354CircleOfFlamePingsGroundAttackersOnly(t *testing.T) {
	run := func(t *testing.T, flying bool) int {
		g := effectstest.NewCatalogGame(t)
		atk, def := effectstest.SeatsForCombat(g)
		effectstest.B12Push(g, def.ID, "Circle of Flame", "Enchantment", oldB354CircleOfFlame, 0, 0)
		c := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Attacker", TypeLine: "Creature — Bear", Power: 1, Toughness: 4, Owner: atk.ID, Controller: atk.ID,
			Keywords: kwIf(flying)})
		effectstest.AdvanceToMain(t, g)
		effectstest.StdB42Attack(t, g, def.ID, c)
		effectstest.PassPriorityAroundTable(t, g)
		return effectstest.DamageOn(t, g, c)
	}
	if got := run(t, false); got != 1 {
		t.Errorf("a ground attacker has %d damage, want 1", got)
	}
	if got := run(t, true); got != 0 {
		t.Errorf("a flying attacker has %d damage, want 0", got)
	}
}

func kwIf(flying bool) []string {
	if flying {
		return []string{"flying"}
	}
	return nil
}

func oldB354MasterHits(t *testing.T, handMV string) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	master := effectstest.B12Push(g, me.ID, "Master of Predicaments", "Creature — Sphinx", oldB354MasterPredicaments, 4, 4)
	effectstest.AdvanceToMain(t, g)
	spell := game.Card{InstanceID: uuid.New(), Name: "Test Card", TypeLine: "Creature — Giant", ManaCost: handMV, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(spell)
	effectstest.StdB42Attack(t, g, opp.ID, master)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, spell.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, opp, spell.InstanceID
}

func TestOldB354MasterOfPredicamentsWrongGuessOffersAFreeCast(t *testing.T) {
	g, me, opp, card := oldB354MasterHits(t, "{4}{G}") // mana value 5: greater than 4
	effectstest.AnswerMayChoice(t, g, opp.ID, false)   // guesses "not greater": wrong
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, card)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, card, game.CastSpellParams{}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card) {
		t.Error("the card was not cast for free")
	}
}

func TestOldB354MasterOfPredicamentsRightGuessOffersNothing(t *testing.T) {
	g, me, opp, card := oldB354MasterHits(t, "{4}{G}")
	effectstest.AnswerMayChoice(t, g, opp.ID, true) // "greater": right
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a free cast was offered after a right guess")
	}
	if !me.Hand.Contains(card) {
		t.Error("the card left the hand")
	}
}

func TestOldB354MasterOfPredicamentsWrongGuessOnALowCardCanBeDeclined(t *testing.T) {
	g, me, opp, card := oldB354MasterHits(t, "{2}{G}") // mana value 3: not greater than 4
	effectstest.AnswerMayChoice(t, g, opp.ID, true)    // guesses "greater": wrong
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, card)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the declined card left the hand")
	}
}
