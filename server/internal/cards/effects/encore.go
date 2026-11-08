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
//
// ONE creation, one group per opponent tagged with that opponent (#4867,
// game.TokenGroup.Tag): "for each opponent, create a token" is one
// instruction, so the tokens are created at the same time (CR 701.7a,
// CR 608.2f) and enter together, and each newcomer sees the others enter
// (CR 603.6a). A doubler doubles each opponent's group, and both of that
// opponent's tokens attack that opponent. A token a replacement adds
// (Xorn's Treasure) is untagged, so it is neither bound nor sacrificed.
// The binding and the sacrifice are the creation's keyed tail, told the
// tokens that landed, so a copy whose entry asks a question (a Clone's
// copy choice) is still bound.
func encoreTokens(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	tmpl, ok := TokenCopyTemplate(g, item.SourceCardID)
	if !ok {
		return nil
	}
	TokenCopyGainsHaste(&tmpl)
	var groups []game.TokenGroup
	for _, opp := range ctx.Opponents() {
		groups = append(groups, game.TokenGroup{Template: copiedToken(tmpl), Count: 1, Tag: opp})
	}
	if len(groups) == 0 {
		return nil
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller, Source: item.SourceCardID, Groups: groups,
	}, ctx.ItemTail(encoreBindTokens))
}

// encoreBindTokens is "that attacks that opponent this turn if able ...
// Sacrifice them at the beginning of the next end step" over the tokens
// the creation made, each bound to the opponent its group was tagged with.
// An on-disk key: never renamed, never reused.
var encoreBindTokens = CardTail("encore/bind-tokens", func(ctx *Context, o game.TailOutcome) error {
	tags := game.TokenTags(o, 0)
	var all []uuid.UUID
	for _, tok := range o.IDs {
		opp, ok := tags[tok]
		if !ok {
			continue
		}
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
	if len(all) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Encore — sacrifice the tokens",
		Cards: all,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
})
