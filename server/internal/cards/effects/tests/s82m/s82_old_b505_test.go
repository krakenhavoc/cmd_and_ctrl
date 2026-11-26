package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b505_test.go — Alara Reborn, batch 505 (#5586). Every test plays
// the real catalog entry through the engine.

const (
	oldB505Bolt        = "3018c0ef-876a-4254-9513-42f351e1da3e"
	oldB505Hemorrhage  = "5cc4ee18-675e-4707-b658-ec746a1c03d8"
	oldB505Crystal     = "3e50d4c9-343b-42e7-b777-9d479e97e776"
	oldB505Sigil       = "92017087-18f4-46ee-8808-5959d31d0011"
	oldB505Thrash      = "5d11cdb2-5a44-4e0e-b357-bb1be59f68fb"
	oldB505Paladin     = "75efff64-ea9b-4e3e-a729-fc003ee3db1e"
	oldB505Identity    = "0abe0584-d876-4827-972b-589da53dbe96"
	oldB505Cloven      = "2b08e5f7-c771-4e03-8560-15de6a8bec71"
	oldB505BearType    = "Creature — Bear"
	oldB505LightningID = effectstest.LightningBoltOracle
)

func TestOldB505EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB505Bolt: "Intimidation Bolt", oldB505Hemorrhage: "Thought Hemorrhage", oldB505Crystal: "Crystallization",
		oldB505Sigil: "Sigil of the Nayan Gods", oldB505Thrash: "Skyclaw Thrash", oldB505Paladin: "Knotvine Paladin",
		oldB505Identity: "Identity Crisis", oldB505Cloven: "Cloven Casting",
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

func TestOldB505IntimidationBoltDamagesAndLetsOnlyTheTargetAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, me.ID, "Target", oldB505BearType, 2, 5)
	other := effectstest.B12Creature(g, me.ID, "Other", oldB505BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Intimidation Bolt", "Instant", oldB505Bolt, effectstest.CardTargets(target))
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, me.ID, "Late", oldB505BearType, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(other, opp.ID); err == nil {
		t.Error("another creature attacked")
	}
	if err := g.DeclareAttacker(late, opp.ID); err == nil {
		t.Error("a creature that arrived later attacked")
	}
	if err := g.DeclareAttacker(target, opp.ID); err != nil {
		t.Errorf("the target could not attack: %v", err)
	}
}

func TestOldB505IntimidationBoltThatKillsItsTargetStillStopsEveryAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, me.ID, "Target", oldB505BearType, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", oldB505BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Intimidation Bolt", "Instant", oldB505Bolt, effectstest.CardTargets(target))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Fatal("the target survived 3 damage")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(other, opp.ID); err == nil {
		t.Error("another creature attacked")
	}
}

func TestOldB505ThoughtHemorrhageDamagesPerRevealedCopyAndExilesThemAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mk := func() game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID}
	}
	var named []uuid.UUID
	var other uuid.UUID
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			h, y, l := mk(), mk(), mk()
			named = append(named, h.InstanceID, y.InstanceID, l.InstanceID)
			opp.Hand.PushTop(h)
			opp.Graveyard.PushTop(y)
			opp.Library.PushTop(l)
		}
		other = uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: other, Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	})
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Thought Hemorrhage", "Sorcery", oldB505Hemorrhage,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "lightning bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-6 {
		t.Errorf("life %d, want %d (3 damage for each of two revealed copies)", got, before-6)
	}
	for _, id := range named {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a Lightning Bolt is in %v, want exile", z)
		}
	}
	if z := g.FindCardZoneForEffect(other); z == nil || z.Kind != game.ZoneHand {
		t.Error("the Shock left the hand")
	}
}

func TestOldB505ThoughtHemorrhageNamingALandDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushHandCardOfType(opp, "Forest", "Basic Land — Forest")
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Thought Hemorrhage", "Sorcery", oldB505Hemorrhage,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "forest"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before {
		t.Errorf("life %d, want %d", got, before)
	}
	if z := g.FindCardZoneForEffect(forest); z == nil || z.Kind != game.ZoneHand {
		t.Error("the Forest was exiled")
	}
}

func TestOldB505CrystallizationStopsCombatAndExilesTheCreatureItIsTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB505BearType, 2, 2)
	aura := effectstest.Enchant(t, g, "Crystallization", oldB505Crystal, bear)
	if r := effectstest.AuraRestrictions(t, g, bear); r&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Fatalf("restrictions %v, want no attack or block", r)
	}
	if effectstest.ZoneOfCard(g, aura) != game.ZoneBattlefield {
		t.Fatal("the Aura is not on the battlefield")
	}
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", oldB505LightningID, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, bear) != game.ZoneExile {
		t.Errorf("the creature is in zone %v, want exile", effectstest.ZoneOfCard(g, bear))
	}
	if effectstest.ZoneOfCard(g, aura) == game.ZoneBattlefield {
		t.Error("the Aura stayed on the battlefield after its creature was exiled")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != before {
		t.Errorf("life changed %d -> %d", before, got)
	}
}

