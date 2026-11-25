package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b248_helpers.go — shared pieces of Amonkhet batch 248 (#5535).

// OldB248YouHaveAtMostCardsInHand is "you have N or fewer cards in hand" for
// `you`, read live from the hand's size (Neheb, the Worthy; Hazoret the
// Fervent).
func OldB248YouHaveAtMostCardsInHand(n int) func(g *game.Game, you uuid.UUID) bool {
	return func(g *game.Game, you uuid.UUID) bool {
		p := g.PlayerByIDForEffect(you)
		return p != nil && p.Hand != nil && p.Hand.Size() <= n
	}
}

// CantAttackUnlessYouHaveAtMostCardsInHand is "This creature can't attack
// unless you have N or fewer cards in hand" (Hazoret the Fervent): the attack
// half of the sentence, a fact about the attacking creature's controller read
// live at declaration (CR 508.1c).
func CantAttackUnlessYouHaveAtMostCardsInHand(n int) game.StaticAbility {
	s := selfAttackTargetRestriction(game.AttackTargetRestriction{ControllerHandAtMost: n})
	s.Label = "Can't attack unless you have " + NumberWord(n) + " or fewer cards in hand."
	return s
}

// OldB248PreventOneFromOpponentsToYou is Protection of the Hekma's "If a source
// an opponent controls would deal damage to you, prevent 1 of that damage."
// The source is read as the damage event snapshotted it (CR 608.2h).
func OldB248PreventOneFromOpponentsToYou() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget != src.Controller {
				return false
			}
			ch, ok := DamageSourceCharacteristics(ev, g)
			return ok && ch.Controller != uuid.Nil && ch.Controller != src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount--
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      "Protection of the Hekma — prevent 1 of that damage",
		RowLabel:   "If a source an opponent controls would deal damage to you, prevent 1 of that damage.",
	}
}

// OldB248PreventDamageBySourcesNamedChosen is Gideon's Intervention's "Prevent
// all damage that would be dealt to you and permanents you control by sources
// with the chosen name": the name is the one chosen as the permanent entered,
// the source is read as the damage event snapshotted it (CR 608.2h), and a
// recipient is you or a permanent you control on the battlefield.
func OldB248PreventDamageBySourcesNamedChosen() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
				return false
			}
			named := strings.TrimSpace(src.ChosenName)
			if named == "" {
				return false
			}
			ch, ok := DamageSourceCharacteristics(ev, g)
			if !ok || !game.CardNameMatches(game.Card{Name: ch.Name}, named) {
				return false
			}
			if ev.DamageTarget == src.Controller {
				return true
			}
			c, ok := g.LookupCardForEffect(ev.DamageTarget)
			return ok && g.Battlefield.Contains(ev.DamageTarget) && c.Controller == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      "Gideon's Intervention — prevent all damage dealt to you and your permanents by sources with the chosen name",
		RowLabel:   "Prevent all damage that would be dealt to you and permanents you control by sources with the chosen name.",
	}
}
