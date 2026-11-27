package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pcy_b673_cards_test.go — Prophecy, S82 OLD batch 673 (#5617): Avatar of Will,
// Windscouter, Troublesome Spirit, Bog Elemental, Rib Cage Spider, Search for
// Survivors, Psychic Theft and Denying Wind, each through the real cast,
// trigger and combat paths. (Flay was already on develop.)

const (
	pcy673Avatar     = "13aa88a0-4b47-4012-ba50-002e51e637cf"
	pcy673Windscout  = "58232c49-f880-41ef-b808-62a6e804c715"
	pcy673Spirit     = "601d9bc3-ca75-43de-beea-ab3f0a49f543"
	pcy673Bog        = "82344b5b-7cd9-460d-9966-f09432ed8e59"
	pcy673Spider     = "906cba93-3dac-4720-a482-987cf1b4e786"
	pcy673Survivors  = "02fd4da6-6c6e-4143-a915-495c668de148"
	pcy673Theft      = "66953dfd-d908-41d3-abb4-309c6a165dc3"
	pcy673DenyingWnd = "8daed4d8-1a43-4452-85dd-51b320714489"
)

func TestPcy673EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		pcy673Avatar: "Avatar of Will", pcy673Windscout: "Windscouter", pcy673Spirit: "Troublesome Spirit",
		pcy673Bog: "Bog Elemental", pcy673Spider: "Rib Cage Spider", pcy673Survivors: "Search for Survivors",
		pcy673Theft: "Psychic Theft", pcy673DenyingWnd: "Denying Wind",
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

func TestPcy673PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		keywords               []string
	}{
		{"Avatar of Will", pcy673Avatar, "Creature — Avatar", []string{"flying"}},
		{"Windscouter", pcy673Windscout, "Creature — Human Scout", []string{"flying"}},
		{"Troublesome Spirit", pcy673Spirit, "Creature — Spirit", []string{"flying"}},
		{"Bog Elemental", pcy673Bog, "Creature — Elemental", []string{"protection from white"}},
		{"Rib Cage Spider", pcy673Spider, "Creature — Spider", []string{"reach"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, false)
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s does not have %s", tc.name, kw)
				}
			}
		})
	}
}

// --- Avatar of Will ---

func pcy673CastAvatar(g *game.Game, me *game.Player) error {
	_, err := effectstest.CastFromHandErr(g, me, "Avatar of Will", "Creature — Avatar", "{6}{U}{U}", pcy673Avatar,
		game.CastSpellParams{Strict: true})
	return err
}

func TestPcy673AvatarOfWillCostsSixLessOnlyWhenAnOpponentHasNoCards(t *testing.T) {
	for _, empty := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		for i, p := range g.Seats {
			if p.ID == me.ID {
				continue
			}
			if empty && i == (g.Turn.ActiveSeat+1)%len(g.Seats) {
				effectstest.RevealHand(p)
				continue
			}
			effectstest.RevealHand(p, effectstest.RhForest())
		}
		effectstest.B06AddMana(me, "U", "U")
		err := pcy673CastAvatar(g, me)
		if empty && err != nil {
			t.Errorf("an opponent has no cards in hand, yet {U}{U} did not cast it: %v", err)
		}
		if !empty && err == nil {
			t.Error("every opponent has cards, yet {U}{U} cast a {6}{U}{U} spell")
		}
	}
}

func TestPcy673AvatarOfWillStillNeedsItsBlueMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	for _, p := range g.Seats {
		if p.ID != me.ID {
			effectstest.RevealHand(p)
		}
	}
	effectstest.B06AddMana(me, "R", "R")
	if err := pcy673CastAvatar(g, me); err == nil {
		t.Error("two red mana paid for {U}{U}")
	}
}

// --- Windscouter ---

func TestPcy673WindscouterReturnsToHandAtEndOfCombatAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scout := effectstest.PushCatalogPermanent(g, me.ID, "Windscouter", "Creature — Human Scout", pcy673Windscout, false)
	effectstest.DeclareAttack(t, g, opp.ID, scout)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scout) {
		t.Fatal("Windscouter left before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(scout) || !me.Hand.Contains(scout) {
		t.Errorf("Windscouter should be in its owner's hand (battlefield %v, hand %v)", g.Battlefield.Contains(scout), me.Hand.Contains(scout))
	}
}

