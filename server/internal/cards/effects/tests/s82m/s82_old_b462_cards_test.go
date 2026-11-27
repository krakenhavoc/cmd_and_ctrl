package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b462_cards_test.go — New Phyrexia, batch 462 (S82, #5575).
// Each test plays the real catalog entry through the engine.

const (
	b462Survivors = "103b92c7-8f8e-4924-8d7e-fc0a01486e86"
	b462Remember  = "7ed3918d-509a-4fe6-a32d-5899df8f9046"
	b462Lunge     = "aacd1aae-1bf0-43e4-99f4-c6d738af812c"
	b462Stance    = "19959b6d-eb8f-4273-94c8-65210e1998ab"
	b462Harvest   = "48dc8f34-8c73-4155-8a3d-b26a35260e6b"
	b462Dose      = "e42d087e-c999-4f9a-bdc5-4327e8ce94b6"
	b462Trainee   = "1d7d340d-fe6f-4314-97ba-e0bf36a4ef6c"
	b462Bite      = "42e4a55b-12c8-4ad0-a27d-28bbdbb06940"
	b462Tangle    = "2a521811-7106-4531-b69c-631e207b8988"
	b462Annex     = "85c0a376-da6f-4f69-853c-a093708d0952"
	b462Slag      = "4c5e38ca-39d7-4e27-8bb7-687d86fd67c4"
	b462Melira    = "a6f72f15-d103-49fb-a808-054b278d00a8"
)

func TestB462EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b462Survivors: "Auriok Survivors", b462Remember: "Remember the Fallen", b462Lunge: "Postmortem Lunge",
		b462Stance: "Defensive Stance", b462Harvest: "Viridian Harvest", b462Dose: "Numbing Dose",
		b462Trainee: "Greenhilt Trainee", b462Bite: "Leeching Bite", b462Tangle: "Chancellor of the Tangle",
		b462Annex: "Chancellor of the Annex", b462Slag: "Slag Fiend", b462Melira: "Melira, Sylvok Outcast",
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

func b462Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func b462Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	params.Strict = true
	return id, g.CastSpell(me.ID, id, params)
}

func b462MustCast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id, err := b462Cast(t, g, me, name, typeLine, cost, oracle, params)
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func b462Gy(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

func TestB462AuriokSurvivorsMayReturnAndAttachAnEquipment(t *testing.T) {
	for _, attach := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, _ := b462Seats(g)
		effectstest.AdvanceToMain(t, g)
		sword := b462Gy(me, "Old Sword", "Artifact — Equipment", "{2}")
		axe := b462Gy(me, "Old Hat", "Artifact", "{2}")
		_ = axe
		surv := b462MustCast(t, g, me, "Auriok Survivors", "Creature — Human Soldier", "", b462Survivors, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PickCard(t, g, me.ID, sword)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.FindBattlefieldCardForTest(g, sword) == nil {
			t.Fatal("the Equipment was not returned to the battlefield")
		}
		effectstest.AnswerMayChoice(t, g, me.ID, attach)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.AttachedTo(t, g, sword, surv); got != attach {
			t.Errorf("attach=%v: attached = %v", attach, got)
		}
	}
}

func TestB462AuriokSurvivorsOnlyTargetsAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b462Seats(g)
	effectstest.AdvanceToMain(t, g)
	hat := b462Gy(me, "Old Hat", "Artifact", "{2}")
	b462MustCast(t, g, me, "Auriok Survivors", "Creature — Human Soldier", "", b462Survivors, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, hat) {
		t.Error("a non-Equipment artifact was offered")
	}
}

func TestB462RememberTheFallenReturnsOneOrBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b462Seats(g)
	bear := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	rock := effectstest.PushGraveyardCardTyped(me, "Old Rock", "Artifact")
	effectstest.CastModal(t, g, "Remember the Fallen", "Sorcery", b462Remember, []int{0, 1}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, bear, 0, 0), effectstest.ModeRef(game.TargetCard, rock, 1, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(rock) {
		t.Errorf("hand has bear=%v rock=%v, want both", me.Hand.Contains(bear), me.Hand.Contains(rock))
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, _ := b462Seats(g2)
	bear2 := effectstest.PushGraveyardCardTyped(me2, "Old Bear", "Creature — Bear")
	rock2 := effectstest.PushGraveyardCardTyped(me2, "Old Rock", "Artifact")
	effectstest.CastModal(t, g2, "Remember the Fallen", "Sorcery", b462Remember, []int{1}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, rock2, 0, 0),
	})
	effectstest.PassPriorityAroundTable(t, g2)
	if !me2.Hand.Contains(rock2) || me2.Hand.Contains(bear2) {
		t.Error("one mode must return only its own card")
	}
	if _, err := effectstest.CastModalErr(t, g2, "Remember the Fallen", "Sorcery", b462Remember, []int{0}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, rock2, 0, 0),
	}); err == nil {
		t.Error("an artifact card was accepted for the creature mode")
	}
}

