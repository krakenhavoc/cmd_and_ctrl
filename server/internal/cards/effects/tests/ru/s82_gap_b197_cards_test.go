package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b197_cards_test.go — Warhammer 40,000 Commander enchantments and
// instants, GAP batch 197 (#5948). Each test plays the real catalog entry
// through the engine.

const (
	gap197WarInHeavenOracle   = "5924cbb0-19b5-449c-a7c5-722e744a54f8"
	gap197RuinousOracle       = "283054ed-5fbd-4202-bc02-294079aaff03"
	gap197HorusOracle         = "fc315174-d063-41dd-9379-11d95c5fd4ed"
	gap197BirthOracle         = "7f0544e3-965f-4b12-baaf-d7dfea8d4d41"
	gap197FleshOracle         = "dafbb927-3d9d-4be7-b427-f6564b295161"
	gap197TyrannicOracle      = "85fd2f65-957e-446b-990a-8711a1e1c4f6"
	gap197LostDamnedOracle    = "22e70ad0-1a9c-4d9e-8cc1-ebea485282ca"
	gap197DefendersOracle     = "7662af07-1ebc-4545-bffd-7a2e0c513b71"
	gap197BileBlightOracle    = "0dd95516-893d-477e-a38c-3cc37c6e7ae0"
	gap197NurgleOracle        = "967f0c23-7af2-4e86-9ebc-dc2f8856c7c5"
	gap197BloodGodOracle      = "a6079dff-9479-4665-aba5-2cddefb9a2bc"
	gap197KillMaimBurnOracle  = "05c7d88d-c081-411b-a530-969b5816d792"
	gap197SagaType            = "Enchantment — Saga"
	gap197CreatureCardInGrave = "Creature — Test"
)

func TestGap197EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap197WarInHeavenOracle:  "The War in Heaven",
		gap197RuinousOracle:      "The Ruinous Powers",
		gap197HorusOracle:        "The Horus Heresy",
		gap197BirthOracle:        "Birth of the Imperium",
		gap197FleshOracle:        "The Flesh Is Weak",
		gap197TyrannicOracle:     "The First Tyrannic War",
		gap197LostDamnedOracle:   "The Lost and the Damned",
		gap197DefendersOracle:    "Defenders of Humanity",
		gap197BileBlightOracle:   "Bile Blight",
		gap197NurgleOracle:       "Nurgle's Conscription",
		gap197BloodGodOracle:     "Blood for the Blood God!",
		gap197KillMaimBurnOracle: "Kill! Maim! Burn!",
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

// gap197Grave puts a creature card with a mana cost into a graveyard.
func gap197Grave(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: gap197CreatureCardInGrave, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// gap197Pick answers the open pick-target prompt of `chooser` with `want`
// (those the prompt offers), returning ResolvePickTargets' error.
func gap197Pick(t *testing.T, g *game.Game, chooser uuid.UUID, want ...uuid.UUID) error {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoicePickTarget || c.Chooser != chooser {
			continue
		}
		var refs []game.TargetRef
		for _, offered := range c.PickTargetCards {
			for _, w := range want {
				if offered == w {
					refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: w})
				}
			}
		}
		return g.ResolvePickTargets(c.ID, chooser, refs)
	}
	t.Fatal("no pick-target prompt is open")
	return nil
}

// gap197Saga casts a Saga and settles chapter I's trigger.
func gap197Saga(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, gap197SagaType, oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// --- Bile Blight ---------------------------------------------------------------

func TestGap197BileBlightShrinksTheTargetAndEveryCreatureWithItsName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 4, 4)
	b := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 4, 4)
	c := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 4, 4)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Bile Blight", "Instant", gap197BileBlightOracle, []game.TargetRef{{Kind: game.TargetCard, ID: a}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 1 || tough != 1 {
			t.Errorf("a Goblin is %d/%d, want 1/1", p, tough)
		}
	}
	if p := effectstest.EffectivePower(t, g, ogre); p != 4 {
		t.Errorf("the Ogre is %d power, want 4", p)
	}
	// "Until end of turn" locks the set at resolution (CR 611.2c).
	late := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 4, 4)
	if p := effectstest.EffectivePower(t, g, late); p != 4 {
		t.Errorf("a Goblin that entered later is %d power, want 4", p)
	}
}

