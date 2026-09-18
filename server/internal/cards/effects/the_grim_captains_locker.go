package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Grim Captain's Locker — Legendary Artifact {3}{B}:
//
//	"{T}: Surveil 1.
//	 {T}: Until end of turn, each creature card in your graveyard gains
//	 'Escape—{3}{B}, Exile four other cards from your graveyard.'"
//
// Past in Flames' shape pointed at escape instead of flashback, and
// activated rather than cast — which changes nothing about the model
// and is the point of writing it. The set is locked when the ABILITY
// resolves (CR 611.2c), so a creature card that reaches the graveyard
// after the activation has no escape, and the Locker has to be
// untapped again to give it one.
//
// Escape deliberately does NOT carry flashback's exile replacement: an
// escaped creature enters the battlefield and dies to the graveyard
// like any other, and can escape again the next time something opens
// the window (CR 702.138). The cost's non-mana half — "exile four
// other cards from your graveyard" — is the same AlternativeCost
// component the printed escape keyword uses, so a granted escape and a
// printed one are paid through the same prompt.
//
// Two abilities rather than one modal: they are printed as two
// separate activations, both tap the Locker, and a player may take
// either one per untap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29931e18-9dac-46fe-b9a4-72d838d79882",
		Name:         "The Grim Captain's Locker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:  "{T}: Surveil 1.",
				Cost:   TapCost(),
				Effect: grimLockerSurveil,
			},
			{
				Label:  "{T}: Until end of turn, each creature card in your graveyard gains \"Escape—{3}{B}, Exile four other cards from your graveyard.\"",
				Cost:   TapCost(),
				Effect: grimLockerGrantEscape,
			},
		},
	})
}

// grimLockerSurveil is the first ability. Package-level so the
// activation captures nothing and survives an undo.
func grimLockerSurveil(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
}

// grimLockerGrantEscape is the second. The set of creature cards is
// locked HERE, as the ability resolves (CR 611.2c).
func grimLockerGrantEscape(g *game.Game, item *game.StackItem) error {
	return GrantCastFromYourGraveyard{
		Filter:                  game.PermissionFilter{CreatureOnly: true},
		AltCostKey:              "escape",
		Cost:                    "{3}{B}",
		ExileOtherFromGraveyard: 4,
		Label:                   "Escape—{3}{B}, Exile four other cards from your graveyard",
	}.Apply(NewContext(g, item))
}
