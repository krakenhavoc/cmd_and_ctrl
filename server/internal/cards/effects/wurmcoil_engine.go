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
// that creates two 3/3 Phyrexian Wurm tokens when it resolves. S21
// sub-PR 1 makes token keywords real, so the two halves are now
// distinct — one deathtouch Wurm, one lifelink Wurm, as printed —
// instead of two copies of a vanilla template. Wurmcoil's own
// printed deathtouch / lifelink ride the S18 keyword pipeline via
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
						ctx := NewContext(g, item)
						for _, tmpl := range []game.Card{
							PhyrexianWurmDeathtouchToken(),
							PhyrexianWurmLifelinkToken(),
						} {
							if err := (CreateToken{
								Controller: item.Controller,
								Template:   tmpl,
								N:          1,
							}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
	})
}