func TestPcy673WindscouterReturnsToHandAtEndOfCombatAfterBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	scout := effectstest.B12Push(g, opp.ID, "Windscouter", "Creature — Human Scout", pcy673Windscout, 3, 3)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, scout)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scout) {
		t.Fatal("Windscouter left before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(scout) || !opp.Hand.Contains(scout) {
		t.Errorf("Windscouter should be in its owner's hand (battlefield %v, hand %v)", g.Battlefield.Contains(scout), opp.Hand.Contains(scout))
	}
}

func TestPcy673WindscouterStaysWhenItDoesNotAttackOrBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	scout := effectstest.B12Push(g, opp.ID, "Windscouter", "Creature — Human Scout", pcy673Windscout, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scout) {
		t.Error("Windscouter that neither attacked nor blocked left the battlefield")
	}
}

// --- Troublesome Spirit ---

func TestPcy673TroublesomeSpiritTapsOnlyYourLandsAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Troublesome Spirit", "Creature — Spirit", pcy673Spirit, false)
	mine := []uuid.UUID{
		effectstest.SeedLand(g, me.ID, "Island", "Basic Land — Island", ""),
		effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", ""),
	}
	theirs := effectstest.SeedLand(g, opp.ID, "Island", "Basic Land — Island", "")
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range mine {
		if !effectstest.IsTapped(g, id) {
			t.Errorf("land %s is untapped after your end step", id)
		}
	}
	if effectstest.IsTapped(g, theirs) {
		t.Error("an opponent's land was tapped")
	}
}

func TestPcy673TroublesomeSpiritDoesNothingOnAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, g.Seats[1].ID, "Troublesome Spirit", "Creature — Spirit", pcy673Spirit, 3, 4)
	land := effectstest.SeedLand(g, me.ID, "Island", "Basic Land — Island", "")
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, land) {
		t.Error("the Spirit's controller's trigger tapped a land on the other player's end step")
	}
}

// --- Bog Elemental ---

func TestPcy673BogElementalUpkeepSacrificesALandOrItself(t *testing.T) {
	t.Run("a land", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bog := effectstest.PushCatalogPermanent(g, me.ID, "Bog Elemental", "Creature — Elemental", pcy673Bog, false)
		land := effectstest.SeedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "")
		effectstest.BeginUpkeep(t, g, me.ID)
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
		effectstest.AnswerChooseCards(t, g, me.ID, land)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) || !g.Battlefield.Contains(bog) {
			t.Error("the land should be sacrificed and the Elemental kept")
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bog := effectstest.PushCatalogPermanent(g, me.ID, "Bog Elemental", "Creature — Elemental", pcy673Bog, false)
		land := effectstest.SeedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "")
		effectstest.BeginUpkeep(t, g, me.ID)
		effectstest.AnswerOptionPick(t, g, me.ID, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(bog) || !g.Battlefield.Contains(land) {
			t.Error("declining should sacrifice the Elemental and keep the land")
		}
	})
	t.Run("no land", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bog := effectstest.PushCatalogPermanent(g, me.ID, "Bog Elemental", "Creature — Elemental", pcy673Bog, false)
		effectstest.BeginUpkeep(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(bog) {
			t.Error("the Elemental stayed with no land to sacrifice")
		}
	})
	t.Run("an opponent's upkeep", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		bog := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Bog Elemental", "Creature — Elemental", pcy673Bog, false)
		effectstest.BeginUpkeep(t, g, g.Seats[1].ID)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(bog) {
			t.Error("the Elemental was sacrificed on an opponent's upkeep")
		}
	})
}

// --- Search for Survivors ---

func TestPcy673SearchForSurvivorsPutsARandomCreatureOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Graveyard.Cards = nil
	bear := effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.CastCatalogSpell(t, g, "Search for Survivors", "Sorcery", pcy673Survivors, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the only card in the graveyard, a creature, was not put onto the battlefield")
	}
}

func TestPcy673SearchForSurvivorsExilesANoncreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Graveyard.Cards = nil
	bolt := effectstest.GraveyardCardOf(g, me, "Bolt", "Instant", "{R}", "")
	effectstest.CastCatalogSpell(t, g, "Search for Survivors", "Sorcery", pcy673Survivors, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bolt) {
		t.Error("the only card in the graveyard, an instant, was not exiled")
	}
}

