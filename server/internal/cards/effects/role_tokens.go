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
// Declared here: Monster, Cursed, Royal, Wicked, Sorcerer, Young Hero,
// Virtuous and Chef. Monster, Cursed, Royal and Wicked are a static
// (+P/+T, a base-P/T set, trample) or a ward, plus Wicked's own "put
// into a graveyard" trigger. Sorcerer, Young Hero and Chef give the
// enchanted creature a TRIGGERED ability whose source is the creature,
// not the Role ("Whenever this creature attacks, scry 1."): each is a
// layer-6 grant of a bundle the token template declares in Grants
// (ADR 0093, amended 2026-10-08), so "this creature" is the host by
// construction and the trigger is the creature's own, controlled by
// the creature's controller. Virtuous is a layer 7c count of the
// enchantments the Role's controller controls.
//
// Questing ("has all the abilities of Questing Beast") is not
// declared: Questing Beast's block restriction and its combat-damage
// prevention rule are card slots (BlockRules, DamageCantBePrevented),
// and an ability bundle has neither.
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

	RoleSorcerer  RoleKind = "sorcerer"
	RoleYoungHero RoleKind = "young-hero"
	RoleVirtuous  RoleKind = "virtuous"
	RoleChef      RoleKind = "chef"
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
	case RoleSorcerer:
		return printedSorcererRoleToken
	case RoleYoungHero:
		return printedYoungHeroRoleToken
	case RoleVirtuous:
		return printedVirtuousRoleToken
	case RoleChef:
		return printedChefRoleToken
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

const (
	sorcererRoleGrant  = "sorcerer-role/scry"
	youngHeroRoleGrant = "young-hero-role/counter"
	chefRoleGrant      = "chef-role/food"
)

func printedSorcererRoleToken() tokenTemplate {
	return tokenTemplate{
		Slug:   "sorcerer-role",
		Card:   roleCard("Sorcerer Role", "U"),
		Static: []game.StaticAbility{PumpAttached(1, 1), GrantAbilitiesToAttached(sorcererRoleGrant)},
		Grants: []AbilityGrant{{
			Key: sorcererRoleGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				}, "Sorcerer Role — scry 1", func(g *game.Game, item *game.StackItem) error {
					return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
			},
			Text: "Whenever this creature attacks, scry 1.",
		}},
		Text: "Enchant creature\nEnchanted creature gets +1/+1 and has \"Whenever this creature attacks, scry 1.\"",
	}
}

func printedYoungHeroRoleToken() tokenTemplate {
	return tokenTemplate{
		Slug:   "young-hero-role",
		Card:   roleCard("Young Hero Role", "W"),
		Static: []game.StaticAbility{GrantAbilitiesToAttached(youngHeroRoleGrant)},
		Grants: []AbilityGrant{{
			Key: youngHeroRoleGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					// Intervening-if, first check (CR 603.4).
					return attackDeclared(ev, source) && source.CurrentToughness() <= 3
				}, "Young Hero Role — a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
					// Second check, on resolution: a pump in response
					// that lifts the toughness past 3 stops the counter.
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || c.CurrentToughness() > 3 {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
			},
			Text: "Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it.",
		}},
		Text: "Enchant creature\nEnchanted creature has \"Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it.\"",
	}
}

func printedVirtuousRoleToken() tokenTemplate {
	return tokenTemplate{
		Slug:   "virtuous-role",
		Card:   roleCard("Virtuous Role", "W"),
		Static: []game.StaticAbility{PumpAttachedPer(1, 1, enchantmentsControlledBy)},
		Text:   "Enchant creature\nEnchanted creature gets +1/+1 for each enchantment you control.",
	}
}

func printedChefRoleToken() tokenTemplate {
	return tokenTemplate{
		Slug:   "chef-role",
		Card:   roleCard("Chef Role", "G"),
		Static: []game.StaticAbility{PumpAttached(1, 1), GrantAbilitiesToAttached(chefRoleGrant)},
		Grants: []AbilityGrant{{
			Key: chefRoleGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				}, "Chef Role — create a Food token", func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
				}),
			},
			Text: "Whenever this creature attacks, create a Food token.",
		}},
		Text: "Enchant creature\nEnchanted creature gets +1/+1 and has \"Whenever this creature attacks, create a Food token.\"",
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
