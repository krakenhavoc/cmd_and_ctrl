package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dubious Challenge — Sorcery {3}{G} (Kaladesh, #5539):
//
//	"Look at the top ten cards of your library, exile up to two creature cards
//	 from among them, then shuffle. Target opponent may choose one of the exiled
//	 cards and put it onto the battlefield under their control. Put the rest
//	 onto the battlefield under your control."
//
// Zimone's Experiment's private look and one choose_cards prompt, over the
// creature cards among the ten with a floor of zero ("up to two"); the picks
// are exiled, then the library is shuffled (also with nothing exiled). The
// targeted opponent is then asked, by a second prompt, for none or one of the
// exiled cards: that card enters under their control, and every other exiled
// card under the caster's, each group as one entry (a card that left exile in
// the meantime is skipped). An opponent who cannot answer chooses none. With
// its one target illegal the spell does nothing at all (CR 608.2b). Both
// prompts carry their players as data, so a table waiting on either is a
// restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f73bf7c5-e43d-4684-8588-2fb400f2ca31",
		Name:         "Dubious Challenge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			opp := effects.FirstLegalTargetPlayer(ctx)
			if opp == uuid.Nil {
				return nil
			}
			g := ctx.Game
			player := ctx.Controller()
			looked := g.LookAtTopOfLibraryForEffect(player, 10)
			var creatures []uuid.UUID
			for _, id := range looked {
				if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
					creatures = append(creatures, id)
				}
			}
			if len(creatures) == 0 {
				return g.ShuffleLibraryForEffect(player)
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  player,
				Source:   ctx.Source(),
				Question: "Dubious Challenge — exile up to two creature cards from among them",
				Cards:    creatures,
				Min:      0,
				Max:      2,
				LiveZone: game.ZoneLibrary,
				ThenKey:  dubiousChallengeExiled,
				Carry:    []uuid.UUID{opp},
			})
			return nil
		},
	})
}

// dubiousChallengeExiled is the first answer: the chosen creature cards are
// exiled, the library is shuffled, and the target opponent is asked which one
// they take. An on-disk key: never renamed, never reused.
var dubiousChallengeExiled = game.RegisterChooseCardsThen("choose-cards/dubious-challenge-exile", func(g *game.Game, r game.CardsChosen) error {
	player := r.Chooser
	picked := effects.CardsStillInALibrary(g, r.Picked)
	g.ExileCardsForEffect(picked)
	if err := g.ShuffleLibraryForEffect(player); err != nil {
		return err
	}
	var exiled []uuid.UUID
	for _, id := range picked {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			exiled = append(exiled, id)
		}
	}
	if len(exiled) == 0 || len(r.Carry) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    r.Carry[0],
		FromPlayer: player,
		Source:     r.Source,
		Question:   "Dubious Challenge — you may choose one of the exiled cards and put it onto the battlefield under your control",
		Cards:      exiled,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneExile,
		ThenKey:    dubiousChallengeChosen,
		Carry:      []uuid.UUID{player},
	})
	return nil
})

// dubiousChallengeChosen is the second answer: the opponent's pick enters under
// their control and the rest under the caster's. An on-disk key.
var dubiousChallengeChosen = game.RegisterChooseCardsThen("choose-cards/dubious-challenge-choose", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	caster := r.Carry[0]
	var mine, theirs []uuid.UUID
	for _, id := range r.Candidates {
		taken := false
		for _, p := range r.Picked {
			taken = taken || p == id
		}
		if taken {
			theirs = append(theirs, id)
		} else {
			mine = append(mine, id)
		}
	}
	if len(theirs) > 1 {
		theirs = theirs[:1]
	}
	if len(theirs) > 0 {
		if err := (effects.ReturnFromExileTogether{Targets: theirs, Controller: r.Chooser}).Apply(effects.KeyedContext(g, r.Chooser, r.Source)); err != nil {
			return err
		}
	}
	if len(mine) == 0 {
		return nil
	}
	return effects.ReturnFromExileTogether{Targets: mine, Controller: caster}.Apply(effects.KeyedContext(g, caster, r.Source))
})
