package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of New Phyrexia — Battle — Siege, defense 6, for {X}{W}{U}:
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, create X 2/2 white and blue Knight
//	 creature tokens with vigilance."
//
// The last of S27's four battles (#92, #626) and the only one whose
// back face is a planeswalker. The bracketed reminder text is engine
// behaviour keyed on the card type (server/internal/game/battle.go),
// not card-effect data; what this Spec carries is the X Knights, and
// the defeated trigger that hands over Teferi Akosa of Zhalfir.
//
// The Knights are an ordinary "when this Siege enters" trigger. X is
// announced at CR 601.2b and the permanent remembers it (CR 400.7d,
// Card.CastX), so Build stamps it on the trigger's Params and the effect
// makes that many tokens as the trigger resolves — after the table has
// had its response window, as printed. What the two faces share is one
// oracle ID, and the back registers under it plus "#1" (see
// deluge_of_the_dead.go).
func init() {
	Register(Spec{
		OracleID:     invasionOfNewPhyrexiaOracleID,
		Name:         "Invasion of New Phyrexia",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 6,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "Invasion of New Phyrexia — create X 2/2 Knights with vigilance",
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Invasion of New Phyrexia — create X 2/2 white and blue Knight creature tokens with vigilance")
					item.Params.Amount = source.CastX()
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					// X can be 0, and CreateToken with N <= 0 makes
					// nothing, which is the printed outcome.
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("2/2 white and blue Knight with vigilance"),
						N:          item.Params.Amount,
					}.Apply(NewContext(g, item))
				},
			},
			DefeatedTrigger("Invasion of New Phyrexia — defeated: exile it, then cast Teferi Akosa of Zhalfir", SiegeDefeated()),
		},
	})
}

// invasionOfNewPhyrexiaOracleID is shared with the back face's spec,
// which registers under this ID plus "#1".
const invasionOfNewPhyrexiaOracleID = "480ea052-87d7-4fea-ac97-b6121b78f863"
