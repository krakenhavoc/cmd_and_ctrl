package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b286_test.go — Shadows over Innistrad activated-ability cards,
// S82 OLD batch 286 (#5543), played through the real catalog entries.

const (
	oldB286Cathar     = "1eb89059-b691-431d-bcda-ab8aa14c67c8"
	oldB286Tormentor  = "78b86f23-650b-4347-917c-3ee8c6007e4e"
	oldB286Malignity  = "795a9f45-e4ee-457c-b260-17b9f09c3ba6"
	oldB286Gargoyle   = "9b5cf592-8385-4115-b27a-98f42bdf6329"
	oldB286Ephemera   = "bfddc488-ab38-4e6d-8dfc-22bf1a8e03fa"
	oldB286Heir       = "d5b9c62d-fe98-4cbc-b4b6-012702bf6657"
	oldB286Bloodsworn = "e2cfbaf5-903c-49fa-b7bf-2c55087f1ffa"
	oldB286Nascency   = "e33dec01-c114-436a-befc-dab4e806c10b"
	oldB286Constable  = "eb3f493a-a5d6-4b9b-9ea0-3c295a67f730"
)

var (
	oldB286TormentorCard = effectstest.IwCard{Oracle: oldB286Tormentor, Front: "Elusive Tormentor", FrontType: "Creature — Vampire Wizard", Cost: "{2}{B}{B}",
		FrontText: "{1}, Discard a card: Transform this creature.", Fp: "4", Ft: "4",
		Back: "Insidious Mist", BackText: "Hexproof, indestructible\nThis creature can't block and can't be blocked.\nWhenever this creature attacks and isn't blocked, you may pay {2}{B}. If you do, transform it.",
		Bp: "0", Bt: "1", Colors: []string{"B"}, Keywords: []string{"Hexproof", "Indestructible"}}
	oldB286GargoyleCard = effectstest.IwCard{Oracle: oldB286Gargoyle, Front: "Thraben Gargoyle", FrontType: "Artifact Creature — Gargoyle", Cost: "{1}",
		FrontText: "Defender\n{6}: Transform this creature.", Fp: "2", Ft: "2",
		Back: "Stonewing Antagonizer", BackText: "Flying", Bp: "4", Bt: "2", Keywords: []string{"Defender"}}
	oldB286HeirCard = effectstest.IwCard{Oracle: oldB286Heir, Front: "Heir of Falkenrath", FrontType: "Creature — Vampire", Cost: "{1}{B}",
		FrontText: "Discard a card: Transform this creature. Activate only once each turn.", Fp: "2", Ft: "1",
		Back: "Heir to the Night", BackText: "Flying", Bp: "3", Bt: "2", Colors: []string{"B"}, Keywords: []string{"Flying"}}
)

func TestOldB286EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB286Cathar: "Veteran Cathar", oldB286Tormentor: "Elusive Tormentor", oldB286Tormentor + "#1": "Insidious Mist",
		oldB286Malignity: "Vessel of Malignity", oldB286Gargoyle: "Thraben Gargoyle", oldB286Gargoyle + "#1": "Stonewing Antagonizer",
		oldB286Ephemera: "Vessel of Ephemera", oldB286Heir: "Heir of Falkenrath", oldB286Heir + "#1": "Heir to the Night",
		oldB286Bloodsworn: "Olivia's Bloodsworn", oldB286Nascency: "Vessel of Nascency", oldB286Constable: "Stern Constable",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB286VeteranCatharGivesAHumanDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cathar := effectstest.MirB1Push(g, me.ID, "Veteran Cathar", "Creature — Human Soldier", oldB286Cathar, 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.MirB1Activate(g, me.ID, cathar, 0, effectstest.CardTarget(human)...); err == nil {
		t.Fatal("activated without paying {3}{W}")
	}
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{W}")
	if err := effectstest.MirB1Activate(g, me.ID, cathar, 0, effectstest.CardTarget(bear)...); err == nil {
		t.Fatal("a non-Human was accepted as the target")
	}
	if err := effectstest.MirB1Activate(g, me.ID, cathar, 0, effectstest.CardTarget(human)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, human, "double strike") {
		t.Error("the Human did not gain double strike")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "double strike") {
		t.Error("the Bear gained double strike")
	}
}

func TestOldB286OliviasBloodswornCantBlockAndGivesAVampireHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sworn := effectstest.MirB1Push(g, me.ID, "Olivia's Bloodsworn", "Creature — Vampire Soldier", oldB286Bloodsworn, 2, 1, "flying")
	vamp := effectstest.B12Creature(g, me.ID, "Vampire", "Creature — Vampire", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AssertRestrictions(t, g, sworn, game.CantBlock)
	effectstest.DcMana(t, g, me.ID, "{R}")
	if err := effectstest.MirB1Activate(g, me.ID, sworn, 0, effectstest.CardTarget(bear)...); err == nil {
		t.Fatal("a non-Vampire was accepted as the target")
	}
	if err := effectstest.MirB1Activate(g, me.ID, sworn, 0, effectstest.CardTarget(vamp)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, vamp, "haste") {
		t.Error("the Vampire did not gain haste")
	}
}

func TestOldB286SternConstableTapsACreatureForTapAndADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	con := effectstest.MirB1Push(g, me.ID, "Stern Constable", "Creature — Human Soldier", oldB286Constable, 1, 1)
	foe := effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, con, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(foe)}); err == nil {
		t.Fatal("activated without discarding a card")
	}
	pitch := effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, con, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(foe), DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, c := effectstest.CardWhere(g, g.Seats[1], foe); !c.Tapped {
		t.Error("the target was not tapped")
	}
	if _, c := effectstest.CardWhere(g, me, con); !c.Tapped {
		t.Error("the Constable did not tap as a cost")
	}
	if me.Hand.Contains(pitch) {
		t.Error("the discarded card is still in hand")
	}
}

