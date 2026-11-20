package testst

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// The enumerator never offers a refused type, and never offers nothing:
// with Goblins the only creature type on the board and Goblin the word
// being replaced, a seat still has an answer that is neither Goblin nor
// Wall.
func TestArtificialEvolutionsSecondPromptOffersNoRefusedType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gob := effectstest.TwPut(g, me.ID, "Goblin Piker", "Creature — Goblin", []string{"R"})
	effectstest.CastCatalogSpell(t, g, "Artificial Evolution", "Instant", effectstest.TwEvolutionOracle, effectstest.TwTargetRef(gob))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	var offered []string
	for _, m := range legal.EnumerateFor(g, me.ID) {
		var p struct {
			CreatureType string `json:"creature_type"`
		}
		if json.Unmarshal(m.Params, &p) == nil && p.CreatureType != "" {
			offered = append(offered, p.CreatureType)
		}
	}
	if len(offered) == 0 {
		t.Fatal("the seat is offered an answer")
	}
	if slices.Contains(offered, "Goblin") || slices.Contains(offered, "Wall") {
		t.Errorf("offered %v, which includes a refused type", offered)
	}
	effectstest.AnswerCreatureType(t, g, me.ID, offered[0])
	if subs := effectstest.EffectiveSubtypes(t, g, gob); !slices.Equal(subs, []string{offered[0]}) {
		t.Errorf("the Goblin is a %v now, want [%s]", subs, offered[0])
	}
}

func twOnBattlefield(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

// The card the issue opened with, on the classic victim: Sleight of
// Mind turns a Doom Blade on the stack from "nonblack" into "nonblue",
// and the blue creature it targeted is no longer a legal target when it
// resolves (CR 608.2b), so the spell does nothing.
func TestSleightOfMindMakesDoomBladeNonblue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	blue := effectstest.TwPut(g, opp.ID, "Blue Bear", "Creature — Bear", []string{"U"})
	blade := effectstest.CastCatalogSpell(t, g, "Doom Blade", "Instant", effectstest.TwDoomBladeOracle, effectstest.TwTargetRef(blue))
	effectstest.CastCatalogSpell(t, g, "Sleight of Mind", "Instant", effectstest.TwSleightOracle, effectstest.TwTargetRef(blade))
	// Sleight resolves first and asks for the words.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	for i := 0; i < 8 && effectstest.LatestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	effectstest.AnswerOptionPick(t, g, me.ID, effectstest.TwBlack)
	effectstest.AnswerOptionPick(t, g, me.ID, effectstest.TwBlue-0) // W, U, R, G: blue is 1
	var changes []string
	g.ReadSnapshot(func() {
		for _, c := range g.Stack.Cards {
			if c.InstanceID == blade {
				changes = c.TextChanges()
			}
		}
	})
	if !slices.Equal(changes, []string{"black → blue"}) {
		t.Fatalf("Doom Blade's changed words = %v, want [black → blue]", changes)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !twOnBattlefield(g, blue) {
		t.Fatal("Doom Blade now says nonblue: its blue target was illegal and it did nothing")
	}
}

// The same change the other way: a Doom Blade made "nonred" can still
// kill the blue creature it targeted — the clause reads the new word, not
// a stale one.
func TestDoomBladeChangedToNonredStillKillsABlueCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	blue := effectstest.TwPut(g, opp.ID, "Blue Bear", "Creature — Bear", []string{"U"})
	blade := effectstest.CastCatalogSpell(t, g, "Doom Blade", "Instant", effectstest.TwDoomBladeOracle, effectstest.TwTargetRef(blue))
	effectstest.CastCatalogSpell(t, g, "Sleight of Mind", "Instant", effectstest.TwSleightOracle, effectstest.TwTargetRef(blade))
	for i := 0; i < 8 && effectstest.LatestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	effectstest.AnswerOptionPick(t, g, me.ID, effectstest.TwBlack)
	effectstest.AnswerOptionPick(t, g, me.ID, 2) // W, U, R, G: red is 2
	effectstest.PassPriorityAroundTable(t, g)
	if twOnBattlefield(g, blue) {
		t.Fatal("a nonred Doom Blade destroys a blue creature")
	}
}