func TestGap197BileBlightKillsAllSameNameCreaturesWithSmallToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Rat", "Creature — Rat", 2, 3)
	b := effectstest.B12Creature(g, g.Seats[2].ID, "Rat", "Creature — Rat", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Bile Blight", "Instant", gap197BileBlightOracle, []game.TargetRef{{Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a Rat survived -3/-3")
	}
}

func TestGap197BileBlightRefusesANonCreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.B12Permanent(g, g.Seats[1].ID, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Bile Blight", "Instant", gap197BileBlightOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("an artifact is not a legal target")
	}
}

// --- Blood for the Blood God! -----------------------------------------------------

func TestGap197BloodCostsOneLessForEachCreatureThatDiedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	if mv := effectstest.SelfPricedMV(t, g, me, gap197BloodGodOracle, "Instant", "{8}{B}{B}{R}"); mv != 11 {
		t.Fatalf("with no deaths the spell costs %d, want 11", mv)
	}
	for i := 0; i < 3; i++ {
		effectstest.B15Destroy(t, g, effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2))
	}
	if mv := effectstest.SelfPricedMV(t, g, me, gap197BloodGodOracle, "Instant", "{8}{B}{B}{R}"); mv != 8 {
		t.Errorf("after three deaths the spell costs %d, want 8", mv)
	}
	for i := 0; i < 9; i++ {
		effectstest.B15Destroy(t, g, effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2))
	}
	if mv := effectstest.SelfPricedMV(t, g, me, gap197BloodGodOracle, "Instant", "{8}{B}{B}{R}"); mv != 3 {
		t.Errorf("after twelve deaths the spell costs %d, want 3 (only generic mana is reduced)", mv)
	}
}

func TestGap197BloodDiscardsDrawsEightDamagesEachOpponentAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B13HandCard(me, "Old A", "Sorcery", "")
	effectstest.B13HandCard(me, "Old B", "Sorcery", "")
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		before[p.ID] = p.Life
	}
	myLife := me.Life
	spell := effectstest.CastCatalogSpell(t, g, "Blood for the Blood God!", "Instant", gap197BloodGodOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 8 {
		t.Errorf("hand is %d cards, want exactly the eight drawn", got)
	}
	for _, p := range g.Seats[1:] {
		if p.Life != before[p.ID]-8 {
			t.Errorf("an opponent is at %d, want %d", p.Life, before[p.ID]-8)
		}
	}
	if me.Life != myLife {
		t.Errorf("the caster's life changed to %d", me.Life)
	}
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("Blood for the Blood God! did not exile itself")
	}
}

// --- Kill! Maim! Burn! -----------------------------------------------------------

func TestGap197KillMaimBurnRunsEveryChosenModeOnItsOwnTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.CastModal(t, g, "Kill! Maim! Burn!", "Instant", gap197KillMaimBurnOracle, []int{0, 1, 2}, []game.TargetRef{
		{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: opp.ID},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(bear) {
		t.Error("a destroy mode left its target on the battlefield")
	}
	if opp.Life != life-3 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-3)
	}
}

func TestGap197KillMaimBurnOneModeTouchesOnlyItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Kill! Maim! Burn!", "Instant", gap197KillMaimBurnOracle, []int{1}, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) {
		t.Error("only the creature should have been destroyed")
	}
}

// --- Nurgle's Conscription -------------------------------------------------------

func TestGap197NurglesConscriptionStealsACreatureTappedAndExilesThatGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := gap197Grave(opp, "Fallen", "{2}{B}")
	rest := gap197Grave(opp, "Other", "{1}")
	mine := gap197Grave(me, "Mine", "{1}")
	effectstest.CastCatalogSpell(t, g, "Nurgle's Conscription", "Instant", gap197NurgleOracle, []game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, target)
	if !ok || c.Controller != me.ID || !c.Tapped {
		t.Fatalf("the creature is not on the battlefield tapped under my control: %+v ok=%v", c, ok)
	}
	if !g.Exile.Contains(rest) || opp.Graveyard.Size() != 0 {
		t.Error("the rest of that player's graveyard was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard was exiled")
	}
}

func TestGap197NurglesConscriptionRefusesYourOwnGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	own := gap197Grave(g.Seats[0], "Mine", "{1}")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Nurgle's Conscription", "Instant", gap197NurgleOracle, []game.TargetRef{{Kind: game.TargetCard, ID: own}}); err == nil {
		t.Fatal("a creature card in your own graveyard is not a legal target")
	}
}

