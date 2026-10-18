package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// innistrad_werewolf_cards_test.go — the original Innistrad werewolves
// and Immerwolf (#4071, ADR 0132 amendment 2026-10-16), on real
// imported cards. The two seams they stand on (the previous turn's
// spells by every player, and "can't transform") are pinned in
// game/werewolf_transform_test.go; these drive them through the catalog
// rows, the real upkeep and the real stack.

const (
	iwFrontText = "At the beginning of each upkeep, if no spells were cast last turn, transform this creature."
	iwBackText  = "At the beginning of each upkeep, if a player cast two or more spells last turn, transform this creature."

	iwImmerwolfOracle = "3c2445ee-84d5-47de-8b09-dbdc89ca2db7"
)

// iwCard is one Innistrad werewolf as the bulk dump has it.
type iwCard struct {
	oracle, front, frontType, cost, frontText, fp, ft string
	back, backText, bp, bt                            string
	colors, keywords                                  []string
}

var (
	iwDeserter = iwCard{"52adc761-1a0a-43d1-a706-41a92bb7880c", "Afflicted Deserter", "Creature — Human Werewolf", "{3}{R}", iwFrontText, "3", "2",
		"Werewolf Ransacker", "Whenever this creature transforms into Werewolf Ransacker, you may destroy target artifact. If that artifact is put into a graveyard this way, this creature deals 3 damage to that artifact's controller.\n" + iwBackText, "5", "4",
		[]string{"R"}, nil}
	iwRanger = iwCard{"280624aa-5f9a-48fd-85ea-815c96c747b3", "Daybreak Ranger", "Creature — Human Archer Ranger Werewolf", "{2}{G}",
		"{T}: This creature deals 2 damage to target creature with flying.\n" + iwFrontText, "2", "2",
		"Nightfall Predator", "{R}, {T}: This creature fights target creature. (Each deals damage equal to its power to the other.)\n" + iwBackText, "4", "4",
		[]string{"G"}, []string{"Fight"}}
	iwRecruiter = iwCard{"87c43776-dd98-4e12-9b95-82d2b8f4f1ab", "Duskwatch Recruiter", "Creature — Human Warrior Werewolf", "{1}{G}",
		"{2}{G}: Look at the top three cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in any order.\n" + iwFrontText, "2", "2",
		"Krallenhorde Howler", "Creature spells you cast cost {1} less to cast.\n" + iwBackText, "3", "3",
		[]string{"G"}, nil}
	iwBandit = iwCard{"64720167-b88b-47f7-bfc1-d5e005c9fd68", "Geier Reach Bandit", "Creature — Human Rogue Werewolf", "{2}{R}",
		"Haste\n" + iwFrontText, "3", "2",
		"Vildin-Pack Alpha", "Whenever a Werewolf you control enters, you may transform it.\n" + iwBackText, "4", "3",
		[]string{"R"}, []string{"Haste"}}
	iwHuntmaster = iwCard{"582328cd-660d-47a4-bb23-e91e80b9a907", "Huntmaster of the Fells", "Creature — Human Werewolf", "{2}{R}{G}",
		"Whenever this creature enters or transforms into Huntmaster of the Fells, create a 2/2 green Wolf creature token and you gain 2 life.\n" + iwFrontText, "2", "2",
		"Ravager of the Fells", "Trample\nWhenever this creature transforms into Ravager of the Fells, it deals 2 damage to target opponent or planeswalker and 2 damage to up to one target creature that player or that planeswalker's controller controls.\n" + iwBackText, "4", "4",
		[]string{"R", "G"}, []string{"Trample"}}
	iwGang = iwCard{"27c2bb87-7e50-4a80-b41c-ff0237a3d648", "Instigator Gang", "Creature — Human Werewolf", "{3}{R}",
		"Attacking creatures you control get +1/+0.\n" + iwFrontText, "2", "3",
		"Wildblood Pack", "Trample\nAttacking creatures you control get +3/+0.\n" + iwBackText, "5", "5",
		[]string{"R"}, []string{"Trample"}}
	iwOutlaw = iwCard{"cf518f59-f083-42a0-9953-fac2757bf804", "Kruin Outlaw", "Creature — Human Rogue Werewolf", "{1}{R}{R}",
		"First strike\n" + iwFrontText, "2", "2",
		"Terror of Kruin Pass", "Double strike\nWerewolves you control have menace. (A creature with menace can't be blocked except by two or more creatures.)\n" + iwBackText, "3", "3",
		[]string{"R"}, []string{"First strike", "Double strike"}}
	iwMayor = iwCard{"6e638587-30b3-4c4b-b463-fb415ea048f7", "Mayor of Avabruck", "Creature — Human Advisor Werewolf", "{1}{G}",
		"Other Human creatures you control get +1/+1.\n" + iwFrontText, "1", "1",
		"Howlpack Alpha", "Each other creature you control that's a Werewolf or a Wolf gets +1/+1.\nAt the beginning of your end step, create a 2/2 green Wolf creature token.\n" + iwBackText, "3", "3",
		[]string{"G"}, nil}
	iwShaman = iwCard{"f94cb722-7e46-4569-b474-9654ecf7ec14", "Mondronen Shaman", "Creature — Human Shaman Werewolf", "{3}{R}",
		iwFrontText, "3", "2",
		"Tovolar's Magehunter", "Whenever an opponent casts a spell, this creature deals 2 damage to that player.\n" + iwBackText, "5", "5",
		[]string{"R"}, nil}
	iwSage = iwCard{"a46454ad-d912-42f7-b228-492796161d82", "Sage of Ancient Lore", "Creature — Human Shaman Werewolf", "{4}{G}",
		"Sage of Ancient Lore's power and toughness are each equal to the number of cards in your hand.\nWhen this creature enters, draw a card.\n" + iwFrontText, "*", "*",
		"Werewolf of Ancient Hunger", "Vigilance, trample\nWerewolf of Ancient Hunger's power and toughness are each equal to the total number of cards in all players' hands.\n" + iwBackText, "*", "*",
		[]string{"G"}, []string{"Vigilance", "Trample"}}
	iwVillager = iwCard{"52855f90-19c1-46c9-8eed-88b3c1722bb0", "Scorned Villager", "Creature — Human Werewolf", "{1}{G}",
		"{T}: Add {G}.\n" + iwFrontText, "1", "1",
		"Moonscarred Werewolf", "Vigilance\n{T}: Add {G}{G}.\n" + iwBackText, "2", "2",
		[]string{"G"}, []string{"Vigilance"}}
)

