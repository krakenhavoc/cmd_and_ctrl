package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bramble Familiar // Fetch Quest — Creature — Elemental Raccoon {1}{G},
// 2/2 // Sorcery — Adventure {5}{G}{G} (Wilds of Eldraine, #4471):
//
//	Bramble Familiar — "{T}: Add {G}.
//	 {1}{G}, {T}, Discard a card: Return this creature to its owner's
//	 hand."
//	Fetch Quest — "Mill seven cards. Then put a creature, enchantment, or
//	 land card from among the milled cards onto the battlefield."
//
// Fetch Quest is Bind to Life's mill and keyed tail with the wider filter:
// among the cards the mill really put into the graveyard that are still
// there, a lone candidate comes back outright and several are offered as an
// exactly-one pick. The return is Aethertide Whale's, with a tap and a
// discard in the cost.
//
// No simplification.
const brambleFamiliarOracleID = "68f36dfa-bbd2-4860-b09b-37e35136be72"

func init() {
	Register(Spec{
		OracleID:     brambleFamiliarOracleID,
		Name:         "Bramble Familiar",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, {T}, Discard a card: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    Plus(ManaCost("{1}{G}"), TapCost(), DiscardACard()),
			Effect:  returnThisPermanentToOwnersHand,
		}},
	})
	Register(Spec{
		OracleID:     brambleFamiliarOracleID + "#1",
		Name:         "Fetch Quest",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return MillToZone{N: 7, ThenTail: ctx.ItemTail(fetchQuestMilled)}.Apply(ctx)
		},
	})
}

// fetchQuestMilled is the clause after the mill: one creature, enchantment
// or land card among the cards milled this way, onto the battlefield.
var fetchQuestMilled = CardTail("fetch-quest/after-mill", func(ctx *Context, o game.TailOutcome) error {
	var picks []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && (c.IsCreature() || c.IsEnchantment() || c.IsLand()) {
			picks = append(picks, id)
		}
	}
	return b17ReanimateOneMilled(ctx.Game, ctx.Item, ctx.Controller(),
		"Fetch Quest — put a creature, enchantment, or land card milled this way onto the battlefield", picks)
})