// --- Defenders of Humanity -------------------------------------------------------

func gap197Warriors(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.IsToken() && c.Name == "Astartes Warrior" {
			n++
		}
	}
	return n
}

func TestGap197DefendersOfHumanityEntersWithXWarriors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	active := me
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Defenders of Humanity", TypeLine: "Enchantment", OracleID: gap197DefendersOracle,
		ManaCost: "{X}{2}{W}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap197Warriors(g, me.ID); got != 3 {
		t.Fatalf("%d Astartes Warriors, want 3", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Astartes Warrior" {
			continue
		}
		if !c.HasSubtype("Astartes") || !c.HasSubtype("Warrior") || !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "vigilance") ||
			effectstest.EffectivePower(t, g, c.InstanceID) != 2 || effectstest.EffectiveToughness(t, g, c.InstanceID) != 2 {
			t.Errorf("token %q is not a 2/2 white vigilant Astartes Warrior", c.TypeLine)
		}
	}
}

func TestGap197DefendersOfHumanityActivationNeedsNoCreaturesAndYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	def := effectstest.PushCatalogPermanent(g, me.ID, "Defenders of Humanity", "Enchantment", gap197DefendersOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, def, 0, game.ActivateAbilityParams{XValue: 2}); err == nil {
		t.Fatal("activated while controlling a creature")
	}
	effectstest.B15Destroy(t, g, bear)
	if err := g.ActivateCatalogAbility(me.ID, def, 0, game.ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("activate with no creatures: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap197Warriors(g, me.ID); got != 2 {
		t.Errorf("%d Warriors, want 2", got)
	}
	if g.Battlefield.Contains(def) || !g.Exile.Contains(def) {
		t.Error("the enchantment should have been exiled as a cost")
	}
}

func TestGap197DefendersOfHumanityCannotBeActivatedOnAnotherTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	def := effectstest.PushCatalogPermanent(g, opp.ID, "Defenders of Humanity", "Enchantment", gap197DefendersOracle, false)
	if err := g.ActivateCatalogAbility(opp.ID, def, 0, game.ActivateAbilityParams{XValue: 1}); err == nil {
		t.Fatal("activated during another player's turn")
	}
}

// --- Birth of the Imperium -------------------------------------------------------

func TestGap197BirthOfTheImperiumRunsAllThreeChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	// a controls fewer creatures than me, b the same, c more.
	effectstest.B12Creature(g, a.ID, "A1", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, b.ID, "B1", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, b.ID, "B2", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, b.ID, "B3", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, b.ID, "B4", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, c.ID, "C1", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, c.ID, "C2", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, c.ID, "C3", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, c.ID, "C4", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, c.ID, "C5", "Creature — Bear", 1, 1)
	saga := gap197Saga(t, g, "Birth of the Imperium", gap197BirthOracle)
	if got := gap197Warriors(g, me.ID); got != 3 {
		t.Fatalf("chapter I made %d Warriors, want 3 (one per opponent)", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	gap197Sacrifice(t, g, a, b, c)
	effectstest.PassPriorityAroundTable(t, g)
	// Chapter II: each opponent sacrificed one creature (A had one, B four, C five).
	if got := countControlledCreatures(g, a.ID); got != 0 {
		t.Errorf("A controls %d creatures, want 0", got)
	}
	if got := countControlledCreatures(g, b.ID); got != 3 {
		t.Errorf("B controls %d creatures, want 3", got)
	}
	if got := countControlledCreatures(g, c.ID); got != 4 {
		t.Errorf("C controls %d creatures, want 4", got)
	}
	if got := countControlledCreatures(g, me.ID); got != 3 {
		t.Errorf("I control %d creatures, want my 3 Warriors", got)
	}
	hand := me.Hand.Size()
	effectstest.HobA9NextChapter(t, g)
	// Chapter III: A has 0 < 3 (fewer), B has 3 (same), C has 4 (more): two cards.
	if got := me.Hand.Size() - hand; got != 2+1 { // +1 for the draw step of my turn
		t.Errorf("drew %d cards over chapter III's turn, want 2 plus the draw step", got)
	}
	_ = saga
}

// gap197Sacrifice answers each seat's open sacrifice prompt with its first candidate.
func gap197Sacrifice(t *testing.T, g *game.Game, seats ...*game.Player) {
	t.Helper()
	for _, p := range seats {
		ch := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, p.ID)
		if ch == nil {
			continue
		}
		if err := g.ResolveChooseCards(ch.ID, p.ID, ch.ChooseCards[:1]); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	}
}