func TestOldB286VesselOfEphemeraSacrificesForTwoSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	v := effectstest.MirB1Push(g, me.ID, "Vessel of Ephemera", "Enchantment", oldB286Ephemera, 0, 0)
	if err := effectstest.MirB1Activate(g, me.ID, v, 0); err == nil {
		t.Fatal("activated without paying {2}{W}")
	}
	effectstest.DcMana(t, g, me.ID, "{C}{C}{W}")
	if err := effectstest.MirB1Activate(g, me.ID, v, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(v) {
		t.Error("the Vessel was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); got != 2 {
		t.Errorf("%d Spirit tokens, want 2", got)
	}
}

func TestOldB286VesselOfMalignityMakesAnOpponentExileTwoAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	keep := effectstest.HandCardFull(opp, "Keep", "Instant", "", "", nil)
	a := effectstest.HandCardFull(opp, "A", "Instant", "", "", nil)
	b := effectstest.HandCardFull(opp, "B", "Instant", "", "", nil)
	effectstest.AdvanceToMain(t, g)
	v := effectstest.MirB1Push(g, me.ID, "Vessel of Malignity", "Enchantment", oldB286Malignity, 0, 0)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := effectstest.MirB1Activate(g, me.ID, v, 0, target...); err == nil {
		t.Fatal("activated without paying {1}{B}")
	}
	effectstest.DcMana(t, g, me.ID, "{C}{B}")
	if err := effectstest.MirB1Activate(g, me.ID, v, 0, target...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(v) {
		t.Error("the Vessel was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, opp.ID)
	if c == nil || c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Fatalf("the opponent's prompt is %+v, want exactly two cards", c)
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) || !opp.Hand.Contains(keep) {
		t.Error("the two chosen cards were not the ones exiled")
	}
}

func TestOldB286VesselOfMalignityIsRefusedOutsideTheMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	v := effectstest.MirB1Push(g, me.ID, "Vessel of Malignity", "Enchantment", oldB286Malignity, 0, 0)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.DcMana(t, g, me.ID, "{C}{B}")
	if err := effectstest.MirB1Activate(g, me.ID, v, 0, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err == nil {
		t.Error("activated in the declare attackers step")
	}
}

func TestOldB286VesselOfNascencyTakesOneCardAndMillsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	v := effectstest.MirB1Push(g, me.ID, "Vessel of Nascency", "Enchantment", oldB286Nascency, 0, 0)
	instant := effectstest.TopOfLibrary(me, "Bolt", "Instant", "{R}")
	sorc := effectstest.TopOfLibrary(me, "Sorcery", "Sorcery", "{1}")
	land := effectstest.TopOfLibrary(me, "Forest", "Basic Land — Forest", "")
	bear := effectstest.TopOfLibrary(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.DcMana(t, g, me.ID, "{C}{G}")
	if err := effectstest.MirB1Activate(g, me.ID, v, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt for the take")
	}
	for _, bad := range []uuid.UUID{instant, sorc} {
		for _, id := range c.ChooseCards {
			if id == bad {
				t.Error("an instant or sorcery card was offered")
			}
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the chosen card is not in hand")
	}
	for _, id := range []uuid.UUID{instant, sorc, land} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %v was not put into the graveyard", id)
		}
	}
}

func TestOldB286HeirOfFalkenrathTransformsOncePerTurn(t *testing.T) {
	g, me := effectstest.IwTable(t)
	heir := effectstest.IwPut(t, g, oldB286HeirCard, me)
	if err := g.ActivateCatalogAbility(me.ID, heir, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without discarding a card")
	}
	pitch := effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, heir, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	pitch2 := effectstest.HandCardFull(me, "Fodder Two", "Sorcery", "{1}", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, heir, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch2}}); err == nil {
		t.Error("activated twice in one turn")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, heir) != 1 {
		t.Errorf("Heir of Falkenrath is on face %d, want Heir to the Night", effectstest.IwFace(g, heir))
	}
}

func TestOldB286ThrabenGargoyleTransformsForSix(t *testing.T) {
	g, me := effectstest.IwTable(t)
	gar := effectstest.IwPut(t, g, oldB286GargoyleCard, me)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, gar, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for five mana")
	}
	effectstest.DcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, gar, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, gar) != 1 {
		t.Errorf("Thraben Gargoyle is on face %d, want Stonewing Antagonizer", effectstest.IwFace(g, gar))
	}
}

func TestOldB286ElusiveTormentorTransformsAndTheMistPaysToTurnBack(t *testing.T) {
	g, me := effectstest.IwTable(t)
	tor := effectstest.IwPut(t, g, oldB286TormentorCard, me)
	effectstest.DcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, tor, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without discarding a card")
	}
	pitch := effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, tor, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, tor) != 1 {
		t.Fatalf("Elusive Tormentor is on face %d, want Insidious Mist", effectstest.IwFace(g, tor))
	}
	effectstest.AssertRestrictions(t, g, tor, game.CantBlock|game.CantBeBlocked)
	if !effectstest.HasKeywordOnBattlefield(t, g, tor, "hexproof") || !effectstest.HasKeywordOnBattlefield(t, g, tor, "indestructible") {
		t.Error("the Mist lacks hexproof or indestructible")
	}
}
