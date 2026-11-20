package testsi

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var (
	iwDeserter = effectstest.IwCard{Oracle: "52adc761-1a0a-43d1-a706-41a92bb7880c", Front: "Afflicted Deserter", FrontType: "Creature — Human Werewolf", Cost: "{3}{R}", FrontText: effectstest.IwFrontText, Fp: "3", Ft: "2",
		Back: "Werewolf Ransacker", BackText: "Whenever this creature transforms into Werewolf Ransacker, you may destroy target artifact. If that artifact is put into a graveyard this way, this creature deals 3 damage to that artifact's controller.\n" + effectstest.IwBackText, Bp: "5", Bt: "4",
		Colors: []string{"R"}, Keywords: nil}
	iwRanger = effectstest.IwCard{Oracle: "280624aa-5f9a-48fd-85ea-815c96c747b3", Front: "Daybreak Ranger", FrontType: "Creature — Human Archer Ranger Werewolf", Cost: "{2}{G}",
		FrontText: "{T}: This creature deals 2 damage to target creature with flying.\n" + effectstest.IwFrontText, Fp: "2", Ft: "2",
		Back: "Nightfall Predator", BackText: "{R}, {T}: This creature fights target creature. (Each deals damage equal to its power to the other.)\n" + effectstest.IwBackText, Bp: "4", Bt: "4",
		Colors: []string{"G"}, Keywords: []string{"Fight"}}
	iwRecruiter = effectstest.IwCard{Oracle: "87c43776-dd98-4e12-9b95-82d2b8f4f1ab", Front: "Duskwatch Recruiter", FrontType: "Creature — Human Warrior Werewolf", Cost: "{1}{G}",
		FrontText: "{2}{G}: Look at the top three cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in any order.\n" + effectstest.IwFrontText, Fp: "2", Ft: "2",
		Back: "Krallenhorde Howler", BackText: "Creature spells you cast cost {1} less to cast.\n" + effectstest.IwBackText, Bp: "3", Bt: "3",
		Colors: []string{"G"}, Keywords: nil}
	iwBandit = effectstest.IwCard{Oracle: "64720167-b88b-47f7-bfc1-d5e005c9fd68", Front: "Geier Reach Bandit", FrontType: "Creature — Human Rogue Werewolf", Cost: "{2}{R}",
		FrontText: "Haste\n" + effectstest.IwFrontText, Fp: "3", Ft: "2",
		Back: "Vildin-Pack Alpha", BackText: "Whenever a Werewolf you control enters, you may transform it.\n" + effectstest.IwBackText, Bp: "4", Bt: "3",
		Colors: []string{"R"}, Keywords: []string{"Haste"}}
	iwHuntmaster = effectstest.IwCard{Oracle: "582328cd-660d-47a4-bb23-e91e80b9a907", Front: "Huntmaster of the Fells", FrontType: "Creature — Human Werewolf", Cost: "{2}{R}{G}",
		FrontText: "Whenever this creature enters or transforms into Huntmaster of the Fells, create a 2/2 green Wolf creature token and you gain 2 life.\n" + effectstest.IwFrontText, Fp: "2", Ft: "2",
		Back: "Ravager of the Fells", BackText: "Trample\nWhenever this creature transforms into Ravager of the Fells, it deals 2 damage to target opponent or planeswalker and 2 damage to up to one target creature that player or that planeswalker's controller controls.\n" + effectstest.IwBackText, Bp: "4", Bt: "4",
		Colors: []string{"R", "G"}, Keywords: []string{"Trample"}}
	iwGang = effectstest.IwCard{Oracle: "27c2bb87-7e50-4a80-b41c-ff0237a3d648", Front: "Instigator Gang", FrontType: "Creature — Human Werewolf", Cost: "{3}{R}",
		FrontText: "Attacking creatures you control get +1/+0.\n" + effectstest.IwFrontText, Fp: "2", Ft: "3",
		Back: "Wildblood Pack", BackText: "Trample\nAttacking creatures you control get +3/+0.\n" + effectstest.IwBackText, Bp: "5", Bt: "5",
		Colors: []string{"R"}, Keywords: []string{"Trample"}}
	iwOutlaw = effectstest.IwCard{Oracle: "cf518f59-f083-42a0-9953-fac2757bf804", Front: "Kruin Outlaw", FrontType: "Creature — Human Rogue Werewolf", Cost: "{1}{R}{R}",
		FrontText: "First strike\n" + effectstest.IwFrontText, Fp: "2", Ft: "2",
		Back: "Terror of Kruin Pass", BackText: "Double strike\nWerewolves you control have menace. (A creature with menace can't be blocked except by two or more creatures.)\n" + effectstest.IwBackText, Bp: "3", Bt: "3",
		Colors: []string{"R"}, Keywords: []string{"First strike", "Double strike"}}
	iwMayor = effectstest.IwCard{Oracle: "6e638587-30b3-4c4b-b463-fb415ea048f7", Front: "Mayor of Avabruck", FrontType: "Creature — Human Advisor Werewolf", Cost: "{1}{G}",
		FrontText: "Other Human creatures you control get +1/+1.\n" + effectstest.IwFrontText, Fp: "1", Ft: "1",
		Back: "Howlpack Alpha", BackText: "Each other creature you control that's a Werewolf or a Wolf gets +1/+1.\nAt the beginning of your end step, create a 2/2 green Wolf creature token.\n" + effectstest.IwBackText, Bp: "3", Bt: "3",
		Colors: []string{"G"}, Keywords: nil}
	iwShaman = effectstest.IwCard{Oracle: "f94cb722-7e46-4569-b474-9654ecf7ec14", Front: "Mondronen Shaman", FrontType: "Creature — Human Shaman Werewolf", Cost: "{3}{R}",
		FrontText: effectstest.IwFrontText, Fp: "3", Ft: "2",
		Back: "Tovolar's Magehunter", BackText: "Whenever an opponent casts a spell, this creature deals 2 damage to that player.\n" + effectstest.IwBackText, Bp: "5", Bt: "5",
		Colors: []string{"R"}, Keywords: nil}
	iwSage = effectstest.IwCard{Oracle: "a46454ad-d912-42f7-b228-492796161d82", Front: "Sage of Ancient Lore", FrontType: "Creature — Human Shaman Werewolf", Cost: "{4}{G}",
		FrontText: "Sage of Ancient Lore's power and toughness are each equal to the number of cards in your hand.\nWhen this creature enters, draw a card.\n" + effectstest.IwFrontText, Fp: "*", Ft: "*",
		Back: "Werewolf of Ancient Hunger", BackText: "Vigilance, trample\nWerewolf of Ancient Hunger's power and toughness are each equal to the total number of cards in all players' hands.\n" + effectstest.IwBackText, Bp: "*", Bt: "*",
		Colors: []string{"G"}, Keywords: []string{"Vigilance", "Trample"}}
	iwVillager = effectstest.IwCard{Oracle: "52855f90-19c1-46c9-8eed-88b3c1722bb0", Front: "Scorned Villager", FrontType: "Creature — Human Werewolf", Cost: "{1}{G}",
		FrontText: "{T}: Add {G}.\n" + effectstest.IwFrontText, Fp: "1", Ft: "1",
		Back: "Moonscarred Werewolf", BackText: "Vigilance\n{T}: Add {G}{G}.\n" + effectstest.IwBackText, Bp: "2", Bt: "2",
		Colors: []string{"G"}, Keywords: []string{"Vigilance"}}
)

