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
		OnResolve: ResolveSteps(
			// The picks wait for a prompt the mills ask (CR 608.2c, #4155).
			// The mills are one library at a time in APNAP order, so a
			// prompt one asks is the only one open (#4266); the picks are
			// parked behind the last of them.
			Instruction(func(ctx *Context) error {
				return millEachPlayerInTurn(ctx, apnapPlayers(ctx.Game), constantMill(10), game.KeyedTail{})
			}),
			Instruction(func(ctx *Context) error {
				return breachAskNext(ctx.Game, ctx.Item, apnapPlayers(ctx.Game), nil)
			}),
		),
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
			ThenKey:    breachPick,
			LiveZone:   game.ZoneGraveyard,
			// The players still to ask and the cards chosen so far are the
			// carry (#4253).
			Carry: packIDLists(rest, chosen),
			Asked: game.NewAskedItem(item),
		})
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:    chosen,
		Controller: item.Controller,
		ThenTail:   breachPhyrexiansTail,
	}.Apply(ctx)
}

// breachPick is one graveyard's pick answered, keyed (#4253): the next
// graveyard is asked, or the chosen cards enter together. Registered in init
// because it asks the next graveyard, which names the key again.
var breachPick game.ChooseCardsThen

func init() {
	breachPick = askedPick("choose-cards/breach-the-multiverse-graveyard", func(ctx *Context, r game.CardsChosen) error {
		rest, chosen := unpackIDPair(r.Carry)
		next := append(append([]uuid.UUID(nil), chosen...), r.Picked...)
		return breachAskNext(ctx.Game, ctx.Item, rest, next)
	})
}

// breachPhyrexiansTail is breachPhyrexians as a keyed tail (#4253 slice 6).
// An on-disk key.
var breachPhyrexiansTail = TogetherTail("breach-the-multiverse/phyrexians",
	func(ctx *Context, entered []uuid.UUID, _ int) error { return breachPhyrexians(ctx, entered) })

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
