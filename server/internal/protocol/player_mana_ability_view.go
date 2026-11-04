package protocol

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// PlayerManaAbilityView is one mana ability a player has, bound to no
// card (#4885): "pay 1 life: add {C}" from Channel. The client draws it
// as a button beside the seat's mana pool and sends
// activate_mana_ability {player_ability: source}.
type PlayerManaAbilityView struct {
	// Source is the granting object's id and the handle the activation
	// names.
	Source string `json:"source"`
	// Label is the granting card's name ("Channel").
	Label string `json:"label"`
	// LifeCost is the life one activation costs.
	LifeCost int `json:"life_cost"`
	// Produced is the mana one activation adds ("{C}").
	Produced string `json:"produced"`
	// Usable is false while the player cannot pay the life (CR 119.4)
	// or may make mana only with lands; the client greys the button.
	Usable bool `json:"usable"`
}

// viewOfPlayerManaAbilities projects a player's live player mana
// abilities. Runs under the game's read lock, like the rest of the
// PlayerView builder.
func viewOfPlayerManaAbilities(g *game.Game, player uuid.UUID) []PlayerManaAbilityView {
	offers := g.PlayerManaAbilitiesLocked(player)
	if len(offers) == 0 {
		return nil
	}
	out := make([]PlayerManaAbilityView, 0, len(offers))
	for _, o := range offers {
		out = append(out, PlayerManaAbilityView{
			Source:   o.Source.String(),
			Label:    o.Label,
			LifeCost: o.LifeCost,
			Produced: o.Produced,
			Usable:   o.Payable,
		})
	}
	return out
}