var iwAll = []effectstest.IwCard{iwDeserter, iwRanger, iwRecruiter, iwBandit, iwHuntmaster, iwGang, iwOutlaw, iwMayor, iwShaman, iwSage, iwVillager}

// iwPutBack is iwPut on its back face, turned over by an instruction (no
// daybound, so an instruction can).
func iwPutBack(t *testing.T, g *game.Game, w effectstest.IwCard, p *game.Player) uuid.UUID {
	t.Helper()
	id := effectstest.IwPut(t, g, w, p)
	effectstest.IwTransform(t, g, id)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("%s did not turn over to %s", w.Front, w.Back)
	}
	return id
}

func TestInnistradWerewolvesAreRegistered(t *testing.T) {
	for _, w := range iwAll {
		front, ok := effects.Lookup(w.Oracle)
		if !ok || front.Name != w.Front || front.Completeness != effects.CompletenessFull {
			t.Errorf("%s: front registered %v as %+v", w.Front, ok, front.Name)
		}
		back, ok := effects.Lookup(w.Oracle + "#1")
		if !ok || back.Name != w.Back || back.Completeness != effects.CompletenessFull {
			t.Errorf("%s: back registered %v as %+v", w.Back, ok, back.Name)
		}
	}
	if im, ok := effects.Lookup(effectstest.IwImmerwolfOracle); !ok || im.Name != "Immerwolf" || im.Completeness != effects.CompletenessFull {
		t.Errorf("Immerwolf registered %v as %q", ok, im.Name)
	}
}