func countControlledCreatures(g *game.Game, p uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p && c.IsCreature() {
			n++
		}
	}
	return n
}

// --- The Flesh Is Weak -----------------------------------------------------------

func TestGap197TheFleshIsWeakCountersMakeArtifactsAndNonartifactsShrink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	golem := effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	theirGolem := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	effectstest.CastCatalogSpell(t, g, "The Flesh Is Weak", "Enchantment", gap197FleshOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, golem} {
		if effectstest.PlusOnes(g, id) != 1 {
			t.Errorf("a creature of mine has %d +1/+1 counters, want 1", effectstest.PlusOnes(g, id))
		}
		if c, _ := effectstest.BattlefieldCard(g, id); !c.IsArtifact() {
			t.Errorf("%s with a +1/+1 counter is not an artifact", c.Name)
		}
	}
	if effectstest.PlusOnes(g, theirs) != 0 {
		t.Error("an opponent's creature got a counter")
	}
	// My Bear: 2/2 + counter, now an artifact so not shrunk = 3/3.
	if c, _ := effectstest.BattlefieldCard(g, mine); c.PowerForComparison() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("my Bear is %d/%d, want 3/3", c.PowerForComparison(), c.CurrentToughness())
	}
	if c, _ := effectstest.BattlefieldCard(g, golem); c.PowerForComparison() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("my Golem is %d/%d, want 3/3", c.PowerForComparison(), c.CurrentToughness())
	}
	// Their nonartifact Bear is -1/-1; their Golem is an artifact and is not.
	if p, tough := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, theirGolem); p != 3 {
		t.Errorf("their Golem is %d power, want 3", p)
	}
}

func TestGap197TheFleshIsWeakKillsAOneOneNonartifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	token := effectstest.B12Creature(g, opp.ID, "Scout", "Creature — Human", 1, 1)
	effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "The Flesh Is Weak", "Enchantment", gap197FleshOracle, false)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(token) {
		t.Error("a nonartifact 1/1 survived -1/-1")
	}
}

// --- The Lost and the Damned -----------------------------------------------------

func gap197Spawns(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.IsToken() && c.Name == "Spawn" {
			n++
		}
	}
	return n
}

func TestGap197TheLostAndTheDamnedTriggersOnALandNotFromYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "The Lost and the Damned", "Enchantment", gap197LostDamnedOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap197Spawns(g, me.ID); got != 0 {
		t.Fatalf("a land played from hand made %d Spawn, want 0", got)
	}
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(land, game.ZoneBattlefield); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap197Spawns(g, me.ID); got != 1 {
		t.Fatalf("a land from the graveyard made %d Spawn, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spawn" && (!c.HasSubtype("Spawn") || !c.IsCreature() ||
			effectstest.EffectivePower(t, g, c.InstanceID) != 3 || effectstest.EffectiveToughness(t, g, c.InstanceID) != 3 ||
			len(c.Colors) != 1 || c.Colors[0] != "R") {
			t.Errorf("the token is %q %v, want a 3/3 red Spawn creature", c.TypeLine, c.Colors)
		}
	}
}

func TestGap197TheLostAndTheDamnedIgnoresACreatureNotFromYourHandAndAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "The Lost and the Damned", "Enchantment", gap197LostDamnedOracle, false)
	creature := gap197Grave(me, "Fallen", "{1}")
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(creature, game.ZoneBattlefield); err != nil {
			t.Fatal(err)
		}
	})
	land := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(land, game.ZoneBattlefield); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap197Spawns(g, me.ID) + gap197Spawns(g, opp.ID); got != 0 {
		t.Errorf("%d Spawn, want 0: a creature is not a land and an opponent's land is not mine", got)
	}
}

