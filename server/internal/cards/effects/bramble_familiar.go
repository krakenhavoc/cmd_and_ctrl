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
//	 land card from among the milled cards onto the battlefield. (Then
//	 exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The mana ability is Llanowar Elves'. The return is Fae of
// Wishes' self-bounce with one discard in its cost.
//
// Fetch Quest is Bind to Life's mill and mandatory pick over a wider
// kind: only the cards this mill put into the graveyard are offered, and
// only those still there (a card a replacement sent
// elsewhere is not among them). With a creature, enchantment or land card
// among them you must put one onto the battlefield; with none, nothing
// happens. An Aura put onto the battlefield this way is attached as it
// enters by the engine's CR 303.4f choice. The Adventure lifecycle is the
// engine's (CR 715.3d).
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

// fetchQuestMilled is the clause after the mill, Bind to Life's with a
// wider kind: the creature, enchantment and land cards this mill put into
// the graveyard that are still there, one of which must go onto the
// battlefield (b17ReanimateOneMilled; a lone candidate comes back
// outright). An on-disk key: never renamed, never reused.
var fetchQuestMilled = CardTail("fetch-quest/after-mill", func(ctx *Context, o game.TailOutcome) error {
	player := ctx.Controller()
	match := Or(Creature(), Enchantment(), Land())
	var offer []uuid.UUID
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && match(ctx.Game, player, c) {
			offer = append(offer, id)
		}
	}
	return b17ReanimateOneMilled(ctx.Game, ctx.Item, player,
		"Fetch Quest — put a creature, enchantment, or land card from among the milled cards onto the battlefield", offer)
})