var iwAll = []iwCard{iwDeserter, iwRanger, iwRecruiter, iwBandit, iwHuntmaster, iwGang, iwOutlaw, iwMayor, iwShaman, iwSage, iwVillager}

// row is the Scryfall `transform` record the deck importer reads.
func (w iwCard) row() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     w.front + " // " + w.back,
		Layout:   "transform",
		TypeLine: w.frontType + " // Creature — Werewolf",
		OracleID: uuid.MustParse(w.oracle),
		Keywords: append([]string{"Transform"}, w.keywords...),
		CardFaces: []cards.CardFace{
			{Name: w.front, TypeLine: w.frontType, ManaCost: w.cost, Colors: w.colors, Power: w.fp, Toughness: w.ft, OracleText: w.frontText},
			{Name: w.back, TypeLine: "Creature — Werewolf", Colors: w.colors, Power: w.bp, Toughness: w.bt, OracleText: w.backText},
		},
	}
}

// iwPut puts the imported werewolf onto p's battlefield front face up,
// able to tap.
func iwPut(t *testing.T, g *game.Game, w iwCard, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(w.row(), p)
	c, ok := cardInZone(p.Hand, id)
	if !ok {
		t.Fatalf("%s did not reach the hand", w.front)
	}
	p.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	return id
}

