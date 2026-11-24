package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Fifteenth Doctor — Legendary Creature — Time Lord Doctor {2}{U}{R}, 3/3:
//
//	"Whenever The Fifteenth Doctor enters or attacks, mill three cards. You may
//	 put an artifact card with mana value 2 or 3 from among them into your hand.
//	 The first nonartifact spell you cast each turn has improvise. (Your
//	 artifacts can help cast that spell. Each artifact you tap after you're done
//	 activating mana abilities pays for {1}.)"
//
// One trigger for both events (WhenThisEntersOrAttacks), Ainok Wayfarer's
// mill-then-take: only the cards this mill put into the graveyard are offered,
// and only an artifact card with mana value 2 or 3 among them. The improvise
// grant is Ironheart's GrantedTapCost, narrowed to the first nonartifact spell
// its controller casts each turn: the cast tally counts every spell and the
// artifact ones apart, and is bumped after the spell is priced (CR 601.2f), so
// a tally with no nonartifact spell in it means this one is the first. An
// artifact spell never has it and never spends it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90aa8344-ffc7-4a1a-a02a-80eceb1babe5",
		Name:         "The Fifteenth Doctor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("The Fifteenth Doctor — mill three cards; you may put an artifact card with mana value 2 or 3 from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(fifteenthDoctorMilled)}.Apply(ctx)
				}),
		},
		GrantedTapCosts: []game.GrantedTapCost{effects.SpellsYouCastHaveImprovise(
			"The first nonartifact spell you cast each turn has improvise.",
			func(g *game.Game, caster uuid.UUID, spell game.Card, _ game.ZoneKind) bool {
				if spell.IsArtifact() {
					return false
				}
				tally := g.CastTallyFor(caster)
				return tally.Total-tally.Artifact == 0
			})},
	})
}

// fifteenthDoctorMilled offers an artifact card with mana value 2 or 3 from
// among the milled cards. An on-disk key.
var fifteenthDoctorMilled = effects.CardTail("the-fifteenth-doctor/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.MayTakeOneFromAmongThem(ctx, o.IDs, effects.And(effects.Artifact(), effects.ManaValueGE(2), effects.ManaValueLE(3)),
		"The Fifteenth Doctor — you may put an artifact card with mana value 2 or 3 from among them into your hand")
})
