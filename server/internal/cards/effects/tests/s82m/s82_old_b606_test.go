package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b606_test.go — Mirrodin, batch 606 (#5606): each test plays the real
// catalog entry through the real cast, activation and trigger paths.

const (
	oldB606OgreOracle   = "17e00d6e-c57d-48ea-a1a1-0c1097083851"
	oldB606SpyOracle    = "d78cb7d8-9cfc-4d16-86ea-d49fc5d65db0"
	oldB606SlugOracle   = "d7c8b636-b7ff-446d-aa51-516e2cfe2730"
	oldB606ShamanOracle = "198d1792-d638-4760-a476-7ec3af495610"
	oldB606ProbeOracle  = "1beb81a8-b4aa-4964-93d2-2b5414fc6ce3"
	oldB606SphereOracle = "790d3ee2-8b93-464c-9b16-afe62db0048c"
	oldB606RustOracle   = "21af6b08-1dd1-4e7f-be7b-cc8e70c578c7"
	oldB606PeaceOracle  = "89390fc2-17af-4754-bb1c-7af905d4a2a8"
	oldB606ElderOracle  = "f85a1752-c6fa-481b-997d-6b5dd4c6b54b"
	oldB606GolemOracle  = "3460e86f-c182-4ca9-bd90-7949b0759c98"
	oldB606HelixOracle  = "9aa8eef1-fc67-4d54-9783-4d0175a76741"
	oldB606PrisonOracle = "da1ea03f-747f-46c9-8ff5-7e85996df804"
)

func TestOldB606CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB606OgreOracle: "Ogre Leadfoot", oldB606SpyOracle: "Neurok Spy", oldB606SlugOracle: "Molder Slug",
		oldB606ShamanOracle: "Krark-Clan Shaman", oldB606ProbeOracle: "Psychogenic Probe",
		oldB606SphereOracle: "Sphere of Purity", oldB606RustOracle: "Rust Elemental",
		oldB606PeaceOracle: "Loxodon Peacekeeper", oldB606ElderOracle: "Leonin Elder",
		oldB606GolemOracle: "Mirror Golem", oldB606HelixOracle: "Spellweaver Helix",
		oldB606PrisonOracle: "Thought Prison",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// --- Ogre Leadfoot ---

func TestOldB606OgreLeadfootDestroysAnArtifactCreatureThatBlocksIt(t *testing.T) {
	for _, artifact := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ogre := effectstest.B12Push(g, me.ID, "Ogre Leadfoot", "Creature — Ogre", oldB606OgreOracle, 3, 3)
		typeLine := "Creature — Wall"
		if artifact {
			typeLine = "Artifact Creature — Golem"
		}
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", typeLine, 0, 9)
		effectstest.AdvanceToMain(t, g)
		effectstest.Pr7bAttack(t, g, opp.ID, ogre)
		effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: ogre})
		if got := g.Battlefield.Contains(blocker); got == artifact {
			t.Errorf("artifact blocker=%v: still on the battlefield=%v", artifact, got)
		}
	}
}

// --- Neurok Spy ---