// iwPutBack is iwPut on its back face, turned over by an instruction (no
// daybound, so an instruction can).
func iwPutBack(t *testing.T, g *game.Game, w iwCard, p *game.Player) uuid.UUID {
	t.Helper()
	id := iwPut(t, g, w, p)
	iwTransform(t, g, id)
	if iwFace(g, id) != 1 {
		t.Fatalf("%s did not turn over to %s", w.front, w.back)
	}
	return id
}

func iwTransform(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.TransformPermanentForEffect(id) })
	if err != nil {
		t.Fatalf("TransformPermanentForEffect: %v", err)
	}
}

func iwFace(g *game.Game, id uuid.UUID) int {
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.ActiveFace
}

// iwNextUpkeep ends the current turn with `casts` (seat -> spells cast
// this turn, standing in for real casts) as its record, walks to the
// next player's upkeep and settles what triggered there. It returns that
// player's seat.
func iwNextUpkeep(t *testing.T, g *game.Game, casts map[int]int) int {
	t.Helper()
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
		for seat, n := range casts {
			g.SpellsCastThisTurn[g.Seats[seat].ID] = game.CastTally{Total: n}
		}
	})
	start := g.Turn.ActiveSeat
	for i := 0; i < 40; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat != start {
			monSettle(t, g)
			return g.Turn.ActiveSeat
		}
	}
	t.Fatal("never reached the next upkeep")
	return -1
}

// iwTable is a catalog game in the active seat's precombat main.
func iwTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat]
}

func TestInnistradWerewolvesAreRegistered(t *testing.T) {
	for _, w := range iwAll {
		front, ok := Lookup(w.oracle)
		if !ok || front.Name != w.front || front.Completeness != CompletenessFull {
			t.Errorf("%s: front registered %v as %+v", w.front, ok, front.Name)
		}
		back, ok := Lookup(w.oracle + "#1")
		if !ok || back.Name != w.back || back.Completeness != CompletenessFull {
			t.Errorf("%s: back registered %v as %+v", w.back, ok, back.Name)
		}
	}
	if im, ok := Lookup(iwImmerwolfOracle); !ok || im.Name != "Immerwolf" || im.Completeness != CompletenessFull {
		t.Errorf("Immerwolf registered %v as %q", ok, im.Name)
	}
}

// A quiet turn turns every front over at the NEXT player's upkeep (an
// opponent's), and the face that comes up brings its own abilities.
func TestInnistradWerewolfTransformsOnAnOpponentsUpkeepAfterAQuietTurn(t *testing.T) {
	g, me := iwTable(t)
	id := iwPut(t, g, iwOutlaw, me)
	if !slices.Contains(effectiveAbilities(t, g, id), "first strike") {
		t.Fatalf("Kruin Outlaw lacks first strike: %v", effectiveAbilities(t, g, id))
	}
	seat := iwNextUpkeep(t, g, nil)
	if g.Seats[seat].ID == me.ID {
		t.Fatal("the next upkeep is the controller's own; the test wants an opponent's")
	}
	if iwFace(g, id) != 1 {
		t.Fatalf("Kruin Outlaw is on face %d after a turn with no spells, want Terror of Kruin Pass", iwFace(g, id))
	}
	abilities := effectiveAbilities(t, g, id)
	if !slices.Contains(abilities, "double strike") || !slices.Contains(abilities, "menace") {
		t.Errorf("Terror of Kruin Pass abilities = %v, want double strike and menace (a Werewolf you control)", abilities)
	}
}

// Any spell, by anyone, keeps the front up.
func TestInnistradWerewolfStaysAfterAnyonesSpell(t *testing.T) {
	g, me := iwTable(t)
	id := iwPut(t, g, iwOutlaw, me)
	opp := (g.Turn.ActiveSeat + 2) % len(g.Seats)
	iwNextUpkeep(t, g, map[int]int{opp: 1})
	if iwFace(g, id) != 0 {
		t.Errorf("an opponent's one spell last turn still turned Kruin Outlaw over (face %d)", iwFace(g, id))
	}
}

