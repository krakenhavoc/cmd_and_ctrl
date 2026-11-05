package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dominion Infiltration — {X}{R}{R}{R} Sorcery (Star Trek, #4973):
//
//	"Choose X target creatures your opponents control. For each of them,
//	 create a token that's a copy of that creature, except it has
//	 changeling. Those tokens gain haste. Exile those tokens at the
//	 beginning of the next end step. (They're every creature type.)"
//
// Curse of the Swine's X-counted target clause with Kindred Charge's
// copies. Every copy is made by ONE creation, a group per still-legal
// target (CreateTokenCopiesTogether, CR 608.2f), so they enter together.
// "Except it has changeling" is an exception to the copy (CR 707.9b), so
// it is part of each token's copiable values; "those tokens gain haste"
// is an effect of the spell, not an exception (CR 707.9a), so haste is a
// grant on each token and a copy of one has none. The grant and the
// exile, a CR 603.7 delayed trigger over exactly the tokens this spell
// made, are the creation's keyed tail.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30597aff-ebb6-4927-b350-fe559fb92c8f",
		Name:         "Dominion Infiltration",
		Completeness: CompletenessFull,
		Targets:      targetsCountedByX(TargetCreature("X target creatures your opponents control", OpponentControls())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var originals []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					originals = append(originals, t.ID)
				}
			}
			if len(originals) == 0 {
				return nil
			}
			return CreateTokenCopiesTogether{
				Controller: item.Controller,
				Copies:     originals,
				Except:     dominionInfiltrationChangeling,
				Then:       ctx.ItemTail(dominionInfiltrationCopiesRest),
			}.Apply(ctx)
		},
	})
}

// dominionInfiltrationChangeling is the "except it has changeling" clause.
func dominionInfiltrationChangeling(t *game.Card) {
	for _, kw := range t.Keywords {
		if kw == game.KeywordChangeling {
			return
		}
	}
	t.Keywords = append(t.Keywords, game.KeywordChangeling)
}

// dominionInfiltrationCopiesRest is "Those tokens gain haste. Exile those
// tokens at the beginning of the next end step.", told the tokens that
// landed. An on-disk key.
var dominionInfiltrationCopiesRest = CardTail("dominion-infiltration/haste-and-exile-copies", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(), Label: "Dominion Infiltration — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		Label: "Dominion Infiltration — exile the tokens",
		Cards: o.IDs,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
})
