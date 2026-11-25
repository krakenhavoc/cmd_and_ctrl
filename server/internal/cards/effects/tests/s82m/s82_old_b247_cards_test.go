package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 Amonkhet batch 247 (#5535): the embalm cards and the instants and
// sorceries.

const (
	akh247Priest     = "4e0b9273-6614-41a8-9738-959349c0717b"
	akh247Guardian   = "867ee3e7-88df-4e4b-b199-f441a8460573"
	akh247Guide      = "bb813ffa-2615-4177-a7fb-6e0dfb819fc2"
	akh247Keeper     = "d22b7d4a-0ac7-4221-80fd-a47edad03b84"
	akh247Temmet     = "e245d5c9-ae3a-4246-bd74-c6fed76b9fad"
	akh247Benefact   = "02da2cbf-50d6-4c37-8a9f-40db1f7954be"
	akh247Mark       = "143d3fcf-2152-4911-b95a-d23b91662f33"
	akh247Dispossess = "9c2bbab4-97a9-4aad-96ea-31d7ef9e28fe"
	akh247Shadow     = "f33d1132-064d-42f1-8814-95d27fde8bf8"
	akh247Bolt       = "4457ed35-7c10-48c8-9776-456485fdf070"
	akh247Beast      = "f05d5632-0a82-45bf-b66e-22cf4080ed5e" // Shadowbeast Sighting: makes a 4/4 Beast token
)

func TestAkh247CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		akh247Priest: "Anointer Priest", akh247Guardian: "Labyrinth Guardian", akh247Guide: "Aven Wind Guide",
		akh247Keeper: "Glyph Keeper", akh247Temmet: "Temmet, Vizier of Naktamun", akh247Benefact: "Benefaction of Rhonas",
		akh247Mark: "Hapatra's Mark", akh247Dispossess: "Dispossess", akh247Shadow: "Shadow of the Grave",
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

// akh247Embalm embalms the graveyard card `id` (activation index `index`) and
// returns the token it made.
func akh247Embalm(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle, mana string) game.Card {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogGraveyardCard(me, name, typeLine, oracle, 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("embalm %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Errorf("the embalmed %s is not in exile", name)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.IsToken() {
			if !strings.Contains(c.TypeLine, "Zombie") {
				t.Errorf("token type line %q is not a Zombie", c.TypeLine)
			}
			return c
		}
	}
	t.Fatalf("no embalm token named %s", name)
	return game.Card{}
}

func TestAkh247AnointerPriestGainsLifeWhenACreatureTokenEntersUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Anointer Priest", "Creature — Human Cleric", akh247Priest, false)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Shadowbeast Sighting", "Sorcery", akh247Beast, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("life %d after a Beast token entered, want %d", me.Life, life+1)
	}
	// A nontoken creature, and a token under an opponent's control, are not "you control".
	effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	effectstest.PushToken(g, opp.ID, game.Card{Name: "Goat", TypeLine: "Token Creature — Goat", Power: 0, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d after a nontoken and an opposing token, want %d", me.Life, life+1)
	}
}

func TestAkh247AnointerPriestEmbalmsIntoAWhiteZombieThatGainsLifeOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	token := akh247Embalm(t, g, me, "Anointer Priest", "Creature — Human Cleric", akh247Priest, "{C}{C}{C}{W}")
	effectstest.PassPriorityAroundTable(t, g)
	if len(token.Colors) != 1 || token.Colors[0] != "W" {
		t.Errorf("token colours %v, want [W]", token.Colors)
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d: the token sees itself enter", me.Life, life+1)
	}
}

func TestAkh247LabyrinthGuardianIsSacrificedWhenASpellTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	guard := effectstest.PushCatalogPermanent(g, me.ID, "Labyrinth Guardian", "Creature — Illusion Warrior", akh247Guardian, false)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", akh247Bolt, effectstest.CardTarget(guard))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, guard) {
		t.Error("the Guardian is still on the battlefield after becoming the target of a spell")
	}
	if !me.Graveyard.Contains(guard) {
		t.Error("the sacrificed Guardian is not in the graveyard")
	}
}

func TestAkh247LabyrinthGuardianIgnoresASpellTargetingSomethingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	guard := effectstest.PushCatalogPermanent(g, me.ID, "Labyrinth Guardian", "Creature — Illusion Warrior", akh247Guardian, false)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", akh247Bolt, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, guard) {
		t.Error("the Guardian was sacrificed by a spell that did not target it")
	}
}

func TestAkh247LabyrinthGuardianEmbalms(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	token := akh247Embalm(t, g, g.Seats[0], "Labyrinth Guardian", "Creature — Illusion Warrior", akh247Guardian, "{C}{C}{C}{U}")
	if len(token.Colors) != 1 || token.Colors[0] != "W" {
		t.Errorf("token colours %v, want [W]", token.Colors)
	}
}

func TestAkh247AvenWindGuideGivesCreatureTokensFlyingAndVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Aven Wind Guide", "Creature — Bird Warrior", akh247Guide, false)
	mine := effectstest.PushToken(g, me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1})
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirs := effectstest.PushToken(g, opp.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1})
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasKeywordOf(g, mine, kw) {
			t.Errorf("my creature token lacks %s", kw)
		}
		if effectstest.HasKeywordOf(g, bear, kw) {
			t.Errorf("a nontoken creature has %s", kw)
		}
		if effectstest.HasKeywordOf(g, theirs, kw) {
			t.Errorf("an opponent's token has %s", kw)
		}
	}
}

