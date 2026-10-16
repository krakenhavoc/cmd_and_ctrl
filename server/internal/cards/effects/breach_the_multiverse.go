package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breach the Multiverse — Sorcery {5}{B}{B}:
//
//	"Each player mills ten cards. For each player, choose a creature or
//	 planeswalker card in that player's graveyard. Put those cards onto
//	 the battlefield under your control. Then each creature you control
//	 becomes a Phyrexian in addition to its other types."
//
// Exhume's walk (b38ReanimateAskNext) with the caster choosing every
// card: one pick per player in turn order, skipping a graveyard with
// nothing to choose, then ONE simultaneous entry under the caster's
// control (ReturnFromGraveyardTogether), so each newcomer sees the others
// enter. The cards are chosen, not targeted, so a milled card is a
// legal pick. "Becomes a Phyrexian" is a layer-4 subtype addition with
// no duration (CR 611.2a), pinned to each creature the caster controls
// once the entry has finished; a creature that leaves loses it, and one
// that enters later never had it.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:     "ebbd5b9c-07ff-4b42-81c2-ba09539f3a42",
		Name:         "Breach the Multiverse",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			order := apnapPlayers(ctx.Game)
			for _, p := range order {
				if err := (MillCards{Player: p, N: 10}).Apply(ctx); err != nil {
					return err
				}
			}
			return breachAskNext(ctx.Game, item, order, nil)
		},
	})
}

func breachCreatureOrPlaneswalker(c game.Card) bool {
	return c.IsCreature() || c.IsPlaneswalker()
}

// breachAskNext asks the caster about the next player whose graveyard
// holds a creature or planeswalker card, carrying the picks so far by
// value, and puts them all onto the battlefield once nobody is left.
func breachAskNext(g *game.Game, item *game.StackItem, order, chosen []uuid.UUID) error {
	ctx := NewContext(g, item)
	for i, owner := range order {
		candidates := graveyardCardIDs(ctx, owner, breachCreatureOrPlaneswalker)
		if len(candidates) == 0 {
			continue
		}
		rest := order[i+1:]
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    item.Controller,
			FromPlayer: owner,
			Source:     item.SourceCardID,
			Question:   "Breach the Multiverse — choose a creature or planeswalker card in this graveyard",
			Cards:      candidates,
			Min:        1,
			Max:        1,
			Zone:       game.ZoneGraveyard,
			Then: func(g *game.Game, picked []uuid.UUID) error {
				next := append(append([]uuid.UUID(nil), chosen...), picked...)
				return breachAskNext(g, item, rest, next)
			},
		})
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:    chosen,
		Controller: item.Controller,
		Then:       breachPhyrexians,
	}.Apply(ctx)
}

// breachPhyrexians makes each creature the caster controls a Phyrexian
// in addition to its other types.
func breachPhyrexians(ctx *Context, _ []uuid.UUID) error {
	caster := ctx.Controller()
	// "Phyrexian" is a word of the spell's text (CR 612.2, #3984).
	phyrexian := ctx.TextSubtypes("Phyrexian")
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != caster || !c.IsCreature() {
			continue
		}
		if err := (ScopedEffectFor{
			Target:   c.InstanceID,
			Mods:     []game.Mod{game.AddSubtypesMod(phyrexian...)},
			Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), c.InstanceID),
			Label:    "Breach the Multiverse — a Phyrexian in addition to its other types",
		}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
