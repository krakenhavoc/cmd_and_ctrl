package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cryotheory Adept — Creature — Human Wizard {1}{U}, 2/1:
//
//	"Prowess
//	 {3}{U}, Exile this card from your graveyard: Tap target creature
//	 and put a stun counter on it. Activate only as a sorcery."
//
// The exile is a cost, so the ability is on the stack with the card
// already gone from the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a9b3ee21-1af1-4e7e-8e48-d7c90d056f5c",
		Name:            "Cryotheory Adept",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Activated: []ActivatedAbility{{
			Label:        "{3}{U}, Exile this card from your graveyard: Tap target creature and put a stun counter on it. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{U}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				if err := (TapTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
					return err
				}
				return AddCounter{Target: ts[0].ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
			},
		}},
	})
}
