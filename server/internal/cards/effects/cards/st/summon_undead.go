package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summon Undead — Sorcery {4}{B} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"You may mill three cards. Then return a creature card from your graveyard
//	 to the battlefield."
//
// Another Chance's shape: the yes/no mill is asked as the spell resolves, only
// when the library holds three cards (a player can't choose to mill more than
// their library holds), then the return either way: a creature card of the
// caster's graveyard as it is now, the milled ones included, chosen by the
// caster (not targeted) and put onto the battlefield under their control.
// A graveyard with no creature card returns nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c86ff0e3-5d6e-4981-a2da-8da70b9daeb0",
		Name:         "Summon Undead",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Library != nil && p.Library.Size() >= 3 {
				return effects.MayChoice{
					Question: "Summon Undead — mill three cards?",
					Then:     summonUndeadMill,
				}.Apply(ctx)
			}
			return summonUndeadReturn(ctx)
		},
	})
}

// summonUndeadMill is the answer to the mill question: mill on a yes, then the
// return either way. An on-disk key.
var summonUndeadMill = effects.MayAnswer("confirm/summon-undead-mill", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := (effects.MillCards{N: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return summonUndeadReturn(ctx)
})

func summonUndeadReturn(ctx *effects.Context) error {
	return effects.ReturnChosenFromGraveyard{
		Player:   ctx.Controller(),
		Question: "Summon Undead — return a creature card from your graveyard to the battlefield",
		Match:    func(c game.Card) bool { return c.IsCreature() },
		Min:      1,
		Max:      1,
	}.Apply(ctx)
}
