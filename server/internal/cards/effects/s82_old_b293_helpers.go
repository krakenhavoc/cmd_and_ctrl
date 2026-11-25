package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b293_helpers.go — shared pieces of Shadows over Innistrad, S82 OLD
// batch 293.

// OldB293ProducedFromGraveyardColors is Corrupted Grafstone's "Choose a color
// of a card in your graveyard. Add one mana of that color.": the pipe string
// of every colour among the cards in the controller's graveyard, so the player
// picks one of them. Colourless cards contribute nothing (CR 105.1), and an
// empty or colourless graveyard produces no mana.
func OldB293ProducedFromGraveyardColors() func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, controller, _ uuid.UUID) string {
		p := g.PlayerByIDForEffect(controller)
		if p == nil || p.Graveyard == nil {
			return ""
		}
		seen := map[string]bool{}
		for i := range p.Graveyard.Cards {
			for _, col := range p.Graveyard.Cards[i].EffectiveColors() {
				seen[strings.ToUpper(col)] = true
			}
		}
		return PipeString(ColorsOnly(OrderedManaSymbols(seen)))
	}
}

// OldB293CantBlockUnlessYouControlAnother is "This creature can't block unless
// you control another <subtype or subtype>" (Howlpack Wolf; CR 509.1b): a
// block rule on the creature itself, asked of its controller live as blockers
// are declared. "Another" excludes the creature, so a lone Wolf cannot count
// itself.
func OldB293CantBlockUnlessYouControlAnother(label string, queries ...game.PermanentQuery) game.BlockRule {
	self := OnSelf()
	return game.BlockRule{
		Reason: game.BlockReasonCantBlockAttacker,
		Label:  label,
		Pair: func(g *game.Game, _, blocker, source *game.Card) bool {
			return self(g, blocker, source) && !g.ControlsMatchingForEffect(source.Controller, source.InstanceID, queries...)
		},
	}
}

// OldB293CantAttackUnlessYouHaveCastANoncreatureSpell is "This creature can't
// attack unless you've cast a noncreature spell this turn" (Silburlind
// Snapper; CR 508.1c): a fact about the creature's controller read live at
// declaration, so it refuses every target while unmet.
func OldB293CantAttackUnlessYouHaveCastANoncreatureSpell() game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{ControllerMustHaveCastNoncreature: true})
}
