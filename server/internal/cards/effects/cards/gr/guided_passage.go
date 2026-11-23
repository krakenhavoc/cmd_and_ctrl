package cardsgr

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guided Passage — Sorcery {G}{U}{R} (Double Masters 2022, #5948):
//
//	"Reveal the cards in your library. An opponent chooses from among them a
//	 creature card, a land card, and a noncreature, nonland card. You put the
//	 chosen cards into your hand. Then shuffle."
//
// The whole library is revealed, then one opponent (you name which when there
// are several) chooses a card for each of the three categories in printed
// order; a category with no card in the library is skipped. The three
// categories are disjoint, so the picks are three different cards. Nothing
// moves until all three are chosen, then the chosen cards go into your hand
// together and you shuffle (also when nothing was chosen).
//
// No simplification.
func init() {
	registerGuidedPassagePicked()
	effects.Register(effects.Spec{
		OracleID:     "f4b30f6c-09ac-407e-9c72-da64d212ce93",
		Name:         "Guided Passage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			g, you := ctx.Game, ctx.Controller()
			if p := g.PlayerByIDForEffect(you); p != nil && p.Library != nil {
				g.RevealTopOfLibraryForEffect(you, ctx.Source(), p.Library.Size(), guidedPassageName)
			}
			opps := ctx.Opponents()
			switch len(opps) {
			case 0:
				return g.ShuffleLibraryForEffect(you)
			case 1:
				return guidedPassageAsk(g, you, ctx.Source(), opps[0], 0, nil)
			}
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: guidedPassageName + " — choose an opponent to choose from your library",
				ThenKey:  guidedPassageOpponent,
			}.Apply(ctx)
		},
	})
}

const guidedPassageName = "Guided Passage"

// guidedPassageKinds are the three categories in printed order.
var guidedPassageKinds = []struct {
	label string
	match func(game.Card) bool
}{
	{"a creature card", func(c game.Card) bool { return c.IsCreature() }},
	{"a land card", func(c game.Card) bool { return c.IsLand() }},
	{"a noncreature, nonland card", func(c game.Card) bool { return !c.IsCreature() && !c.IsLand() }},
}

// guidedPassageOpponent runs once the controller has named the opponent. An
// on-disk key: never renamed, never reused.
var guidedPassageOpponent = effects.ChoosePlayerThen("option-pick/guided-passage-opponent", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return guidedPassageAsk(ctx.Game, ctx.Controller(), ctx.Source(), chosen, 0, nil)
})

// guidedPassageAsk puts the prompt for the first category from `stage` on that
// has a card in the library to the opponent; with none left it finishes.
func guidedPassageAsk(g *game.Game, you, source, chooser uuid.UUID, stage int, picked []uuid.UUID) error {
	p := g.PlayerByIDForEffect(you)
	for ; stage < len(guidedPassageKinds) && p != nil && p.Library != nil; stage++ {
		var cands []uuid.UUID
		for _, c := range p.Library.Cards {
			if guidedPassageKinds[stage].match(c) {
				cands = append(cands, c.InstanceID)
			}
		}
		if len(cands) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    chooser,
			FromPlayer: you,
			Source:     source,
			Question:   guidedPassageName + " — choose " + guidedPassageKinds[stage].label + " from the revealed library",
			Cards:      cands,
			Min:        1,
			Max:        1,
			ThenKey:    guidedPassagePicked,
			LiveZone:   game.ZoneLibrary,
			Carry:      picked,
			CarryWord:  strconv.Itoa(stage),
		})
		return nil
	}
	if len(picked) == 0 {
		return g.ShuffleLibraryForEffect(you)
	}
	return g.TakeFromLibraryToHandKeyedForEffect(you, picked, game.KeyedTail{Ref: guidedPassageShuffle, Player: you, Source: source})
}

// guidedPassagePicked is the keyed answer: it records the pick and asks about
// the next category. An on-disk key: never renamed, never reused.
var guidedPassagePicked game.ChooseCardsThen

func registerGuidedPassagePicked() {
	guidedPassagePicked = game.RegisterChooseCardsThen("choose-cards/guided-passage-picked", func(g *game.Game, r game.CardsChosen) error {
		stage, _ := strconv.Atoi(r.Word)
		picked := append(append([]uuid.UUID{}, r.Carry...), r.Picked...)
		return guidedPassageAsk(g, r.From, r.Source, r.Chooser, stage+1, picked)
	})
}

// guidedPassageShuffle is "then shuffle", after the chosen cards landed. An
// on-disk key: never renamed, never reused.
var guidedPassageShuffle = game.RegisterTail("then/guided-passage-shuffle", func(g *game.Game, o game.TailOutcome) error {
	return g.ShuffleLibraryForEffect(o.Tail.Player)
})