// The back face turns back only when ONE player cast two: two players
// casting one each is two spells, but not "a player cast two".
func TestInnistradWerewolfTurnsBackOnlyWhenOnePlayerCastTwo(t *testing.T) {
	g, me := iwTable(t)
	id := iwPutBack(t, g, iwOutlaw, me)
	a, b := g.Turn.ActiveSeat, (g.Turn.ActiveSeat+1)%len(g.Seats)
	iwNextUpkeep(t, g, map[int]int{a: 1, b: 1})
	if iwFace(g, id) != 1 {
		t.Fatalf("two players casting one spell each turned Terror of Kruin Pass back (face %d)", iwFace(g, id))
	}
	iwNextUpkeep(t, g, map[int]int{b: 2})
	if iwFace(g, id) != 0 {
		t.Errorf("one player cast two spells and Terror of Kruin Pass stayed (face %d)", iwFace(g, id))
	}
}

// The real casts count: the active player casting two spells in their
// own main phase turns the back face over at the next upkeep.
func TestInnistradWerewolfReadsRealCasts(t *testing.T) {
	g, me := iwTable(t)
	id := iwPutBack(t, g, iwOutlaw, me)
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	start := g.Turn.ActiveSeat
	for i := 0; i < 40 && (g.Turn.Step != game.StepUpkeep || g.Turn.ActiveSeat == start); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	monSettle(t, g)
	if iwFace(g, id) != 0 {
		t.Errorf("two real casts last turn left Terror of Kruin Pass up (face %d)", iwFace(g, id))
	}
}

// --- Immerwolf -----------------------------------------------------------

// Immerwolf holds a back face (a non-Human Werewolf) but not a front
// (a Human Werewolf), and pumps the other Wolves and Werewolves.
func TestImmerwolfHoldsTheBackFaceOnly(t *testing.T) {
	g, me := iwTable(t)
	wolf := b12Push(g, me.ID, "Immerwolf", "Creature — Wolf", iwImmerwolfOracle, 2, 2)
	terror := iwPutBack(t, g, iwOutlaw, me) // the Human front may still turn over
	outlaw := iwPut(t, g, iwOutlaw, me)
	if p, _ := effectivePTOf(t, g, terror); p != 4 {
		t.Errorf("Terror of Kruin Pass power = %d, want 3 + Immerwolf's 1", p)
	}
	if p, _ := effectivePTOf(t, g, wolf); p != 2 {
		t.Errorf("Immerwolf power = %d, want 2: it pumps only the OTHER Wolves and Werewolves", p)
	}

	seat := g.Turn.ActiveSeat
	iwNextUpkeep(t, g, map[int]int{seat: 2})
	if iwFace(g, terror) != 1 {
		t.Errorf("Terror of Kruin Pass turned back under Immerwolf (face %d)", iwFace(g, terror))
	}
	if iwFace(g, outlaw) != 0 {
		t.Fatalf("the front was not up to begin with (face %d)", iwFace(g, outlaw))
	}
	iwNextUpkeep(t, g, nil)
	if iwFace(g, outlaw) != 1 {
		t.Errorf("Kruin Outlaw, a Human Werewolf, did not transform under Immerwolf (face %d)", iwFace(g, outlaw))
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(wolf); err != nil {
			t.Fatalf("destroy Immerwolf: %v", err)
		}
	})
	iwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
	if iwFace(g, terror) != 0 || iwFace(g, outlaw) != 0 {
		t.Errorf("with Immerwolf gone both should turn back: faces %d and %d", iwFace(g, terror), iwFace(g, outlaw))
	}
}