// A printed protection from a colour is text: Sleight of Mind on it
// changes the colour (CR 612.2). A protection GRANTED by something else
// is not the permanent's text and would not change (CR 612.3).
func TestSleightOfMindChangesPrintedProtection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	knight := effectstest.TwPut(g, opp.ID, "White Knight", "Creature — Human Knight", []string{"W"}, "protection from black")
	effectstest.TwChange(t, g, "Sleight of Mind", effectstest.TwSleightOracle, knight, effectstest.TwBlack, 2) // red
	if effectstest.HasEffectiveKeyword(t, g, knight, "protection from black") {
		t.Error("the knight's protection from black now reads protection from red")
	}
	if !effectstest.HasEffectiveKeyword(t, g, knight, "protection from red") {
		t.Error("the knight has protection from red")
	}
	if got := effectstest.TwTextChanges(t, g, knight); !slices.Equal(got, []string{"black → red"}) {
		t.Errorf("text_changes = %v", got)
	}
}

// A catalog card's printed keywords are its text too, though they reach
// the permanent through the catalog rather than the importer's list:
// Burrenton Forge-Tender's protection from red changes
// with the rest.
func TestSleightOfMindChangesACatalogPrintedProtection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	tender := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Burrenton Forge-Tender", TypeLine: "Creature — Kithkin Wizard",
		OracleID: "70bb275b-3458-4690-a50f-b231fbf0bccb", Colors: []string{"W"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	if !effectstest.HasEffectiveKeyword(t, g, tender, "protection from red") {
		t.Fatal("setup: the Forge-Tender has protection from red")
	}
	effectstest.TwChange(t, g, "Sleight of Mind", effectstest.TwSleightOracle, tender, effectstest.TwRed, 1) // W, U, B, G: blue is 1
	if effectstest.HasEffectiveKeyword(t, g, tender, "protection from red") {
		t.Error("its protection from red now reads protection from blue")
	}
	if !effectstest.HasEffectiveKeyword(t, g, tender, "protection from blue") {
		t.Error("the Forge-Tender has protection from blue")
	}
}

// Glamerdye and Alter Reality are Sleight of Mind with retrace and
// flashback: the same change, cast from hand here.
func TestGlamerdyeAndAlterRealityChangeAColourWord(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Glamerdye", effectstest.TwGlamerdyeOracle},
		{"Alter Reality", effectstest.TwAlterOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
			knight := effectstest.TwPut(g, opp.ID, "White Knight", "Creature — Human Knight", []string{"W"}, "protection from black")
			effectstest.TwChange(t, g, tc.name, tc.oracle, knight, effectstest.TwBlack, 1) // blue
			if !effectstest.HasEffectiveKeyword(t, g, knight, "protection from blue") {
				t.Error("protection from black became protection from blue")
			}
			spec, ok := effects.Lookup(tc.oracle)
			if !ok || len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].FromZone != game.ZoneGraveyard {
				t.Errorf("%s is castable from the graveyard", tc.name)
			}
		})
	}
}

// Magical Hack on a basic Forest: it is an Island now (the type line is
// text), taps for blue, and is still named Forest (CR 612.2).
func TestMagicalHackTurnsAForestIntoAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID,
	})
	// Plains, Island, Swamp, Mountain, Forest; then the four others.
	effectstest.TwChange(t, g, "Magical Hack", effectstest.TwHackOracle, forest, 4, 1) // Forest → Island
	subs := effectstest.EffectiveSubtypes(t, g, forest)
	if !slices.Equal(subs, []string{"Island"}) {
		t.Fatalf("the Forest's subtypes = %v, want [Island]", subs)
	}
	c := effectstest.GaLayered(t, g, forest)
	if c.Name != "Forest" {
		t.Errorf("the name is not text a text change touches: %q", c.Name)
	}
	if !c.HasSupertype("basic") {
		t.Error("the land is still basic")
	}
	var produced []string
	for _, ab := range game.ManaAbilitiesForCard(c) {
		produced = append(produced, ab.Produced)
	}
	if !slices.Equal(produced, []string{"{U}"}) {
		t.Errorf("the land taps for %v, want [{U}] (CR 305.6)", produced)
	}
}