// A quiet turn turns every front over at the NEXT player's upkeep (an
// opponent's), and the face that comes up brings its own abilities.
func TestInnistradWerewolfTransformsOnAnOpponentsUpkeepAfterAQuietTurn(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, iwOutlaw, me)
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, id), "first strike") {
		t.Fatalf("Kruin Outlaw lacks first strike: %v", effectstest.EffectiveAbilities(t, g, id))
	}
	seat := effectstest.IwNextUpkeep(t, g, nil)
	if g.Seats[seat].ID == me.ID {
		t.Fatal("the next upkeep is the controller's own; the test wants an opponent's")
	}
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("Kruin Outlaw is on face %d after a turn with no spells, want Terror of Kruin Pass", effectstest.IwFace(g, id))
	}
	abilities := effectstest.EffectiveAbilities(t, g, id)
	if !slices.Contains(abilities, "double strike") || !slices.Contains(abilities, "menace") {
		t.Errorf("Terror of Kruin Pass abilities = %v, want double strike and menace (a Werewolf you control)", abilities)
	}
}

// Any spell, by anyone, keeps the front up.
func TestInnistradWerewolfStaysAfterAnyonesSpell(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, iwOutlaw, me)
	opp := (g.Turn.ActiveSeat + 2) % len(g.Seats)
	effectstest.IwNextUpkeep(t, g, map[int]int{opp: 1})
	if effectstest.IwFace(g, id) != 0 {
		t.Errorf("an opponent's one spell last turn still turned Kruin Outlaw over (face %d)", effectstest.IwFace(g, id))
	}
}

// The back face turns back only when ONE player cast two: two players
// casting one each is two spells, but not "a player cast two".
func TestInnistradWerewolfTurnsBackOnlyWhenOnePlayerCastTwo(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := iwPutBack(t, g, iwOutlaw, me)
	a, b := g.Turn.ActiveSeat, (g.Turn.ActiveSeat+1)%len(g.Seats)
	effectstest.IwNextUpkeep(t, g, map[int]int{a: 1, b: 1})
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("two players casting one spell each turned Terror of Kruin Pass back (face %d)", effectstest.IwFace(g, id))
	}
	effectstest.IwNextUpkeep(t, g, map[int]int{b: 2})
	if effectstest.IwFace(g, id) != 0 {
		t.Errorf("one player cast two spells and Terror of Kruin Pass stayed (face %d)", effectstest.IwFace(g, id))
	}
}

// The real casts count: the active player casting two spells in their
// own main phase turns the back face over at the next upkeep.
func TestInnistradWerewolfReadsRealCasts(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := iwPutBack(t, g, iwOutlaw, me)
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	start := g.Turn.ActiveSeat
	for i := 0; i < 40 && (g.Turn.Step != game.StepUpkeep || g.Turn.ActiveSeat == start); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.MonSettle(t, g)
	if effectstest.IwFace(g, id) != 0 {
		t.Errorf("two real casts last turn left Terror of Kruin Pass up (face %d)", effectstest.IwFace(g, id))
	}
}

// --- Immerwolf -----------------------------------------------------------

