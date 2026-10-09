package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twisted Sewer-Witch — Creature — Human Warlock {3}{B}{B}, 3/4:
//
//	"When this creature enters, create a 1/1 black Rat creature token
//	 with "This creature can't block." Then for each Rat you control,
//	 create a Wicked Role token attached to that Rat. (If you control
//	 another Role on it, put that one into the graveyard. Enchanted
//	 creature gets +1/+0. When this token is put into a graveyard, each
//	 opponent loses 1 life.)"
//
// "Each Rat you control" is counted after the token exists, so the new
// Rat gets a Role and so does every Rat that was already there (a
// changeling counts, CR 702.73a). The set is read once, then one Role
// each.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "99c9920a-f84b-4626-9fe5-d94cc62cd277",
		Name:         "Twisted Sewer-Witch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Twisted Sewer-Witch — create a Rat token, then a Wicked Role on each Rat you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CreateToken{Controller: item.Controller, Template: CantBlockRatToken(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					var rats []game.Card
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype("Rat") {
							rats = append(rats, c)
						}
					}
					for _, c := range rats {
						if err := (CreateRoleToken{Role: RoleWicked, Host: c.InstanceID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}

const cantBlockRatSlug = "rat-cant-block"

// printedCantBlockRatToken is Twisted Sewer-Witch's 1/1 black Rat with
// "This creature can't block." — a token template because the ability
// is a static the token carries, which a plain tokens_table row cannot
// declare.
func printedCantBlockRatToken() tokenTemplate {
	return tokenTemplate{
		Slug: cantBlockRatSlug,
		Card: game.Card{
			Name:      "Rat",
			TypeLine:  "Token Creature — Rat",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"B"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "This creature can't block.",
	}
}

// CantBlockRatToken is the Rat's card value for CreateToken.
func CantBlockRatToken() game.Card { return tokenFromCatalog(printedCantBlockRatToken) }
