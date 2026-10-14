package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestLilianaMinusFourSacrificesCommanderOnceBeforeCommanderChoice(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Turn.ActiveSeat
	owner := g.Seats[active]
	liliana := pushWalkerForTest(g, owner.ID, "Liliana, Dreadhorde General", wrLilianaDGOracle, 6)

	commander := game.NewCommander("Test Commander", owner.ID)
	commander.TypeLine = "Legendary Creature — Bear"
	commander.Power, commander.Toughness = 2, 2
	commander.Controller = owner.ID
	g.Battlefield.PushTop(commander)
	ordinary := pushVanillaCreature(g, owner.ID, "Fodder A", 1, 1)
	unselected := pushVanillaCreature(g, owner.ID, "Fodder B", 1, 1)

	advanceToMainOf(t, g, active)
	if err := g.ActivateCatalogAbility(owner.ID, liliana, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Liliana −4: %v", err)
	}
	passPriorityAroundTable(t, g)

	prompt := latestChooseCardsFor(g, owner.ID)
	if prompt == nil || prompt.ChooseMin != 2 || prompt.ChooseMax != 2 {
		t.Fatalf("Liliana choice = %+v, want a choice of exactly two", prompt)
	}
	answerChooseCards(t, g, owner.ID, commander.InstanceID, ordinary)

	if g.Battlefield.Contains(commander.InstanceID) || g.Battlefield.Contains(ordinary) {
		t.Fatal("chosen permanents did not leave the battlefield before the commander choice")
	}
	if !owner.Graveyard.Contains(commander.InstanceID) || !owner.Graveyard.Contains(ordinary) {
		t.Fatal("chosen permanents were not sacrificed into the graveyard")
	}
	if !g.Battlefield.Contains(unselected) {
		t.Fatal("the unselected creature was sacrificed")
	}
	var commanderChoices []*game.PendingChoice
	for _, choice := range g.PendingChoices {
		if choice != nil && choice.Kind == game.PendingChoiceCommanderReturn {
			commanderChoices = append(commanderChoices, choice)
		}
	}
	if len(commanderChoices) != 1 || commanderChoices[0].Chooser != owner.ID || commanderChoices[0].Source != commander.InstanceID {
		t.Fatalf("commander-return choices = %+v, want one prompt for the sacrificed commander", commanderChoices)
	}
	if got := lilianaPendingTriggersFrom(g, liliana); got != 2 {
		t.Fatalf("Liliana death triggers while commander choice is open = %d, want 2", got)
	}

	var sacrifices, deaths int
	deathBatches := map[uuid.UUID]uint64{}
	for _, event := range g.EventsThisTurn() {
		if event.CardID != commander.InstanceID && event.CardID != ordinary {
			continue
		}
		if event.Kind == game.EventSacrifice {
			sacrifices++
		}
		if event.Kind == game.EventLTB && event.NewZone == game.ZoneGraveyard {
			deaths++
			deathBatches[event.CardID] = event.Batch
		}
	}
	if sacrifices != 2 || deaths != 2 {
		t.Fatalf("before commander answer: sacrifice events=%d, deaths=%d; want 2 each", sacrifices, deaths)
	}
	if deathBatches[commander.InstanceID] == 0 || deathBatches[commander.InstanceID] != deathBatches[ordinary] {
		t.Fatalf("commander and ordinary creature death batches = %d and %d, want the same nonzero simultaneous batch", deathBatches[commander.InstanceID], deathBatches[ordinary])
	}

	if err := g.ResolveCommanderReturn(commanderChoices[0].ID, owner.ID, true); err != nil {
		t.Fatalf("resolve commander-return choice: %v", err)
	}
	if !owner.Command.Contains(commander.InstanceID) || owner.Graveyard.Contains(commander.InstanceID) {
		t.Fatal("commander's answer did not move it from graveyard to command zone")
	}
	if got := lilianaPendingTriggersFrom(g, liliana); got != 2 {
		t.Fatalf("Liliana death triggers after commander answer = %d, want exactly 2", got)
	}
	if len(commanderReturnChoicesForTest(g)) != 0 {
		t.Fatal("commander-return prompt remained or was reopened after its answer")
	}
	var commanderDeaths, commanderZoneMoves int
	for _, event := range g.EventsThisTurn() {
		if event.CardID != commander.InstanceID {
			continue
		}
		if event.Kind == game.EventLTB && event.NewZone == game.ZoneGraveyard {
			commanderDeaths++
		}
		if event.Kind == game.EventZoneMove && event.OldZone == game.ZoneGraveyard && event.NewZone == game.ZoneCommand {
			commanderZoneMoves++
		}
	}
	if commanderDeaths != 1 || commanderZoneMoves != 1 {
		t.Fatalf("commander events: deaths=%d, graveyard-to-command=%d; want one each", commanderDeaths, commanderZoneMoves)
	}
}

