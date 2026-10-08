package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// role_tokens.go — Role tokens (CR 111.10, 303.7), #1945.
//
// A Role is a token Aura: "Token Enchantment — Aura Role" with an
// implicit "enchant creature". It is created already attached
// (CR 303.7), and a player controlling two Roles on one permanent
// keeps only the newest (CR 303.7a, enforced by the CR 704.5z branch
// of game.attachmentSBALocked). The Role subtype, the attach and the
// state-based action are engine; this file is the printed token
// definitions and the CreateRoleToken primitive that makes one.
//
// Declared here: Monster, Cursed, Royal and Wicked. Their text is a
// static (+P/+T, a base-P/T set, trample) or a ward granted to the
// enchanted creature, plus Wicked's own "put into a graveyard"
// trigger — every one is an existing shape. Sorcerer, Young Hero,
// Virtuous, Questing and Chef each GRANT the enchanted creature a
// triggered ability or count enchantments; none is declared, so a
// card making one stays off the catalog.
//
// # A Role is not created when it has nowhere legal to go
//
// CR 303.4f / 303.7: a Role token created attached to a permanent
// that is not a creature on the battlefield is not created, rather
// than entering and being put into the graveyard a beat later. The
// player sees the same board either way; the difference is a Wicked
// Role's graveyard trigger, which must not drain anyone for a token
// that never existed.

// RoleKind names a Role token definition.
type RoleKind string

const (
	RoleMonster RoleKind = "monster"
	RoleCursed  RoleKind = "cursed"
	RoleRoyal   RoleKind = "royal"
	RoleWicked  RoleKind = "wicked"
)

func roleBuilder(k RoleKind) tokenTemplateBuilder {
	switch k {
	case RoleMonster:
		return printedMonsterRoleToken
	case RoleCursed:
		return printedCursedRoleToken
	case RoleRoyal:
		return printedRoyalRoleToken
	case RoleWicked:
		return printedWickedRoleToken
	}
	panic("effects.CreateRoleToken: unknown role " + string(k))
}

func roleCard(name string, colors ...string) game.Card {
	return game.Card{
		Name:     name,
		TypeLine: "Token Enchantment — Aura Role",
		Colors:   colors,
	}
}

func printedMonsterRoleToken() tokenTemplate {
	return tokenTemplate{
		Slug:   "monster-role",
		Card:   roleCard("Monster Role", "R"),
		Static: []game.StaticAbility{PumpAttached(1, 1), GrantToAttached("trample")},
		Text:   "Enchant creature\nEnchanted creature gets +1/+1 and has trample.",
	}
}

func printedCursedRoleToken() tokenTemplate {
	return tokenTemplate{
		Slug:   "cursed-role",
		Card:   roleCard("Cursed Role", "B"),
		Static: []game.StaticAbility{SetAttachedBasePT(1, 1)},
		Text:   "Enchant creature\nEnchanted creature is 1/1.",
	}
}

func printedRoyalRoleToken() tokenTemplate {
	return tokenTemplate{
		Slug:   "royal-role",
		Card:   roleCard("Royal Role", "W"),
		Static: []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{1}"), "Royal Role — ward {1}", AttachedToSource),
		},
		Text: "Enchant creature\nEnchanted creature gets +1/+1 and has ward {1}.",
	}
}

func printedWickedRoleToken() tokenTemplate {
	return tokenTemplate{
		Slug:   "wicked-role",
		Card:   roleCard("Wicked Role", "B"),
		Static: []game.StaticAbility{PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Wicked Role — each opponent loses 1 life", func(g *game.Game, item *game.StackItem) error {
				return eachOpponentLosesLife(g, item, 1)
			}),
		},
		Text: "Enchant creature\nEnchanted creature gets +1/+0.\nWhen this token is put into a graveyard, each opponent loses 1 life.",
	}
}

// RoleToken is the named Role's template, for a test or a card that
// needs the card value rather than the primitive.
func RoleToken(k RoleKind) game.Card { return tokenFromCatalog(roleBuilder(k)) }

// CreateRoleToken is "create a <Role> token attached to <Host>".
// Controller defaults to the ability's controller. A Host that is not
// a creature on the battlefield creates nothing (see the file
// comment). The Role enters attached and takes a fresh attach
// timestamp, so it is the newest Role on the host (CR 303.7a) and any
// older Role the same player controls there is put into the graveyard
// by the next state-based check.
type CreateRoleToken struct {
	Role       RoleKind
	Host       uuid.UUID
	Controller uuid.UUID
}

func (r CreateRoleToken) Apply(ctx *Context) error {
	g := ctx.Game
	z := g.FindCardZoneForEffect(r.Host)
	if r.Host == uuid.Nil || z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	creature := false
	for i := range z.Cards {
		if z.Cards[i].InstanceID == r.Host {
			creature = z.Cards[i].IsCreature()
		}
	}
	if !creature {
		return nil
	}
	controller := r.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	host := game.TargetRef{Kind: game.TargetCard, ID: r.Host}
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: controller,
		Source:     ctx.Source(),
		Groups: []game.TokenGroup{{
			Template: tokenFromCatalog(roleBuilder(r.Role)),
			Count:    1,
		}},
	}, func(g *game.Game, created []uuid.UUID) error {
		for _, id := range created {
			if err := g.AttachForEffect(id, host); err != nil {
				return err
			}
		}
		return nil
	})
}

// AttachedToThis is the sacrifice-cost clause "an Aura attached to
// this creature" (Faunsbane Troll): the activator's own permanents that
// match `preds`, restricted to those attached to the paying permanent.
func AttachedToThis(label string, preds ...CardPredicate) *game.TargetSpec {
	spec := sacrificeSpec(label, preds...)
	spec.AttachedToSource = true
	return spec
}