func TestPcy673SearchForSurvivorsTakesExactlyOneCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Graveyard.Cards = nil
	var ids []uuid.UUID
	for _, n := range []string{"A", "B", "C", "D"} {
		ids = append(ids, effectstest.GraveyardCardOf(g, me, n, "Sorcery", "{R}", ""))
	}
	effectstest.CastCatalogSpell(t, g, "Search for Survivors", "Sorcery", pcy673Survivors, nil)
	effectstest.PassPriorityAroundTable(t, g)
	exiled := 0
	for _, id := range ids {
		if g.Exile.Contains(id) {
			exiled++
		}
	}
	if exiled != 1 {
		t.Errorf("%d of four graveyard cards were exiled, want exactly one", exiled)
	}
}

func TestPcy673SearchForSurvivorsWithAnEmptyGraveyardDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Graveyard.Cards = nil
	effectstest.CastCatalogSpell(t, g, "Search for Survivors", "Sorcery", pcy673Survivors, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Exile.Cards) != 0 {
		t.Errorf("%d cards were exiled from an empty graveyard", len(g.Exile.Cards))
	}
}

// --- Psychic Theft ---

func pcy673CastTheft(t *testing.T, g *game.Game, opp *game.Player) {
	t.Helper()
	effectstest.CastAt(t, g, "Psychic Theft", "Sorcery", pcy673Theft, effectstest.PbPlayer(opp.ID))
}

func pcy673TakeBolt(t *testing.T, g *game.Game, me *game.Player, opp *game.Player) uuid.UUID {
	t.Helper()
	hand := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt())
	pcy673CastTheft(t, g, opp)
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, []uuid.UUID{hand[1]}) {
		t.Fatalf("options = %v, want the Bolt only", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	g.SettleResolution()
	if !g.Exile.Contains(hand[1]) {
		t.Fatal("the chosen card was not exiled")
	}
	return hand[1]
}

func TestPcy673PsychicTheftReturnsTheCardAtTheNextEndStepIfNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	bolt := pcy673TakeBolt(t, g, me, opp)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(bolt) || !opp.Hand.Contains(bolt) {
		t.Errorf("the card should be back in its owner's hand (exile %v, hand %v)", g.Exile.Contains(bolt), opp.Hand.Contains(bolt))
	}
}

func TestPcy673PsychicTheftLetsYouCastTheCardAndThenItIsNotReturned(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	bolt := pcy673TakeBolt(t, g, me, opp)
	g.Turn.Step = game.StepPrecombatMain
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "exile", Strict: true, Targets: effectstest.PlayerRefs(opp.ID)}); err != nil {
		t.Fatalf("cast the exiled Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(bolt) {
		t.Error("a card that was cast came back to its owner's hand")
	}
}

func TestPcy673PsychicTheftCannotTakeALandWhenTheHandHasNoInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	effectstest.RevealHand(opp, effectstest.RhForest())
	pcy673CastTheft(t, g, opp)
	effectstest.NoVariant(t, g)
	if len(g.Exile.Cards) != 0 {
		t.Error("something was exiled from a hand with no instant or sorcery")
	}
}

// --- Denying Wind ---

func TestPcy673DenyingWindExilesUpToSevenCardsFromTheTargetsLibrary(t *testing.T) {
	for _, take := range []int{7, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E", "F", "G", "H", "I")
		before := opp.Library.Size()
		effectstest.CastCatalogSpell(t, g, "Denying Wind", "Sorcery", pcy673DenyingWnd, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.SearchChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatal("the caster was not asked to search the library")
		}
		var ids []uuid.UUID
		for _, id := range pick.SearchCards {
			if len(ids) < take {
				ids = append(ids, id)
			}
		}
		if len(ids) != take {
			t.Fatalf("only %d cards were offered", len(ids))
		}
		effectstest.AnswerSearchByID(t, g, me.ID, ids...)
		effectstest.PassPriorityAroundTable(t, g)
		if got := opp.Library.Size(); got != before-take {
			t.Errorf("take %d: the library has %d cards, want %d", take, got, before-take)
		}
		for _, id := range ids {
			if !g.Exile.Contains(id) {
				t.Errorf("take %d: %s was not exiled", take, id)
			}
		}
	}
}
