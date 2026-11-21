package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set VOW (Innistrad: Crimson Vow and its Commander set), batch 12
// (#5495). Each test plays the real catalog entry through the engine.

const (
	vowB12LibraryOracle    = "e787ec12-ba0f-4a85-af6b-422423c90155"
	vowB12MillicentOracle  = "71766a05-62a5-4839-b9bf-f5dcde710251"
	vowB12PriestOracle     = "726dbbce-cd5e-4605-aa1d-ae5ec80cc8cf"
	vowB12WizardOracle     = "321e101c-a426-4d18-be0b-c26c5348b70d"
	vowB12HeronOracle      = "9c5897cd-f209-4e51-adc2-d791e5671029"
	vowB12NurturingOracle  = "3f03c834-2cb3-4c54-bee6-c64f3a5fce47"
	vowB12DrogskolOracle   = "6d174448-52f5-4439-9586-3abc5af940d7"
	vowB12RhodaOracle      = "3579e40c-6ec5-4927-8252-5a188db400ab"
	vowB12TiminOracle      = "0a836332-5905-4afa-a442-47d0bcf8acea"
	vowB12BreathOracle     = "eb887067-4466-45da-8387-4444fdc723f8"
	vowB12BoltOracle       = "4457ed35-7c10-48c8-9776-456485fdf070"
	vowB12SpiritTokenName  = "Spirit"
	vowB12SpiritTypeLine   = "Creature — Spirit"
	vowB12MillicentTypeStr = "Legendary Creature — Spirit Soldier"
)

func TestVowB12EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		vowB12LibraryOracle:   "Haunted Library",
		vowB12MillicentOracle: "Millicent, Restless Revenant",
		vowB12PriestOracle:    "Priest of the Blessed Graf",
		vowB12WizardOracle:    "Whispering Wizard",
		vowB12HeronOracle:     "Heron-Blessed Geist",
		vowB12NurturingOracle: "Nurturing Presence",
		vowB12DrogskolOracle:  "Drogskol Reinforcements",
		vowB12RhodaOracle:     "Rhoda, Geist Avenger",
		vowB12TiminOracle:     "Timin, Youthful Geist",
		vowB12BreathOracle:    "Breath of the Sleepless",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// vowB12Spirits counts the Spirit tokens `owner` controls.
func vowB12Spirits(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == vowB12SpiritTokenName && c.IsToken() {
			n++
		}
	}
	return n
}

// --- Haunted Library ------------------------------------------------------------

func TestVowB12HauntedLibraryPaysOneForASpiritWhenAnOpponentsCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Haunted Library", "Enchantment", vowB12LibraryOracle, 0, 0)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent's creature dying asks the controller to pay {1}")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if got := vowB12Spirits(g, me.ID); got != 1 {
		t.Fatalf("paying makes one Spirit, got %d", got)
	}
	effectstest.AssertKeywords(t, g, vowB12FirstSpirit(g, me.ID), "flying")

	second := effectstest.B12Creature(g, opp.ID, "Second", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if got := vowB12Spirits(g, me.ID); got != 1 {
		t.Errorf("declining makes nothing, got %d Spirits", got)
	}
}

func vowB12FirstSpirit(g *game.Game, owner uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == vowB12SpiritTokenName && c.IsToken() {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestVowB12HauntedLibraryIgnoresYourOwnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Haunted Library", "Enchantment", vowB12LibraryOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("your own creature dying does not trigger it")
	}
}

// --- Millicent, Restless Revenant -------------------------------------------------

func TestVowB12MillicentMakesASpiritWhenANontokenSpiritDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	millicent := effectstest.B27Push(g, me.ID, "Millicent, Restless Revenant", vowB12MillicentTypeStr, vowB12MillicentOracle, "{5}{W}{U}", 4, 4, "W", "U")
	effectstest.AssertKeywords(t, g, millicent, "flying")
	spirit := effectstest.B12Creature(g, me.ID, "Ghost", vowB12SpiritTypeLine, 1, 1)
	effectstest.KillCreature(t, g, me.ID, spirit)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 1 {
		t.Fatalf("a nontoken Spirit dying makes one Spirit token, got %d", got)
	}
	// A Spirit token dying does not.
	token := vowB12FirstSpirit(g, me.ID)
	effectstest.KillCreature(t, g, me.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 0 {
		t.Errorf("a Spirit token dying makes nothing, got %d", got)
	}
	// Neither does a non-Spirit, or an opponent's Spirit.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Ghost", vowB12SpiritTypeLine, 1, 1)
	effectstest.KillCreature(t, g, g.Seats[1].ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 0 {
		t.Errorf("a Bear or an opposing Spirit dying makes nothing, got %d", got)
	}
}

func TestVowB12MillicentMakesASpiritWhenSheOrANontokenSpiritConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Millicent, Restless Revenant", vowB12MillicentTypeStr, vowB12MillicentOracle, "{5}{W}{U}", 4, 4, "W", "U")
	ghost := effectstest.B12Creature(g, me.ID, "Ghost", vowB12SpiritTypeLine, 1, 1)
	effectstest.AttackWith(t, g, opp.ID, ghost)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 1 {
		t.Errorf("a nontoken Spirit dealing combat damage to a player makes one token, got %d", got)
	}
}