func TestOldB505SigilOfTheNayanGodsPumpsPerCreatureYouControlAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB505BearType, 2, 2)
	effectstest.B12Creature(g, me.ID, "Second", oldB505BearType, 1, 1)
	effectstest.B12Creature(g, opp.ID, "Theirs", oldB505BearType, 1, 1)
	effectstest.Enchant(t, g, "Sigil of the Nayan Gods", oldB505Sigil, bear)
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("the creature is %d/%d, want 4/4 (+2/+2 for two creatures you control)", p, tt)
	}
	effectstest.B12Creature(g, me.ID, "Third", oldB505BearType, 1, 1)
	if p, tt := effectstest.PtOf(t, g, bear); p != 5 || tt != 5 {
		t.Errorf("the creature is %d/%d, want 5/5 after another creature arrives", p, tt)
	}
	spec, _ := effects.Lookup(oldB505Sigil)
	if len(spec.Activated) != 1 || !spec.Activated[0].Cycling {
		t.Errorf("activated abilities %+v, want one cycling ability", spec.Activated)
	}
}

func TestOldB505KnotvinePaladinGetsPlusOnePerUntappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	paladin := effectstest.B12Push(g, me.ID, "Knotvine Paladin", "Creature — Human Knight", oldB505Paladin, 2, 2)
	effectstest.B12Creature(g, me.ID, "Idle One", oldB505BearType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Idle Two", oldB505BearType, 1, 1)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped", oldB505BearType, 1, 1)
	effectstest.B12Creature(g, opp.ID, "Theirs", oldB505BearType, 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, paladin)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, paladin); p != 4 || tt != 4 {
		t.Errorf("the Paladin is %d/%d, want 4/4 (two untapped creatures; it is tapped by attacking)", p, tt)
	}
}

func TestOldB505SkyclawThrashPumpsAndFliesOnlyOnAWonFlip(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 40 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		thrash := effectstest.B12Push(g, me.ID, "Skyclaw Thrash", "Artifact Creature — Lizard Warrior", oldB505Thrash, 4, 4)
		effectstest.DeclareAttack(t, g, opp.ID, thrash)
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) != flips+1 {
			t.Fatalf("%d coins were flipped, want 1", len(evs)-flips)
		}
		p, tt := effectstest.PtOf(t, g, thrash)
		flying := effectstest.HasEffectiveKeyword(t, g, thrash, "flying")
		if evs[len(evs)-1].Won {
			won++
			if p != 5 || tt != 5 || !flying {
				t.Fatalf("a won flip left %d/%d flying=%v, want 5/5 flying", p, tt, flying)
			}
		} else {
			lost++
			if p != 4 || tt != 4 || flying {
				t.Fatalf("a lost flip left %d/%d flying=%v, want 4/4 without flying", p, tt, flying)
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("saw %d won and %d lost flips; want both", won, lost)
	}
}

func TestOldB505IdentityCrisisExilesTheTargetsHandAndGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var theirs []uuid.UUID
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			h := game.Card{InstanceID: uuid.New(), Name: "Hand Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID}
			y := game.Card{InstanceID: uuid.New(), Name: "Yard Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID}
			theirs = append(theirs, h.InstanceID, y.InstanceID)
			opp.Hand.PushTop(h)
			opp.Graveyard.PushTop(y)
		}
	})
	mine := effectstest.PushHandCardOfType(me, "My Card", "Instant")
	libBefore := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Identity Crisis", "Sorcery", oldB505Identity,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range theirs {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a card is in %v, want exile", z)
		}
	}
	if opp.Hand.Size() != 0 || opp.Graveyard.Size() != 0 {
		t.Errorf("hand %d, graveyard %d, want both empty", opp.Hand.Size(), opp.Graveyard.Size())
	}
	if opp.Library.Size() != libBefore {
		t.Error("the library changed")
	}
	if z := g.FindCardZoneForEffect(mine); z == nil || z.Kind != game.ZoneHand {
		t.Error("the caster's own card left their hand")
	}
}

func oldB505CastBolt(t *testing.T, g *game.Game, me *game.Player, colors []string, at uuid.UUID) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: oldB505LightningID,
		Colors: colors, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: at}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
}

func TestOldB505ClovenCastingCopiesAMulticoloredSpellForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Cloven Casting", "Enchantment", oldB505Cloven, false)
	before := effectstest.LifeOf(g, opp.ID)
	oldB505CastBolt(t, g, me, []string{"R", "W"}, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4; i++ {
		if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
			if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("retarget: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != before-6 {
		t.Errorf("life %d, want %d (the Bolt and its copy)", got, before-6)
	}
}

func TestOldB505ClovenCastingIgnoresAMonocoloredSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Cloven Casting", "Enchantment", oldB505Cloven, false)
	oldB505CastBolt(t, g, me, []string{"R"}, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("Cloven Casting asked for {1} for a monocolored spell")
	}
}

func TestOldB505ClovenCastingDeclinedMakesNoCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Cloven Casting", "Enchantment", oldB505Cloven, false)
	before := effectstest.LifeOf(g, opp.ID)
	oldB505CastBolt(t, g, me, []string{"R", "W"}, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-3 {
		t.Errorf("life %d, want %d (no copy)", got, before-3)
	}
}
