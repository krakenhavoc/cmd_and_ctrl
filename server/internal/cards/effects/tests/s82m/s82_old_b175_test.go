package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB175Collision  = "197ce21b-3387-4e76-a49c-04b9c371c031"
	oldB175Boar       = "2b04b60b-0c4c-4fee-a54b-7052ca5190e1"
	oldB175GatesAblaz = "3b97bf8e-d5c4-4eae-a453-73456e0461a1"
	oldB175Carnival   = "b996ae86-3bf6-4e84-8302-6ec6edaa6ff0"
	oldB175CultGuild  = "c65f8949-5e37-4929-994a-ec037814bf63"
	oldB175Roustabout = "c696d82e-6dee-4766-bc08-33af00f61717"
	oldB175Immolation = "cf681c9e-a644-44c7-b744-34faafc67d50"
	oldB175Bolrac     = "dbfa5409-5706-4de8-8152-e74b8d93910a"
	oldB175Cavalcade  = "e29be33c-e5b0-428d-a057-b1cd61127557"
	oldB175SyndGuild  = "ec484fd1-aa9a-4056-becf-1d8c710272fd"
	oldB175Thrull     = "ec8a14d4-6ca0-43b1-abf6-0c129daf8350"
)

func TestOldB175AreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB175Collision: "Collision", oldB175Collision + "#1": "Colossus",
		oldB175Boar: "Charging War Boar", oldB175GatesAblaz: "Gates Ablaze",
		oldB175Carnival: "Carnival", oldB175Carnival + "#1": "Carnage",
		oldB175CultGuild: "Cult Guildmage", oldB175Roustabout: "Rakdos Roustabout",
		oldB175Immolation: "Immolation Shaman", oldB175Bolrac: "Bolrac-Clan Crusher",
		oldB175Cavalcade: "Cavalcade of Calamity", oldB175SyndGuild: "Syndicate Guildmage",
		oldB175Thrull: "Grasping Thrull",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB175CastSplit casts a free split card from the active player's hand.
func oldB175CastSplit(t *testing.T, g *game.Game, oracle string, face int, leftType, rightType string, targets []game.TargetRef, strict bool) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Left", TypeLine: leftType, ManaCost: "{0}"},
			{Name: "Right", TypeLine: rightType, ManaCost: "{0}"},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: face, Targets: targets, Strict: strict})
}

func oldB175Damage(g *game.Game, id uuid.UUID) int {
	n := 0
	g.WithWriteLock(func() { n = effectstest.FindBattlefieldCardForTest(g, id).DamageMarked })
	return n
}

func TestOldB175CollisionHitsOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 9)
	flyer := effectstest.B12Creature(g, opp.ID, "Flyer", "Creature — Bird", 2, 9)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, flyer).Keywords = []string{"flying"} })
	if err := oldB175CastSplit(t, g, oldB175Collision, 0, "Instant", "Instant", effectstest.CardTarget(ground), true); err == nil {
		t.Fatal("Collision accepted a creature without flying")
	}
	if err := oldB175CastSplit(t, g, oldB175Collision, 0, "Instant", "Instant", effectstest.CardTarget(flyer), true); err != nil {
		t.Fatalf("Collision on a flyer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB175Damage(g, flyer); got != 6 {
		t.Errorf("damage %d, want 6", got)
	}
}

