package effects

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// The enumerator never offers a refused type, and never offers nothing:
// with Goblins the only creature type on the board and Goblin the word
// being replaced, a seat still has an answer that is neither Goblin nor
// Wall.
func TestArtificialEvolutionsSecondPromptOffersNoRefusedType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gob := twPut(g, me.ID, "Goblin Piker", "Creature — Goblin", []string{"R"})
	castCatalogSpell(t, g, "Artificial Evolution", "Instant", twEvolutionOracle, twTargetRef(gob))
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Goblin")
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
	answerCreatureType(t, g, me.ID, offered[0])
	if subs := effectiveSubtypes(t, g, gob); !slices.Equal(subs, []string{offered[0]}) {
		t.Errorf("the Goblin is a %v now, want [%s]", subs, offered[0])
	}
}

// text_words_test.go — changing one word of an object's text (CR 612.2,
// ADR 0160's 2026-10-15 amendment, #2779): Sleight of Mind, Glamerdye,
// Alter Reality, Magical Hack, Mind Bend, Crystal Spray, Whim of Volrath
// and Artificial Evolution, and the readers that consult the change.

const (
	twSleightOracle    = "99dba614-40d3-41c1-a3b2-edc8777b010f"
	twGlamerdyeOracle  = "570a5e60-ed0f-4c14-8cc3-af819648030b"
	twAlterOracle      = "054c00b4-c940-455a-ad88-b8aa35931ecf"
	twHackOracle       = "cba229fa-9035-405b-b091-3798898a37ee"
	twMindBendOracle   = "d8377869-f8f7-4b4b-93ad-3ad3c4826b1a"
	twSprayOracle      = "fe371e3a-976f-41d8-ad58-4b3b61415d71"
	twWhimOracle       = "51a20c9f-68d2-471b-8622-67d481b53b92"
	twEvolutionOracle  = "c5ab6dc9-34af-434f-ae8e-b2389ee6c5ae"
	twDoomBladeOracle  = "59e7f2ae-4535-4191-98be-3e65b6b2befa"
	twGoblinKingOracle = "d236b3fc-0d3f-4d99-875d-e32a33fe5767"
)

// The option indices the two prompts offer: colours WUBRG, then the
// basic land types in CR 305.6 order. The second prompt leaves the
// first answer out.
const (
	twWhite, twBlue, twBlack, twRed, twGreen = 0, 1, 2, 3, 4
)

// twPut puts a vanilla permanent onto the battlefield.
func twPut(g *game.Game, owner uuid.UUID, name, typeLine string, colors []string, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Colors: colors, Keywords: keywords,
		Owner: owner, Controller: owner,
	})
}

func twTargetRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// twChange casts a text-changing spell on `target`, resolves it and
// answers its two prompts with the given indices.
func twChange(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID, from, to int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, name, "Instant", oracle, twTargetRef(target))
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, from)
	answerOptionPick(t, g, me.ID, to)
}

func twTextChanges(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	c := gaLayered(t, g, id)
	return c.TextChanges()
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
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	blue := twPut(g, opp.ID, "Blue Bear", "Creature — Bear", []string{"U"})
	blade := castCatalogSpell(t, g, "Doom Blade", "Instant", twDoomBladeOracle, twTargetRef(blue))
	castCatalogSpell(t, g, "Sleight of Mind", "Instant", twSleightOracle, twTargetRef(blade))
	// Sleight resolves first and asks for the words.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	for i := 0; i < 8 && latestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	answerOptionPick(t, g, me.ID, twBlack)
	answerOptionPick(t, g, me.ID, twBlue-0) // W, U, R, G: blue is 1
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
	passPriorityAroundTable(t, g)
	if !twOnBattlefield(g, blue) {
		t.Fatal("Doom Blade now says nonblue: its blue target was illegal and it did nothing")
	}
}

// The same change the other way: a Doom Blade made "nonred" can still
// kill the blue creature it targeted — the clause reads the new word, not
// a stale one.
func TestDoomBladeChangedToNonredStillKillsABlueCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	blue := twPut(g, opp.ID, "Blue Bear", "Creature — Bear", []string{"U"})
	blade := castCatalogSpell(t, g, "Doom Blade", "Instant", twDoomBladeOracle, twTargetRef(blue))
	castCatalogSpell(t, g, "Sleight of Mind", "Instant", twSleightOracle, twTargetRef(blade))
	for i := 0; i < 8 && latestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	answerOptionPick(t, g, me.ID, twBlack)
	answerOptionPick(t, g, me.ID, 2) // W, U, R, G: red is 2
	passPriorityAroundTable(t, g)
	if twOnBattlefield(g, blue) {
		t.Fatal("a nonred Doom Blade destroys a blue creature")
	}
}

