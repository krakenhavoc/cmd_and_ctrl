package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Edgar, Master Machinist — Legendary Creature — Human Artificer Noble
// {2}{R}{W}, 2/4 (deck request #3655):
//
//	"Once during each of your turns, you may cast an artifact spell from
//	 your graveyard. If you cast a spell this way, that artifact enters
//	 tapped.
//	 Tools — Whenever Edgar attacks, it gets +X/+0 until end of turn,
//	 where X is the greatest mana value among artifacts you control."
//
// The cast is Gisa and Geralf's standing graveyard permission (ADR 0066)
// with a one-use budget of the "artifact" type per Edgar object per turn
// (CastPermission.PerType), closed off your turn (TimingYourTurnOnly) and
// leaving the artifact's own timing and costs in force. "That artifact
// enters tapped" is the permission's EntersTapped (#3768): the spell
// carries it to its entry, so the artifact enters tapped through the
// CR 614 pipeline rather than entering untapped and being tapped. An
// artifact cast from the graveyard by its own text (a printed
// flashback) or by another permission does not use Edgar's and enters
// as it would.
//
// Tools reads the greatest mana value as the trigger resolves, Edgar
// himself included if he is an artifact by then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75348d55-99cd-4dab-97be-1f818c43a69d",
		Name:         "Edgar, Master Machinist",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:         game.ZoneGraveyard,
			CastOnly:     true,
			Filter:       game.PermissionFilter{CardType: "artifact"},
			PerType:      []string{game.PermissionTypeArtifact},
			Timing:       game.TimingYourTurnOnly,
			EntersTapped: true,
			Label:        "Once during each of your turns, cast an artifact spell from your graveyard; it enters tapped (Edgar, Master Machinist)",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Edgar, Master Machinist — Tools: +X/+0, X the greatest mana value among your artifacts", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{
					Target: ctx.Source(),
					Power:  b27GreatestManaValueAmongArtifactsControlled(g, item.Controller),
					Label:  "Edgar, Master Machinist — Tools",
				}.Apply(ctx)
			}),
		},
	})
}
