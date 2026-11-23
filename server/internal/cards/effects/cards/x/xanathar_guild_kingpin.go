package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Xanathar, Guild Kingpin — Legendary Creature — Beholder {4}{U}{B}, 5/6
// (Adventures in the Forgotten Realms):
//
//	"At the beginning of your upkeep, choose target opponent. Until end of
//	 turn, that player can't cast spells, you may look at the top card of
//	 their library any time, you may play the top card of their library, and
//	 you may spend mana as though it were mana of any color to cast spells
//	 this way."
//
// The opponent is a target, chosen as the trigger goes on the stack
// (CR 603.3d) and re-checked as it resolves; with it gone nothing happens
// (CR 608.2b). Both halves are one-turn effects of the resolving ability
// (CR 611.2): a cast ban on that player (RestrictCasting, Avatar's Wrath's
// seam) and one stored standing permission over their library
// (game.CastPermission with ZoneOwner and SeesLibraryTop, the cross-seat
// grant the engine already carries for exactly this clause): "look at the
// top card any time" is SeesLibraryTop, "play the top card" is the
// permission itself (lands included, CR 305.1), and AnyColor spends mana as
// though it were any colour to cast spells this way (CR 609.4b). A spell
// cast this way is owned by the library's owner, who is the one barred
// from casting.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a168ba5-8e11-41dd-8f8a-8cb23bbde056",
		Name:         "Xanathar, Guild Kingpin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourUpkeep("Xanathar, Guild Kingpin — choose target opponent: until end of turn they can't cast spells, and you may look at and play the top card of their library",
				xanatharUpkeep),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

func xanatharUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := (effects.RestrictCasting{
			Player:   t.ID,
			Rule:     game.CastBanRule{Kind: game.CastBanOutright},
			Label:    "Xanathar, Guild Kingpin — can't cast spells this turn",
			Duration: effects.DurationUntilEndOfTurn(ctx),
		}).Apply(ctx); err != nil {
			return err
		}
		g.GrantCastPermissionForEffect(game.CastPermission{
			Player:           ctx.Controller(),
			Zone:             game.ZoneLibrary,
			Scope:            game.ScopeStanding,
			ZoneOwner:        t.ID,
			TopOfLibraryOnly: true,
			SeesLibraryTop:   true,
			AnyColor:         true,
			Duration:         g.UntilEndOfTurnDuration(),
			Source:           ctx.Source(),
			SourceName:       "Xanathar, Guild Kingpin",
			Label:            "Play the top card of their library (Xanathar, Guild Kingpin)",
		})
		return nil
	}
	return nil
}