func TestOldB606NeurokSpyIsUnblockableWhileTheDefenderControlsAnArtifact(t *testing.T) {
	for _, artifact := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		spy := effectstest.B12Push(g, me.ID, "Neurok Spy", "Creature — Human Rogue", oldB606SpyOracle, 2, 2)
		wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
		if artifact {
			effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		} else {
			effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(spy, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(wall, spy)
		if artifact && err == nil {
			t.Error("the Spy was blocked although the defender controls an artifact")
		}
		if !artifact && err != nil {
			t.Errorf("the Spy could not be blocked without a defending artifact: %v", err)
		}
	}
}

// --- Molder Slug ---

func TestOldB606MolderSlugMakesEachPlayerSacrificeAnArtifactInTheirUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Molder Slug", "Creature — Slug Beast", oldB606SlugOracle, 4, 6)
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	mineToo := effectstest.B12Permanent(g, me.ID, "My Other Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent kept their only artifact through their own upkeep")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(mineToo) {
		t.Error("my artifacts were sacrificed in the opponent's upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, mineToo)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mineToo) || !g.Battlefield.Contains(mine) {
		t.Error("the artifact I chose was not the one sacrificed")
	}
}

// --- Krark-Clan Shaman ---

func TestOldB606KrarkClanShamanDamagesEveryCreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := effectstest.B12Push(g, me.ID, "Krark-Clan Shaman", "Creature — Goblin Shaman", oldB606ShamanOracle, 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	bird := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, shaman, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if g.Battlefield.Contains(shaman) {
		t.Error("the 1/1 Shaman survived its own damage")
	}
	if !g.Battlefield.Contains(ground) {
		t.Error("a 2/2 died to 1 damage")
	}
	if !g.Battlefield.Contains(bird) {
		t.Error("a creature with flying was damaged to death")
	}
	if got := effectstest.DamageMarkedOn(g, ground); got != 1 {
		t.Errorf("the creature without flying has %d damage, want 1", got)
	}
}

// --- Psychogenic Probe ---

func TestOldB606PsychogenicProbeHitsAnyPlayerWhoseLibraryIsShuffledByAnEffect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, me.ID, "Probe", "Artifact")
	effectstest.PushCatalogPermanent(g, me.ID, "Psychogenic Probe", "Artifact", oldB606ProbeOracle, false)
	before, myBefore := opp.Life, me.Life
	g.WithWriteLock(func() { _ = g.ShuffleLibraryForEffect(opp.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("the player whose library was shuffled has %d life, want %d", opp.Life, before-2)
	}
	if me.Life != myBefore {
		t.Errorf("the Probe's controller lost life: %d", me.Life)
	}
	g.WithWriteLock(func() { _ = g.ShuffleLibraryForEffect(me.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myBefore-2 {
		t.Errorf("a controller's own shuffle left them at %d, want %d", me.Life, myBefore-2)
	}
}

// --- Sphere of Purity ---

func TestOldB606SphereOfPurityPreventsOneDamageFromAnArtifactOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sphere of Purity", "Enchantment", oldB606SphereOracle, false)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock Golem", "Artifact Creature — Golem")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	before := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(rock, me.ID, 3) })
	if me.Life != before-2 {
		t.Errorf("an artifact's 3 damage left me at %d, want %d", me.Life, before-2)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bear, me.ID, 3) })
	if me.Life != before-5 {
		t.Errorf("a non-artifact's 3 damage left me at %d, want %d", me.Life, before-5)
	}
	opponentLife := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(rock, opp.ID, 3) })
	if opp.Life != opponentLife-3 {
		t.Errorf("the Sphere protected another player: %d, want %d", opp.Life, opponentLife-3)
	}
}

// --- Rust Elemental ---

func TestOldB606RustElementalSacrificesAnotherArtifactOrTapsAndCostsFourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elem := effectstest.B12Push(g, me.ID, "Rust Elemental", "Artifact Creature — Elemental", oldB606RustOracle, 4, 4)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	before := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Fatal("the other artifact was not sacrificed")
	}
	if me.Life != before || effectstest.IsTapped(g, elem) {
		t.Fatalf("with an artifact to sacrifice, life %d (want %d), tapped=%v", me.Life, before, effectstest.IsTapped(g, elem))
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-4 {
		t.Errorf("life %d, want %d once there is nothing to sacrifice", me.Life, before-4)
	}
	if !g.Battlefield.Contains(elem) || !effectstest.IsTapped(g, elem) {
		t.Error("the Elemental did not tap")
	}
}

// --- Loxodon Peacekeeper ---

func TestOldB606LoxodonPeacekeeperGoesToTheLowestLifeTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keeper := effectstest.B12Push(g, me.ID, "Loxodon Peacekeeper", "Creature — Elephant Soldier", oldB606PeaceOracle, 4, 4)
	g.WithWriteLock(func() { opp.Life = 3 })
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, keeper); got != opp.ID {
		t.Errorf("controller %s, want the player with the lowest life total", got)
	}
}

func TestOldB606LoxodonPeacekeeperLetsItsControllerBreakATie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	keeper := effectstest.B12Push(g, me.ID, "Loxodon Peacekeeper", "Creature — Elephant Soldier", oldB606PeaceOracle, 4, 4)
	g.WithWriteLock(func() { a.Life, b.Life = 5, 5 })
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, keeper); got != b.ID {
		t.Errorf("controller %s, want the tied player I chose", got)
	}
}

// --- Leonin Elder ---

func TestOldB606LeoninElderGainsLifeForAnyArtifactEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Leonin Elder", "Creature — Cat Cleric", oldB606ElderOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	effectstest.CastFromHandForTest(t, g, me, "My Rock", "Artifact", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Fatalf("life %d, want %d after my artifact entered", me.Life, before+1)
	}
	// An opponent's artifact counts too, and a nonartifact permanent does not.
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, opp, "Their Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a creature entering triggered the Elder")
	}
	effectstest.CastFromHandForTest(t, g, opp, "Their Rock", "Artifact", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d, want %d: the optional gain was declined", me.Life, before+1)
	}
}

// --- Mirror Golem ---

