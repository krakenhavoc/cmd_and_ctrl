package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wurmcoil Engine — 6/6 Artifact Creature — Phyrexian Wurm for {6}:
//
//	"Deathtouch, lifelink.
//	When Wurmcoil Engine dies, create a 3/3 colorless Phyrexian Wurm
//	artifact creature token with deathtouch and a 3/3 colorless
//	Phyrexian Wurm artifact creature token with lifelink."
//
// S19 sub-PR 4 implements the dies half: a mandatory LTB trigger
// that creates two 3/3 Phyrexian Wurm tokens when it resolves. The
// real card splits deathtouch onto one and lifelink onto the other;
// token keywords are cosmetic in the sandbox, so both come from the
// shared PhyrexianWurmToken template (N: 2). Wurmcoil's own printed
// deathtouch / lifelink ride the S18 keyword pipeline via
// PrintedKeywords. cardDied gates the trigger to graveyard-only.
func init() {
	Register(Spec{
		OracleID:        "d1a60f44-7696-49ee-91fb-cab5b3102962",
		Name:            "Wurmcoil Engine",
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Wurmcoil Engine — create two 3/3 Wurms",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   PhyrexianWurmToken(),
							N:          2,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