func TestAkh247AvenWindGuideEmbalmsIntoATokenThatAlsoHasThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	token := akh247Embalm(t, g, me, "Aven Wind Guide", "Creature — Bird Warrior", akh247Guide, "{C}{C}{C}{C}{W}{U}")
	if !effectstest.HasKeywordOf(g, token.InstanceID, "flying") || !effectstest.HasKeywordOf(g, token.InstanceID, "vigilance") {
		t.Error("the embalmed Guide token lacks flying or vigilance")
	}
}

func TestAkh247GlyphKeeperCountersOnlyTheFirstTargetingEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Glyph Keeper", "Creature — Sphinx", akh247Keeper, false)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", akh247Bolt, effectstest.CardTarget(keeper))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, keeper) {
		t.Fatal("the first spell to target the Keeper was not countered")
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", akh247Bolt, effectstest.CardTarget(keeper))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, keeper) {
		t.Error("the second spell to target the Keeper in a turn was countered too")
	}
}

func TestAkh247TemmetPumpsAndUnblockablesATokenAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Temmet, Vizier of Naktamun", "Legendary Creature — Human Cleric", akh247Temmet, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	token := effectstest.PushToken(g, me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1})
	p0, t0 := effectstest.EffectivePTOf(t, g, token)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for Temmet's trigger")
	}
	if !effectstest.HasID(pick.PickTargetCards, token) || effectstest.HasID(pick.PickTargetCards, bear) {
		t.Errorf("offered %v: want the creature token only", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, token); p != p0+1 || tough != t0+1 {
		t.Errorf("token is %d/%d, want %d/%d", p, tough, p0+1, t0+1)
	}
	if effectstest.RestrictionsOf(t, g, token)&game.CantBeBlocked == 0 {
		t.Error("the token can still be blocked")
	}
	if effectstest.RestrictionsOf(t, g, bear)&game.CantBeBlocked != 0 {
		t.Error("the nontoken creature became unblockable")
	}
}

func TestAkh247BenefactionOfRhonasTakesACreatureAndAnEnchantmentAndMillsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Deep", TypeLine: "Land"},
		game.Card{Name: "Land B", TypeLine: "Land"},
		game.Card{Name: "Sorcery A", TypeLine: "Sorcery"},
		game.Card{Name: "Aura", TypeLine: "Enchantment — Aura"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	// Top of the library is the last card pushed: Bear, Aura, Sorcery A, Land B, Deep.
	effectstest.CastCatalogSpell(t, g, "Benefaction of Rhonas", "Sorcery", akh247Benefact, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatalf("no card choice: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[4], ids[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[4]) || !me.Hand.Contains(ids[3]) {
		t.Error("the Bear and the Aura are not both in hand")
	}
	for _, id := range ids[1:3] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %s from the top five was not put into the graveyard", id)
		}
	}
	if me.Library.Contains(ids[0]) || !me.Graveyard.Contains(ids[0]) {
		// five cards are revealed, so the fifth (the bottom one) goes to the graveyard too
		t.Error("the fifth revealed card was not put into the graveyard")
	}
}

func TestAkh247HapatrasMarkGrantsHexproofAndRemovesEveryMinusOneCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.PutCounters(t, g, bear, game.CounterMinusOne, 2)
	effectstest.CastCatalogSpell(t, g, "Hapatra's Mark", "Instant", akh247Mark, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, "hexproof") {
		t.Error("the creature did not gain hexproof")
	}
	if got := effectstest.CounterCount(g, bear, game.CounterMinusOne); got != 0 {
		t.Errorf("%d -1/-1 counters left, want 0", got)
	}
}

func TestAkh247HapatrasMarkCantTargetAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Hapatra's Mark", "Instant", akh247Mark, effectstest.CardTarget(theirs)); err == nil {
		t.Error("Hapatra's Mark targeted a creature its caster does not control")
	}
}

func TestAkh247DispossessExilesEveryArtifactWithTheChosenNameFromThreeZones(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inLib := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Sol Ring", TypeLine: "Artifact"})
	inHand := effectstest.PushCatalogHandCard(opp, "Sol Ring", "Artifact", "")
	inYard := effectstest.PushCatalogGraveyardCard(opp, "Sol Ring", "Artifact", "", 0, 0)
	other := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Mana Vault", TypeLine: "Artifact"})
	notArtifact := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Sol Ring", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	effectstest.CastCatalogSpell(t, g, "Dispossess", "Sorcery", akh247Dispossess, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatalf("no card-name prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Sol Ring"); err != nil {
		t.Fatalf("ResolveCardName: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if search := effectstest.SearchChoiceFor(g, me.ID); search != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, inLib, inHand, inYard)
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, id := range []uuid.UUID{inLib, inHand, inYard} {
		if !g.Exile.Contains(id) {
			t.Errorf("a Sol Ring (%s) was not exiled", id)
		}
	}
	if g.Exile.Contains(other) || g.Exile.Contains(notArtifact) {
		t.Error("a card that is not an artifact named Sol Ring, or another artifact, was exiled")
	}
}

func TestAkh247ShadowOfTheGraveReturnsOnlyCardsYouCycledOrDiscardedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Land"})
	cycled, _ := effectstest.CycleFromHand(t, g, "Archfiend of Ifnir", "Creature — Demon", "535e2af0-7b08-4026-b1bf-87626407dd42", "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	older := effectstest.PushCatalogGraveyardCard(me, "Old Card", "Creature — Bear", "", 2, 2)
	effectstest.S58HandCards(me, 1)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(cycled) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.CastCatalogSpell(t, g, "Shadow of the Grave", "Instant", akh247Shadow, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cycled) {
		t.Error("the cycled card was not returned")
	}
	if me.Graveyard.Contains(older) == false {
		t.Error("a card that was in the graveyard before the turn was returned")
	}
}