func TestGap197TheLostAndTheDamnedTriggersOnASpellCastNotFromYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "The Lost and the Damned", "Enchantment", gap197LostDamnedOracle, false)
	// From the hand: nothing.
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap197Spawns(g, me.ID); got != 0 {
		t.Fatalf("a spell cast from hand made %d Spawn", got)
	}
	// From exile and from the graveyard: one each, as the cast event names its origin.
	for _, from := range []game.ZoneKind{game.ZoneExile, game.ZoneGraveyard} {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventCast, Actor: me.ID, CardID: uuid.New(), OldZone: from})
		})
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap197Spawns(g, me.ID); got != 2 {
		t.Errorf("%d Spawn after two casts from other zones, want 2", got)
	}
	// An opponent's cast is not mine.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: g.Seats[1].ID, CardID: uuid.New(), OldZone: game.ZoneExile})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap197Spawns(g, me.ID); got != 2 {
		t.Errorf("an opponent's cast made a Spawn: %d", got)
	}
}

// --- The Ruinous Powers ----------------------------------------------------------

// gap197TopOf stacks the top of a library with copies of one card: the
// opponents draw on the turns that pass before my upkeep, so one copy would
// be gone.
func gap197TopOf(p *game.Player, name, typeLine, cost string) {
	for i := 0; i < 6; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	}
}

// gap197AdvanceToMyUpkeep walks the turns around the table to seat 0's next upkeep.
func gap197AdvanceToMyUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 {
			effectstest.PassPriorityAroundTable(t, g)
			return
		}
	}
	t.Fatal("never reached my upkeep")
}

// gap197ExiledCards lists the cards in exile, whoever owns them.
func gap197ExiledCards(g *game.Game) []game.Card {
	return append([]game.Card(nil), g.Exile.Cards...)
}

func TestGap197TheRuinousPowersExilesAnOpponentsTopCardAndCostsThemItsManaValueWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Every opponent's top card is a 3-mana creature, so the random pick does not matter.
	for _, p := range g.Seats[1:] {
		gap197TopOf(p, "Their Ogre", "Creature — Ogre", "{2}{R}")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "The Ruinous Powers", "Enchantment", gap197RuinousOracle, false)
	gap197AdvanceToMyUpkeep(t, g)
	exiled := gap197ExiledCards(g)
	if len(exiled) != 1 || exiled[0].Name != "Their Ogre" {
		t.Fatalf("exile holds %d cards, want exactly one opponent's Ogre: %+v", len(exiled), exiled)
	}
	owner := g.PlayerByID(exiled[0].Owner)
	if owner == nil || owner.ID == me.ID {
		t.Fatal("the exiled card is not an opponent's")
	}
	life := owner.Life
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, exiled[0].InstanceID, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast the exiled card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if owner.Life != life-3 {
		t.Errorf("the owner is at %d, want %d (its mana value, 3, lost)", owner.Life, life-3)
	}
}

func TestGap197TheRuinousPowersPermissionLastsOnlyUntilEndOfTurnAndNeedsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats[1:] {
		gap197TopOf(p, "Their Ogre", "Creature — Ogre", "{2}{R}")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "The Ruinous Powers", "Enchantment", gap197RuinousOracle, false)
	gap197AdvanceToMyUpkeep(t, g)
	exiled := gap197ExiledCards(g)
	if len(exiled) != 1 {
		t.Fatalf("exile holds %d cards, want 1", len(exiled))
	}
	// An opponent (the card's owner) may not cast it: the permission is mine.
	owner := g.PlayerByID(exiled[0].Owner)
	if err := g.CastSpell(owner.ID, exiled[0].InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("the card's owner cast it without a permission")
	}
	// Once my turn is over the card can no longer be cast.
	for i := 0; i < 400 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, exiled[0].InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("the exiled card was still castable after the turn ended")
	}
}

func TestGap197TheRuinousPowersDoesNotCostALifeForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats[1:] {
		gap197TopOf(p, "Their Forest", "Basic Land — Forest", "")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "The Ruinous Powers", "Enchantment", gap197RuinousOracle, false)
	gap197AdvanceToMyUpkeep(t, g)
	exiled := gap197ExiledCards(g)
	if len(exiled) != 1 {
		t.Fatalf("exile holds %d cards, want 1", len(exiled))
	}
	owner := g.PlayerByID(exiled[0].Owner)
	life := owner.Life
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, exiled[0].InstanceID, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the exiled land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if owner.Life != life {
		t.Errorf("a land played this way cost its owner %d life", life-owner.Life)
	}
}