func TestVowB12MillicentDiesAndMakesASpiritToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	millicent := effectstest.B27Push(g, me.ID, "Millicent, Restless Revenant", vowB12MillicentTypeStr, vowB12MillicentOracle, "{5}{W}{U}", 4, 4, "W", "U")
	effectstest.KillCreature(t, g, me.ID, millicent)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 1 {
		t.Errorf("Millicent dying makes a Spirit token, got %d", got)
	}
}

func TestVowB12MillicentCostsOneLessPerSpiritYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Ghost A", vowB12SpiritTypeLine, 1, 1)
	effectstest.B12Creature(g, me.ID, "Ghost B", vowB12SpiritTypeLine, 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.ToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Millicent, Restless Revenant", TypeLine: vowB12MillicentTypeStr,
		OracleID: vowB12MillicentOracle, ManaCost: "{5}{W}{U}", Power: 4, Toughness: 4,
		Colors: []string{"W", "U"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.B06AddMana(me, "W", "U", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("two Spirits make it cost {3}{W}{U}: four mana is not enough")
	}
	effectstest.B06AddMana(me, "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("five mana pays {3}{W}{U} after two Spirits: %v", err)
	}
}

// --- Priest of the Blessed Graf ----------------------------------------------------

func TestVowB12PriestMakesASpiritPerOpponentWithMoreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.B12Push(g, me.ID, "Priest of the Blessed Graf", "Creature — Human Cleric", vowB12PriestOracle, 1, 2)
	effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.B12Permanent(g, o1.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, o1.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, o2.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, o2.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, o2.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, o3.ID, "Island", "Basic Land — Island") // equal: not more
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 2 {
		t.Errorf("two opponents control more lands: %d Spirits, want 2", got)
	}
}

func TestVowB12PriestMakesNothingWhenNobodyHasMoreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Priest of the Blessed Graf", "Creature — Human Cleric", vowB12PriestOracle, 1, 2)
	effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 0 {
		t.Errorf("no opponent has more lands: %d Spirits, want 0", got)
	}
}

// --- Whispering Wizard -------------------------------------------------------------

func TestVowB12WhisperingWizardTriggersOnlyOncePerTurnAndOnlyForNoncreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Whispering Wizard", "Creature — Human Wizard", vowB12WizardOracle, 3, 2)
	effectstest.ToMain(t, g)
	effectstest.JmpB4CastCreature(t, g, me, "Plain Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 0 {
		t.Fatalf("a creature spell does not trigger it, got %d Spirits", got)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", vowB12BoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 1 {
		t.Fatalf("the first noncreature spell makes a Spirit, got %d", got)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", vowB12BoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 1 {
		t.Errorf("the second noncreature spell of the turn does not trigger it, got %d", got)
	}
}

// --- Heron-Blessed Geist -----------------------------------------------------------

func TestVowB12HeronBlessedGeistMakesTwoSpiritsFromTheGraveyardWithAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Permanent(g, me.ID, "Some Enchantment", "Enchantment")
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Heron-Blessed Geist", "Creature — Spirit",
		vowB12HeronOracle, 3, 3, "{W}{C}{C}{C}", game.ActivateAbilityParams{})
	if !g.Exile.Contains(id) {
		t.Fatal("the cost exiles the card from the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 2 {
		t.Errorf("two Spirit tokens, got %d", got)
	}
}

func TestVowB12HeronBlessedGeistNeedsAnEnchantmentAndSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogGraveyardCard(me, "Heron-Blessed Geist", "Creature — Spirit", vowB12HeronOracle, 3, 3)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("with no enchantment the ability cannot be activated")
	}
	effectstest.B12Permanent(g, me.ID, "Some Enchantment", "Enchantment")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("outside a main phase the ability cannot be activated")
	}
}

// --- Nurturing Presence ------------------------------------------------------------

func TestVowB12NurturingPresenceMakesASpiritAndPumpsTheEnchantedCreatureOnEntries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Nurturing Presence", "Enchantment — Aura", vowB12NurturingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB12Spirits(g, me.ID); got != 1 {
		t.Fatalf("the Aura's enters trigger makes a Spirit, got %d", got)
	}
	if got := effectstest.AttachmentHostOf(t, g, vowB12AuraID(g, me.ID)).ID; got != host {
		t.Fatalf("the Aura is on %s, want the host", got)
	}
	// The Spirit token entered before the grant was resolved or after; now another creature enters.
	pre := effects.B31CurrentPowerOf(g, host)
	effectstest.ToMain(t, g)
	effectstest.JmpB4CastCreature(t, g, me, "Plain Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effects.B31CurrentPowerOf(g, host); got != pre+1 {
		t.Errorf("host power %d → %d after a creature entered, want +1", pre, got)
	}
	// An opponent's creature entering does not.
	opp := g.Seats[1]
	prePower := effects.B31CurrentPowerOf(g, host)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effects.B31CurrentPowerOf(g, host); got != prePower {
		t.Errorf("an opponent's creature changed the host's power %d → %d", prePower, got)
	}
}

