package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dauthi Voidwalker — Creature — Dauthi Rogue {B}{B}, 3/2:
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 If a card would be put into an opponent's graveyard from anywhere,
//	 instead exile it with a void counter on it.
//	 {T}, Sacrifice this creature: Choose an exiled card an opponent owns
//	 with a void counter on it. You may play it this turn without paying
//	 its mana cost."
//
// The replacement is GraveyardBecomesExile (the family comment names this
// card), with ExileCounter "void": the counter is stamped on the card
// once it has landed in exile, so a move a later replacement redirected
// (a commander to the command zone) is not marked. "From anywhere"
// covers a dying creature (it never died, CR 700.4), a discard, a mill
// and a countered spell. Tokens are not cards and are not touched. The
// graveyard is always the card's owner's, so a creature you stole and
// lost goes to its owner's graveyard and is exiled if that owner is an
// opponent of the Voidwalker's controller.
//
// The ability chooses at resolution, not as a target (the card says
// "choose"), so hexproof does not matter and the card can be any
// opponent's void-countered exile, including one exiled by an earlier
// Voidwalker. Sacrificing is a cost, so the Voidwalker is gone before
// the choice. The play permission is for this turn and has no mana cost
// ("{0}", which also fixes X at 0), and is not cast-only, so a land can
// be played from exile through it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f1c2dbe2-fbe0-4058-bdf1-91d1b1832786",
		Name:            "Dauthi Voidwalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Replacements: []game.ReplacementEffect{GraveyardBecomesExile{
			OpponentsOnly: true,
			ExileCounter:  dauthiVoidCounter,
			Label:         "Dauthi Voidwalker: exile with a void counter instead of an opponent's graveyard",
		}.Build()},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Choose an exiled card an opponent owns with a void counter on it. You may play it this turn without paying its mana cost.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  dauthiVoidwalkerPlay,
		}},
	})
}

const dauthiVoidCounter = "void"

func dauthiVoidwalkerPlay(g *game.Game, item *game.StackItem) error {
	you := item.Controller
	var ids []uuid.UUID
	if g.Exile != nil {
		for _, c := range g.Exile.Cards {
			if c.Counters[dauthiVoidCounter] > 0 && c.Owner != you && g.PlayerByIDForEffect(c.Owner) != nil {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  you,
		Source:   item.SourceCardID,
		Question: "Dauthi Voidwalker — choose an exiled card with a void counter to play this turn without paying its mana cost",
		Cards:    ids,
		Min:      1,
		Max:      1,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for _, id := range picked {
				z := g.FindCardZoneForEffect(id)
				c, ok := g.LookupCardForEffect(id)
				if !ok || z == nil || z.Kind != game.ZoneExile {
					continue
				}
				g.GrantCastPermissionToCardsForEffect(game.CastPermission{
					Player:     you,
					Zone:       game.ZoneExile,
					Duration:   g.UntilEndOfTurnDuration(),
					Cost:       "{0}",
					Source:     item.SourceCardID,
					SourceName: "Dauthi Voidwalker",
					Label:      "Dauthi Voidwalker — play it without paying its mana cost",
				}, []game.Card{c})
			}
			return nil
		},
	})
	return nil
}
