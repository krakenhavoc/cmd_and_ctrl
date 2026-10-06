package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Conduit of Worlds — Artifact {2}{G}{G}:
//
//	"You may play lands from your graveyard.
//	{T}: Choose target nonland permanent card in your graveyard. If you
//	haven't cast a spell this turn, you may cast that card. If you do,
//	you can't cast additional spells this turn. Activate only as a
//	sorcery."
//
// The first line is Crucible of Worlds' standing graveyard permission.
//
// The ability is a stored cast permission over the one targeted card
// (#2173, game/cast_follow_up.go), the shape cascade uses for "cast it
// as part of the resolution" (CR 608.2g): flash-timed, closed by the
// holder's next priority pass, an uncast card simply staying in the
// graveyard. Three of the printed clauses are data on it:
//
//   - "If you haven't cast a spell this turn" is read as the ability
//     resolves (no grant at all otherwise) AND kept honest by
//     RequiresNoSpellsCast, so a spell cast before the holder passes
//     ends the offer rather than letting a second spell through.
//   - "you may cast that card" pays the card's mana cost: the grant
//     carries no price of its own.
//   - "If you do, you can't cast additional spells this turn" is the
//     permission's FollowUp, which runs only when the cast is made
//     through it. A declined or lapsed offer binds nobody.
//
// A card the graveyard already lets you cast (flashback, a Gravecrawler)
// is cast through its own text rather than this permission when the
// holder claims that path, which does not run the follow-up; the ban is
// a cost of THIS permission only.
//
// No further simplification.
var conduitOfWorldsNoMoreSpells = game.RegisterCastFollowUp("conduit-of-worlds-no-more-spells",
	func(g *game.Game, f game.CastFollowUp) error {
		g.GrantCastBanForEffect(f.Player, game.CastBanRule{Kind: game.CastBanOutright},
			"Conduit of Worlds — can't cast additional spells this turn", f.Source, g.UntilEndOfTurnDuration())
		return nil
	})

func init() {
	Register(Spec{
		OracleID:     "ed14be15-8f8d-4fe3-a147-f5da8ed873bf",
		Name:         "Conduit of Worlds",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true},
			Label:  "Play a land from your graveyard (Conduit of Worlds)",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{T}: Choose target nonland permanent card in your graveyard. If you haven't cast a spell this turn, you may cast that card. If you do, you can't cast additional spells this turn.",
			Cost:         TapCost(),
			SorcerySpeed: true,
			Targets: TargetCardInGraveyard("target nonland permanent card in your graveyard",
				YouOwn(), Permanent(), Nonland()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if g.CastTallyFor(ctx.Controller()).Total > 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					c, ok := g.LookupCardForEffect(t.ID)
					if z := g.FindCardZoneForEffect(t.ID); !ok || z == nil || z.Kind != game.ZoneGraveyard {
						continue
					}
					g.GrantCastPermissionToCardsForEffect(game.CastPermission{
						Player:               ctx.Controller(),
						Zone:                 game.ZoneGraveyard,
						Duration:             g.UntilEndOfTurnDuration(),
						Timing:               game.TimingFlash,
						CastOnly:             true,
						LapseOnPass:          game.LapseStaysInExile,
						RequiresNoSpellsCast: true,
						FollowUp:             conduitOfWorldsNoMoreSpells,
						Source:               ctx.Source(),
						SourceName:           "Conduit of Worlds",
						Label:                "Conduit of Worlds — cast it from your graveyard",
					}, []game.Card{c})
				}
				return nil
			},
		}},
	})
}
