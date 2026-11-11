package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b19_helpers.go — shared bodies for Commander wave 2, batches 19
// and 20 (#5368).

// cmdr2B19AttachedCantAttackYou is "Enchanted creature … can't attack you
// or planeswalkers you control" (Vow of Duty, Vow of Lightning, CR 508.1c):
// a layer 6 restriction on the attached creature naming the Aura's
// controller, read when the layers run, so a stolen Aura protects its new
// controller. A battle that player protects is not covered.
func cmdr2B19AttachedCantAttackYou() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Enchanted creature can't attack you or planeswalkers you control.",
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source == nil {
				return
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
				Source:     source.InstanceID,
				SourceName: source.Name,
				NotPlayer:  source.Controller,
			})
		},
	}
}

// cmdr2B19HumanMonkToken is Careful Cultivation's 1/1 green Human Monk
// creature token with "{T}: Add {G}."
func cmdr2B19HumanMonkToken() game.Card { return tokenFromCatalog(printedCmdr2B19HumanMonkToken) }

// printedCmdr2B19HumanMonkToken is that token as PRINTED (Neon Dynasty's
// token card, tneo 10). Its slug is an on-disk identity.
func printedCmdr2B19HumanMonkToken() tokenTemplate {
	return tokenTemplate{
		Slug: "human-monk-tap-g",
		Card: game.Card{
			Name:      "Human Monk",
			TypeLine:  "Token Creature — Human Monk",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"G"},
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{G}",
			Label:    "{T}: Add {G}",
		}},
		Text: "{T}: Add {G}.",
	}
}

// cmdr2B19AttachedIsCreature is the condition of Careful Cultivation's
// "As long as enchanted permanent is a creature": the host as the earlier
// layers have left it this pass.
func cmdr2B19AttachedIsCreature(target *game.Card, g *game.Game, source *game.Card) bool {
	return AttachedToSource(target, g, source) && target.IsCreature()
}

// cmdr2B19AttachedAssignsAsThoughUnblocked is "Enchanted creature's
// controller may have it assign its combat damage as though it weren't
// blocked" (Indomitable Might, CR 510.1c): the creature-side flag Wolverine,
// Claws Out sets on itself, set here on the attached creature.
func cmdr2B19AttachedAssignsAsThoughUnblocked() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Enchanted creature's controller may have it assign its combat damage as though it weren't blocked.",
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.AssignsAsThoughUnblocked = true
		},
	}
}

const cmdr2B19DaybreakCoronetOracle = "0f5ac364-447b-4c43-9aed-6098dc0bab82"

// cmdr2B19HasAnotherAuraAttached is Daybreak Coronet's "Enchant creature
// with another Aura attached to it". An enchant clause is one predicate on
// the candidate host that serves both the cast and the CR 704.5m check
// that reruns it once the Coronet is attached, and neither knows which
// Aura it is judging. It is answered from the host's attachments instead:
//
//   - any attached Aura that is not a Daybreak Coronet is "another Aura" to
//     every Coronet, so the host is legal;
//   - with only Coronets attached, each is another Aura to the rest, so two
//     of them hold each other up and a lone one (the other Aura left) does
//     not: the host is legal only with two or more.
//
// A lone Coronet is the only state where the two readings differ (the
// Coronet itself is on the host at the check, and is not yet at the cast),
// and a lone Coronet on an otherwise bare creature is exactly the state
// the rule removes. Auras are read as they are now (effective types).
func cmdr2B19HasAnotherAuraAttached() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		if g == nil || g.Battlefield == nil {
			return false
		}
		coronets := 0
		for i := range g.Battlefield.Cards {
			a := &g.Battlefield.Cards[i]
			if !a.IsAura() || !a.IsAttachedTo(c.InstanceID) {
				continue
			}
			if a.OracleID != cmdr2B19DaybreakCoronetOracle {
				return true
			}
			coronets++
		}
		return coronets >= 2
	}
}

// cmdr2B19AurasAttachedTo counts the Auras attached to a permanent.
// It runs inside a layer pass that may hold only the read lock, so it
// walks the battlefield directly (see PumpAttachedPer). Auras are read as
// the layers have them, so an enchantment that became an Aura counts.
func cmdr2B19AurasAttachedTo(g *game.Game, id uuid.UUID) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.IsAura() && a.IsAttachedTo(id) {
			n++
		}
	}
	return n
}

// cmdr2B19AurasOnTheBattlefield counts every Aura on the battlefield, any
// controller's (Aura Gnarlid), by the same walk.
func cmdr2B19AurasOnTheBattlefield(g *game.Game) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].IsAura() {
			n++
		}
	}
	return n
}

