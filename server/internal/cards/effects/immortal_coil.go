package effects

import (
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Immortal Coil — Artifact {2}{B}{B}:
//
//	"{T}, Exile two cards from your graveyard: Draw a card.
//	 If damage would be dealt to you, prevent that damage. Exile a card from your graveyard for each 1 damage prevented this way.
//	 When there are no cards in your graveyard, you lose the game."
//
// ADR 0108 §8 (#1906): a prevention static with an additional effect, one
// application per recipient (you) in a damage instance, counting the
// damage PREVENTED this way — so damage that can't be prevented is dealt
// and exiles nothing (CR 615.12). All the damage is prevented however few
// cards are left (the ruling): when the count covers the graveyard, all
// of it is exiled at once; otherwise its controller chooses which cards.
// The last line is a CR 603.8 state trigger (ADR 0107 PR 1): checked only
// as it triggers, and its controller loses as it resolves (the rulings).
//
// Simplification: when there is a choice to make, it is put to the
// player as a prompt, which the engine answers after the spell or ability
// that dealt the damage has finished, rather than immediately (CR 615.5).
// The candidates are the cards in the graveyard when the damage was
// prevented, and a pick that has left by the answer is made up from the
// rest of them, so the count is never lower than printed.
func init() {
	Register(Spec{
		OracleID:     "c85fc672-12d8-4ab0-b18d-ccd590ac063c",
		Name:         "Immortal Coil",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"When you must choose which cards to exile, you choose after the spell or ability that dealt the damage has finished.",
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Exile two cards from your graveyard: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), ExileFromGraveyard(2, "two cards", nil)),
			Effect:  Do(DrawCards{N: 1}),
		}},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToYou,
				Then:  immortalCoilExileBody,
				Label: "Immortal Coil — prevent the damage and exile a card from your graveyard for each 1 prevented",
			}),
		},
		Triggered: []game.TriggeredAbility{
			WhenState("Immortal Coil — you lose the game", immortalCoilGraveyardEmpty, Do(LoseTheGame{})),
		},
	})
}

// immortalCoilGraveyardEmpty is "When there are no cards in your
// graveyard".
func immortalCoilGraveyardEmpty(g *game.Game, _ *game.Card, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Graveyard != nil && p.Graveyard.Size() == 0
}

// The additional effect: exile a card from your graveyard for each 1
// damage prevented this way.
var immortalCoilExileBody = game.DelayedBody("immortal-coil/exile-per-prevented", immortalCoilExile)

func immortalCoilExile(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	n := p.Amount
	you := g.PlayerByIDForEffect(item.Controller)
	if n <= 0 || you == nil || you.Graveyard == nil || you.Graveyard.Size() == 0 {
		return nil
	}
	cards := make([]uuid.UUID, 0, you.Graveyard.Size())
	for _, c := range you.Graveyard.Cards {
		cards = append(cards, c.InstanceID)
	}
	if n >= len(cards) {
		g.ExileCardsForEffect(cards)
		return nil
	}
	owner := you.ID
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  owner,
		Source:   item.SourceCardID,
		Question: fmt.Sprintf("Immortal Coil — exile %d cards from your graveyard", n),
		Cards:    cards,
		Min:      n,
		Max:      n,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			return immortalCoilExilePicked(g, owner, cards, picked, n)
		},
	})
	return nil
}

// immortalCoilExilePicked exiles the picks still in the owner's
// graveyard, and makes any shortfall up from the other candidates still
// there, in graveyard order: a card that left before the answer never
// lowers the count.
func immortalCoilExilePicked(g *game.Game, owner uuid.UUID, candidates, picked []uuid.UUID, n int) error {
	inGraveyard := func(id uuid.UUID) bool {
		z := g.FindCardZoneForEffect(id)
		return z != nil && z.Kind == game.ZoneGraveyard && z.Owner == owner
	}
	var exile []uuid.UUID
	for _, id := range picked {
		if len(exile) < n && inGraveyard(id) && !slices.Contains(exile, id) {
			exile = append(exile, id)
		}
	}
	for _, id := range candidates {
		if len(exile) < n && inGraveyard(id) && !slices.Contains(exile, id) {
			exile = append(exile, id)
		}
	}
	g.ExileCardsForEffect(exile)
	return nil
}
