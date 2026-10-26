package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ardenn, Intrepid Archaeologist — Legendary Creature — Kor Scout
// {2}{W}, 2/2 (the Squall deck request, #3918):
//
//	"At the beginning of combat on your turn, you may attach any number
//	 of Auras and Equipment you control to target permanent or player.
//	 Partner"
//
// Partner is a deck-construction rule internal/deck reads off the
// oracle text (ADR 0144); nothing here.
//
// The target is chosen as the trigger goes on the stack: any permanent
// or any player, hexproof and protection judged as usual. As it
// resolves, the controller picks any number of the Auras and Equipment
// they control, none being the "you may", and each is attached to the
// still-legal target (CR 608.2b).
//
// Only an attachment that could legally be attached there is offered
// (CR 701.3b): an Equipment only to a creature, an Aura only to what its
// enchant clause and the host's protection allow, never a permanent to
// itself. A player can carry only an Aura that enchants a player;
// Equipment is never offered for a player. The attach itself re-checks
// the same rule, and one already on the target stays where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c521612-21c6-4c4b-b10e-66e70928609b",
		Name:         "Ardenn, Intrepid Archaeologist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Ardenn, Intrepid Archaeologist — attach any number of Auras and Equipment you control to target permanent or player",
					ardennAttach),
				ardennTargetPermanentOrPlayer()),
		},
	})
}

// ardennTargetPermanentOrPlayer is "target permanent or player".
func ardennTargetPermanentOrPlayer() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:    "any",
		Label:   "target permanent or player",
		Players: true,
		Zones:   []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, _ game.Card, _ game.ZoneKind) bool {
			return true
		},
		Min: 1, Max: 1,
	}
}

// ardennAttach offers the Auras and Equipment that could go on the
// target and attaches the ones picked.
func ardennAttach(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	host := legal[0]
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || c.InstanceID == host.ID {
			continue
		}
		isAura, isEquipment := c.HasSubtype("Aura"), c.HasSubtype("Equipment")
		if !isAura && !isEquipment {
			continue
		}
		if c.AttachedTo == host {
			continue // already there: attaching it again does nothing
		}
		switch host.Kind {
		case game.TargetCard:
			if !g.AttachLegalForEffect(c.InstanceID, host.ID) {
				continue
			}
		case game.TargetPlayer:
			if !isAura || !ardennAuraEnchantsPlayer(g, c, host.ID) {
				continue
			}
		default:
			continue
		}
		candidates = append(candidates, c.InstanceID)
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Ardenn, Intrepid Archaeologist — attach any number of Auras and Equipment you control to that " + ardennHostWord(host),
		Cards:    candidates,
		Min:      0,
		Max:      len(candidates),
		ThenKey:  ardennAttachPick,
		LiveZone: game.ZoneBattlefield,
		// The host the pick attaches to: its ID is the carry and its kind
		// ("permanent" or "player") the word.
		Carry:     []uuid.UUID{host.ID},
		CarryWord: ardennHostWord(host),
	})
	return nil
}

// ardennAttachPick is the picked Auras and Equipment being attached to the
// host, keyed (#4253): the host is plain data on the prompt.
var ardennAttachPick = game.RegisterChooseCardsThen("choose-cards/ardenn-intrepid-archaeologist-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	host := game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]}
	if r.Word == "player" {
		host.Kind = game.TargetPlayer
	}
	for _, id := range r.Picked {
		if err := g.AttachForEffect(id, host); err != nil {
			return err
		}
	}
	return nil
})

// ardennAuraEnchantsPlayer reports whether an Aura's enchant clause
// admits this player (a Curse's "Enchant player", "Enchant opponent"),
// judged for the Aura's controller. Protection from the Aura is the
// attach's own re-check.
func ardennAuraEnchantsPlayer(g *game.Game, c game.Card, player uuid.UUID) bool {
	spec, ok := Lookup(game.CatalogKey(c))
	if !ok || spec.Targets == nil || !spec.Targets.Players {
		return false
	}
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return false
	}
	return spec.Targets.PlayerOK == nil || spec.Targets.PlayerOK(g, c.Controller, p)
}

func ardennHostWord(host game.TargetRef) string {
	if host.Kind == game.TargetPlayer {
		return "player"
	}
	return "permanent"
}