func TestB462PostmortemLungeReturnsAnXCreatureWithHasteAndExilesItAtTheEnd(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b462Seats(g)
	effectstest.AdvanceToMain(t, g)
	three := b462Gy(me, "Three Drop", "Creature — Bear", "{2}{R}")
	two := b462Gy(me, "Two Drop", "Creature — Bear", "{1}{R}")
	effectstest.ApaMana(me, "B", "C", "C", "C")
	if _, err := b462Cast(t, g, me, "Postmortem Lunge", "Sorcery", "{X}{B/P}", b462Lunge,
		game.CastSpellParams{XValue: 3, Targets: effectstest.CardTarget(two)}); err == nil {
		t.Fatal("a creature with mana value 2 was accepted for X = 3")
	}
	effectstest.ApaMana(me, "B", "C", "C", "C")
	b462MustCast(t, g, me, "Postmortem Lunge", "Sorcery", "{X}{B/P}", b462Lunge,
		game.CastSpellParams{XValue: 3, Targets: effectstest.CardTarget(three)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, three) == nil {
		t.Fatal("the creature did not return")
	}
	if !effectstest.HasEffectiveKeyword(t, g, three, "haste") {
		t.Error("the returned creature has no haste")
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, three) {
		t.Error("the creature was not exiled at the next end step")
	}
}

func TestB462DefensiveStanceShrinksPowerAndGrowsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b462Seats(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "U")
	b462MustCast(t, g, me, "Defensive Stance", "Enchantment — Aura", "{U}", b462Stance, game.CastSpellParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 1 || tough != 3 {
		t.Errorf("enchanted Bear is %d/%d, want 1/3", p, tough)
	}
}

func TestB462ViridianHarvestPaysSixWhenTheArtifactGoesToAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b462Seats(g)
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "G")
	if _, err := b462Cast(t, g, me, "Viridian Harvest", "Enchantment — Aura", "{G}", b462Harvest, game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("it enchanted a non-artifact creature")
	}
	effectstest.ApaMana(me, "G")
	b462MustCast(t, g, me, "Viridian Harvest", "Enchantment — Aura", "{G}", b462Harvest, game.CastSpellParams{Targets: effectstest.CardTarget(rock)})
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	effectstest.B14Kill(g, rock)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 6 {
		t.Errorf("gained %d, want 6 (the Aura's controller, not the artifact's)", got)
	}
}

