package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mole Man, Moloid Master — {2}{G} Legendary Creature — Human Villain 1/1
// (Marvel Super Heroes, #4455):
//
//	"You may play lands from your graveyard.
//	 Landfall — Whenever a land you control enters, create a 1/1 green Minion
//	 creature token named Moloid with "Whenever this token attacks, you may
//	 mill a card.""
//
// The permission is Crucible of Worlds' (a land play from the graveyard
// counts against the land drop). The Moloid is a catalog token template
// (ADR 0083), so its attack trigger travels with the token and resolves as
// a real stack item, the way Falcon's Redwing does; the mill is optional.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34994475-b914-44f6-8ca1-c2747b6d5956",
		Name:         "Mole Man, Moloid Master",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true},
			Label:  "Play a land from your graveyard (Mole Man)",
		}},
		Triggered: []game.TriggeredAbility{
			Landfall("Mole Man — create a 1/1 green Minion named Moloid",
				Do(CreateToken{Template: mshB6MoloidToken(), N: 1})),
		},
	})
}

// mshB6MoloidToken is the 1/1 green Minion named Moloid with "Whenever this
// token attacks, you may mill a card."
func mshB6MoloidToken() game.Card { return tokenFromCatalog(printedMshB6MoloidToken) }

// printedMshB6MoloidToken is the Moloid as PRINTED.
func printedMshB6MoloidToken() tokenTemplate {
	return tokenTemplate{
		Slug: "moloid",
		Card: game.Card{
			Name:      "Moloid",
			TypeLine:  "Token Creature — Minion",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"G"},
		},
		Triggered: []game.TriggeredAbility{
			Optional(WheneverThisAttacks("Moloid — mill a card", Do(MillCards{N: 1})), "Moloid — mill a card?"),
		},
		Text: "Whenever this token attacks, you may mill a card.",
	}
}
