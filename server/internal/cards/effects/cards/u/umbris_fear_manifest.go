package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Umbris, Fear Manifest — Legendary Creature — Nightmare Horror {3}{U}{B},
// 1/1 (Innistrad: Crimson Vow):
//
//	"Umbris gets +1/+1 for each card your opponents own in exile.
//	 Whenever Umbris or another Nightmare or Horror you control enters,
//	 target opponent exiles cards from the top of their library until they
//	 exile a land card."
//
// The bonus is a layer 7c count over the exile zone: every card an opponent
// of Umbris's controller owns there, face up or down, a token never (a token
// ceases to exist in exile). The enters trigger targets an opponent; it
// exiles from the top of that player's library one card at a time until a
// land card is exiled, or the library is gone, and leaves them all in exile,
// where they feed the count. Umbris's own entry counts, and so does every
// other Nightmare or Horror it controls (a changeling is both).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af31febb-af53-4b48-b930-9e1f9e5d8f26",
		Name:         "Umbris, Fear Manifest",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Umbris gets +1/+1 for each card your opponents own in exile.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, e := range g.Exile.Cards {
					if e.Owner != source.Controller && !e.IsToken() {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, umbrisOrAnotherNightmareOrHorrorEntered,
					"Umbris, Fear Manifest — target opponent exiles cards from the top of their library until they exile a land card",
					umbrisExileUntilLand),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

func umbrisOrAnotherNightmareOrHorrorEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
	if !ok {
		return false
	}
	return c.InstanceID == source.InstanceID ||
		c.HasSubtype(source.TextSubtype("Nightmare")) || c.HasSubtype(source.TextSubtype("Horror"))
}

func umbrisExileUntilLand(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	opp := effects.TargetedPlayer(ctx)
	if opp == uuid.Nil {
		return nil
	}
	return g.ExileTopUntilForEffect(opp, game.Card.IsLand)
}