func TestSacrificeAllContinuationWaitsForReplacementOrder(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	creature := pushVanillaCreature(g, owner.ID, "Replacement Test", 2, 2)
	var firstCalls, secondCalls int
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(event *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return event.Kind == game.RepEventMove && event.OldZone == game.ZoneBattlefield
			},
			Replace: func(event *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				firstCalls++
				event.NewZone = game.ZoneExile
				return nil
			},
			Label: "Redirect the sacrifice move",
		})
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(event *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return event.Kind == game.RepEventMove && event.OldZone == game.ZoneBattlefield
			},
			Replace: func(_ *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				secondCalls++
				return nil
			},
			Label: "Observe the redirected move",
		})
	})

	var tailCalls int
	var sacrificed []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ChoosePermanentsInOrderThenForEffect([]game.ChooseCardsPrompt{{
			Chooser: owner.ID, FromPlayer: owner.ID, Cards: []uuid.UUID{creature},
			Question: "Choose a permanent", Min: 0, Max: 1, Zone: game.ZoneBattlefield,
		}}, func(g *game.Game, picks game.PromptedPicks) error {
			if len(picks.Cards()) == 0 {
				return nil
			}
			return g.SacrificeAllThenForEffect(uuid.Nil, picks.Cards(), func(_ *game.Game, ids []uuid.UUID) error {
				tailCalls++
				sacrificed = append([]uuid.UUID(nil), ids...)
				return nil
			})
		}); err != nil {
			t.Fatalf("ChoosePermanentsInOrderThenForEffect: %v", err)
		}
	})
	selection := latestChooseCardsFor(g, owner.ID)
	if selection == nil || g.Battlefield.Contains(creature) == false || tailCalls != 0 || firstCalls != 0 || secondCalls != 0 {
		t.Fatalf("before permanent selection: prompt=%+v, on battlefield=%v, tail=%d, replacements=%d/%d; want an open selection and no continuation", selection, g.Battlefield.Contains(creature), tailCalls, firstCalls, secondCalls)
	}
	answerChooseCards(t, g, owner.ID, creature)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder || len(g.PendingChoices[0].ReplacementEffectIDs) != 2 {
		t.Fatalf("pending choices after permanent selection = %+v, want one CR 616 ordering prompt for both replacements", g.PendingChoices)
	}
	if !g.Battlefield.Contains(creature) || tailCalls != 0 || firstCalls != 0 || secondCalls != 0 {
		t.Fatalf("while replacement order is open: on battlefield=%v, tail=%d, replacements=%d/%d; want unchanged and no continuation", g.Battlefield.Contains(creature), tailCalls, firstCalls, secondCalls)
	}
	choice := g.PendingChoices[0]
	if err := g.ResolveReplacementOrder(choice.ID, choice.Chooser, append([]game.ReplacementEffectID(nil), choice.ReplacementEffectIDs...)); err != nil {
		t.Fatalf("resolve replacement order: %v", err)
	}
	if tailCalls != 1 || len(sacrificed) != 1 || sacrificed[0] != creature {
		t.Fatalf("continuation calls=%d, sacrificed=%v; want one continuation with the one moved permanent", tailCalls, sacrificed)
	}
	if firstCalls != 1 || secondCalls != 1 {
		t.Fatalf("replacement calls=%d/%d, want one each", firstCalls, secondCalls)
	}
	if g.Battlefield.Contains(creature) || !g.Exile.Contains(creature) {
		t.Fatal("replacement-resumed sacrifice did not finish in exile")
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d pending choices after replacement resolution, want none", len(g.PendingChoices))
	}
}

func commanderReturnChoicesForTest(g *game.Game) []*game.PendingChoice {
	var out []*game.PendingChoice
	for _, choice := range g.PendingChoices {
		if choice != nil && choice.Kind == game.PendingChoiceCommanderReturn {
			out = append(out, choice)
		}
	}
	return out
}