// Immerwolf holds a back face (a non-Human Werewolf) but not a front
// (a Human Werewolf), and pumps the other Wolves and Werewolves.
func TestImmerwolfHoldsTheBackFaceOnly(t *testing.T) {
	g, me := effectstest.IwTable(t)
	wolf := effectstest.B12Push(g, me.ID, "Immerwolf", "Creature — Wolf", effectstest.IwImmerwolfOracle, 2, 2)
	terror := iwPutBack(t, g, iwOutlaw, me) // the Human front may still turn over
	outlaw := effectstest.IwPut(t, g, iwOutlaw, me)
	if p, _ := effectstest.EffectivePTOf(t, g, terror); p != 4 {
		t.Errorf("Terror of Kruin Pass power = %d, want 3 + Immerwolf's 1", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, wolf); p != 2 {
		t.Errorf("Immerwolf power = %d, want 2: it pumps only the OTHER Wolves and Werewolves", p)
	}

	seat := g.Turn.ActiveSeat
	effectstest.IwNextUpkeep(t, g, map[int]int{seat: 2})
	if effectstest.IwFace(g, terror) != 1 {
		t.Errorf("Terror of Kruin Pass turned back under Immerwolf (face %d)", effectstest.IwFace(g, terror))
	}
	if effectstest.IwFace(g, outlaw) != 0 {
		t.Fatalf("the front was not up to begin with (face %d)", effectstest.IwFace(g, outlaw))
	}
	effectstest.IwNextUpkeep(t, g, nil)
	if effectstest.IwFace(g, outlaw) != 1 {
		t.Errorf("Kruin Outlaw, a Human Werewolf, did not transform under Immerwolf (face %d)", effectstest.IwFace(g, outlaw))
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(wolf); err != nil {
			t.Fatalf("destroy Immerwolf: %v", err)
		}
	})
	effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
	if effectstest.IwFace(g, terror) != 0 || effectstest.IwFace(g, outlaw) != 0 {
		t.Errorf("with Immerwolf gone both should turn back: faces %d and %d", effectstest.IwFace(g, terror), effectstest.IwFace(g, outlaw))
	}
}

// Tovolar, the Midnight Scourge is a nightbound Werewolf and not a
// Human, so under Immerwolf day does not turn it back.
func TestImmerwolfKeepsTovolarAtNightByDay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	id := effectstest.ImportAndCast(t, g, effectstest.DnbWerewolf(effectstest.DnbTovolar).Row(), me)
	effectstest.MonSettle(t, g)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("Tovolar cast at night is on face %d, want the Midnight Scourge", effectstest.IwFace(g, id))
	}
	effectstest.B12Push(g, me.ID, "Immerwolf", "Creature — Wolf", effectstest.IwImmerwolfOracle, 2, 2)
	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	if effectstest.IwFace(g, id) != 1 {
		t.Errorf("day turned Tovolar, the Midnight Scourge over under Immerwolf (face %d)", effectstest.IwFace(g, id))
	}
}

// --- the cards -----------------------------------------------------------

// Huntmaster: a Wolf and 2 life as it enters, the Ravager's two
// related targets as it turns over, and the Huntmaster's trigger again
// as it turns back.
func TestHuntmasterOfTheFellsAndRavager(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := effectstest.ImportAndCast(t, g, iwHuntmaster.Row(), me)
	effectstest.MonSettle(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf"); n != 1 || me.Life != life+2 {
		t.Fatalf("on entering: %d Wolves and life %d, want 1 and %d", n, me.Life, life+2)
	}

	nextSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	opp := g.Seats[nextSeat]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 4)
	elsewhere := effectstest.B12Creature(g, other.ID, "Another Bear", "Creature — Bear", 2, 4)
	oppLife := opp.Life

	effectstest.IwNextUpkeep(t, g, nil)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("Huntmaster did not turn over after a quiet turn (face %d)", effectstest.IwFace(g, id))
	}
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Ravager of the Fells asked for no target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("target the opponent: %v", err)
	}
	pick = effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the Ravager's creature")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: elsewhere, Slot: 1}}); err == nil {
		t.Fatal("a creature another player controls was accepted as the second target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 1}}); err != nil {
		t.Fatalf("target the opponent's creature: %v", err)
	}
	effectstest.MonSettle(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, oppLife-2)
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.DamageMarked != 2 {
		t.Errorf("their creature has %d damage, want 2", c.DamageMarked)
	}

	life = me.Life
	effectstest.IwNextUpkeep(t, g, map[int]int{nextSeat: 2})
	if effectstest.IwFace(g, id) != 0 {
		t.Fatalf("Ravager did not turn back after a two-spell turn (face %d)", effectstest.IwFace(g, id))
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf"); n != 2 || me.Life != life+2 {
		t.Errorf("turning back into Huntmaster: %d Wolves and life %d, want 2 and %d", n, me.Life, life+2)
	}
}