// --- The War in Heaven -----------------------------------------------------------

func TestGap197TheWarInHeavenDrawsMillsAndReturnsNecrodermisArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := gap197Grave(me, "Alpha", "{2}{B}")
	b := gap197Grave(me, "Beta", "{2}{B}")
	c := gap197Grave(me, "Gamma", "{1}{B}")
	d := gap197Grave(me, "Delta", "{3}{B}")
	hand, life := me.Hand.Size(), me.Life
	gap197Saga(t, g, "The War in Heaven", gap197WarInHeavenOracle)
	if me.Hand.Size()-hand != 3 || me.Life != life-3 {
		t.Fatalf("chapter I: hand +%d life %d, want +3 and %d", me.Hand.Size()-hand, me.Life, life-3)
	}
	lib := me.Library.Size()
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - me.Library.Size(); got != 3+1 { // three milled and my draw step
		t.Fatalf("chapter II: library shrank by %d, want 3 plus the draw step", got)
	}
	// The milled cards are not creatures with a cost, so the four seeded ones are still the candidates.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	// A total of 3+3+2+4 = 12 is refused; 3+3+2 = 8 is allowed.
	if err := gap197Pick(t, g, me.ID, a, b, c, d); err == nil {
		t.Fatal("four creature cards with total mana value 12 were accepted")
	}
	if err := gap197Pick(t, g, me.ID, a, b, c); err != nil {
		t.Fatalf("three creature cards with total mana value 8 refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		card, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Errorf("%s did not return", card.Name)
			continue
		}
		if card.Controller != me.ID || !card.IsCreature() || !card.IsArtifact() || effectstest.CounterCount(g, id, "necrodermis") != 1 {
			t.Errorf("%s: creature %v artifact %v necrodermis %d", card.Name, card.IsCreature(), card.IsArtifact(), effectstest.CounterCount(g, id, "necrodermis"))
		}
	}
	if !me.Graveyard.Contains(d) {
		t.Error("the unchosen card left the graveyard")
	}
}

// --- The First Tyrannic War ------------------------------------------------------

func gap197Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Permanent(g, owner, "Forest", "Basic Land — Forest")
	}
}

func gap197FirstTyrannicWarPick(t *testing.T, g *game.Game, me *game.Player, pick ...uuid.UUID) {
	t.Helper()
	ch := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if ch == nil {
		t.Fatal("no creature prompt for chapter I")
	}
	if ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want a may-put of one", ch.ChooseMin, ch.ChooseMax)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, pick); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGap197TheFirstTyrannicWarPutsAnXCreatureInWithLandCountCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gap197Lands(g, me.ID, 3)
	hydra := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: hydra, Name: "Hydra", TypeLine: "Creature — Hydra", ManaCost: "{X}{G}", Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	gap197Saga(t, g, "The First Tyrannic War", gap197TyrannicOracle)
	gap197FirstTyrannicWarPick(t, g, me, hydra)
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("the Hydra did not enter")
	}
	if got := effectstest.PlusOnes(g, hydra); got != 3 {
		t.Errorf("the Hydra has %d +1/+1 counters, want 3 (the lands I control)", got)
	}
	if !me.Hand.Contains(bear) {
		t.Error("the unpicked creature left my hand")
	}
}

func TestGap197TheFirstTyrannicWarPutsAPlainCreatureInWithNoCountersAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gap197Lands(g, me.ID, 3)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	gap197Saga(t, g, "The First Tyrannic War", gap197TyrannicOracle)
	gap197FirstTyrannicWarPick(t, g, me, bear)
	if !g.Battlefield.Contains(bear) || effectstest.PlusOnes(g, bear) != 0 {
		t.Errorf("the Bear: on battlefield %v, counters %d, want in play with none", g.Battlefield.Contains(bear), effectstest.PlusOnes(g, bear))
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	kept := uuid.New()
	me2.Hand.PushTop(game.Card{InstanceID: kept, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me2.ID, Controller: me2.ID})
	gap197Saga(t, g2, "The First Tyrannic War", gap197TyrannicOracle)
	gap197FirstTyrannicWarPick(t, g2, me2)
	if g2.Battlefield.Contains(kept) || !me2.Hand.Contains(kept) {
		t.Error("declining should leave the creature in hand")
	}
}