// Tovolar, the Midnight Scourge is a nightbound Werewolf and not a
// Human, so under Immerwolf day does not turn it back.
func TestImmerwolfKeepsTovolarAtNightByDay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	id := importAndCast(t, g, dnbWerewolf(dnbTovolar).row(), me)
	monSettle(t, g)
	if iwFace(g, id) != 1 {
		t.Fatalf("Tovolar cast at night is on face %d, want the Midnight Scourge", iwFace(g, id))
	}
	b12Push(g, me.ID, "Immerwolf", "Creature — Wolf", iwImmerwolfOracle, 2, 2)
	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	if iwFace(g, id) != 1 {
		t.Errorf("day turned Tovolar, the Midnight Scourge over under Immerwolf (face %d)", iwFace(g, id))
	}
}

// --- the cards -----------------------------------------------------------

// Huntmaster: a Wolf and 2 life as it enters, the Ravager's two
// related targets as it turns over, and the Huntmaster's trigger again
// as it turns back.
func TestHuntmasterOfTheFellsAndRavager(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := importAndCast(t, g, iwHuntmaster.row(), me)
	monSettle(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Wolf"); n != 1 || me.Life != life+2 {
		t.Fatalf("on entering: %d Wolves and life %d, want 1 and %d", n, me.Life, life+2)
	}

	nextSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	opp := g.Seats[nextSeat]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 4)
	elsewhere := b12Creature(g, other.ID, "Another Bear", "Creature — Bear", 2, 4)
	oppLife := opp.Life

	iwNextUpkeep(t, g, nil)
	if iwFace(g, id) != 1 {
		t.Fatalf("Huntmaster did not turn over after a quiet turn (face %d)", iwFace(g, id))
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Ravager of the Fells asked for no target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("target the opponent: %v", err)
	}
	pick = latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the Ravager's creature")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: elsewhere, Slot: 1}}); err == nil {
		t.Fatal("a creature another player controls was accepted as the second target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 1}}); err != nil {
		t.Fatalf("target the opponent's creature: %v", err)
	}
	monSettle(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, oppLife-2)
	}
	if c, _ := battlefieldCard(g, theirs); c.DamageMarked != 2 {
		t.Errorf("their creature has %d damage, want 2", c.DamageMarked)
	}

	life = me.Life
	iwNextUpkeep(t, g, map[int]int{nextSeat: 2})
	if iwFace(g, id) != 0 {
		t.Fatalf("Ravager did not turn back after a two-spell turn (face %d)", iwFace(g, id))
	}
	if n, _ := countTokensNamed(g, me.ID, "Wolf"); n != 2 || me.Life != life+2 {
		t.Errorf("turning back into Huntmaster: %d Wolves and life %d, want 2 and %d", n, me.Life, life+2)
	}
}

// The Ravager's second target is optional.
func TestRavagerOfTheFellsMayTargetNoCreature(t *testing.T) {
	g, me := iwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := iwPut(t, g, iwHuntmaster, me)
	oppLife := opp.Life
	iwNextUpkeep(t, g, nil)
	if iwFace(g, id) != 1 {
		t.Fatalf("no turn-over (face %d)", iwFace(g, id))
	}
	answerPickTargetPlayer(t, g, opp.ID)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("declining the creature: %v", err)
		}
	}
	monSettle(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, oppLife-2)
	}
}

