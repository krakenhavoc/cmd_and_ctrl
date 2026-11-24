package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coram, the Undertaker — {1}{B}{R}{G} Legendary Creature — Human Warrior 0/5
//
//	"Coram gets +X/+0, where X is the greatest power among creature cards
//	 in all graveyards.
//	 Whenever Coram attacks, each player mills a card.
//	 During each of your turns, you may play a land and cast a spell from
//	 among cards in graveyards that were put there from libraries this
//	 turn."
//
// The first line is a layer 7c pump read off every graveyard, as printed
// characteristics (a card in a graveyard has no layers applied).
//
// The third line is two STANDING graveyard permissions (ADR 0066), derived
// off the battlefield on every query, so they end the moment Coram leaves:
//
//   - CastPermission.AllGraveyards (#6092) reaches a card in ANY player's
//     graveyard, not only Coram's controller's;
//   - PermissionFilter.FromLibraryToGraveyardThisTurn (Kagha's clause)
//     narrows it to cards that arrived from a library this turn — a mill,
//     a surveil, any move out of a library. A card that came from a hand
//     or the battlefield, or that left and came back (CR 400.7), does not
//     qualify;
//   - "A land and a spell" is one of each, not one of either: the land
//     half is a LandsOnly grant with a one-use budget under the type
//     "land", the spell half a CastOnly grant with a one-use budget under
//     a different type, so the two are tallied apart (per_type_permission.go
//     keys the spent use by holder, granting object and type);
//   - "During each of your turns" is TimingYourTurnOnly; the card's own
//     timing still applies, so a land needs a main phase, an empty stack
//     and an unused land drop, and a sorcery is cast at sorcery speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51c511c0-783a-456d-a7e2-959f6e8ddfc8",
		Name:         "Coram, the Undertaker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			Label:     "Coram gets +X/+0, where X is the greatest power among creature cards in all graveyards.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				best := 0
				for _, p := range g.Seats {
					if p == nil {
						continue
					}
					if _, power := effects.Cmdr2S30CreatureCardsIn(g, p.ID); power > best {
						best = power
					}
				}
				c.Power += best
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Coram, the Undertaker — each player mills a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B02EachPlayerMills(g, item, 1)
				}),
		},
		CastPermissions: []game.CastPermission{
			{
				Zone:          game.ZoneGraveyard,
				AllGraveyards: true,
				Filter:        game.PermissionFilter{CardInZone: true, FromLibraryToGraveyardThisTurn: true, LandsOnly: true},
				PerType:       []string{game.PermissionTypeLand},
				Timing:        game.TimingYourTurnOnly,
				Label:         "During your turn, play a land from among cards in graveyards that were put there from libraries this turn (Coram, the Undertaker)",
			},
			{
				Zone:          game.ZoneGraveyard,
				AllGraveyards: true,
				Filter:        game.PermissionFilter{CardInZone: true, FromLibraryToGraveyardThisTurn: true},
				CastOnly:      true,
				PerType:       []string{game.PermissionTypePermanent},
				Timing:        game.TimingYourTurnOnly,
				Label:         "During your turn, cast a spell from among cards in graveyards that were put there from libraries this turn (Coram, the Undertaker)",
			},
		},
	})
}