func TestB462NumbingDoseStopsUntappingAndDrainsTheControllerEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b462Seats(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.ApaMana(me, "U", "U", "C", "C", "C")
	b462MustCast(t, g, me, "Numbing Dose", "Enchantment — Aura", "{3}{U}{U}", b462Dose, game.CastSpellParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	life := opp.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, oppSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 1 {
		t.Errorf("the controller lost %d life, want 1", got)
	}
	if !effectstest.TappedOnBattlefield(t, g, bear) {
		t.Error("the enchanted creature untapped")
	}
	if me.Life != g.Settings.StartingLife {
		t.Error("the Aura's controller lost life")
	}
}

func TestB462GreenhiltTraineeNeedsFourPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b462Seats(g)
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Push(g, me.ID, "Greenhilt Trainee", "Creature — Elf Warrior", b462Trainee, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.RavB2Activate(t, g, me.ID, small, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("activated with power 2")
	}
	big := effectstest.B12Push(g, me.ID, "Greenhilt Trainee", "Creature — Elf Warrior", b462Trainee, 4, 3)
	effectstest.RavB2MustActivate(t, g, me.ID, big, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if p, tough := effectstest.PtOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("target is %d/%d, want 6/6", p, tough)
	}
}

func TestB462LeechingBiteShrinksOneAndGrowsAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b462Seats(g)
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.ApaMana(me, "G", "C")
	if _, err := b462Cast(t, g, me, "Leeching Bite", "Instant", "{1}{G}", b462Bite, game.CastSpellParams{Targets: effectstest.CardRefs(mine, mine)}); err == nil {
		t.Fatal("the same creature was both targets")
	}
	effectstest.ApaMana(me, "G", "C")
	b462MustCast(t, g, me, "Leeching Bite", "Instant", "{1}{G}", b462Bite, game.CastSpellParams{Targets: effectstest.CardRefs(mine, theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("first target %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("second target %d/%d, want 2/2", p, tough)
	}
}

func b462Reveal(t *testing.T, g *game.Game, seat int, card game.Card) *game.Player {
	t.Helper()
	p := g.Seats[seat]
	effectstest.HandWith(g, seat, card, effectstest.FillerCard("Spare"))
	effectstest.KeepAll(t, g)
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c.Kind == game.PendingChoiceConfirm && c.Chooser == p.ID {
			offer = c
		}
	}
	if offer == nil {
		t.Fatal("not offered the reveal")
	}
	if err := g.ResolveConfirm(offer.ID, p.ID, true); err != nil {
		t.Fatalf("reveal: %v", err)
	}
	return p
}

func TestB462ChancellorOfTheTangleAddsGAtYourFirstMainPhase(t *testing.T) {
	g := effectstest.OpeningTable(t)
	starter := b462Reveal(t, g, g.StartingSeat, game.Card{Name: "Chancellor of the Tangle", TypeLine: "Creature — Phyrexian Beast",
		ManaCost: "{4}{G}{G}{G}", OracleID: b462Tangle, Power: 6, Toughness: 7})
	effectstest.AdvanceToMainOf(t, g, g.StartingSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if starter.ManaPool.Total() != 1 {
		t.Errorf("pool %v, want one {G}", effectstest.PoolColors(starter))
	}
}

func TestB462ChancellorOfTheAnnexCountersTheFirstSpellUnlessPaid(t *testing.T) {
	g := effectstest.OpeningTable(t)
	starter := b462Reveal(t, g, g.StartingSeat, game.Card{Name: "Chancellor of the Annex", TypeLine: "Creature — Phyrexian Angel",
		ManaCost: "{4}{W}{W}{W}", OracleID: b462Annex, Power: 5, Toughness: 6})
	opp := g.Seats[(g.StartingSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMainOf(t, g, g.StartingSeat)
	first := effectstest.Batch01OpponentCasts(t, g, opp, "First Spell", "", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P7Countered(g, first) {
		t.Fatal("the opponent's first spell was not countered")
	}
	// Once only (CR 603.7b), and the card is still in hand: the second spell is untaxed.
	second := effectstest.Batch01OpponentCasts(t, g, opp, "Second Spell", "", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.P7Countered(g, second) {
		t.Error("the second spell was countered by the opening-hand reveal")
	}
	_ = starter
}

func TestB462ChancellorOfTheAnnexOnTheBattlefieldTaxesEverySpell(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := b462Seats(g)
		effectstest.AdvanceToMain(t, g)
		effectstest.B12Push(g, me.ID, "Chancellor of the Annex", "Creature — Phyrexian Angel", b462Annex, 5, 6)
		if pay {
			effectstest.ApaMana(opp, "C")
		}
		spell := effectstest.Batch01OpponentCasts(t, g, opp, "Opp Spell", "", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.P7Countered(g, spell); got == pay {
			t.Errorf("pay=%v: countered=%v", pay, got)
		}
		// Its own controller's spells are not taxed.
		mine := effectstest.CastFromSeat(t, g, me, "My Spell", "Instant", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.P7Countered(g, mine) {
			t.Error("the Chancellor's controller's spell was countered")
		}
	}
}

func TestB462SlagFiendCountsArtifactCardsInEveryGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b462Seats(g)
	fiend := effectstest.B12Push(g, me.ID, "Slag Fiend", "Creature — Phyrexian Construct", b462Slag, 0, 0)
	if p, tough := effectstest.PtOf(t, g, fiend); p != 0 || tough != 0 {
		t.Fatalf("with empty graveyards %d/%d, want 0/0", p, tough)
	}
	for _, d := range []struct {
		owner    uuid.UUID
		name, tl string
	}{{me.ID, "Rock", "Artifact"}, {opp.ID, "Golem", "Artifact Creature — Golem"}, {opp.ID, "Bear", "Creature — Bear"}, {opp.ID, "Forest", "Basic Land — Forest"}} {
		effectstest.B14Kill(g, effectstest.B12Push(g, d.owner, d.name, d.tl, "", 1, 1))
	}
	g.RunStateChecksForTest()
	if p, tough := effectstest.PtOf(t, g, fiend); p != 2 || tough != 2 {
		t.Errorf("two artifact cards in graveyards: %d/%d, want 2/2", p, tough)
	}
}

func TestB462MeliraSylvokOutcastShieldsYouFromPoisonAndMinusCountersAndStripsInfect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b462Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Melira, Sylvok Outcast", "Legendary Creature — Human Scout", b462Melira, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	viper := effectstest.B12Creature(g, opp.ID, "Viper", "Creature — Snake", 1, 1)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, viper).Keywords = append(effectstest.FindBattlefieldCardForTest(g, viper).Keywords, "infect")
	})
	if effectstest.HasEffectiveKeyword(t, g, viper, "infect") {
		t.Error("an opponent's creature kept infect")
	}
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(me.ID, game.CounterPoison, 2); err != nil {
			t.Errorf("poison: %v", err)
		}
		if err := g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 2); err != nil {
			t.Errorf("poison: %v", err)
		}
		if err := g.AddCounterForEffect(bear, game.CounterMinusOne, 1); err != nil {
			t.Errorf("-1/-1: %v", err)
		}
		if err := g.AddCounterForEffect(theirs, game.CounterMinusOne, 1); err != nil {
			t.Errorf("-1/-1: %v", err)
		}
	})
	if got := me.Counters[game.CounterPoison]; got != 0 {
		t.Errorf("you got %d poison counters, want 0", got)
	}
	if got := opp.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("an opponent got %d poison counters, want 2", got)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterMinusOne); got != 0 {
		t.Errorf("your creature has %d -1/-1 counters, want 0", got)
	}
	if got := effectstest.CounterCount(g, theirs, game.CounterMinusOne); got != 1 {
		t.Errorf("an opponent's creature has %d -1/-1 counters, want 1", got)
	}
}