// Landwalk is text: Magical Hack changes "swampwalk" to "islandwalk".
func TestMagicalHackChangesLandwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	wraith := effectstest.TwPut(g, opp.ID, "Bog Wraith", "Creature — Wraith", []string{"B"}, "swampwalk")
	effectstest.TwChange(t, g, "Magical Hack", effectstest.TwHackOracle, wraith, 2, 1) // Swamp → Island
	if effectstest.HasEffectiveKeyword(t, g, wraith, "swampwalk") || !effectstest.HasEffectiveKeyword(t, g, wraith, "islandwalk") {
		t.Error("swampwalk became islandwalk")
	}
}

// Magical Hack on a lord: the land type in the keyword it GRANTS is a
// word of the lord's text too. Goblin King's mountainwalk grant becomes
// an islandwalk grant.
func TestMagicalHackChangesAGrantedLandwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	king := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin King", TypeLine: "Creature — Goblin",
		OracleID: effectstest.TwGoblinKingOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	gob := effectstest.TwPut(g, me.ID, "Goblin Piker", "Creature — Goblin", []string{"R"})
	if !effectstest.HasEffectiveKeyword(t, g, gob, "mountainwalk") {
		t.Fatal("setup: Goblin King grants mountainwalk")
	}
	effectstest.TwChange(t, g, "Magical Hack", effectstest.TwHackOracle, king, 3, 1) // Mountain → Island
	if effectstest.HasEffectiveKeyword(t, g, gob, "mountainwalk") || !effectstest.HasEffectiveKeyword(t, g, gob, "islandwalk") {
		t.Error("Goblin King's grant now reads islandwalk")
	}
}

// Mind Bend asks for ONE word from either class, then offers only the
// other words of the class picked.
func TestMindBendOffersEitherClassThenTheSameClass(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wraith := effectstest.TwPut(g, me.ID, "Bog Wraith", "Creature — Wraith", []string{"B"}, "swampwalk")
	effectstest.CastCatalogSpell(t, g, "Mind Bend", "Instant", effectstest.TwMindBendOracle, effectstest.TwTargetRef(wraith))
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.LatestOptionPickFor(g, me.ID)
	if first == nil || len(first.PickOptions) != 10 {
		t.Fatalf("the first prompt offers the five colours and the five land types: %+v", first)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 7) // Swamp
	second := effectstest.LatestOptionPickFor(g, me.ID)
	var labels []string
	for _, o := range second.PickOptions {
		labels = append(labels, o.Label)
	}
	if !slices.Equal(labels, []string{"Plains", "Island", "Mountain", "Forest"}) {
		t.Fatalf("the second prompt offers the other land types: %v", labels)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 3) // Forest
	if !effectstest.HasEffectiveKeyword(t, g, wraith, "forestwalk") {
		t.Error("swampwalk became forestwalk")
	}
}

// Crystal Spray changes the word until end of turn, then draws.
func TestCrystalSprayLastsTheTurnAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := effectstest.TwPut(g, me.ID, "White Knight", "Creature — Human Knight", []string{"W"}, "protection from black")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Crystal Spray", "Instant", effectstest.TwSprayOracle, effectstest.TwTargetRef(knight))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("the card is drawn after the words are chosen, not before (hand %d → %d)", hand, me.Hand.Size())
	}
	effectstest.AnswerOptionPick(t, g, me.ID, effectstest.TwBlack)
	effectstest.AnswerOptionPick(t, g, me.ID, effectstest.TwGreen-1) // W, U, R, G: green is 3
	if me.Hand.Size() != hand+1 {
		t.Errorf("Crystal Spray draws a card (hand %d → %d)", hand, me.Hand.Size())
	}
	if !effectstest.HasEffectiveKeyword(t, g, knight, "protection from green") {
		t.Fatal("protection from black reads protection from green this turn")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, knight, "protection from black") {
		t.Error("the change ended at cleanup")
	}
}