// A printed protection from a colour is text: Sleight of Mind on it
// changes the colour (CR 612.2). A protection GRANTED by something else
// is not the permanent's text and would not change (CR 612.3).
func TestSleightOfMindChangesPrintedProtection(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	knight := twPut(g, opp.ID, "White Knight", "Creature — Human Knight", []string{"W"}, "protection from black")
	twChange(t, g, "Sleight of Mind", twSleightOracle, knight, twBlack, 2) // red
	if hasEffectiveKeyword(t, g, knight, "protection from black") {
		t.Error("the knight's protection from black now reads protection from red")
	}
	if !hasEffectiveKeyword(t, g, knight, "protection from red") {
		t.Error("the knight has protection from red")
	}
	if got := twTextChanges(t, g, knight); !slices.Equal(got, []string{"black → red"}) {
		t.Errorf("text_changes = %v", got)
	}
}

// A catalog card's printed keywords are its text too, though they reach
// the permanent through the layer-6 printed-keyword row rather than the
// importer's list: Burrenton Forge-Tender's protection from red changes
// with the rest.
func TestSleightOfMindChangesACatalogPrintedProtection(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	tender := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Burrenton Forge-Tender", TypeLine: "Creature — Kithkin Wizard",
		OracleID: "70bb275b-3458-4690-a50f-b231fbf0bccb", Colors: []string{"W"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	if !hasEffectiveKeyword(t, g, tender, "protection from red") {
		t.Fatal("setup: the Forge-Tender has protection from red")
	}
	twChange(t, g, "Sleight of Mind", twSleightOracle, tender, twRed, 1) // W, U, B, G: blue is 1
	if hasEffectiveKeyword(t, g, tender, "protection from red") {
		t.Error("its protection from red now reads protection from blue")
	}
	if !hasEffectiveKeyword(t, g, tender, "protection from blue") {
		t.Error("the Forge-Tender has protection from blue")
	}
}

// Glamerdye and Alter Reality are Sleight of Mind with retrace and
// flashback: the same change, cast from hand here.
func TestGlamerdyeAndAlterRealityChangeAColourWord(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Glamerdye", twGlamerdyeOracle},
		{"Alter Reality", twAlterOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
			knight := twPut(g, opp.ID, "White Knight", "Creature — Human Knight", []string{"W"}, "protection from black")
			twChange(t, g, tc.name, tc.oracle, knight, twBlack, 1) // blue
			if !hasEffectiveKeyword(t, g, knight, "protection from blue") {
				t.Error("protection from black became protection from blue")
			}
			spec, ok := Lookup(tc.oracle)
			if !ok || len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].FromZone != game.ZoneGraveyard {
				t.Errorf("%s is castable from the graveyard", tc.name)
			}
		})
	}
}

// Magical Hack on a basic Forest: it is an Island now (the type line is
// text), taps for blue, and is still named Forest (CR 612.2).
func TestMagicalHackTurnsAForestIntoAnIsland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID,
	})
	// Plains, Island, Swamp, Mountain, Forest; then the four others.
	twChange(t, g, "Magical Hack", twHackOracle, forest, 4, 1) // Forest → Island
	subs := effectiveSubtypes(t, g, forest)
	if !slices.Equal(subs, []string{"Island"}) {
		t.Fatalf("the Forest's subtypes = %v, want [Island]", subs)
	}
	c := gaLayered(t, g, forest)
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
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	wraith := twPut(g, opp.ID, "Bog Wraith", "Creature — Wraith", []string{"B"}, "swampwalk")
	twChange(t, g, "Magical Hack", twHackOracle, wraith, 2, 1) // Swamp → Island
	if hasEffectiveKeyword(t, g, wraith, "swampwalk") || !hasEffectiveKeyword(t, g, wraith, "islandwalk") {
		t.Error("swampwalk became islandwalk")
	}
}

// Magical Hack on a lord: the land type in the keyword it GRANTS is a
// word of the lord's text too. Goblin King's mountainwalk grant becomes
// an islandwalk grant.
func TestMagicalHackChangesAGrantedLandwalk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	king := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin King", TypeLine: "Creature — Goblin",
		OracleID: twGoblinKingOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	gob := twPut(g, me.ID, "Goblin Piker", "Creature — Goblin", []string{"R"})
	if !hasEffectiveKeyword(t, g, gob, "mountainwalk") {
		t.Fatal("setup: Goblin King grants mountainwalk")
	}
	twChange(t, g, "Magical Hack", twHackOracle, king, 3, 1) // Mountain → Island
	if hasEffectiveKeyword(t, g, gob, "mountainwalk") || !hasEffectiveKeyword(t, g, gob, "islandwalk") {
		t.Error("Goblin King's grant now reads islandwalk")
	}
}

// Mind Bend asks for ONE word from either class, then offers only the
// other words of the class picked.
func TestMindBendOffersEitherClassThenTheSameClass(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wraith := twPut(g, me.ID, "Bog Wraith", "Creature — Wraith", []string{"B"}, "swampwalk")
	castCatalogSpell(t, g, "Mind Bend", "Instant", twMindBendOracle, twTargetRef(wraith))
	passPriorityAroundTable(t, g)
	first := latestOptionPickFor(g, me.ID)
	if first == nil || len(first.PickOptions) != 10 {
		t.Fatalf("the first prompt offers the five colours and the five land types: %+v", first)
	}
	answerOptionPick(t, g, me.ID, 7) // Swamp
	second := latestOptionPickFor(g, me.ID)
	var labels []string
	for _, o := range second.PickOptions {
		labels = append(labels, o.Label)
	}
	if !slices.Equal(labels, []string{"Plains", "Island", "Mountain", "Forest"}) {
		t.Fatalf("the second prompt offers the other land types: %v", labels)
	}
	answerOptionPick(t, g, me.ID, 3) // Forest
	if !hasEffectiveKeyword(t, g, wraith, "forestwalk") {
		t.Error("swampwalk became forestwalk")
	}
}

