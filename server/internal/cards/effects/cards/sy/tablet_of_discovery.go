package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tablet of Discovery — Artifact {2}{R}:
//
//	"When this artifact enters, mill a card. You may play that card
//	 this turn.
//	 {T}: Add {R}.
//	 {T}: Add {R}{R}. Spend this mana only to cast instant and sorcery
//	 spells."
//
// The entry trigger mills through MillToZone, so a replacement that
// sends the card elsewhere leaves nothing to play, and grants a play
// permission over the milled card where it landed: Emry, Lurker of the
// Loch's graveyard grant, but "play" rather than "cast", so a milled
// land can be played as the turn's land drop. The zero Duration is
// "this turn", and the grant names the card object (CR 400.7), so it
// lapses if the card leaves the graveyard.
//
// The second mana ability is one tag, an OR of the two types
// (ManaRestrictAnyType), on a cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19cf5798-4600-4a66-b1c7-de77bde157d0",
		Name:         "Tablet of Discovery",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tablet of Discovery — mill a card. You may play that card this turn.", tabletOfDiscoveryMill),
		},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{R}",
				Label:    "Add {R}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{R}{R}",
				Label:    "Add {R}{R} (instant and sorcery spells only)",
				Restrictions: []string{
					effects.ManaRestrictCast,
					effects.ManaRestrictAnyType("Instant", "Sorcery"),
				},
			},
		},
	})
}

func tabletOfDiscoveryMill(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	return effects.MillToZone{
		Player:   controller,
		N:        1,
		ThenTail: effects.NewContext(g, item).ItemTail(tabletOfDiscoveryMilled),
	}.Apply(effects.NewContext(g, item))
}

// tabletOfDiscoveryMilled is the clause after the mill (#4253 slice 9).
var tabletOfDiscoveryMilled = effects.CardTail("tablet-of-discovery/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	controller := ctx.Controller()
	milled := o.IDs

	for _, id := range milled {
		ctx.Game.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player: controller,
			Zone:   game.ZoneGraveyard,
			Source: ctx.Source(),
			Label:  "Tablet of Discovery — you may play that card this turn",
		})
	}
	return nil
})