// Whim of Volrath is Mind Bend until end of turn, with buyback.
func TestWhimOfVolrathLastsTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wraith := effectstest.TwPut(g, me.ID, "Bog Wraith", "Creature — Wraith", []string{"B"}, "swampwalk")
	effectstest.TwChange(t, g, "Whim of Volrath", effectstest.TwWhimOracle, wraith, 7, 0) // Swamp → Plains
	if !effectstest.HasEffectiveKeyword(t, g, wraith, "plainswalk") {
		t.Fatal("swampwalk reads plainswalk this turn")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, wraith, "swampwalk") {
		t.Error("the change ended at cleanup")
	}
	spec, _ := effects.Lookup(effectstest.TwWhimOracle)
	if len(spec.OptionalCosts) != 1 {
		t.Error("Whim of Volrath has buyback")
	}
}

// Artificial Evolution on a lord changes the tribe it pumps, and on a
// creature its type line; the new type can't be Wall.
func TestArtificialEvolutionRewritesALordsTribe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	king := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin King", TypeLine: "Creature — Goblin",
		OracleID: effectstest.TwGoblinKingOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	gob := effectstest.TwPut(g, me.ID, "Goblin Piker", "Creature — Goblin", []string{"R"})
	elf := effectstest.TwPut(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", []string{"G"})
	effectstest.CastCatalogSpell(t, g, "Artificial Evolution", "Instant", effectstest.TwEvolutionOracle, effectstest.TwTargetRef(king))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	prompt := effectstest.PendingOfKind(g, game.PendingChoiceCreatureType)
	if prompt == nil || !slices.Contains(prompt.ExcludedTypes, "Wall") || !slices.Contains(prompt.ExcludedTypes, "Goblin") {
		t.Fatalf("the new type can't be Wall or the old one: %+v", prompt)
	}
	if err := g.ResolveCreatureTypeChoice(prompt.ID, me.ID, "Wall"); err == nil {
		t.Fatal("Wall is refused as the new type")
	}
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	if got := effectstest.EffectivePower(t, g, elf); got != 3 {
		t.Errorf("Goblin King is a lord of Elves now: the Elf's power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, gob); got != 2 {
		t.Errorf("and no longer of Goblins: the Goblin's power = %d, want 2", got)
	}
	if subs := effectstest.EffectiveSubtypes(t, g, king); !slices.Equal(subs, []string{"Elf"}) {
		t.Errorf("Goblin King's type line is text too: %v, want [Elf]", subs)
	}
}

// A change made to a permanent SPELL stays with the permanent it becomes
// (CR 400.7a; "the permanent will retain the text change").
func TestATextChangeOnAPermanentSpellCarriesOntoThePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := effectstest.TwPut(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", []string{"G"})
	king := effectstest.CastCatalogSpell(t, g, "Goblin King", "Creature — Goblin", effectstest.TwGoblinKingOracle, nil)
	effectstest.CastCatalogSpell(t, g, "Artificial Evolution", "Instant", effectstest.TwEvolutionOracle, effectstest.TwTargetRef(king))
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceCreatureType) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	effectstest.PassPriorityAroundTable(t, g)
	if !twOnBattlefield(g, king) {
		t.Fatal("Goblin King resolved")
	}
	if got := effectstest.EffectivePower(t, g, elf); got != 3 {
		t.Errorf("the permanent kept the spell's text change: the Elf's power = %d, want 3", got)
	}
}
