package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// encore.go — encore (CR 702.141, #3172): "{cost}, Exile this card from
// your graveyard: For each opponent, create a token that's a copy of it
// that attacks that opponent this turn if able. They gain haste.
// Sacrifice them at the beginning of the next end step. Activate only as
// a sorcery."
//
// An activated ability that works from the graveyard and whose cost
// exiles its source, the same shape as embalm (embalm.go): the token
// copies are made from the card in exile, where the cost just put it.
// Each token is made for ONE opponent and carries the requirement to
// attack that opponent (game.ModAttackPlayerIfAble, layer 6, until end
// of turn) as the creature is declared, so a table of three opponents
// gets three tokens that each must go at their own player if able. Haste
// rides the token copy's keywords (TokenCopyGainsHaste). One delayed
// trigger sacrifices every token at the beginning of the next end step.
//
// The tokens are the card's copiable values (CR 707.2), so a copied
// creature's triggers, statics and abilities all work, and an opponent
// who has left the game (or cannot be attacked) gets no token.

// Encore is "Encore <cost>" — CR 702.141a.
//
//	Activated: []ActivatedAbility{Encore("{3}{R}")},
func Encore(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label: "Encore " + cost + " (" + cost + ", Exile this card from your graveyard: For each opponent, " +
			"create a token that's a copy of it that attacks that opponent this turn if able. They gain haste. " +
			"Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)",
		Cost:         Plus(ManaCost(cost), ExileThis()),
		Zones:        []game.ZoneKind{game.ZoneGraveyard},
		SorcerySpeed: true,
		Effect:       encoreTokens,
	}
}

// encoreTokens makes one hasty token copy per opponent, binds each to its
// opponent, and schedules their sacrifice. Package-level and
// capture-free, like every ability body.
func encoreTokens(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var all []uuid.UUID
	for _, opp := range ctx.Opponents() {
		cursor := b25LastEventSeq(g)
		if err := (CreateTokenCopy{
			Controller: item.Controller,
			Copy:       item.SourceCardID,
			N:          1,
			Except:     TokenCopyGainsHaste,
		}).Apply(ctx); err != nil {
			return err
		}
		for _, tok := range b27TokensCreatedByAfter(g, item.Controller, cursor) {
			if err := (ScopedEffectFor{
				Target:   tok,
				Mods:     []game.Mod{game.AttackPlayerIfAbleMod(opp)},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Encore — attacks that opponent this turn if able",
			}).Apply(ctx); err != nil {
				return err
			}
			all = append(all, tok)
		}
	}
	if len(all) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Encore — sacrifice the tokens",
		Cards: all,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
