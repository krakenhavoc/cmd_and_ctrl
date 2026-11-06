package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Izoni, Center of the Web — Legendary Creature — Elf Detective
// {4}{B}{G}, 5/4 (Murders at Karlov Manor, #5193):
//
//	"Menace
//	 Whenever Izoni enters or attacks, you may collect evidence 4. If you
//	 do, create two 2/1 black and green Spider creature tokens with reach
//	 and menace.
//	 Sacrifice four tokens: Surveil 2, then draw two cards. You gain 2
//	 life."
//
// The trigger is one ability watching two events (WhenThisEntersOrAttacks).
// Collect evidence is the effect form (RegisterCollectEvidence): as it
// resolves Izoni's controller may exile graveyard cards with total mana
// value 4 or more, and only then are the Spiders created. A graveyard
// holding less than 4 in all asks nothing (CR 701.59b).
//
// The sacrifice ability takes any four tokens you control, creature or not
// (CR 111.1). Surveil comes first, so the cards drawn come from the library
// as arranged; the life follows the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8074afbf-190b-4a24-b939-cc559bdffe38",
		Name:            "Izoni, Center of the Web",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Izoni, Center of the Web — you may collect evidence 4; if you do, create two 2/1 Spiders with reach and menace",
				func(g *game.Game, item *game.StackItem) error {
					return izoniSpiders.Ask(NewContext(g, item), "Izoni, Center of the Web — collect evidence 4 (exile cards with total mana value 4 or more from your graveyard) to create two Spiders? Choose none to decline.")
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice four tokens: Surveil 2, then draw two cards. You gain 2 life.",
			Cost:    SacrificeN(4, "four tokens", IsTokenPredicate()),
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 2},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return Surveil{Player: item.Controller, N: 2, ThenTail: ctx.Tail(izoniDrawGain)}.Apply(ctx)
			},
		}},
	})
}

// izoniSpiders is "you may collect evidence 4. If you do, create two
// Spiders". An on-disk key: never renamed, never reused.
var izoniSpiders = RegisterCollectEvidence("choose-cards/izoni-center-of-the-web-collect-evidence", 4,
	func(ctx *Context, _ []uuid.UUID) error {
		return CreateToken{Template: TokenCard("2/1 black and green Spider with menace and reach"), N: 2}.Apply(ctx)
	})

// izoniDrawGain is "then draw two cards. You gain 2 life", keyed (#4253):
// an on-disk tail the surveil prompt carries.
var izoniDrawGain = game.RegisterTail("then/izoni-center-of-the-web-draw-two-gain-two", func(g *game.Game, o game.TailOutcome) error {
	if err := g.DrawNForEffect(o.Tail.Player, 2); err != nil {
		return err
	}
	return g.ChangePlayerLifeForEffect(o.Tail.Source, o.Tail.Player, 2)
})