func TestGap197TheFirstTyrannicWarDoublesEveryKindOfCounterOnChaptersTwoAndThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	target := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ox", TypeLine: "Creature — Ox",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 1, "shield": 2}})
	gap197Saga(t, g, "The First Tyrannic War", gap197TyrannicOracle)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	_ = gap197Pick(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, target) != 2 || effectstest.CounterCount(g, target, "shield") != 4 {
		t.Fatalf("after chapter II: %d +1/+1 and %d shield, want 2 and 4", effectstest.PlusOnes(g, target), effectstest.CounterCount(g, target, "shield"))
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	_ = gap197Pick(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, target) != 4 || effectstest.CounterCount(g, target, "shield") != 8 {
		t.Errorf("after chapter III: %d +1/+1 and %d shield, want 4 and 8", effectstest.PlusOnes(g, target), effectstest.CounterCount(g, target, "shield"))
	}
}

// --- The Horus Heresy ------------------------------------------------------------

func TestGap197TheHorusHeresyStealsOneNonlegendaryCreaturePerOpponentWhileItRemains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	aBear := effectstest.B12Creature(g, a.ID, "A Bear", "Creature — Bear", 2, 2)
	aKing := effectstest.B12Creature(g, a.ID, "A King", "Legendary Creature — Human", 5, 5)
	bOx := effectstest.B12Creature(g, b.ID, "B Ox", "Creature — Ox", 3, 3)
	saga := effectstest.CastCatalogSpell(t, g, "The Horus Heresy", gap197SagaType, gap197HorusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			for _, offered := range c.PickTargetCards {
				if offered == aKing {
					t.Error("a legendary creature was offered")
				}
			}
		}
	}
	effectstest.HobAnswerTargets(t, g, me.ID, aBear, bOx)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{aBear, bOx} {
		if effectstest.ControllerOf(t, g, id) != me.ID {
			t.Errorf("a stolen creature is not under my control")
		}
	}
	if effectstest.ControllerOf(t, g, aKing) != a.ID {
		t.Error("the legendary creature changed hands")
	}
	effectstest.B15Destroy(t, g, saga)
	if effectstest.ControllerOf(t, g, aBear) != a.ID || effectstest.ControllerOf(t, g, bOx) != b.ID {
		t.Error("the creatures did not go home when the Saga left")
	}
}

func TestGap197TheHorusHeresyDrawsForCreaturesYouControlButDontOwnThenEachPlayerChoosesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	aBear := effectstest.B12Creature(g, a.ID, "A Bear", "Creature — Bear", 2, 2)
	bOx := effectstest.B12Creature(g, b.ID, "B Ox", "Creature — Ox", 3, 3)
	cWolf := effectstest.B12Creature(g, c.ID, "C Wolf", "Creature — Wolf", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Cat", "Creature — Cat", 1, 1)
	effectstest.CastCatalogSpell(t, g, "The Horus Heresy", gap197SagaType, gap197HorusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.HobAnswerTargets(t, g, me.ID, aBear, bOx)
	effectstest.PassPriorityAroundTable(t, g)

	hand := me.Hand.Size()
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2+1 { // two creatures I don't own, plus the draw step
		t.Errorf("chapter II drew %d, want 2 plus the draw step", got)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	// Starting with me, each player chooses a creature: I and the first opponent choose the same one.
	picks := []struct {
		p  *game.Player
		id uuid.UUID
	}{{me, cWolf}, {a, cWolf}, {b, mine}, {c, aBear}}
	for _, step := range picks {
		ch := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, step.p.ID)
		if ch == nil {
			t.Fatalf("no prompt for seat %s; pending: %d", step.p.Name, len(g.PendingChoices))
		}
		if len(ch.ChooseCards) != 4 {
			t.Errorf("seat %s was offered %d creatures, want all 4", step.p.Name, len(ch.ChooseCards))
		}
		if err := g.ResolveChooseCards(ch.ID, step.p.ID, []uuid.UUID{step.id}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"C Wolf": cWolf, "My Cat": mine, "A Bear": aBear} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was chosen and should be destroyed", name)
		}
	}
	if !g.Battlefield.Contains(bOx) {
		t.Error("B Ox was not chosen and should remain")
	}
}