func TestOldB175ColossusGivesPlusFourPlusTwoAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB175CastSplit(t, g, oldB175Collision, 1, "Instant", "Instant", effectstest.CardTarget(bear), true); err != nil {
		t.Fatalf("Colossus: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p, tough := effectstest.EffectivePTOf(t, g, bear)
	trample := effectstest.HasKeywordOf(g, bear, "trample")
	if p != 6 || tough != 4 || !trample {
		t.Errorf("got %d/%d trample=%v, want 6/4 with trample", p, tough, trample)
	}
}

func TestOldB175ChargingWarBoarNeedsADomriPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	boar := effectstest.PushCatalogPermanent(g, me.ID, "Charging War Boar", "Creature — Boar", oldB175Boar, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, boar)
		c.Power, c.Toughness = 3, 1
	})
	read := func() (int, int, bool) {
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		p, tough := effectstest.EffectivePTOf(t, g, boar)
		return p, tough, effectstest.HasKeywordOf(g, boar, "trample")
	}
	if p, tough, tr := read(); p != 3 || tough != 1 || tr {
		t.Errorf("alone: %d/%d trample=%v, want 3/1 without trample", p, tough, tr)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Chandra", "Legendary Planeswalker — Chandra", "", false)
	if p, tough, tr := read(); p != 3 || tough != 1 || tr {
		t.Errorf("with a non-Domri walker: %d/%d trample=%v, want 3/1 without trample", p, tough, tr)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Domri Rade", "Legendary Planeswalker — Domri", "", false)
	if p, tough, tr := read(); p != 4 || tough != 2 || !tr {
		t.Errorf("with Domri: %d/%d trample=%v, want 4/2 with trample", p, tough, tr)
	}
}

func TestOldB175GatesAblazeDealsDamageEqualToYourGates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Gate One", "Land — Gate", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Gate Two", "Land — Gate", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Gate", "Land — Gate", "", false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Ogre", 1, 9)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 1, 9)
	effectstest.CastCatalogSpell(t, g, "Gates Ablaze", "Sorcery", oldB175GatesAblaz, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if got := oldB175Damage(g, id); got != 2 {
			t.Errorf("damage %d, want 2 (two Gates you control)", got)
		}
	}
}

func TestOldB175CarnivalHitsThePermanentAndItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 9)
	life := opp.Life
	if err := oldB175CastSplit(t, g, oldB175Carnival, 0, "Instant", "Sorcery", effectstest.CardTarget(bear), true); err != nil {
		t.Fatalf("Carnival: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB175Damage(g, bear); got != 1 {
		t.Errorf("creature damage %d, want 1", got)
	}
	if opp.Life != life-1 {
		t.Errorf("controller life %d, want %d", opp.Life, life-1)
	}
}

func TestOldB175CarnageDamagesAnOpponentAndMakesThemDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.HandCardOf(g, opp, "A", "Sorcery", "", "")
	b := effectstest.HandCardOf(g, opp, "B", "Sorcery", "", "")
	effectstest.HandCardOf(g, opp, "C", "Sorcery", "", "")
	if err := oldB175CastSplit(t, g, oldB175Carnival, 1, "Instant", "Sorcery", effectstest.Pr6Player(me.ID), true); err == nil {
		t.Fatal("Carnage accepted its own caster as the target opponent")
	}
	life := opp.Life
	if err := oldB175CastSplit(t, g, oldB175Carnival, 1, "Instant", "Sorcery", effectstest.Pr6Player(opp.ID), true); err != nil {
		t.Fatalf("Carnage: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, a, b)
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) {
		t.Error("the two chosen cards were not discarded")
	}
}

func TestOldB175CultGuildmageAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.PushCatalogPermanent(g, me.ID, "Cult Guildmage", "Creature — Human Shaman", oldB175CultGuild, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 1)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Targets: effectstest.Pr6Player(me.ID), Strict: true}); err == nil {
		t.Error("the damage ability targeted its own controller")
	}
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Targets: effectstest.Pr6Player(opp.ID), Strict: true}); err != nil {
		t.Fatalf("{R},{T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
}

func TestOldB175CultGuildmageDiscardIsSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.PushCatalogPermanent(g, me.ID, "Cult Guildmage", "Creature — Human Shaman", oldB175CultGuild, false)
	card := effectstest.HandCardOf(g, opp, "Victim", "Sorcery", "", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 4)
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Player(opp.ID), Strict: true}); err != nil {
		t.Fatalf("{3}{B},{T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, card)
	if !opp.Graveyard.Contains(card) {
		t.Error("the target player did not discard")
	}
	// With the ability on the stack the activation window is closed.
	mage2 := effectstest.PushCatalogPermanent(g, me.ID, "Cult Guildmage", "Creature — Human Shaman", oldB175CultGuild, false)
	effectstest.FillPoolColored(me, "B", 4)
	effectstest.CastCatalogSpell(t, g, "Hold", "Sorcery", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, mage2, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Player(opp.ID), Strict: true}); err == nil {
		t.Error("the discard ability was activated with a spell on the stack")
	}
}

func TestOldB175SyndicateGuildmageTapsOnlyPowerFourOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.PushCatalogPermanent(g, me.ID, "Syndicate Guildmage", "Creature — Human Cleric", oldB175SyndGuild, false)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "W", 2)
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(small), Strict: true}); err == nil {
		t.Fatal("tapped a creature with power 3")
	}
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(big), Strict: true}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, big) {
		t.Error("the big creature is not tapped")
	}
}

