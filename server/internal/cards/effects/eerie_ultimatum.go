package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eerie Ultimatum — Sorcery {W}{W}{B}{B}{B}{G}{G} (EDHREC rank
// 1534):
//
//	"Return any number of permanent cards with different names from
//	 your graveyard to the battlefield."
//
// The seven-mana mass reanimation. It does not target (#2524): the
// choice is made as the spell RESOLVES, through a choose-cards prompt
// over the permanent cards in the caster's graveyard at that moment,
// so a response cannot strand a card the way a target clause would,
// and targeting rules never come into it. The cards come back under
// their owner's control — the caster's — together, as one entry
// (#1867): each card's own enters-tapped clause runs, and every ETB
// trigger sees the whole batch (CR 603.6a).
//
// "With different names" is the prompt's Validate hook: a set holding
// two cards of one name is not an answer the engine accepts or the
// enumerator offers. The name compared is the card's current one
// (Effective), the reading EachDifferentName takes for the targeted
// reanimations.
//
// "Any number" includes none, so the spell can be cast with an empty
// graveyard and do nothing, and "choose nothing" is always a legal
// answer to the prompt.
func init() {
	Register(Spec{
		OracleID:     "5674f6ae-ed5d-441e-a534-b5dd415165fd",
		Name:         "Eerie Ultimatum",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			candidates := graveyardCardIDs(ctx, item.Controller, game.Card.IsPermanent)
			if len(candidates) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:    item.Controller,
				FromPlayer: item.Controller,
				Source:     item.SourceCardID,
				Question:   "Eerie Ultimatum — return any number of permanent cards with different names from your graveyard to the battlefield",
				Cards:      candidates,
				Min:        0,
				Max:        0,
				Zone:       game.ZoneGraveyard,
				Validate:   allDifferentNames,
				Then: func(g *game.Game, picked []uuid.UUID) error {
					return ReturnFromGraveyardTogether{Targets: picked}.Apply(NewContext(g, item))
				},
			})
			return nil
		},
	})
}

// allDifferentNames is "with different names" over a picked set: no
// two cards share a (case-folded) current name. A card with no name
// conflicts with nothing, the lenient reading EachDifferentName takes.
func allDifferentNames(picked []game.Card) bool {
	seen := make(map[string]bool, len(picked))
	for _, c := range picked {
		name := strings.ToLower(strings.TrimSpace(c.Effective().Name))
		if name == "" {
			continue
		}
		if seen[name] {
			return false
		}
		seen[name] = true
	}
	return true
}

// returnLegalGraveyardTargetsToBattlefield puts every still-legal
// graveyard target onto the battlefield under its owner's control, all
// of them as one entry (#1867, CR 603.6a) — the body of the set-rule
// reanimations (#1559): Agadeem's Awakening, Behold the Sinister Six!.
// The set rule itself is the clause's; by the time this runs the CR
// 608.2b re-check has already dropped any pick that breaks it.
func returnLegalGraveyardTargetsToBattlefield(ctx *Context) error {
	return ReturnFromGraveyardTogether{Targets: legalTargetCardIDs(ctx)}.Apply(ctx)
}