// Werewolf Ransacker: the transform targets an artifact and asks; yes
// destroys it and burns its controller.
func TestWerewolfRansackerDestroysAnArtifactAndBurnsItsController(t *testing.T) {
	g, me := iwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	id := iwPut(t, g, iwDeserter, me)
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	oppLife := opp.Life
	iwNextUpkeep(t, g, nil)
	if iwFace(g, id) != 1 {
		t.Fatalf("Afflicted Deserter did not turn over (face %d)", iwFace(g, id))
	}
	answerPickTarget(t, g, relic)
	monSettle(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	monSettle(t, g)
	if _, ok := battlefieldCard(g, relic); ok {
		t.Fatal("the artifact was not destroyed")
	}
	if opp.Life != oppLife-3 {
		t.Errorf("the artifact's controller is at %d, want %d", opp.Life, oppLife-3)
	}
}

// Vildin-Pack Alpha asks about a Werewolf that can transform and says
// nothing about a daybound one.
func TestVildinPackAlphaMayTransformAnEnteringWerewolf(t *testing.T) {
	g, me := iwTable(t)
	iwPutBack(t, g, iwBandit, me)
	outlaw := importAndCast(t, g, iwOutlaw.row(), me)
	monSettle(t, g)
	answerMayChoice(t, g, me.ID, true)
	monSettle(t, g)
	if iwFace(g, outlaw) != 1 {
		t.Fatalf("the entering Kruin Outlaw was not transformed (face %d)", iwFace(g, outlaw))
	}

	villager := importAndCast(t, g, fearfulVillagerRow(), me)
	monSettle(t, g)
	if c := latestConfirmFor(g, me.ID); c != nil {
		t.Errorf("a daybound Werewolf, which nothing but day and night may transform, raised a yes/no prompt %v", c.ID)
	}
	if iwFace(g, villager) != 0 {
		t.Errorf("the daybound Werewolf turned over (face %d)", iwFace(g, villager))
	}
}

// Sage of Ancient Lore counts your hand; Werewolf of Ancient Hunger
// counts every hand.
func TestSageOfAncientLoreCountsHands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := importAndCast(t, g, iwSage.row(), me)
	monSettle(t, g)
	if p, tough := effectivePTOf(t, g, id); p != me.Hand.Size() || tough != me.Hand.Size() {
		t.Errorf("Sage is %d/%d with %d cards in hand", p, tough, me.Hand.Size())
	}
	iwTransform(t, g, id)
	total := 0
	for _, p := range g.Seats {
		total += p.Hand.Size()
	}
	if p, tough := effectivePTOf(t, g, id); p != total || tough != total {
		t.Errorf("Werewolf of Ancient Hunger is %d/%d with %d cards in all hands", p, tough, total)
	}
	abilities := effectiveAbilities(t, g, id)
	if !slices.Contains(abilities, "vigilance") || !slices.Contains(abilities, "trample") {
		t.Errorf("Werewolf of Ancient Hunger abilities = %v", abilities)
	}
}

func TestSageOfAncientLoreDrawsAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	importAndCast(t, g, iwSage.row(), me) // into the hand, then onto the stack
	monSettle(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand = %d, want %d: the entry trigger draws one", me.Hand.Size(), before+1)
	}
}

// Krallenhorde Howler discounts your creature spells only.
func TestKrallenhordeHowlerDiscountsCreatureSpells(t *testing.T) {
	g, me := iwTable(t)
	iwPutBack(t, g, iwRecruiter, me)
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{2}{G}"); got != 2 {
		t.Errorf("a creature spell costs %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Growth", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("a sorcery costs %d, want 3", got)
	}
}