func TestOldB606MirrorGolemGainsProtectionFromTheExiledCardsTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	exiled := effectstest.GraveyardCardOf(g, me, "Relic Beast", "Artifact Creature — Beast", "{3}", "")
	other := effectstest.GraveyardCardOf(g, me, "Bolt", "Instant", "{R}", "")
	golem := effectstest.CastHoldCreature(t, g, "Mirror Golem", "Artifact Creature — Golem", oldB606GolemOracle, 3, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, exiled)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(exiled); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the card was not exiled: %v", z)
	}
	for _, kw := range []string{"protection from artifacts", "protection from creatures"} {
		if !effectstest.HasEffectiveKeyword(t, g, golem, kw) {
			t.Errorf("Mirror Golem lacks %q", kw)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, golem, "protection from instants") {
		t.Error("protection from a type of a card that was not exiled")
	}
	if !me.Graveyard.Contains(other) {
		t.Error("the other graveyard card was exiled")
	}
}

// --- Thought Prison ---

func TestOldB606ThoughtPrisonExilesANonlandCardAndPunishesSharedColorOrValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	forest := uuid.New()
	exiled := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: exiled, Name: "Red Three", TypeLine: "Creature — Bear", ManaCost: "{2}{R}",
		Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Thought Prison", "Artifact", oldB606PrisonOracle, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	var c *game.PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceRevealedHandPick {
			c = pc
		}
	}
	if c == nil {
		t.Fatalf("no revealed-hand pick is open: %+v", g.PendingChoices)
	}
	for _, id := range c.DiscardOptions {
		if id == forest {
			t.Error("a land was offered")
		}
	}
	if err := g.ResolvePendingChoice(c.ID, me.ID, []uuid.UUID{exiled}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(exiled); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the chosen card is in %v, want exile", z)
	}
	cast := func(name, typeLine, cost string, colors []string) int {
		before := opp.Life
		id := uuid.New()
		g.WithWriteLock(func() {
			opp.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors, Owner: opp.ID, Controller: opp.ID})
		})
		if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		return before - opp.Life
	}
	if got := cast("Red One", "Instant", "{R}", []string{"R"}); got != 2 {
		t.Errorf("a red spell cost %d life, want 2", got)
	}
	if got := cast("Blue Three", "Instant", "{2}{U}", []string{"U"}); got != 2 {
		t.Errorf("a mana value 3 spell cost %d life, want 2", got)
	}
	if got := cast("Blue Two", "Instant", "{1}{U}", []string{"U"}); got != 0 {
		t.Errorf("a spell sharing neither cost %d life, want 0", got)
	}
}

// --- Spellweaver Helix ---

func oldB606Helix(t *testing.T) (g *game.Game, me *game.Player, helix, twin, other uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	twin = effectstest.GraveyardCardOf(g, me, "Twin Spell", "Sorcery", "{1}{R}", "")
	other = effectstest.GraveyardCardOf(g, me, "Other Spell", "Sorcery", "{2}{G}", "")
	effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.AdvanceToMain(t, g)
	helix = effectstest.CastCatalogSpell(t, g, "Spellweaver Helix", "Artifact", oldB606HelixOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return
}

func TestOldB606SpellweaverHelixImprintsTwoSorceriesFromOneGraveyard(t *testing.T) {
	g, me, _, twin, other := oldB606Helix(t)
	oldB606PickTwo(t, g, me, twin, other)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{twin, other} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a sorcery card is in %v, want exile", z)
		}
	}
}

func TestOldB606SpellweaverHelixCopiesTheOtherCardWhenOneIsCast(t *testing.T) {
	g, me, _, twin, other := oldB606Helix(t)
	oldB606PickTwo(t, g, me, twin, other)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// An unrelated spell does nothing.
	effectstest.CastFromHandForTest(t, g, me, "Unrelated", "Sorcery", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a spell with a different name triggered the Helix")
	}
	// A card named like one of the exiled cards does.
	effectstest.CastFromHandForTest(t, g, me, "Twin Spell", "Sorcery", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	var copyID uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.CardCopy {
				copyID = c.InstanceID
				if c.Name != "Other Spell" {
					t.Errorf("the copy is of %q, want the other exiled card", c.Name)
				}
			}
		}
	})
	if copyID == uuid.Nil {
		t.Fatal("no copy of the other exiled card was made")
	}
	if err := g.CastSpell(me.ID, copyID, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("the copy could not be cast without paying its mana cost: %v", err)
	}
}

// oldB606PickTwo answers a two-target pick with both cards.
func oldB606PickTwo(t *testing.T, g *game.Game, me *game.Player, a, b uuid.UUID) {
	t.Helper()
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}
