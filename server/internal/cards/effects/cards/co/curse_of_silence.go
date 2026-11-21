package cardsco

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curse of Silence — Enchantment — Aura Curse {W}:
//
//	"Enchant player
//	 As this Aura enters, choose a card name.
//	 Spells with the chosen name enchanted player casts cost {2} more to
//	 cast.
//	 Whenever enchanted player casts a spell with the chosen name, you
//	 may sacrifice this Aura. If you do, draw a card."
//
// The name is chosen as the Aura enters (CR 614.12) through the card-name
// prompt Pithing Needle uses (ChooseCardNameAsEnters), stored on the
// permanent; until it is answered nothing is named, so nothing costs more
// and nothing triggers. The tax is a cost increase (CR 601.2f) on a spell
// the enchanted player casts whose name matches on any face (CR 201.2b);
// it is not an ability, so a special action or an activation is not taxed.
// The trigger watches the same cast, asks "you may sacrifice", and draws
// only if the Aura really was sacrificed (SacrificeThisThen: an Aura that
// left in response is not sacrificed, so nothing is drawn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36e267c5-ea98-4e26-b2b4-3faa05493eb4",
		Name:         "Curse of Silence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		AsEnters:     effects.ChooseCardNameAsEnters("Curse of Silence — choose a card name"),
		CostModifiers: []game.CostModifier{
			effects.CostsMore(2, "Spells with the chosen name enchanted player casts cost {2} more to cast.",
				func(q game.CostQuery) bool {
					named := strings.TrimSpace(q.Source.ChosenName)
					return named != "" && q.Ability == nil && q.SpecialAction == nil &&
						q.Source.AttachedTo.Kind == game.TargetPlayer && q.Controller == q.Source.AttachedTo.ID &&
						game.CardNameMatches(q.Card, named)
				}),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					named := strings.TrimSpace(source.ChosenName)
					if named == "" || source.AttachedTo.Kind != game.TargetPlayer || ev.Actor != source.AttachedTo.ID {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && game.CardNameMatches(c, named)
				}, "Curse of Silence — you may sacrifice it; if you do, draw a card", curseOfSilenceDraw),
				"Sacrifice Curse of Silence to draw a card?"),
		},
	})
}

// curseOfSilenceDraw is "sacrifice this Aura. If you do, draw a card". An
// on-disk key.
var curseOfSilenceDraw = effects.SacrificeThisThen("then/curse-of-silence-draw", func(ctx *effects.Context) error {
	return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