// Duskwatch Recruiter looks at three, takes a creature and has the
// player order the rest on the bottom.
func TestDuskwatchRecruiterTakesACreature(t *testing.T) {
	g, me := iwTable(t)
	id := iwPut(t, g, iwRecruiter, me)
	bear := uuid.New()
	ids := seedLibrary(me, "A", "B")
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	floatMana(t, g, me, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 1 {
		t.Fatalf("no take-one prompt: %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("take the Bear: %v", err)
	}
	if !me.Hand.Contains(bear) {
		t.Fatal("the creature card is not in hand")
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom {
		t.Fatalf("the rest are not ordered on the bottom: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[1], ids[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := libraryBottomIDs(me, 2); !sameIDs(got, []uuid.UUID{ids[1], ids[0]}) {
		t.Errorf("bottom = %v, want the player's order", got)
	}
}

// Tovolar's Magehunter burns an opponent who casts, not its controller.
func TestTovolarsMagehunterBurnsACaster(t *testing.T) {
	g, me := iwTable(t)
	iwPutBack(t, g, iwShaman, me)
	myLife := me.Life
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Errorf("the controller's own spell cost them %d life", myLife-me.Life)
	}
	seat := iwNextUpkeep(t, g, nil)
	opp := g.Seats[seat]
	advanceToMainOf(t, g, seat)
	oppLife := opp.Life
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the casting opponent is at %d, want %d", opp.Life, oppLife-2)
	}
}

// Mayor of Avabruck pumps your other Humans; Howlpack Alpha your other
// Wolves and Werewolves, and makes a Wolf at your end step.
func TestMayorOfAvabruckAndHowlpackAlpha(t *testing.T) {
	g, me := iwTable(t)
	mayor := iwPut(t, g, iwMayor, me)
	human := b12Creature(g, me.ID, "Villager", "Creature — Human", 1, 1)
	if p, _ := effectivePTOf(t, g, human); p != 2 {
		t.Errorf("a Human under the Mayor has power %d, want 2", p)
	}
	if p, _ := effectivePTOf(t, g, mayor); p != 1 {
		t.Errorf("the Mayor pumps itself: power %d", p)
	}
	iwTransform(t, g, mayor)
	if p, _ := effectivePTOf(t, g, human); p != 1 {
		t.Errorf("the Human kept the Mayor's bonus after it turned over: power %d", p)
	}
	advanceTo(t, g, game.StepEnd)
	monSettle(t, g)
	n, _ := countTokensNamed(g, me.ID, "Wolf")
	if n != 1 {
		t.Fatalf("Howlpack Alpha made %d Wolves at the end step, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Wolf" {
			if p, _ := effectivePTOf(t, g, c.InstanceID); p != 3 {
				t.Errorf("the Wolf token is %d power under the Alpha, want 3", p)
			}
		}
	}
}

// Instigator Gang pumps your attackers.
func TestInstigatorGangPumpsAttackers(t *testing.T) {
	g, me := iwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	iwPut(t, g, iwGang, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Fatalf("a creature not attacking has power %d, want 2", p)
	}
	declareAttack(t, g, opp.ID, bear)
	if p, _ := effectivePTOf(t, g, bear); p != 3 {
		t.Errorf("an attacking Bear has power %d, want 3", p)
	}
}

// Daybreak Ranger shoots a flier; Nightfall Predator fights.
func TestDaybreakRangerAndNightfallPredator(t *testing.T) {
	g, me := iwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ranger := iwPut(t, g, iwRanger, me)
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 3,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, ranger, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bird}},
	}); err != nil {
		t.Fatalf("activate the Ranger: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, bird); c.DamageMarked != 2 {
		t.Errorf("the flier has %d damage, want 2", c.DamageMarked)
	}

	iwTransform(t, g, ranger)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ranger {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, ranger, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate the Predator: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, bear); ok {
		t.Error("the Bear survived 4 damage")
	}
	if c, _ := battlefieldCard(g, ranger); c.DamageMarked != 2 {
		t.Errorf("Nightfall Predator has %d damage, want the Bear's 2", c.DamageMarked)
	}
}

// Scorned Villager taps for {G}; Moonscarred Werewolf for {G}{G}.
func TestScornedVillagerManaAbilities(t *testing.T) {
	front, _ := Lookup(iwVillager.oracle)
	back, _ := Lookup(iwVillager.oracle + "#1")
	if len(front.ManaAbilities) != 1 || front.ManaAbilities[0].Produced != "{G}" {
		t.Errorf("Scorned Villager mana abilities = %+v", front.ManaAbilities)
	}
	if len(back.ManaAbilities) != 1 || back.ManaAbilities[0].Produced != "{G}{G}" {
		t.Errorf("Moonscarred Werewolf mana abilities = %+v", back.ManaAbilities)
	}
	if !slices.Contains(back.PrintedKeywords, "vigilance") {
		t.Errorf("Moonscarred Werewolf keywords = %v", back.PrintedKeywords)
	}
}
