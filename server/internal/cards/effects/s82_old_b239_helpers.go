package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB239YouHaveADesert is "you control a Desert or there is a Desert card in
// your graveyard" (Hour of Devastation's Desert payoffs): a Desert permanent
// under `you`, or a card with the Desert subtype in `you`'s graveyard. It reads
// only board state, so a static ability may call it.
func OldB239YouHaveADesert(g *game.Game, you uuid.UUID) bool {
	if len(PermanentsControlledByMatching(g, you, Subtype("Desert"))) > 0 {
		return true
	}
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Graveyard == nil {
		return false
	}
	for _, c := range p.Graveyard.Cards {
		if c.HasSubtype("Desert") {
			return true
		}
	}
	return false
}

// OldB239EnchantedPlayer is the player an Aura that enchants a player is
// attached to.
func OldB239EnchantedPlayer(aura *game.Card) (uuid.UUID, bool) {
	if aura == nil || aura.AttachedTo.Kind != game.TargetPlayer || aura.AttachedTo.ID == uuid.Nil {
		return uuid.Nil, false
	}
	return aura.AttachedTo.ID, true
}

// OldB239EnchantedPlayersCreatures is the AppliesTo of "creatures enchanted
// player controls": every creature controlled by the player the source Aura
// enchants.
func OldB239EnchantedPlayersCreatures(target *game.Card, _ *game.Game, source *game.Card) bool {
	player, ok := OldB239EnchantedPlayer(source)
	return ok && target.IsCreature() && target.Controller == player
}

// OldB239EnchantedPlayerCantActivateNonMana is "enchanted player can't
// activate abilities that aren't mana abilities or loyalty abilities"
// (CR 602.5): a restriction on the Aura that forbids every activation by the
// enchanted player except a mana ability (CR 605.1a) or a loyalty ability.
func OldB239EnchantedPlayerCantActivateNonMana(label string) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			player, ok := OldB239EnchantedPlayer(&q.Source)
			if !ok || q.Controller != player {
				return false
			}
			return !q.Ability.Mana && !q.Ability.Loyalty
		},
	}
}