// The Ravager's second target is optional.
func TestRavagerOfTheFellsMayTargetNoCreature(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.IwPut(t, g, iwHuntmaster, me)
	oppLife := opp.Life
	effectstest.IwNextUpkeep(t, g, nil)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("no turn-over (face %d)", effectstest.IwFace(g, id))
	}
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("declining the creature: %v", err)
		}
	}
	effectstest.MonSettle(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, oppLife-2)
	}
}

// Werewolf Ransacker: the transform targets an artifact and asks; yes
// destroys it and burns its controller.
func TestWerewolfRansackerDestroysAnArtifactAndBurnsItsController(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	id := effectstest.IwPut(t, g, iwDeserter, me)
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	oppLife := opp.Life
	effectstest.IwNextUpkeep(t, g, nil)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("Afflicted Deserter did not turn over (face %d)", effectstest.IwFace(g, id))
	}
	effectstest.AnswerPickTarget(t, g, relic)
	effectstest.MonSettle(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.MonSettle(t, g)
	if _, ok := effectstest.BattlefieldCard(g, relic); ok {
		t.Fatal("the artifact was not destroyed")
	}
	if opp.Life != oppLife-3 {
		t.Errorf("the artifact's controller is at %d, want %d", opp.Life, oppLife-3)
	}
}

// Vildin-Pack Alpha asks about a Werewolf that can transform and says
// nothing about a daybound one.
func TestVildinPackAlphaMayTransformAnEnteringWerewolf(t *testing.T) {
	g, me := effectstest.IwTable(t)
	iwPutBack(t, g, iwBandit, me)
	outlaw := effectstest.ImportAndCast(t, g, iwOutlaw.Row(), me)
	effectstest.MonSettle(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.MonSettle(t, g)
	if effectstest.IwFace(g, outlaw) != 1 {
		t.Fatalf("the entering Kruin Outlaw was not transformed (face %d)", effectstest.IwFace(g, outlaw))
	}

	villager := effectstest.ImportAndCast(t, g, effectstest.FearfulVillagerRow(), me)
	effectstest.MonSettle(t, g)
	if c := effectstest.LatestConfirmFor(g, me.ID); c != nil {
		t.Errorf("a daybound Werewolf, which nothing but day and night may transform, raised a yes/no prompt %v", c.ID)
	}
	if effectstest.IwFace(g, villager) != 0 {
		t.Errorf("the daybound Werewolf turned over (face %d)", effectstest.IwFace(g, villager))
	}
}

// Sage of Ancient Lore counts your hand; Werewolf of Ancient Hunger
// counts every hand.
func TestSageOfAncientLoreCountsHands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.ImportAndCast(t, g, iwSage.Row(), me)
	effectstest.MonSettle(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != me.Hand.Size() || tough != me.Hand.Size() {
		t.Errorf("Sage is %d/%d with %d cards in hand", p, tough, me.Hand.Size())
	}
	effectstest.IwTransform(t, g, id)
	total := 0
	for _, p := range g.Seats {
		total += p.Hand.Size()
	}
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != total || tough != total {
		t.Errorf("Werewolf of Ancient Hunger is %d/%d with %d cards in all hands", p, tough, total)
	}
	abilities := effectstest.EffectiveAbilities(t, g, id)
	if !slices.Contains(abilities, "vigilance") || !slices.Contains(abilities, "trample") {
		t.Errorf("Werewolf of Ancient Hunger abilities = %v", abilities)
	}
}

func TestSageOfAncientLoreDrawsAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	effectstest.ImportAndCast(t, g, iwSage.Row(), me) // into the hand, then onto the stack
	effectstest.MonSettle(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand = %d, want %d: the entry trigger draws one", me.Hand.Size(), before+1)
	}
}

// Krallenhorde Howler discounts your creature spells only.
func TestKrallenhordeHowlerDiscountsCreatureSpells(t *testing.T) {
	g, me := effectstest.IwTable(t)
	iwPutBack(t, g, iwRecruiter, me)
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{2}{G}"); got != 2 {
		t.Errorf("a creature spell costs %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Growth", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("a sorcery costs %d, want 3", got)
	}
}