// Crystal Spray changes the word until end of turn, then draws.
func TestCrystalSprayLastsTheTurnAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := twPut(g, me.ID, "White Knight", "Creature — Human Knight", []string{"W"}, "protection from black")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Crystal Spray", "Instant", twSprayOracle, twTargetRef(knight))
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("the card is drawn after the words are chosen, not before (hand %d → %d)", hand, me.Hand.Size())
	}
	answerOptionPick(t, g, me.ID, twBlack)
	answerOptionPick(t, g, me.ID, twGreen-1) // W, U, R, G: green is 3
	if me.Hand.Size() != hand+1 {
		t.Errorf("Crystal Spray draws a card (hand %d → %d)", hand, me.Hand.Size())
	}
	if !hasEffectiveKeyword(t, g, knight, "protection from green") {
		t.Fatal("protection from black reads protection from green this turn")
	}
	advanceToNextSeatsTurn(t, g)
	if !hasEffectiveKeyword(t, g, knight, "protection from black") {
		t.Error("the change ended at cleanup")
	}
}

// Whim of Volrath is Mind Bend until end of turn, with buyback.
func TestWhimOfVolrathLastsTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wraith := twPut(g, me.ID, "Bog Wraith", "Creature — Wraith", []string{"B"}, "swampwalk")
	twChange(t, g, "Whim of Volrath", twWhimOracle, wraith, 7, 0) // Swamp → Plains
	if !hasEffectiveKeyword(t, g, wraith, "plainswalk") {
		t.Fatal("swampwalk reads plainswalk this turn")
	}
	advanceToNextSeatsTurn(t, g)
	if !hasEffectiveKeyword(t, g, wraith, "swampwalk") {
		t.Error("the change ended at cleanup")
	}
	spec, _ := Lookup(twWhimOracle)
	if len(spec.OptionalCosts) != 1 {
		t.Error("Whim of Volrath has buyback")
	}
}

// Artificial Evolution on a lord changes the tribe it pumps, and on a
// creature its type line; the new type can't be Wall.
func TestArtificialEvolutionRewritesALordsTribe(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	king := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin King", TypeLine: "Creature — Goblin",
		OracleID: twGoblinKingOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	gob := twPut(g, me.ID, "Goblin Piker", "Creature — Goblin", []string{"R"})
	elf := twPut(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", []string{"G"})
	castCatalogSpell(t, g, "Artificial Evolution", "Instant", twEvolutionOracle, twTargetRef(king))
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Goblin")
	prompt := pendingOfKind(g, game.PendingChoiceCreatureType)
	if prompt == nil || !slices.Contains(prompt.ExcludedTypes, "Wall") || !slices.Contains(prompt.ExcludedTypes, "Goblin") {
		t.Fatalf("the new type can't be Wall or the old one: %+v", prompt)
	}
	if err := g.ResolveCreatureTypeChoice(prompt.ID, me.ID, "Wall"); err == nil {
		t.Fatal("Wall is refused as the new type")
	}
	answerCreatureType(t, g, me.ID, "Elf")
	if got := effectivePower(t, g, elf); got != 3 {
		t.Errorf("Goblin King is a lord of Elves now: the Elf's power = %d, want 3", got)
	}
	if got := effectivePower(t, g, gob); got != 2 {
		t.Errorf("and no longer of Goblins: the Goblin's power = %d, want 2", got)
	}
	if subs := effectiveSubtypes(t, g, king); !slices.Equal(subs, []string{"Elf"}) {
		t.Errorf("Goblin King's type line is text too: %v, want [Elf]", subs)
	}
}

// A change made to a permanent SPELL stays with the permanent it becomes
// (CR 400.7a; "the permanent will retain the text change").
func TestATextChangeOnAPermanentSpellCarriesOntoThePermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := twPut(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", []string{"G"})
	king := castCatalogSpell(t, g, "Goblin King", "Creature — Goblin", twGoblinKingOracle, nil)
	castCatalogSpell(t, g, "Artificial Evolution", "Instant", twEvolutionOracle, twTargetRef(king))
	for i := 0; i < 8 && pendingOfKind(g, game.PendingChoiceCreatureType) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	answerCreatureType(t, g, me.ID, "Goblin")
	answerCreatureType(t, g, me.ID, "Elf")
	passPriorityAroundTable(t, g)
	if !twOnBattlefield(g, king) {
		t.Fatal("Goblin King resolved")
	}
	if got := effectivePower(t, g, elf); got != 3 {
		t.Errorf("the permanent kept the spell's text change: the Elf's power = %d, want 3", got)
	}
}