// cmdr2B19LesserPowerThanAttacker is "creatures with power less than this
// creature's power": a blocker whose power is below the attacker's, both
// read as the comparison rules do (CR 208.1), live as the block is
// declared. The attacker is the creature itself, so a pump (Aura Gnarlid's
// own) raises the bar.
func cmdr2B19LesserPowerThanAttacker() BlockerAgainstAttacker {
	return func(_ *game.Game, attacker, blocker *game.Card) bool {
		return blocker.PowerForComparison() < attacker.PowerForComparison()
	}
}

// cmdr2B19YourAuraBecameAttachedToYourCreature is "Whenever an Aura you
// control becomes attached to a creature you control": both read as the
// Aura attaches.
func cmdr2B19YourAuraBecameAttachedToYourCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttach || ev.Label != "Aura" || ev.AlreadyAttached || source == nil {
		return false
	}
	aura, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || aura.Controller != source.Controller {
		return false
	}
	host, ok := g.LookupCardForEffect(ev.Target)
	return ok && host.IsCreature() && host.Controller == source.Controller && host.Controller != uuid.Nil
}

// cmdr2B19OfferAuraForNextToken is Liberated Livestock's walk over its
// tokens: for the first token in `tokens` that is still on the battlefield,
// ask `player` to pick an Aura card from their hand and/or graveyard (or
// none), put it onto the battlefield attached to that token, and go on to
// the rest once the entry is complete. A token with no Aura card to offer is
// passed over without a question. The tokens still to come ride the prompt
// and the tail as data (Carry).
func cmdr2B19OfferAuraForNextToken(g *game.Game, player, source uuid.UUID, tokens []uuid.UUID) error {
	for len(tokens) > 0 {
		token := tokens[0]
		rest := tokens[1:]
		tokens = rest
		if c, ok := g.LookupCardForEffect(token); !ok || !c.IsCreature() {
			continue
		}
		p := g.PlayerByIDForEffect(player)
		if p == nil || p.Eliminated {
			return nil
		}
		var cands []uuid.UUID
		for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
			if z == nil {
				continue
			}
			for _, c := range z.Cards {
				if c.IsAura() && !c.IsToken() {
					cands = append(cands, c.InstanceID)
				}
			}
		}
		if len(cands) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Destination: game.ChooseOntoBattlefield,
			Chooser:     player,
			Source:      source,
			Question:    "Liberated Livestock — you may put an Aura card from your hand and/or graveyard onto the battlefield attached to that token",
			Cards:       cands,
			Min:         0,
			Max:         1,
			ThenKey:     cmdr2B19LivestockPick,
			Carry:       append([]uuid.UUID{token}, rest...),
		})
		return nil
	}
	return nil
}

// cmdr2B19LivestockPick is the pick's continuation: Carry is the token this
// offer is for, then the tokens still to come.
var cmdr2B19LivestockPick game.ChooseCardsThen

// cmdr2B19LivestockNext runs once the put has entered (or been refused) and
// offers the next token its Aura.
var cmdr2B19LivestockNext game.TailRef

// The pick and the tail refer to each other through the walk, so they are
// registered here rather than by package-level initializers (an
// initialization cycle).
func init() {
	cmdr2B19LivestockPick = game.RegisterChooseCardsThen("choose-cards/cmdr2-b19-liberated-livestock-aura",
		func(g *game.Game, r game.CardsChosen) error {
			if len(r.Carry) == 0 {
				return nil
			}
			token, rest := r.Carry[0], r.Carry[1:]
			if len(r.Picked) == 0 {
				return cmdr2B19OfferAuraForNextToken(g, r.Chooser, r.Source, rest)
			}
			from := game.ZoneHand
			if z := g.FindCardZoneForEffect(r.Picked[0]); z != nil {
				from = z.Kind
			}
			return g.PutOntoBattlefieldTogetherKeyedForEffect(
				[]game.BatchEntry{{CardID: r.Picked[0], From: from}},
				game.ZoneEntryOptions{Controller: r.Chooser, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: token}},
				game.KeyedTail{Ref: cmdr2B19LivestockNext, Player: r.Chooser, Source: r.Source, Carry: rest})
		})
	cmdr2B19LivestockNext = CardTail("cmdr2-b19/liberated-livestock-next", func(ctx *Context, o game.TailOutcome) error {
		return cmdr2B19OfferAuraForNextToken(ctx.Game, o.Tail.Player, o.Tail.Source, o.Tail.Carry)
	})
}