// Duskwatch Recruiter looks at three, takes a creature and has the
// player order the rest on the bottom.
func TestDuskwatchRecruiterTakesACreature(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, iwRecruiter, me)
	bear := uuid.New()
	ids := effectstest.SeedLibrary(me, "A", "B")
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 1 {
		t.Fatalf("no take-one prompt: %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("take the Bear: %v", err)
	}
	if !me.Hand.Contains(bear) {
		t.Fatal("the creature card is not in hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom {
		t.Fatalf("the rest are not ordered on the bottom: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[1], ids[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LibraryBottomIDs(me, 2); !effectstest.SameIDs(got, []uuid.UUID{ids[1], ids[0]}) {
		t.Errorf("bottom = %v, want the player's order", got)
	}
}

// Tovolar's Magehunter burns an opponent who casts, not its controller.
func TestTovolarsMagehunterBurnsACaster(t *testing.T) {
	g, me := effectstest.IwTable(t)
	iwPutBack(t, g, iwShaman, me)
	myLife := me.Life
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Errorf("the controller's own spell cost them %d life", myLife-me.Life)
	}
	seat := effectstest.IwNextUpkeep(t, g, nil)
	opp := g.Seats[seat]
	effectstest.AdvanceToMainOf(t, g, seat)
	oppLife := opp.Life
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the casting opponent is at %d, want %d", opp.Life, oppLife-2)
	}
}

// Mayor of Avabruck pumps your other Humans; Howlpack Alpha your other
// Wolves and Werewolves, and makes a Wolf at your end step.
func TestMayorOfAvabruckAndHowlpackAlpha(t *testing.T) {
	g, me := effectstest.IwTable(t)
	mayor := effectstest.IwPut(t, g, iwMayor, me)
	human := effectstest.B12Creature(g, me.ID, "Villager", "Creature — Human", 1, 1)
	if p, _ := effectstest.EffectivePTOf(t, g, human); p != 2 {
		t.Errorf("a Human under the Mayor has power %d, want 2", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, mayor); p != 1 {
		t.Errorf("the Mayor pumps itself: power %d", p)
	}
	effectstest.IwTransform(t, g, mayor)
	if p, _ := effectstest.EffectivePTOf(t, g, human); p != 1 {
		t.Errorf("the Human kept the Mayor's bonus after it turned over: power %d", p)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.MonSettle(t, g)
	n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf")
	if n != 1 {
		t.Fatalf("Howlpack Alpha made %d Wolves at the end step, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Wolf" {
			if p, _ := effectstest.EffectivePTOf(t, g, c.InstanceID); p != 3 {
				t.Errorf("the Wolf token is %d power under the Alpha, want 3", p)
			}
		}
	}
}

// Instigator Gang pumps your attackers.
func TestInstigatorGangPumpsAttackers(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.IwPut(t, g, iwGang, me)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Fatalf("a creature not attacking has power %d, want 2", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 3 {
		t.Errorf("an attacking Bear has power %d, want 3", p)
	}
}

// Daybreak Ranger shoots a flier; Nightfall Predator fights.
func TestDaybreakRangerAndNightfallPredator(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ranger := effectstest.IwPut(t, g, iwRanger, me)
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 3,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, ranger, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bird}},
	}); err != nil {
		t.Fatalf("activate the Ranger: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bird); c.DamageMarked != 2 {
		t.Errorf("the flier has %d damage, want 2", c.DamageMarked)
	}

	effectstest.IwTransform(t, g, ranger)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ranger {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, ranger, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate the Predator: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the Bear survived 4 damage")
	}
	if c, _ := effectstest.BattlefieldCard(g, ranger); c.DamageMarked != 2 {
		t.Errorf("Nightfall Predator has %d damage, want the Bear's 2", c.DamageMarked)
	}
}

// Scorned Villager taps for {G}; Moonscarred Werewolf for {G}{G}.
func TestScornedVillagerManaAbilities(t *testing.T) {
	front, _ := effects.Lookup(iwVillager.Oracle)
	back, _ := effects.Lookup(iwVillager.Oracle + "#1")
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