func vowB12AuraID(g *game.Game, owner uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Nurturing Presence" {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// --- Drogskol Reinforcements -------------------------------------------------------

func TestVowB12DrogskolGivesOtherSpiritsMeleeAndKeepsItsOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drogskol := effectstest.B27Push(g, me.ID, "Drogskol Reinforcements", "Creature — Spirit Soldier", vowB12DrogskolOracle, "{W}{C}{C}{C}", 2, 2, "W")
	ghost := effectstest.B12Creature(g, me.ID, "Ghost", vowB12SpiritTypeLine, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ghost", vowB12SpiritTypeLine, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, drogskol, game.KeywordMelee) || !effectstest.HasEffectiveKeyword(t, g, ghost, game.KeywordMelee) {
		t.Error("the Reinforcements and your other Spirits have melee")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, game.KeywordMelee) || effectstest.HasEffectiveKeyword(t, g, theirs, game.KeywordMelee) {
		t.Error("a non-Spirit, and an opponent's Spirit, do not")
	}
}

func TestVowB12DrogskolPreventsNoncombatDamageToYourSpiritsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drogskol := effectstest.B27Push(g, me.ID, "Drogskol Reinforcements", "Creature — Spirit Soldier", vowB12DrogskolOracle, "{W}{C}{C}{C}", 2, 2, "W")
	ghost := effectstest.B12Creature(g, me.ID, "Ghost", vowB12SpiritTypeLine, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ghost", vowB12SpiritTypeLine, 1, 1)
	for _, id := range []uuid.UUID{ghost, drogskol} {
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", vowB12BoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: id}})
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(id) {
			t.Error("noncombat damage to one of your Spirits is prevented")
		}
	}
	for name, id := range map[string]uuid.UUID{"a Bear": bear, "an opposing Spirit": theirs} {
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", vowB12BoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: id}})
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(id) {
			t.Errorf("damage to %s is not prevented", name)
		}
	}
}

// --- Rhoda, Geist Avenger ----------------------------------------------------------

func TestVowB12RhodaGrowsWhenAnOpponentsCreatureIsTappedButNotYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhoda := effectstest.B27Push(g, me.ID, "Rhoda, Geist Avenger", "Legendary Creature — Human Soldier", vowB12RhodaOracle, "{W}{C}{C}{C}", 3, 3, "W")
	effectstest.AssertKeywords(t, g, rhoda, "vigilance")
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := g.TapTargetForEffect(mine); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, rhoda, game.CounterPlusOne); got != 0 {
		t.Fatalf("your own creature becoming tapped gives no counter, got %d", got)
	}
	if err := g.TapTargetForEffect(theirs); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, rhoda, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's creature becoming tapped gives one counter, got %d", got)
	}
}

func TestVowB12RhodaIgnoresAnOpponentsCreatureDeclaredAsAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhoda := effectstest.B27Push(g, me.ID, "Rhoda, Geist Avenger", "Legendary Creature — Human Soldier", vowB12RhodaOracle, "{W}{C}{C}{C}", 3, 3, "W")
	raider := effectstest.B12Creature(g, opp.ID, "Raider", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, rhoda, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature tapped as it is declared an attacker gives no counter, got %d", got)
	}
}

// --- Timin, Youthful Geist ---------------------------------------------------------

func TestVowB12TiminTapsACreatureAtTheBeginningOfEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	timin := effectstest.B27Push(g, me.ID, "Timin, Youthful Geist", "Legendary Creature — Spirit", vowB12TiminOracle, "{4}{U}", 3, 4, "U")
	effectstest.AssertKeywords(t, g, timin, "flying")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !vowB12Tapped(g, victim) {
		t.Error("the target creature was not tapped")
	}
}

func vowB12Tapped(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}

// --- Breath of the Sleepless -------------------------------------------------------

func TestVowB12BreathTapsACreatureWhenYouCastACreatureSpellOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Breath of the Sleepless", "Enchantment", vowB12BreathOracle, 0, 0)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	// On your own turn a creature spell does not trigger it.
	effectstest.ToMain(t, g)
	effectstest.JmpB4CastCreature(t, g, me, "Own Turn Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("casting on your own turn does not trigger it")
	}
	if vowB12Tapped(g, victim) {
		t.Fatal("nothing taps the victim on your own turn")
	}
	// On an opponent's turn, a Spirit spell can be cast at instant speed (flash)
	// and its cast taps up to one target creature.
	effectstest.AdvanceToMainOf(t, g, 1)
	ghost := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: ghost, Name: "Quick Ghost", TypeLine: vowB12SpiritTypeLine,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, ghost, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a Spirit spell is castable with flash on an opponent's turn: %v", err)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !vowB12Tapped(g, victim) {
		t.Error("the target creature was not tapped")
	}
}
