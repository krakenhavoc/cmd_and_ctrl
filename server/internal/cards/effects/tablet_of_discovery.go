package effects

import (
	"github.com/google/uuid"
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
	Register(Spec{
		OracleID:     "19cf5798-4600-4a66-b1c7-de77bde157d0",
		Name:         "Tablet of Discovery",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tablet of Discovery — mill a card. You may play that card this turn.", tabletOfDiscoveryMill),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{R}",
				Label:    "Add {R}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{R}{R}",
				Label:    "Add {R}{R} (instant and sorcery spells only)",
				Restrictions: []string{
					ManaRestrictCast,
					ManaRestrictAnyType("Instant", "Sorcery"),
				},
			},
		},
	})
}

func tabletOfDiscoveryMill(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	return MillToZone{
		Player: controller,
		N:      1,
		Then: func(ctx *Context, milled []uuid.UUID) error {
			for _, id := range milled {
				ctx.Game.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
					Player: controller,
					Zone:   game.ZoneGraveyard,
					Source: ctx.Source(),
					Label:  "Tablet of Discovery — you may play that card this turn",
				})
			}
			return nil
		},
	}.Apply(NewContext(g, item))
}