func TestOldB175SyndicateGuildmageDealsTwoToAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.PushCatalogPermanent(g, me.ID, "Syndicate Guildmage", "Creature — Human Cleric", oldB175SyndGuild, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 5)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Targets: effectstest.Pr6Player(opp.ID), Strict: true}); err != nil {
		t.Fatalf("{4}{B},{T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
}

func TestOldB175RakdosRoustaboutHitsTheDefenderOnceHoweverManyBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	rakdos := effectstest.PushCatalogPermanent(g, me.ID, "Rakdos Roustabout", "Creature — Ogre Warrior", oldB175Roustabout, false)
	a := effectstest.PushVanillaCreature(g, them.ID, "Wall A", 0, 5)
	b := effectstest.PushVanillaCreature(g, them.ID, "Wall B", 0, 5)
	life := them.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rakdos, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{a, b} {
		if err := g.DeclareBlocker(w, rakdos); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	if them.Life != life-1 {
		t.Errorf("life %d, want %d (one point however many block)", them.Life, life-1)
	}
}

func TestOldB175ImmolationShamanPunishesAnOpponentsNonManaActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Immolation Shaman", "Creature — Lizard Shaman", oldB175Immolation, false)
	theirKrenko := effectstest.ActPushKrenko(g, them.ID)
	myKrenko := effectstest.ActPushKrenko(g, me.ID)
	before := them.Life
	if err := g.ActivateCatalogAbility(me.ID, myKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("own Krenko: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != before {
		t.Errorf("their life %d after our own activation, want %d", them.Life, before)
	}
	if err := g.ActivateCatalogAbility(them.ID, theirKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("their Krenko: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if them.Life != before-1 || me.Life != 40 {
		t.Errorf("lives %d / %d, want %d / 40", them.Life, me.Life, before-1)
	}
}

func TestOldB175ImmolationShamanPumpsAndGainsMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shaman := effectstest.PushCatalogPermanent(g, me.ID, "Immolation Shaman", "Creature — Lizard Shaman", oldB175Immolation, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, shaman)
		c.Power, c.Toughness = 1, 3
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 5)
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p, tough := effectstest.EffectivePTOf(t, g, shaman)
	menace := effectstest.HasKeywordOf(g, shaman, "menace")
	if p != 4 || tough != 6 || !menace {
		t.Errorf("got %d/%d menace=%v, want 4/6 with menace", p, tough, menace)
	}
}

func TestOldB175BolracClanCrusherRemovesACounterToDealTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusher := effectstest.PushCatalogPermanent(g, me.ID, "Bolrac-Clan Crusher", "Creature — Ogre Warrior", oldB175Bolrac, false)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, crusher, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Player(opp.ID), Strict: true, CounterSourceIDs: []uuid.UUID{other}}); err == nil {
		t.Fatal("activated with no +1/+1 counter to remove")
	}
	effectstest.B28AddCounter(t, g, other, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, crusher, 0, game.ActivateAbilityParams{Targets: effectstest.Pr6Player(opp.ID), Strict: true, CounterSourceIDs: []uuid.UUID{other}}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	if !effectstest.B16Tapped(t, g, crusher) {
		t.Error("the crusher was not tapped as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	if got := effectstest.CounterCount(g, other, game.CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters left %d, want 0", got)
	}
}

func TestOldB175CavalcadeOfCalamityPingsForEachAttackerWithPowerOneOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Cavalcade of Calamity", "Enchantment", oldB175Cavalcade, 0, 0)
	small := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	zero := effectstest.B12Creature(g, me.ID, "Zero", "Creature — Elf", 0, 3)
	two := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, small, zero, two)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d (power 1 and 0 attack, power 2 does not)", opp.Life, life-2)
	}
}

func TestOldB175GraspingThrullDrainsOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	oppLives := effectstest.LifeOfOpponents(g)
	effectstest.CastCatalogSpell(t, g, "Grasping Thrull", "Creature — Thrull", oldB175Thrull, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	now := effectstest.LifeOfOpponents(g)
	for i := range now {
		if now[i] != oppLives[i]-2 {
			t.Errorf("opponent %d life %d, want %d", i, now[i], oppLives[i]-2)
		}
	}
}
